package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// FirebaseAuthMiddleware creates authentication middleware using Firebase and SCS sessions
func FirebaseAuthMiddleware(sessionManager *auth.SCSSessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this path should skip authentication requirement
			isPublicPath := shouldSkipAuthRequirement(r.URL.Path)

			// Get request ID for logging
			reqID := middleware.GetReqID(r.Context())

			// Try to get user from session
			userID := sessionManager.GetUserID(r)
			if userID == "" {
				if isPublicPath {
					// Public path with no session - continue without user context
					// Skip logging for health endpoint to reduce noise
					if r.URL.Path != "/health" {
						log.Debug().
							Str("request_id", reqID).
							Str("path", r.URL.Path).
							Msg("Public path accessed without session")
					}
					next.ServeHTTP(w, r)
					return
				}

				log.Debug().
					Str("request_id", reqID).
					Str("path", r.URL.Path).
					Msg("No active session - redirecting to login")

				// Protected path with no session - redirect to login
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			// Get full session data for context
			sessionData, err := sessionManager.GetSession(r)
			if err != nil {
				if isPublicPath {
					// Public path with invalid session - clear session and continue without user context
					log.Debug().
						Err(err).
						Str("request_id", reqID).
						Str("user_id", userID).
						Str("path", r.URL.Path).
						Msg("Invalid session on public path - clearing session")
					sessionManager.DestroySession(w, r)
					next.ServeHTTP(w, r)
					return
				}

				log.Warn().
					Err(err).
					Str("request_id", reqID).
					Str("user_id", userID).
					Msg("Failed to get session data - redirecting to login")

				// Protected path with invalid session - clear and redirect
				sessionManager.DestroySession(w, r)
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			// Add user info to request context using existing pattern
			ctx := context.WithValue(r.Context(), userContextKey{}, &sessionData.UserInfo)

			log.Debug().
				Str("request_id", reqID).
				Str("user_id", userID).
				Str("user_email", sessionData.UserInfo.Email).
				Str("path", r.URL.Path).
				Msg("Request authenticated via session")

			// Continue with authenticated request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// shouldSkipAuthRequirement returns true if the path should skip authentication requirement
// These paths are accessible without login but will still add user context if available
func shouldSkipAuthRequirement(path string) bool {
	// Public paths that don't require authentication
	publicPaths := []string{
		"/", // Home page should be public but show user menu if logged in
		"/login",
		"/signup", // Signup page should also be public
		"/auth/",
		"/health",
		"/favicon.ico",
		"/css/",
		"/js/",
		"/images/",
		"/static/",
		"/gallery", // Gallery should be publicly accessible
		"/about",   // About page should be publicly accessible
	}

	for _, publicPath := range publicPaths {
		// Special case for exact root path match
		if publicPath == "/" && path == "/" {
			return true
		}
		// For other paths, use prefix matching but skip root to avoid matching everything
		if publicPath != "/" && strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}

// APIAuthMiddleware adds Firebase ID token to API requests
func APIAuthMiddleware(sessionManager *auth.SCSSessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to API requests
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			// Get Firebase ID token from session
			idToken := sessionManager.GetIDToken(r)
			if idToken != "" {
				// Add Authorization header for API calls
				r.Header.Set("Authorization", "Bearer "+idToken)

				log.Debug().
					Str("user_id", sessionManager.GetUserID(r)).
					Str("path", r.URL.Path).
					Msg("Added Firebase token to API request")
			}

			next.ServeHTTP(w, r)
		})
	}
}
