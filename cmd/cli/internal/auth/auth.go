package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// Service handles authentication operations
type Service struct {
	OAuthConfig *oauth2.Config
}

// NewService creates a new auth service
func NewService() *Service {
	oauthConfig := &oauth2.Config{
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

	return &Service{
		OAuthConfig: oauthConfig,
	}
}

// LoginResult represents the result of the login process
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Login handles the OAuth2 login process
func (s *Service) Login() (*LoginResult, error) {
	// Generate a random state that's URL safe (alphanumeric only)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %v", err)
	}

	// Convert random bytes to alphanumeric only
	state := make([]byte, 16)
	for i := range b {
		state[i] = letters[int(b[i])%len(letters)]
	}
	stateStr := string(state)

	// Create channels for communication
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
		s.OAuthConfig.Endpoint.AuthURL,
		s.OAuthConfig.ClientID,
		strings.Join(s.OAuthConfig.Scopes, " "),
		s.OAuthConfig.RedirectURL,
		audience,
		stateStr,
	)

	// Open browser
	fmt.Println("Opening browser for authentication...")
	err = OpenBrowser(authURL)
	if err != nil {
		fmt.Printf("Please open this URL in your browser: %s\n", authURL)
	}

	// Wait for code or error
	select {
	case tokenString := <-codeCh:
		// In SPA flow, we get the token directly
		expiresAt, err := ParseJWTExpiry(tokenString)
		if err != nil {
			return nil, err
		}

		return &LoginResult{
			AccessToken:  tokenString,
			RefreshToken: "", // No refresh token in implicit flow
			ExpiresAt:    expiresAt,
		}, nil

	case err := <-errCh:
		return nil, err

	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("authentication timed out")
	}
}

// ParseJWTExpiry extracts the expiration time from a JWT token
func ParseJWTExpiry(tokenString string) (time.Time, error) {
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

// OpenBrowser opens a browser to the specified URL
func OpenBrowser(url string) error {
	return browser.OpenURL(url)
}
