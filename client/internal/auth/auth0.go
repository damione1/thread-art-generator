package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Auth0Service handles authentication with Auth0
type Auth0Service struct {
	config         *Config
	sessionManager *SessionManager
	clientFactory  ClientFactory // Use interface instead of concrete type
}

// NewAuth0Service creates a new Auth0 service
func NewAuth0Service(config *Config, sessionManager *SessionManager, clientFactory ClientFactory) *Auth0Service {
	return &Auth0Service{
		config:         config,
		sessionManager: sessionManager,
		clientFactory:  clientFactory,
	}
}

// LoginHandler redirects the user to Auth0 login page
func (a *Auth0Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate state parameter for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to generate random state")
		return
	}

	// Get environment for cookie configuration
	environment := os.Getenv("ENVIRONMENT")
	cookieDomain := ""
	if environment == "production" || environment == "staging" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}

	// Store state in cookie for verification
	stateCookie := &http.Cookie{
		Name:     "auth_state",
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
		Secure:   environment != "development",
		Path:     "/",
		Domain:   cookieDomain,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, stateCookie)

	// Build Auth0 authorization URL
	authURL := fmt.Sprintf("https://%s/authorize", a.config.Domain)
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", a.config.ClientID)
	params.Set("redirect_uri", a.config.CallbackURL)
	params.Set("scope", "openid profile email offline_access")
	params.Set("state", state)
	params.Set("audience", a.config.Audience)

	redirectURL := authURL + "?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// CallbackHandler processes the callback from Auth0
func (a *Auth0Service) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("request_url", r.URL.String()).
		Str("referer", r.Header.Get("Referer")).
		Str("config_callback", a.config.CallbackURL).
		Msg("Auth0 callback received")

	// Verify state parameter
	stateCookie, err := r.Cookie("auth_state")
	if err != nil {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		log.Error().Err(err).Msg("State cookie not found")
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" || stateParam != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		log.Error().Str("expected", stateCookie.Value).Str("received", stateParam).Msg("State mismatch")
		return
	}

	// Get environment for cookie configuration
	environment := os.Getenv("ENVIRONMENT")
	cookieDomain := ""
	if environment == "production" || environment == "staging" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   environment != "development",
		Path:     "/",
		Domain:   cookieDomain,
		SameSite: http.SameSiteLaxMode,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		log.Error().Msg("No code in callback")
		return
	}

	tokenData, err := a.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to exchange code for token")
		return
	}

	// Create a temporary user info with minimal data from token
	// We'll get complete user info from our API
	minimalUserInfo, err := a.extractMinimalUserInfoFromToken(tokenData.AccessToken)
	if err != nil {
		http.Error(w, "Failed to extract user ID from token", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to extract user ID from token")
		return
	}

	// Create session with minimal data first
	expiresAt := time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second)
	sessionData := &SessionData{
		UserID:       minimalUserInfo.ID,
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		ExpiresAt:    expiresAt,
		UserInfo:     *minimalUserInfo,
	}

	// Create session so we can make an authenticated call to our API
	err = a.sessionManager.CreateSession(w, sessionData)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to create session")
		return
	}

	// Create a context with the token we've just received
	ctx := r.Context()
	ctx = a.clientFactory.AddTokenToContext(ctx, tokenData.AccessToken)

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}

// LogoutHandler logs the user out
func (a *Auth0Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session
	err := a.sessionManager.DestroySession(w, r)
	if err != nil {
		log.Error().Err(err).Msg("Error destroying session")
	}

	// Redirect to Auth0 logout
	logoutURL := fmt.Sprintf("https://%s/v2/logout", a.config.Domain)
	params := url.Values{}
	params.Set("client_id", a.config.ClientID)
	params.Set("returnTo", a.config.LogoutURL)

	redirectURL := logoutURL + "?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// TokenResponse represents the response from Auth0 token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// exchangeCodeForToken exchanges the authorization code for tokens
func (a *Auth0Service) exchangeCodeForToken(code string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth/token", a.config.Domain)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", a.config.ClientID)
	data.Set("client_secret", a.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", a.config.CallbackURL)

	// Create request with form data in the body, not query parameters
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Debug().
			Int("status_code", resp.StatusCode).
			Str("response_body", string(bodyBytes)).
			Msg("Auth0 token exchange failed")
		return nil, fmt.Errorf("failed to get token: %s - %s", resp.Status, string(bodyBytes))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(bodyBytes, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// extractMinimalUserInfoFromToken extracts minimal user info (ID only) from a JWT token
// We only need this to create an initial session before calling our API
func (a *Auth0Service) extractMinimalUserInfoFromToken(accessToken string) (*UserInfo, error) {
	// Split the token into parts
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// Decode the payload (second part)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// Parse the payload
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	// Extract minimal info needed for initial session
	userID, _ := payload["sub"].(string)
	if userID == "" {
		return nil, errors.New("could not extract user ID from token")
	}

	return &UserInfo{
		ID: userID,
	}, nil
}

// generateRandomState generates a random state parameter for CSRF protection
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
