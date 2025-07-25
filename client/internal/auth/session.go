package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/core/util"
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

// SessionStore defines the interface for session storage backends
type SessionStore interface {
	Set(ctx context.Context, sessionID string, data *SessionData) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	Delete(ctx context.Context, sessionID string) error
	Extend(ctx context.Context, sessionID string) error
}

// RedisSessionStore implements SessionStore using Redis
type RedisSessionStore struct {
	redis *redis.Client
}

// MemorySessionStore implements SessionStore using in-memory storage
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
}

type sessionEntry struct {
	data      *SessionData
	expiresAt time.Time
}

// SessionManager handles user sessions
type SessionManager struct {
	store        SessionStore
	secureCookie *securecookie.SecureCookie
}

// NewRedisSessionStore creates a new Redis session store
func NewRedisSessionStore(redisAddr string) (*RedisSessionStore, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Check Redis connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisSessionStore{
		redis: redisClient,
	}, nil
}

// NewMemorySessionStore creates a new in-memory session store
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*sessionEntry),
	}
}

// NewSessionManager creates a new session manager with the appropriate storage backend
func NewSessionManager(config *util.Config, hashKey, blockKey []byte) (*SessionManager, error) {
	var store SessionStore
	var err error

	switch config.Session.StorageType {
	case "redis":
		if !config.Session.RedisEnabled {
			return nil, fmt.Errorf("Redis storage type selected but Redis is disabled")
		}
		store, err = NewRedisSessionStore(config.Session.RedisAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis session store: %w", err)
		}
	case "memory":
		store = NewMemorySessionStore()
	default:
		return nil, fmt.Errorf("unsupported session storage type: %s", config.Session.StorageType)
	}

	sc := securecookie.New(hashKey, blockKey)

	return &SessionManager{
		store:        store,
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

// RedisSessionStore methods

func (r *RedisSessionStore) Set(ctx context.Context, sessionID string, data *SessionData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, sessionPrefix+sessionID, jsonData, sessionExpiration).Err()
}

func (r *RedisSessionStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	data, err := r.redis.Get(ctx, sessionPrefix+sessionID).Result()
	if err != nil {
		return nil, err
	}

	var sessionData SessionData
	err = json.Unmarshal([]byte(data), &sessionData)
	if err != nil {
		return nil, err
	}

	return &sessionData, nil
}

func (r *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	return r.redis.Del(ctx, sessionPrefix+sessionID).Err()
}

func (r *RedisSessionStore) Extend(ctx context.Context, sessionID string) error {
	return r.redis.Expire(ctx, sessionPrefix+sessionID, sessionExpiration).Err()
}

// MemorySessionStore methods

func (m *MemorySessionStore) Set(ctx context.Context, sessionID string, data *SessionData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[sessionID] = &sessionEntry{
		data:      data,
		expiresAt: time.Now().Add(sessionExpiration),
	}
	return nil
}

func (m *MemorySessionStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(entry.expiresAt) {
		delete(m.sessions, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return entry.data, nil
}

func (m *MemorySessionStore) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

func (m *MemorySessionStore) Extend(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	entry.expiresAt = time.Now().Add(sessionExpiration)
	return nil
}

// CreateSession creates a new session for the user
func (sm *SessionManager) CreateSession(w http.ResponseWriter, data *SessionData) error {
	sessionID, err := sm.GenerateSessionID()
	if err != nil {
		return err
	}

	// Store session data using the configured storage backend
	ctx := context.Background()
	err = sm.store.Set(ctx, sessionID, data)
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

	// Get session data using the configured storage backend
	ctx := context.Background()
	sessionData, err := sm.store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if token is expired and needs refresh
	if time.Now().After(sessionData.ExpiresAt) {
		// TODO: Implement token refresh logic
		log.Warn().Msg("Token expired - refresh functionality needed")
	}

	// Extend session expiration on activity
	sm.store.Extend(ctx, sessionID)

	return sessionData, nil
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

	// Update session data using the configured storage backend
	ctx := context.Background()
	err = sm.store.Set(ctx, sessionID, data)
	if err != nil {
		return err
	}

	// Extend expiration
	sm.store.Extend(ctx, sessionID)

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

	// Delete using the configured storage backend
	ctx := context.Background()
	sm.store.Delete(ctx, sessionID)

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
