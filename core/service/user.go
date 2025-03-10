package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	var err error

	pbUser := req.GetUser()
	pbUser.Name = "" // name is not allowed to be set by the user
	dbUser := pbx.ProtoUserToDb(pbUser)

	err = dbUser.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("user.email", errors.New(pbErrors.ErrEmailAlreadyExists)),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		return nil, pbErrors.InternalError("failed to insert user", err)
	}

	accountActivation := &models.AccountActivation{
		UserID:          dbUser.ID,
		UserEmail:       dbUser.Email,
		ActivationToken: util.RandomInt(1000000, 9999999),
	}
	if err = accountActivation.Insert(ctx, server.config.DB, boil.Infer()); err != nil {
		return nil, pbErrors.InternalError("failed to insert account validation", err)
	}

	if err = server.mailService.SendValidateEmail(dbUser.Email, dbUser.FirstName, dbUser.LastName.String, accountActivation.ActivationToken); err != nil {
		return nil, pbErrors.InternalError("failed to send email", err)
	}

	pbUser = pbx.DbUserToProto(dbUser)

	return pbUser, nil
}

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	pbUser := req.GetUser()

	userId, err := pbx.GetResourceIDByType(pbUser.GetName(), pbx.RessourceTypeUsers)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", pbErrors.ErrValidationPrefix, pbErrors.ErrInvalidResourceName, err)
	}

	userIdFromToken := middleware.FromAdminContext(ctx).UserPayload.UserID
	if userId != userIdFromToken {
		return nil, pbErrors.PermissionDeniedError("cannot update other user's info")
	}

	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	updateMask := req.GetUpdateMask()
	if updateMask != nil && len(updateMask.GetPaths()) > 0 {
		for _, path := range updateMask.GetPaths() {
			switch path {
			case "first_name":
				if pbUser.GetFirstName() != "" {
					user.FirstName = pbUser.GetFirstName()
				}
			case "last_name":
				user.LastName.String = pbUser.GetLastName()
				user.LastName.Valid = false
				if pbUser.GetLastName() != "" {
					user.LastName.Valid = true
				}
			case "email":
				if pbUser.GetEmail() != "" {
					user.Email = pbUser.GetEmail()
				}
			default:
				return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
					pbErrors.FieldViolation("updateMask", errors.New(fmt.Sprintf("invalid field mask: %s", path))),
				})
			}
		}
	}

	if _, err = user.Update(ctx, server.config.DB, boil.Infer()); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("email", errors.New(pbErrors.ErrEmailAlreadyExists)),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		return nil, pbErrors.InternalError("failed to update user", err)
	}

	return pbx.DbUserToProto(user), nil
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if clientIPS := md.Get(xForwardedForHeader); len(clientIPS) > 0 {
			mtdt.ClientIP = clientIPS[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = p.Addr.String()
	}

	return mtdt
}

func (server *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	userId, err := pbx.GetResourceIDByType(req.GetName(), pbx.RessourceTypeUsers)
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New(pbErrors.ErrInvalidResourceName)),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	userIdFromToken := middleware.FromAdminContext(ctx).UserPayload.UserID
	if userId != userIdFromToken {
		return nil, pbErrors.PermissionDeniedError("cannot get other user's info")
	}

	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	return pbx.DbUserToProto(user), nil
}
