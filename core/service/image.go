package service

import (
	"bytes"
	"errors"
	"strings"

	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/storage"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

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
