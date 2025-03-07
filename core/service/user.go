package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

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
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	dbUser.Password, err = util.HashPassword(pbUser.GetPassword())
	if err != nil {
		return nil, pbErrors.InternalError("failed to hash password", err)
	}

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
			case "password":
				if pbUser.GetPassword() != "" {
					hashedPassword, err := util.HashPassword(pbUser.GetPassword())
					if err != nil {
						return nil, pbErrors.InternalError("failed to hash password", err)
					}
					user.Password = hashedPassword
				}
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

func (server *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	user, err := models.Users(models.UserWhere.Email.EQ(req.GetEmail())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Print(fmt.Sprintf("User %s not found", req.GetEmail()))
			// Create field violations for both email and password
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("email", errors.New(pbErrors.ErrIncorrectCredentials)),
				pbErrors.FieldViolation("password", errors.New(pbErrors.ErrIncorrectCredentials)),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	err = util.CheckPassword(req.GetPassword(), user.Password)
	if err != nil {
		log.Print(fmt.Sprintf("User %s password incorrect", req.GetEmail()))
		// Create field violations for both email and password
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("email", errors.New(pbErrors.ErrIncorrectCredentials)),
			pbErrors.FieldViolation("password", errors.New(pbErrors.ErrIncorrectCredentials)),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if !user.Active {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("user", errors.New(pbErrors.ErrUserNotActive)),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, pbErrors.InternalError("failed to create access token", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, pbErrors.InternalError("failed to create refresh token", err)
	}

	metadata := server.extractMetadata(ctx)

	session := &models.Session{
		ID:           refreshPayload.ID.String(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		ClientIP:     metadata.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpireTime,
	}
	err = session.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to insert session", err)
	}

	return &pb.CreateSessionResponse{
		User:                   pbx.DbUserToProto(user),
		SessionId:              session.ID,
		AccessToken:            accessToken,
		RefreshToken:           refreshToken,
		AccessTokenExpireTime:  timestamppb.New(accessPayload.ExpireTime),
		RefreshTokenExpireTime: timestamppb.New(refreshPayload.ExpireTime),
	}, nil
}

func (server *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	refreshPayload, err := server.tokenMaker.ValidateToken(req.GetRefreshToken())
	if err != nil {
		return nil, pbErrors.InternalError("failed to verify token", err)
	}

	session, err := models.Sessions(models.SessionWhere.RefreshToken.EQ(req.GetRefreshToken())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("session not found")
		}
		return nil, pbErrors.InternalError("failed to get session", err)
	}

	if session.IsBlocked {
		return nil, pbErrors.PermissionDeniedError(pbErrors.ErrSessionBlocked)
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		session.UserID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, pbErrors.InternalError("failed to create access token", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		session.UserID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, pbErrors.InternalError("failed to create refresh token", err)
	}

	session.RefreshToken = refreshToken
	session.ExpiresAt = refreshPayload.ExpireTime
	_, err = session.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to update session", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:            accessToken,
		AccessTokenExpireTime:  timestamppb.New(accessPayload.ExpireTime),
		RefreshToken:           refreshToken,
		RefreshTokenExpireTime: timestamppb.New(refreshPayload.ExpireTime),
	}, nil
}

func (server *Server) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	session, err := models.Sessions(models.SessionWhere.RefreshToken.EQ(req.GetRefreshToken())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("session not found")
		}
		return nil, pbErrors.InternalError("failed to get session", err)
	}

	_, err = session.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to delete session", err)
	}

	return &emptypb.Empty{}, nil
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

func (server *Server) ValidateEmail(ctx context.Context, req *pb.ValidateEmailRequest) (*emptypb.Empty, error) {
	activation, err := models.AccountActivations(
		models.AccountActivationWhere.UserEmail.EQ(req.GetEmail()),
		models.AccountActivationWhere.ActivationToken.EQ(int(req.GetValidationNumber())),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create field violations for both email and validation_number
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("email", errors.New("account activation not found")),
				pbErrors.FieldViolation("validationNumber", errors.New("account activation not found")),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		return nil, pbErrors.InternalError("failed to get account activation", err)
	}

	user, err := models.Users(models.UserWhere.ID.EQ(activation.UserID)).One(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	user.Active = true
	_, err = user.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to update user", err)
	}

	_, err = activation.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to delete account activation", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) SendValidationEmail(ctx context.Context, req *pb.SendValidationEmailRequest) (*emptypb.Empty, error) {
	dbUser, err := models.Users(models.UserWhere.Email.EQ(req.GetEmail())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			violations := []*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("email", errors.New(pbErrors.ErrUserNotFound)),
			}
			return nil, pbErrors.InvalidArgumentError(violations)
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}
	if dbUser.Active {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("email", errors.New("user is already active")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	activationsRequestCount, err := models.AccountActivations(
		models.AccountActivationWhere.UserEmail.EQ(req.GetEmail()),
	).Count(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to count account activations", err)
	}

	if activationsRequestCount > 5 {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("email", errors.New(pbErrors.ErrTooManyValidationRequests)),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
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

	return &emptypb.Empty{}, nil
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
