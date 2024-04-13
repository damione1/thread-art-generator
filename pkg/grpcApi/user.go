package grpcApi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"

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

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if err := validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	user := &models.User{
		Email:    req.GetEmail(),
		Password: hashedPassword,
		Name:     req.GetName(),
	}

	err = user.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return nil, status.Errorf(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to insert user: %s", err)
	}

	pbUser := pbx.DbUserToProto(user)

	return &pb.CreateUserResponse{
		User: pbUser,
	}, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) error {
	return validation.ValidateStruct(req,
		// Name cannot be empty, and the length must be between 5 and 20
		validation.Field(&req.Name, validation.Required, validation.Length(5, 20)),
		// Email cannot be empty and should be in a valid email format
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required, validation.Length(8, 100), validation.By(checkPassword)),
	)
}

// update user
func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err = validateUpdateUserRequest(req); err != nil {
		return nil, err
	}

	pbUser := req.GetUser()

	if authPayload.UserID != pbUser.GetId() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	user, err := models.Users(models.UserWhere.Email.EQ(pbUser.GetEmail())).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if pbUser.GetPassword() == "" {
		hashedPassword, err := util.HashPassword(pbUser.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
		}
		user.Password = hashedPassword
	}

	if pbUser.GetName() != "" {
		user.Name = pbUser.GetName()
	}

	if pbUser.GetEmail() != "" {
		user.Email = pbUser.GetEmail()
	}

	_, err = user.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	return &pb.UpdateUserResponse{
		User: pbx.DbUserToProto(user),
	}, nil

}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.User, validation.Required, validation.By(
			func(value interface{}) error {
				user := value.(*pb.User)
				return validation.ValidateStruct(user,
					validation.Field(&user.Email, validation.Required, is.Email),
					validation.Field(&user.Name, validation.Required),
					validation.Field(&user.Password, validation.Length(8, 100), validation.Match(regexp.MustCompile(`^(?=.*[A-Z].*[A-Z])(?=.*[!@#$&*])(?=.*[0-9].*[0-9])(?=.*[a-z].*[a-z].*[a-z]).{8}$`))),
				)
			},
		)),
	)
}

// login user
func (server *Server) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	err := validateLoginUserRequest(req)
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

	fmt.Println("server.config.AccessTokenDuration", server.config.AccessTokenDuration)

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

	return &pb.LoginResponse{
		User:                   pbx.DbUserToProto(user),
		SessionId:              session.ID,
		AccessToken:            accessToken,
		RefreshToken:           refreshToken,
		AccessTokenExpireTime:  timestamppb.New(accessPayload.ExpireTime),
		RefreshTokenExpireTime: timestamppb.New(refreshPayload.ExpireTime),
	}, nil

}

// refresh token service
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

// logout
func (server *Server) LogoutUser(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {

	err := validation.ValidateStruct(req, validation.Field(&req.RefreshToken, validation.Required))
	if err != nil {
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

	return &pb.LogoutResponse{
		Success: true,
	}, nil
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

// func validateCreateUserRequest(req *pb.CreateUserRequest) error {
// 	return validation.ValidateStruct(req,
// 		// Name cannot be empty, and the length must be between 5 and 20
// 		validation.Field(&req.Name, validation.Required, validation.Length(5, 20)),
// 		// Email cannot be empty and should be in a valid email format
// 		validation.Field(&req.Email, validation.Required, is.Email),
// 		validation.Field(&req.Password, validation.Required, validation.Length(8, 100), validation.By(checkPassword)),
// 	)
// }

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
	return nil
}

func validateLoginUserRequest(req *pb.LoginRequest) error {
	return validation.ValidateStruct(req,
		// Email cannot be empty and should be in a valid email format
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
	)
}
