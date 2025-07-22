package client

import (
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

// FirebaseAuthTransport is an http.RoundTripper that adds Firebase ID tokens from SCS sessions
type FirebaseAuthTransport struct {
	SessionManager *auth.SCSSessionManager
	Base           http.RoundTripper
}

// NewFirebaseAuthTransport creates a new HTTP transport that adds Firebase auth headers
func NewFirebaseAuthTransport(sessionManager *auth.SCSSessionManager) http.RoundTripper {
	return &FirebaseAuthTransport{
		SessionManager: sessionManager,
		Base:           http.DefaultTransport,
	}
}

// RoundTrip implements http.RoundTripper for Firebase authentication
func (t *FirebaseAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get Firebase ID token from SCS session
	idToken := t.SessionManager.GetIDToken(req)
	if idToken != "" {
		// Add authorization header with Firebase ID token
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", idToken))
	}

	// Add Origin header to prevent CORS errors
	req.Header.Set("Origin", "http://localhost:8080")

	return t.Base.RoundTrip(req)
}
