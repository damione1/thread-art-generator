package auth

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// AuthClaims represents common claims that can be extracted from a token
type AuthClaims struct {
	UserID string
}

// Authenticator defines the interface for authentication providers
type Authenticator interface {
	// ValidateToken validates a token and returns the claims
	ValidateToken(ctx context.Context, token string) (*AuthClaims, error)

	// GetAuthMiddleware should return any middleware-specific data
	GetAuthMiddleware() interface{}
}

// Auth0Configuration holds Auth0-specific configuration
type Auth0Configuration struct {
	Domain       string
	Audience     string
	ClientID     string
	ClientSecret string
}

// Auth0Authenticator implements the Authenticator interface for Auth0
type Auth0Authenticator struct {
	validator *validator.Validator
}

// NewAuth0Authenticator creates a new Auth0 authenticator
func NewAuth0Authenticator(config Auth0Configuration) (Authenticator, error) {
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

	return &Auth0Authenticator{validator: jwtValidator}, nil
}

// customClaims contains custom data we want from the Auth0 token
type customClaims struct {
	Auth0ID string `json:"sub"`
}

// Validate does nothing but is required for the validator interface
func (c customClaims) Validate(ctx context.Context) error {
	return nil
}

// ValidateToken validates the token and returns the claims
func (a *Auth0Authenticator) ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error) {
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
func (a *Auth0Authenticator) GetAuthMiddleware() interface{} {
	return a.validator
}
