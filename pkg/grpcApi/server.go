package grpcApi

import (
	"fmt"

	"gocloud.dev/blob"

	mailService "github.com/Damione1/thread-art-generator/pkg/mail"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/storage"
	"github.com/Damione1/thread-art-generator/pkg/token"
	"github.com/Damione1/thread-art-generator/pkg/util"
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
		server.bucket, err = storage.NewMinioBlobStorage("minio:9000", "minio", "miniosecret", "local-bucket", false) //Default Minio credentials
		if err != nil {
			return nil, fmt.Errorf("failed to create minio storage. %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown environment: %s", config.Environment)
	}

	return server, nil
}

func (s *Server) Close() error {
	if s.bucket != nil {
		return s.bucket.Close()
	}
	return nil
}
