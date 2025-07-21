package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// ServiceAccountConfig holds Firebase service account configuration
type ServiceAccountConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// InitializeFirebaseApp initializes Firebase app with proper authentication
// For local development with emulator, it detects emulator environment automatically
// For production, it requires either:
// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable pointing to service account JSON file
// 2. Service account configuration via environment variables
// 3. Default application credentials (when running on Google Cloud)
func InitializeFirebaseApp(ctx context.Context) (*firebase.App, *auth.Client, error) {
	var app *firebase.App
	var err error

	// Check if we're in emulator mode
	if isEmulatorMode() {
		// In emulator mode, we can initialize without credentials
		app, err = firebase.NewApp(ctx, &firebase.Config{
			ProjectID: getProjectID(),
		})
	} else {
		// Production mode - check for credentials
		credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credentialsPath != "" {
			// Use service account file
			opt := option.WithCredentialsFile(credentialsPath)
			app, err = firebase.NewApp(ctx, &firebase.Config{
				ProjectID: getProjectID(),
			}, opt)
		} else {
			// Try to build credentials from environment variables
			serviceAccount, configErr := buildServiceAccountFromEnv()
			if configErr != nil {
				// Fall back to default application credentials
				app, err = firebase.NewApp(ctx, &firebase.Config{
					ProjectID: getProjectID(),
				})
				if err != nil {
					return nil, nil, fmt.Errorf("failed to initialize Firebase app with default credentials: %v (also failed to build from env: %v)", err, configErr)
				}
			} else {
				// Use constructed service account
				credentialsJSON, jsonErr := json.Marshal(serviceAccount)
				if jsonErr != nil {
					return nil, nil, fmt.Errorf("failed to marshal service account credentials: %v", jsonErr)
				}

				opt := option.WithCredentialsJSON(credentialsJSON)
				app, err = firebase.NewApp(ctx, &firebase.Config{
					ProjectID: serviceAccount.ProjectID,
				}, opt)
			}
		}
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	// Get Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}

	return app, authClient, nil
}

// isEmulatorMode checks if we're running in Firebase emulator mode
func isEmulatorMode() bool {
	// Check for Firebase Auth emulator environment variable
	authEmulator := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if authEmulator != "" {
		return true
	}

	// Check for legacy environment variable
	legacyEmulator := os.Getenv("FIREBASE_EMULATOR_HOST")
	if legacyEmulator != "" {
		return true
	}

	// Check if environment is explicitly set to development with demo project
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	environment := os.Getenv("ENVIRONMENT")

	return projectID == "demo-thread-art-generator" && environment == "development"
}

// getProjectID returns the Firebase project ID from environment
func getProjectID() string {
	// Try Firebase-specific environment variable first
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID != "" {
		return projectID
	}

	// Fall back to Google Cloud project ID
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID != "" {
		return projectID
	}

	// Default for local development
	return "demo-thread-art-generator"
}

// buildServiceAccountFromEnv constructs a service account config from environment variables
func buildServiceAccountFromEnv() (*ServiceAccountConfig, error) {
	// Check if all required environment variables are set
	requiredVars := map[string]string{
		"FIREBASE_ADMIN_TYPE":                        "type",
		"FIREBASE_ADMIN_PROJECT_ID":                  "project_id",
		"FIREBASE_ADMIN_PRIVATE_KEY_ID":              "private_key_id",
		"FIREBASE_ADMIN_PRIVATE_KEY":                 "private_key",
		"FIREBASE_ADMIN_CLIENT_EMAIL":                "client_email",
		"FIREBASE_ADMIN_CLIENT_ID":                   "client_id",
		"FIREBASE_ADMIN_AUTH_URI":                    "auth_uri",
		"FIREBASE_ADMIN_TOKEN_URI":                   "token_uri",
		"FIREBASE_ADMIN_AUTH_PROVIDER_X509_CERT_URL": "auth_provider_x509_cert_url",
		"FIREBASE_ADMIN_CLIENT_X509_CERT_URL":        "client_x509_cert_url",
	}

	config := &ServiceAccountConfig{}
	missingVars := []string{}

	for envVar, field := range requiredVars {
		value := os.Getenv(envVar)
		if value == "" {
			missingVars = append(missingVars, envVar)
			continue
		}

		switch field {
		case "type":
			config.Type = value
		case "project_id":
			config.ProjectID = value
		case "private_key_id":
			config.PrivateKeyID = value
		case "private_key":
			config.PrivateKey = value
		case "client_email":
			config.ClientEmail = value
		case "client_id":
			config.ClientID = value
		case "auth_uri":
			config.AuthURI = value
		case "token_uri":
			config.TokenURI = value
		case "auth_provider_x509_cert_url":
			config.AuthProviderX509CertURL = value
		case "client_x509_cert_url":
			config.ClientX509CertURL = value
		}
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return config, nil
}
