package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type userContextKey struct{}

func UserFromContext(ctx context.Context) (*auth.UserInfo, bool) {
	user, ok := ctx.Value(userContextKey{}).(*auth.UserInfo)
	return user, ok
}

func WithUser(ctx context.Context, user *auth.UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// SessionAuthMiddleware gates HTML routes on the SCS cookie. Public paths still
// pick up a user if a session exists. IdentityInterceptor is the API gate.
func SessionAuthMiddleware(sessionManager *auth.SCSSessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isPublicPath := shouldSkipAuthRequirement(r.URL.Path)
			reqID := chiMiddleware.GetReqID(r.Context())

			userID := sessionManager.GetUserID(r)
			if userID == "" {
				if isPublicPath {
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

				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			sessionData, err := sessionManager.GetSession(r)
			if err != nil {
				if isPublicPath {
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

				sessionManager.DestroySession(w, r)
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, &sessionData.UserInfo)

			log.Debug().
				Str("request_id", reqID).
				Str("user_id", userID).
				Str("user_email", sessionData.UserInfo.Email).
				Str("path", r.URL.Path).
				Msg("Request authenticated via session")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func shouldSkipAuthRequirement(path string) bool {
	publicPaths := []string{
		"/",
		"/login",
		"/signup",
		"/forgot-password",
		"/reset-password",
		"/verify",
		"/check-email",
		"/auth/",
		"/logout",
		"/health",
		"/favicon.ico",
		"/css/",
		"/js/",
		"/images/",
		"/static/",
		"/gallery",
		"/about",
		"/rpc",
	}

	for _, publicPath := range publicPaths {
		if publicPath == "/" && path == "/" {
			return true
		}
		if publicPath != "/" && strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}
