package client

import (
	"context"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

// Factory creates API clients
// Implements auth.ClientFactory interface
type Factory struct {
	baseURL        string
	sessionManager *auth.SessionManager
}

// NewFactory creates a new client factory
func NewFactory(baseURL string, sessionManager *auth.SessionManager) *Factory {
	return &Factory{
		baseURL:        baseURL,
		sessionManager: sessionManager,
	}
}

// NewGRPCClient creates a new gRPC client
func (f *Factory) NewGRPCClient() auth.APIClient {
	return NewGRPCClient(f.baseURL, f.sessionManager)
}

// AddTokenToContext adds an access token to the context
func (f *Factory) AddTokenToContext(ctx context.Context, token string) context.Context {
	// Create a mock session to pass in the context
	sessionData := &auth.SessionData{
		AccessToken: token,
	}
	return context.WithValue(ctx, "session", sessionData)
}
