package auth

import (
	"context"
	"net/http"
)

// ClientFactory is an interface for creating API clients
type ClientFactory interface {
	// NewGRPCClient creates a new gRPC client
	NewGRPCClient() APIClient

	// AddTokenToContext adds an access token to the context
	AddTokenToContext(ctx context.Context, token string) context.Context
}

// APIClient is an interface for API operations
type APIClient interface {
	// GetCurrentUser fetches the current user from the API
	GetCurrentUser(ctx context.Context, req *http.Request) (*APIUser, error)

	// CheckSessionToken checks if a session has a valid token
	CheckSessionToken(req *http.Request) error
}

// APIUser represents user data returned from the API
type APIUser struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Avatar    string
}
