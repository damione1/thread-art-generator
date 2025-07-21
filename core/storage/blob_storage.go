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
	Provider         StorageProvider
	Bucket           string
	Region           string
	InternalEndpoint string // Internal endpoint for operations (e.g., http://minio:9000)
	ExternalEndpoint string // External endpoint for signed URLs (e.g., http://localhost:9000)
	UseSSL           bool   // Used for internal connections
	ForceExternalSSL bool   // Whether to force HTTPS for external URLs
	// Auth credentials
	AccessKey string
	SecretKey string
	// GCS specific
	GCPProjectID   string
	GCPCredentials []byte // Optional JSON credentials
}

// For backward compatibility
// PublicURL is now deprecated in favor of ExternalEndpoint
func (c BlobStorageConfig) PublicURL() string {
	if c.ExternalEndpoint != "" {
		return stripProtocol(c.ExternalEndpoint)
	}
	return ""
}

// BlobStorage provides a unified interface for blob storage operations
// with support for different cloud providers
type BlobStorage struct {
	*blob.Bucket
	provider         StorageProvider
	s3Client         *s3.S3 // For S3/MinIO signed URLs
	bucketName       string
	publicURL        string
	forceExternalSSL bool
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
		Bucket:           bucket,
		provider:         config.Provider,
		bucketName:       config.Bucket,
		publicURL:        config.PublicURL(),
		forceExternalSSL: config.ForceExternalSSL,
	}

	// Create signing client for S3/MinIO specifically for generating external URLs
	if config.Provider == ProviderS3 || config.Provider == ProviderMinIO {
		// External endpoint is required for signed URLs
		if config.ExternalEndpoint == "" {
			return nil, fmt.Errorf("external endpoint is required for S3/MinIO storage")
		}

		// Strip protocol if present, we'll explicitly set the protocol based on config
		externalEndpoint := stripProtocol(config.ExternalEndpoint)

		// Create a separate AWS session configured for generating external URLs
		awsConfig := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
			Region:           aws.String(config.Region),
			Endpoint:         aws.String(externalEndpoint),
			DisableSSL:       aws.Bool(!config.ForceExternalSSL),
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

// newS3BlobStorage creates a new S3-compatible blob storage for internal operations
func newS3BlobStorage(ctx context.Context, config BlobStorageConfig) (*blob.Bucket, error) {
	// Ensure we have an internal endpoint
	if config.InternalEndpoint == "" {
		return nil, fmt.Errorf("internal endpoint is required for S3/MinIO storage")
	}

	// For internal operations, we use the internal endpoint with specified SSL setting
	internalEndpoint := stripProtocol(config.InternalEndpoint)

	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		Region:           aws.String(config.Region),
		Endpoint:         aws.String(internalEndpoint),
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
// with security constraints for image uploads
func (b *BlobStorage) s3SignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	// Default to PUT if method not specified
	method := opts.Method
	if method == "" {
		method = "PUT"
	}

	// Validate and enforce security constraints for uploads
	if method == "PUT" {
		// Ensure content type is specified and is an image
		if opts.ContentType == "" {
			return "", fmt.Errorf("content type is required for uploads")
		}
		
		// Validate image content types
		validImageTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		}
		
		if !validImageTypes[opts.ContentType] {
			return "", fmt.Errorf("invalid content type: %s. Only image files are allowed", opts.ContentType)
		}
		
		// Enforce 1-minute expiration for security
		if opts.Expiry > time.Minute {
			opts.Expiry = time.Minute
		}
		if opts.Expiry == 0 {
			opts.Expiry = time.Minute
		}
	}

	// For GET requests, use longer expiration for viewing
	if method == "GET" && opts.Expiry == 0 {
		opts.Expiry = 15 * time.Minute
	}

	// Create a secure request with proper content type validation
	var req *request.Request
	switch method {
	case "GET":
		input := &s3.GetObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(key),
		}
		req, _ = b.s3Client.GetObjectRequest(input)
	case "PUT":
		input := &s3.PutObjectInput{
			Bucket: aws.String(b.bucketName),
			Key:    aws.String(key),
		}
		// Don't include ContentType or ContentLength in the input to avoid signing complications
		// The validation is done at the API layer, and the client will send the content-type header
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

	// Set URL scheme based on configuration
	if b.forceExternalSSL {
		req.HTTPRequest.URL.Scheme = "https"
	} else {
		req.HTTPRequest.URL.Scheme = "http"
	}

	// Don't set headers before signing to avoid signature mismatches
	// The client will send the Content-Type header which won't be part of the signature
	
	// Generate the signed URL with security constraints
	url, err := req.Presign(opts.Expiry)
	if err != nil {
		return "", fmt.Errorf("failed to sign request: %v", err)
	}

	// Only force HTTPS if configured to do so
	if b.forceExternalSSL && strings.HasPrefix(url, "http://") {
		url = "https://" + strings.TrimPrefix(url, "http://")
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

// GetPublicURL returns a direct URL for public access without signing
func (b *BlobStorage) GetPublicURL(key string) string {
	protocol := "http"
	if b.forceExternalSSL {
		protocol = "https"
	}

	// Use the external endpoint for public URLs
	if b.publicURL == "" {
		return ""
	}

	// For MinIO and S3, the public URL format is: {protocol}://{endpoint}/{bucket}/{key}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, b.publicURL, b.bucketName, key)
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
