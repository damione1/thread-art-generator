package main

import (
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/core/auth"
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

	// Initialize PASETO service for BFF → API authentication
	pasetoService, err := createPasetoService(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize PASETO service")
	}

	// Define our Connect interceptors - using PASETO for secure, stateless authentication
	interceptorChain := connect.WithInterceptors(
		interceptors.ConnectLogger(),
		interceptors.PasetoAuthMiddleware(pasetoService), // PASETO-only auth for internal communication
	)

	// Create Connect adapters
	adapter := service.NewConnectAdapter(server)
	functionsAdapter := service.NewFirebaseFunctionsConnectAdapter(server)

	// Create API handlers
	path, handler := pbconnect.NewArtGeneratorServiceHandler(adapter, interceptorChain)
	functionsPath, functionsHandler := pbconnect.NewFirebaseFunctionsServiceHandler(functionsAdapter, interceptorChain)

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Adjust this in production
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Connect-Protocol-Version",
			"X-Requested-With",
			"X-User-Agent",
			"X-Grpc-Web",
			"Origin",
			"Access-Control-Request-Method",
			"Access-Control-Request-Headers",
		},
		ExposedHeaders: []string{
			"Connect-Protocol-Version",
			"Grpc-Status",
			"Grpc-Message",
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

	// Register the services
	mux.Handle(path, corsHandler.Handler(handler))
	mux.Handle(functionsPath, corsHandler.Handler(functionsHandler))

	// Create the server
	serverPort := config.HTTPServerPort
	if serverPort == "" {
		serverPort = config.GRPCServerPort // Fallback to GRPC port if HTTP port not set
		log.Warn().Msg("HTTP_SERVER_PORT not set, using GRPC_SERVER_PORT instead")
	}
	addr := fmt.Sprintf("0.0.0.0:%s", serverPort)
	log.Print("🍩 Starting to listen on " + addr)

	err = http.ListenAndServe(addr, interceptors.HttpLogger(mux))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func createPasetoService(config util.Config) (*auth.PasetoService, error) {
	// Validate PASETO configuration at startup
	if config.Paseto.SecretKey == "" {
		return nil, fmt.Errorf("PASETO_SECRET_KEY is not configured")
	}

	if len(config.Paseto.SecretKey) != 32 {
		return nil, fmt.Errorf("PASETO_SECRET_KEY must be exactly 32 bytes, got %d bytes", len(config.Paseto.SecretKey))
	}

	if config.Paseto.Issuer == "" {
		log.Warn().Msg("PASETO_ISSUER not set, using default 'thread-art-generator'")
		config.Paseto.Issuer = "thread-art-generator"
	}

	if config.Paseto.TTLMinutes <= 0 {
		log.Warn().Msg("PASETO_TTL_MINUTES not set or invalid, using default 15 minutes")
		config.Paseto.TTLMinutes = 15
	}

	pasetoConfig := auth.PasetoConfig{
		SecretKey:  config.Paseto.SecretKey,
		Issuer:     config.Paseto.Issuer,
		TTLMinutes: config.Paseto.TTLMinutes,
	}

	log.Info().
		Str("issuer", pasetoConfig.Issuer).
		Int("ttl_minutes", pasetoConfig.TTLMinutes).
		Msg("PASETO service configured successfully")

	return auth.NewPasetoService(pasetoConfig)
}
