package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

// StorageProvider identifies which cloud storage provider to use
type StorageProvider string

const (
	ProviderS3    StorageProvider = "s3"
	ProviderMinIO StorageProvider = "minio"
	ProviderGCS   StorageProvider = "gcs"
)

// BlobStorageConfig holds configuration for different storage providers
type BlobStorageConfig struct {
	Provider  StorageProvider
	Bucket    string
	Region    string
	Endpoint  string // Internal endpoint for operations
	PublicURL string // Public URL for signed URLs (no protocol)
	UseSSL    bool
	// Auth credentials
	AccessKey string
	SecretKey string
	// GCS specific
	GCPProjectID   string
	GCPCredentials []byte // Optional JSON credentials
}

// BlobStorage provides a unified interface for blob storage operations
// with support for different cloud providers
type BlobStorage struct {
	*blob.Bucket
	provider   StorageProvider
	s3Client   *s3.S3 // For S3/MinIO signed URLs
	bucketName string
	publicURL  string
}

// NewBlobStorage creates a new blob storage client based on provider
func NewBlobStorage(ctx context.Context, config BlobStorageConfig) (*BlobStorage, error) {
	var bucket *blob.Bucket
	var err error

	switch config.Provider {
	case ProviderS3, ProviderMinIO:
		bucket, err = newS3BlobStorage(ctx, config)
	case ProviderGCS:
		bucket, err = newGCSBlobStorage(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}

	if err != nil {
		return nil, err
	}

	// Verify bucket is accessible
	if ok, err := bucket.IsAccessible(ctx); err != nil {
		_ = bucket.Close()
		return nil, fmt.Errorf("bucket accessibility check failed: %v", err)
	} else if !ok {
		_ = bucket.Close()
		return nil, fmt.Errorf("bucket %s is not accessible", config.Bucket)
	}

	bs := &BlobStorage{
		Bucket:     bucket,
		provider:   config.Provider,
		bucketName: config.Bucket,
		publicURL:  config.PublicURL,
	}

	// Create signing client for S3/MinIO
	if config.Provider == ProviderS3 || config.Provider == ProviderMinIO {
		var signingEndpoint string
		if config.PublicURL != "" {
			signingEndpoint = config.PublicURL
		} else {
			signingEndpoint = config.Endpoint
		}

		// Strip protocol if present
		signingEndpoint = stripProtocol(signingEndpoint)

		awsConfig := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
			Region:           aws.String(config.Region),
			Endpoint:         aws.String(signingEndpoint),
			DisableSSL:       aws.Bool(!config.UseSSL),
			S3ForcePathStyle: aws.Bool(true),
		}

		sess, err := session.NewSession(awsConfig)
		if err != nil {
			_ = bucket.Close()
			return nil, fmt.Errorf("failed to create signing session: %v", err)
		}
		bs.s3Client = s3.New(sess)
	}

	return bs, nil
}

// newS3BlobStorage creates a new S3-compatible blob storage
func newS3BlobStorage(ctx context.Context, config BlobStorageConfig) (*blob.Bucket, error) {
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		Region:           aws.String(config.Region),
		Endpoint:         aws.String(config.Endpoint),
		DisableSSL:       aws.Bool(!config.UseSSL),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	return s3blob.OpenBucket(ctx, sess, config.Bucket, nil)
}

// newGCSBlobStorage creates a new Google Cloud Storage blob storage
func newGCSBlobStorage(ctx context.Context, config BlobStorageConfig) (*blob.Bucket, error) {
	var creds *google.Credentials
	var err error
	var scopes []string

	if len(config.GCPCredentials) > 0 {
		creds, err = google.CredentialsFromJSON(ctx, config.GCPCredentials, scopes...)
	} else {
		creds, err = google.FindDefaultCredentials(ctx, scopes...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get GCP credentials: %v", err)
	}

	client, err := gcp.NewHTTPClient(gcp.DefaultTransport(), creds.TokenSource)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %v", err)
	}

	return gcsblob.OpenBucket(ctx, client, config.Bucket, nil)
}

// SignedURL generates a pre-signed URL for the given key
func (b *BlobStorage) SignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	if opts == nil {
		opts = &blob.SignedURLOptions{}
	}
	if opts.Expiry == 0 {
		opts.Expiry = 15 * time.Minute
	}

	// For S3/MinIO, use our custom implementation for better control
	if b.provider == ProviderS3 || b.provider == ProviderMinIO {
		return b.s3SignedURL(ctx, key, opts)
	}

	// For other providers, rely on the gocloud.dev implementation
	return b.Bucket.SignedURL(ctx, key, opts)
}

// s3SignedURL generates a pre-signed URL specifically for S3/MinIO
func (b *BlobStorage) s3SignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	// Default to PUT if method not specified
	method := opts.Method
	if method == "" {
		method = "PUT"
	}

	var req *request.Request
	switch method {
	case "GET":
		req, _ = b.s3Client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(key),
		})
	case "PUT":
		input := &s3.PutObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(key),
		}
		// Add content type in input
		if opts.ContentType != "" {
			input.ContentType = aws.String(opts.ContentType)
		}
		req, _ = b.s3Client.PutObjectRequest(input)
	case "DELETE":
		req, _ = b.s3Client.DeleteObjectRequest(&s3.DeleteObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(key),
		})
	case "HEAD":
		req, _ = b.s3Client.HeadObjectRequest(&s3.HeadObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(key),
		})
	default:
		return "", fmt.Errorf("unsupported method: %s", method)
	}

	if req == nil {
		return "", fmt.Errorf("failed to create request for method %s", method)
	}

	// Ensure we're using HTTPS in the signed URL
	req.HTTPRequest.Header.Set("X-Forwarded-Proto", "https")

	url, err := req.Presign(opts.Expiry)
	if err != nil {
		return "", fmt.Errorf("failed to sign request: %v", err)
	}

	return url, nil
}

// Upload uploads data to a blob
func (b *BlobStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	w, err := b.Bucket.NewWriter(ctx, key, &blob.WriterOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to create writer: %v", err)
	}

	_, err = io.Copy(w, data)
	closeErr := w.Close()
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close writer: %v", closeErr)
	}

	return nil
}

// Download downloads data from a blob
func (b *BlobStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return b.Bucket.NewReader(ctx, key, nil)
}

// Helper to strip protocol from URL
func stripProtocol(url string) string {
	if strings.HasPrefix(url, "https://") {
		return strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		return strings.TrimPrefix(url, "http://")
	}
	return url
}
