package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/rs/zerolog/log"
)

// FirebaseConfiguration holds Firebase-specific configuration
type FirebaseConfiguration struct {
	ProjectID             string
	ServiceAccountKeyPath string
	EmulatorHost          string // For local development with Firebase emulator
}

// userInfoCacheEntry stores cached user info with expiration
type userInfoCacheEntry struct {
	info    *UserInfo
	expires time.Time
}

// FirebaseAuthService implements both Authenticator and UserProvider interfaces
type FirebaseAuthService struct {
	config     FirebaseConfiguration
	authClient *auth.Client
	httpClient *http.Client

	// User info cache
	userInfoCache      map[string]*userInfoCacheEntry
	userInfoCacheMutex sync.RWMutex
}

// NewFirebaseAuthService creates a new Firebase auth service implementing AuthService
func NewFirebaseAuthService(config FirebaseConfiguration) (AuthService, error) {
	ctx := context.Background()

	// Use the new service account configuration system
	_, authClient, err := InitializeFirebaseApp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	service := &FirebaseAuthService{
		config:        config,
		authClient:    authClient,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		userInfoCache: make(map[string]*userInfoCacheEntry),
	}

	// Start cache cleanup goroutine
	go service.startCacheCleanup()

	log.Info().
		Str("project_id", getProjectID()).
		Bool("emulator_mode", isEmulatorMode()).
		Msg("Firebase Auth service initialized successfully")

	return service, nil
}

// ValidateToken validates the Firebase ID token and returns the claims
func (f *FirebaseAuthService) ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error) {
	// Verify the ID token
	token, err := f.authClient.VerifyIDToken(ctx, tokenString)
	if err != nil {
		log.Debug().Err(err).Msg("Firebase token validation failed")
		return nil, fmt.Errorf("invalid Firebase token: %v", err)
	}

	// Extract claims from Firebase token
	name := ""
	if nameVal, ok := token.Claims["name"].(string); ok {
		name = nameVal
	}

	email := ""
	if emailVal, ok := token.Claims["email"].(string); ok {
		email = emailVal
	}

	picture := ""
	if pictureVal, ok := token.Claims["picture"].(string); ok {
		picture = pictureVal
	}

	// Extract provider information
	provider := "firebase"
	if providerData, ok := token.Claims["firebase"].(map[string]interface{}); ok {
		if identities, ok := providerData["identities"].(map[string]interface{}); ok {
			// Get the first provider from identities
			for providerKey := range identities {
				if providerKey != "email" { // Skip email provider, prefer social providers
					provider = providerKey
					break
				}
			}
		}
	}

	// Create auth claims from Firebase token
	authClaims := &AuthClaims{
		UserID:   token.UID,
		Email:    email,
		Name:     name,
		Picture:  picture,
		Provider: provider,
	}

	return authClaims, nil
}

// GetUserInfoFromToken retrieves user information directly from the Firebase token
func (f *FirebaseAuthService) GetUserInfoFromToken(ctx context.Context, token string) (*UserInfo, error) {
	// Extract claims from token
	claims, err := f.ValidateToken(ctx, token)
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

	// Create user info from token claims
	result := &UserInfo{
		ID:        claims.UserID,
		Email:     claims.Email,
		Name:      claims.Name,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   claims.Picture,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
		Provider:  claims.Provider,
	}

	return result, nil
}

// GetUserInfo retrieves user information from cache or Firebase Auth
func (f *FirebaseAuthService) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	// First check service-level cache
	if cachedInfo, ok := f.getCachedUserInfo(userID); ok {
		log.Debug().Str("user_id", userID).Msg("Using cached user info from service cache")
		return cachedInfo, nil
	}

	// Try to get user record from Firebase Auth
	userRecord, err := f.authClient.GetUser(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Failed to get user info from Firebase")

		// Return minimal info when we can't fetch from Firebase
		result := &UserInfo{
			ID:        userID,
			Provider:  "firebase",
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		f.setCachedUserInfo(userID, result)
		return result, nil
	}

	// Parse name into components
	firstName, lastName := "", ""
	if userRecord.DisplayName != "" {
		nameParts := strings.SplitN(userRecord.DisplayName, " ", 2)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
			if len(nameParts) > 1 {
				lastName = nameParts[1]
			}
		}
	}

	// Extract provider from user record
	provider := "firebase"
	if len(userRecord.ProviderUserInfo) > 0 {
		// Use the first provider (prioritize social providers)
		for _, providerInfo := range userRecord.ProviderUserInfo {
			if providerInfo.ProviderID != "password" && providerInfo.ProviderID != "firebase" {
				provider = providerInfo.ProviderID
				break
			}
		}
	}

	result := &UserInfo{
		ID:        userRecord.UID,
		Email:     userRecord.Email,
		Name:      userRecord.DisplayName,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   userRecord.PhotoURL,
		CreatedAt: time.Unix(userRecord.UserMetadata.CreationTimestamp/1000, 0).Format(time.RFC3339),
		UpdatedAt: time.Unix(userRecord.UserMetadata.LastLogInTimestamp/1000, 0).Format(time.RFC3339),
		Provider:  provider,
	}

	// Cache the result
	f.setCachedUserInfo(userID, result)
	return result, nil
}

// GetAuthMiddleware returns the Firebase auth client for middleware integration
func (f *FirebaseAuthService) GetAuthMiddleware() interface{} {
	return f.authClient
}

// GetUserInfoFromAPI is not needed for Firebase as GetUserInfo already uses the Firebase API
func (f *FirebaseAuthService) GetUserInfoFromAPI(ctx context.Context, token string) (*UserInfo, error) {
	// For Firebase, we can use the token directly to get user info
	return f.GetUserInfoFromToken(ctx, token)
}

// UpdateUserPassword is not implemented for Firebase (handled client-side)
func (f *FirebaseAuthService) UpdateUserPassword(ctx context.Context, userID string, newPassword string) error {
	log.Warn().Str("user_id", userID).Msg("UpdateUserPassword called but not implemented - should be handled client-side")
	return nil
}

// UpdateUserProfile is not implemented for Firebase (handled client-side)
func (f *FirebaseAuthService) UpdateUserProfile(ctx context.Context, userID string, profile UserProfile) error {
	log.Warn().
		Str("user_id", userID).
		Str("name", profile.Name).
		Msg("UpdateUserProfile called but not implemented - should be handled client-side")
	return nil
}

// startCacheCleanup runs a periodic cleanup of expired cache entries
func (f *FirebaseAuthService) startCacheCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		f.cleanupExpiredCache()
	}
}

// cleanupExpiredCache removes expired entries from the user info cache
func (f *FirebaseAuthService) cleanupExpiredCache() {
	f.userInfoCacheMutex.Lock()
	defer f.userInfoCacheMutex.Unlock()

	now := time.Now()
	expired := 0

	for userID, entry := range f.userInfoCache {
		if now.After(entry.expires) {
			delete(f.userInfoCache, userID)
			expired++
		}
	}

	if expired > 0 {
		log.Info().Int("removed_entries", expired).Msg("Cleaned expired user info cache entries")
	}
}

// getCachedUserInfo retrieves user info from service-level cache
func (f *FirebaseAuthService) getCachedUserInfo(userID string) (*UserInfo, bool) {
	f.userInfoCacheMutex.RLock()
	defer f.userInfoCacheMutex.RUnlock()

	entry, exists := f.userInfoCache[userID]
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
func (f *FirebaseAuthService) setCachedUserInfo(userID string, info *UserInfo) {
	f.userInfoCacheMutex.Lock()
	defer f.userInfoCacheMutex.Unlock()

	// Cache for 24 hours
	f.userInfoCache[userID] = &userInfoCacheEntry{
		info:    info,
		expires: time.Now().Add(24 * time.Hour),
	}
}
