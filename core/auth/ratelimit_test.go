package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthLimiterAllowsThenBlocks(t *testing.T) {
	t.Parallel()
	now := time.Now()
	l := NewAuthLimiter(time.Minute, 2)
	l.now = func() time.Time { return now }

	require.True(t, l.Allow("login:1.1.1.1:a@b.c"))
	require.True(t, l.Allow("login:1.1.1.1:a@b.c"))
	require.False(t, l.Allow("login:1.1.1.1:a@b.c"))
	require.True(t, l.Allow("login:1.1.1.1:other@b.c"))
}

func TestAuthLimiterWindowExpires(t *testing.T) {
	t.Parallel()
	now := time.Now()
	l := NewAuthLimiter(time.Minute, 1)
	l.now = func() time.Time { return now }
	require.True(t, l.Allow("k"))
	require.False(t, l.Allow("k"))
	l.now = func() time.Time { return now.Add(time.Minute + time.Second) }
	require.True(t, l.Allow("k"))
}

func TestAuthLimiterNilAllows(t *testing.T) {
	t.Parallel()
	var l *AuthLimiter
	require.True(t, l.Allow("x"))
}
