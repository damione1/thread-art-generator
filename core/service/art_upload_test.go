package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
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
	jpeg := []byte{0xff, 0xd8, 0xff, 0xe0}
	png := append([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}, 0, 0)
	gif := []byte("GIF89a....")
	webp := []byte("RIFF....WEBP....")
	tests := []struct {
		name    string
		info    *storage.ObjectInfo
		magic   []byte
		wantMsg string
	}{
		{name: "nil", wantMsg: "image not found"},
		{
			name:    "not image",
			info:    &storage.ObjectInfo{ContentType: "application/pdf", Size: 100},
			magic:   []byte("%PDF"),
			wantMsg: "not an allowed image type",
		},
		{
			name:    "svg rejected",
			info:    &storage.ObjectInfo{ContentType: "image/svg+xml", Size: 100},
			magic:   []byte("<svg"),
			wantMsg: "not an allowed image type",
		},
		{
			name:    "empty content type",
			info:    &storage.ObjectInfo{ContentType: "", Size: 1},
			magic:   jpeg,
			wantMsg: "missing a content type",
		},
		{
			name:    "too large",
			info:    &storage.ObjectInfo{ContentType: "image/jpeg", Size: maxArtImageBytes + 1},
			magic:   jpeg,
			wantMsg: "exceeds 10MB",
		},
		{
			name:    "magic mismatch",
			info:    &storage.ObjectInfo{ContentType: "image/jpeg", Size: 12},
			magic:   png,
			wantMsg: "does not match",
		},
		{
			name:  "jpeg at cap",
			info:  &storage.ObjectInfo{ContentType: "image/jpeg", Size: maxArtImageBytes},
			magic: jpeg,
		},
		{
			name:  "png",
			info:  &storage.ObjectInfo{ContentType: "image/png", Size: 2048},
			magic: png,
		},
		{
			name:  "gif",
			info:  &storage.ObjectInfo{ContentType: "image/gif", Size: 2048},
			magic: gif,
		},
		{
			name:  "webp",
			info:  &storage.ObjectInfo{ContentType: "image/webp", Size: 2048},
			magic: webp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateUploadedObject(tt.info, tt.magic)
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

func TestRequireAllowedImageContentType(t *testing.T) {
	t.Parallel()
	require.NoError(t, requireAllowedImageContentType("image/jpeg"))
	require.NoError(t, requireAllowedImageContentType("image/png; charset=binary"))
	err := requireAllowedImageContentType("")
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	err = requireAllowedImageContentType("application/pdf")
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
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

	err = inspectUploadedOriginal(ctx, b, key)
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))

	jpeg := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10}
	require.NoError(t, b.Put(ctx, key, bytes.NewReader(jpeg), storage.PutOptions{ContentType: "image/jpeg"}))
	require.NoError(t, inspectUploadedOriginal(ctx, b, key))
}

func TestHeadUploadedOriginalRejectsNonImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := storage.NewMemoryBucket()
	require.NoError(t, b.Put(ctx, "k", bytes.NewReader([]byte("%PDF")), storage.PutOptions{ContentType: "application/pdf"}))
	err := inspectUploadedOriginal(ctx, b, "k")
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "not an allowed image type")
}

func TestInspectUploadedOriginalRejectsMagicMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := storage.NewMemoryBucket()
	require.NoError(t, b.Put(ctx, "k", bytes.NewReader([]byte("%PDF-1.4")), storage.PutOptions{ContentType: "image/jpeg"}))
	err := inspectUploadedOriginal(ctx, b, "k")
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "does not match")
}

func TestSignCompositionDownloadsPresignsKeys(t *testing.T) {
	t.Parallel()
	s := &Server{bucket: storage.NewMemoryBucket()}
	c := &pb.Composition{
		GcodeUrl:    "users/u/arts/a/compositions/c/gcode",
		PathlistUrl: "users/u/arts/a/compositions/c/pathlist",
		PreviewUrl:  "users/u/arts/a/compositions/c/preview",
	}
	require.NoError(t, s.signCompositionDownloads(context.Background(), c))
	require.Contains(t, c.GcodeUrl, "memory.local/get/")
	require.Contains(t, c.PathlistUrl, "memory.local/get/")
	require.Contains(t, c.PreviewUrl, "memory.local/get/")
}

func TestSignArtDownloadsPresignsKey(t *testing.T) {
	t.Parallel()
	s := &Server{bucket: storage.NewMemoryBucket()}
	art := &pb.Art{ImageUrl: "users/u/arts/a/original"}
	require.NoError(t, s.signArtDownloads(context.Background(), art))
	require.Contains(t, art.ImageUrl, "memory.local/get/")
}

func TestSignCompositionDownloadsNilSafe(t *testing.T) {
	t.Parallel()
	s := &Server{}
	require.NoError(t, s.signCompositionDownloads(context.Background(), nil))
	require.NoError(t, s.signCompositionDownloads(context.Background(), &pb.Composition{}))
}

func TestHoopOriginalJPEGGrayscaleCircle(t *testing.T) {
	t.Parallel()
	src := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	red := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.SetNRGBA(x, y, red)
		}
	}
	var pngBuf bytes.Buffer
	require.NoError(t, png.Encode(&pngBuf, src))

	out, err := hoopOriginalJPEG(pngBuf.Bytes())
	require.NoError(t, err)
	img, err := jpeg.Decode(bytes.NewReader(out))
	require.NoError(t, err)

	center := color.NRGBAModel.Convert(img.At(4, 4)).(color.NRGBA)
	require.Equal(t, center.R, center.G)
	require.Equal(t, center.G, center.B)
	require.InDelta(t, 76, center.R, 8)

	corner := color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA)
	require.Greater(t, int(corner.R), 250)
	require.Equal(t, corner.R, corner.G)
	require.Equal(t, corner.G, corner.B)
}

func TestStoreHoopOriginalOverwritesColorUpload(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := storage.NewMemoryBucket()
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}
	var pngBuf bytes.Buffer
	require.NoError(t, png.Encode(&pngBuf, src))
	require.NoError(t, b.Put(ctx, "k", bytes.NewReader(pngBuf.Bytes()), storage.PutOptions{ContentType: "image/png"}))

	require.NoError(t, storeHoopOriginal(ctx, b, "k"))

	reader, info, err := b.Get(ctx, "k")
	require.NoError(t, err)
	defer reader.Close()
	require.Equal(t, "image/jpeg", info.ContentType)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}))

	img, err := jpeg.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	center := color.NRGBAModel.Convert(img.At(2, 2)).(color.NRGBA)
	require.Equal(t, center.R, center.G)
	require.Equal(t, center.G, center.B)
}
