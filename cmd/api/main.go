package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/cache"
	database "github.com/Damione1/thread-art-generator/core/db"
	"github.com/Damione1/thread-art-generator/core/interceptors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/service"
	"github.com/Damione1/thread-art-generator/core/util"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}

	_, err = database.ConnectDb(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to connect to database")
	}

	go cache.CleanExpiredCacheEntries()
	runGrpcServer(config)
}

func runGrpcServer(config util.Config) {
	log.Print("🍩 Starting gRPC server...")
	server, err := service.NewServer(config)
	if err != nil {
		log.Print(fmt.Sprintf("Failed to create gRPC server. %v", err))
	}
	defer server.Close()
	log.Print("🍩 gRPC server created")

	authService, err := createAuthService(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize auth service")
	}

	chainedInterceptors := grpc.ChainUnaryInterceptor(
		interceptors.GrpcLogger,
		interceptors.AuthInterceptor(authService, config.DB),
	)
	grpcServer := grpc.NewServer(chainedInterceptors)
	pb.RegisterArtGeneratorServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	log.Print("🍩 Starting to listen on port " + config.GRPCServerPort)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", config.GRPCServerPort))
	if err != nil {
		log.Print(fmt.Sprintf("🍩 Failed to listen. %v", err))
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Print(fmt.Sprintf("🍩 Failed to serve gRPC server over port %s. %v", listener.Addr().String(), err))
	}
}

func createAuthService(config util.Config) (auth.AuthService, error) {
	auth0Config := auth.Auth0Configuration{
		Domain:       config.Auth0.Domain,
		Audience:     config.Auth0.Audience,
		ClientID:     config.Auth0.ClientID,
		ClientSecret: config.Auth0.ClientSecret,
	}

	return auth.NewAuth0Service(auth0Config)
}
