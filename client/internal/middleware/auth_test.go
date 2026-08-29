package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type userPayload struct {
	OK    bool   `json:"ok"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func testRouter(t *testing.T) (*chi.Mux, *auth.SCSSessionManager) {
	t.Helper()
	sm := auth.NewInMemorySessionManager()
	r := chi.NewRouter()
	r.Use(sm.GetSessionManager().LoadAndSave)
	r.Use(SessionAuthMiddleware(sm, nil))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if !ok || user == nil {
			_ = json.NewEncoder(w).Encode(userPayload{OK: false})
			return
		}
		_ = json.NewEncoder(w).Encode(userPayload{OK: true, Name: user.Name, Email: user.Email})
	})
	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, sm.CreateSession(w, r, "user-1", auth.UserInfoFromIdentity(coreauth.Identity{
			UserID:    "user-1",
			Email:     "ada@example.com",
			FirstName: "Ada",
			LastName:  "Lovelace",
		}), 1))
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return r, sm
}

func TestGuestHomeHasNoUser(t *testing.T) {
	r, _ := testRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	var got userPayload
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	require.False(t, got.OK)
}

func TestDashboardRedirectsWhenLoggedOut(t *testing.T) {
	r, _ := testRouter(t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))
	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	require.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestLoginHydratesUserOnHome(t *testing.T) {
	r, _ := testRouter(t)

	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/auth/login", nil))
	require.Equal(t, http.StatusOK, loginRec.Code)
	cookies := loginRec.Result().Cookies()
	require.NotEmpty(t, cookies)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var got userPayload
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	require.True(t, got.OK)
	require.Equal(t, "Ada Lovelace", got.Name)
	require.Equal(t, "ada@example.com", got.Email)
}

func TestShouldSkipAuthRequirement(t *testing.T) {
	require.True(t, shouldSkipAuthRequirement("/"))
	require.True(t, shouldSkipAuthRequirement("/login"))
	require.True(t, shouldSkipAuthRequirement("/logout"))
	require.True(t, shouldSkipAuthRequirement("/forgot-password"))
	require.True(t, shouldSkipAuthRequirement("/auth/login"))
	require.True(t, shouldSkipAuthRequirement("/confirm-email"))
	require.False(t, shouldSkipAuthRequirement("/dashboard"))
}

func TestSessionEpochMismatchLogsOut(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	r := chi.NewRouter()
	r.Use(sm.GetSessionManager().LoadAndSave)
	r.Use(SessionAuthMiddleware(sm, func(context.Context, string) (int, error) {
		return 2, nil
	}))
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, sm.CreateSession(w, r, "user-1", auth.UserInfoFromIdentity(coreauth.Identity{
			UserID: "user-1", Email: "ada@example.com", FirstName: "Ada",
		}), 1))
		w.WriteHeader(http.StatusOK)
	})

	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, httptest.NewRequest(http.MethodPost, "/auth/login", nil))
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(loginRec.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	require.Equal(t, "/login", rec.Header().Get("Location"))
}
