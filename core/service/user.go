package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/bufbuild/protovalidate-go"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader            = "user-agent"
	xForwardedForHeader        = "x-forwarded-for"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	pbUser := req.GetUser()

	userResource, err := resource.ParseResourceName(pbUser.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("user.name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	user, ok := userResource.(*resource.User)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("user.name", errors.New("invalid user resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Get user ID from context using the same key used in auth interceptor
	userIdFromContext, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	if user.ID != userIdFromContext {
		return nil, pbErrors.PermissionDeniedError("cannot update other user's info")
	}

	userDb, err := models.Users(models.UserWhere.ID.EQ(user.ID)).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	if pbUser.GetFirstName() != "" {
		userDb.FirstName = pbUser.GetFirstName()
	}

	userDb.LastName.Valid = false
	userDb.LastName.String = pbUser.GetLastName()
	if pbUser.GetLastName() != "" {
		userDb.LastName.Valid = true
	}

	if pbUser.GetEmail() != "" {
		userDb.Email.Valid = true
		userDb.Email.String = pbUser.GetEmail()
	}

	// If avatar is provided in the request, update it
	// This allows clients to set a custom avatar if needed
	if pbUser.GetAvatar() != "" && pbUser.GetAvatar() != userDb.AvatarID.String {
		userDb.AvatarID.Valid = true
		userDb.AvatarID.String = pbUser.GetAvatar()
	}
	// Note: We don't reset AvatarID if it's not provided to preserve the Auth0 avatar

	if _, err = userDb.Update(ctx, server.config.DB, boil.Infer()); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("email", errors.New(pbErrors.ErrEmailAlreadyExists)),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		return nil, pbErrors.InternalError("failed to update user", err)
	}

	return pbx.DbUserToProto(userDb), nil
}

func (server *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	userResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	user, ok := userResource.(*resource.User)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid user resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Get user from database - user should already exist from auth sync
	userDb, err := server.getUserFromFirebaseUID(ctx, firebaseUID)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("GetUser failed to get user - user should have been created during auth sync")
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	// Ensure the current user has permission to access the requested user
	if user.ID != userDb.ID {
		return nil, pbErrors.PermissionDeniedError("cannot get other user's info")
	}

	return pbx.DbUserToProto(userDb), nil
}

// GetCurrentUser retrieves the current authenticated user based on the context
func (server *Server) GetCurrentUser(ctx context.Context, req *pb.GetCurrentUserRequest) (*pb.User, error) {
	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.UnauthenticatedError("user not authenticated")
	}

	// Get user from database - user should already exist from auth sync
	user, err := server.getUserFromFirebaseUID(ctx, firebaseUID)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("GetCurrentUser failed to get user - user should have been created during auth sync")
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	return pbx.DbUserToProto(user), nil
}

// SyncUserFromFirebase creates or updates a user from Firebase Authentication data
// This endpoint is called by Firebase Cloud Functions when a user is created/updated
func (server *Server) SyncUserFromFirebase(ctx context.Context, req *pb.SyncUserFromFirebaseRequest) (*pb.User, error) {
	// Note: Internal API key validation is now handled by the Connect adapter

	// Validate the request
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	log.Info().
		Str("firebase_uid", req.FirebaseUid).
		Str("email", req.Email).
		Str("display_name", req.DisplayName).
		Msg("SyncUserFromFirebase: Syncing user from Firebase")

	// Check if user already exists (idempotency)
	existingUser, err := server.getUserFromFirebaseUID(ctx, req.FirebaseUid)
	if err == nil && existingUser != nil {
		log.Info().
			Str("firebase_uid", req.FirebaseUid).
			Str("user_id", existingUser.ID).
			Msg("SyncUserFromFirebase: User already exists, returning existing user")
		return pbx.DbUserToProto(existingUser), nil
	}

	// Parse display name into first and last names
	firstName, lastName := server.parseDisplayName(req.DisplayName)

	// Create new user in PostgreSQL
	user, err := server.createUserFromFirebaseClaims(
		ctx,
		req.FirebaseUid,
		req.Email,
		firstName+" "+lastName, // Full name for compatibility
		req.PhotoUrl,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("firebase_uid", errors.New("user with this Firebase UID already exists")),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		log.Error().Err(err).
			Str("firebase_uid", req.FirebaseUid).
			Msg("SyncUserFromFirebase: Failed to create user")
		return nil, pbErrors.InternalError("failed to create user", err)
	}

	log.Info().
		Str("firebase_uid", req.FirebaseUid).
		Str("user_id", user.ID).
		Msg("SyncUserFromFirebase: Successfully created new user")

	return pbx.DbUserToProto(user), nil
}
