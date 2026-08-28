package id

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUUIDNewIsV4(t *testing.T) {
	t.Parallel()
	var g UUID
	parsed, err := uuid.Parse(g.New())
	require.NoError(t, err)
	require.Equal(t, uuid.Version(4), parsed.Version())
}
