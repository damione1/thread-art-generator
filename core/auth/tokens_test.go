package auth

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashTokenStable(t *testing.T) {
	t.Parallel()
	a := HashToken("abc")
	b := HashToken("abc")
	require.Equal(t, a, b)
	require.Len(t, a, 64)
	_, err := hex.DecodeString(a)
	require.NoError(t, err)
	require.NotEqual(t, HashToken("abc"), HashToken("abd"))
}

func TestNewRawTokenUnpredictable(t *testing.T) {
	t.Parallel()
	a, err := newRawToken()
	require.NoError(t, err)
	b, err := newRawToken()
	require.NoError(t, err)
	require.NotEmpty(t, a)
	require.NotEqual(t, a, b)
}
