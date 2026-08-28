package interceptors

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/middleware"
)

// IdentityInterceptor fills auth.Identity from Service HMAC or a session cookie.
//
// Dual-run with PasetoAuthMiddleware (phase C drops PASETO):
//   - Authorization: Service … → ServiceAuth (fail closed)
//   - else cookie via Sessions.LoadFromCookie if Sessions != nil
//   - else Bearer → ignored here (PASETO interceptor still validates)
//
// Health procedures are skipped. SyncUserFromFirebase / ConfirmArtImageUploadFromFunction
// stay on the old PASETO whitelist until phase C — not this interceptor's job.
//
// When stacking, run this BEFORE PasetoAuthMiddleware: the PASETO interceptor
// rejects non-Bearer schemes, so Service HMAC would 401 if PASETO ran first.
func IdentityInterceptor(sessions auth.Sessions, services auth.ServiceAuth) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx, err := resolveIdentity(ctx, req.Spec().Procedure, req.Header(), sessions, services)
			if err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}

func resolveIdentity(ctx context.Context, procedure string, header http.Header, sessions auth.Sessions, services auth.ServiceAuth) (context.Context, error) {
	if skipIdentity(procedure) {
		return ctx, nil
	}

	authz := header.Get("Authorization")
	if isServiceAuthorization(authz) {
		if services == nil {
			return ctx, unauthenticated(auth.ErrInvalidServiceCred)
		}
		id, err := services.Authorize(ctx, authz)
		if err != nil {
			return ctx, unauthenticated(err)
		}
		return attachIdentity(ctx, id), nil
	}

	if sessions != nil {
		sess, err := sessions.LoadFromCookie(ctx, &http.Request{Header: header})
		if err == nil && sess.UserID != "" {
			return attachIdentity(ctx, auth.Identity{
				UserID: sess.UserID,
				Email:  sess.Email,
				Kind:   auth.PrincipalUser,
			}), nil
		}
	}

	if isBearerAuthorization(authz) {
		// PASETO still lives in PasetoAuthMiddleware. Do not validate here.
		return ctx, nil
	}

	return ctx, unauthenticated(errors.New("authorization token is not provided"))
}

func attachIdentity(ctx context.Context, id auth.Identity) context.Context {
	ctx = auth.WithIdentity(ctx, id)
	// Leftover service code (CreateArt) still reads middleware.AuthKey.
	ctx = context.WithValue(ctx, middleware.AuthKey, id.UserID)
	return ctx
}

func skipIdentity(procedure string) bool {
	if strings.HasSuffix(procedure, "Health") {
		return true
	}
	if i := strings.LastIndex(procedure, "/"); i > 0 && strings.HasSuffix(procedure[:i], "Health") {
		return true
	}
	switch procedure {
	case "/pb.ArtGeneratorService/SyncUserFromFirebase":
		return true
	}
	return false
}

func isServiceAuthorization(header string) bool {
	scheme, _, ok := strings.Cut(header, " ")
	return ok && strings.EqualFold(scheme, "Service")
}

func isBearerAuthorization(header string) bool {
	scheme, _, ok := strings.Cut(header, " ")
	return ok && strings.EqualFold(scheme, "Bearer")
}

func unauthenticated(err error) error {
	if err == nil {
		err = errors.New("unauthenticated")
	}
	return connect.NewError(connect.CodeUnauthenticated, err)
}
