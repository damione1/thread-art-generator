package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// StorageProvider identifies which cloud storage provider to use
type StorageProvider string

const (
	ProviderS3    StorageProvider = "s3"
	ProviderMinIO StorageProvider = "minio"
	ProviderGCS   StorageProvider = "gcs"
)

// BlobStorageConfig holds configuration for S3-compatible storage.
type BlobStorageConfig struct {
	Provider         StorageProvider
	Bucket           string
	Region           string
	InternalEndpoint string // ops: minio:9000
	ExternalEndpoint string // sign: localhost:9000 (browser)
	UseSSL           bool
	ForceExternalSSL bool
	AccessKey        string
	SecretKey        string
	GCPProjectID     string
	GCPCredentials   []byte
}

// SignedURLOptions is the local stand-in for gocloud blob.SignedURLOptions.
type SignedURLOptions struct {
	Expiry      time.Duration
	Method      string
	ContentType string
}

// BlobStorage is the legacy dual-bucket handle. Internals are AWS SDK v2 Bucket.
type BlobStorage struct {
	bucket           Bucket
	provider         StorageProvider
	bucketName       string
	publicURL        string
	forceExternalSSL bool
}

// NewBlobStorage creates a BlobStorage backed by the S3 driver (MinIO/R2/AWS).
func NewBlobStorage(ctx context.Context, config BlobStorageConfig) (*BlobStorage, error) {
	switch config.Provider {
	case ProviderS3, ProviderMinIO, "":
	case ProviderGCS:
		return nil, fmt.Errorf("storage: GCS is not supported; use S3-compatible (MinIO/R2/AWS)")
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	publicBase := strings.TrimRight(config.ExternalEndpoint, "/")
	if publicBase != "" && config.Bucket != "" && !strings.HasSuffix(publicBase, "/"+config.Bucket) {
		publicBase = publicBase + "/" + config.Bucket
	}

	b, err := NewS3Bucket(ctx, Config{
		Endpoint:       config.InternalEndpoint,
		Region:         config.Region,
		Bucket:         config.Bucket,
		AccessKey:      config.AccessKey,
		SecretKey:      config.SecretKey,
		PublicBaseURL:  publicBase,
		ForcePathStyle: true,
		UseTLS:         config.UseSSL || config.ForceExternalSSL,
	})
	if err != nil {
		return nil, err
	}

	return &BlobStorage{
		bucket:           b,
		provider:         config.Provider,
		bucketName:       config.Bucket,
		publicURL:        stripProtocol(config.ExternalEndpoint),
		forceExternalSSL: config.ForceExternalSSL,
	}, nil
}

// Bucket returns the cloud-agnostic driver (PresignPut with headers).
func (b *BlobStorage) Bucket() Bucket {
	return b.bucket
}

// SignedURL generates a pre-signed URL. PUT does not return headers — use Bucket().PresignPut.
func (b *BlobStorage) SignedURL(ctx context.Context, key string, opts *SignedURLOptions) (string, error) {
	if opts == nil {
		opts = &SignedURLOptions{}
	}
	method := strings.ToUpper(opts.Method)
	if method == "" {
		method = "PUT"
	}
	ttl := opts.Expiry
	if ttl == 0 {
		if method == "GET" {
			ttl = 15 * time.Minute
		} else {
			ttl = time.Minute
		}
	}
	switch method {
	case "PUT":
		out, err := b.bucket.PresignPut(ctx, key, PresignPutOptions{
			ContentType: opts.ContentType,
			TTL:         ttl,
		})
		if err != nil {
			return "", err
		}
		return out.URL, nil
	default:
		return b.bucket.PresignGet(ctx, key, ttl)
	}
}

func (b *BlobStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	return b.bucket.Put(ctx, key, data, PutOptions{ContentType: contentType})
}

func (b *BlobStorage) UploadWithPublicRead(ctx context.Context, key string, data io.Reader, contentType string) error {
	return b.Upload(ctx, key, data, contentType)
}

func (b *BlobStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	r, _, err := b.bucket.Get(ctx, key)
	return r, err
}

func (b *BlobStorage) Delete(ctx context.Context, key string) error {
	return b.bucket.Delete(ctx, key)
}

func (b *BlobStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := b.bucket.Head(ctx, key)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b *BlobStorage) GetPublicURL(key string) string {
	protocol := "http"
	if b.forceExternalSSL {
		protocol = "https"
	}
	if b.publicURL == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, b.publicURL, b.bucketName, key)
}

func (b *BlobStorage) Close() error {
	return nil
}

func stripProtocol(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	return u
}
