package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	mailService "github.com/Damione1/thread-art-generator/core/mail"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/token"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Server struct {
	config      util.Config
	tokenMaker  token.Maker
	bucket      *storage.BlobStorage
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

	// Initialize blob storage based on environment and configuration
	ctx := context.Background()

	// Convert provider string to StorageProvider type
	var provider storage.StorageProvider
	switch config.Storage.Provider {
	case "s3":
		provider = storage.ProviderS3
	case "minio":
		provider = storage.ProviderMinIO
	case "gcs":
		provider = storage.ProviderGCS
	default:
		// Default to MinIO in development, S3 in production
		if config.Environment == "development" {
			provider = storage.ProviderMinIO
		} else {
			provider = storage.ProviderS3
		}
	}

	// Create storage configuration from environment variables
	storageConfig := storage.BlobStorageConfig{
		Provider:         provider,
		Bucket:           config.Storage.Bucket,
		Region:           config.Storage.Region,
		InternalEndpoint: config.Storage.InternalEndpoint,
		ExternalEndpoint: config.Storage.ExternalEndpoint,
		UseSSL:           config.Storage.UseSSL,
		ForceExternalSSL: config.Storage.ForceExternalSSL,
		AccessKey:        config.Storage.AccessKey,
		SecretKey:        config.Storage.SecretKey,
		GCPProjectID:     config.Storage.GCPProjectID,
	}

	// If config values are missing, provide reasonable defaults based on environment
	if storageConfig.Bucket == "" {
		storageConfig.Bucket = "local-bucket"
	}

	if storageConfig.Region == "" {
		storageConfig.Region = "us-east-1" // Default region for S3/MinIO
	}

	// Set up endpoints based on environment if not provided
	if config.Environment == "development" && provider == storage.ProviderMinIO {
		if storageConfig.InternalEndpoint == "" {
			storageConfig.InternalEndpoint = "http://minio:9000"
		}
		if storageConfig.ExternalEndpoint == "" {
			storageConfig.ExternalEndpoint = "http://localhost:9000"
		}
	}

	server.bucket, err = storage.NewBlobStorage(ctx, storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob storage: %v", err)
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

func (s *Server) GetTokenMaker() token.Maker {
	return s.tokenMaker
}

func (s *Server) Close() error {
	var err error

	// Close bucket connection
	if s.bucket != nil && s.bucket.Bucket != nil {
		if bucketErr := s.bucket.Close(); bucketErr != nil {
			err = bucketErr
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

	// Create new user model
	userDb := &models.User{
		ID:          firebaseUID, // Use Firebase UID as primary key for consistency
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
