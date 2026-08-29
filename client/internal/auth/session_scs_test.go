package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/stretchr/testify/require"
)

func TestUserInfoFromIdentity(t *testing.T) {
	info := UserInfoFromIdentity(coreauth.Identity{
		UserID:    "user-1",
		Email:     "ada@example.com",
		FirstName: "Ada",
		LastName:  "Lovelace",
	})
	require.Equal(t, "user-1", info.ID)
	require.Equal(t, "ada@example.com", info.Email)
	require.Equal(t, "Ada Lovelace", info.Name)
	require.Equal(t, "Ada", info.FirstName)
	require.Equal(t, "Lovelace", info.LastName)
	require.Contains(t, info.Picture, "gravatar.com/avatar/")
}

func TestUserInfoFromIdentityEmailFallback(t *testing.T) {
	info := UserInfoFromIdentity(coreauth.Identity{
		UserID: "user-1",
		Email:  "ada@example.com",
	})
	require.Equal(t, "ada@example.com", info.Name)
	require.Contains(t, info.Picture, "gravatar.com/avatar/")
}

func TestGetSessionFallsBackToEmailKeys(t *testing.T) {
	sm := NewInMemorySessionManager()
	handler := sm.GetSessionManager().LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			sm.GetSessionManager().Put(r.Context(), sessionKeyUserID, "user-1")
			sm.GetSessionManager().Put(r.Context(), "email", "ada@example.com")
			w.WriteHeader(http.StatusOK)
			return
		}

		session, err := sm.GetSession(r)
		require.NoError(t, err)
		require.Equal(t, "user-1", session.UserID)
		require.Equal(t, "ada@example.com", session.UserInfo.Email)
		require.Equal(t, "ada@example.com", session.UserInfo.Name)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", nil))
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestCreateSessionStoresVersionAndCSRF(t *testing.T) {
	sm := NewInMemorySessionManager()
	var csrf string
	handler := sm.GetSessionManager().LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			require.NoError(t, sm.CreateSession(w, r, "user-1", UserInfoFromIdentity(coreauth.Identity{
				UserID: "user-1", Email: "ada@example.com", FirstName: "Ada",
			}), 4))
			w.WriteHeader(http.StatusOK)
			return
		}
		session, err := sm.GetSession(r)
		require.NoError(t, err)
		require.Equal(t, 4, session.Version)
		var genErr error
		csrf, genErr = sm.EnsureCSRFToken(r)
		require.NoError(t, genErr)
		require.NotEmpty(t, csrf)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", nil))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(rec.Result().Cookies()[0])
	handler.ServeHTTP(httptest.NewRecorder(), req)
	require.NotEmpty(t, csrf)
}

func TestCreateSessionRoundTrip(t *testing.T) {
	sm := NewInMemorySessionManager()
	handler := sm.GetSessionManager().LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			require.NoError(t, sm.CreateSession(w, r, "user-1", UserInfoFromIdentity(coreauth.Identity{
				UserID:    "user-1",
				Email:     "ada@example.com",
				FirstName: "Ada",
				LastName:  "Lovelace",
			}), 1))
			w.WriteHeader(http.StatusOK)
			return
		}

		session, err := sm.GetSession(r)
		require.NoError(t, err)
		require.Equal(t, "Ada Lovelace", session.UserInfo.Name)
		require.Equal(t, "Ada", session.UserInfo.FirstName)
		require.NotEmpty(t, session.UserInfo.Picture)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", nil))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(rec.Result().Cookies()[0])
	handler.ServeHTTP(httptest.NewRecorder(), req)
}
