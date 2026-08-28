package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyDefaultsQueueAndHMAC(t *testing.T) {
	t.Parallel()
	c := Config{TokenSymmetricKey: "token-symmetric-key-32-bytes!!!!"}
	c.applyDefaults()
	require.Equal(t, "postgres", c.QueueProvider)
	require.Equal(t, c.TokenSymmetricKey, c.ServiceHMACSecret)
	require.Equal(t, "local-public", c.Storage.PublicBucket)
	require.Equal(t, "local-private", c.Storage.PrivateBucket)
}

func TestApplyDefaultsDoesNotOverrideQueueProvider(t *testing.T) {
	t.Parallel()
	c := Config{QueueProvider: "rabbitmq", ServiceHMACSecret: "explicit-hmac-secret-32-bytes!!"}
	c.applyDefaults()
	require.Equal(t, "rabbitmq", c.QueueProvider)
	require.Equal(t, "explicit-hmac-secret-32-bytes!!", c.ServiceHMACSecret)
}
