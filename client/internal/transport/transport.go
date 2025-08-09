package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/rs/zerolog/log"
)

// contextKey type for storing values in context
type contextKey string

const (
	sessionRequestKey contextKey = "session_request"
)

// PasetoAuthTransport is an http.RoundTripper that uses PASETO tokens for API authentication
// This provides secure, stateless authentication for BFF → API communication
type PasetoAuthTransport struct {
	PasetoConverter *auth.PasetoConverter
	Base            http.RoundTripper
}

// NewPasetoAuthTransport creates a new HTTP transport that uses PASETO tokens for authentication
func NewPasetoAuthTransport(pasetoConverter *auth.PasetoConverter) http.RoundTripper {
	return &PasetoAuthTransport{
		PasetoConverter: pasetoConverter,
		Base:            http.DefaultTransport,
	}
}

// RoundTrip implements http.RoundTripper for PASETO-based authentication
func (t *PasetoAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a clone of the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Try to get the original session request from context
	if sessionReq, ok := req.Context().Value(sessionRequestKey).(*http.Request); ok {
		// Use the new fallback method to get any available token
		token, tokenType, err := t.PasetoConverter.GetFallbackToken(sessionReq)
		if err != nil {
			log.Warn().
				Err(err).
				Str("url", req.URL.String()).
				Msg("Failed to get authentication token - API request will be unauthenticated")
		} else if token != "" {
			reqClone.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			log.Debug().
				Str("url", req.URL.String()).
				Str("token_type", tokenType).
				Msg("Successfully added authentication token to API request")
		} else {
			log.Warn().
				Str("url", req.URL.String()).
				Msg("Empty token returned - API request will be unauthenticated")
		}
	} else {
		log.Warn().
			Str("url", req.URL.String()).
			Msg("No session request found in context - API request will be unauthenticated")
	}

	// Add required headers for proper API communication
	if reqClone.Header.Get("Content-Type") == "" {
		reqClone.Header.Set("Content-Type", "application/json")
	}
	
	// Add Origin header to prevent CORS errors
	reqClone.Header.Set("Origin", "http://localhost:8080")
	
	// Add User-Agent for better request identification
	if reqClone.Header.Get("User-Agent") == "" {
		reqClone.Header.Set("User-Agent", "ThreadArtGenerator-BFF/1.0")
	}

	// Log the request for debugging (without sensitive headers)
	log.Debug().
		Str("method", reqClone.Method).
		Str("url", reqClone.URL.String()).
		Bool("has_auth", reqClone.Header.Get("Authorization") != "").
		Msg("Making API request")

	// Perform the request with error handling
	resp, err := t.Base.RoundTrip(reqClone)
	if err != nil {
		// Handle authentication errors specifically
		if sessionReq, ok := req.Context().Value(sessionRequestKey).(*http.Request); ok {
			authErr := t.PasetoConverter.HandleAuthError(sessionReq, err, "transport_roundtrip")
			log.Error().
				Err(authErr).
				Str("method", reqClone.Method).
				Str("url", reqClone.URL.String()).
				Msg("API request failed with authentication context")
		} else {
			log.Error().
				Err(err).
				Str("method", reqClone.Method).
				Str("url", reqClone.URL.String()).
				Msg("API request failed without session context")
		}
		return nil, err
	}

	// Log response status for debugging
	log.Debug().
		Str("method", reqClone.Method).
		Str("url", reqClone.URL.String()).
		Int("status", resp.StatusCode).
		Msg("API request completed")

	// Handle specific HTTP error responses
	if resp.StatusCode == http.StatusUnauthorized {
		log.Warn().
			Str("method", reqClone.Method).
			Str("url", reqClone.URL.String()).
			Msg("API returned 401 Unauthorized - authentication may have failed")
		
		// If we have session context, try to handle the auth error
		if sessionReq, ok := req.Context().Value(sessionRequestKey).(*http.Request); ok {
			authErr := fmt.Errorf("API returned 401 Unauthorized")
			t.PasetoConverter.HandleAuthError(sessionReq, authErr, "api_unauthorized")
		}
	} else if resp.StatusCode >= 500 {
		log.Error().
			Str("method", reqClone.Method).
			Str("url", reqClone.URL.String()).
			Int("status", resp.StatusCode).
			Msg("API returned server error")
	}

	return resp, nil
}

// WithSessionRequest adds the original session request to the context for authentication
func WithSessionRequest(ctx context.Context, sessionReq *http.Request) context.Context {
	return context.WithValue(ctx, sessionRequestKey, sessionReq)
}