package commands

import (
	"fmt"

	"github.com/Damione1/thread-art-generator/cmd/cli/internal/auth"
	"github.com/Damione1/thread-art-generator/cmd/cli/internal/config"
)

// LoginCmd handles authentication
type LoginCmd struct{}

// LogoutCmd handles logging out
type LogoutCmd struct{}

// Run executes the login command
func (cmd *LoginCmd) Run(configManager *config.Manager) error {
	authService := auth.NewService()

	// Perform login
	result, err := authService.Login()
	if err != nil {
		return err
	}

	// Save token info to config
	configManager.Config.AccessToken = result.AccessToken
	configManager.Config.RefreshToken = result.RefreshToken
	configManager.Config.ExpiresAt = result.ExpiresAt

	// Save config
	if err := configManager.Save(); err != nil {
		return err
	}

	fmt.Println("Successfully logged in!")
	return nil
}

// Run executes the logout command
func (cmd *LogoutCmd) Run(configManager *config.Manager) error {
	// Clear config
	if err := configManager.Clear(); err != nil {
		return err
	}

	fmt.Println("Successfully logged out!")
	return nil
}
