package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/rs/zerolog/log"
)

const (
	sessionKeyUserID   = "user_id"
	sessionKeyUserInfo = "user_info"
	sessionKeyCSRF     = "csrf_token"
	sessionKeyVersion  = "session_version"
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
	Version   int       `json:"version"`
}

func NewSCSSessionManager(db *sql.DB) (*SCSSessionManager, error) {
	store := postgresstore.New(db)
	sessionManager := scs.New()
	sessionManager.Store = store
	configureSessionCookie(sessionManager)
	return &SCSSessionManager{sessionManager: sessionManager}, nil
}

// NewInMemorySessionManager is for tests: same cookie keys, no Postgres.
func NewInMemorySessionManager() *SCSSessionManager {
	sessionManager := scs.New()
	configureSessionCookie(sessionManager)
	return &SCSSessionManager{sessionManager: sessionManager}
}

func configureSessionCookie(sessionManager *scs.SessionManager) {
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = !util.IsDevelopment(os.Getenv("ENVIRONMENT"))
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
}

// UserInfoFromIdentity builds the session profile shown in the header menu.
func UserInfoFromIdentity(identity coreauth.Identity) UserInfo {
	info := UserInfo{
		ID:        identity.UserID,
		Email:     identity.Email,
		FirstName: strings.TrimSpace(identity.FirstName),
		LastName:  strings.TrimSpace(identity.LastName),
	}
	info.Name = DisplayName(info.FirstName, info.LastName, info.Email)
	if info.Email != "" {
		g := util.NewGravatarFromEmail(info.Email)
		g.Size = 80
		g.Default = "identicon"
		info.Picture = g.GetURL()
	}
	return info
}

// DisplayName prefers first+last, then email.
func DisplayName(first, last, email string) string {
	full := strings.TrimSpace(strings.TrimSpace(first) + " " + strings.TrimSpace(last))
	if full != "" {
		return full
	}
	return strings.TrimSpace(email)
}

func (s *SCSSessionManager) GetSessionManager() *scs.SessionManager {
	return s.sessionManager
}

func (s *SCSSessionManager) CreateSession(w http.ResponseWriter, r *http.Request, userID string, userInfo SessionUserInfo, sessionVersion int) error {
	if err := s.sessionManager.RenewToken(r.Context()); err != nil {
		return fmt.Errorf("failed to renew session token: %w", err)
	}
	if sessionVersion <= 0 {
		sessionVersion = 1
	}
	s.sessionManager.Put(r.Context(), sessionKeyUserID, userID)
	s.sessionManager.Put(r.Context(), "email", userInfo.Email)
	s.sessionManager.Put(r.Context(), sessionKeyVersion, sessionVersion)

	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %v", err)
	}
	s.sessionManager.Put(r.Context(), sessionKeyUserInfo, string(userInfoJSON))

	log.Info().
		Str("user_id", userID).
		Str("user_email", userInfo.Email).
		Int("session_version", sessionVersion).
		Msg("Created new session")
	return nil
}

func (s *SCSSessionManager) GetSession(r *http.Request) (*SessionData, error) {
	userID := s.sessionManager.GetString(r.Context(), sessionKeyUserID)
	if userID == "" {
		return nil, fmt.Errorf("no active session")
	}

	userInfo := UserInfo{ID: userID}
	userInfoJSON := s.sessionManager.GetString(r.Context(), sessionKeyUserInfo)
	if userInfoJSON != "" {
		if err := json.Unmarshal([]byte(userInfoJSON), &userInfo); err != nil {
			log.Warn().Err(err).Str("user_id", userID).Msg("Invalid user_info in session, falling back to email keys")
			userInfo = UserInfo{ID: userID}
		}
	}
	if userInfo.ID == "" {
		userInfo.ID = userID
	}
	if userInfo.Email == "" {
		userInfo.Email = s.sessionManager.GetString(r.Context(), "email")
	}
	if userInfo.Name == "" {
		userInfo.Name = DisplayName(userInfo.FirstName, userInfo.LastName, userInfo.Email)
	}

	return &SessionData{
		UserID:    userID,
		UserInfo:  userInfo,
		ExpiresAt: time.Now().Add(s.sessionManager.Lifetime),
		Version:   s.sessionManager.GetInt(r.Context(), sessionKeyVersion),
	}, nil
}

func (s *SCSSessionManager) EnsureCSRFToken(r *http.Request) (string, error) {
	if existing := s.sessionManager.GetString(r.Context(), sessionKeyCSRF); existing != "" {
		return existing, nil
	}
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(b[:])
	s.sessionManager.Put(r.Context(), sessionKeyCSRF, token)
	return token, nil
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
