package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/mail"
	"github.com/stretchr/testify/require"
)

type fakeTokens struct {
	raw     string
	userID  string
	payload string
	err     error
}

func (f *fakeTokens) Issue(context.Context, string, coreauth.TokenPurpose, time.Duration) (string, error) {
	return f.raw, f.err
}
func (f *fakeTokens) IssueWithPayload(_ context.Context, _ string, _ coreauth.TokenPurpose, _ time.Duration, payload string) (string, error) {
	f.payload = payload
	return f.raw, f.err
}
func (f *fakeTokens) Consume(context.Context, string, coreauth.TokenPurpose) (string, error) {
	return f.userID, f.err
}
func (f *fakeTokens) ConsumeWithPayload(context.Context, string, coreauth.TokenPurpose) (string, string, error) {
	return f.userID, f.payload, f.err
}

type recordingMailer struct {
	last mail.Message
}

func (r *recordingMailer) Send(_ context.Context, msg mail.Message) error {
	r.last = msg
	return nil
}

type settingsIdentities struct {
	current    coreauth.Identity
	byEmailErr error
}

func (s *settingsIdentities) ByEmail(_ context.Context, email string) (coreauth.Identity, string, error) {
	if email == s.current.Email {
		return s.current, "hash", nil
	}
	if s.byEmailErr != nil {
		return coreauth.Identity{}, "", s.byEmailErr
	}
	return coreauth.Identity{}, "", coreauth.ErrIdentityNotFound
}
func (s *settingsIdentities) ByID(context.Context, string) (coreauth.Identity, string, error) {
	return s.current, "hash", nil
}
func (s *settingsIdentities) Create(context.Context, string, string, string, string) (coreauth.Identity, error) {
	return s.current, nil
}
func (s *settingsIdentities) UpdatePassword(context.Context, string, string) error { return nil }
func (s *settingsIdentities) UpdateProfile(_ context.Context, _, first, last, email string) error {
	s.current.FirstName = first
	s.current.LastName = last
	s.current.Email = email
	return nil
}
func (s *settingsIdentities) SetActive(context.Context, string, bool) error { return nil }
func (s *settingsIdentities) BumpSessionVersion(context.Context, string) (int, error) {
	s.current.SessionVersion++
	return s.current.SessionVersion, nil
}

func TestUpdateProfileDoesNotChangeEmailUntilConfirm(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	ids := &settingsIdentities{current: coreauth.Identity{
		UserID: "user-1", Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace", SessionVersion: 1,
	}}
	tokens := &fakeTokens{raw: "tok"}
	recMail := &recordingMailer{}
	emails := mail.NewEmails(recMail, mail.Address{Email: "noreply@localhost"}, "http://localhost:8080")
	h := NewSettingsHandler(ids, tokens, emails, sm)

	form := url.Values{"first_name": {"Ada"}, "last_name": {"Lovelace"}, "email": {"new@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/settings/profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(middleware.WithUser(req.Context(), &auth.UserInfo{ID: "user-1", Email: "ada@example.com"}))

	rec := httptest.NewRecorder()
	sm.GetSessionManager().LoadAndSave(http.HandlerFunc(h.UpdateProfile)).ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/settings?saved=email", rec.Header().Get("Location"))
	require.Equal(t, "new@example.com", tokens.payload)
	require.Equal(t, "ada@example.com", ids.current.Email)
	require.Contains(t, recMail.last.HTML, "/confirm-email?token=tok")
}

func TestConfirmEmailChangeAppliesPayload(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	ids := &settingsIdentities{current: coreauth.Identity{
		UserID: "user-1", Email: "ada@example.com", FirstName: "Ada", SessionVersion: 1,
	}}
	tokens := &fakeTokens{userID: "user-1", payload: "new@example.com"}
	h := NewSettingsHandler(ids, tokens, mail.NewEmails(&recordingMailer{}, mail.Address{Email: "n@l"}, "http://localhost:8080"), sm)

	req := httptest.NewRequest(http.MethodGet, "/confirm-email?token=abc", nil)
	req = req.WithContext(middleware.WithUser(req.Context(), &auth.UserInfo{ID: "user-1"}))
	rec := httptest.NewRecorder()
	sm.GetSessionManager().LoadAndSave(http.HandlerFunc(h.ConfirmEmailChange)).ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/settings?saved=email-confirmed", rec.Header().Get("Location"))
	require.Equal(t, "new@example.com", ids.current.Email)
	require.Equal(t, 2, ids.current.SessionVersion)
}
