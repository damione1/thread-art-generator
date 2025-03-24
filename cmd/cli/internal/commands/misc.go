package commands

import (
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/cmd/cli/internal/client"
	"github.com/Damione1/thread-art-generator/core/pb"
)

// StatusCmd checks connection status
type StatusCmd struct{}

// GenerateCmd generates new thread art
type GenerateCmd struct {
	ArtID string `arg:"" help:"Art ID to generate thread art for"`
}

// Run executes the status command
func (cmd *StatusCmd) Run(clientService *client.Service) error {
	if !clientService.ConfigManager.IsTokenValid() {
		fmt.Println("Not logged in or token expired")
		return nil
	}

	// Attempt to get current user to verify the token
	grpcClient, err := clientService.GetClient()
	if err != nil {
		return err
	}

	ctx, err := clientService.GetAuthContext()
	if err != nil {
		return err
	}

	user, err := grpcClient.GetCurrentUser(ctx, &pb.GetCurrentUserRequest{})
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	fmt.Printf("Logged in as %s (ID: %s)\n", user.Name, user.Name)
	fmt.Printf("Token valid until %s\n", clientService.ConfigManager.Config.ExpiresAt.Format(time.RFC1123))
	return nil
}

// Run executes the generate command
func (cmd *GenerateCmd) Run(clientService *client.Service) error {
	// This will be implemented when the thread art generation API is ready
	fmt.Println("Thread art generation not yet implemented in the CLI")
	return nil
}
