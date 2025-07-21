package services

import (
	"github.com/Damione1/thread-art-generator/client/internal/auth"
	coreErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
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
