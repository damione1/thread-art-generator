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
