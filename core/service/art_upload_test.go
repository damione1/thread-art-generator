package service

import (
	"net/http"
	"testing"

	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFlattenPresignHeadersTakesFirstValue(t *testing.T) {
	t.Parallel()
	got := flattenPresignHeaders(http.Header{
		"Content-Type":   []string{"image/jpeg", "text/plain"},
		"Content-Length": []string{"12"},
		"Empty":          []string{},
	})
	require.Equal(t, "image/jpeg", got["Content-Type"])
	require.Equal(t, "12", got["Content-Length"])
	_, ok := got["Empty"]
	require.False(t, ok)
}

func TestValidateUploadedObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		info    *storage.ObjectInfo
		wantMsg string
	}{
		{name: "nil", wantMsg: "image not found"},
		{
			name:    "not image",
			info:    &storage.ObjectInfo{ContentType: "application/pdf", Size: 100},
			wantMsg: "not an image",
		},
		{
			name:    "too large",
			info:    &storage.ObjectInfo{ContentType: "image/jpeg", Size: maxArtImageBytes + 1},
			wantMsg: "exceeds 10MB",
		},
		{
			name: "empty content type allowed",
			info: &storage.ObjectInfo{ContentType: "", Size: 1},
		},
		{
			name: "jpeg at cap",
			info: &storage.ObjectInfo{ContentType: "image/jpeg", Size: maxArtImageBytes},
		},
		{
			name: "png",
			info: &storage.ObjectInfo{ContentType: "image/png", Size: 2048},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateUploadedObject(tt.info)
			if tt.wantMsg == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Equal(t, codes.FailedPrecondition, status.Code(err))
			require.Contains(t, err.Error(), tt.wantMsg)
		})
	}
}
