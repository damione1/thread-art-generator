package service

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
			require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
			require.Contains(t, err.Error(), tt.wantMsg)
		})
	}
}

func TestRequireParentIdentity(t *testing.T) {
	t.Parallel()
	uid := "11111111-1111-1111-1111-111111111111"
	require.NoError(t, requireParentIdentity("users/"+uid, uid))
	err := requireParentIdentity("users/other", uid)
	require.True(t, connect.CodeOf(err) == connect.CodePermissionDenied)
	err = requireParentIdentity("users/"+uid+"/arts/x", uid)
	require.True(t, connect.CodeOf(err) == connect.CodeInvalidArgument)
}

func TestApplyArtUpdateMask(t *testing.T) {
	t.Parallel()
	src := &pb.Art{Title: "New"}
	dst := &models.Art{Title: "Old"}
	cols, err := applyArtUpdateMask(&fieldmaskpb.FieldMask{Paths: []string{"title"}}, src, dst)
	require.NoError(t, err)
	require.Equal(t, []string{models.ArtColumns.Title}, cols)
	require.Equal(t, "New", dst.Title)

	_, err = applyArtUpdateMask(&fieldmaskpb.FieldMask{Paths: []string{"status"}}, src, dst)
	require.True(t, connect.CodeOf(err) == connect.CodeInvalidArgument)

	dst.Title = "Old"
	cols, err = applyArtUpdateMask(&fieldmaskpb.FieldMask{Paths: []string{"*"}}, src, dst)
	require.NoError(t, err)
	require.Equal(t, []string{models.ArtColumns.Title}, cols)
}

func TestPresignAndHeadWithMemoryBucket(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := storage.NewMemoryBucket()
	key := "users/u/arts/a/original"

	resp, err := presignArtOriginal(ctx, b, key, "image/jpeg")
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, resp.Method)
	require.Equal(t, "image/jpeg", resp.Headers["Content-Type"])
	require.Contains(t, resp.UploadUrl, key)

	err = headUploadedOriginal(ctx, b, key)
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))

	require.NoError(t, b.Put(ctx, key, bytes.NewReader([]byte("jpeg-bytes")), storage.PutOptions{ContentType: "image/jpeg"}))
	require.NoError(t, headUploadedOriginal(ctx, b, key))
}

func TestHeadUploadedOriginalRejectsNonImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := storage.NewMemoryBucket()
	require.NoError(t, b.Put(ctx, "k", bytes.NewReader([]byte("%PDF")), storage.PutOptions{ContentType: "application/pdf"}))
	err := headUploadedOriginal(ctx, b, "k")
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "not an image")
}

func TestSignCompositionDownloadsPresignsKeys(t *testing.T) {
	t.Parallel()
	s := &Server{bucket: storage.NewMemoryBucket()}
	c := &pb.Composition{
		GcodeUrl:    "users/u/arts/a/compositions/c/gcode",
		PathlistUrl: "users/u/arts/a/compositions/c/pathlist",
		PreviewUrl:  "http://localhost:9000/thread-art/preview.png",
	}
	require.NoError(t, s.signCompositionDownloads(context.Background(), c))
	require.Contains(t, c.GcodeUrl, "memory.local/get/")
	require.Contains(t, c.PathlistUrl, "memory.local/get/")
	require.Equal(t, "http://localhost:9000/thread-art/preview.png", c.PreviewUrl)
}

func TestSignCompositionDownloadsNilSafe(t *testing.T) {
	t.Parallel()
	s := &Server{}
	require.NoError(t, s.signCompositionDownloads(context.Background(), nil))
	require.NoError(t, s.signCompositionDownloads(context.Background(), &pb.Composition{}))
}
