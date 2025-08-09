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
	contextKeyClaims contextKey = "paseto_claims"
	contextKeyToken  contextKey = "paseto_token"
)

const (
	authorizationHeader = "Authorization"
	authorizationBearer = "Bearer"
)

var whiteListedPaths = []string{
	"/pb.ArtGeneratorService/SyncUserFromFirebase",
	"/pb.FirebaseFunctionsService/ConfirmArtImageUploadFromFunction", // Allow Firebase Functions to confirm art uploads
}

// PasetoAuthMiddleware creates a Connect middleware that validates PASETO tokens for internal BFF → API communication
// This provides secure, stateless authentication optimized for high performance
func PasetoAuthMiddleware(pasetoService *auth.PasetoService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for whitelisted paths
			if isWhiteListedPath(req.Spec().Procedure) {
				return next(ctx, req)
			}

			// Check if user ID is already in context (e.g., from previous middleware)
			if userID, ok := ctx.Value(middleware.AuthKey).(string); ok && userID != "" {
				log.Debug().Str("user_id", userID).Msg("User already authenticated via context, skipping auth check")
				return next(ctx, req)
			}

			// Create HTTP request for validation
			httpReq := &http.Request{Header: make(http.Header)}
			for k, v := range req.Header() {
				httpReq.Header[k] = v
			}

			// Extract and validate PASETO token
			claims, token, err := authorizeUserFromPasetoHeaders(ctx, httpReq.Header, pasetoService)
			if err != nil {
				// Audit log: Failed authentication attempt
				log.Warn().
					Err(err).
					Str("endpoint", req.Spec().Procedure).
					Str("user_agent", req.Header().Get("User-Agent")).
					Msg("PASETO authentication failed")
				return nil, err
			}

			// Audit log: Successful PASETO authentication
			log.Debug().
				Str("user_id", claims.UserID).
				Str("endpoint", req.Spec().Procedure).
				Str("user_email", claims.Email).
				Str("provider", claims.Provider).
				Msg("PASETO authentication successful")

			// Create context with PASETO claims for downstream use
			ctxWithClaims := context.WithValue(ctx, contextKeyClaims, claims)
			ctxWithToken := context.WithValue(ctxWithClaims, contextKeyToken, token)
			ctxWithUser := context.WithValue(ctxWithToken, middleware.AuthKey, claims.UserID)

			return next(ctxWithUser, req)
		}
	}
}

// authorizeUserFromPasetoHeaders extracts and validates PASETO token from HTTP headers
func authorizeUserFromPasetoHeaders(_ context.Context, headers http.Header, pasetoService *auth.PasetoService) (*auth.AuthClaims, string, error) {
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

	// Validate PASETO token
	claims, err := pasetoService.ValidateToken(token)
	if err != nil {
		log.Debug().Err(err).Msg("PASETO token validation failed")
		return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
	}

	return claims, token, nil
}

func isWhiteListedPath(path string) bool {
	return slices.Contains(whiteListedPaths, path)
}