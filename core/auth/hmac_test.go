package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testSecret() string {
	return strings.Repeat("s", 32)
}

func TestNewHMACServiceAuthRejectsShortSecret(t *testing.T) {
	_, err := NewHMACServiceAuth(strings.Repeat("x", 31))
	require.ErrorIs(t, err, ErrInvalidServiceSecret)
}

func TestHMACServiceAuthRoundTrip(t *testing.T) {
	a, err := NewHMACServiceAuth(testSecret())
	require.NoError(t, err)

	header, err := a.Sign("worker-1")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(header, "Service worker-1:"))

	id, err := a.Authorize(context.Background(), header)
	require.NoError(t, err)
	require.Equal(t, "worker-1", id.UserID)
	require.Equal(t, PrincipalService, id.Kind)
	require.Empty(t, id.Email)
}

func TestHMACServiceAuthTamperedMACFails(t *testing.T) {
	a, err := NewHMACServiceAuth(testSecret())
	require.NoError(t, err)

	header, err := a.Sign("worker-1")
	require.NoError(t, err)

	_, rest, _ := strings.Cut(header, ":")
	mac, err := hex.DecodeString(rest)
	require.NoError(t, err)
	mac[0] ^= 0xff
	tampered := "Service worker-1:" + hex.EncodeToString(mac)

	_, err = a.Authorize(context.Background(), tampered)
	require.ErrorIs(t, err, ErrInvalidServiceCred)
}

func TestHMACServiceAuthWrongIDFails(t *testing.T) {
	a, err := NewHMACServiceAuth(testSecret())
	require.NoError(t, err)

	header, err := a.Sign("worker-1")
	require.NoError(t, err)
	_, mac, _ := strings.Cut(header, ":")

	_, err = a.Authorize(context.Background(), "Service worker-2:"+mac)
	require.ErrorIs(t, err, ErrInvalidServiceCred)
}

func TestHMACServiceAuthConstantTimeComparePath(t *testing.T) {
	secret := []byte(testSecret())
	id := "worker-1"
	good := hmacSHA256(secret, id)
	bad := hmacSHA256(secret, "worker-2")

	require.Equal(t, 1, subtle.ConstantTimeCompare(good, hmacSHA256(secret, id)))
	require.Equal(t, 0, subtle.ConstantTimeCompare(good, bad))
	require.Equal(t, sha256.Size, len(good))

	a, err := NewHMACServiceAuth(testSecret())
	require.NoError(t, err)
	_, err = a.Authorize(context.Background(), "Service worker-1:"+hex.EncodeToString(bad))
	require.ErrorIs(t, err, ErrInvalidServiceCred)
}

func TestHMACServiceAuthMalformedHeader(t *testing.T) {
	a, err := NewHMACServiceAuth(testSecret())
	require.NoError(t, err)

	for _, h := range []string{
		"",
		"Bearer abc",
		"Service",
		"Service ",
		"Service nocolon",
		"Service :deadbeef",
		"Service id:zzzz",
		"Service id with space:aa",
	} {
		_, err := a.Authorize(context.Background(), h)
		require.ErrorIs(t, err, ErrInvalidServiceCred, h)
	}
}

func TestHMACSignRejectsBadID(t *testing.T) {
	a, err := NewHMACServiceAuth(testSecret())
	require.NoError(t, err)
	_, err = a.Sign("")
	require.ErrorIs(t, err, ErrInvalidServiceID)
	_, err = a.Sign("has space")
	require.ErrorIs(t, err, ErrInvalidServiceID)
	_, err = a.Sign("has:colon")
	require.ErrorIs(t, err, ErrInvalidServiceID)
}

func hmacSHA256(secret []byte, msg string) []byte {
	m := hmac.New(sha256.New, secret)
	_, _ = m.Write([]byte(msg))
	return m.Sum(nil)
}
