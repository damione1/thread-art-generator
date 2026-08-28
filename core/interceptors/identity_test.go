package interceptors

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/middleware"
)

type dummyMsg struct{}

type fakeSessions struct {
	sess auth.Session
	err  error
}

func (f *fakeSessions) Issue(context.Context, http.ResponseWriter, *http.Request, auth.Session) error {
	return nil
}
func (f *fakeSessions) Load(context.Context, *http.Request) (auth.Session, error) {
	return f.sess, f.err
}
func (f *fakeSessions) Destroy(context.Context, http.ResponseWriter, *http.Request) error {
	return nil
}
func (f *fakeSessions) LoadFromCookie(context.Context, *http.Request) (auth.Session, error) {
	return f.sess, f.err
}

func TestIdentityInterceptorServiceHeader(t *testing.T) {
	svc, err := auth.NewHMACServiceAuth(strings.Repeat("k", 32))
	require.NoError(t, err)
	hdr, err := svc.Sign("worker-1")
	require.NoError(t, err)

	var saw auth.Identity
	var sawOK bool
	var legacy string
	next := connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		saw, sawOK = auth.IdentityFromContext(ctx)
		legacy, _ = ctx.Value(middleware.AuthKey).(string)
		return connect.NewResponse(&dummyMsg{}), nil
	})

	req := connect.NewRequest(&dummyMsg{})
	req.Header().Set("Authorization", hdr)

	_, err = IdentityInterceptor(nil, svc)(next)(context.Background(), req)
	require.NoError(t, err)
	require.True(t, sawOK)
	require.Equal(t, "worker-1", saw.UserID)
	require.Equal(t, auth.PrincipalService, saw.Kind)
	require.Equal(t, "worker-1", legacy)
}

func TestIdentityInterceptorTamperedService(t *testing.T) {
	svc, err := auth.NewHMACServiceAuth(strings.Repeat("k", 32))
	require.NoError(t, err)
	hdr, err := svc.Sign("worker-1")
	require.NoError(t, err)
	tampered := hdr[:len(hdr)-2] + "00"

	called := false
	next := connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		called = true
		return connect.NewResponse(&dummyMsg{}), nil
	})
	req := connect.NewRequest(&dummyMsg{})
	req.Header().Set("Authorization", tampered)

	_, err = IdentityInterceptor(nil, svc)(next)(context.Background(), req)
	require.Error(t, err)
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	require.False(t, called)
}

func TestResolveIdentityCookie(t *testing.T) {
	ctx, err := resolveIdentity(context.Background(), "/pb.ArtGeneratorService/CreateArt", http.Header{},
		&fakeSessions{sess: auth.Session{UserID: "user-uuid", Email: "a@b.c"}}, nil)
	require.NoError(t, err)
	id, ok := auth.IdentityFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "user-uuid", id.UserID)
	require.Equal(t, "a@b.c", id.Email)
	require.Equal(t, auth.PrincipalUser, id.Kind)
	legacy, _ := ctx.Value(middleware.AuthKey).(string)
	require.Equal(t, "user-uuid", legacy)
}

func TestResolveIdentityBearerPassthrough(t *testing.T) {
	h := make(http.Header)
	h.Set("Authorization", "Bearer paseto-still-elsewhere")
	ctx, err := resolveIdentity(context.Background(), "/pb.ArtGeneratorService/CreateArt", h, nil, nil)
	require.NoError(t, err)
	_, ok := auth.IdentityFromContext(ctx)
	require.False(t, ok)
}

func TestResolveIdentityMissingAuth(t *testing.T) {
	_, err := resolveIdentity(context.Background(), "/pb.ArtGeneratorService/CreateArt", http.Header{}, nil, nil)
	require.Error(t, err)
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

func TestSkipIdentity(t *testing.T) {
	require.True(t, skipIdentity("/grpc.health.v1.Health/Check"))
	require.True(t, skipIdentity("/grpc.health.v1.Health/Watch"))
	require.True(t, skipIdentity("/connectrpc.health.v1.Health"))
	require.False(t, skipIdentity("/pb.ArtGeneratorService/CreateArt"))
	require.True(t, skipIdentity("/pb.ArtGeneratorService/SyncUserFromFirebase"))
	require.False(t, skipIdentity("/pb.FirebaseFunctionsService/ConfirmArtImageUploadFromFunction"))
}

func TestResolveIdentityHealthSkip(t *testing.T) {
	ctx, err := resolveIdentity(context.Background(), "/grpc.health.v1.Health/Check", http.Header{}, nil, nil)
	require.NoError(t, err)
	_, ok := auth.IdentityFromContext(ctx)
	require.False(t, ok)
}
