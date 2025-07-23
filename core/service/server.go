package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	mailService "github.com/Damione1/thread-art-generator/core/mail"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

	// Initialize dual bucket storage system
	ctx := context.Background()
	server.storage, err = storage.NewDualBucketStorage(ctx, config.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create dual bucket storage: %v", err)
	}

	// Initialize queue client if URL is provided
	if config.Queue.URL != "" {
		server.queueClient, err = queue.NewRabbitMQClient(config.Queue.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to create queue client: %v", err)
		}
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

// getUserFromFirebaseUID is a helper method to get the internal user from Firebase UID
func (s *Server) getUserFromFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	user, err := models.Users(
		models.UserWhere.FirebaseUID.EQ(null.StringFrom(firebaseUID)),
	).One(ctx, s.config.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pbErrors.NotFoundError("user not found")
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}
	return user, nil
}

// createUserFromFirebaseClaims creates a new user record from Firebase auth claims
func (s *Server) createUserFromFirebaseClaims(ctx context.Context, firebaseUID, email, name, picture string) (*models.User, error) {
	// Parse name into first/last name components
	firstName := "User"
	var lastName null.String

	if name != "" {
		nameParts := strings.SplitN(name, " ", 2)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
		}
		if len(nameParts) > 1 {
			lastName = null.StringFrom(nameParts[1])
		}
	}

	// Create new user model with UUID primary key
	userDb := &models.User{
		ID:          uuid.New().String(), // Use UUID for primary key
		FirebaseUID: null.StringFrom(firebaseUID),
		Active:      true,
		Role:        models.RoleEnumUser,
		FirstName:   firstName,
		LastName:    lastName,
	}

	// Set optional fields
	if email != "" {
		userDb.Email = null.StringFrom(email)
	}

	if picture != "" {
		userDb.AvatarID = null.StringFrom(picture)
	}

	// Insert user into database
	if err := userDb.Insert(ctx, s.config.DB, boil.Infer()); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return userDb, nil
}

// validateInternalAPIKeyFromHeaders validates the internal API key from Connect-RPC HTTP headers
func (s *Server) validateInternalAPIKeyFromHeaders(headers http.Header) bool {
	// Get authorization header
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		log.Debug().Msg("No Authorization header found for internal API key validation")
		return false
	}

	// Extract Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Debug().Str("auth_header", authHeader).Msg("Authorization header doesn't start with 'Bearer '")
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	expectedToken := s.config.InternalAPIKey

	// Validate token (must be non-empty and match)
	isValid := token != "" && expectedToken != "" && token == expectedToken

	if !isValid {
		log.Warn().Msg("Internal API key validation failed")
	} else {
		log.Debug().Msg("Internal API key validation successful")
	}

	return isValid
}

// parseDisplayName parses a display name into first and last names
func (s *Server) parseDisplayName(displayName string) (firstName, lastName string) {
	if displayName == "" {
		return "User", ""
	}

	parts := strings.SplitN(strings.TrimSpace(displayName), " ", 2)
	firstName = parts[0]
	if len(parts) > 1 {
		lastName = parts[1]
	}

	// Ensure first name is not empty
	if firstName == "" {
		firstName = "User"
	}

	return firstName, lastName
}
