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
	Domain                    string
	Audience                  string
	ClientID                  string
	ClientSecret              string
	ManagementApiClientID     string
	ManagementApiClientSecret string
}

// Auth0Service implements both Authenticator and UserProvider interfaces
type Auth0Service struct {
	config     Auth0Configuration
	validator  *validator.Validator
	httpClient *http.Client

	// Token cache
	managementToken    string
	managementTokenExp time.Time
	tokenCacheMutex    sync.RWMutex

	// User info cache
	userInfoCache      map[string]*userInfoCacheEntry
	userInfoCacheMutex sync.RWMutex
}

// userInfoCacheEntry represents a cached user info with expiration
type userInfoCacheEntry struct {
	info    *UserInfo
	expires time.Time
}

// customClaims contains custom data we want from the Auth0 token
type customClaims struct {
	Auth0ID   string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Nickname  string `json:"nickname"`   // Added for GitHub which might use nickname
	GivenName string `json:"given_name"` // Some providers use given_name
	// Raw data mapping for fallback
	RawUserInfo map[string]interface{} `json:"-"`
}

// Validate checks claims and populates missing fields from raw data
func (c *customClaims) Validate(ctx context.Context) error {
	// If name is empty but nickname exists, use nickname
	if c.Name == "" && c.Nickname != "" {
		c.Name = c.Nickname
	}

	// If name is still empty but given_name exists, use that
	if c.Name == "" && c.GivenName != "" {
		c.Name = c.GivenName
	}

	// Add more fallbacks as needed

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

	service := &Auth0Service{
		config:        config,
		validator:     jwtValidator,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		userInfoCache: make(map[string]*userInfoCacheEntry),
	}

	// Start a goroutine to clean up expired cache entries
	go service.startCacheCleanup()

	return service, nil
}

// startCacheCleanup runs a periodic cleanup of expired cache entries
func (a *Auth0Service) startCacheCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		a.cleanupExpiredCache()
	}
}

// cleanupExpiredCache removes expired entries from the user info cache
func (a *Auth0Service) cleanupExpiredCache() {
	a.userInfoCacheMutex.Lock()
	defer a.userInfoCacheMutex.Unlock()

	now := time.Now()
	expired := 0

	for userID, entry := range a.userInfoCache {
		if now.After(entry.expires) {
			delete(a.userInfoCache, userID)
			expired++
		}
	}

	if expired > 0 {
		log.Info().Int("removed_entries", expired).Msg("Cleaned expired user info cache entries")
	}
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

	// Extract provider from Auth0 ID
	parts := strings.Split(customClaims.Auth0ID, "|")
	provider := "auth0"
	if len(parts) > 1 {
		provider = parts[0]
	}

	// For GitHub and other social providers, we might need special handling
	name := customClaims.Name
	if name == "" && provider == "github" {
		// For GitHub, we might have a nickname but no name
		name = customClaims.Nickname
	}

	// Create auth claims directly from token without extra fetching
	authClaims := &AuthClaims{
		UserID:   customClaims.Auth0ID,
		Email:    customClaims.Email,
		Name:     name,
		Picture:  customClaims.Picture,
		Provider: provider,
	}
	return authClaims, nil
}

// GetUserInfoFromToken retrieves user information directly from the token without extra fetching
func (a *Auth0Service) GetUserInfoFromToken(ctx context.Context, token string) (*UserInfo, error) {
	// Extract the auth0 identifier from token claims
	claims, err := a.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}

	// Parse name into first/last name components if possible
	firstName, lastName := "", ""
	if claims.Name != "" {
		nameParts := strings.SplitN(claims.Name, " ", 2)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
			if len(nameParts) > 1 {
				lastName = nameParts[1]
			}
		}
	}

	// Just use what's in the token
	result := &UserInfo{
		ID:        claims.UserID,
		Email:     claims.Email,
		Name:      claims.Name,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   claims.Picture,
		CreatedAt: time.Now().Format(time.RFC3339), // Use current time as we don't have this info
		UpdatedAt: time.Now().Format(time.RFC3339),
		Provider:  claims.Provider,
	}

	return result, nil
}

// GetUserInfo retrieves user information from the token and caches it
func (a *Auth0Service) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	// First check service-level cache
	if cachedInfo, ok := a.getCachedUserInfo(userID); ok {
		log.Debug().Str("user_id", userID).Msg("Using cached user info from service cache")
		return cachedInfo, nil
	}

	// Check if we have a token in context
	token, ok := ctx.Value("token").(string)
	if ok && token != "" {
		// Try to use the token directly
		userInfo, err := a.GetUserInfoFromToken(ctx, token)
		if err == nil {
			// Cache the result
			a.setCachedUserInfo(userID, userInfo)
			return userInfo, nil
		}
		// Log error but don't fail - we'll return minimal info
		log.Warn().Err(err).Str("user_id", userID).Msg("Failed to get user info from token")
	}

	// Extract the auth0 identifier
	parts := strings.Split(userID, "|")
	provider := "auth0"
	if len(parts) > 1 {
		provider = parts[0]
	}

	// Return minimal info when we don't have a token
	result := &UserInfo{
		ID:        userID,
		Provider:  provider,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Cache the minimal info too
	a.setCachedUserInfo(userID, result)
	return result, nil
}

// GetAuthMiddleware returns the validator for middleware integration
func (a *Auth0Service) GetAuthMiddleware() interface{} {
	return a.validator
}

// getUserInfoCache implements a simple per-request cache for user info lookups
type userInfoCache struct {
	cache map[string]*UserInfo
	mu    sync.RWMutex
}

// contextKey is a private type for context keys to avoid collisions
type contextKey string

// userInfoCacheKey is the context key for the user info cache
const userInfoCacheKey contextKey = "userInfoCache"

// getCachedUserInfo retrieves user info from service-level cache
func (a *Auth0Service) getCachedUserInfo(userID string) (*UserInfo, bool) {
	a.userInfoCacheMutex.RLock()
	defer a.userInfoCacheMutex.RUnlock()

	entry, exists := a.userInfoCache[userID]
	if !exists {
		return nil, false
	}

	// Check if entry is expired
	if time.Now().After(entry.expires) {
		return nil, false
	}

	return entry.info, true
}

// setCachedUserInfo stores user info in service-level cache with 24-hour expiration
func (a *Auth0Service) setCachedUserInfo(userID string, info *UserInfo) {
	a.userInfoCacheMutex.Lock()
	defer a.userInfoCacheMutex.Unlock()

	// Cache for 24 hours
	a.userInfoCache[userID] = &userInfoCacheEntry{
		info:    info,
		expires: time.Now().Add(24 * time.Hour),
	}
}

func (a *Auth0Service) UpdateUserPassword(ctx context.Context, userID string, newPassword string) error {
	// With SPA authentication, password updates should be handled client-side
	// This method becomes a no-op in this implementation
	log.Warn().Str("user_id", userID).Msg("UpdateUserPassword called but not implemented in SPA mode")
	return nil
}

// UpdateUserProfile updates a user's profile information in Auth0
func (a *Auth0Service) UpdateUserProfile(ctx context.Context, userID string, profile UserProfile) error {
	// With SPA authentication, profile updates should be handled client-side
	// This method becomes a no-op in this implementation
	log.Warn().
		Str("user_id", userID).
		Str("name", profile.Name).
		Msg("UpdateUserProfile called but not implemented in SPA mode")
	return nil
}

// GetUserInfoFromAPI retrieves user information directly from Auth0's userinfo endpoint
func (a *Auth0Service) GetUserInfoFromAPI(ctx context.Context, token string) (*UserInfo, error) {
	// Make request to Auth0 userinfo endpoint
	userInfoURL := fmt.Sprintf("https://%s/userinfo", a.config.Domain)
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	client := a.httpClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Warn().
			Int("status_code", resp.StatusCode).
			Str("response", string(bodyBytes)).
			Msg("Failed to get user info from Auth0")
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	// Parse user info
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	// Extract relevant fields
	info := &UserInfo{
		ID:        getStringMapValue(userInfo, "sub"),
		Email:     getStringMapValue(userInfo, "email"),
		Name:      getStringMapValue(userInfo, "name"),
		FirstName: getStringMapValue(userInfo, "given_name"),
		LastName:  getStringMapValue(userInfo, "family_name"),
		Picture:   getStringMapValue(userInfo, "picture"),
		Provider:  "auth0",
	}

	// For GitHub users, name may be missing but nickname might be available
	if info.Name == "" {
		if nickname := getStringMapValue(userInfo, "nickname"); nickname != "" {
			info.Name = nickname
			log.Debug().
				Str("user_id", info.ID).
				Str("nickname", nickname).
				Msg("Using nickname as name for GitHub user")
		}
	}

	// If no first/last name but we have a name, split it
	if info.FirstName == "" && info.Name != "" {
		parts := strings.SplitN(info.Name, " ", 2)
		if len(parts) > 0 {
			info.FirstName = parts[0]
			if len(parts) > 1 {
				info.LastName = parts[1]
			}
		}
	}

	// Extract provider from ID
	if info.ID != "" {
		parts := strings.Split(info.ID, "|")
		if len(parts) > 1 {
			info.Provider = parts[0]
		}
	}

	// Set created/updated times
	info.CreatedAt = time.Now().Format(time.RFC3339)
	info.UpdatedAt = time.Now().Format(time.RFC3339)

	log.Debug().
		Str("user_id", info.ID).
		Str("name", info.Name).
		Str("email", info.Email).
		Str("provider", info.Provider).
		Msg("Retrieved user info from Auth0 API")

	return info, nil
}

// Helper function to safely extract string values from a map
func getStringMapValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}
