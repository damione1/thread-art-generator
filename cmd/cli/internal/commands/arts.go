package commands

import (
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/cmd/cli/internal/client"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/util"
)

// ArtsCmd is the parent command for arts operations
type ArtsCmd struct {
	List   ArtsListCmd   `cmd:"" help:"List all your arts"`
	Get    ArtsGetCmd    `cmd:"" help:"Get a specific art by ID"`
	Create ArtsCreateCmd `cmd:"" help:"Create a new art"`
	Delete ArtsDeleteCmd `cmd:"" help:"Delete an art by ID"`
}

// ArtsListCmd lists all arts
type ArtsListCmd struct {
	PageSize int32 `help:"Number of arts to return" default:"10"`
}

// ArtsGetCmd gets a specific art
type ArtsGetCmd struct {
	ID string `arg:"" help:"Art ID to retrieve"`
}

// ArtsCreateCmd creates a new art
type ArtsCreateCmd struct {
	Title string `arg:"" help:"Title of the art"`
}

// ArtsDeleteCmd deletes an art
type ArtsDeleteCmd struct {
	ID string `arg:"" help:"Art ID to delete"`
}

// Run executes the arts list command
func (cmd *ArtsListCmd) Run(clientService *client.Service) error {
	grpcClient, err := clientService.GetClient()
	if err != nil {
		return err
	}

	ctx, err := clientService.GetAuthContext()
	if err != nil {
		return err
	}

	resp, err := grpcClient.ListArts(ctx, &pb.ListArtsRequest{
		PageSize: cmd.PageSize,
	})
	if err != nil {
		return fmt.Errorf("failed to list arts: %v", err)
	}

	if len(resp.Arts) == 0 {
		fmt.Println("No arts found")
		return nil
	}

	fmt.Println("Your arts:")
	for i, art := range resp.Arts {
		fmt.Printf("%d. %s (ID: %s, Status: %s)\n", i+1, art.Title, art.Name, art.Status)
	}

	if resp.NextPageToken != "" {
		fmt.Println("\nMore arts available. Use a higher page size to see more.")
	}

	return nil
}

// Run executes the arts get command
func (cmd *ArtsGetCmd) Run(clientService *client.Service) error {
	grpcClient, err := clientService.GetClient()
	if err != nil {
		return err
	}

	ctx, err := clientService.GetAuthContext()
	if err != nil {
		return err
	}

	// Get current user to construct resource name
	user, err := grpcClient.GetCurrentUser(ctx, &pb.GetCurrentUserRequest{})
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// Construct resource name
	resourceName := fmt.Sprintf("users/%s/arts/%s", util.ExtractUserID(user.Name), cmd.ID)

	art, err := grpcClient.GetArt(ctx, &pb.GetArtRequest{
		Name: resourceName,
	})
	if err != nil {
		return fmt.Errorf("failed to get art: %v", err)
	}

	fmt.Printf("Art Details:\n")
	fmt.Printf("  ID: %s\n", art.Name)
	fmt.Printf("  Title: %s\n", art.Title)
	fmt.Printf("  Status: %s\n", art.Status)
	fmt.Printf("  Created At: %s\n", art.CreateTime.AsTime().Format(time.RFC1123))
	if art.ImageUrl != "" {
		fmt.Printf("  Image URL: %s\n", art.ImageUrl)
	}

	return nil
}

// Run executes the arts create command
func (cmd *ArtsCreateCmd) Run(clientService *client.Service) error {
	grpcClient, err := clientService.GetClient()
	if err != nil {
		return err
	}

	ctx, err := clientService.GetAuthContext()
	if err != nil {
		return err
	}

	art, err := grpcClient.CreateArt(ctx, &pb.CreateArtRequest{
		Art: &pb.Art{
			Title: cmd.Title,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create art: %v", err)
	}

	fmt.Printf("Art created successfully!\n")
	fmt.Printf("  ID: %s\n", art.Name)
	fmt.Printf("  Title: %s\n", art.Title)
	fmt.Printf("  Status: %s\n", art.Status)

	return nil
}

// Run executes the arts delete command
func (cmd *ArtsDeleteCmd) Run(clientService *client.Service) error {
	grpcClient, err := clientService.GetClient()
	if err != nil {
		return err
	}

	ctx, err := clientService.GetAuthContext()
	if err != nil {
		return err
	}

	// Get current user to construct resource name
	user, err := grpcClient.GetCurrentUser(ctx, &pb.GetCurrentUserRequest{})
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// Construct resource name
	resourceName := fmt.Sprintf("users/%s/arts/%s", util.ExtractUserID(user.Name), cmd.ID)

	_, err = grpcClient.DeleteArt(ctx, &pb.DeleteArtRequest{
		Name: resourceName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete art: %v", err)
	}

	fmt.Printf("Art %s deleted successfully\n", cmd.ID)
	return nil
}
