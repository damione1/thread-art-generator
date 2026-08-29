# Thread Art Generator

![Thread Art Generator](https://github.com/Damione1/thread-art-generator/assets/14912510/6b6ef9e1-9bad-4dd7-8579-17fe55ae9c13)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Turn images into circular thread art: upload, tweak settings, compare compositions, export GCode.

## Stack

```
[Web UI] <--> [API] <--> [Postgres jobs] <--> [Worker]
                |                                  |
                v                                  v
           [Postgres] <---------------------> [S3 / MinIO]
```

- **UI** — Go + Templ + HTMX. Email/password, session cookie.
- **API** — Connect-RPC (h2c). Browser talks protobuf on same-origin `/rpc`.
- **Worker** — consumes Postgres `SKIP LOCKED` jobs.
- **Storage** — one S3-compatible bucket (MinIO locally).

## Local

Needs Docker, Tilt, Go 1.25+, Node, [Buf](https://buf.build/docs/installation).

```bash
make setup   # .env from sample, templ, npm, proto tools
make up      # tilt up
```

- UI: http://localhost:8080
- API: http://localhost:9090 (`/health`)
- MinIO console: http://localhost:9001

```bash
make down
make proto      # buf generate + lint
make test
make psql
```

Auth is `/auth/login` and `/auth/signup`. The worker uses `Authorization: Service …` (HMAC). Object env is `S3_*` in `.env`.

## Hardware

Machine designs come from [Bdring's StringArt](https://github.com/bdring/StringArt). FluidNC configs live under `machine/`.

## License

MIT — see [LICENSE](LICENSE).
