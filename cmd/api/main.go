package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

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

	go runHttpServer(config)
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

	// Pass the tokenMaker to the AuthInterceptor
	chainedInterceptors := grpc.ChainUnaryInterceptor(
		interceptors.GrpcLogger,
		interceptors.AuthInterceptor(server.GetTokenMaker()),
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

func runHttpServer(config util.Config) {
	log.Print("🍦 Starting HTTP server...")

	server, err := service.NewServer(config)
	if err != nil {
		log.Fatal().Err(err).Msg("🍦 Failed to create HTTP server.")
	}
	defer server.Close()

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

	if err := pb.RegisterArtGeneratorServiceHandlerServer(ctx, grpcMux, server); err != nil {
		log.Fatal().Err(err).Msg("🍦 Failed to register HTTP gateway server.")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger", fs))
	log.Print(fmt.Sprintf("🍨 Swagger UI server started on http://localhost:%s/swagger/", config.HTTPServerPort))

	// Pass the server to the file upload handler
	mux.Handle("/v1/upload", service.HandleBinaryFileUpload(server))

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", config.HTTPServerPort))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen.")
	}

	log.Print(fmt.Sprintf("🍦 HTTP server started on http://localhost:%s/v1/", config.HTTPServerPort))

	handler := interceptors.HttpLogger(mux)
	handler = interceptors.HttpAuthInterceptor(server.GetTokenMaker(), handler)

	// Set CORS headers
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", config.FrontendUrl)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	// Graceful shutdown
	srv := &http.Server{
		Handler: corsHandler(handler),
	}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg(fmt.Sprintf("🍦 Failed to serve HTTP gateway server over port %s.", listener.Addr().String()))
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown

	log.Print("🍦 Shutting down HTTP server...")

	timeoutCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Fatal().Err(err).Msg("🍦 HTTP server graceful shutdown failed.")
	}
	log.Print("🍦 HTTP server gracefully stopped.")
}
