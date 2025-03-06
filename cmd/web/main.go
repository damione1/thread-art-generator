package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/Damione1/thread-art-generator/web/client"
	"github.com/Damione1/thread-art-generator/web/router"
)

func main() {
	// Load configuration
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize gRPC client
	grpcClient, err := client.NewGrpcClient(config)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcClient.Close()

	// Initialize router
	r := router.NewRouter(grpcClient)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.HTTPServerPort),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🌐 Web server started on http://localhost:%s", config.HTTPServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
