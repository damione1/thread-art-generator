package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/Damione1/thread-art-generator/cmd/cli/internal/client"
	"github.com/Damione1/thread-art-generator/cmd/cli/internal/commands"
	"github.com/Damione1/thread-art-generator/cmd/cli/internal/config"
)

// CLI represents the command-line interface structure
type CLI struct {
	// Commands
	Login    LoginCmd    `cmd:"" help:"Log in with Auth0"`
	Logout   LogoutCmd   `cmd:"" help:"Log out and clear credentials"`
	Arts     ArtsCmd     `cmd:"" help:"Manage your arts"`
	Status   StatusCmd   `cmd:"" help:"Show connection status"`
	Generate GenerateCmd `cmd:"" help:"Generate a new thread art"`
}

// LoginCmd handles authentication
type LoginCmd struct {
	cmdLogin commands.LoginCmd
}

// Run executes the login command
func (cmd *LoginCmd) Run() error {
	configManager := config.NewManager()
	return cmd.cmdLogin.Run(configManager)
}

// LogoutCmd handles logging out
type LogoutCmd struct {
	cmdLogout commands.LogoutCmd
}

// Run executes the logout command
func (cmd *LogoutCmd) Run() error {
	configManager := config.NewManager()
	return cmd.cmdLogout.Run(configManager)
}

// StatusCmd checks connection status
type StatusCmd struct {
	cmdStatus commands.StatusCmd
}

// Run executes the status command
func (cmd *StatusCmd) Run() error {
	configManager := config.NewManager()
	clientService := client.NewService(configManager)
	return cmd.cmdStatus.Run(clientService)
}

// GenerateCmd generates new thread art
type GenerateCmd struct {
	ArtID  string `arg:"" help:"Art ID to generate thread art for"`
	cmdGen commands.GenerateCmd
}

// Run executes the generate command
func (cmd *GenerateCmd) Run() error {
	configManager := config.NewManager()
	clientService := client.NewService(configManager)

	// Set ArtID from command line args
	cmd.cmdGen.ArtID = cmd.ArtID

	return cmd.cmdGen.Run(clientService)
}

// ArtsCmd is a wrapper for the arts commands
type ArtsCmd struct {
	List   ArtsListCmd   `cmd:"" help:"List all your arts"`
	Get    ArtsGetCmd    `cmd:"" help:"Get a specific art by ID"`
	Create ArtsCreateCmd `cmd:"" help:"Create a new art"`
	Delete ArtsDeleteCmd `cmd:"" help:"Delete an art by ID"`
}

// ArtsListCmd lists all arts
type ArtsListCmd struct {
	PageSize int32 `help:"Number of arts to return" default:"10"`
	cmdList  commands.ArtsListCmd
}

// Run executes the arts list command
func (cmd *ArtsListCmd) Run() error {
	configManager := config.NewManager()
	clientService := client.NewService(configManager)

	// Set PageSize from command line args
	cmd.cmdList.PageSize = cmd.PageSize

	return cmd.cmdList.Run(clientService)
}

// ArtsGetCmd gets a specific art
type ArtsGetCmd struct {
	ID     string `arg:"" help:"Art ID to retrieve"`
	cmdGet commands.ArtsGetCmd
}

// Run executes the arts get command
func (cmd *ArtsGetCmd) Run() error {
	configManager := config.NewManager()
	clientService := client.NewService(configManager)

	// Set ID from command line args
	cmd.cmdGet.ID = cmd.ID

	return cmd.cmdGet.Run(clientService)
}

// ArtsCreateCmd creates a new art
type ArtsCreateCmd struct {
	Title     string `arg:"" help:"Title of the art"`
	cmdCreate commands.ArtsCreateCmd
}

// Run executes the arts create command
func (cmd *ArtsCreateCmd) Run() error {
	configManager := config.NewManager()
	clientService := client.NewService(configManager)

	// Set Title from command line args
	cmd.cmdCreate.Title = cmd.Title

	return cmd.cmdCreate.Run(clientService)
}

// ArtsDeleteCmd deletes an art
type ArtsDeleteCmd struct {
	ID        string `arg:"" help:"Art ID to delete"`
	cmdDelete commands.ArtsDeleteCmd
}

// Run executes the arts delete command
func (cmd *ArtsDeleteCmd) Run() error {
	configManager := config.NewManager()
	clientService := client.NewService(configManager)

	// Set ID from command line args
	cmd.cmdDelete.ID = cmd.ID

	return cmd.cmdDelete.Run(clientService)
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("thread-art-cli"),
		kong.Description("CLI for Thread Art Generator"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Execute command
	err := ctx.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
