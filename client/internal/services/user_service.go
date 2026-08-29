package services

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/auth"
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

// GetCurrentUser gets the current user via the API (cookie forwarded by SessionTransport).
func (s *UserService) GetCurrentUser(ctx context.Context, _ *http.Request) (*auth.UserInfo, error) {
	req := connect.NewRequest(&pb.GetCurrentUserRequest{})

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

	return &auth.UserInfo{
		ID:        resp.Msg.GetName(),
		FirstName: resp.Msg.GetFirstName(),
		LastName:  resp.Msg.GetLastName(),
		Email:     resp.Msg.GetEmail(),
		Picture:   resp.Msg.GetAvatar(),
	}, nil
}
