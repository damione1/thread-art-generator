package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Damione1/thread-art-generator/cmd/cli/internal/config"
	"github.com/Damione1/thread-art-generator/core/pb"
)

// Service handles gRPC client operations
type Service struct {
	ConfigManager *config.Manager
	ServerAddress string
}

// NewService creates a new client service
func NewService(configManager *config.Manager) *Service {
	return &Service{
		ConfigManager: configManager,
		ServerAddress: "tag.local:9090", // Could make this configurable
	}
}

// GetClient creates a new gRPC client with authentication
func (s *Service) GetClient() (pb.ArtGeneratorServiceClient, error) {
	// Check if token is valid
	if !s.ConfigManager.IsTokenValid() {
		return nil, fmt.Errorf("not authenticated or token expired")
	}

	// Create a connection with the gRPC server
	conn, err := grpc.Dial(s.ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to gRPC server: %v", err)
	}

	// Create and return client
	return pb.NewArtGeneratorServiceClient(conn), nil
}

// GetAuthContext creates a context with auth metadata
func (s *Service) GetAuthContext() (context.Context, error) {
	if s.ConfigManager.Config.AccessToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	ctx := context.Background()
	md := metadata.Pairs("authorization", "Bearer "+s.ConfigManager.Config.AccessToken)
	return metadata.NewOutgoingContext(ctx, md), nil
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
