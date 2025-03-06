package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/web/client"
)

type contextKey string

const (
	// UserContextKey is the key used to store the user in the context
	UserContextKey contextKey = "user"
)

// RequireAuth is a middleware that checks if the user is authenticated
func RequireAuth(grpcClient *client.GrpcClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the session token from the request
			token := client.GetSessionToken(r)
			if token == "" {
				// If there's no token, try to refresh using the refresh token
				refreshToken := client.GetRefreshToken(r)
				if refreshToken == "" {
					// No refresh token either, redirect to login
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				// Try to refresh the token
				ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()

				refreshReq := &pb.RefreshTokenRequest{
					RefreshToken: refreshToken,
				}
				refreshResp, err := grpcClient.GetClient().RefreshToken(ctx, refreshReq)
				if err != nil {
					// Failed to refresh, clear cookies and redirect to login
					client.ClearSessionCookies(w)
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				// Set the refreshed cookies
				client.SetRefreshedCookies(w, refreshResp)
				token = refreshResp.AccessToken
			}

			// Get the current user
			ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Add auth to context
			authCtx := client.WithAuth(ctx, token)

			userReq := &pb.GetUserRequest{
				Name: "me",
			}
			user, err := grpcClient.GetClient().GetUser(authCtx, userReq)
			if err != nil {
				// Failed to get user, clear cookies and redirect to login
				client.ClearSessionCookies(w)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Store the user in the context
			ctx = context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext gets the user from the context
func GetUserFromContext(ctx context.Context) *pb.User {
	user, ok := ctx.Value(UserContextKey).(*pb.User)
	if !ok {
		return nil
	}
	return user
}
