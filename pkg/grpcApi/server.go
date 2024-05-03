package grpcApi

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"

	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/storage"
	"github.com/Damione1/thread-art-generator/pkg/token"
	"github.com/Damione1/thread-art-generator/pkg/util"
)

type Server struct {
	pb.UnimplementedArtGeneratorServiceServer
	config     util.Config
	tokenMaker token.Maker
	bucket     *blob.Bucket
}

func NewServer(config util.Config) (*Server, error) {
	var bucket *blob.Bucket
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token maker. %v", err)
	}

	log.Print("Current environment:", config.Environment)

	switch config.Environment {
	case "production":
		//To be implemented
	case "development":
		bucket, err = storage.NewMinioBlobStorage("minio:9000", "minio", "miniosecret", "local-bucket", false) //Default Minio credentials
		if err != nil {
			return nil, fmt.Errorf("failed to create minio storage. %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown environment: %s", config.Environment)
	}

	server := &Server{
		config:     config,
		tokenMaker: tokenMaker,
		bucket:     bucket,
	}

	return server, nil
}

func (s *Server) Close() error {
	if s.bucket != nil {
		return s.bucket.Close()
	}
	return nil
}
