package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/Damione1/thread-art-generator/cmd/cli/internal/config"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
)

// Service handles Connect-RPC client operations
type Service struct {
	ConfigManager *config.Manager
	ServerAddress string
}

// NewService creates a new client service
func NewService(configManager *config.Manager) *Service {
	return &Service{
		ConfigManager: configManager,
		ServerAddress: "http://tag.local:9090", // Now requires http:// prefix
	}
}

// GetClient creates a new Connect-RPC client with authentication
func (s *Service) GetClient() (pbconnect.ArtGeneratorServiceClient, error) {
	// Check if token is valid
	if !s.ConfigManager.IsTokenValid() {
		return nil, fmt.Errorf("not authenticated or token expired")
	}

	// Create an HTTP client with auth interceptor
	httpClient := &http.Client{
		Transport: &authTransport{
			token: s.ConfigManager.Config.AccessToken,
			base:  http.DefaultTransport,
		},
	}

	// Create and return Connect client with gRPC protocol (full binary)
	// gRPC mode gives us better performance for CLI tools
	return pbconnect.NewArtGeneratorServiceClient(
		httpClient,
		s.ServerAddress,
		connect.WithGRPC(),                  // Use gRPC protocol for full binary mode
		connect.WithSendCompression("gzip"), // Use gzip compression for sending
	), nil
}

// authTransport is an http.RoundTripper that adds auth headers
type authTransport struct {
	token string
	base  http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

// GetAuthContext creates a context with auth metadata - no longer needed for Connect
// but kept for backwards compatibility
func (s *Service) GetAuthContext() (context.Context, error) {
	if s.ConfigManager.Config.AccessToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	// Connect client doesn't use context for auth, but we'll keep this method
	// to maintain backward compatibility with existing code
	return context.Background(), nil
}

// GetAuthContextWithTimeout creates a context with auth metadata and timeout
func (s *Service) GetAuthContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc, error) {
	ctx, err := s.GetAuthContext()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, cancel, nil
}
