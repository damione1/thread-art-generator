package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rs/zerolog/log"
)

// Auth0Configuration holds Auth0-specific configuration
type Auth0Configuration struct {
	Domain       string
	Audience     string
	ClientID     string
	ClientSecret string
}

// Auth0Service implements both Authenticator and UserProvider interfaces
type Auth0Service struct {
	config     Auth0Configuration
	validator  *validator.Validator
	tokenCache *managementTokenCache
}

// managementTokenCache caches the Auth0 management API token
type managementTokenCache struct {
	sync.RWMutex
	token     string
	expiresAt time.Time
}

// customClaims contains custom data we want from the Auth0 token
type customClaims struct {
	Auth0ID string `json:"sub"`
}

// Validate does nothing but is required for the validator interface
func (c customClaims) Validate(ctx context.Context) error {
	return nil
}

// NewAuth0Service creates a new Auth0 service implementing AuthService
func NewAuth0Service(config Auth0Configuration) (AuthService, error) {
	issuerURL, err := url.Parse(fmt.Sprintf("https://%s/", config.Domain))
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{config.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &customClaims{}
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %v", err)
	}

	return &Auth0Service{
		config:     config,
		validator:  jwtValidator,
		tokenCache: &managementTokenCache{},
	}, nil
}

// ValidateToken validates the token and returns the claims
func (a *Auth0Service) ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error) {
	claims, err := a.validator.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	validatedClaims, ok := claims.(*validator.ValidatedClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast claims to ValidatedClaims")
	}

	customClaims, ok := validatedClaims.CustomClaims.(*customClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast to custom claims")
	}

	return &AuthClaims{
		UserID: customClaims.Auth0ID,
	}, nil
}

// GetAuthMiddleware returns the validator for middleware integration
func (a *Auth0Service) GetAuthMiddleware() interface{} {
	return a.validator
}

// GetUserInfo retrieves user information from Auth0
func (a *Auth0Service) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	token, err := a.getManagementToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get management token: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://%s/api/v2/users/%s", a.config.Domain, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s (status: %d)", string(body), resp.StatusCode)
	}

	var auth0User struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Nickname  string `json:"nickname"`
		Picture   string `json:"picture"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&auth0User); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:        auth0User.UserID,
		Email:     auth0User.Email,
		Name:      auth0User.Name,
		FirstName: auth0User.FirstName,
		LastName:  auth0User.LastName,
		Picture:   auth0User.Picture,
		CreatedAt: auth0User.CreatedAt,
		UpdatedAt: auth0User.UpdatedAt,
	}, nil
}

// UpdateUserPassword updates a user's password in Auth0
func (a *Auth0Service) UpdateUserPassword(ctx context.Context, userID string, newPassword string) error {
	token, err := a.getManagementToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get management token: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://%s/api/v2/users/%s", a.config.Domain, userID)

	payload := strings.NewReader(fmt.Sprintf(`{
		"password": "%s",
		"connection": "Username-Password-Authentication"
	}`, newPassword))

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, apiURL, payload)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update password: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// UpdateUserProfile updates a user's profile information in Auth0
func (a *Auth0Service) UpdateUserProfile(ctx context.Context, userID string, profile UserProfile) error {
	token, err := a.getManagementToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get management token: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://%s/api/v2/users/%s", a.config.Domain, userID)

	// Create payload with only the fields that are present
	payloadMap := make(map[string]string)

	if profile.Name != "" {
		payloadMap["name"] = profile.Name
	}

	if profile.FirstName != "" {
		payloadMap["given_name"] = profile.FirstName
	}

	if profile.LastName != "" {
		payloadMap["family_name"] = profile.LastName
	}

	if profile.Picture != "" {
		payloadMap["picture"] = profile.Picture
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, apiURL, strings.NewReader(string(payloadJSON)))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update profile: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// getManagementToken gets a token for the Auth0 Management API with caching
func (a *Auth0Service) getManagementToken(ctx context.Context) (string, error) {
	// Check cache first
	a.tokenCache.RLock()
	if a.tokenCache.token != "" && time.Now().Before(a.tokenCache.expiresAt) {
		token := a.tokenCache.token
		a.tokenCache.RUnlock()
		return token, nil
	}
	a.tokenCache.RUnlock()

	// No valid token, acquire write lock and fetch a new one
	a.tokenCache.Lock()
	defer a.tokenCache.Unlock()

	// Double-check in case another goroutine refreshed the token while we were waiting
	if a.tokenCache.token != "" && time.Now().Before(a.tokenCache.expiresAt) {
		return a.tokenCache.token, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	tokenURL := fmt.Sprintf("https://%s/oauth/token", a.config.Domain)

	payload := strings.NewReader(fmt.Sprintf(`{
		"client_id": "%s",
		"client_secret": "%s",
		"audience": "https://%s/api/v2/",
		"grant_type": "client_credentials"
	}`, a.config.ClientID, a.config.ClientSecret, a.config.Domain))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get management token: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Cache the token with a bit of buffer (90% of the actual expiry time)
	expiryDuration := time.Duration(float64(result.ExpiresIn)*0.9) * time.Second
	a.tokenCache.token = result.AccessToken
	a.tokenCache.expiresAt = time.Now().Add(expiryDuration)

	log.Info().Msg("Refreshed Auth0 Management API token")

	return result.AccessToken, nil
}
