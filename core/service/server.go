package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	mailService "github.com/Damione1/thread-art-generator/core/mail"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Server struct {
	config      util.Config
	storage     *storage.DualBucketStorage
	mailService mailService.MailService
	queueClient queue.QueueClient
}

func NewServer(config util.Config) (*Server, error) {
	var err error
	server := &Server{
		config: config,
	}

	server.mailService, err = mailService.NewSendInBlueMailService(config.SendInBlueAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail service. %v", err)
	}

	ctx := context.Background()
	server.storage, err = storage.NewDualBucketStorage(ctx, config.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create dual bucket storage: %v", err)
	}

	if config.DB != nil {
		server.queueClient = queue.NewPostgresQueue(config.DB, queue.PostgresOptions{})
	}

	return server, nil
}

func (s *Server) Close() error {
	var err error

	// Close storage connections
	if s.storage != nil {
		if storageErr := s.storage.Close(); storageErr != nil {
			err = storageErr
		}
	}

	// Close queue connection
	if s.queueClient != nil {
		if queueErr := s.queueClient.Close(); queueErr != nil {
			if err == nil {
				err = queueErr
			} else {
				err = fmt.Errorf("%v; %v", err, queueErr)
			}
		}
	}

	return err
}

// currentUser resolves the Postgres user from cookie/HMAC identity (UUID only).
func (s *Server) currentUser(ctx context.Context) (*models.User, error) {
	authID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	if id, found := auth.IdentityFromContext(ctx); found && id.UserID != "" {
		if id.Kind == auth.PrincipalService {
			return nil, pbErrors.PermissionDeniedError("service credentials cannot act as a user")
		}
		authID = id.UserID
	}
	if !isPostgresUserID(authID) {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	user, err := models.Users(models.UserWhere.ID.EQ(authID)).One(ctx, s.config.DB)
	if err == nil {
		return user, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pbErrors.NotFoundError("user not found")
	}
	return nil, pbErrors.InternalError("failed to get user", err)
}

func isPostgresUserID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
