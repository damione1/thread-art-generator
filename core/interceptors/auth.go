package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/token"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

var whiteListedMethods = []string{
	"/pb.ArtGeneratorService/GetMediaUploadUrl",
	"/pb.ArtGeneratorService/GetMediaDownloadUrl",
	"/pb.ArtGeneratorService/CreateUser",
	"/pb.ArtGeneratorService/CreateSession",
	"/pb.ArtGeneratorService/RefreshToken",
	"/pb.ArtGeneratorService/DeleteSession",
	"/pb.ArtGeneratorService/ResetPassword",
	"/pb.ArtGeneratorService/ValidateEmail",
	"/pb.ArtGeneratorService/SendValidationEmail",
}

var whiteListedHttpEndpoints = []string{
	"/swagger",
	"/v1/tokens:refresh",
	"/v1/sessions",
	"/v1/users/<string>:resetPassword",
	"/v1/users/<string>:validateEmail",
	"/v1/users/<string>:sendValidationEmail",
}

// Helper function to authorize user from gRPC metadata
func authorizeUserFromContext(ctx context.Context, tokenMaker token.Maker) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	return authorizeUserFromHeader(md.Get(authorizationHeader), tokenMaker)
}

// Helper function to extract and validate token from HTTP header
func authorizeUserFromHeader(authHeader []string, tokenMaker token.Maker) (*token.Payload, error) {
	if len(authHeader) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	fields := strings.Fields(authHeader[0])
	if len(fields) < 2 || strings.ToLower(fields[0]) != authorizationBearer {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	accessToken := fields[1]
	return tokenMaker.ValidateToken(accessToken)
}

// gRPC Auth Interceptor
func AuthInterceptor(tokenMaker token.Maker) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Skip authentication if the method is white-listed
		log.Debug().Str("method", info.FullMethod).Msg("checking white-listed methods")
		if isWhiteListedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		payload, err := authorizeUserFromContext(ctx, tokenMaker)
		if err != nil {
			return nil, pbErrors.UnauthenticatedError("unauthorized: " + err.Error())
		}

		adminCtx := middleware.NewAdminContext(ctx, payload)
		return handler(adminCtx, req)
	}
}

// HTTP Auth Interceptor
func HttpAuthInterceptor(tokenMaker token.Maker, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
		}

		// Skip authentication if the endpoint is white-listed
		log.Debug().Str("path", req.URL.Path).Msg("checking white-listed endpoints")
		if isWhiteListedEndpoint(req.URL.Path) {
			handler.ServeHTTP(rec, req)
			return
		}

		authHeader := req.Header.Get(authorizationHeader)
		payload, err := authorizeUserFromHeader([]string{authHeader}, tokenMaker)
		if err != nil {
			respondWithError(res, http.StatusUnauthorized, err.Error())
			log.Error().Err(err).Msg("invalid access token")
			return
		}

		adminCtx := middleware.NewAdminContext(req.Context(), payload)
		handler.ServeHTTP(rec, req.WithContext(adminCtx))

		duration := time.Since(startTime)
		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error()
		}

		logger.Str("protocol", "http").
			Str("method", req.Method).
			Str("path", req.URL.Path).
			Int("status_code", rec.StatusCode).
			Dur("duration", duration).
			Msg("received an HTTP request")
	})
}

func respondWithError(res http.ResponseWriter, code int, message string) {
	res.WriteHeader(code)
	json.NewEncoder(res).Encode(map[string]string{"error": message})
}

// Check if the method is white-listed
func isWhiteListedMethod(method string) bool {
	for _, m := range whiteListedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// Check if the endpoint is white-listed
func isWhiteListedEndpoint(path string) bool {
	for _, p := range whiteListedHttpEndpoints {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
