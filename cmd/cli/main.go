package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/util"
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
type LoginCmd struct{}

// LogoutCmd handles logging out
type LogoutCmd struct{}

// ArtsCmd handles art operations
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

// StatusCmd checks connection status
type StatusCmd struct{}

// GenerateCmd generates new thread art
type GenerateCmd struct {
	ArtID string `arg:"" help:"Art ID to generate thread art for"`
}

// ConfigFile represents the structure of our config file
type ConfigFile struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Global variables
var (
	configFilePath = os.ExpandEnv("$HOME/.thread-art-cli.json")
	config         ConfigFile
	oauthConfig    *oauth2.Config
)

func init() {
	// Initialize OAuth2 config for Auth0
	oauthConfig = &oauth2.Config{
		ClientID: os.Getenv("AUTH0_CLIENT_ID"),
		// No client secret needed for SPA flow
		RedirectURL: "http://localhost:8085/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/authorize", os.Getenv("AUTH0_DOMAIN")),
			TokenURL: fmt.Sprintf("https://%s/oauth/token", os.Getenv("AUTH0_DOMAIN")),
		},
		// Match scopes with the web SPA application
		Scopes: []string{"openid", "profile", "email"},
	}

	// Load config if exists
	loadConfig()
}

// loadConfig loads config from file
func loadConfig() {
	file, err := os.Open(configFilePath)
	if err != nil {
		// If file doesn't exist, that's okay
		if os.IsNotExist(err) {
			return
		}
		log.Printf("Warning: could not open config file: %v", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Printf("Warning: could not decode config file: %v", err)
	}
}

// saveConfig saves config to file
func saveConfig() error {
	file, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("could not create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("could not encode config: %v", err)
	}
	return nil
}

// getGRPCClient creates a new gRPC client with auth token
func getGRPCClient() (pb.ArtGeneratorServiceClient, error) {
	// Check token expiration and ask user to log in again if needed
	if config.ExpiresAt.Before(time.Now()) {
		fmt.Println("Token expired. Please log in again.")
		fmt.Println("Run: thread-art-cli login")
		return nil, fmt.Errorf("token expired")
	}

	// Create a connection with the gRPC server
	conn, err := grpc.Dial("tag.local:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to gRPC server: %v", err)
	}

	// Create and return client
	return pb.NewArtGeneratorServiceClient(conn), nil
}

// getAuthContext creates a context with auth metadata
func getAuthContext() (context.Context, error) {
	if config.AccessToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	ctx := context.Background()
	md := metadata.Pairs("authorization", "Bearer "+config.AccessToken)
	return metadata.NewOutgoingContext(ctx, md), nil
}

// Run executes the login command
func (cmd *LoginCmd) Run() error {
	// Generate a random state that's URL safe and won't have encoding issues
	// Use alphanumeric characters only to avoid encoding problems
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Errorf("failed to generate state: %v", err)
	}

	// Convert random bytes to alphanumeric only
	state := make([]byte, 16)
	for i := range b {
		state[i] = letters[int(b[i])%len(letters)]
	}
	stateStr := string(state)

	// Create a channel to receive the authorization code
	codeCh := make(chan string)
	errCh := make(chan error)

	// Start a local server to handle the callback
	srv := &http.Server{Addr: ":8085"}

	// Setup HTTP handlers on DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()

	// Set up a handler for token fragment callback - this is our main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// This page will extract the token from URL fragment and process it directly
		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Auth0 CLI Login</title>
			<style>
				body { font-family: sans-serif; margin: 2em; max-width: 500px; margin: 0 auto; padding: 2em; }
				.error { color: #e53e3e; }
				.success { color: #38a169; }
				pre { background: #f7fafc; padding: 1em; border-radius: 5px; }
			</style>
			<script>
				// Extract token from URL fragment
				function getHashParams() {
					var hashParams = {};
					var e,
						r = /([^&;=]+)=?([^&;]*)/g,
						q = window.location.hash.substring(1);
					while (e = r.exec(q)) {
						hashParams[e[1]] = decodeURIComponent(e[2]);
					}
					return hashParams;
				}

				// Main function
				function handleAuth() {
					var params = getHashParams();
					var expectedState = "%s";

					// Check if we have an access token and state
					if (params.access_token && params.state) {
						// Compare states directly
						if (params.state === expectedState) {
							// Send token to our backend
							var xhr = new XMLHttpRequest();
							xhr.open('POST', '/process-token', true);
							xhr.setRequestHeader('Content-Type', 'application/json');
							xhr.onreadystatechange = function() {
								if (xhr.readyState === 4) {
									if (xhr.status === 200) {
										document.getElementById('result').innerHTML = '<div class="success">Authentication successful! You can close this window.</div>';
									} else {
										document.getElementById('result').innerHTML = '<div class="error">Error from server: ' + xhr.responseText + '</div>';
									}
								}
							};
							xhr.send(JSON.stringify({
								access_token: params.access_token,
								state: params.state
							}));
							document.getElementById('result').innerHTML = 'Processing...';
						} else {
							document.getElementById('result').innerHTML = '<div class="error">Error: Authentication failed. State mismatch.</div>';
						}
					} else if (params.error) {
						document.getElementById('result').innerHTML = '<div class="error">Error: ' + params.error + '<br>' + params.error_description + '</div>';
					} else {
						document.getElementById('result').innerHTML = '<div class="error">No token received. Please try again.</div>';
					}
				}

				// Run when page loads
				window.onload = handleAuth;
			</script>
		</head>
		<body>
			<h2>Auth0 CLI Login</h2>
			<div id="result">Processing authentication...</div>
		</body>
		</html>
		`, stateStr)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Add an endpoint to process the token directly from the frontend
	http.HandleFunc("/process-token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON request
		var tokenRequest struct {
			AccessToken string `json:"access_token"`
			State       string `json:"state"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&tokenRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Failed to parse token")
			errCh <- fmt.Errorf("failed to parse token request: %v", err)
			return
		}

		// Validate state
		if tokenRequest.State != stateStr {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Authentication failed")
			errCh <- fmt.Errorf("invalid state")
			return
		}

		// Handle the token
		codeCh <- tokenRequest.AccessToken
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Authentication successful!")

		// Shutdown server
		go func() {
			time.Sleep(500 * time.Millisecond)
			srv.Shutdown(context.Background())
		}()
	})

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Generate authorization URL
	// Use response_type=token for implicit flow (SPA approach)
	audience := os.Getenv("AUTH0_AUDIENCE")

	// Construct the URL with properly escaped parameters
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=token&scope=%s&redirect_uri=%s&audience=%s&state=%s",
		oauthConfig.Endpoint.AuthURL,
		oauthConfig.ClientID,
		strings.Join(oauthConfig.Scopes, " "),
		oauthConfig.RedirectURL,
		audience,
		stateStr,
	)

	// Open browser
	fmt.Println("Opening browser for authentication...")
	openBrowser(authURL)

	// Wait for code or error
	select {
	case tokenString := <-codeCh:
		// In SPA flow, we get the token directly (not a code to exchange)
		// The token is already the access token

		// Parse expiry from JWT (simplified)
		expiresAt, err := parseJWTExpiry(tokenString)
		if err != nil {
			return err
		}

		// Save token to config
		config.AccessToken = tokenString
		config.RefreshToken = "" // No refresh token in implicit flow
		config.ExpiresAt = expiresAt

		// Save config
		if err := saveConfig(); err != nil {
			return err
		}

		fmt.Println("Successfully logged in!")
		return nil

	case err := <-errCh:
		return err

	case <-time.After(2 * time.Minute):
		return fmt.Errorf("authentication timed out")
	}
}

// Run executes the logout command
func (cmd *LogoutCmd) Run() error {
	// Clear config
	config = ConfigFile{}

	// Save empty config
	if err := saveConfig(); err != nil {
		return err
	}

	fmt.Println("Successfully logged out!")
	return nil
}

// Run executes the status command
func (cmd *StatusCmd) Run() error {
	if config.AccessToken == "" {
		fmt.Println("Not logged in")
		return nil
	}

	if config.ExpiresAt.Before(time.Now()) {
		fmt.Println("Token expired, please log in again")
		return nil
	}

	// Attempt to get current user to verify the token
	client, err := getGRPCClient()
	if err != nil {
		return err
	}

	ctx, err := getAuthContext()
	if err != nil {
		return err
	}

	user, err := client.GetCurrentUser(ctx, &pb.GetCurrentUserRequest{})
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	fmt.Printf("Logged in as %s (ID: %s)\n", user.Name, user.Name)
	fmt.Printf("Token valid until %s\n", config.ExpiresAt.Format(time.RFC1123))
	return nil
}

// Run executes the arts list command
func (cmd *ArtsListCmd) Run() error {
	client, err := getGRPCClient()
	if err != nil {
		return err
	}

	ctx, err := getAuthContext()
	if err != nil {
		return err
	}

	resp, err := client.ListArts(ctx, &pb.ListArtsRequest{
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
func (cmd *ArtsGetCmd) Run() error {
	client, err := getGRPCClient()
	if err != nil {
		return err
	}

	ctx, err := getAuthContext()
	if err != nil {
		return err
	}

	// Get current user to construct resource name
	user, err := client.GetCurrentUser(ctx, &pb.GetCurrentUserRequest{})
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// Construct resource name
	resourceName := fmt.Sprintf("users/%s/arts/%s", util.ExtractUserID(user.Name), cmd.ID)

	art, err := client.GetArt(ctx, &pb.GetArtRequest{
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
func (cmd *ArtsCreateCmd) Run() error {
	client, err := getGRPCClient()
	if err != nil {
		return err
	}

	ctx, err := getAuthContext()
	if err != nil {
		return err
	}

	art, err := client.CreateArt(ctx, &pb.CreateArtRequest{
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
func (cmd *ArtsDeleteCmd) Run() error {
	client, err := getGRPCClient()
	if err != nil {
		return err
	}

	ctx, err := getAuthContext()
	if err != nil {
		return err
	}

	// Get current user to construct resource name
	user, err := client.GetCurrentUser(ctx, &pb.GetCurrentUserRequest{})
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// Construct resource name
	resourceName := fmt.Sprintf("users/%s/arts/%s", util.ExtractUserID(user.Name), cmd.ID)

	_, err = client.DeleteArt(ctx, &pb.DeleteArtRequest{
		Name: resourceName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete art: %v", err)
	}

	fmt.Printf("Art %s deleted successfully\n", cmd.ID)
	return nil
}

// Run executes the generate command
func (cmd *GenerateCmd) Run() error {
	// This will be implemented when the thread art generation API is ready
	fmt.Println("Thread art generation not yet implemented in the CLI")
	return nil
}

// openBrowser opens a browser to the specified URL
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = browser.OpenURL(url)
	}

	if err != nil {
		fmt.Printf("Please open this URL in your browser: %s\n", url)
	}
}

// parseJWTExpiry extracts the expiration time from a JWT token
func parseJWTExpiry(tokenString string) (time.Time, error) {
	// Split the token
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return time.Now().Add(24 * time.Hour), fmt.Errorf("invalid token format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Now().Add(24 * time.Hour), fmt.Errorf("failed to decode token payload")
	}

	// Parse JSON payload
	var claims struct {
		Exp int64 `json:"exp"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Now().Add(24 * time.Hour), fmt.Errorf("failed to parse token payload")
	}

	// If the exp claim is present
	if claims.Exp > 0 {
		return time.Unix(claims.Exp, 0), nil
	}

	// Default 24h expiry as fallback
	return time.Now().Add(24 * time.Hour), nil
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
