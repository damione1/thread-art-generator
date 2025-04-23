package main

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/cache"
	database "github.com/Damione1/thread-art-generator/core/db"
	"github.com/Damione1/thread-art-generator/core/interceptors"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
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
	runConnectServer(config)
}

func runConnectServer(config util.Config) {
	log.Print("🍩 Starting Connect server...")
	server, err := service.NewServer(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}
	defer server.Close()
	log.Print("🍩 Server created")

	authService, err := createAuthService(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize auth service")
	}

	// Define our Connect interceptors
	interceptorChain := connect.WithInterceptors(
		interceptors.ConnectLogger(),
		interceptors.AuthMiddleware(authService, config.DB),
	)

	// Create Connect adapter
	adapter := service.NewConnectAdapter(server)

	// Create API handler
	path, handler := pbconnect.NewArtGeneratorServiceHandler(adapter, interceptorChain)

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Adjust this in production
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
			"X-Requested-With",
			"X-User-Agent",
			"X-Grpc-Web",
			"Origin",
			"Access-Control-Request-Method",
			"Access-Control-Request-Headers",
			"Grpc-Accept-Encoding",
			"Grpc-Timeout",
			"Grpc-Status-Details-Bin",
		},
		ExposedHeaders: []string{
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
			"Grpc-Status",
			"Grpc-Message",
			"Grpc-Accept-Encoding",
			"Grpc-Timeout",
			"Grpc-Status-Details-Bin",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Credentials",
		},
		AllowCredentials: true,
		MaxAge:           86400,                               // 24 hours
		Debug:            config.Environment == "development", // Enable debug for development
	})

	// Create a mux for routing
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Register the service
	mux.Handle(path, corsHandler.Handler(handler))

	// Create the server
	serverPort := config.HTTPServerPort
	if serverPort == "" {
		serverPort = config.GRPCServerPort // Fallback to GRPC port if HTTP port not set
		log.Warn().Msg("HTTP_SERVER_PORT not set, using GRPC_SERVER_PORT instead")
	}
	addr := fmt.Sprintf("0.0.0.0:%s", serverPort)
	log.Print("🍩 Starting to listen on " + addr)

	// Configure HTTP/2 server
	srv := &http.Server{
		Addr:    addr,
		Handler: interceptors.HttpLogger(mux),
		// Configure TLS for HTTP/2
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			MinVersion: tls.VersionTLS12,
		},
	}

	// Get certificate file paths from config
	certFile := config.TLSCertFile
	keyFile := config.TLSKeyFile

	// Determine whether to use HTTP/2 with TLS or fallback to HTTP/1.1
	if certFile != "" && keyFile != "" {
		// Enable HTTP/2 with TLS using certificates
		log.Info().
			Str("certFile", certFile).
			Str("keyFile", keyFile).
			Msg("Starting server with HTTP/2 and TLS")

		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		// Fallback to HTTP/1.1 without TLS
		log.Warn().Msg("TLS certificates not specified. Starting with HTTP/1.1 without TLS")
		err = srv.ListenAndServe()
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func createAuthService(config util.Config) (auth.AuthService, error) {
	auth0Config := auth.Auth0Configuration{
		Domain:                    config.Auth0.Domain,
		Audience:                  config.Auth0.Audience,
		ClientID:                  config.Auth0.ClientID,
		ClientSecret:              config.Auth0.ClientSecret,
		ManagementApiClientID:     config.Auth0.ManagementApiClientID,
		ManagementApiClientSecret: config.Auth0.ManagementApiClientSecret,
	}

	return auth.NewAuth0Service(auth0Config)
}
