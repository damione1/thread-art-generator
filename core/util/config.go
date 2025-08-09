package util

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

// FirebaseConfig stores Firebase-specific configuration
type FirebaseConfig struct {
	ProjectID                   string `mapstructure:"FIREBASE_PROJECT_ID"`
	EmulatorHost                string `mapstructure:"FIREBASE_AUTH_EMULATOR_HOST"`
	StorageEmulatorHost         string `mapstructure:"FIREBASE_STORAGE_EMULATOR_HOST"`
	StorageEmulatorExternalHost string `mapstructure:"FIREBASE_STORAGE_EMULATOR_EXTERNAL_HOST"`
	GCSEmulatorHost             string `mapstructure:"STORAGE_EMULATOR_HOST"`
	WebAPIKey                   string `mapstructure:"FIREBASE_WEB_API_KEY"`
	AuthDomain                  string `mapstructure:"FIREBASE_AUTH_DOMAIN"`
}

// StorageConfig stores storage provider-specific configuration
type StorageConfig struct {
	Provider string `mapstructure:"STORAGE_PROVIDER"`
	Bucket   string `mapstructure:"STORAGE_BUCKET"`
	Region   string `mapstructure:"STORAGE_REGION"`
}

// StorageServiceConfig stores storage service configuration
type StorageServiceConfig struct {
	Port                string `mapstructure:"STORAGE_SERVICE_PORT"`
	URL                 string `mapstructure:"STORAGE_SERVICE_URL"`
	SignedURLTTLMinutes int    `mapstructure:"SIGNED_URL_TTL_MINUTES"`
	MaxFileSizeMB       int    `mapstructure:"MAX_FILE_SIZE_MB"`
}

// PubSubConfig stores Google Cloud Pub/Sub configuration
type PubSubConfig struct {
	ProjectID    string `mapstructure:"PUBSUB_PROJECT_ID"`
	EmulatorHost string `mapstructure:"PUBSUB_EMULATOR_HOST"`
	TopicPrefix  string `mapstructure:"PUBSUB_TOPIC_PREFIX"`
}

// SessionConfig stores session storage configuration
type SessionConfig struct {
	StorageType  string `mapstructure:"SESSION_STORAGE_TYPE"`
	RedisAddr    string `mapstructure:"REDIS_ADDR"`
	RedisEnabled bool   `mapstructure:"REDIS_ENABLED"`
	CookieDomain string `mapstructure:"COOKIE_DOMAIN"`
}

// PasetoConfig stores PASETO token configuration for BFF → API communication
type PasetoConfig struct {
	SecretKey  string `mapstructure:"PASETO_SECRET_KEY"`
	Issuer     string `mapstructure:"PASETO_ISSUER"`
	TTLMinutes int    `mapstructure:"PASETO_TTL_MINUTES"`
}

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment                  string               `mapstructure:"ENVIRONMENT"`
	GRPCServerPort               string               `mapstructure:"GRPC_SERVER_PORT"`
	HTTPServerPort               string               `mapstructure:"HTTP_SERVER_PORT"`
	FrontendPort                 string               `mapstructure:"FRONTEND_PORT"`
	WorkerPort                   string               `mapstructure:"WORKER_PORT"`
	ApiURL                       string               `mapstructure:"API_URL"`
	TokenSymmetricKey            string               `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	InternalAPIKey               string               `mapstructure:"INTERNAL_API_KEY"`
	EmailSenderName              string               `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress           string               `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword          string               `mapstructure:"EMAIL_SENDER_PASSWORD"`
	PostgresHost                 string               `mapstructure:"POSTGRES_HOST"`
	PostgresUser                 string               `mapstructure:"POSTGRES_USER"`
	PostgresPassword             string               `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDb                   string               `mapstructure:"POSTGRES_DB"`
	PostgresIAMAuth              bool                 `mapstructure:"POSTGRES_IAM_AUTH"`
	DB                           *sql.DB              `mapstructure:"-"`
	AdminEmail                   string               `mapstructure:"ADMIN_EMAIL"`
	GCSBucketName                string               `mapstructure:"GCS_BUCKET_NAME"`
	SendInBlueAPIKey             string               `mapstructure:"SENDINBLUE_API_KEY"`
	FrontendUrl                  string               `mapstructure:"FRONTEND_URL"`
	GoogleApplicationCredentials string               `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
	GoogleCloudProject           string               `mapstructure:"GOOGLE_CLOUD_PROJECT"`
	Firebase                     FirebaseConfig       `mapstructure:",squash"`
	Storage                      StorageConfig        `mapstructure:",squash"`
	StorageService               StorageServiceConfig `mapstructure:",squash"`
	PubSub                       PubSubConfig         `mapstructure:",squash"`
	Session                      SessionConfig        `mapstructure:",squash"`
	Paseto                       PasetoConfig         `mapstructure:",squash"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("MIGRATION_PATH")
	viper.BindEnv("GRPC_SERVER_PORT")
	viper.BindEnv("HTTP_SERVER_PORT")
	viper.BindEnv("FRONTEND_PORT")
	viper.BindEnv("WORKER_PORT")
	viper.BindEnv("API_URL")
	viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	viper.BindEnv("INTERNAL_API_KEY")
	viper.BindEnv("EMAIL_SENDER_NAME")
	viper.BindEnv("EMAIL_SENDER_ADDRESS")
	viper.BindEnv("EMAIL_SENDER_PASSWORD")
	viper.BindEnv("POSTGRES_HOST")
	viper.BindEnv("POSTGRES_USER")
	viper.BindEnv("POSTGRES_PASSWORD")
	viper.BindEnv("POSTGRES_DB")
	viper.BindEnv("POSTGRES_IAM_AUTH")
	viper.BindEnv("ADMIN_EMAIL")
	viper.BindEnv("GCS_BUCKET_NAME")
	viper.BindEnv("SENDINBLUE_API_KEY")
	viper.BindEnv("FRONTEND_URL")
	viper.BindEnv("GOOGLE_APPLICATION_CREDENTIALS")
	viper.BindEnv("GOOGLE_CLOUD_PROJECT")
	// Firebase configuration
	viper.BindEnv("FIREBASE_PROJECT_ID")
	viper.BindEnv("FIREBASE_AUTH_EMULATOR_HOST")
	viper.BindEnv("FIREBASE_STORAGE_EMULATOR_HOST")
	viper.BindEnv("FIREBASE_STORAGE_EMULATOR_EXTERNAL_HOST")
	viper.BindEnv("FIREBASE_WEB_API_KEY")
	viper.BindEnv("FIREBASE_AUTH_DOMAIN")
	
	// Google Cloud Storage emulator host for Admin SDK compatibility
	viper.BindEnv("STORAGE_EMULATOR_HOST")

	// Storage configuration
	viper.BindEnv("STORAGE_PROVIDER")
	viper.BindEnv("STORAGE_BUCKET")
	viper.BindEnv("STORAGE_REGION")

	// Storage service configuration
	viper.BindEnv("STORAGE_SERVICE_PORT")
	viper.BindEnv("STORAGE_SERVICE_URL")
	viper.BindEnv("SIGNED_URL_TTL_MINUTES")
	viper.BindEnv("MAX_FILE_SIZE_MB")

	// Session configuration
	viper.BindEnv("SESSION_STORAGE_TYPE")
	viper.BindEnv("REDIS_ADDR")
	viper.BindEnv("REDIS_ENABLED")
	viper.BindEnv("COOKIE_DOMAIN")

	// Pub/Sub configuration
	viper.BindEnv("PUBSUB_PROJECT_ID")
	viper.BindEnv("PUBSUB_EMULATOR_HOST")
	viper.BindEnv("PUBSUB_TOPIC_PREFIX")

	// PASETO configuration
	viper.BindEnv("PASETO_SECRET_KEY")
	viper.BindEnv("PASETO_ISSUER")
	viper.BindEnv("PASETO_TTL_MINUTES")

	if err = viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for missing values
	config.applyDefaults()

	return config, nil
}

// applyDefaults sets default values for configuration fields that are empty
func (c *Config) applyDefaults() {
	if c.FrontendPort == "" {
		c.FrontendPort = "8080"
	}
	if c.WorkerPort == "" {
		c.WorkerPort = "8080" // Default port for Cloud Run
	}
	if c.PostgresHost == "" {
		c.PostgresHost = "db"
	}
	if c.PostgresUser == "" {
		c.PostgresUser = "postgres"
	}
	if c.PostgresPassword == "" {
		c.PostgresPassword = "postgres"
	}
	if c.PostgresDb == "" {
		c.PostgresDb = "threadmachine"
	}
	// Default to false for local development
	if c.Environment == "development" || c.Environment == "" {
		c.PostgresIAMAuth = false
	}
	if c.ApiURL == "" {
		c.ApiURL = "http://api:9090"
	}
	if c.Firebase.ProjectID == "" {
		c.Firebase.ProjectID = "demo-thread-art-generator"
	}

	// Storage defaults - set default bucket names if not explicitly configured
	// Firebase Storage bucket uses .appspot.com suffix in emulator and production
	if c.Storage.Bucket == "" {
		c.Storage.Bucket = c.Firebase.ProjectID + ".appspot.com"
	}

	// Storage service defaults
	if c.StorageService.Port == "" {
		c.StorageService.Port = "9091"
	}
	if c.StorageService.URL == "" {
		c.StorageService.URL = "http://storage:9091"
	}
	if c.StorageService.SignedURLTTLMinutes <= 0 {
		c.StorageService.SignedURLTTLMinutes = 15 // Default 15 minutes
	}
	if c.StorageService.MaxFileSizeMB <= 0 {
		c.StorageService.MaxFileSizeMB = 10 // Default 10MB
	}

	// Session defaults
	if c.Session.StorageType == "" {
		if c.Session.RedisEnabled {
			c.Session.StorageType = "redis"
		} else {
			c.Session.StorageType = "memory"
		}
	}
	if c.Session.RedisAddr == "" {
		c.Session.RedisAddr = "localhost:6379"
	}

	// PubSub defaults
	if c.PubSub.ProjectID == "" && c.Firebase.ProjectID != "" {
		c.PubSub.ProjectID = c.Firebase.ProjectID
	}
	if c.PubSub.TopicPrefix == "" {
		c.PubSub.TopicPrefix = "thread-art"
	}

	// PASETO defaults
	if c.Paseto.Issuer == "" {
		c.Paseto.Issuer = "thread-art-generator"
	}
	if c.Paseto.TTLMinutes <= 0 {
		c.Paseto.TTLMinutes = 15 // Default 15 minutes
	}
}

// GetPostgresDSN builds the PostgreSQL connection string from configuration
func (c *Config) GetPostgresDSN() string {
	// For Cloud SQL with IAM auth (production/staging)
	if c.PostgresIAMAuth {
		// Use Unix socket for Cloud SQL with IAM authentication
		return fmt.Sprintf("host=/cloudsql/%s user=%s dbname=%s sslmode=disable",
			c.PostgresHost, c.PostgresUser, c.PostgresDb)
	}

	// For local development or password-based connections
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresUser, c.PostgresPassword, c.PostgresDb)
}

// GetFirebaseConfigForFrontend converts the core Firebase config to frontend-compatible format
func (c *Config) GetFirebaseConfigForFrontend() *FirebaseClientConfig {
	// Check if we're in emulator mode
	isEmulator := c.Firebase.EmulatorHost != "" || c.Environment == "development"

	config := &FirebaseClientConfig{
		ProjectID:  c.Firebase.ProjectID,
		APIKey:     c.Firebase.WebAPIKey,
		AuthDomain: c.Firebase.AuthDomain,
		IsEmulator: isEmulator,
	}

	if isEmulator {
		// For emulator, always use localhost for browser access
		config.EmulatorHost = "localhost:9099"
		config.EmulatorUI = "localhost:4000"
		config.APIKey = "demo-api-key" // Emulator doesn't need real API key
		config.ProjectID = "demo-thread-art-generator"
	}

	// Generate authDomain from projectID if not provided
	if config.AuthDomain == "" && config.ProjectID != "" {
		config.AuthDomain = fmt.Sprintf("%s.firebaseapp.com", config.ProjectID)
	}

	return config
}

// IsEmulatorMode returns true if we're running in emulator mode for Firebase services
func (c *Config) IsEmulatorMode() bool {
	return c.Environment == "development" || c.Firebase.EmulatorHost != "" || c.PubSub.EmulatorHost != ""
}

// FirebaseClientConfig represents Firebase configuration for frontend clients
type FirebaseClientConfig struct {
	ProjectID    string `json:"projectId"`
	APIKey       string `json:"apiKey"`
	AuthDomain   string `json:"authDomain"`
	EmulatorHost string `json:"emulatorHost,omitempty"`
	EmulatorUI   string `json:"emulatorUI,omitempty"`
	IsEmulator   bool   `json:"isEmulator"`
}