package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"github.com/rs/zerolog/log"
)

// PasetoConfig holds PASETO service configuration
type PasetoConfig struct {
	SecretKey  string
	Issuer     string
	TTLMinutes int
}

// PasetoClaims represents our internal PASETO claims
type PasetoClaims struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Provider string    `json:"provider"`
	Issuer   string    `json:"iss"`
	Subject  string    `json:"sub"`
	Audience string    `json:"aud"`
	IssuedAt time.Time `json:"iat"`
	Expires  time.Time `json:"exp"`
}

// PasetoService handles PASETO token operations for internal BFF → API communication
// PASETO is more secure than JWT - no algorithm confusion, symmetric crypto, built-in expiration
type PasetoService struct {
	pasetoMaker *paseto.V2
	secretKey   []byte
	issuer      string
	ttl         time.Duration
}

// NewPasetoService creates a new PASETO service with symmetric key
func NewPasetoService(config PasetoConfig) (*PasetoService, error) {
	if len(config.SecretKey) != 32 {
		return nil, fmt.Errorf("secret key must be exactly 32 bytes, got %d", len(config.SecretKey))
	}

	ttl := time.Duration(config.TTLMinutes) * time.Minute
	if ttl == 0 {
		ttl = 15 * time.Minute // Default 15 minutes
	}

	return &PasetoService{
		pasetoMaker: paseto.NewV2(),
		secretKey:   []byte(config.SecretKey),
		issuer:      config.Issuer,
		ttl:         ttl,
	}, nil
}

// GenerateToken creates a new PASETO token from auth claims
func (p *PasetoService) GenerateToken(claims *AuthClaims) (string, error) {
	now := time.Now()
	expiresAt := now.Add(p.ttl)
	
	pasetoClaims := &PasetoClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Name:     claims.Name,
		Provider: claims.Provider,
		Issuer:   p.issuer,
		Subject:  claims.UserID,
		Audience: "api-service",
		IssuedAt: now,
		Expires:  expiresAt,
	}

	// Marshal claims to JSON
	payload, err := json.Marshal(pasetoClaims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PASETO claims: %v", err)
	}

	// Create PASETO token (v2.local - symmetric encryption)
	token, err := p.pasetoMaker.Encrypt(p.secretKey, payload, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create PASETO token: %v", err)
	}

	log.Debug().
		Str("user_id", claims.UserID).
		Time("expires_at", expiresAt).
		Msg("Generated new internal PASETO token")

	return token, nil
}

// ValidateToken validates and parses a PASETO token
func (p *PasetoService) ValidateToken(tokenString string) (*AuthClaims, error) {
	// Decrypt PASETO token
	var payload []byte
	err := p.pasetoMaker.Decrypt(tokenString, p.secretKey, &payload, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt PASETO token: %v", err)
	}

	// Unmarshal claims
	var claims PasetoClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal PASETO claims: %v", err)
	}

	// Validate expiration
	if time.Now().After(claims.Expires) {
		return nil, fmt.Errorf("PASETO token has expired")
	}

	// Validate issuer
	if claims.Issuer != p.issuer {
		return nil, fmt.Errorf("invalid token issuer: %s", claims.Issuer)
	}

	// Validate audience
	if claims.Audience != "api-service" {
		return nil, fmt.Errorf("invalid token audience: %s", claims.Audience)
	}

	// Convert to AuthClaims
	authClaims := &AuthClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Name:     claims.Name,
		Provider: claims.Provider,
	}

	return authClaims, nil
}

// GetTokenExpiration returns the expiration time for tokens issued by this service
func (p *PasetoService) GetTokenExpiration() time.Duration {
	return p.ttl
}

// RefreshToken creates a new token with the same claims but updated expiration
func (p *PasetoService) RefreshToken(tokenString string) (string, error) {
	// First validate the existing token to get claims
	claims, err := p.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %v", err)
	}

	// Generate new token with same claims
	return p.GenerateToken(claims)
}