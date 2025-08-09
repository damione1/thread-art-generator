package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"

	"github.com/Damione1/thread-art-generator/core/util"
)

// InitializeFirebaseApp initializes Firebase app with proper authentication
// Uses environment-based configuration for seamless emulator/production switching
func InitializeFirebaseApp(ctx context.Context, config *util.Config) (*firebase.App, *auth.Client, error) {
	// Firebase SDK automatically detects emulator mode from FIREBASE_AUTH_EMULATOR_HOST
	// which is already set in the environment and loaded into config.Firebase.EmulatorHost
	if config.Firebase.EmulatorHost != "" {
		log.Info().
			Str("emulator_host", config.Firebase.EmulatorHost).
			Msg("Using Firebase Auth emulator")
	}

	// Initialize Firebase app - SDK automatically detects emulator vs production
	var app *firebase.App
	var err error

	projectID := getProjectID(config)
	firebaseConfig := &firebase.Config{
		ProjectID: projectID,
	}

	if config.GoogleApplicationCredentials != "" {
		// Use service account file if explicitly provided
		log.Info().
			Str("project_id", projectID).
			Str("credentials_path", config.GoogleApplicationCredentials).
			Msg("Initializing Firebase with service account credentials")

		opt := option.WithCredentialsFile(config.GoogleApplicationCredentials)
		app, err = firebase.NewApp(ctx, firebaseConfig, opt)
	} else {
		// Use default application credentials (ADC)
		// Works with emulator (no auth needed) and production (IAM roles)
		log.Info().
			Str("project_id", projectID).
			Msg("Initializing Firebase with default application credentials")

		app, err = firebase.NewApp(ctx, firebaseConfig)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	// Get Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}

	log.Info().
		Str("project_id", projectID).
		Bool("emulator_configured", config.Firebase.EmulatorHost != "").
		Msg("Firebase Admin SDK initialized successfully")

	return app, authClient, nil
}

// getProjectID returns the Firebase project ID from config
func getProjectID(config *util.Config) string {
	// Try Firebase-specific project ID from config first
	if config.Firebase.ProjectID != "" {
		return config.Firebase.ProjectID
	}

	// Fall back to Google Cloud project ID from config
	if config.GoogleCloudProject != "" {
		return config.GoogleCloudProject
	}

	// Default for local development
	return "demo-thread-art-generator"
}
