# Rebuild Plan — Thread Art Generator

Source de vérité pour remettre l’app dans la vision : **gRPC partout**, **storage S3-compatible**, **cloud-agnostic**, **auth simple qui cohabite avec les signed URLs**, **interfaces d’abord**.

Ce fichier se met à jour. La section [Journal](#journal) est append-only.

---

## 0. Vision

Produit : upload d’images → compositions thread-art → preview / GCode.

Contraintes non négociables :

- Le **contrat** c’est protobuf + AIP (resource names, standard methods, field masks, pagination opaque).
- Le **transport** c’est gRPC. Browser en **gRPC-Web binaire**. Inter-services en **gRPC natif (h2c en local, h2+TLS en prod)**.
- Le **bucket** parle **S3**. Local = MinIO. Prod = R2 ou S3. Pas de Firebase Storage, pas de GCS SDK.
- **Aucune** logique `if emulator` dans le code applicatif. Un driver, des env vars.
- Identité applicative **orthogonale** aux signed URLs. L’URL signée *est* le token d’accès objet. L’auth app décide *qui* a le droit d’en émettre une.
- Max d’**interfaces** dans `core/`. Les `cmd/` et `client/` ne parlent qu’aux interfaces.

Non-buts (volontaires) :

- Pas de retour à Next.js / React pour débloquer le gRPC. HTMX + islands TS + connect-web.
- Pas d’Envoy. Connect-Go parle gRPC + gRPC-Web + Connect sur le même handler.
- Pas de Workers Cloudflare comme runtime de l’API Go. R2 = bucket + CDN. L’API reste un binaire Go.
- Pas de smart-router Laravel tant que le contrat API n’est pas stable. `URLFor()` suffit.
- Pas de transcoding REST `google.api.http` au runtime. OpenAPI optionnel plus tard, pas un contrat.

---

## 1. Architecture cible

```
                    ┌─────────────────────────────────────┐
  Browser           │  Traefik / same-origin              │
  HTML (HTMX)  ────►│  :8080  pages, session cookie       │
  JS islands   ────►│  /rpc   → API gRPC-Web binaire      │
  PUT bytes    ────►│  MinIO/R2/S3  (presigned URL)       │
                    └──────────────┬──────────────────────┘
                                   │ cookie (session)
                                   ▼
                    ┌─────────────────────────────────────┐
                    │  API  :9090                         │
                    │  Connect handler ×3 protocoles      │
                    │  h2c: gRPC + gRPC-Web + Connect     │
                    │  interceptors: session | svc creds  │
                    └──────┬──────────┬──────────┬────────┘
                           │          │          │
                     Postgres      Bucket      Queue
                           │          │          │
                           │          │          ▼
                           │          │     Worker (gRPC interne
                           │          │     ou lecture queue
                           │          │     + Bucket.Get/Put)
                           │          │
                    MinIO | R2 | S3   (même API S3)
```

Réseau local : `docker-compose` = `db` + `minio` + `redis` + `api` + `worker` + `client`. Plus de Firebase emulator, plus de Java, plus de Pub/Sub emulator.

Prod (jour 1 cloud) :

| Pièce | Cloudflare | AWS |
|---|---|---|
| Compute API/worker | Fly / CF Containers plus tard | ECS/Fargate |
| Postgres | Neon + Hyperdrive (si CF) ou RDS | RDS |
| Bucket | **R2** | **S3** |
| CDN | R2 custom domain | CloudFront |
| Queue | Postgres `SKIP LOCKED` ou NATS | SQS *via l’interface* `queue.Queue` |

Le code ne connaît pas ces noms. Il connaît `Bucket`, `Queue`, `Sessions`, `Identities`.

---

## 2. Protocole — gRPC full

### Serveur

Un seul `http.Server` Connect :

- HTTP/1.1 + **h2c** (`Protocols.SetUnencryptedHTTP2(true)`).
- Timeouts : `ReadHeaderTimeout`, `IdleTimeout`. Plus de `ListenAndServe` nu.
- Reflection (`connectrpc.com/grpcreflect`) + health (`connectrpc.com/grpchealth`).
- CORS : **interdit** sur l’API. Same-origin via proxy `/rpc`.

Handlers = interfaces Connect directement. **Supprimer** `ConnectAdapter` (wrap mort du gRPC-go).

### Clients

| Hop | Transport | Auth |
|---|---|---|
| Browser → API | `createGrpcWebTransport({ useBinaryFormat: true, credentials: 'include' })` | session cookie |
| BFF → API (si encore des reads HTML) | `connect.WithGRPC()` | cookie forward **ou** identity metadata depuis session déjà validée côté BFF, sur réseau privé |
| Worker → API | `connect.WithGRPC()` | `ServiceCredential` (HMAC) |
| CLI / grpcurl | gRPC plaintext local | Bearer session ou rien en dev |

### Proto / AIP (contrat)

Package proto reste `pb` en phase 1 (blast radius). `buf managed` fixe `go_package`. Plus tard : `threadart.v1`.

À honorer (aujourd’hui violé) :

- `parent` sur Create/List — 403 si ≠ identity.
- `update_mask` appliqué (einride `fieldmask`).
- Pagination einride, tokens **opaques signés** (pas un offset `"10"`).
- `page_size = 0` → default. CEL actuel `gt: 0` à corriger (`gte: 0` ou pas de contrainte, clamp serveur).
- `order_by` unique style AIP-132 (`create_time desc`), drop `order_direction`.
- Resource type : plus `art.example.com`. `threadart.local/Art` (dev) / domaine réel en prod.
- HTTP mappings Users cassés (`/v1/users/{name}` avec `name=users/x`) : **les dropper** (pas de gateway).
- `ListUsers` / `DeleteUser` : `Unimplemented` ou implémentés. Plus de no-op 200.
- Storage n’est **pas** un resource AIP plat `user_id`+`art_id`. Custom method :

```
rpc StartArtUpload (StartArtUploadRequest) returns (StartArtUploadResponse);
  // name = users/{user}/arts/{art}

rpc CompleteArtUpload (CompleteArtUploadRequest) returns (Art);
```

Internal (`SyncUserFromFirebase`, `ConfirmArtImageUploadFromFunction`) : service `Internal` séparé **ou** suppression (signup crée le user, `CompleteArtUpload` remplace la Function).

Erreurs : **uniquement** `*connect.Error` + `errdetails` (AIP-193). Interceptor protovalidate. Drop `google.golang.org/grpc/status` dans les services.

---

## 3. Génération proto — boring

Cause actuelle : deux templates, plugins locaux `@latest`, TS vers `web/` mort, `connect-es` v1 disparu.

Cible : **un** `proto/buf.gen.yaml` v2, **remote plugins pinés**, zéro binaire `protoc-gen-*` local.

```yaml
# proto/buf.yaml
version: v2
modules:
  - path: .
deps:
  - buf.build/googleapis/googleapis
  - buf.build/bufbuild/protovalidate
lint:
  use: [STANDARD]
breaking:
  use: [FILE]
```

```yaml
# proto/buf.gen.yaml
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/Damione1/thread-art-generator/core/pb
plugins:
  - remote: buf.build/protocolbuffers/go:v1.36.6
    out: ../core/pb
    opt: paths=source_relative
  - remote: buf.build/connectrpc/go:v1.18.1
    out: ../core/pb
    opt: paths=source_relative
  - remote: buf.build/bufbuild/es:v2.10.0
    out: ../client/src/gen
    opt: target=ts
```

`make proto` = `cd proto && buf generate && buf lint`.
CI : `buf breaking --against '.git#branch=main'`.
Tilt : watch `proto/**/*.proto` uniquement. Drop watch `client/internal/pb`.
Supprimer `proto/buf.gen.make.yaml`.
Drop plugin `openapiv2` (swagger de routes inexistantes).

---

## 4. Storage — S3 only

### Interface (`core/storage`)

Étroite. Pas de `GetPublicURL` sur le driver. Pas de validation MIME dans le driver.

```go
type Bucket interface {
    Put(ctx context.Context, key string, r io.Reader, opts PutOptions) error
    Get(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error)
    Head(ctx context.Context, key string) (*ObjectInfo, error)
    Delete(ctx context.Context, key string) error
    PresignPut(ctx context.Context, key string, opts PresignPutOptions) (*PresignPut, error)
    PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type PresignPut struct {
    URL     string
    Method  string            // PUT
    Headers http.Header       // headers que le browser DOIT rejouer (Content-Type)
    Expires time.Time
}

type PresignPutOptions struct {
    ContentType string
    MaxBytes    int64
    TTL         time.Duration
}
```

Public URL = `config.PublicBaseURL + "/" + key` dans la couche API, pas dans le driver.

### Implémentation

Un seul driver : **AWS SDK v2** `s3` + `PresignClient`.

Deux clients internes (le split de `3bd30e7` qui était juste) :

| Client | Endpoint | Rôle |
|---|---|---|
| `ops` | `S3_ENDPOINT` (`minio:9000`, `*.r2.cloudflarestorage.com`, AWS default) | Head/Get/Put/Delete serveur |
| `sign` | host de `S3_PUBLIC_BASE_URL` | Presign uniquement |

`S3_FORCE_PATH_STYLE=true` pour MinIO/R2. `false` pour AWS.

Keys (serveur only) :

```
users/{user_id}/arts/{art_id}/original
users/{user_id}/arts/{art_id}/compositions/{composition_id}/preview
users/{user_id}/arts/{art_id}/compositions/{composition_id}/gcode
users/{user_id}/arts/{art_id}/compositions/{composition_id}/pathlist
```

DB stocke le **key**, jamais l’URL.

### Flow upload

1. `CreateArt` → `PENDING_IMAGE`.
2. `StartArtUpload(name)` → auth + ownership + status → `PresignPut` (TTL 10 min, Content-Type exact, key fixé).
3. Browser `PUT` vers MinIO/R2/S3 avec **les headers retournés**. Progress XHR. **Pas de PATCH.**
4. `CompleteArtUpload(name)` → `Head` : exists, size ≤ max, content-type image → status `COMPLETE`.

Plus de Cloud Function `onFinalize`. Plus de `user_id` dans le body.

Reads ListArts : concat public URL, **zéro** I/O storage dans `pbx`.

### Compose local

MinIO `:9000` API, `:9001` console. Bucket `thread-art` créé au boot. Policy : public-read sur prefix arts **ou** tout privé + presign GET (préférer public-read sur `*/original` et `*/preview` pour le CDN ; gcode/pathlist privés).

---

## 5. Auth — drop PASETO, pas de 4e stack

### Verdict

PASETO v2.local n’est **pas** le mauvais algo. C’est le mauvais *rôle* ici :

- Régénéré à **chaque** RoundTrip BFF.
- Fallback Firebase ID token que l’API refuse.
- Doublon `core/token` (mort) + `core/auth`.
- Écosystème browser / grpcurl / connect-web = **JWT ou cookies**, pas PASETO.
- Les signed URLs S3 **ne doivent pas** connaître PASETO/JWT. HMAC bucket, point.

Pas de Keycloak / Ory / Zitadel / Clerk. Trop lourd, lock-in, et ça n’aide pas les presign.

### Bundle retenu : cookies + S3 presign + service HMAC

Trois jetons, trois jobs, zéro recouvrement :

| Jeton | Quoi | Où |
|---|---|---|
| **Session cookie** httpOnly, Secure, SameSite=Lax, Redis store (SCS déjà là, Redis déjà dans compose, **non branché**) | Identité user browser ↔ API / BFF | `core/auth/session` |
| **S3 presign** | Accès **un** objet, **une** méthode, TTL court | `core/storage` |
| **Service HMAC** (`Authorization: Service <id>:<mac>`) | Worker / jobs internes | `core/auth/service` |

Librairies :

- Sessions : `alexedwards/scs/v2` + `scs/redisstore` (déjà deps). Fallback mémoire si Redis down en dev.
- Interceptor Connect : petit code maison (`IdentityInterceptor`). Pas `connectrpc.com/authn` obligatoire — wrapper 40 lignes.
- Passwords : `core/util/password.go` (bcrypt) déjà là. Signup/login email sans Firebase.
- Plus tard OIDC (Google) : interface `auth.FederatedProvider`, pas Firebase. Un seul provider impl.

gRPC-Web browser : `credentials: 'include'` → cookie. Same-origin `/rpc`.

Worker : pas de user token. Le message queue porte `art_id` / `composition_id`. Le worker s’authentifie en **service**.

BFF HTML : valide la session **lui-même** (cookie). S’il appelle l’API en gRPC, il forward le cookie **ou** il passe `x-identity-user-id` **seulement** si l’API n’écoute pas le monde (réseau docker). Préférer forward cookie : un seul chemin d’auth sur l’API.

### Identité publique dans les resource names

**Un** ID : UUID interne Postgres. Plus de Firebase UID dans le path.

```
users/{uuid}/arts/{uuid}
```

`firebase_uid` colonne droppée après coupure Firebase. Mapping transitoire OK le temps du dual-run.

---

## 6. Queue — agnostic

Interface déjà là (`QueueClient`). Trop liée Pub/Sub.

Cible :

```go
type Queue interface {
    Publish(ctx context.Context, topic string, body []byte) error
    Subscribe(ctx context.Context, topic, consumer string, h Handler) error
    Close() error
}
```

Impls :

1. `postgres` — table `jobs`, `FOR UPDATE SKIP LOCKED`. **Default local + petit prod.** Zero infra.
2. `nats` / `sqs` plus tard derrière la même interface.

Drop Firebase Pub/Sub emulator. Le worker arrête de créer la subscription au boot (side effect).

---

## 7. Couper Firebase

Ordre (ne pas tout couper le jour 0) :

1. Storage → MinIO (Functions storage trigger meurt).
2. `CompleteArtUpload` remplace `onArtImageUpload`.
3. Signup/login cookie remplace Firebase Auth JS + emu.
4. `syncUserOnCreate` meurt : `CreateUser` dans le handler signup.
5. Supprimer `functions/`, `storage.rules`, `storage.production.rules`, terraform module firebase, env `FIREBASE_*`, `PASETO_*`, `TOKEN_SYMMETRIC_KEY`.
6. Pub/Sub → Postgres jobs.

Auth0 comments / `GetUserInfoFromAPI` : tombent avec le ménage `core/auth`.

---

## 8. Frontend

- Pages : HTMX + templ. Pas de rewrite SPA.
- RPC : islands TS (`art-upload.ts`, plus tard poll composition) via `client/src/gen` (buf es).
- `createGrpcWebTransport({ baseUrl: '/rpc', useBinaryFormat: true, credentials: 'include' })`.
- Upload : `StartArtUpload` → PUT presign + headers → `CompleteArtUpload`. **Interdit** : PATCH metadata, `waitForFirebaseReady`, `window.firebase`.
- Proxy : Traefik ou le BFF reverse-proxy `/rpc` → `api:9090`. Same-origin cookies.
- Smart router : freeze. Pas de migration handlers.

---

## 9. Sécurité (revue)

P0 :

- API pas exposée avec CORS `*`.
- `CompleteArtUpload` / `StartArtUpload` : identity == resource user, status machine.
- Presign : key **serveur**, Content-Type dans la signature, TTL court, `Head` au complete (taille, type).
- Cookie : httpOnly, Secure (prod), SameSite, rotation, Redis.
- Timing-safe compare pour service HMAC (`subtle.ConstantTimeCompare`).
- `http.Server` timeouts. Max body sur les RPC (pas sur le PUT S3).
- Plus d’internal API key en Bearer générique sur des paths whitelistés. Service creds dédiés, procedures listées.
- Logs : pas de dump proto, pas de body HTTP error complet, pas d’email en Info.

P1 :

- CSRF : same-site cookie + gRPC-Web depuis JS same-origin. Forms HTMX : token existant à auditer.
- Bucket : gcode/pathlist jamais public-read.
- Resource names : 403 si parent ≠ caller (pas de leak d’existence via des messages différents si possible — NotFound uniforme).
- Delete storage : même schéma de key que l’upload (bug actuel UUID vs Firebase UID).

P2 :

- Rate limit `StartArtUpload`.
- Content-length-range si on passe en POST policy (R2 capricieux → **PUT seulement**).
- mTLS service-to-service en prod plus tard (remplace HMAC).

---

## 10. Catalogue d’interfaces (`core/`)

Tout nouveau comportement passe par une interface. Implémentations dans le même package ou `core/<x>/<provider>`.

| Package | Interface | Impls |
|---|---|---|
| `core/storage` | `Bucket` | `s3` (MinIO/R2/AWS) |
| `core/auth` | `Sessions`, `Passwords`, `Identities` | scs+redis, bcrypt, postgres |
| `core/auth` | `ServiceAuth` | hmac |
| `core/auth` | `FederatedProvider` | none day 1, oidc later |
| `core/queue` | `Queue` | postgres, (nats) |
| `core/resource` | `Parser`, builders | einride (déjà) |
| `core/errors` | unifier sur Connect | drop grpc/status |
| `core/service` | handlers Connect par resource | `ArtServer`, `CompositionServer`, `UserServer` — plus de god `Server` |
| `core/clock` | `Clock` | `Real`, `Fake` (tests) |
| `core/id` | `Generator` | uuid |

`cmd/api`, `cmd/worker`, `client/` : wiring only.

`util.Config` : plus de `*sql.DB` dedans. Config = valeurs. Le main injecte.

---

## 11. Phases & ownership fichiers

Ne **pas** tout faire dans un seul PR. Ownership pour éviter les collisions agents.

### Phase A — Contrats (préalable, séquentiel)

- `docs/rebuild-plan.md` (ce fichier)
- Interfaces `core/storage`, `core/auth`, `core/queue` resserrées
- Proto : `StartArtUpload` / `CompleteArtUpload`, drop storage plat si possible
- buf v2 remote plugins

### Phase B — Fondations parallèles

| Stream | Fichiers autorisés | Interdit |
|---|---|---|
| **B1 proto-gen** | `proto/buf.yaml`, `proto/buf.gen.yaml`, `Makefile` (cible proto), `Tiltfile` (proto-generate), delete `buf.gen.make.yaml` | proto métier, go services |
| **B2 bucket** | `core/storage/**`, `docker-compose.yml` (service minio + env S3_*), tests storage | firebase_*.go consumers ailleurs : adapter l’interface, laisser un stub de compile |
| **B3 identity** | `core/auth/**` (nouveau), `client/internal/auth/session*.go` vers Redis | paseto : marquer deprecated, ne plus appeler depuis api/main |
| **B4 queue-pg** | `core/queue/**`, migration `jobs` | ne pas casser pubsub tant que worker pas basculé |

### Phase C — API runtime

- `cmd/api` : h2c, reflection, identity interceptor, timeouts, **no CORS**
- Services : `*connect.Error`, parent/mask, Complete/Start upload
- Worker : `Bucket` + `Queue` postgres
- Drop `ConnectAdapter`

### Phase D — Browser RPC

- `client/src/gen` + `@connectrpc/connect-web`
- `art-upload.ts` rewrite
- Proxy `/rpc`
- Drop `window.firebase` de l’upload

### Phase E — Couper le GCP

- Functions, firebase emu, paseto, pubsub emu, terraform firebase
- Signup/login password
- Docs README

### Phase F — Qualité

- Tests service art/upload/auth interceptor
- `buf lint` + breaking en CI
- Freeze smart-router ou delete hybrid mort

---

## 12. Critères “terminé” (vision)

- `grpcurl -plaintext localhost:9090 list` marche.
- Browser Network : `application/grpc-web+proto` sur `/rpc`.
- Upload : PUT MinIO, **aucun** hop Firebase, image visible après `CompleteArtUpload`.
- `make proto` no-diff, sans GOPATH plugins.
- `grep -r firebase --include='*.go'` vide (sauf journal/plan).
- `grep -r paseto --include='*.go'` vide.
- Un `.env` : Postgres + MinIO + Redis + keys cookie/HMAC. Plus de `FIREBASE_*`.
- Worker traite une composition E2E local.

---

## Journal

Format : `YYYY-MM-DD HH:MM TZ | type | texte`. Types : `decision` `discovery` `progress` `blocker`. Append-only. Ne pas réécrire l’historique.

### 2026-08-28 ~11:44 EDT | discovery | Review initiale

Contrat AIP (resource names, protovalidate, einride partiel) vs runtime Connect vs BFF HTMX vs annotations `google.api.http` mortes. Triple mental model.

### 2026-08-28 ~11:44 EDT | discovery | Bugs runtime

- `ListUsers`/`DeleteUser` no-op 200 (`connect_adapter.go`).
- `CreateArt`/`ListArts` ignorent `parent`.
- `update_mask` jamais appliqué.
- Double identité Firebase UID (path) vs UUID (DB). `DeleteArt` storage key avec UUID → images probablement pas supprimées.
- BFF `SyncUserFromFirebase` : path whitelist PASETO + adapter exige API key → retry sync mort.
- `ListCompositions` BFF sans `parent` (CEL fail).
- `StorageService` proto non monté ; `GenerateUploadURL` new Firebase client **par request**.
- `user_id` body trust (IDOR si API exposée).
- Deux systèmes d’erreurs (grpc/status vs connect.Error).
- Page tokens = offset clair. `page_size gt:0` contredit AIP-158.

### 2026-08-28 ~11:46 EDT | discovery | Git proto / gRPC browser

`609c85f Working grpc Call` : Next + `createGrpcWebTransport({ useBinaryFormat: true })` + Envoy grpc-web → `grpc.NewServer`. Cible déjà atteinte une fois.

`16cf382 remove envoy` **pendant** que l’API était encore gRPC natif → browser cassé (pas de h2 gRPC).

Puis `createConnectTransport` vers un serveur pas encore Connect. Puis `c56cf61` Next→HTMX. Puis l’API devient Connect **sans** client browser.

Connect-Go parle gRPC-Web binaire **sans Envoy**. Envoy est un fossile.

Proto gen : `buf.gen.yaml` TS → `web/` supprimé ; `make proto` utilise `buf.gen.make.yaml` (Go only) ; plugins `es`/`connect-es` locaux non installés ; `@latest` ; `connect-es` v1 mort (protobuf-es v2 = `bufbuild/es`).

### 2026-08-28 ~11:50 EDT | discovery | Signed URLs / Firebase

Firebase emulator **ne signe pas**. Code : URL REST brute + `art-upload.ts` PUT sans Content-Type + PATCH metadata. Hors contrat S3.

Dual endpoint MinIO (`minio:9000` vs host browser) = `SignatureDoesNotMatch`. `3bd30e7 Improve blob.bucket` avait le bon split ops/sign. Envoy routait `/storage`. Envoy viré → signatures cassées. gocloud n’était pas le problème.

Functions `onFinalize` = 3e hop inutile en S3 (`Head` au complete).

`GetPublicURL` / `pbx` fait de l’I/O. ListArts N+1 storage.

### 2026-08-28 ~11:54 EDT | decision | Auth bundle

**Pas PASETO. Pas Firebase. Pas IdP lourd.**

Bundle : **SCS cookie (Redis) + S3 presign + HMAC service**. Trois jetons, trois jobs. Signed URLs hors auth app.

Raison : same-origin gRPC-Web a besoin de cookies ; grpcurl/JWT écosystème ; presign S3 est déjà le capability-token objet ; PASETO n’apporte rien sur ces deux hops et n’a pas de client browser.

OIDC social = `FederatedProvider` plus tard.

### 2026-08-28 ~11:54 EDT | decision | Storage = S3 only

Driver unique AWS SDK v2. Local MinIO. Prod R2 ou S3. Cloudflare = R2+CDN, **pas** le runtime Go. Postgres jobs default queue.

### 2026-08-28 ~11:54 EDT | decision | Identité resource names

UUID Postgres partout dans les names. Plus de Firebase UID dans le path.

### 2026-08-28 ~11:54 EDT | decision | Phasage agents

Phase A contrats → Phase B 4 streams fichiers disjoints → C API → D browser → E coupe GCP. Smart-router gelé.

### 2026-08-28 ~11:55 EDT | progress | Plan écrit

Ce fichier créé.

### 2026-08-28 ~11:59 EDT | progress | Phase B agents lancés

Streams B1 proto-gen, B2 bucket S3, B3 identity/HMAC, B4 queue postgres. `go.mod` déjà pourvu d’aws-sdk-go-v2. Agents interdits de `go get` concurrent.

### 2026-08-28 ~11:58 EDT | progress | Phase A contrats

Interfaces posées :
- `core/storage/bucket.go` — `Bucket`, `PresignPut`
- `core/auth/identity.go` `sessions.go` `service_cred.go`
- `core/queue/topics.go` — `Queue` (postgres default visé)
- `core/clock`, `core/id`

`StorageProvider` (Firebase) encore là pour compiler. Dual-run jusqu’à Phase E.

### 2026-08-28 | progress | B2 bucket S3

Implemented `storage.NewS3Bucket` / `NewBucket` (AWS SDK v2). Dual clients: **ops** = `S3_ENDPOINT` (`http://minio:9000`), **sign** = host of `S3_PUBLIC_BASE_URL` (`localhost:9000`). `GetPublicURL` is not on the driver; legacy `StorageProvider` adapter concatenates PublicBaseURL for dual-run until phase C.

Compose: services `minio` (API 9000, console 9001) + `minio-init` (`mc mb local/thread-art`). api+worker get `STORAGE_PROVIDER=s3` and `S3_*`. FIREBASE_* left in place.

Discovery:
- MinIO public object URL is path-style `http://localhost:9000/thread-art/{key}`. Signer `BaseEndpoint` must be `http://localhost:9000` — if you pass `http://localhost:9000/thread-art` the SDK doubles the bucket (`/thread-art/thread-art/key`).
- SDK v2 `PresignPutObject` strips `Content-Type` when body/Content-Length is 0 (`RemoveContentTypeHeader`). Restored in Finalize before Signing so the browser MUST replay it. Checksums set to `WhenRequired` so `x-amz-checksum-*` is not signed (browsers won't send it).
- `S3_FORCE_PATH_STYLE=true` required for MinIO (and typically R2). AWS native: false.
- cmd/api still calls `NewStorageProvider` (adapter), not `Bucket`. Phase C wires the interface.
- `go.mod` did **not** already have aws-sdk-go-v2. Bare `go mod tidy` resolved s3 **v1.27 / sdk v1.16** (no `BaseEndpoint`). Pinned `aws-sdk-go-v2 v1.45.1` + `service/s3 v1.109.1` + `credentials v1.20.1`. No `go get` of other modules.

### 2026-08-28 ~12:10 EDT | discovery | B3 identity — sessions

Wrapping `*scs.SessionManager` in `core/auth` is cycle-free (scs already in go.mod; client already imports core). Plan claimed `scs/redisstore` is a dep — it is **not** in go.mod (only `postgresstore`). No `go get` this round.

`SCSSessions` adapter exists (mem default via `scs.New()`). Redis/postgres store wiring deferred to cmd/client (phase C). Cookie name follows the injected manager (client today uses `session_id`, SCS default is `session`).

### 2026-08-28 ~12:10 EDT | discovery | B3 identity — Bearer/PASETO

`IdentityInterceptor` ignores Bearer (does not validate PASETO). PASETO still lives in `PasetoAuthMiddleware`. When stacked, IdentityInterceptor **must run first**: PASETO rejects non-Bearer, so Service HMAC would 401 if PASETO ran first. Phase C drops PASETO; this interceptor is the only gate. SyncUserFromFirebase / ConfirmArtImageUploadFromFunction stay on the old whitelist until then.

### 2026-08-28 ~12:10 EDT | progress | B3 identity

HMACServiceAuth (crypto/hmac+sha256, subtle.ConstantTimeCompare), BcryptPasswords (wraps core/util), SCSSessions adapter, IdentityInterceptor (Service HMAC → cookie → Bearer passthrough; Health skip; sets `auth.Identity` + `middleware.AuthKey`). Not wired in cmd/api this round. paseto_service.go marked Deprecated, still compiles.

### 2026-08-28 ~12:05 EDT | progress | Dependabot security bumps

Go (go.mod toolchain → 1.25.0, machine locale 1.25.2) :
- golang.org/x/crypto 0.52.0 (critical SSH/agent)
- google.golang.org/grpc 1.82.1
- golang.org/x/net 0.55.0, x/image 0.41.0
- chi v5.2.4, go-jose/v4 4.1.4, cel-go 0.29.0, otel sdk 1.43.0
- aws-sdk-go-v2 s3 1.97.3 (aligné Dependabot + B2)

npm client : 0 vuln (`overrides` + postcss 8.5.23, webpack 5.104.1).

npm functions : high/critical down. Restent 8 **moderate** tous dans l’arbre `firebase-admin@12` — fix officiel = v14 (major). Won’t-fix : Functions meurent en Phase E.

Won’t-fix Go :
- `github.com/golang-jwt/jwt` v3 (unmaintained, no patch). Dead code `core/token/jwt_maker.go`. Part avec PASETO.
- `github.com/disintegration/imaging` (no patch). `x/image` bumpé en dessous.

Alertes GitHub ne se ferment qu’après **push**. PRs Dependabot ouverts (viper, grpc-gateway, gocloud, protovalidate) = noise / libs qu’on drop. Ne pas merger gocloud.

### 2026-08-28 ~12:16 EDT | progress | B3 identity tests

`go test ./core/auth/ ./core/interceptors/` pass. `go build ./core/auth/ ./core/interceptors/` pass (paseto_service.go still compiles). `go build ./cmd/api` currently fails on B1 proto import paths (`core/pb/buf/validate`) — not this stream.

### 2026-08-28 ~12:12 EDT | progress | B4 queue-pg

`PostgresQueue` implements `queue.Queue` with `jobs` + `FOR UPDATE SKIP LOCKED`. Dual-run: `pubsub.go` untouched, worker stays on Pub/Sub. Live tests: `DATABASE_URL=... go test -tags integration ./core/queue`.

Migration: `000013_jobs.up.sql` / `000013_jobs.down.sql`.

### 2026-08-28 | progress | B1 proto-gen

`make proto` = `buf` on PATH + `buf generate && buf lint`. Remote pins: protocolbuffers/go v1.36.6, connectrpc/go v1.17.0 (BSR has it, matches go.mod), bufbuild/es v2.10.0. OpenAPI plugin dropped. grpc-gateway stays as a **dep** (not a plugin) because existing `.proto` files import `protoc-gen-openapiv2` annotations — cannot strip without a proto edit.

lint excepts (contract/layout, not this stream): `PACKAGE_DIRECTORY_MATCH`, `PACKAGE_VERSION_SUFFIX`, `IMPORT_USED`, `RPC_REQUEST_RESPONSE_UNIQUE`, `RPC_RESPONSE_STANDARD_NAME`. Managed mode disabled on googleapis/protovalidate/grpc-gateway so Go imports stay on genproto + protovalidate SDK. es plugin `include_imports: true` so TS file descriptors resolve.

### 2026-08-28 ~12:05 EDT | progress | Dependabot security merged

PR https://github.com/Damione1/thread-art-generator/pull/93 squash-merged to `master` as `c3a5547`. Branch `fix/dependabot-security` deleted. Worktree `/tmp/tag-security` removed. Feature WIP was not included.

### 2026-08-28 ~12:30 EDT | progress | Phase C kickoff

`StartArtUpload` / `CompleteArtUpload` added to proto + generated. `page_size` CEL `gte: 0`. Object key helpers in `core/resource`. `SERVICE_HMAC_SECRET` / `QUEUE_PROVIDER` (default postgres). aws-sdk-go-v2 + grpcreflect/grpchealth in go.mod. Adapter stubs so `cmd/api` compiles. MinIO already in compose.

Dual-run identity: session `user_id` is still **Firebase UID**. Resource names stay UID this wave. Object keys use Postgres UUID (`ArtOriginalObjectKey`). Phase E flips names.

ConnectAdapter kept this wave (drop later). Smart-router frozen.

### 2026-08-28 ~12:20 EDT | progress | Phase C API runtime (`cmd/api`)

h2c via `http.Protocols` + `ReadHeaderTimeout`/`IdleTimeout`. Dropped CORS `*` (same-origin `/rpc`). IdentityInterceptor (SCS cookie `session_id` + postgresstore + HMAC) stacked before PASETO. grpcreflect v1/v1alpha + grpchealth on `pb.ArtGeneratorService`. `/health` JSON kept. ConnectAdapter + FirebaseFunctions still mounted. Session `user_id` still Firebase UID.

### 2026-08-28 ~12:09 EDT | progress | C worker postgres-queue

Worker default path is `queue.PostgresQueue` on `TopicCompositionProcessing` (`composition-processing`). `QUEUE_PROVIDER=pubsub` still boots the old emulator subscription. Object keys now use Postgres UUID + helpers (`ArtOriginalObjectKey` `/original`, composition preview/gcode/pathlist). Firebase UID no longer required for storage paths. Health server got `ReadHeaderTimeout`. No new migration.

### 2026-08-28 ~12:13 EDT | progress | Phase D browser RPC

Same-origin `/rpc` reverse-proxy on the BFF (`hybrid_setup.go` → `config.ApiURL`, default `http://api:9090`). Host rewrite + `FlushInterval=-1`. No CORS. SecurityPublic; cookie forwarded, IdentityInterceptor is the gate. `/api/storage/upload-url` kept (dual-run).

`art-upload.ts` → `createClient` + `createGrpcWebTransport({ baseUrl: '/rpc', useBinaryFormat: true })`. connect-web v2 dropped `credentials`; cookies via `fetch` override. Flow: `startArtUpload` → S3 PUT replaying response `headers` (no PATCH) → `completeArtUpload`. No Firebase in the upload path. Resource name still `users/{firebaseUid}/arts/{artId}` this wave.

npm: `@bufbuild/protobuf` `@connectrpc/connect` `@connectrpc/connect-web` (v2). `tsc --noEmit`, webpack, `go build ./client/cmd/frontend` pass.

### 2026-08-28 ~12:40 EDT | progress | feature/rebuild from master

New integration branch `feature/rebuild` off `master` (`c3a5547`). Smart-router stays on `feature/routing-improvment`. Stacked PRs target `feature/rebuild`, not master.

Ported onto master DualBucket: AWS SDK v2 `BlobStorage` wrapper (dropped gocloud at runtime), `StartArtUpload`/`CompleteArtUpload` (headers + Head), IdentityInterceptor before Firebase AuthMiddleware, h2c + reflect/health, PostgresQueue default, worker UUID object keys, BFF `/rpc` + `art-upload.ts`. Dual-run: `GetArtUploadUrl`/`ConfirmArtImageUpload` + RabbitMQ if `QUEUE_PROVIDER=rabbitmq`. Session `user_id` still Firebase UID; resource names already Postgres UUID.

Stacked PRs merged into `feature/rebuild` (not master):
- #94 S3 + proto-gen
- #95 HMAC + IdentityInterceptor + Postgres queue
- #96 h2c API + Postgres worker
- #97 gRPC-Web upload island + `/rpc`

### 2026-08-28 ~13:10 EDT | progress | CompleteArtUpload size cap

`Head.Size` rejected above 10MB (same as the browser island). gocloud / aws-sdk-go v1 / rs/cors already gone from go.mod after tidy.

Remaining for §12: Firebase Auth JS + emu + `FIREBASE_*` (Phase E), RabbitMQ dual-run, DualBucket → one bucket, session `user_id` still Firebase UID.

