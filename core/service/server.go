package service

import (
	"fmt"

	"gocloud.dev/blob"

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
	bucket      *blob.Bucket
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

	switch config.Environment {
	case "production":
		//To be implemented
	case "development":
		// Use https://tag.local/storage as the public URL for signed URLs
		// while still connecting to minio:9000 internally
		server.bucket, err = storage.NewMinioBlobStorage(
			"minio:9000",                // Internal endpoint for operations
			"minio",                     // Access key
			"minio123",                  // Secret key
			"local-bucket",              // Bucket name
			false,                       // Use SSL
			"https://tag.local/storage", // Public URL for signed URLs
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create minio storage. %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown environment: %s", config.Environment)
	}

	return server, nil
}

func (s *Server) GetTokenMaker() token.Maker {
	return s.tokenMaker
}

func (s *Server) Close() error {
	if s.bucket != nil {
		return s.bucket.Close()
	}
	return nil
}
