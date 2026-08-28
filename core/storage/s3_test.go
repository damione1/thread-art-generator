package storage

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testConfig(mods ...func(*Config)) Config {
	cfg := Config{
		Endpoint:       "http://minio:9000",
		Region:         "us-east-1",
		Bucket:         "thread-art",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		PublicBaseURL:  "http://localhost:9000/thread-art",
		ForcePathStyle: true,
		UseTLS:         false,
	}
	for _, m := range mods {
		m(&cfg)
	}
	return cfg
}

func TestWithScheme(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		raw    string
		useTLS bool
		want   string
	}{
		{name: "host only http", raw: "minio:9000", useTLS: false, want: "http://minio:9000"},
		{name: "host only https", raw: "minio:9000", useTLS: true, want: "https://minio:9000"},
		{name: "already http", raw: "http://minio:9000", useTLS: true, want: "http://minio:9000"},
		{name: "already https", raw: "https://s3.amazonaws.com", useTLS: false, want: "https://s3.amazonaws.com"},
		{name: "trailing slash", raw: "http://minio:9000/", useTLS: false, want: "http://minio:9000"},
		{name: "empty", raw: "", useTLS: false, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, withScheme(tt.raw, tt.useTLS))
		})
	}
}

func TestSignerBaseEndpoint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "minio public strips bucket path",
			cfg:  testConfig(),
			want: "http://localhost:9000",
		},
		{
			name: "empty public uses ops",
			cfg:  testConfig(func(c *Config) { c.PublicBaseURL = "" }),
			want: "http://minio:9000",
		},
		{
			name: "https public keeps scheme",
			cfg: testConfig(func(c *Config) {
				c.PublicBaseURL = "https://cdn.example.com/thread-art"
				c.UseTLS = true
			}),
			want: "https://cdn.example.com",
		},
		{
			name: "host without scheme",
			cfg: testConfig(func(c *Config) {
				c.PublicBaseURL = "localhost:9000/thread-art"
				c.UseTLS = false
			}),
			want: "http://localhost:9000",
		},
		{
			name: "r2 api host",
			cfg: testConfig(func(c *Config) {
				c.Endpoint = "https://abc.r2.cloudflarestorage.com"
				c.PublicBaseURL = "https://abc.r2.cloudflarestorage.com"
				c.UseTLS = true
			}),
			want: "https://abc.r2.cloudflarestorage.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := signerBaseEndpoint(tt.cfg)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestObjectURLPathStyle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		base      string
		bucket    string
		key       string
		pathStyle bool
		want      string
	}{
		{
			name:      "minio path style",
			base:      "http://localhost:9000",
			bucket:    "thread-art",
			key:       "users/u/arts/a/original",
			pathStyle: true,
			want:      "http://localhost:9000/thread-art/users/u/arts/a/original",
		},
		{
			name:      "virtual hosted",
			base:      "https://s3.us-east-1.amazonaws.com",
			bucket:    "thread-art",
			key:       "users/u/arts/a/original",
			pathStyle: false,
			want:      "https://thread-art.s3.us-east-1.amazonaws.com/users/u/arts/a/original",
		},
		{
			name:      "public base is prefix not endpoint",
			base:      "http://localhost:9000",
			bucket:    "thread-art",
			key:       "k",
			pathStyle: true,
			want:      "http://localhost:9000/thread-art/k",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, objectURL(tt.base, tt.bucket, tt.key, tt.pathStyle))
		})
	}
}

func TestPresignTTL(t *testing.T) {
	t.Parallel()
	require.Equal(t, defaultPresignTTL, presignTTL(0))
	require.Equal(t, 5*time.Minute, presignTTL(5*time.Minute))
}

func TestNewS3BucketRequiresCreds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, err := NewS3Bucket(ctx, Config{Bucket: "thread-art"})
	require.Error(t, err)
	_, err = NewS3Bucket(ctx, Config{AccessKey: "a", SecretKey: "b"})
	require.Error(t, err)
}

func TestPresignPutIncludesContentType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, err := NewS3Bucket(ctx, testConfig())
	require.NoError(t, err)

	tests := []struct {
		name        string
		contentType string
		key         string
	}{
		{name: "jpeg", contentType: "image/jpeg", key: "users/u/arts/a/original"},
		{name: "png", contentType: "image/png", key: "users/u/arts/b/original"},
		{name: "webp", contentType: "image/webp", key: "users/u/arts/c/original"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := b.PresignPut(ctx, tt.key, PresignPutOptions{
				ContentType: tt.contentType,
				TTL:         10 * time.Minute,
			})
			require.NoError(t, err)
			require.Equal(t, http.MethodPut, out.Method)
			require.Equal(t, tt.contentType, out.Headers.Get("Content-Type"))
			require.Contains(t, strings.ToLower(out.URL), "x-amz-signature")
			require.Contains(t, out.URL, "localhost:9000")
			require.NotContains(t, out.URL, "minio:9000")
			require.Contains(t, out.URL, "thread-art")
			require.Contains(t, out.URL, tt.key)
			require.True(t, out.Expires.After(time.Now()))
			// SignedHeaders query must list content-type so the browser replay is required.
			require.Contains(t, strings.ToLower(out.URL), "content-type")
		})
	}
}

func TestPresignGetUsesSignerHost(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, err := NewS3Bucket(ctx, testConfig())
	require.NoError(t, err)
	url, err := b.PresignGet(ctx, "users/u/arts/a/original", 5*time.Minute)
	require.NoError(t, err)
	require.Contains(t, url, "localhost:9000")
	require.NotContains(t, url, "minio:9000")
	require.Contains(t, strings.ToLower(url), "x-amz-signature")
}

func TestReplayHeadersDropsHost(t *testing.T) {
	t.Parallel()
	signed := http.Header{}
	signed.Set("Host", "localhost:9000")
	signed.Set("Content-Type", "image/jpeg")
	h := replayHeaders(signed, "image/jpeg")
	require.Empty(t, h.Get("Host"))
	require.Equal(t, "image/jpeg", h.Get("Content-Type"))
}
