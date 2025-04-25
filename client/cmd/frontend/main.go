package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Damione1/thread-art-generator/client/internal/handlers"
	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Get port from environment or use default
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

	// Create server
	mux := http.NewServeMux()

	// Setup specific static file routes
	mux.HandleFunc("GET /static/css/{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /static/js/{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /static/images/{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))).ServeHTTP(w, r)
	})

	// Register handlers
	handlers.RegisterHandlers(mux)

	// Apply middleware
	handler := middleware.ChainMiddleware(
		mux,
		middleware.LoggingMiddleware,
		middleware.RecoveryMiddleware,
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Msgf("Starting frontend server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-stop
	log.Info().Msg("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server gracefully stopped")
}
