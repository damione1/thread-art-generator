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
	require.Equal(t, 587, c.SMTP.Port)
	require.Equal(t, "none", c.SMTP.TLSMode)
	require.Equal(t, "ThreadArt", c.SMTP.FromName)
	require.Equal(t, "noreply@localhost", c.SMTP.FromAddr)
	require.Equal(t, "http://localhost:8080", c.FrontendUrl)
}

func TestApplyDefaultsDoesNotOverrideHMAC(t *testing.T) {
	t.Parallel()
	c := Config{ServiceHMACSecret: "explicit-hmac-secret-32-bytes!!"}
	c.applyDefaults()
	require.Equal(t, "explicit-hmac-secret-32-bytes!!", c.ServiceHMACSecret)
}
