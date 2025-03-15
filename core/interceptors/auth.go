package interceptors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

// AuthUser holds Auth0 user profile information
type AuthUser struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	Picture     string `json:"picture"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LastLogin   string `json:"last_login"`
	LastIP      string `json:"last_ip"`
	LoginsCount int    `json:"logins_count"`
}

// Auth0 management token cache
var (
	auth0TokenMutex    sync.RWMutex
	auth0Token         string
	auth0TokenExpiry   time.Time
	auth0TokenLifetime = 23 * time.Hour // Management tokens typically last 24 hours, use 23 to be safe
)

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

// getOrCreateUser checks if a user exists in the database and creates one if it doesn't
func getOrCreateUser(ctx context.Context, db *sql.DB, auth0ID string, config util.Auth0Config) (string, error) {
	// Try to find user by Auth0 ID - using null.StringFrom to convert string to null.String
	user, err := models.Users(
		models.UserWhere.Auth0ID.EQ(null.StringFrom(auth0ID)),
	).One(ctx, db)

	if err == nil {
		// User found, return internal ID
		return user.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Unexpected DB error
		return "", err
	}

	// User not found, fetch details from Auth0
	authUser, err := fetchUserFromAuth0(ctx, auth0ID, config)
	if err != nil {
		return "", err
	}

	// Parse name into first name and last name
	firstName, lastName := parseNameFromAuth0(authUser.Name)

	// Create new user
	internalID := uuid.New().String()
	newUser := &models.User{
		ID:        internalID,
		Auth0ID:   null.StringFrom(authUser.UserID),
		Email:     authUser.Email,
		Password:  "", // Empty password since Auth0 handles authentication
		FirstName: firstName,
		LastName:  null.StringFrom(lastName),
		AvatarID:  null.StringFrom(""), // We could store the picture URL from Auth0 if needed
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      models.RoleEnumUser, // Default role
	}

	err = newUser.Insert(ctx, db, boil.Infer())
	if err != nil {
		return "", err
	}

	log.Info().Str("user_id", internalID).Str("auth0_id", auth0ID).Msg("Created new user from Auth0")
	return internalID, nil
}

// parseNameFromAuth0 splits a full name into first and last name
func parseNameFromAuth0(fullName string) (firstName, lastName string) {
	parts := strings.Split(fullName, " ")
	if len(parts) == 0 {
		return "", ""
	}

	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], strings.Join(parts[1:], " ")
}

// fetchUserFromAuth0 fetches user details from Auth0 Management API
func fetchUserFromAuth0(ctx context.Context, auth0ID string, config util.Auth0Config) (*AuthUser, error) {
	// Get management API token
	token, err := getAuth0ManagementToken(ctx, config)
	if err != nil {
		return nil, err
	}

	// Create request to Auth0 Management API
	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://%s/api/v2/users/%s", config.Domain, auth0ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s, status: %d", string(body), resp.StatusCode)
	}

	// Parse response
	var authUser AuthUser
	if err := json.NewDecoder(resp.Body).Decode(&authUser); err != nil {
		return nil, err
	}

	return &authUser, nil
}

// getAuth0ManagementToken gets a token for the Auth0 Management API with caching
func getAuth0ManagementToken(ctx context.Context, config util.Auth0Config) (string, error) {
	// Check if we have a valid cached token
	auth0TokenMutex.RLock()
	if auth0Token != "" && time.Now().Before(auth0TokenExpiry) {
		token := auth0Token
		auth0TokenMutex.RUnlock()
		return token, nil
	}
	auth0TokenMutex.RUnlock()

	// No valid token, acquire write lock and fetch a new one
	auth0TokenMutex.Lock()
	defer auth0TokenMutex.Unlock()

	// Double-check in case another goroutine refreshed the token while we were waiting
	if auth0Token != "" && time.Now().Before(auth0TokenExpiry) {
		return auth0Token, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	tokenURL := fmt.Sprintf("https://%s/oauth/token", config.Domain)

	payload := strings.NewReader(fmt.Sprintf(`{
		"client_id": "%s",
		"client_secret": "%s",
		"audience": "https://%s/api/v2/",
		"grant_type": "client_credentials"
	}`, config.ClientID, config.ClientSecret, config.Domain))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get management token: %s, status: %d", string(body), resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Cache the token
	auth0Token = result.AccessToken
	auth0TokenExpiry = time.Now().Add(auth0TokenLifetime)

	log.Info().Msg("Refreshed Auth0 Management API token")

	return result.AccessToken, nil
}

// AuthInterceptor creates a gRPC interceptor for authentication
func AuthInterceptor(authenticator auth.Authenticator, db *sql.DB, config util.Auth0Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Info().Str("method", info.FullMethod).Msg("Intercepting request")

		if isWhiteListedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		claims, err := authorizeUserFromContext(ctx, authenticator)
		if err != nil {
			return nil, err
		}

		// Get or create user in our database
		internalID, err := getOrCreateUser(ctx, db, claims.UserID, config)
		if err != nil {
			log.Error().Err(err).Str("auth0_id", claims.UserID).Msg("Failed to get or create user")
			return nil, status.Errorf(codes.Internal, "internal error")
		}

		// Add internal user ID to context
		newCtx := context.WithValue(ctx, middleware.AuthKey, internalID)
		return handler(newCtx, req)
	}
}

// HttpAuthInterceptor creates HTTP middleware for authentication
func HttpAuthInterceptor(authenticator auth.Authenticator, db *sql.DB, config util.Auth0Config, handler http.Handler) http.Handler {
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

		// Get or create user in our database
		internalID, err := getOrCreateUser(r.Context(), db, claims.UserID, config)
		if err != nil {
			log.Error().Err(err).Str("auth0_id", claims.UserID).Msg("Failed to get or create user")
			respondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// Add internal user ID to context
		ctx := context.WithValue(r.Context(), middleware.AuthKey, internalID)
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
