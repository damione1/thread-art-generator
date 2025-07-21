package services

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/client"
	coreErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/rs/zerolog/log"
)

// GeneratorService handles all service interactions with the API
type GeneratorService struct {
	client         pbconnect.ArtGeneratorServiceClient
	sessionManager *auth.SCSSessionManager
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(client pbconnect.ArtGeneratorServiceClient, sessionManager *auth.SCSSessionManager) *GeneratorService {
	return &GeneratorService{
		client:         client,
		sessionManager: sessionManager,
	}
}

// parseErrorToFieldErrors converts Connect errors to form field errors
func (s *GeneratorService) parseErrorToFieldErrors(err error) map[string][]string {
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
func (s *GeneratorService) parseErrorForLogging(err error) *coreErrors.StandardError {
	parser := coreErrors.NewErrorParser()
	return parser.ParseConnectError(err)
}

// ArtFormData represents the form data for creating art
type ArtFormData struct {
	Title   string
	Errors  map[string][]string
	Success bool
}

// UserProfile represents user information within the application
type UserProfile struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Avatar    string
}

// GetCurrentUser gets the current user
func (s *GeneratorService) GetCurrentUser(ctx context.Context, r *http.Request) (*client.User, error) {
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

// CreateArt creates a new art resource
func (s *GeneratorService) CreateArt(ctx context.Context, createArtRequest *pb.CreateArtRequest) (*pb.Art, map[string][]string, error) {
	req := connect.NewRequest(createArtRequest)

	resp, err := s.client.CreateArt(ctx, req)
	if err != nil {
		fieldErrors := s.parseErrorToFieldErrors(err)
		return nil, fieldErrors, err
	}

	return resp.Msg, nil, nil
}

// GetArt gets a specific art by its resource name
func (s *GeneratorService) GetArt(ctx context.Context, userID, artID string) (*pb.Art, error) {
	artName := resource.BuildArtResourceName(userID, artID)

	req := connect.NewRequest(&pb.GetArtRequest{
		Name: artName,
	})

	resp, err := s.client.GetArt(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("art_name", artName).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to get art")
		return nil, fmt.Errorf("failed to get art: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// GetArtUploadUrl gets a signed URL for uploading an image to an art
func (s *GeneratorService) GetArtUploadUrl(ctx context.Context, userID, artID string) (*pb.GetArtUploadUrlResponse, error) {
	artName := resource.BuildArtResourceName(userID, artID)

	req := connect.NewRequest(&pb.GetArtUploadUrlRequest{
		Name: artName,
	})

	resp, err := s.client.GetArtUploadUrl(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("art_name", artName).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to get art upload URL")
		return nil, fmt.Errorf("failed to get art upload URL: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// ConfirmArtImageUpload confirms that an image has been uploaded for an art
func (s *GeneratorService) ConfirmArtImageUpload(ctx context.Context, artName string) (*pb.Art, error) {
	req := connect.NewRequest(&pb.ConfirmArtImageUploadRequest{
		Name: artName,
	})

	resp, err := s.client.ConfirmArtImageUpload(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("art_name", artName).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to confirm art image upload")
		return nil, fmt.Errorf("failed to confirm art image upload: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// ListArts gets a list of arts for the authenticated user
func (s *GeneratorService) ListArts(ctx context.Context, userID string, pageSize int, pageToken string, orderBy, orderDirection string) (*pb.ListArtsResponse, error) {
	// Create the request payload with parent field
	req := connect.NewRequest(&pb.ListArtsRequest{
		Parent:         resource.BuildUserResourceName(userID),
		PageSize:       int32(pageSize),
		PageToken:      pageToken,
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
	})

	// Make the API call through the authenticated client
	resp, err := s.client.ListArts(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("userID", userID).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to list arts")
		return nil, fmt.Errorf("failed to list arts: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// ListCompositions gets a list of compositions for the authenticated user
func (s *GeneratorService) ListCompositions(ctx context.Context, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
	req := connect.NewRequest(&pb.ListCompositionsRequest{
		PageSize:  int32(pageSize),
		PageToken: pageToken,
	})

	resp, err := s.client.ListCompositions(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to list compositions")
		return nil, fmt.Errorf("failed to list compositions: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// CreateComposition creates a new composition
func (s *GeneratorService) CreateComposition(ctx context.Context, createRequest *pb.CreateCompositionRequest) (*pb.Composition, map[string][]string, error) {
	req := connect.NewRequest(createRequest)

	resp, err := s.client.CreateComposition(ctx, req)
	if err != nil {
		fieldErrors := s.parseErrorToFieldErrors(err)
		return nil, fieldErrors, err
	}

	return resp.Msg, nil, nil
}
