package services

import (
	"context"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/client"
)

// UserService handles user-related operations
type UserService struct {
	clientFactory auth.ClientFactory
}

// NewUserService creates a new user service
func NewUserService(clientFactory auth.ClientFactory) *UserService {
	return &UserService{
		clientFactory: clientFactory,
	}
}

// GetCurrentUser gets the current user
func (s *UserService) GetCurrentUser(ctx context.Context, r *http.Request) (*client.User, error) {
	// Get GRPC client
	grpcClient := s.clientFactory.NewGRPCClient()

	// Get APIUser from the client
	apiUser, err := grpcClient.GetCurrentUser(ctx, r)
	if err != nil {
		return nil, err
	}

	// Convert to internal User type
	return &client.User{
		ID:        apiUser.ID,
		FirstName: apiUser.FirstName,
		LastName:  apiUser.LastName,
		Email:     apiUser.Email,
		Avatar:    apiUser.Avatar,
	}, nil
}

// UserProfile represents user information within the application
type UserProfile struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Avatar    string
}
