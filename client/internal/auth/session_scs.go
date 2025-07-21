package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/rs/zerolog/log"
)

const (
	// Session keys
	sessionKeyUserID   = "user_id"
	sessionKeyUserInfo = "user_info"
	sessionKeyIDToken  = "id_token"
	sessionKeyExpiry   = "token_expiry"
)

// SCSSessionManager handles user sessions using alexedwards/scs
type SCSSessionManager struct {
	sessionManager *scs.SessionManager
}

// SessionUserInfo contains basic user profile information stored in session
type SessionUserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// NewSCSSessionManager creates a new session manager with PostgreSQL store
func NewSCSSessionManager(db *sql.DB) (*SCSSessionManager, error) {
	// Create PostgreSQL store
	store := postgresstore.New(db)

	// Create session manager
	sessionManager := scs.New()
	sessionManager.Store = store
	sessionManager.Lifetime = 24 * time.Hour // 24 hour session lifetime
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = false // Set to true in production with HTTPS
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	// Initialize the store (creates tables if they don't exist)
	// The PostgreSQL store will automatically create tables and handle cleanup

	return &SCSSessionManager{
		sessionManager: sessionManager,
	}, nil
}

// GetSessionManager returns the underlying SCS session manager for middleware
func (s *SCSSessionManager) GetSessionManager() *scs.SessionManager {
	return s.sessionManager
}

// CreateSession creates a new session with Firebase user data
func (s *SCSSessionManager) CreateSession(w http.ResponseWriter, r *http.Request, userID string, userInfo SessionUserInfo, idToken string, tokenExpiry time.Time) error {
	// Store user ID
	s.sessionManager.Put(r.Context(), sessionKeyUserID, userID)

	// Store user info as JSON
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %v", err)
	}
	s.sessionManager.Put(r.Context(), sessionKeyUserInfo, string(userInfoJSON))

	// Store Firebase ID token (for API calls)
	s.sessionManager.Put(r.Context(), sessionKeyIDToken, idToken)

	// Store token expiry
	s.sessionManager.Put(r.Context(), sessionKeyExpiry, tokenExpiry.Unix())

	log.Info().
		Str("user_id", userID).
		Str("user_name", userInfo.Name).
		Str("user_email", userInfo.Email).
		Msg("Created new session")

	return nil
}

// GetSession retrieves the user's session data
func (s *SCSSessionManager) GetSession(r *http.Request) (*SessionData, error) {
	// Check if user is authenticated
	userID := s.sessionManager.GetString(r.Context(), sessionKeyUserID)
	if userID == "" {
		return nil, fmt.Errorf("no active session")
	}

	// Get user info
	userInfoJSON := s.sessionManager.GetString(r.Context(), sessionKeyUserInfo)
	if userInfoJSON == "" {
		return nil, fmt.Errorf("no user info in session")
	}

	var userInfo SessionUserInfo
	if err := json.Unmarshal([]byte(userInfoJSON), &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	// Get token data
	tokenExpiryUnix := s.sessionManager.GetInt64(r.Context(), sessionKeyExpiry)
	tokenExpiry := time.Unix(tokenExpiryUnix, 0)

	// Check if token is expired
	if time.Now().After(tokenExpiry) {
		log.Warn().Str("user_id", userID).Msg("Firebase token expired in session")
		// Note: In a full implementation, we'd implement token refresh here
		// For now, we'll let the API handle token validation
	}

	// Create legacy UserInfo for compatibility
	legacyUserInfo := UserInfo{
		ID:        userInfo.ID,
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Picture:   userInfo.Picture,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
	}

	sessionData := &SessionData{
		UserID:    userID,
		UserInfo:  legacyUserInfo,
		ExpiresAt: tokenExpiry,
		// Note: We're not exposing the actual tokens for security
		AccessToken:  "", // Not used with Firebase
		RefreshToken: "", // Not used with Firebase
	}

	return sessionData, nil
}

// GetUserID returns the user ID from the session
func (s *SCSSessionManager) GetUserID(r *http.Request) string {
	return s.sessionManager.GetString(r.Context(), sessionKeyUserID)
}

// GetIDToken returns the Firebase ID token from the session (for API calls)
func (s *SCSSessionManager) GetIDToken(r *http.Request) string {
	return s.sessionManager.GetString(r.Context(), sessionKeyIDToken)
}

// UpdateSession updates an existing session with new data
func (s *SCSSessionManager) UpdateSession(w http.ResponseWriter, r *http.Request, userInfo SessionUserInfo, idToken string, tokenExpiry time.Time) error {
	// Update user info
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %v", err)
	}
	s.sessionManager.Put(r.Context(), sessionKeyUserInfo, string(userInfoJSON))

	// Update token data
	s.sessionManager.Put(r.Context(), sessionKeyIDToken, idToken)
	s.sessionManager.Put(r.Context(), sessionKeyExpiry, tokenExpiry.Unix())

	userID := s.sessionManager.GetString(r.Context(), sessionKeyUserID)
	log.Info().Str("user_id", userID).Msg("Updated session")

	return nil
}

// DestroySession removes the user's session
func (s *SCSSessionManager) DestroySession(w http.ResponseWriter, r *http.Request) error {
	userID := s.sessionManager.GetString(r.Context(), sessionKeyUserID)

	// Destroy the session
	err := s.sessionManager.Destroy(r.Context())
	if err != nil {
		return fmt.Errorf("failed to destroy session: %v", err)
	}

	log.Info().Str("user_id", userID).Msg("Destroyed session")
	return nil
}

// RenewToken renews the session token (extends lifetime)
func (s *SCSSessionManager) RenewToken(w http.ResponseWriter, r *http.Request) error {
	return s.sessionManager.RenewToken(r.Context())
}
