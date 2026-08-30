package util

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

// StorageConfig is the S3-compatible bucket (MinIO local, R2, AWS).
type StorageConfig struct {
	Endpoint       string `mapstructure:"S3_ENDPOINT"`
	Region         string `mapstructure:"S3_REGION"`
	Bucket         string `mapstructure:"S3_BUCKET"`
	AccessKey      string `mapstructure:"S3_ACCESS_KEY"`
	SecretKey      string `mapstructure:"S3_SECRET_KEY"`
	PublicBaseURL  string `mapstructure:"S3_PUBLIC_BASE_URL"`
	ForcePathStyle bool   `mapstructure:"S3_FORCE_PATH_STYLE"`
	UseTLS         bool   `mapstructure:"S3_USE_TLS"`
}

// SMTPConfig is the outbound mail transport. Secrets: SMTP_USERNAME / SMTP_PASSWORD.
type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	FromName string `mapstructure:"SMTP_FROM_NAME"`
	FromAddr string `mapstructure:"SMTP_FROM_ADDRESS"`
	TLSMode  string `mapstructure:"SMTP_TLS_MODE"` // none | starttls | tls
}

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment         string        `mapstructure:"ENVIRONMENT"`
	GRPCServerPort      string        `mapstructure:"GRPC_SERVER_PORT"`
	HTTPServerPort      string        `mapstructure:"HTTP_SERVER_PORT"`
	FrontendPort        string        `mapstructure:"FRONTEND_PORT"`
	ApiURL              string        `mapstructure:"API_URL"`
	EmailSenderName     string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress  string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	PostgresHost        string        `mapstructure:"POSTGRES_HOST"`
	PostgresUser        string        `mapstructure:"POSTGRES_USER"`
	PostgresPassword    string        `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDb          string        `mapstructure:"POSTGRES_DB"`
	DB                  *sql.DB       `mapstructure:"-"`
	AdminEmail          string        `mapstructure:"ADMIN_EMAIL"`
	FrontendUrl         string        `mapstructure:"FRONTEND_URL"`
	Storage             StorageConfig `mapstructure:",squash"`
	SMTP                SMTPConfig    `mapstructure:",squash"`
	ServiceHMACSecret   string        `mapstructure:"SERVICE_HMAC_SECRET"`
	RembgURL            string        `mapstructure:"REMBG_URL"`
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
	viper.BindEnv("EMAIL_SENDER_NAME")
	viper.BindEnv("EMAIL_SENDER_ADDRESS")
	viper.BindEnv("EMAIL_SENDER_PASSWORD")
	viper.BindEnv("POSTGRES_HOST")
	viper.BindEnv("POSTGRES_USER")
	viper.BindEnv("POSTGRES_PASSWORD")
	viper.BindEnv("POSTGRES_DB")
	viper.BindEnv("ADMIN_EMAIL")
	viper.BindEnv("FRONTEND_URL")
	viper.BindEnv("S3_ENDPOINT")
	viper.BindEnv("S3_REGION")
	viper.BindEnv("S3_BUCKET")
	viper.BindEnv("S3_ACCESS_KEY")
	viper.BindEnv("S3_SECRET_KEY")
	viper.BindEnv("S3_PUBLIC_BASE_URL")
	viper.BindEnv("S3_FORCE_PATH_STYLE")
	viper.BindEnv("S3_USE_TLS")
	viper.BindEnv("SERVICE_HMAC_SECRET")
	viper.BindEnv("SMTP_HOST")
	viper.BindEnv("SMTP_PORT")
	viper.BindEnv("SMTP_USERNAME")
	viper.BindEnv("SMTP_PASSWORD")
	viper.BindEnv("SMTP_FROM_NAME")
	viper.BindEnv("SMTP_FROM_ADDRESS")
	viper.BindEnv("SMTP_TLS_MODE")
	viper.BindEnv("REMBG_URL")

	if err = viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	config.applyDefaults()
	return config, nil
}

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
	if c.FrontendUrl == "" {
		c.FrontendUrl = "http://localhost:8080"
	}
	if c.Storage.Bucket == "" {
		c.Storage.Bucket = "thread-art"
	}
	if c.Storage.Region == "" {
		c.Storage.Region = "us-east-1"
	}
	if c.Storage.PublicBaseURL == "" {
		c.Storage.PublicBaseURL = "http://localhost:9000/thread-art"
	}
	if c.SMTP.Port == 0 {
		c.SMTP.Port = 587
	}
	if c.SMTP.TLSMode == "" {
		c.SMTP.TLSMode = "none"
	}
	if c.SMTP.FromName == "" {
		c.SMTP.FromName = c.EmailSenderName
	}
	if c.SMTP.FromName == "" {
		c.SMTP.FromName = "ThreadArt"
	}
	if c.SMTP.FromAddr == "" {
		c.SMTP.FromAddr = c.EmailSenderAddress
	}
	if c.SMTP.FromAddr == "" {
		c.SMTP.FromAddr = "noreply@localhost"
	}
	if c.SMTP.Password == "" {
		c.SMTP.Password = c.EmailSenderPassword
	}
}

// GetPostgresDSN builds the PostgreSQL connection string from configuration
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresUser, c.PostgresPassword, c.PostgresDb)
}
