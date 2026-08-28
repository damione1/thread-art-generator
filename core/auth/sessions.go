package auth

import (
	"context"
	"net/http"
	"time"
)

// Session is the browser identity grant. Transported as an httpOnly cookie.
// Orthogonal to S3 presigns and to service HMAC.
type Session struct {
	UserID    string
	Email     string
	ExpiresAt time.Time
}

// Sessions issues and loads sessions. Implementation: SCS + Redis.
type Sessions interface {
	Issue(ctx context.Context, w http.ResponseWriter, r *http.Request, s Session) error
	Load(ctx context.Context, r *http.Request) (Session, error)
	Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	// LoadFromCookie is for gRPC-Web on the API (no BFF): cookie on the RPC request.
	LoadFromCookie(ctx context.Context, r *http.Request) (Session, error)
}

// Passwords hashes and checks credentials. Implementation: bcrypt (core/util).
type Passwords interface {
	Hash(password string) (string, error)
	Check(password, hash string) error
}

// Identities looks up users for login/signup. Implementation: Postgres.
type Identities interface {
	ByEmail(ctx context.Context, email string) (Identity, string, error) // identity + password hash
	Create(ctx context.Context, email, passwordHash, first, last string) (Identity, error)
}
