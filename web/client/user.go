package client

import (
	"context"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
)

// GetUser gets a user by name
func (c *GrpcClient) GetUser(ctx context.Context, name string) (*pb.User, error) {
	ctx, cancel := WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.GetUserRequest{
		Name: name,
	}

	return c.client.GetUser(ctx, req)
}

// GetCurrentUser gets the current user from the context
func (c *GrpcClient) GetCurrentUser(ctx context.Context, token string) (*pb.User, error) {
	ctx = WithAuth(ctx, token)
	ctx, cancel := WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// The backend should extract the user from the token
	// and return the current user
	req := &pb.GetUserRequest{
		Name: "me",
	}

	return c.client.GetUser(ctx, req)
}
