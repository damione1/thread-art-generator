package interceptors

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/middleware"
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

// Helper function to extract and validate token from gRPC metadata
func authorizeUserFromContext(ctx context.Context, authenticator auth.Authenticator) (*auth.AuthClaims, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	bearerToken := values[0]
	fields := strings.Fields(bearerToken)
	if len(fields) < 2 {
		return nil, "", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, "", status.Errorf(codes.Unauthenticated, "unsupported authorization type: %s", authType)
	}

	token := fields[1]
	claims, err := authenticator.ValidateToken(context.Background(), token)
	if err != nil {
		return nil, "", status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims, token, nil
}

// getOrCreateUser checks if a user exists in the database and creates one if it doesn't
func getOrCreateUser(ctx context.Context, db *sql.DB, auth0ID string, userProvider auth.UserProvider) (string, error) {
	// Try to find user by Auth0 ID - using null.StringFrom to convert string to null.String
	user, err := models.Users(
		models.UserWhere.Auth0ID.EQ(auth0ID),
	).One(ctx, db)

	if err == nil {
		// User found, we'll use what we have
		return user.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Unexpected DB error
		return "", err
	}

	// User not found, get info from token and create new user
	claims, ok := ctx.Value("claims").(*auth.AuthClaims)
	if !ok {
		// If claims aren't in context, just create with minimal info
		log.Warn().Str("auth0_id", auth0ID).Msg("Creating user with minimal info, no claims in context")
		return createMinimalUser(ctx, db, auth0ID)
	}

	// Parse name into first name and last name
	firstName, lastName := parseNameFromAuth0(claims.Name)

	// If Auth0 didn't provide a name, use a default
	if firstName == "" {
		firstName = "New User"
	}

	// Get avatar URL from Auth0 token or default to empty
	avatarURL := ""
	if claims.Picture != "" {
		avatarURL = claims.Picture
		log.Debug().Str("auth0_id", auth0ID).Str("avatar_url", avatarURL).Msg("Using avatar from Auth0")
	}

	// Create new user with info from token
	internalID := uuid.New().String()
	newUser := &models.User{
		ID:        internalID,
		Auth0ID:   auth0ID,
		Email:     null.StringFrom(claims.Email),
		FirstName: firstName,
		LastName:  null.StringFrom(lastName),
		AvatarID:  null.StringFrom(avatarURL),
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      models.RoleEnumUser, // Default role
	}

	err = newUser.Insert(ctx, db, boil.Infer())
	if err != nil {
		return "", err
	}

	log.Info().Str("user_id", internalID).Str("auth0_id", auth0ID).Msg("Created new user from token claims")
	return internalID, nil
}

// createMinimalUser creates a user with minimal information
func createMinimalUser(ctx context.Context, db *sql.DB, auth0ID string) (string, error) {
	internalID := uuid.New().String()
	newUser := &models.User{
		ID:        internalID,
		Auth0ID:   auth0ID,
		Email:     null.String{}, // Empty email
		FirstName: "New User",
		LastName:  null.String{},
		AvatarID:  null.String{}, // Empty avatar, will fall back to Gravatar
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      models.RoleEnumUser,
	}

	err := newUser.Insert(ctx, db, boil.Infer())
	if err != nil {
		return "", err
	}

	log.Info().Str("user_id", internalID).Str("auth0_id", auth0ID).Msg("Created new user with minimal info")
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

// AuthInterceptor creates a gRPC interceptor for authentication
func AuthInterceptor(authService auth.AuthService, db *sql.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for whitelisted methods
		if isWhiteListedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Check if user ID is already in context (e.g., from previous middleware)
		if userID, ok := ctx.Value(middleware.AuthKey).(string); ok && userID != "" {
			log.Debug().Str("user_id", userID).Msg("User already authenticated, skipping auth check")
			// User already authenticated, proceed with handler
			return handler(ctx, req)
		}

		claims, token, err := authorizeUserFromContext(ctx, authService)
		if err != nil {
			return nil, err
		}

		// Store claims in context for user creation if needed
		ctxWithClaims := context.WithValue(ctx, "claims", claims)

		// Get or create user in our database
		internalID, err := getOrCreateUser(ctxWithClaims, db, claims.UserID, authService)
		if err != nil {
			log.Error().Err(err).Str("auth0_id", claims.UserID).Msg("Failed to get or create user")
			return nil, status.Errorf(codes.Internal, "internal error")
		}

		// Add internal user ID and raw token to context
		newCtx := context.WithValue(ctxWithClaims, middleware.AuthKey, internalID)
		// Store the raw token in context for later use
		newCtx = context.WithValue(newCtx, "token", token)
		// Pass the server instance through the context
		newCtx = context.WithValue(newCtx, "server", info.Server)
		return handler(newCtx, req)
	}
}

func isWhiteListedMethod(method string) bool {
	for _, whiteListedMethod := range whiteListedMethods {
		if whiteListedMethod == method {
			return true
		}
	}
	return false
}
