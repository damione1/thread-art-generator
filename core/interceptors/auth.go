package interceptors

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

var whiteListedMethods = []string{
	"/pb.ArtGeneratorService/GetMediaUploadUrl",
	"/pb.ArtGeneratorService/GetMediaDownloadUrl",
	"/pb.ArtGeneratorService/CreateUser",
}

var whiteListedHttpEndpoints = []string{
	"/swagger",
}

// Helper function to extract and validate token from gRPC metadata
func authorizeUserFromContext(ctx context.Context, authenticator auth.Authenticator) (*auth.AuthClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	return authorizeUserFromHeader(values, authenticator)
}

// Helper function to extract and validate token from authorization header
func authorizeUserFromHeader(authHeader []string, authenticator auth.Authenticator) (*auth.AuthClaims, error) {
	if len(authHeader) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	bearerToken := authHeader[0]
	fields := strings.Fields(bearerToken)
	if len(fields) < 2 {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, status.Errorf(codes.Unauthenticated, "unsupported authorization type: %s", authType)
	}

	token := fields[1]
	claims, err := authenticator.ValidateToken(context.Background(), token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims, nil
}

// AuthInterceptor creates a gRPC interceptor for authentication
func AuthInterceptor(authenticator auth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Info().Str("method", info.FullMethod).Msg("Intercepting request")

		if isWhiteListedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		claims, err := authorizeUserFromContext(ctx, authenticator)
		if err != nil {
			return nil, err
		}

		// Add user info to context
		newCtx := context.WithValue(ctx, middleware.AuthKey, claims.UserID)
		return handler(newCtx, req)
	}
}

// HttpAuthInterceptor creates HTTP middleware for authentication
func HttpAuthInterceptor(authenticator auth.Authenticator, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isWhiteListedEndpoint(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get(authorizationHeader)
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "authorization token is not provided")
			return
		}

		claims, err := authorizeUserFromHeader([]string{authHeader}, authenticator)
		if err != nil {
			statusErr, ok := status.FromError(err)
			if !ok {
				respondWithError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			respondWithError(w, http.StatusUnauthorized, statusErr.Message())
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), middleware.AuthKey, claims.UserID)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondWithError(res http.ResponseWriter, code int, message string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	json.NewEncoder(res).Encode(map[string]string{"error": message})
}

func isWhiteListedMethod(method string) bool {
	for _, whiteListedMethod := range whiteListedMethods {
		if whiteListedMethod == method {
			return true
		}
	}
	return false
}

func isWhiteListedEndpoint(path string) bool {
	for _, whiteListedEndpoint := range whiteListedHttpEndpoints {
		if strings.HasPrefix(path, whiteListedEndpoint) {
			return true
		}
	}
	return false
}
