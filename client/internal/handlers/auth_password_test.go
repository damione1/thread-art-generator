package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/stretchr/testify/require"
)

type fakeIdentities struct {
	identity coreauth.Identity
	hash     string
	err      error
}

func (f *fakeIdentities) ByEmail(context.Context, string) (coreauth.Identity, string, error) {
	return f.identity, f.hash, f.err
}

func (f *fakeIdentities) Create(context.Context, string, string, string, string) (coreauth.Identity, error) {
	return f.identity, f.err
}

func TestLoginHydratesSessionProfile(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	hash, err := coreauth.Argon2idPasswords{}.Hash("password12")
	require.NoError(t, err)

	h := NewPasswordAuthHandler(&fakeIdentities{
		identity: coreauth.Identity{
			UserID:    "user-1",
			Email:     "ada@example.com",
			FirstName: "Ada",
			LastName:  "Lovelace",
			Kind:      coreauth.PrincipalUser,
		},
		hash: hash,
	}, sm)

	mux := http.NewServeMux()
	mux.Handle("/auth/login", sm.GetSessionManager().LoadAndSave(http.HandlerFunc(h.Login)))
	mux.Handle("/auth/status", sm.GetSessionManager().LoadAndSave(http.HandlerFunc(h.Status)))

	body, _ := json.Marshal(map[string]string{"email": "ada@example.com", "password": "password12"})
	loginRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(loginRec, req)
	require.Equal(t, http.StatusOK, loginRec.Code)

	statusReq := httptest.NewRequest(http.MethodGet, "/auth/status", nil)
	statusReq.AddCookie(loginRec.Result().Cookies()[0])
	statusRec := httptest.NewRecorder()
	mux.ServeHTTP(statusRec, statusReq)
	require.Equal(t, http.StatusOK, statusRec.Code)

	var got AuthSyncResponse
	require.NoError(t, json.NewDecoder(statusRec.Body).Decode(&got))
	require.True(t, got.Success)
	require.NotNil(t, got.User)
	require.Equal(t, "Ada Lovelace", got.User.Name)
	require.Equal(t, "Ada", got.User.FirstName)
	require.Equal(t, "ada@example.com", got.User.Email)
	require.NotEmpty(t, got.User.Picture)
}
