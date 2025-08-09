package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/db/models"
	mailService "github.com/Damione1/thread-art-generator/core/mail"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/token"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Server struct {
	config      util.Config
	tokenMaker  token.Maker
	storage     storage.StorageProvider
	mailService mailService.MailService
	queueClient queue.QueueClient
}

func NewServer(config util.Config) (*Server, error) {
	var err error
	server := &Server{
		config: config,
	}

	server.tokenMaker, err = token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token maker. %v", err)
	}

	server.mailService, err = mailService.NewSendInBlueMailService(config.SendInBlueAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail service. %v", err)
	}

	// Initialize single storage provider using the new interface approach
	ctx := context.Background()
	
	server.storage, err = storage.NewStorageProviderFromUtil(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage provider: %v", err)
	}

	// Initialize queue client - prefer Pub/Sub if configured
	if config.PubSub.ProjectID != "" {
		server.queueClient, err = queue.NewPubSubClient(ctx, config.PubSub.ProjectID, config.PubSub.EmulatorHost, config.Environment)
		if err != nil {
			return nil, fmt.Errorf("failed to create Pub/Sub client: %v", err)
		}
	}

	return server, nil
}

// GetStorage returns the storage provider for all resource types
// Resource-based security is handled through Firebase storage rules and paths
func (s *Server) GetStorage() storage.StorageProvider {
	return s.storage
}

// GenerateUploadURL generates a signed URL for secure file uploads
func (s *Server) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
	// Create a storage service instance and delegate to it
	storageService, err := NewStorageService(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %w", err)
	}
	defer storageService.Close()

	// Create a connect request wrapper and call the storage service
	connectReq := connect.NewRequest(req)
	response, err := storageService.GenerateUploadURL(ctx, connectReq)
	if err != nil {
		return nil, err
	}
	
	return response.Msg, nil
}

// GenerateDownloadURL generates a signed URL for secure file downloads
func (s *Server) GenerateDownloadURL(ctx context.Context, req *pb.GenerateDownloadURLRequest) (*pb.GenerateDownloadURLResponse, error) {
	// Create a storage service instance and delegate to it
	storageService, err := NewStorageService(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %w", err)
	}
	defer storageService.Close()

	// Create a connect request wrapper and call the storage service
	connectReq := connect.NewRequest(req)
	response, err := storageService.GenerateDownloadURL(ctx, connectReq)
	if err != nil {
		return nil, err
	}
	
	return response.Msg, nil
}

func (s *Server) Close() error {
	var err error

	// Close storage connection
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