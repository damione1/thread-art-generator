package services

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/client"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/rs/zerolog/log"
)

// UserService handles user-related operations
type UserService struct {
	*BaseService
}

// NewUserService creates a new user service
func NewUserService(baseService *BaseService) *UserService {
	return &UserService{
		BaseService: baseService,
	}
}

// GetCurrentUser gets the current user
func (s *UserService) GetCurrentUser(ctx context.Context, r *http.Request) (*client.User, error) {
	// Use session from request if available
	if r != nil {
		// Extract session from request and add to context
		sessionData, err := s.sessionManager.GetSession(r)
		if err == nil && sessionData.AccessToken != "" {
			// Add session to context
			ctx = context.WithValue(ctx, "session", sessionData)
		}
	}

	// Create the Connect request
	req := connect.NewRequest(&pb.GetCurrentUserRequest{})

	// Make the API call directly using s.client
	resp, err := s.client.GetCurrentUser(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to get current user")
		return nil, fmt.Errorf("failed to get current user: %s", standardErr.Message)
	}

	// Convert the response to our User type
	return &client.User{
		ID:        resp.Msg.GetName(),
		FirstName: resp.Msg.GetFirstName(),
		LastName:  resp.Msg.GetLastName(),
		Email:     resp.Msg.GetEmail(),
		Avatar:    resp.Msg.GetAvatar(),
	}, nil
}
