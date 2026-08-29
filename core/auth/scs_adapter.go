package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

const (
	scsUserIDKey  = "user_id"
	scsEmailKey   = "email"
	scsVersionKey = "session_version"
)

// ErrNoSession is returned when Load/LoadFromCookie finds no user id.
var ErrNoSession = errors.New("no session")

// SCSSessions wraps *scs.SessionManager. Store is injected with the manager.
type SCSSessions struct {
	sm *scs.SessionManager
}

var _ Sessions = (*SCSSessions)(nil)

// NewSCSSessions wraps an existing SCS manager. sm must be non-nil.
func NewSCSSessions(sm *scs.SessionManager) (*SCSSessions, error) {
	if sm == nil {
		return nil, errors.New("nil SessionManager")
	}
	return &SCSSessions{sm: sm}, nil
}

func (s *SCSSessions) loadCtx(ctx context.Context, r *http.Request) (context.Context, error) {
	var token string
	if c, err := r.Cookie(s.sm.Cookie.Name); err == nil {
		token = c.Value
	}
	return s.sm.Load(ctx, token)
}

func (s *SCSSessions) Issue(ctx context.Context, w http.ResponseWriter, r *http.Request, sess Session) error {
	ctx, err := s.loadCtx(ctx, r)
	if err != nil {
		return err
	}
	s.sm.Put(ctx, scsUserIDKey, sess.UserID)
	s.sm.Put(ctx, scsEmailKey, sess.Email)
	ver := sess.SessionVersion
	if ver <= 0 {
		ver = 1
	}
	s.sm.Put(ctx, scsVersionKey, ver)
	if !sess.ExpiresAt.IsZero() {
		s.sm.SetDeadline(ctx, sess.ExpiresAt)
	}
	token, expiry, err := s.sm.Commit(ctx)
	if err != nil {
		return err
	}
	s.sm.WriteSessionCookie(ctx, w, token, expiry)
	return nil
}

func (s *SCSSessions) Load(ctx context.Context, r *http.Request) (Session, error) {
	return s.LoadFromCookie(ctx, r)
}

func (s *SCSSessions) LoadFromCookie(ctx context.Context, r *http.Request) (Session, error) {
	ctx, err := s.loadCtx(ctx, r)
	if err != nil {
		return Session{}, err
	}
	userID := s.sm.GetString(ctx, scsUserIDKey)
	if userID == "" {
		return Session{}, ErrNoSession
	}
	return Session{
		UserID:         userID,
		Email:          s.sm.GetString(ctx, scsEmailKey),
		ExpiresAt:      s.sm.Deadline(ctx),
		SessionVersion: s.sm.GetInt(ctx, scsVersionKey),
	}, nil
}

func (s *SCSSessions) Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, err := s.loadCtx(ctx, r)
	if err != nil {
		return err
	}
	if err := s.sm.Destroy(ctx); err != nil {
		return err
	}
	s.sm.WriteSessionCookie(ctx, w, "", time.Time{})
	return nil
}
