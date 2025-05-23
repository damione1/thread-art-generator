package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/client"
	"github.com/Damione1/thread-art-generator/client/internal/handlers"
	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8080"
	}

	// Get the current working directory to determine static file path
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get working directory")
	}

	// Determine the static files directory based on environment
	staticDir := filepath.Join(workDir, "client/public")
	// Check if we're running in Docker
	if _, err := os.Stat("/app/client/public"); err == nil {
		staticDir = "/app/client/public"
	}

	log.Info().Str("staticDir", staticDir).Msg("Using static files directory")

	// Load Auth0 configuration
	auth0Config := auth.NewConfig()

	// Get cookie encryption keys from environment or generate them
	// These should be set in production for consistency across restarts and multiple instances
	var hashKey, blockKey []byte

	hashKeyStr := os.Getenv("COOKIE_HASH_KEY")
	if hashKeyStr == "" {
		log.Warn().Msg("COOKIE_HASH_KEY not set, generating random key. Sessions will be invalidated on restart.")
		hashKey = generateRandomKey(32)
	} else {
		// Ensure hash key is exactly 32	 bytes
		hashKey = []byte(hashKeyStr)
		if len(hashKey) < 32 {
			// Pad key if too short
			paddedKey := make([]byte, 32)
			copy(paddedKey, hashKey)
			hashKey = paddedKey
			log.Warn().Msg("COOKIE_HASH_KEY was shorter than 32 bytes, padded with zeros")
		} else if len(hashKey) > 32 {
			// Truncate key if too long
			hashKey = hashKey[:32]
			log.Warn().Msg("COOKIE_HASH_KEY was longer than 32 bytes, truncated to 32 bytes")
		}
	}

	blockKeyStr := os.Getenv("COOKIE_BLOCK_KEY")
	if blockKeyStr == "" {
		log.Warn().Msg("COOKIE_BLOCK_KEY not set, generating random key. Sessions will be invalidated on restart.")
		blockKey = generateRandomKey(32)
	} else {
		// Ensure block key is exactly 32 bytes
		blockKey = []byte(blockKeyStr)
		if len(blockKey) < 32 {
			// Pad key if too short
			paddedKey := make([]byte, 32)
			copy(paddedKey, blockKey)
			blockKey = paddedKey
			log.Warn().Msg("COOKIE_BLOCK_KEY was shorter than 32 bytes, padded with zeros")
		} else if len(blockKey) > 32 {
			// Truncate key if too long
			blockKey = blockKey[:32]
			log.Warn().Msg("COOKIE_BLOCK_KEY was longer than 32 bytes, truncated to 32 bytes")
		}
	}

	// Create session manager
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	sessionManager, err := auth.NewSessionManager(redisAddr, hashKey, blockKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create session manager")
	}

	// Create HTTP client with auth transport
	httpClient := &http.Client{
		Transport: &client.AuthTransport{
			SessionManager: sessionManager,
			Base:           http.DefaultTransport,
		},
	}

	// Create connect client directly
	artGeneratorClient := pbconnect.NewArtGeneratorServiceClient(
		httpClient,
		auth0Config.APIBaseURL,
	)

	// Create Auth0 service
	auth0Service := auth.NewAuth0Service(auth0Config, sessionManager)

	// Create generator service
	generatorService := services.NewGeneratorService(artGeneratorClient, sessionManager)

	// Create handlers
	authHandler := handlers.NewAuthHandler(auth0Service)
	pageHandler := handlers.NewPageHandler(generatorService)
	artHandler := handlers.NewArtHandler(generatorService)

	// Create router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.WithAuthInfo(sessionManager))
	r.Use(middleware.EnrichUser(generatorService))
	// Process CSRF/Auth token from HTMX requests
	r.Use(middleware.ProcessAuthToken(sessionManager))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", authHandler.Login)
			r.Get("/callback", authHandler.Callback)
			r.Get("/logout", authHandler.Logout)
			r.Get("/status", authHandler.Status)
		})

		// Public home page
		r.Get("/", pageHandler.HomePage)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		// Apply auth middleware for protected routes
		r.Use(middleware.RequireAuth(sessionManager, "/auth/login"))

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/", pageHandler.DashboardPage)
			r.Route("/arts", func(r chi.Router) {
				r.Get("/new", artHandler.NewArtPage)
				r.Post("/new", artHandler.CreateArt)
			})
		})

		// Protected API routes
		r.Route("/api", func(r chi.Router) {
			r.Get("/user", func(w http.ResponseWriter, r *http.Request) {
				// Get user from context (added by middleware)
				user, ok := middleware.UserFromContext(r.Context())

				w.Header().Set("Content-Type", "application/json")
				if !ok || user == nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"Unauthorized"}`))
					return
				}

				// Return user info as JSON
				w.Write([]byte(fmt.Sprintf(`{"id":"%s","name":"%s","email":"%s"}`,
					user.ID, user.Name, user.Email)))
			})
		})
	})

	// Static files
	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Run the server in a goroutine
	go func() {
		log.Info().Str("port", port).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}

	log.Info().Msg("Server gracefully stopped")
}

// generateRandomKey creates a random key for session encryption
func generateRandomKey(length int) []byte {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal().Err(err).Msg("Failed to generate random key")
	}
	return bytes
}
