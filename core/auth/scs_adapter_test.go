package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/require"
)

func TestSCSSessionsRoundTrip(t *testing.T) {
	s, err := NewSCSSessions(scs.New())
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, s.Issue(req.Context(), rec, req, Session{
		UserID: "user-uuid",
		Email:  "a@b.c",
	}))

	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(cookies[0])
	got, err := s.LoadFromCookie(req2.Context(), req2)
	require.NoError(t, err)
	require.Equal(t, "user-uuid", got.UserID)
	require.Equal(t, "a@b.c", got.Email)
}

func TestSCSSessionsMissingCookie(t *testing.T) {
	s, err := NewSCSSessions(scs.New())
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err = s.LoadFromCookie(req.Context(), req)
	require.ErrorIs(t, err, ErrNoSession)
}

func TestNewSCSSessionsNil(t *testing.T) {
	_, err := NewSCSSessions(nil)
	require.Error(t, err)
}
