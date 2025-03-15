package auth

import (
	"context"
)

// AuthClaims represents common claims that can be extracted from a token
type AuthClaims struct {
	UserID string
}

// UserInfo contains user profile information
type UserInfo struct {
	ID        string
	Email     string
	Name      string
	FirstName string
	LastName  string
	Picture   string
	CreatedAt string
	UpdatedAt string
}

// UserProfile contains updatable user profile information
type UserProfile struct {
	FirstName string
	LastName  string
	Name      string
	Picture   string
}

// Authenticator handles token validation and basic auth operations
type Authenticator interface {
	// ValidateToken validates a token and returns the claims
	ValidateToken(ctx context.Context, token string) (*AuthClaims, error)

	// GetAuthMiddleware should return any middleware-specific data
	GetAuthMiddleware() interface{}
}

// UserProvider handles user profile management operations
type UserProvider interface {
	// GetUserInfo retrieves user information from the auth provider
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)

	// UpdateUserPassword updates a user's password
	UpdateUserPassword(ctx context.Context, userID string, newPassword string) error

	// UpdateUserProfile updates a user's profile information
	UpdateUserProfile(ctx context.Context, userID string, profile UserProfile) error
}

// AuthService combines authentication and user management capabilities
type AuthService interface {
	Authenticator
	UserProvider
}
