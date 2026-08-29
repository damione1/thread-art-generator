package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/mail"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Server struct {
	config        util.Config
	bucket        storage.Bucket
	publicBaseURL string
	mailer        mail.Mailer
	queueClient   queue.Queue
}

func NewServer(config util.Config) (*Server, error) {
	var err error
	server := &Server{
		config:        config,
		publicBaseURL: config.Storage.PublicBaseURL,
	}

	server.mailer, err = mail.NewMailer(mail.ConfigFromUtil(config))
	if err != nil {
		return nil, fmt.Errorf("failed to create mailer: %w", err)
	}

	ctx := context.Background()
	server.bucket, err = storage.NewBucket(ctx, storage.BucketConfigFromUtil(config))
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket: %v", err)
	}

	if config.DB != nil {
		server.queueClient = queue.NewPostgresQueue(config.DB, queue.PostgresOptions{})
	}

	return server, nil
}

func (s *Server) Close() error {
	if s.queueClient != nil {
		return s.queueClient.Close()
	}
	return nil
}

// currentUser resolves the Postgres user from cookie/HMAC identity (UUID only).
func (s *Server) currentUser(ctx context.Context) (*models.User, error) {
	authID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	var sessionVersion int
	if id, found := auth.IdentityFromContext(ctx); found && id.UserID != "" {
		if id.Kind == auth.PrincipalService {
			return nil, pbErrors.PermissionDeniedError("service credentials cannot act as a user")
		}
		authID = id.UserID
		sessionVersion = id.SessionVersion
	}
	if !isPostgresUserID(authID) {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	user, err := models.Users(models.UserWhere.ID.EQ(authID)).One(ctx, s.config.DB)
	if err == nil {
		if err := s.requireCurrentSessionVersion(ctx, user.ID, sessionVersion); err != nil {
			return nil, err
		}
		return user, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pbErrors.NotFoundError("user not found")
	}
	return nil, pbErrors.InternalError("failed to get user", err)
}

func (s *Server) requireCurrentSessionVersion(ctx context.Context, userID string, sessionVersion int) error {
	if s == nil || s.config.DB == nil {
		return nil
	}
	var dbVersion int
	err := s.config.DB.QueryRowContext(ctx, `SELECT session_version FROM users WHERE id = $1`, userID).Scan(&dbVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return pbErrors.NotFoundError("user not found")
	}
	if err != nil {
		// Pre-migration databases have no session_version; skip the epoch check.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42703" {
			return nil
		}
		return pbErrors.InternalError("failed to get session version", err)
	}
	if sessionVersion <= 0 {
		sessionVersion = 1
	}
	if dbVersion <= 0 {
		dbVersion = 1
	}
	if sessionVersion != dbVersion {
		return pbErrors.UnauthenticatedError("session expired")
	}
	return nil
}

func isPostgresUserID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
