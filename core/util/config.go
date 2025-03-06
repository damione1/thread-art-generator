package util

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	HTTPServerPort       string        `mapstructure:"HTTP_SERVER_PORT"`
	GRPCServerPort       string        `mapstructure:"GRPC_SERVER_PORT"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	PostgresUser         string        `mapstructure:"POSTGRES_USER"`
	PostgresPassword     string        `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDb           string        `mapstructure:"POSTGRES_DB"`
	DB                   *sql.DB       `mapstructure:"-"`
	AdminEmail           string        `mapstructure:"ADMIN_EMAIL"`
	GCSBucketName        string        `mapstructure:"GCS_BUCKET_NAME"`
	SendInBlueAPIKey     string        `mapstructure:"SENDINBLUE_API_KEY"`
	FrontendUrl          string        `mapstructure:"FRONTEND_URL"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("MIGRATION_PATH")
	viper.BindEnv("HTTP_SERVER_PORT")
	viper.BindEnv("GRPC_SERVER_PORT")
	viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	viper.BindEnv("ACCESS_TOKEN_DURATION")
	viper.BindEnv("REFRESH_TOKEN_DURATION")
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

	if err = viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return config, nil
}
