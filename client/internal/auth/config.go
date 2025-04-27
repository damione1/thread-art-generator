package auth

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds Auth0 configuration values
type Config struct {
	Domain       string `mapstructure:"AUTH0_DOMAIN"`
	ClientID     string `mapstructure:"AUTH0_CLIENT_ID"`
	ClientSecret string `mapstructure:"AUTH0_CLIENT_SECRET"`
	Audience     string `mapstructure:"AUTH0_AUDIENCE"`
	CallbackURL  string `mapstructure:"-"` // Derived field
	LogoutURL    string `mapstructure:"-"` // Derived field
	APIBaseURL   string `mapstructure:"API_URL"`
	FrontendURL  string `mapstructure:"FRONTEND_URL"`
}

// LoadConfig reads configuration from environment variables using Viper
func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	// Bind required environment variables
	viper.BindEnv("AUTH0_DOMAIN")
	viper.BindEnv("AUTH0_CLIENT_ID")
	viper.BindEnv("AUTH0_CLIENT_SECRET")
	viper.BindEnv("AUTH0_AUDIENCE")
	viper.BindEnv("API_URL")
	viper.BindEnv("FRONTEND_URL")

	if err = viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal auth config: %w", err)
	}

	// Set derived fields
	config.CallbackURL = config.FrontendURL + "/auth/callback"
	config.LogoutURL = config.FrontendURL

	return config, nil
}

// NewConfig creates a new Auth0 configuration from environment variables
func NewConfig() *Config {
	config, err := LoadConfig()
	if err != nil {
		// Log error and fall back to os.Getenv for backward compatibility
		return &Config{
			Domain:       viper.GetString("AUTH0_DOMAIN"),
			ClientID:     viper.GetString("AUTH0_CLIENT_ID"),
			ClientSecret: viper.GetString("AUTH0_CLIENT_SECRET"),
			Audience:     viper.GetString("AUTH0_AUDIENCE"),
			CallbackURL:  viper.GetString("FRONTEND_URL") + "/auth/callback",
			LogoutURL:    viper.GetString("FRONTEND_URL"),
			APIBaseURL:   viper.GetString("API_URL"),
		}
	}
	return &config
}
