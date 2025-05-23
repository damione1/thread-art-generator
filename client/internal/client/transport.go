package client

import (
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

// AuthTransport is an http.RoundTripper that adds auth headers from the session
type AuthTransport struct {
	SessionManager *auth.SessionManager
	Base           http.RoundTripper
}

// NewAuthTransport creates a new HTTP transport that adds auth headers
func NewAuthTransport(sessionManager *auth.SessionManager) http.RoundTripper {
	return &AuthTransport{
		SessionManager: sessionManager,
		Base:           http.DefaultTransport,
	}
}

// RoundTrip implements http.RoundTripper
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// First check for token directly in the context via TokenFromContext
	token, ok := auth.TokenFromContext(req.Context())
	if ok && token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	} else if req.Context().Value("session") != nil {
		// Fall back to session from context if available
		session := req.Context().Value("session").(*auth.SessionData)
		if session.AccessToken != "" {
			// Add authorization header with token
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.AccessToken))
		}
	}

	// Add Origin header to prevent CORS errors
	req.Header.Set("Origin", "http://localhost:8080") // Set to your client's origin

	return t.Base.RoundTrip(req)
}
