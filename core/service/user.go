package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/bufbuild/protovalidate-go"
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

func (server *Server) updateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
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

	target, ok := userResource.(*resource.User)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("user.name", errors.New("invalid user resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	current, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if target.ID != current.ID {
		return nil, pbErrors.PermissionDeniedError("cannot update other user's info")
	}

	userDb, err := models.Users(models.UserWhere.ID.EQ(current.ID)).One(ctx, server.config.DB)
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

func (server *Server) getUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
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

	current, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	if user.ID != current.ID {
		return nil, pbErrors.PermissionDeniedError("cannot get other user's info")
	}

	return pbx.DbUserToProto(current), nil
}

func (server *Server) getCurrentUser(ctx context.Context, req *pb.GetCurrentUserRequest) (*pb.User, error) {
	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	return pbx.DbUserToProto(user), nil
}
