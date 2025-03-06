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
