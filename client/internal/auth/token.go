package auth

import (
	"context"
)

// Key for the request context
type tokenContextKey struct{}

// WithToken adds a token to the context
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

// TokenFromContext extracts the token from context
func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey{}).(string)
	return token, ok
}
