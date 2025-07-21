package interceptors

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Context keys for type-safe context operations
type contextKey string

const (
	contextKeyClaims contextKey = "firebase_claims"
	contextKeyToken  contextKey = "firebase_token"
)

const (
	authorizationHeader = "Authorization"
	authorizationBearer = "Bearer"
)

var whiteListedPaths = []string{
	"/pb.ArtGeneratorService/GetMediaUploadUrl",
	"/pb.ArtGeneratorService/GetMediaDownloadUrl",
	"/pb.ArtGeneratorService/CreateUser",
}

// Helper function to extract and validate token from HTTP headers
func authorizeUserFromHeaders(ctx context.Context, headers http.Header, authenticator auth.Authenticator) (*auth.AuthClaims, string, error) {
	authHeader := headers.Get(authorizationHeader)
	if authHeader == "" {
		log.Debug().Msg("No Authorization header found")
		return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("authorization token is not provided"))
	}

	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		log.Debug().Str("header", authHeader).Msg("Invalid Authorization header format")
		return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization header format"))
	}

	authType := strings.ToLower(fields[0])
	if authType != strings.ToLower(authorizationBearer) {
		log.Debug().Str("auth_type", authType).Msg("Unsupported authorization type")
		return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("unsupported authorization type"))
	}

	token := fields[1]

	claims, err := authenticator.ValidateToken(ctx, token)
	if err != nil {
		log.Debug().Err(err).Msg("Token validation failed")
		return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
	}
	return claims, token, nil
}

// AuthMiddleware creates a simplified Connect middleware for Firebase authentication
// Uses Firebase UID directly without database user management complexity
func AuthMiddleware(authService auth.AuthService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for whitelisted paths
			if isWhiteListedPath(req.Spec().Procedure) {
				return next(ctx, req)
			}

			// Check if user ID is already in context (e.g., from previous middleware)
			if userID, ok := ctx.Value(middleware.AuthKey).(string); ok && userID != "" {
				log.Debug().Str("firebase_uid", userID).Msg("User already authenticated, skipping auth check")
				return next(ctx, req)
			}

			// Validate Firebase token and extract claims
			claims, token, err := authorizeUserFromHeaders(ctx, req.Header(), authService)
			if err != nil {
				// Audit log: Failed authentication attempt
				log.Warn().
					Err(err).
					Str("endpoint", req.Spec().Procedure).
					Str("user_agent", req.Header().Get("User-Agent")).
					Msg("Firebase authentication failed")
				return nil, err
			}

			// Audit log: Successful authentication
			log.Info().
				Str("firebase_uid", claims.UserID).
				Str("endpoint", req.Spec().Procedure).
				Str("user_email", claims.Email).
				Str("provider", claims.Provider).
				Msg("Firebase authentication successful")

			// Create context with Firebase claims and token for downstream use
			ctxWithClaims := context.WithValue(ctx, contextKeyClaims, claims)
			ctxWithToken := context.WithValue(ctxWithClaims, contextKeyToken, token)

			// Use Firebase UID directly as the user identifier
			ctxWithUser := context.WithValue(ctxWithToken, middleware.AuthKey, claims.UserID)

			return next(ctxWithUser, req)
		}
	}
}

func isWhiteListedPath(path string) bool {
	return slices.Contains(whiteListedPaths, path)
}
