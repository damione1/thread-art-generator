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
func (s *CompositionService) ListCompositions(ctx context.Context, r *http.Request, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

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
func (s *CompositionService) ListCompositionsForArt(ctx context.Context, r *http.Request, userID, artID string, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

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
func (s *CompositionService) DeleteComposition(ctx context.Context, r *http.Request, compositionName string) error {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

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
func (s *CompositionService) CreateComposition(ctx context.Context, r *http.Request, userID, artID string, composition *pb.Composition) (*pb.Composition, map[string][]string, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

	// Build the parent art resource name
	parent := fmt.Sprintf("users/%s/arts/%s", userID, artID)

	req := connect.NewRequest(&pb.CreateCompositionRequest{
		Parent:      parent,
		Composition: composition,
	})

	resp, err := s.client.CreateComposition(ctx, req)
	if err != nil {
		// Check if this is a field validation error
		fieldErrors := s.parseErrorToFieldErrors(err)
		if len(fieldErrors) > 0 {
			return nil, fieldErrors, err
		}

		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Str("parent", parent).
			Msg("Failed to create composition")
		return nil, nil, fmt.Errorf("failed to create composition: %s", standardErr.Message)
	}

	return resp.Msg, nil, nil
}

// GetComposition retrieves a specific composition
func (s *CompositionService) GetComposition(ctx context.Context, r *http.Request, compositionID string) (*pb.Composition, error) {
	// Add the session request to context so the transport layer can access it
	if r != nil {
		ctx = transport.WithSessionRequest(ctx, r)
	}

	req := connect.NewRequest(&pb.GetCompositionRequest{
		Name: compositionID, // Assume compositionID is already a full resource name
	})

	resp, err := s.client.GetComposition(ctx, req)
	if err != nil {
		standardErr := s.parseErrorForLogging(err)
		log.Error().
			Err(err).
			Str("errorType", string(standardErr.Type)).
			Str("message", standardErr.Message).
			Str("compositionName", compositionID).
			Msg("Failed to get composition")
		return nil, fmt.Errorf("failed to get composition: %s", standardErr.Message)
	}

	return resp.Msg, nil
}