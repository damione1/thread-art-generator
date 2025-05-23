package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/client"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/bufbuild/connect-go"
	"github.com/rs/zerolog/log"
)

// GeneratorService handles all service interactions with the API
type GeneratorService struct {
	client         pbconnect.ArtGeneratorServiceClient
	sessionManager *auth.SessionManager
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(client pbconnect.ArtGeneratorServiceClient, sessionManager *auth.SessionManager) *GeneratorService {
	return &GeneratorService{
		client:         client,
		sessionManager: sessionManager,
	}
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
		return nil, fmt.Errorf("failed to get current user: %w", err)
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
	// Create the request payload
	req := connect.NewRequest(createArtRequest)

	// Make the API call through the authenticated client
	// The context already contains authentication info added by middleware
	resp, err := s.client.CreateArt(ctx, req)
	if err != nil {
		// Check if it's a connect error and extract field violations
		if connectErr, ok := err.(*connect.Error); ok {
			fieldErrors := make(map[string][]string)

			// Convert error details to field errors
			for _, detail := range connectErr.Details() {
				// Log raw detail for debugging
				log.Debug().Interface("detail", detail).Msg("Error detail")

				// Check if we have field violations in the error message
				if connectErr.Code() == connect.CodeInvalidArgument {
					// For field validation errors, map them to form fields
					if createArtRequest.Art.Title == "" {
						fieldErrors["art.title"] = []string{"Title is required"}
					} else {
						fieldErrors["art.title"] = []string{connectErr.Message()}
					}
				}
			}

			if len(fieldErrors) == 0 {
				// Fallback if no specific field errors were found
				fieldErrors["_form"] = []string{connectErr.Message()}
			}

			log.Debug().
				Interface("fieldErrors", fieldErrors).
				Msg("Processed field errors")

			return nil, fieldErrors, err
		}

		log.Error().Err(err).Msg("Failed to create art")
		return nil, nil, err
	}

	return resp.Msg, nil, nil
}

// ListArts gets a list of arts for the authenticated user
func (s *GeneratorService) ListArts(ctx context.Context, user *auth.UserInfo, pageSize int, pageToken string, orderBy, orderDirection string) (*pb.ListArtsResponse, error) {
	// Create the request payload with parent field
	req := connect.NewRequest(&pb.ListArtsRequest{
		Parent:         user.ID, // User ID already includes the "users/" prefix
		PageSize:       int32(pageSize),
		PageToken:      pageToken,
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
	})

	// Log request details for debugging
	log.Debug().
		Str("parent", user.ID).
		Int32("pageSize", int32(pageSize)).
		Str("pageToken", pageToken).
		Str("orderBy", orderBy).
		Str("orderDirection", orderDirection).
		Msg("Sending ListArts request")

	// Make the API call through the authenticated client
	// The context already contains authentication info added by middleware
	resp, err := s.client.ListArts(ctx, req)
	if err != nil {
		// Check if it's a connect error and extract field violations
		if connectErr, ok := err.(*connect.Error); ok {
			fieldErrors := make(map[string][]string)

			// Convert error details to field errors
			for _, detail := range connectErr.Details() {
				// Log raw detail for debugging
				log.Debug().Interface("detail", detail).Msg("Error detail")

				// Check if we have field violations in the error message
				if connectErr.Code() == connect.CodeInvalidArgument {
					// For field validation errors, map them to form fields
				}
			}

			if len(fieldErrors) == 0 {
				// Fallback if no specific field errors were found
				fieldErrors["_form"] = []string{connectErr.Message()}
			}

			log.Debug().
				Interface("fieldErrors", fieldErrors).
				Msg("Processed field errors")

			return nil, err
		}

		log.Error().Err(err).Msg("Failed to list arts")
		return nil, err
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
		log.Error().Err(err).Msg("Failed to list compositions")
		return nil, err
	}

	return resp.Msg, nil
}

// CreateComposition creates a new composition
func (s *GeneratorService) CreateComposition(ctx context.Context, createRequest *pb.CreateCompositionRequest) (*pb.Composition, map[string][]string, error) {
	req := connect.NewRequest(createRequest)

	resp, err := s.client.CreateComposition(ctx, req)
	if err != nil {
		if connectErr, ok := err.(*connect.Error); ok {
			fieldErrors := make(map[string][]string)

			// Handle field violations similar to CreateArt
			if connectErr.Code() == connect.CodeInvalidArgument {
				fieldErrors["_form"] = []string{connectErr.Message()}
			}

			log.Debug().
				Interface("fieldErrors", fieldErrors).
				Msg("Processed composition field errors")

			return nil, fieldErrors, err
		}

		log.Error().Err(err).Msg("Failed to create composition")
		return nil, nil, err
	}

	return resp.Msg, nil, nil
}
