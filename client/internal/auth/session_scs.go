package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/rs/zerolog/log"
)

const (
	sessionKeyUserID   = "user_id"
	sessionKeyUserInfo = "user_info"
)

type SCSSessionManager struct {
	sessionManager *scs.SessionManager
}

type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SessionUserInfo = UserInfo

type SessionData struct {
	UserID    string    `json:"user_id"`
	UserInfo  UserInfo  `json:"user_info"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewSCSSessionManager(db *sql.DB) (*SCSSessionManager, error) {
	store := postgresstore.New(db)
	sessionManager := scs.New()
	sessionManager.Store = store
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = os.Getenv("ENVIRONMENT") != "" && os.Getenv("ENVIRONMENT") != "development"
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	return &SCSSessionManager{sessionManager: sessionManager}, nil
}

func (s *SCSSessionManager) GetSessionManager() *scs.SessionManager {
	return s.sessionManager
}

func (s *SCSSessionManager) CreateSession(w http.ResponseWriter, r *http.Request, userID string, userInfo SessionUserInfo) error {
	if err := s.sessionManager.RenewToken(r.Context()); err != nil {
		return fmt.Errorf("failed to renew session token: %w", err)
	}
	s.sessionManager.Put(r.Context(), sessionKeyUserID, userID)
	s.sessionManager.Put(r.Context(), "email", userInfo.Email)

	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %v", err)
	}
	s.sessionManager.Put(r.Context(), sessionKeyUserInfo, string(userInfoJSON))

	log.Info().
		Str("user_id", userID).
		Str("user_email", userInfo.Email).
		Msg("Created new session")
	return nil
}

func (s *SCSSessionManager) GetSession(r *http.Request) (*SessionData, error) {
	userID := s.sessionManager.GetString(r.Context(), sessionKeyUserID)
	if userID == "" {
		return nil, fmt.Errorf("no active session")
	}

	userInfoJSON := s.sessionManager.GetString(r.Context(), sessionKeyUserInfo)
	if userInfoJSON == "" {
		return nil, fmt.Errorf("no user info in session")
	}

	var userInfo UserInfo
	if err := json.Unmarshal([]byte(userInfoJSON), &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %v", err)
	}

	return &SessionData{
		UserID:    userID,
		UserInfo:  userInfo,
		ExpiresAt: time.Now().Add(s.sessionManager.Lifetime),
	}, nil
}

func (s *SCSSessionManager) GetUserID(r *http.Request) string {
	return s.sessionManager.GetString(r.Context(), sessionKeyUserID)
}

func (s *SCSSessionManager) DestroySession(w http.ResponseWriter, r *http.Request) error {
	userID := s.sessionManager.GetString(r.Context(), sessionKeyUserID)
	if err := s.sessionManager.Destroy(r.Context()); err != nil {
		return fmt.Errorf("failed to destroy session: %v", err)
	}
	log.Info().Str("user_id", userID).Msg("Destroyed session")
	return nil
}

func (s *SCSSessionManager) RenewToken(w http.ResponseWriter, r *http.Request) error {
	return s.sessionManager.RenewToken(r.Context())
}
