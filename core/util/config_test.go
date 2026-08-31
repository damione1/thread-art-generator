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

func TestIsDevelopment(t *testing.T) {
	t.Parallel()
	require.True(t, IsDevelopment(""))
	require.True(t, IsDevelopment("development"))
	require.True(t, IsDevelopment("Development"))
	require.False(t, IsDevelopment("production"))
	require.False(t, IsDevelopment("staging"))
}

func TestApplyDefaultsDoesNotOverrideHMAC(t *testing.T) {
	t.Parallel()
	c := Config{ServiceHMACSecret: "explicit-hmac-secret-32-bytes!!"}
	c.applyDefaults()
	require.Equal(t, "explicit-hmac-secret-32-bytes!!", c.ServiceHMACSecret)
}

func TestLoadConfigSMTPFromEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "mailhog")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_FROM_ADDRESS", "noreply@threadart.local")
	t.Setenv("SMTP_TLS_MODE", "none")
	t.Setenv("SMTP_FROM_NAME", "ThreadArt")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "mailhog", cfg.SMTP.Host)
	require.Equal(t, 1025, cfg.SMTP.Port)
	require.Equal(t, "none", cfg.SMTP.TLSMode)
	require.Equal(t, "noreply@threadart.local", cfg.SMTP.FromAddr)
}

func TestLoadConfigRembgURL(t *testing.T) {
	t.Setenv("REMBG_URL", "http://rembg:7000")
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "http://rembg:7000", cfg.RembgURL)
}
