package service

import (
	"context"
	"fmt"

	mailService "github.com/Damione1/thread-art-generator/core/mail"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/token"
	"github.com/Damione1/thread-art-generator/core/util"
)

type Server struct {
	pb.UnimplementedArtGeneratorServiceServer
	config      util.Config
	tokenMaker  token.Maker
	bucket      *storage.BlobStorage
	mailService mailService.MailService
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

	// Initialize blob storage based on environment
	ctx := context.Background()
	switch config.Environment {
	case "production":
		// In production, use GCS or S3 based on configuration
		// if config.StorageProvider == "gcs" {
		// 	// GCS configuration
		// 	storageConfig := storage.BlobStorageConfig{
		// 		Provider:     storage.ProviderGCS,
		// 		Bucket:       config.BucketName,
		// 		GCPProjectID: config.GCPProjectID,
		// 		// GCP credentials will be loaded from environment variables or instance metadata
		// 	}
		// 	server.bucket, err = storage.NewBlobStorage(ctx, storageConfig)
		// } else {
		// 	// Default to S3 for production if not GCS
		// 	storageConfig := storage.BlobStorageConfig{
		// 		Provider:  storage.ProviderS3,
		// 		Bucket:    config.BucketName,
		// 		Region:    config.AWSRegion,
		// 		AccessKey: config.AWSAccessKey,
		// 		SecretKey: config.AWSSecretKey,
		// 		PublicURL: config.StoragePublicURL,
		// 		UseSSL:    true,
		// 	}
		// 	server.bucket, err = storage.NewBlobStorage(ctx, storageConfig)
		// }
	case "development":
		// In development, use MinIO with separate internal and external endpoints
		storageConfig := storage.BlobStorageConfig{
			Provider:         storage.ProviderMinIO,
			Bucket:           "local-bucket",
			Region:           "us-east-1",
			InternalEndpoint: "http://minio:9000",         // Internal HTTP endpoint within Docker
			ExternalEndpoint: "https://storage.tag.local", // External HTTPS endpoint for clients
			UseSSL:           false,                       // Don't use SSL for internal communication
			AccessKey:        "minio",
			SecretKey:        "minio123",
		}
		server.bucket, err = storage.NewBlobStorage(ctx, storageConfig)
	default:
		return nil, fmt.Errorf("unknown environment: %s", config.Environment)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create blob storage: %v", err)
	}

	return server, nil
}

func (s *Server) GetTokenMaker() token.Maker {
	return s.tokenMaker
}

func (s *Server) Close() error {
	if s.bucket != nil && s.bucket.Bucket != nil {
		return s.bucket.Close()
	}
	return nil
}
