package services

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
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

// CreateArt creates a new art resource
func (s *ArtService) CreateArt(ctx context.Context, createArtRequest *pb.CreateArtRequest) (*pb.Art, map[string][]string, error) {
	req := connect.NewRequest(createArtRequest)

	resp, err := s.client.CreateArt(ctx, req)
	if err != nil {
		fieldErrors := s.parseErrorToFieldErrors(err)
		return nil, fieldErrors, err
	}

	return resp.Msg, nil, nil
}

// GetArt gets a specific art by its resource name
func (s *ArtService) GetArt(ctx context.Context, userID, artID string) (*pb.Art, error) {
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
func (s *ArtService) GetArtUploadUrl(ctx context.Context, userID, artID string) (*pb.GetArtUploadUrlResponse, error) {
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
func (s *ArtService) ConfirmArtImageUpload(ctx context.Context, artName string) (*pb.Art, error) {
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
func (s *ArtService) ListArts(ctx context.Context, userID string, pageSize int, pageToken string, orderBy, orderDirection string) (*pb.ListArtsResponse, error) {
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
