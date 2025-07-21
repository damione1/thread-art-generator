package main

import (
	"context"
	"database/sql"
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
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "github.com/lib/pq"
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

	// Connect to PostgreSQL database for sessions
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "db"
	}
	dbUser := os.Getenv("POSTGRES_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "threadmachine"
	}

	dbDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbHost, dbUser, dbPassword, dbName)
	
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	// Create SCS session manager with PostgreSQL store
	sessionManager, err := auth.NewSCSSessionManager(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create SCS session manager")
	}

	// Create HTTP client with auth transport (updated for SCS session manager)
	httpClient := &http.Client{
		Transport: &client.FirebaseAuthTransport{
			SessionManager: sessionManager,
			Base:           http.DefaultTransport,
		},
	}

	// Get API URL from environment
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://api:9090"
	}

	// Create connect client directly
	artGeneratorClient := pbconnect.NewArtGeneratorServiceClient(
		httpClient,
		apiURL,
	)

	// Initialize Firebase auth service
	emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	environment := os.Getenv("ENVIRONMENT")
	isEmulator := emulatorHost != "" || environment == "development"
	
	firebaseConfig := coreauth.FirebaseConfiguration{
		ProjectID: "demo-thread-art-generator", // Default for emulator
	}
	
	if isEmulator {
		firebaseConfig.EmulatorHost = "host.docker.internal:9099"
	} else {
		firebaseConfig.ProjectID = os.Getenv("FIREBASE_PROJECT_ID")
	}
	
	firebaseAuth, err := coreauth.NewFirebaseAuthService(firebaseConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Firebase auth service")
	}

	// Create generator service
	generatorService := services.NewGeneratorService(artGeneratorClient, sessionManager)

	// Create Firebase auth handler with all services
	authHandler := handlers.NewFirebaseAuthHandlerWithServices(firebaseAuth, sessionManager, generatorService, db)
	pageHandler := handlers.NewPageHandler(generatorService)
	artHandler := handlers.NewArtHandler(generatorService)

	// Create router
	r := chi.NewRouter()

	// Global middleware - updated for Firebase and SCS sessions
	r.Use(sessionManager.GetSessionManager().LoadAndSave)
	r.Use(middleware.FirebaseAuthMiddleware(sessionManager))
	r.Use(middleware.APIAuthMiddleware(sessionManager))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		// Firebase Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/sync", authHandler.AuthSync)
			r.Post("/logout", authHandler.Logout)
			r.Get("/logout", authHandler.Logout) // Support GET for logout links
			r.Get("/status", authHandler.Status)
		})

		// Public home page
		r.Get("/", pageHandler.HomePage)
		
		// Auth pages
		r.Get("/login", pageHandler.LoginPage)
		r.Get("/signup", pageHandler.SignupPage)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		// Firebase auth is handled by the global middleware - all routes here require authentication

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/", pageHandler.DashboardPage)
			r.Route("/arts", func(r chi.Router) {
				r.Get("/new", artHandler.NewArtPage)
				r.Post("/new", artHandler.CreateArt)
				r.Get("/{artId}", artHandler.ViewArtPage)
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

			// Art upload API routes
			r.Get("/get-upload-url/{artId}", artHandler.GetArtUploadUrl)
			r.Post("/confirm-upload/{artId}", artHandler.ConfirmArtImageUpload)
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
