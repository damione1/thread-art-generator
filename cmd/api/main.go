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

	//go runGatewayServer(config, db)
	runGrpcServer(config, db)

}

func runGrpcServer(config util.Config, store *sql.DB) {
	log.Print("🍩 Starting gRPC server...")
	server, err := grpcApi.NewServer(config)
	if err != nil {
		log.Print(fmt.Sprintf("Failed to create gRPC server. %v", err))
	}
	log.Print("🍩 gRPC server created")
	gprcLogger := grpc.UnaryInterceptor(grpcApi.GrpcLogger)
	grpcServer := grpc.NewServer(gprcLogger)
	pb.RegisterArtGeneratorServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	log.Print("🍩 Starting to listen on port " + config.GRPCServerPort)

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", config.GRPCServerPort))
	if err != nil {
		log.Print(fmt.Sprintf("🍩 Failed to listen. %v", err))
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Print(fmt.Sprintf("🍩 Failed to serve gRPC server over port %s. %v", listener.Addr().String(), err))
	}
}

func runGatewayServer(config util.Config, store *sql.DB) {
	log.Print("🍦 Starting HTTP server...")
	server, err := grpcApi.NewServer(config)
	if err != nil {
		log.Print(fmt.Sprintf("🍦 Failed to create HTTP server. %v", err))
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
	err = pb.RegisterArtGeneratorServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("🍦 Failed to register HTTP gateway server.")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger", fs))
	log.Print(fmt.Sprintf("🍨 Swagger UI server started on http://localhost:%s/swagger/", config.HTTPServerPort))

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s/swagger/", config.HTTPServerPort))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen.")
	}

	log.Print(fmt.Sprintf("🍦 HTTP server started on http://localhost:%s/v1/", config.HTTPServerPort))
	handler := grpcApi.HttpLogger(mux)
	// Add request logging middleware
	handler = logRequest(handler)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("🍦 Failed to serve HTTP gateway server over port %s.", listener.Addr().String()))
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📡 Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
