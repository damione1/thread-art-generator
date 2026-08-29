package storage

import (
	"testing"

	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/stretchr/testify/require"
)

func TestBucketConfigFromUtil(t *testing.T) {
	t.Parallel()
	got := BucketConfigFromUtil(util.Config{
		Storage: util.StorageConfig{
			Endpoint:       "http://minio:9000",
			Region:         "us-east-1",
			Bucket:         "thread-art",
			AccessKey:      "minioadmin",
			SecretKey:      "minioadmin",
			PublicBaseURL:  "http://localhost:9000/thread-art",
			ForcePathStyle: true,
			UseTLS:         false,
		},
	})
	require.Equal(t, "http://minio:9000", got.Endpoint)
	require.Equal(t, "thread-art", got.Bucket)
	require.Equal(t, "http://localhost:9000/thread-art", got.PublicBaseURL)
	require.True(t, got.ForcePathStyle)
	require.False(t, got.UseTLS)
}
