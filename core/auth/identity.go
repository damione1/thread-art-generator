package auth

import "context"

// Identity is the authenticated principal on a request.
// UserID is the public resource id (Postgres UUID).
type Identity struct {
	UserID         string
	Email          string
	FirstName      string
	LastName       string
	Active         bool
	Kind           PrincipalKind
	SessionVersion int
}

// PrincipalKind distinguishes browser sessions from internal workers.
type PrincipalKind int

const (
	PrincipalUser PrincipalKind = iota + 1
	PrincipalService
)

type identityContextKey struct{}

// WithIdentity stores identity on the context (Connect interceptors).
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, id)
}

// IdentityFromContext returns the request identity if present.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityContextKey{}).(Identity)
	return id, ok
}
