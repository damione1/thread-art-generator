package services

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/transport"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/rs/zerolog/log"
)

// ArtService handles art-related operations
type ArtService struct {
	*BaseService
}

// NewArtService creates a new art service
func NewArtService(baseService *BaseService) *ArtService {
	return &ArtService{
		BaseService: baseService,
	}
}

// CreateArt creates a new art for the authenticated user
func (s *ArtService) CreateArt(ctx context.Context, r *http.Request, userID string, title string) (*pb.Art, map[string][]string, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

	// Create the request payload with parent field (Google AIP pattern)
	req := connect.NewRequest(&pb.CreateArtRequest{
		Art: &pb.Art{
			Title: title,
		},
		Parent: resource.BuildUserResourceName(userID), // "users/{user_id}"
	})

	// Make the API call through the authenticated client
	resp, err := s.client.CreateArt(ctx, req)
	if err != nil {
		fieldErrors := s.parseErrorToFieldErrors(err)
		return nil, fieldErrors, err
	}

	return resp.Msg, nil, nil
}

// GetArt gets a specific art by ID for the authenticated user
func (s *ArtService) GetArt(ctx context.Context, r *http.Request, userID, artID string) (*pb.Art, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

	// Create the request payload
	req := connect.NewRequest(&pb.GetArtRequest{
		Name: resource.BuildArtResourceName(userID, artID),
	})

	// Make the API call through the authenticated client
	resp, err := s.client.GetArt(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("userID", userID).
			Str("artID", artID).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Msg("Failed to get art")
		return nil, fmt.Errorf("failed to get art: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// ListArts gets a list of arts for the authenticated user
func (s *ArtService) ListArts(ctx context.Context, r *http.Request, userID string, pageSize int, pageToken string, orderBy, orderDirection string) (*pb.ListArtsResponse, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

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