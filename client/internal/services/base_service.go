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

func NewBaseService(client pbconnect.ArtGeneratorServiceClient, sessionManager *auth.SCSSessionManager) *BaseService {
	return &BaseService{
		client:         client,
		sessionManager: sessionManager,
	}
}

func (s *BaseService) parseErrorToFieldErrors(err error) map[string][]string {
	return coreErrors.FormFields(err)
}

func (s *BaseService) parseErrorForLogging(err error) *coreErrors.StandardError {
	return coreErrors.FromConnectError(err)
}
