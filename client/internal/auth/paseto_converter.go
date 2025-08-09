package auth

import (
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/rs/zerolog/log"
)

// PasetoConverter converts Firebase authentication sessions to PASETO tokens for BFF → API communication
// This provides secure, stateless authentication for internal service communication
type PasetoConverter struct {
	pasetoService  *auth.PasetoService
	sessionManager *SCSSessionManager
}

// NewPasetoConverter creates a new converter that transforms Firebase sessions into PASETO tokens
func NewPasetoConverter(pasetoService *auth.PasetoService, sessionManager *SCSSessionManager) *PasetoConverter {
	return &PasetoConverter{
		pasetoService:  pasetoService,
		sessionManager: sessionManager,
	}
}

// GetSessionManager returns the underlying session manager for fallback operations
func (pc *PasetoConverter) GetSessionManager() *SCSSessionManager {
	return pc.sessionManager
}

// ConvertSessionToPasetoToken extracts Firebase claims from session and converts to PASETO token
func (pc *PasetoConverter) ConvertSessionToPasetoToken(req *http.Request) (string, error) {
	// Get session data which contains user claims
	sessionData, err := pc.sessionManager.GetSession(req)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get session data for PASETO token generation")
		return "", fmt.Errorf("no valid session found: %v", err)
	}

	// Extract user claims from session
	userID := sessionData.UserID
	email := sessionData.UserInfo.Email
	name := sessionData.UserInfo.Name
	provider := "firebase" // Firebase provider

	// Validate required fields
	if userID == "" {
		log.Warn().Msg("Session missing user_id - cannot generate PASETO token")
		return "", fmt.Errorf("no user_id found in session")
	}
	if email == "" {
		log.Warn().Str("user_id", userID).Msg("Session missing user_email - cannot generate PASETO token")
		return "", fmt.Errorf("no user_email found in session")
	}

	// Create auth claims from session data
	claims := &auth.AuthClaims{
		UserID:   userID,
		Email:    email,
		Name:     name,
		Provider: provider,
	}

	// Generate PASETO token for internal API communication
	token, err := pc.pasetoService.GenerateToken(claims)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("email", email).
			Msg("Failed to generate PASETO token from session")
		return "", fmt.Errorf("failed to generate PASETO token: %v", err)
	}

	log.Debug().
		Str("user_id", userID).
		Str("email", email).
		Msg("Successfully converted Firebase session to PASETO token")

	return token, nil
}

// ValidateSessionHasClaims checks if the session contains the required user claims
func (pc *PasetoConverter) ValidateSessionHasClaims(req *http.Request) bool {
	sessionData, err := pc.sessionManager.GetSession(req)
	if err != nil {
		log.Debug().Err(err).Msg("Session validation failed")
		return false
	}

	hasRequiredClaims := sessionData.UserID != "" && sessionData.UserInfo.Email != ""
	if !hasRequiredClaims {
		log.Debug().
			Bool("has_user_id", sessionData.UserID != "").
			Bool("has_email", sessionData.UserInfo.Email != "").
			Msg("Session missing required claims")
	}

	return hasRequiredClaims
}

// RefreshTokenIfNeeded checks if the token needs refreshing and does so if necessary
func (pc *PasetoConverter) RefreshTokenIfNeeded(req *http.Request) error {
	// Check if session is valid
	if !pc.sessionManager.IsSessionValid(req) {
		return fmt.Errorf("session is invalid or expired")
	}

	// For PASETO tokens, we generate them fresh each time, so no refresh needed
	// The session itself handles Firebase token refresh
	return nil
}

// ClearSessionData clears authentication session data
func (pc *PasetoConverter) ClearSessionData(w http.ResponseWriter, req *http.Request) error {
	return pc.sessionManager.DestroySession(w, req)
}

// HandleAuthError provides centralized error handling for authentication failures
// This includes proper logging and fallback strategies
func (pc *PasetoConverter) HandleAuthError(req *http.Request, err error, context string) error {
	userID := pc.sessionManager.GetUserID(req)

	log.Warn().
		Err(err).
		Str("user_id", userID).
		Str("context", context).
		Str("url", req.URL.String()).
		Msg("Authentication error occurred")

	// Check if session is still valid after error
	if userID != "" && !pc.sessionManager.IsSessionValid(req) {
		log.Info().
			Str("user_id", userID).
			Str("context", context).
			Msg("Session invalid after auth error - may need cleanup")
		return fmt.Errorf("session invalid: %v", err)
	}

	return err
}

// GetFallbackToken attempts to get any available authentication token for API calls
// This method tries PASETO first, then falls back to Firebase ID token
func (pc *PasetoConverter) GetFallbackToken(req *http.Request) (string, string, error) {
	// Try PASETO token first
	if pc.ValidateSessionHasClaims(req) {
		token, err := pc.ConvertSessionToPasetoToken(req)
		if err == nil && token != "" {
			return token, "paseto", nil
		}
		log.Debug().Err(err).Msg("PASETO token generation failed, trying Firebase ID token")
	}

	// Fallback to Firebase ID token
	idToken := pc.sessionManager.GetIDToken(req)
	if idToken != "" {
		return idToken, "firebase", nil
	}

	return "", "", fmt.Errorf("no valid authentication tokens available")
}
