package client

import (
	"context"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
)

// CreateSession creates a new session (login)
func (c *GrpcClient) CreateSession(ctx context.Context, email, password string) (*pb.CreateSessionResponse, error) {
	ctx, cancel := WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.CreateSessionRequest{
		Email:    email,
		Password: password,
	}

	return c.client.CreateSession(ctx, req)
}

// RefreshToken refreshes an access token
func (c *GrpcClient) RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error) {
	ctx, cancel := WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	return c.client.RefreshToken(ctx, req)
}

// DeleteSession deletes a session (logout)
func (c *GrpcClient) DeleteSession(ctx context.Context, refreshToken string) error {
	ctx, cancel := WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.DeleteSessionRequest{
		RefreshToken: refreshToken,
	}

	_, err := c.client.DeleteSession(ctx, req)
	return err
}
