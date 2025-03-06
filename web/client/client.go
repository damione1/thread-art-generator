package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GrpcClient wraps the gRPC connection and client
type GrpcClient struct {
	conn   *grpc.ClientConn
	client pb.ArtGeneratorServiceClient
	config util.Config
}

// NewGrpcClient creates a new gRPC client
func NewGrpcClient(config util.Config) (*GrpcClient, error) {
	// Set up a connection to the server
	addr := fmt.Sprintf("api:%s", config.GRPCServerPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	fmt.Println("Connected to gRPC server at", addr)

	// Create a client
	client := pb.NewArtGeneratorServiceClient(conn)

	return &GrpcClient{
		conn:   conn,
		client: client,
		config: config,
	}, nil
}

// Close closes the gRPC connection
func (c *GrpcClient) Close() error {
	return c.conn.Close()
}

// GetClient returns the gRPC client
func (c *GrpcClient) GetClient() pb.ArtGeneratorServiceClient {
	return c.client
}

// WithAuth adds authentication metadata to the context
func WithAuth(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("Bearer %s", token))
}

// WithTimeout adds a timeout to the context
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
