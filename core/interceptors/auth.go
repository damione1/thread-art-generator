package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func authorizeUser(ctx context.Context, tokenMaker token.Maker) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", authType)
	}

	accessToken := fields[1]
	payload, err := tokenMaker.ValidateToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}

func AuthInterceptor(tokenMaker token.Maker) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Skip authentication if the method is in the white list
		if isWhiteListedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		payload, err := authorizeUser(ctx, tokenMaker)
		if err != nil {
			return nil, err
		}

		adminCtx := middleware.NewAdminContext(ctx, payload)
		return handler(adminCtx, req)
	}
}

func isWhiteListedMethod(method string) bool {
	for _, m := range whiteListedMethods {
		if m == method {
			return true
		}
	}
	return false
}

var whiteListedMethods = []string{
	"/pb.ArtGeneratorService/GetMediaUploadUrl",
	"/pb.ArtGeneratorService/GetMediaDownloadUrl",
	"/pb.ArtGeneratorService/CreateSession",
	"/pb.ArtGeneratorService/RefreshToken",
	"/pb.ArtGeneratorService/DeleteSession",
	"/pb.ArtGeneratorService/ResetPassword",
	"/pb.ArtGeneratorService/ValidateEmail",
	"/pb.ArtGeneratorService/SendValidationEmail",
}
