package services

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/rs/zerolog/log"
)

// CompositionService handles composition-related operations
type CompositionService struct {
	*BaseService
}

// NewCompositionService creates a new composition service
func NewCompositionService(baseService *BaseService) *CompositionService {
	return &CompositionService{
		BaseService: baseService,
	}
}

// ListCompositions gets a list of compositions for the authenticated user
func (s *CompositionService) ListCompositions(ctx context.Context, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
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

// ListCompositionsForArt gets a list of compositions for a specific art
func (s *CompositionService) ListCompositionsForArt(ctx context.Context, userID, artID string, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
	// Build the art resource name as parent
	artResourceName := fmt.Sprintf("users/%s/arts/%s", userID, artID)

	req := connect.NewRequest(&pb.ListCompositionsRequest{
		Parent:    artResourceName,
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
			Str("parent", artResourceName).
			Msg("Failed to list compositions for art")
		return nil, fmt.Errorf("failed to list compositions for art: %s", standardErr.Message)
	}

	return resp.Msg, nil
}

// DeleteComposition deletes a composition
func (s *CompositionService) DeleteComposition(ctx context.Context, compositionName string) error {
	req := connect.NewRequest(&pb.DeleteCompositionRequest{
		Name: compositionName,
	})

	_, err := s.client.DeleteComposition(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Str("compositionName", compositionName).
			Msg("Failed to delete composition")
		return fmt.Errorf("failed to delete composition: %s", standardErr.Message)
	}

	return nil
}

// CreateComposition creates a new composition
func (s *CompositionService) CreateComposition(ctx context.Context, createRequest *pb.CreateCompositionRequest) (*pb.Composition, map[string][]string, error) {
	req := connect.NewRequest(createRequest)

	resp, err := s.client.CreateComposition(ctx, req)
	if err != nil {
		fieldErrors := s.parseErrorToFieldErrors(err)
		return nil, fieldErrors, err
	}

	return resp.Msg, nil, nil
}

// GetComposition retrieves a specific composition
func (s *CompositionService) GetComposition(ctx context.Context, userID, artID, compositionID string) (*pb.Composition, error) {
	// Build the composition resource name
	compositionName := fmt.Sprintf("users/%s/arts/%s/compositions/%s", userID, artID, compositionID)

	req := connect.NewRequest(&pb.GetCompositionRequest{
		Name: compositionName,
	})

	resp, err := s.client.GetComposition(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Str("compositionName", compositionName).
			Msg("Failed to get composition")
		return nil, fmt.Errorf("failed to get composition: %s", standardErr.Message)
	}

	return resp.Msg, nil
}
