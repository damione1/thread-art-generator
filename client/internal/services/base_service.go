package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/transport"
	coreErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/rs/zerolog/log"
)

// BaseService provides shared functionality for all domain services
type BaseService struct {
	client         pbconnect.ArtGeneratorServiceClient
	sessionManager *auth.SCSSessionManager
}

// NewBaseService creates a new base service with shared dependencies
func NewBaseService(client pbconnect.ArtGeneratorServiceClient, sessionManager *auth.SCSSessionManager) *BaseService {
	return &BaseService{
		client:         client,
		sessionManager: sessionManager,
	}
}

// parseErrorToFieldErrors converts Connect errors to form field errors
func (s *BaseService) parseErrorToFieldErrors(err error) map[string][]string {
	parser := coreErrors.NewErrorParser()
	standardErr := parser.ParseConnectError(err)

	fieldErrors := make(map[string][]string)

	// Convert field-level errors to form format
	for field, messages := range standardErr.Fields {
		fieldErrors[field] = messages
	}

	// Add global error if present
	if standardErr.GlobalError != "" {
		fieldErrors["_form"] = []string{standardErr.GlobalError}
	}

	// If no specific errors were parsed, use the raw error message
	if len(fieldErrors) == 0 {
		fieldErrors["_form"] = []string{standardErr.Message}
	}

	return fieldErrors
}

// parseErrorForLogging converts Connect errors to structured error for logging
func (s *BaseService) parseErrorForLogging(err error) *coreErrors.StandardError {
	parser := coreErrors.NewErrorParser()
	return parser.ParseConnectError(err)
}

// isUserNotFoundError checks if the error indicates the user was not found in the database
func (s *BaseService) isUserNotFoundError(err error) bool {
	standardErr := s.parseErrorForLogging(err)
	return standardErr.Type == coreErrors.ErrorTypeNotFound ||
		(standardErr.Type == coreErrors.ErrorTypeInternal && 
		(strings.Contains(standardErr.Message, "user not found") || 
		 strings.Contains(standardErr.Message, "failed to get user")))
}

// syncUserFromFirebaseWithRetry attempts to sync the current user from Firebase authentication
// This is a fallback mechanism when the Firebase Cloud Function fails to sync automatically
func (s *BaseService) syncUserFromFirebaseWithRetry(ctx context.Context, r *http.Request, maxRetries int) error {
	if r == nil {
		return fmt.Errorf("HTTP request required for user sync")
	}

	// Get session data to extract Firebase user information
	sessionData, err := s.sessionManager.GetSession(r)
	if err != nil {
		log.Error().Err(err).Msg("syncUserFromFirebaseWithRetry: Failed to get session data")
		return fmt.Errorf("failed to get session data: %w", err)
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

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Info().
			Str("firebase_uid", sessionData.UserID).
			Str("email", sessionData.UserInfo.Email).
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Msg("syncUserFromFirebaseWithRetry: Attempting user sync")

		// Call the sync API
		resp, err := s.client.SyncUserFromFirebase(ctxWithSession, syncReq)
		if err != nil {
			lastErr = err
			standardErr := s.parseErrorForLogging(err)
			
			log.Warn().
				Err(err).
				Str("firebase_uid", sessionData.UserID).
				Str("error_type", string(standardErr.Type)).
				Str("error_message", standardErr.Message).
				Int("attempt", attempt).
				Msg("syncUserFromFirebaseWithRetry: Sync attempt failed")
			
			// If it's a validation error with message about existing user, that's actually success
			if standardErr.Type == coreErrors.ErrorTypeValidation && 
			   strings.Contains(standardErr.Message, "user with this Firebase UID already exists") {
				log.Info().
					Str("firebase_uid", sessionData.UserID).
					Msg("syncUserFromFirebaseWithRetry: User already exists, sync successful")
				return nil
			}
			
			// If this is the last attempt, return the error
			if attempt == maxRetries {
				break
			}
			
			// Wait before retrying (exponential backoff could be added here)
			continue
		}

		// Sync successful
		log.Info().
			Str("firebase_uid", sessionData.UserID).
			Str("user_resource_name", resp.Msg.GetName()).
			Str("email", resp.Msg.GetEmail()).
			Int("attempt", attempt).
			Msg("syncUserFromFirebaseWithRetry: User sync successful")
		
		return nil
	}

	// All retries failed
	standardErr := s.parseErrorForLogging(lastErr)
	log.Error().
		Err(lastErr).
		Str("firebase_uid", sessionData.UserID).
		Str("error_type", string(standardErr.Type)).
		Str("error_message", standardErr.Message).
		Int("max_retries", maxRetries).
		Msg("syncUserFromFirebaseWithRetry: All sync attempts failed")
	
	return fmt.Errorf("failed to sync user after %d attempts: %s", maxRetries, standardErr.Message)
}

// tryWithUserSync wraps an API call and automatically retries with user sync if user not found
func (s *BaseService) tryWithUserSync(ctx context.Context, r *http.Request, operation func() error) error {
	// First attempt
	err := operation()
	if err == nil {
		return nil // Success
	}

	// Check if it's a user not found error
	if !s.isUserNotFoundError(err) {
		return err // Different error, don't retry
	}

	log.Info().
		Err(err).
		Msg("tryWithUserSync: User not found, attempting automatic sync")

	// Try to sync the user
	syncErr := s.syncUserFromFirebaseWithRetry(ctx, r, 2) // Max 2 retry attempts
	if syncErr != nil {
		log.Error().
			Err(syncErr).
			Msg("tryWithUserSync: Failed to sync user, returning original error")
		return err // Return original error if sync fails
	}

	// Retry the original operation after successful sync
	log.Info().Msg("tryWithUserSync: User sync successful, retrying original operation")
	retryErr := operation()
	if retryErr != nil {
		log.Error().
			Err(retryErr).
			Msg("tryWithUserSync: Operation still failed after user sync")
	}
	
	return retryErr
}