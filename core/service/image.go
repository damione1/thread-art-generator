package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"strings"

	_ "image/gif"
	_ "image/png"

	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/disintegration/imaging"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	_ "golang.org/x/image/webp"
)

const hoopOriginalEdge = 1600

var (
	magicJPEG = []byte{0xff, 0xd8, 0xff}
	magicPNG  = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	magicGIF  = []byte("GIF8")
	magicRIFF = []byte("RIFF")
	magicWEBP = []byte("WEBP")
)

func validateUploadedObject(info *storage.ObjectInfo, magic []byte) error {
	if info == nil {
		return pbErrors.FailedPreconditionError("image not found in storage, upload first")
	}
	if info.Size > maxArtImageBytes {
		return pbErrors.FailedPreconditionError("uploaded object exceeds 10MB")
	}
	ct := strings.ToLower(strings.TrimSpace(info.ContentType))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct == "" {
		return pbErrors.FailedPreconditionError("uploaded object is missing a content type")
	}
	kind, ok := allowedImageContentType(ct)
	if !ok {
		return pbErrors.FailedPreconditionError("uploaded object is not an allowed image type")
	}
	if !magicMatches(kind, magic) {
		return pbErrors.FailedPreconditionError("uploaded object does not match its content type")
	}
	return nil
}

func requireAllowedImageContentType(contentType string) error {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct == "" {
		return pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("content_type", errors.New("content type is required")),
		})
	}
	if _, ok := allowedImageContentType(ct); !ok {
		return pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("content_type", errors.New("not an allowed image type")),
		})
	}
	return nil
}

func allowedImageContentType(ct string) (string, bool) {
	switch ct {
	case "image/jpeg", "image/jpg":
		return "jpeg", true
	case "image/png":
		return "png", true
	case "image/gif":
		return "gif", true
	case "image/webp":
		return "webp", true
	default:
		return "", false
	}
}

func magicMatches(kind string, magic []byte) bool {
	switch kind {
	case "jpeg":
		return bytes.HasPrefix(magic, magicJPEG)
	case "png":
		return bytes.HasPrefix(magic, magicPNG)
	case "gif":
		return bytes.HasPrefix(magic, magicGIF)
	case "webp":
		return len(magic) >= 12 && bytes.HasPrefix(magic, magicRIFF) && bytes.Equal(magic[8:12], magicWEBP)
	default:
		return false
	}
}

// storeHoopOriginal validates the uploaded object, then overwrites it with a
// square circular grayscale JPEG. Contrast stays a composition-time knob.
func storeHoopOriginal(ctx context.Context, bucket storage.Bucket, key string) error {
	reader, info, err := bucket.Get(ctx, key)
	if err != nil {
		return pbErrors.FailedPreconditionError("image not found in storage, upload first")
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxArtImageBytes+1))
	if err != nil {
		return pbErrors.InternalError("failed to read uploaded image", err)
	}
	if int64(len(data)) > maxArtImageBytes {
		return pbErrors.FailedPreconditionError("uploaded object exceeds 10MB")
	}
	if info == nil {
		info = &storage.ObjectInfo{}
	}
	info.Size = int64(len(data))
	magic := data
	if len(magic) > imageSniffBytes {
		magic = magic[:imageSniffBytes]
	}
	if err := validateUploadedObject(info, magic); err != nil {
		return err
	}
	jpegBytes, err := hoopOriginalJPEG(data)
	if err != nil {
		return pbErrors.FailedPreconditionError("uploaded object is not a valid image")
	}
	if err := bucket.Put(ctx, key, bytes.NewReader(jpegBytes), storage.PutOptions{ContentType: "image/jpeg"}); err != nil {
		return pbErrors.InternalError("failed to store processed image", err)
	}
	return nil
}

func hoopOriginalJPEG(src []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 1 || h < 1 {
		return nil, errors.New("empty image")
	}
	side := w
	if h < side {
		side = h
	}
	square := imaging.CropAnchor(img, side, side, imaging.Center)
	if side > hoopOriginalEdge {
		square = imaging.Resize(square, hoopOriginalEdge, hoopOriginalEdge, imaging.Lanczos)
	}
	gray := imaging.Clone(imaging.Grayscale(square))
	maskCircleWhite(gray)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, gray, &jpeg.Options{Quality: 92}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func maskCircleWhite(img *image.NRGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	midX, midY := w/2, h/2
	r2 := midX * midX
	if midY*midY < r2 {
		r2 = midY * midY
	}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 0; y < h; y++ {
		dy := y - midY
		for x := 0; x < w; x++ {
			dx := x - midX
			if dx*dx+dy*dy > r2 {
				img.SetNRGBA(b.Min.X+x, b.Min.Y+y, white)
			}
		}
	}
}
