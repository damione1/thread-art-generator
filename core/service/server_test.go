package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPostgresUserID(t *testing.T) {
	t.Parallel()
	require.True(t, isPostgresUserID("11111111-1111-1111-1111-111111111111"))
	require.False(t, isPostgresUserID("firebaseUidNotAUUID"))
	require.False(t, isPostgresUserID(""))
}
