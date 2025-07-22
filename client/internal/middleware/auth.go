package middleware

import (
	"context"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/core/resource"
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
			// Get session and check if it exists
			session, err := sessionManager.GetSession(r)
			if err != nil {
				log.Debug().Err(err).Str("path", r.URL.Path).Msg("No valid session, redirecting to login")
				http.Redirect(w, r, loginPath, http.StatusTemporaryRedirect)
				return
			}

			// Add auth token to context for API calls
			ctx := r.Context()
			if session.AccessToken != "" {
				ctx = auth.WithToken(ctx, session.AccessToken)
				log.Debug().Str("path", r.URL.Path).Msg("Added auth token to context for protected route")
			}

			// Continue with context that includes the auth token
			next.ServeHTTP(w, r.WithContext(ctx))
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

				// Add user info to context
				ctx := WithUser(r.Context(), &session.UserInfo)

				// Also add auth token to context if available
				if session.AccessToken != "" {
					ctx = auth.WithToken(ctx, session.AccessToken)
					log.Debug().Str("path", r.URL.Path).Msg("Added auth token to context for authenticated user")
				}

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// User not authenticated, continue without user info
			next.ServeHTTP(w, r)
		})
	}
}

// EnrichUser middleware enriches user data from the API when authenticated
func EnrichUser(generatorService *services.GeneratorService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if ok && user != nil && generatorService != nil {
				apiUser, err := generatorService.GetCurrentUser(r.Context(), r)
				if err == nil && apiUser != nil {
					// Extract user ID from resource name if it's in resource format
					userID := apiUser.ID
					if apiUser.ID != "" {
						userResource, parseErr := resource.ParseResourceName(apiUser.ID)
						if parseErr == nil {
							if parsedUser, ok := userResource.(*resource.User); ok {
								userID = parsedUser.ID
							}
						}
					}

					// Create a full name, ensuring it's never empty
					fullName := apiUser.FirstName + " " + apiUser.LastName
					if fullName == " " || fullName == "" {
						// Use email or ID as fallback if name components are empty
						if apiUser.Email != "" {
							fullName = apiUser.Email
						} else if user.Email != "" {
							fullName = user.Email
						} else if userID != "" {
							fullName = userID
						} else {
							fullName = "User" // Last resort fallback
						}
					}

					// Update user info with API data using extracted ID
					enrichedUser := &auth.UserInfo{
						ID:        userID,
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

// ProcessAuthToken middleware to handle CSRF token and add auth token to API calls
func ProcessAuthToken(sessionManager *auth.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is an HTMX request with a CSRF token
			csrfToken := r.Header.Get("X-CSRF-Token")

			// If we have a CSRF token and the route requires authentication
			if csrfToken != "" {
				// Get the session to extract the auth token
				session, err := sessionManager.GetSession(r)
				if err == nil && session != nil && session.AccessToken != "" {
					// Create a new request with the auth token
					ctx := auth.WithToken(r.Context(), session.AccessToken)
					// Continue with the request with auth token context
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else {
					log.Debug().
						Err(err).
						Str("path", r.URL.Path).
						Str("method", r.Method).
						Msg("CSRF token provided but session not found or invalid")
				}
			}

			// Proceed with the original request if no CSRF token or not authenticated
			next.ServeHTTP(w, r)
		})
	}
}
