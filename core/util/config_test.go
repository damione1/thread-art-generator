package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyDefaultsStorage(t *testing.T) {
	t.Parallel()
	c := Config{}
	c.applyDefaults()
	require.Equal(t, "postgres", c.PostgresUser)
	require.Equal(t, "thread-art", c.Storage.Bucket)
	require.Equal(t, "http://localhost:9000/thread-art", c.Storage.PublicBaseURL)
	require.False(t, c.Storage.ForcePathStyle)
	require.Empty(t, c.Storage.Endpoint)
}

func TestApplyDefaultsDoesNotOverrideHMAC(t *testing.T) {
	t.Parallel()
	c := Config{ServiceHMACSecret: "explicit-hmac-secret-32-bytes!!"}
	c.applyDefaults()
	require.Equal(t, "explicit-hmac-secret-32-bytes!!", c.ServiceHMACSecret)
}
