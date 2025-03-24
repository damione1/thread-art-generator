package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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
	config    Auth0Configuration
	validator *validator.Validator
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
		config:    config,
		validator: jwtValidator,
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
	// With SPA authentication, we can create a basic user profile from the token ID
	// This removes the need for the Management API token

	// Extract the auth0 identifier - typically in format "provider|userid"
	parts := strings.Split(userID, "|")
	provider := "auth0"
	if len(parts) > 1 {
		provider = parts[0]
	}

	// Create a basic user profile with the information we have
	return &UserInfo{
		ID:        userID,
		Email:     "", // Can be populated later when user logs in
		Name:      "", // Can be populated later when user logs in
		FirstName: "",
		LastName:  "",
		Picture:   "",
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
		Provider:  provider,
	}, nil
}

// UpdateUserPassword updates a user's password in Auth0
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
