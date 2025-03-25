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
		// User found, only update if really necessary
		// Check if we have all required fields
		if user.Email != "" && user.FirstName != "" {
			// We have all the essential info, no need to fetch from Auth0 again
			log.Debug().Str("user_id", user.ID).Msg("User found with complete profile, skipping Auth0 fetch")
			return user.ID, nil
		}

		// Missing essential info, need to update
		needsUpdate := false

		// Check if token is in context for direct userinfo
		token, tokenOk := ctx.Value("token").(string)

		// Get latest Auth0 info anyway to ensure we have the most up-to-date data
		var authUser *auth.UserInfo
		var authErr error

		// Try first with direct token method if available (more reliable)
		if tokenOk && token != "" {
			// Try to use the token-based method which uses userinfo endpoint
			if auth0Service, ok := userProvider.(auth.AuthService); ok {
				authUser, authErr = auth0Service.GetUserInfoFromToken(ctx, token)
				if authErr != nil {
					log.Warn().Err(authErr).Str("auth0_id", auth0ID).Msg("Failed to get user info from token, will try regular method")
				}
			}
		}

		// Fall back to standard method if token method failed
		if authUser == nil {
			authUser, authErr = userProvider.GetUserInfo(ctx, auth0ID)
		}

		if authErr == nil {
			// Only update if Auth0 has data and our local data is empty
			if authUser.Email != "" && user.Email == "" {
				user.Email = authUser.Email
				needsUpdate = true
				log.Debug().Str("user_id", user.ID).Str("email", authUser.Email).Msg("Updating user email from Auth0")
			}

			if authUser.Name != "" && (user.FirstName == "" || user.LastName.String == "") {
				firstName, lastName := parseNameFromAuth0(authUser.Name)
				if firstName != "" && user.FirstName == "" {
					user.FirstName = firstName
					needsUpdate = true
					log.Debug().Str("user_id", user.ID).Str("first_name", firstName).Msg("Updating user first name from Auth0")
				}
				if lastName != "" && user.LastName.String == "" {
					user.LastName = null.StringFrom(lastName)
					needsUpdate = true
					log.Debug().Str("user_id", user.ID).Str("last_name", lastName).Msg("Updating user last name from Auth0")
				}
			}

			if needsUpdate {
				user.UpdatedAt = time.Now()
				_, updateErr := user.Update(ctx, db, boil.Infer())
				if updateErr != nil {
					log.Warn().Err(updateErr).Str("user_id", user.ID).Msg("Failed to update user with Auth0 info")
					// Don't fail the operation just because we couldn't update
				} else {
					log.Info().Str("user_id", user.ID).Msg("Updated user with latest Auth0 info")
				}
			}
		} else {
			// Just log the error but continue
			log.Warn().Err(authErr).Str("auth0_id", auth0ID).Msg("Couldn't fetch updated user info from Auth0")
		}

		// Return internal ID even if update failed
		return user.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Unexpected DB error
		return "", err
	}

	// User not found, fetch details from Auth0
	// Try first with direct token method if available (more reliable)
	var authUser *auth.UserInfo
	var authErr error

	token, tokenOk := ctx.Value("token").(string)
	if tokenOk && token != "" {
		// Try to use the token-based method which uses userinfo endpoint
		if auth0Service, ok := userProvider.(auth.AuthService); ok {
			authUser, authErr = auth0Service.GetUserInfoFromToken(ctx, token)
			if authErr != nil {
				log.Warn().Err(authErr).Str("auth0_id", auth0ID).Msg("Failed to get user info from token, will try regular method")
			}
		}
	}

	// Fall back to standard method if token method failed
	if authUser == nil {
		authUser, authErr = userProvider.GetUserInfo(ctx, auth0ID)
		if authErr != nil {
			return "", authErr
		}
	}

	log.Info().
		Str("id", authUser.ID).
		Str("email", authUser.Email).
		Str("name", authUser.Name).
		Str("provider", authUser.Provider).
		Msg("Creating new user from Auth0 info")

	// Parse name into first name and last name
	firstName, lastName := parseNameFromAuth0(authUser.Name)

	// If Auth0 didn't provide a name, use a default based on provider
	if firstName == "" {
		if authUser.Provider != "" {
			firstName = "User from " + authUser.Provider
		} else {
			firstName = "New User"
		}
	}

	// Create new user
	internalID := uuid.New().String()
	newUser := &models.User{
		ID:        internalID,
		Auth0ID:   authUser.ID,
		Email:     authUser.Email,
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

		// Get or create user in our database
		internalID, err := getOrCreateUser(ctx, db, claims.UserID, authService)
		if err != nil {
			log.Error().Err(err).Str("auth0_id", claims.UserID).Msg("Failed to get or create user")
			return nil, status.Errorf(codes.Internal, "internal error")
		}

		// Add internal user ID and raw token to context
		newCtx := context.WithValue(ctx, middleware.AuthKey, internalID)
		// Store the raw token in context for later use in GetUserInfo
		newCtx = context.WithValue(newCtx, "token", token)
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
