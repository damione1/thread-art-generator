package services

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/transport"
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

// GetCurrentUser gets the current user with automatic fallback sync
func (s *UserService) GetCurrentUser(ctx context.Context, r *http.Request) (*User, error) {
	var resultUser *User

	// Define the operation that we want to retry with user sync
	operation := func() error {
		// Add the session request to context so the transport layer can access it
		ctxWithSession := ctx
		if r != nil {
			ctxWithSession = transport.WithSessionRequest(ctx, r)
		}

		// Create the Connect request
		req := connect.NewRequest(&pb.GetCurrentUserRequest{})

		// Make the API call - the transport layer will automatically add PASETO token
		resp, err := s.client.GetCurrentUser(ctxWithSession, req)
		if err != nil {
			return err
		}

		// Convert the response to our User type
		resultUser = &User{
			ID:        resp.Msg.GetName(),
			FirstName: resp.Msg.GetFirstName(),
			LastName:  resp.Msg.GetLastName(),
			Email:     resp.Msg.GetEmail(),
			Avatar:    resp.Msg.GetAvatar(),
		}
		return nil
	}

	// Use the base service's tryWithUserSync wrapper
	err := s.tryWithUserSync(ctx, r, operation)
	if err != nil {
		// Log the final error with context
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("GetCurrentUser: Failed to get current user after sync attempts")
		
		return nil, fmt.Errorf("failed to get current user: %s", standardErr.Message)
	}

	// Success - log the result
	log.Debug().
		Str("user_id", resultUser.ID).
		Str("email", resultUser.Email).
		Msg("GetCurrentUser: Successfully retrieved current user")

	return resultUser, nil
}

// SyncUserFromFirebase directly calls the sync API (for manual sync operations)
func (s *UserService) SyncUserFromFirebase(ctx context.Context, r *http.Request) (*User, error) {
	// Get session data to extract Firebase user information
	sessionData, err := s.sessionManager.GetSession(r)
	if err != nil {
		log.Error().Err(err).Msg("SyncUserFromFirebase: Failed to get session data")
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	// Prepare sync request with available user data
	syncReq := connect.NewRequest(&pb.SyncUserFromFirebaseRequest{
		FirebaseUid: sessionData.UserID, // Firebase UID from session
		Email:       sessionData.UserInfo.Email,
		DisplayName: sessionData.UserInfo.Name,
		PhotoUrl:    sessionData.UserInfo.Picture,
	})

	// Add the session request to context for authentication
	ctxWithSession := transport.WithSessionRequest(ctx, r)

	log.Info().
		Str("firebase_uid", sessionData.UserID).
		Str("email", sessionData.UserInfo.Email).
		Msg("SyncUserFromFirebase: Manually syncing user from Firebase")

	// Call the sync API
	resp, err := s.client.SyncUserFromFirebase(ctxWithSession, syncReq)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		
		log.Error().
			Err(err).
			Str("firebase_uid", sessionData.UserID).
			Str("error_type", string(standardErr.Type)).
			Str("error_message", standardErr.Message).
			Msg("SyncUserFromFirebase: Manual sync failed")
		
		return nil, fmt.Errorf("failed to sync user: %s", standardErr.Message)
	}

	// Sync successful - convert to client user
	user := &User{
		ID:        resp.Msg.GetName(),
		FirstName: resp.Msg.GetFirstName(),
		LastName:  resp.Msg.GetLastName(),
		Email:     resp.Msg.GetEmail(),
		Avatar:    resp.Msg.GetAvatar(),
	}

	log.Info().
		Str("firebase_uid", sessionData.UserID).
		Str("user_resource_name", resp.Msg.GetName()).
		Str("email", resp.Msg.GetEmail()).
		Msg("SyncUserFromFirebase: Manual sync successful")
	
	return user, nil
}