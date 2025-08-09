package middleware

import (
	"context"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

// User is a key for the request context
type userContextKey struct{}

// UserFromContext extracts the user info from context
func UserFromContext(ctx context.Context) (*auth.UserInfo, bool) {
	user, ok := ctx.Value(userContextKey{}).(*auth.UserInfo)
	return user, ok
}

// WithUser adds user info to the context
func WithUser(ctx context.Context, user *auth.UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}
