package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/middleware"
)

// AuthHandler handles authentication-related routes
type AuthHandler struct {
	auth0Service *auth.Auth0Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(auth0Service *auth.Auth0Service) *AuthHandler {
	return &AuthHandler{
		auth0Service: auth0Service,
	}
}

// Login handles login requests
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.auth0Service.LoginHandler(w, r)
}

// Callback handles Auth0 callback requests
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	h.auth0Service.CallbackHandler(w, r)
}

// Logout handles logout requests
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.auth0Service.LogoutHandler(w, r)
}

// Status returns the user's login status
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	// Check if user is in context (should be set by middleware)
	user, ok := middleware.UserFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if ok && user != nil {
		w.Write([]byte(`{"authenticated": true, "user": {"name": "` + user.Name + `", "email": "` + user.Email + `"}}`))
		return
	}

	w.Write([]byte(`{"authenticated": false}`))
}
