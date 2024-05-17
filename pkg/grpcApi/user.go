package grpcApi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/pbx"
	"github.com/Damione1/thread-art-generator/pkg/util"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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
	if err := validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	pbUser := req.GetUser()
	pbUser.Name = "" // name is not allowed to be set by the user
	dbUser := pbx.ProtoUserToDb(pbUser)

	dbUser.Password, err = util.HashPassword(pbUser.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	err = dbUser.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to insert user: %s", err)
	}

	accountActivation := &models.AccountActivation{
		UserID:          dbUser.ID,
		UserEmail:       dbUser.Email,
		ActivationToken: util.RandomInt(1000000, 9999999),
	}
	if err = accountActivation.Insert(ctx, server.config.DB, boil.Infer()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert account validation: %s", err)
	}

	if err := server.mailService.SendValidateEmail(dbUser.Email, dbUser.FirstName, dbUser.LastName.String, accountActivation.ActivationToken); err != nil {
		status.Errorf(codes.Internal, "failed to send email: %s", err)
	}

	pbUser = pbx.DbUserToProto(dbUser)

	return pbUser, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.User, validation.Required, validation.By(
			func(value interface{}) error {
				user := value.(*pb.User)
				return validation.ValidateStruct(user,
					validation.Field(&user.Email, validation.Required, is.Email),
					validation.Field(&user.Password, validation.By(checkPassword)),
					validation.Field(&user.Name, validation.Length(0, 0)), // name is not allowed to be set by the user
					validation.Field(&user.FirstName, validation.Required, validation.Length(1, 255)),
					validation.Field(&user.LastName, validation.NilOrNotEmpty, validation.Length(1, 255)),
				)
			},
		)),
	)
}

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err = validateUpdateUserRequest(req); err != nil {
		return nil, err

	}
	pbUser := req.GetUser()

	userId, err := pbx.GetResourceIDByType(pbUser.GetName(), pbx.RessourceTypeUsers)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ressource name: %s", err)
	}

	if userId != authPayload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	updateMask := req.GetUpdateMask()
	if updateMask != nil && len(updateMask.GetPaths()) > 0 {
		for _, path := range updateMask.GetPaths() {
			switch path {
			case "password":
				if pbUser.GetPassword() != "" {
					hashedPassword, err := util.HashPassword(pbUser.GetPassword())
					if err != nil {
						return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
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
				return nil, status.Errorf(codes.InvalidArgument, "Invalid field mask: %s", path)
			}
		}
	}

	if _, err = user.Update(ctx, server.config.DB, boil.Infer()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	return pbx.DbUserToProto(user), nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) error {
	// Ensure the UpdateMask and User fields are provided
	err := validation.ValidateStruct(req,
		validation.Field(&req.UpdateMask, validation.Required),
		validation.Field(&req.User, validation.Required),
	)
	if err != nil {
		return err
	}

	user := req.GetUser()
	updateMaskPaths := req.GetUpdateMask().GetPaths()

	// Dynamically build validation rules based on the fields present in the UpdateMask
	var rules []*validation.FieldRules

	rules = append(rules, validation.Field(&user.Name, validation.Required))

	if slices.Contains(updateMaskPaths, "email") {
		rules = append(rules, validation.Field(&user.Email, validation.Required, is.Email))
	}

	if slices.Contains(updateMaskPaths, "first_name") {
		rules = append(rules, validation.Field(&user.FirstName, validation.Required, validation.Length(2, 255)))
	}

	if slices.Contains(updateMaskPaths, "last_name") {
		rules = append(rules, validation.Field(&user.LastName, validation.Length(0, 255)))
	}

	if slices.Contains(updateMaskPaths, "password") {
		rules = append(rules, validation.Field(&user.Password, validation.Required, validation.By(checkPassword)))
	}

	// Validate the user struct based on the dynamically built rules
	return validation.ValidateStruct(user, rules...)
}

func (server *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	err := validateCreateSessionRequest(req)
	if err != nil {
		return nil, err
	}

	user, err := models.Users(models.UserWhere.Email.EQ(req.GetEmail())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Print(fmt.Sprintf("User %s not found", req.GetEmail()))
			return nil, status.Errorf(codes.Unauthenticated, "incorrect email or password")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	err = util.CheckPassword(req.GetPassword(), user.Password)
	if err != nil {
		log.Print(fmt.Sprintf("User %s password incorrect", req.GetEmail()))
		return nil, status.Errorf(codes.Unauthenticated, "incorrect email or password")
	}

	if !user.Active {
		return nil, status.Errorf(codes.PermissionDenied, "user is not active")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %s", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.RefreshTokenDuration,
	)

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
		return nil, status.Errorf(codes.Internal, "failed to insert session: %s", err)
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

func validateCreateSessionRequest(req *pb.CreateSessionRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
	)
}

func (server *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	err := validation.ValidateStruct(req, validation.Field(&req.RefreshToken, validation.Required))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err)
	}

	session, err := models.Sessions(models.SessionWhere.RefreshToken.EQ(req.GetRefreshToken())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "session not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get session: %s", err)
	}

	if session.IsBlocked {
		return nil, status.Errorf(codes.PermissionDenied, "session is blocked")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		session.UserID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %s", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		session.UserID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %s", err)
	}

	session.RefreshToken = refreshToken
	session.ExpiresAt = refreshPayload.ExpireTime
	_, err = session.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update session: %s", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:            accessToken,
		AccessTokenExpireTime:  timestamppb.New(accessPayload.ExpireTime),
		RefreshToken:           refreshToken,
		RefreshTokenExpireTime: timestamppb.New(refreshPayload.ExpireTime),
	}, nil

}

func (server *Server) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	if err := validation.ValidateStruct(req, validation.Field(&req.RefreshToken, validation.Required)); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err)
	}

	if req.GetRefreshToken() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	session, err := models.Sessions(models.SessionWhere.RefreshToken.EQ(req.GetRefreshToken())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "session not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get session: %s", err)
	}

	_, err = session.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete session: %s", err)
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

func checkPassword(value interface{}) error {
	password, _ := value.(string)
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`\d`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}
	if len(password) < 10 || len(password) > 255 {
		return fmt.Errorf("password must be between 10 and 255 characters")
	}
	return nil
}

func (server *Server) ValidateEmail(ctx context.Context, req *pb.ValidateEmailRequest) (*emptypb.Empty, error) {
	if err := validateValidateEmail(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err)
	}

	activation, err := models.AccountActivations(
		models.AccountActivationWhere.UserEmail.EQ(req.GetEmail()),
		models.AccountActivationWhere.ActivationToken.EQ(int(req.GetValidationNumber())),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "account activation not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get account activation: %s", err)
	}

	user, err := models.Users(models.UserWhere.ID.EQ(activation.UserID)).One(ctx, server.config.DB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %s", err)
	}

	user.Active = true
	_, err = user.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	_, err = activation.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete account activation: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func validateValidateEmail(req *pb.ValidateEmailRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.ValidationNumber, validation.Required, validation.By(
			func(value interface{}) error {
				validationNumber := value.(int64)
				if validationNumber < 1000000 || validationNumber > 9999999 {
					return fmt.Errorf("validation number must be 7 digits")
				}
				return nil
			},
		)),
	)
}

func (server *Server) SendValidationEmail(ctx context.Context, req *pb.SendValidationEmailRequest) (*emptypb.Empty, error) {
	if err := validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err)
	}

	dbUser, err := models.Users(models.UserWhere.Email.EQ(req.GetEmail())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %s", err)
	}
	if dbUser.Active {
		return nil, status.Errorf(codes.PermissionDenied, "user is already active")
	}

	if activationsRequestCount, err := models.AccountActivations(
		models.AccountActivationWhere.UserEmail.EQ(req.GetEmail()),
	).Count(ctx, server.config.DB); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count account activations: %s", err)
	} else if activationsRequestCount > 5 {
		return nil, status.Errorf(codes.PermissionDenied, "too many requests")
	}

	accountActivation := &models.AccountActivation{
		UserID:          dbUser.ID,
		UserEmail:       dbUser.Email,
		ActivationToken: util.RandomInt(1000000, 9999999),
	}
	if err = accountActivation.Insert(ctx, server.config.DB, boil.Infer()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert account validation: %s", err)
	}

	if err := server.mailService.SendValidateEmail(dbUser.Email, dbUser.FirstName, dbUser.LastName.String, accountActivation.ActivationToken); err != nil {
		status.Errorf(codes.Internal, "failed to send email: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	userId, err := pbx.GetResourceIDByType(req.GetName(), pbx.RessourceTypeUsers)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ressource name: %s", err)
	}

	if userId != authPayload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "cannot get other user's info")
	}

	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return pbx.DbUserToProto(user), nil
}
