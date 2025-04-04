package util

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

// Auth0Config stores Auth0-specific configuration
type Auth0Config struct {
	Domain                    string `mapstructure:"AUTH0_DOMAIN"`
	Audience                  string `mapstructure:"AUTH0_AUDIENCE"`
	ClientID                  string `mapstructure:"AUTH0_CLIENT_ID"`
	ClientSecret              string `mapstructure:"AUTH0_CLIENT_SECRET"`
	ManagementApiClientID     string `mapstructure:"AUTH0_MANAGEMENT_API_CLIENT_ID"`
	ManagementApiClientSecret string `mapstructure:"AUTH0_MANAGEMENT_API_CLIENT_SECRET"`
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

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment         string        `mapstructure:"ENVIRONMENT"`
	GRPCServerPort      string        `mapstructure:"GRPC_SERVER_PORT"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	EmailSenderName     string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress  string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	PostgresUser        string        `mapstructure:"POSTGRES_USER"`
	PostgresPassword    string        `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDb          string        `mapstructure:"POSTGRES_DB"`
	DB                  *sql.DB       `mapstructure:"-"`
	AdminEmail          string        `mapstructure:"ADMIN_EMAIL"`
	GCSBucketName       string        `mapstructure:"GCS_BUCKET_NAME"`
	SendInBlueAPIKey    string        `mapstructure:"SENDINBLUE_API_KEY"`
	FrontendUrl         string        `mapstructure:"FRONTEND_URL"`
	Auth0               Auth0Config   `mapstructure:",squash"`
	Storage             StorageConfig `mapstructure:",squash"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("MIGRATION_PATH")
	viper.BindEnv("GRPC_SERVER_PORT")
	viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	viper.BindEnv("EMAIL_SENDER_NAME")
	viper.BindEnv("EMAIL_SENDER_ADDRESS")
	viper.BindEnv("EMAIL_SENDER_PASSWORD")
	viper.BindEnv("POSTGRES_USER")
	viper.BindEnv("POSTGRES_PASSWORD")
	viper.BindEnv("POSTGRES_DB")
	viper.BindEnv("ADMIN_EMAIL")
	viper.BindEnv("GCS_BUCKET_NAME")
	viper.BindEnv("SENDINBLUE_API_KEY")
	viper.BindEnv("FRONTEND_URL")
	viper.BindEnv("AUTH0_DOMAIN")
	viper.BindEnv("AUTH0_AUDIENCE")
	viper.BindEnv("AUTH0_CLIENT_ID")
	viper.BindEnv("AUTH0_CLIENT_SECRET")
	viper.BindEnv("AUTH0_MANAGEMENT_API_CLIENT_ID")
	viper.BindEnv("AUTH0_MANAGEMENT_API_CLIENT_SECRET")

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

	if err = viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return config, nil
}
