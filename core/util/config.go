package util

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

// FirebaseConfig stores Firebase-specific configuration
type FirebaseConfig struct {
	ProjectID    string `mapstructure:"FIREBASE_PROJECT_ID"`
	EmulatorHost string `mapstructure:"FIREBASE_AUTH_EMULATOR_HOST"`
	WebAPIKey    string `mapstructure:"FIREBASE_WEB_API_KEY"`
	AuthDomain   string `mapstructure:"FIREBASE_AUTH_DOMAIN"`
}

// StorageConfig stores storage provider-specific configuration
type StorageConfig struct {
	Provider         string `mapstructure:"STORAGE_PROVIDER"`
	Bucket           string `mapstructure:"STORAGE_BUCKET"`
	Region           string `mapstructure:"STORAGE_REGION"`
	InternalEndpoint string `mapstructure:"STORAGE_INTERNAL_ENDPOINT"`
	ExternalEndpoint string `mapstructure:"STORAGE_EXTERNAL_ENDPOINT"`
	UseSSL           bool   `mapstructure:"STORAGE_USE_SSL"`
	ForceExternalSSL bool   `mapstructure:"STORAGE_FORCE_EXTERNAL_SSL"`
	AccessKey        string `mapstructure:"STORAGE_ACCESS_KEY"`
	SecretKey        string `mapstructure:"STORAGE_SECRET_KEY"`
	GCPProjectID     string `mapstructure:"GCP_PROJECT_ID"`
}

// QueueConfig stores queue-specific configuration
type QueueConfig struct {
	URL                   string `mapstructure:"RABBITMQ_URL"`
	User                  string `mapstructure:"RABBITMQ_USER"`
	Password              string `mapstructure:"RABBITMQ_PASSWORD"`
	CompositionProcessing string `mapstructure:"QUEUE_COMPOSITION_PROCESSING"`
}

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment         string         `mapstructure:"ENVIRONMENT"`
	GRPCServerPort      string         `mapstructure:"GRPC_SERVER_PORT"`
	HTTPServerPort      string         `mapstructure:"HTTP_SERVER_PORT"`
	FrontendPort        string         `mapstructure:"FRONTEND_PORT"`
	ApiURL              string         `mapstructure:"API_URL"`
	TokenSymmetricKey   string         `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	InternalAPIKey      string         `mapstructure:"INTERNAL_API_KEY"`
	EmailSenderName     string         `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress  string         `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword string         `mapstructure:"EMAIL_SENDER_PASSWORD"`
	PostgresHost        string         `mapstructure:"POSTGRES_HOST"`
	PostgresUser        string         `mapstructure:"POSTGRES_USER"`
	PostgresPassword    string         `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDb          string         `mapstructure:"POSTGRES_DB"`
	DB                  *sql.DB        `mapstructure:"-"`
	AdminEmail          string         `mapstructure:"ADMIN_EMAIL"`
	GCSBucketName       string         `mapstructure:"GCS_BUCKET_NAME"`
	SendInBlueAPIKey    string         `mapstructure:"SENDINBLUE_API_KEY"`
	FrontendUrl         string         `mapstructure:"FRONTEND_URL"`
	Firebase            FirebaseConfig `mapstructure:",squash"`
	Storage             StorageConfig  `mapstructure:",squash"`
	Queue               QueueConfig    `mapstructure:",squash"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("MIGRATION_PATH")
	viper.BindEnv("GRPC_SERVER_PORT")
	viper.BindEnv("HTTP_SERVER_PORT")
	viper.BindEnv("FRONTEND_PORT")
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
	viper.BindEnv("ADMIN_EMAIL")
	viper.BindEnv("GCS_BUCKET_NAME")
	viper.BindEnv("SENDINBLUE_API_KEY")
	viper.BindEnv("FRONTEND_URL")
	// Firebase configuration
	viper.BindEnv("FIREBASE_PROJECT_ID")
	viper.BindEnv("FIREBASE_AUTH_EMULATOR_HOST")
	viper.BindEnv("FIREBASE_WEB_API_KEY")
	viper.BindEnv("FIREBASE_AUTH_DOMAIN")

	// Storage configuration
	viper.BindEnv("STORAGE_PROVIDER")
	viper.BindEnv("STORAGE_BUCKET")
	viper.BindEnv("STORAGE_REGION")
	viper.BindEnv("STORAGE_INTERNAL_ENDPOINT")
	viper.BindEnv("STORAGE_EXTERNAL_ENDPOINT")
	viper.BindEnv("STORAGE_USE_SSL")
	viper.BindEnv("STORAGE_FORCE_EXTERNAL_SSL")
	viper.BindEnv("STORAGE_ACCESS_KEY")
	viper.BindEnv("STORAGE_SECRET_KEY")
	viper.BindEnv("GCP_PROJECT_ID")

	// Queue configuration
	viper.BindEnv("RABBITMQ_URL")
	viper.BindEnv("RABBITMQ_USER")
	viper.BindEnv("RABBITMQ_PASSWORD")
	viper.BindEnv("QUEUE_COMPOSITION_PROCESSING")

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
	if c.ApiURL == "" {
		c.ApiURL = "http://api:9090"
	}
	if c.Firebase.ProjectID == "" {
		c.Firebase.ProjectID = "demo-thread-art-generator"
	}
}

// GetPostgresDSN builds the PostgreSQL connection string from configuration
func (c *Config) GetPostgresDSN() string {
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

// FirebaseClientConfig represents Firebase configuration for frontend clients
type FirebaseClientConfig struct {
	ProjectID    string `json:"projectId"`
	APIKey       string `json:"apiKey"`
	AuthDomain   string `json:"authDomain"`
	EmulatorHost string `json:"emulatorHost,omitempty"`
	EmulatorUI   string `json:"emulatorUI,omitempty"`
	IsEmulator   bool   `json:"isEmulator"`
}
