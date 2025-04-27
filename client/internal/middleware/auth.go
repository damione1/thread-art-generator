package middleware

import (
	"context"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/rs/zerolog/log"
)

// User is a key for the request context
type userContextKey struct{}

// UserFromContext extracts the user info from context
func UserFromContext(ctx context.Context) (*auth.UserInfo, bool) {
	user, ok := ctx.Value(userContextKey{}).(*auth.UserInfo)
	return user, ok
}

// WithUser adds user info to the context
func WithUser(ctx context.Context, user *auth.UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// RequireAuth middleware requires authentication for protected routes
// but doesn't overwrite user context if it already exists
func RequireAuth(sessionManager *auth.SessionManager, loginPath string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Just check if session exists, don't replace context
			_, err := sessionManager.GetSession(r)
			if err != nil {
				log.Debug().Err(err).Str("path", r.URL.Path).Msg("No valid session, redirecting to login")
				http.Redirect(w, r, loginPath, http.StatusTemporaryRedirect)
				return
			}

			// Continue with existing context (which should already have user info from WithAuthInfo)
			next.ServeHTTP(w, r)
		})
	}
}

// WithAuthInfo middleware adds user info to context if authenticated but doesn't require auth
func WithAuthInfo(sessionManager *auth.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionManager.GetSession(r)
			if err == nil && session != nil {
				// User is authenticated, ensure UserInfo has valid data
				if session.UserInfo.Name == "" {
					session.UserInfo.Name = "User" // Fallback name if empty
				}

				// Add info to context
				ctx := WithUser(r.Context(), &session.UserInfo)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// User not authenticated, continue without user info
			next.ServeHTTP(w, r)
		})
	}
}

// EnrichUser middleware enriches user data from the API when authenticated
func EnrichUser(userService *services.UserService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if ok && user != nil && userService != nil {
				apiUser, err := userService.GetCurrentUser(r.Context(), r)
				if err == nil && apiUser != nil {
					// Create a full name, ensuring it's never empty
					fullName := apiUser.FirstName + " " + apiUser.LastName
					if fullName == " " || fullName == "" {
						// Use email or ID as fallback if name components are empty
						if apiUser.Email != "" {
							fullName = apiUser.Email
						} else if user.Email != "" {
							fullName = user.Email
						} else if apiUser.ID != "" {
							fullName = apiUser.ID
						} else {
							fullName = "User" // Last resort fallback
						}
					}

					// Update user info with API data
					enrichedUser := &auth.UserInfo{
						ID:        apiUser.ID,
						Name:      fullName,
						Email:     apiUser.Email,
						Picture:   apiUser.Avatar,
						FirstName: apiUser.FirstName,
						LastName:  apiUser.LastName,
					}

					// Replace user in context with enriched data
					ctx := WithUser(r.Context(), enrichedUser)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else {
					log.Error().Err(err).Msg("Failed to enrich user data from API")

					// Ensure existing user has valid data even if API call failed
					if user.Name == "" {
						if user.Email != "" {
							user.Name = user.Email
						} else if user.ID != "" {
							user.Name = user.ID
						} else {
							user.Name = "User" // Fallback
						}

						// Update context with the fixed user data
						ctx := WithUser(r.Context(), user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Continue with existing context
			next.ServeHTTP(w, r)
		})
	}
}
