package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"
)

const (
	sessionCookieName = "session_id"
	sessionPrefix     = "session:"
	sessionExpiration = 24 * time.Hour
)

// SessionData holds the data stored in the session
type SessionData struct {
	UserID       string    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserInfo     UserInfo  `json:"user_info"`
}

// UserInfo contains basic user profile information
type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// SessionManager handles user sessions
type SessionManager struct {
	redis        *redis.Client
	secureCookie *securecookie.SecureCookie
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisAddr string, hashKey, blockKey []byte) (*SessionManager, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Check Redis connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	sc := securecookie.New(hashKey, blockKey)

	return &SessionManager{
		redis:        redisClient,
		secureCookie: sc,
	}, nil
}

// GenerateSessionID creates a new random session ID
func (sm *SessionManager) GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session for the user
func (sm *SessionManager) CreateSession(w http.ResponseWriter, data *SessionData) error {
	sessionID, err := sm.GenerateSessionID()
	if err != nil {
		return err
	}

	// Store session data in Redis
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = sm.redis.Set(ctx, sessionPrefix+sessionID, jsonData, sessionExpiration).Err()
	if err != nil {
		return err
	}

	// Set secure cookie with session ID
	encoded, err := sm.secureCookie.Encode(sessionCookieName, sessionID)
	if err != nil {
		return err
	}

	// Get environment for domain configuration
	environment := os.Getenv("ENVIRONMENT")
	cookieDomain := ""
	if environment == "production" || environment == "staging" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		Domain:   cookieDomain,
		HttpOnly: true,
		Secure:   environment != "development", // Only disable for local development
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionExpiration.Seconds()),
	}

	http.SetCookie(w, cookie)
	return nil
}

// GetSession retrieves the user's session
func (sm *SessionManager) GetSession(r *http.Request) (*SessionData, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}

	var sessionID string
	err = sm.secureCookie.Decode(sessionCookieName, cookie.Value, &sessionID)
	if err != nil {
		return nil, err
	}

	// Get session data from Redis
	ctx := context.Background()
	data, err := sm.redis.Get(ctx, sessionPrefix+sessionID).Result()
	if err != nil {
		return nil, err
	}

	var sessionData SessionData
	err = json.Unmarshal([]byte(data), &sessionData)
	if err != nil {
		return nil, err
	}

	// Check if token is expired and needs refresh
	if time.Now().After(sessionData.ExpiresAt) {
		// TODO: Implement token refresh logic
		log.Warn().Msg("Token expired - refresh functionality needed")
	}

	// Extend session expiration on activity
	sm.redis.Expire(ctx, sessionPrefix+sessionID, sessionExpiration)

	return &sessionData, nil
}

// UpdateSession updates an existing session with new data
func (sm *SessionManager) UpdateSession(w http.ResponseWriter, r *http.Request, data *SessionData) error {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return err
	}

	var sessionID string
	err = sm.secureCookie.Decode(sessionCookieName, cookie.Value, &sessionID)
	if err != nil {
		return err
	}

	// Update session data in Redis
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = sm.redis.Set(ctx, sessionPrefix+sessionID, jsonData, sessionExpiration).Err()
	if err != nil {
		return err
	}

	// Extend expiration
	sm.redis.Expire(ctx, sessionPrefix+sessionID, sessionExpiration)

	return nil
}

// DestroySession removes the user's session
func (sm *SessionManager) DestroySession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sessionCookieName)
	if err == http.ErrNoCookie {
		return nil // No session to destroy
	}
	if err != nil {
		return err
	}

	var sessionID string
	err = sm.secureCookie.Decode(sessionCookieName, cookie.Value, &sessionID)
	if err != nil {
		return err
	}

	// Delete from Redis
	ctx := context.Background()
	sm.redis.Del(ctx, sessionPrefix+sessionID)

	// Get environment for domain configuration - same as in CreateSession
	environment := os.Getenv("ENVIRONMENT")
	cookieDomain := ""
	if environment == "production" || environment == "staging" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}

	// Clear the cookie - use same settings as when creating
	expiredCookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain,
		HttpOnly: true,
		Secure:   environment != "development", // Match the setting used in CreateSession
		MaxAge:   -1,
	}

	http.SetCookie(w, expiredCookie)
	return nil
}
