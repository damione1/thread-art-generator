package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	database "github.com/Damione1/thread-art-generator/pkg/db"
	"github.com/Damione1/thread-art-generator/pkg/grpcApi"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/util"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}

	db, err := database.ConnectDb(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to connect to database")
	}

	go runGatewayServer(config, db)
	runGrpcServer(config, db)

}

func runGrpcServer(config util.Config, store *sql.DB) {
	log.Print("🍩 Starting gRPC server...")
	server, err := grpcApi.NewServer(config)
	if err != nil {
		log.Print(fmt.Printf("Failed to create gRPC server. %v", err))
	}
	log.Print("🍩 gRPC server created")
	gprcLogger := grpc.UnaryInterceptor(grpcApi.GrpcLogger)
	grpcServer := grpc.NewServer(gprcLogger)
	pb.RegisterPortfolioServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	log.Print("🍩 Starting to listen on port " + config.GRPCServerAddress)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Print(fmt.Printf("🍩 Failed to listen. %v", err))
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Print(fmt.Printf("🍩 Failed to serve gRPC server over port %s. %v", listener.Addr().String(), err))
	}
}

func runGatewayServer(config util.Config, store *sql.DB) {
	log.Print("🍦 Starting HTTP server...")
	server, err := grpcApi.NewServer(config)
	if err != nil {
		log.Print(fmt.Printf("🍦 Failed to create HTTP server. %v", err))
	}

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterPortfolioServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("🍦 Failed to register HTTP gateway server.")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger", fs))
	log.Print("🍨 Swagger UI server started on http://" + config.HTTPServerAddress + "/swagger/ ")

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen.")
	}

	log.Print("🍦 HTTP server started on http://" + config.HTTPServerAddress + "/v1/ ")
	handler := grpcApi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("🍦 Failed to serve HTTP gateway server over port %s.", listener.Addr().String()))
	}
}
