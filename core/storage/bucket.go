package storage

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Bucket is the cloud-agnostic object store. One implementation: S3 API
// (MinIO local, R2, AWS S3). No emulator branches. No public-URL logic.
//
// Public URLs are PublicBaseURL + "/" + key at the API layer, not here.
type Bucket interface {
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) error
	Get(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error)
	Head(ctx context.Context, key string) (*ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	PresignPut(ctx context.Context, key string, opts PresignPutOptions) (*PresignPut, error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// PutOptions controls a server-side upload.
type PutOptions struct {
	ContentType string
	MaxBytes    int64
}

// ObjectInfo is a subset of S3 HeadObject.
type ObjectInfo struct {
	Key         string
	Size        int64
	ContentType string
	ETag        string
}

// PresignPutOptions is the capability scope for a browser PUT.
type PresignPutOptions struct {
	ContentType string
	MaxBytes    int64
	TTL         time.Duration
}

// PresignPut is a scoped upload grant. The client MUST replay Headers
// (especially Content-Type) or the signature fails. This is the feature.
type PresignPut struct {
	URL     string
	Method  string
	Headers http.Header
	Expires time.Time
}

// Config is S3-compatible bucket configuration. Empty Endpoint = AWS default.
type Config struct {
	Endpoint       string // minio:9000, <account>.r2.cloudflarestorage.com, or empty
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	PublicBaseURL  string // host the browser sees; used only by the signer
	ForcePathStyle bool
	UseTLS         bool
}
