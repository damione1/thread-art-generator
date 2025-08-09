package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

// FirebaseStorage provides a unified interface for Firebase Storage operations
type FirebaseStorage struct {
	client               *storage.Client
	bucketName           string
	projectID            string
	isEmulator           bool
	emulatorHost         string
	emulatorExternalHost string // Host accessible from browsers
}

// FirebaseStorageConfig holds configuration for Firebase Storage
type FirebaseStorageConfig struct {
	ProjectID            string
	BucketName           string
	EmulatorHost         string // For local development with emulator (internal host)
	EmulatorExternalHost string // Host accessible from browsers (external host)
}

// NewFirebaseStorage creates a new Firebase storage client
func NewFirebaseStorage(ctx context.Context, config FirebaseStorageConfig) (*FirebaseStorage, error) {
	var opts []option.ClientOption

	// Check if we're using the emulator
	isEmulator := config.EmulatorHost != ""

	if isEmulator {
		// Configure for emulator usage
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("http://%s/storage/v1/", config.EmulatorHost)))
		opts = append(opts, option.WithoutAuthentication())
	}

	// Create storage client
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase Storage client: %v", err)
	}

	fs := &FirebaseStorage{
		client:               client,
		bucketName:           config.BucketName,
		projectID:            config.ProjectID,
		isEmulator:           isEmulator,
		emulatorHost:         config.EmulatorHost,
		emulatorExternalHost: config.EmulatorExternalHost,
	}

	// Verify bucket is accessible (skip for emulator as it auto-creates buckets)
	if !isEmulator {
		bucket := client.Bucket(config.BucketName)
		if _, err := bucket.Attrs(ctx); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("bucket %s is not accessible: %v", config.BucketName, err)
		}
	}

	return fs, nil
}

// NewFirebaseStorageFromConfig creates a new Firebase storage client from util.Config
// This helper function bridges between the main config and Firebase storage requirements
func NewFirebaseStorageFromConfig(ctx context.Context, config StorageConfig) (*FirebaseStorage, error) {
	// Convert storage config to Firebase storage config
	fbConfig := FirebaseStorageConfig{
		ProjectID:            config.ProjectID,
		BucketName:           config.Bucket,
		EmulatorHost:         config.EmulatorHost,
		EmulatorExternalHost: config.EmulatorExternalHost,
	}

	// If no bucket specified, return error
	if fbConfig.BucketName == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	return NewFirebaseStorage(ctx, fbConfig)
}

// Upload uploads data to Firebase Storage
func (fs *FirebaseStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)

	w := obj.NewWriter(ctx)
	w.ContentType = contentType

	_, err := io.Copy(w, data)
	if err != nil {
		_ = w.Close()
		return fmt.Errorf("failed to write data: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	return nil
}

// UploadWithPublicRead uploads data with public read permissions
func (fs *FirebaseStorage) UploadWithPublicRead(ctx context.Context, key string, data io.Reader, contentType string) error {
	if err := fs.Upload(ctx, key, data, contentType); err != nil {
		return err
	}

	// For emulator, skip ACL setting as it's not fully supported
	if fs.isEmulator {
		return nil
	}

	// Set public read ACL
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)
	acl := obj.ACL()

	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return fmt.Errorf("failed to set public read ACL: %v", err)
	}

	return nil
}

// Download downloads data from Firebase Storage
func (fs *FirebaseStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %v", err)
	}

	return reader, nil
}

// Delete deletes an object from Firebase Storage
func (fs *FirebaseStorage) Delete(ctx context.Context, key string) error {
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

// SignedURL generates a pre-signed URL for the given key
func (fs *FirebaseStorage) SignedURL(ctx context.Context, key string, opts *SignedURLOptions) (string, error) {
	if opts == nil {
		opts = &SignedURLOptions{}
	}
	if opts.Expiry == 0 {
		opts.Expiry = 15 * time.Minute
	}
	if opts.Method == "" {
		opts.Method = "GET"
	}

	// For emulator, return direct URL using external host for browser access
	if fs.isEmulator {
		protocol := "http"
		// Use external host if available, otherwise fall back to internal host
		host := fs.emulatorExternalHost
		if host == "" {
			host = fs.emulatorHost
		}

		if opts.Method == "GET" {
			return fmt.Sprintf("%s://%s/v0/b/%s/o/%s?alt=media", protocol, host, fs.bucketName, url.PathEscape(key)), nil
		}
		// For uploads (PUT), use the Firebase Storage emulator simple REST API format
		// The emulator uses a different endpoint structure than production
		if opts.Method == "PUT" {
			// Firebase Storage emulator uses the simple v0 API format for uploads
			// File extension in the path should help emulator detect MIME type automatically
			return fmt.Sprintf("%s://%s/v0/b/%s/o/%s", protocol, host, fs.bucketName, url.PathEscape(key)), nil
		}
		// Fallback for other methods
		return fmt.Sprintf("%s://%s/v0/b/%s/o/%s", protocol, host, fs.bucketName, url.PathEscape(key)), nil
	}

	// For production, use proper signed URLs
	// Convert our options to Google Cloud Storage options
	method := opts.Method
	if method == "PUT" {
		if opts.ContentType == "" {
			return "", fmt.Errorf("content type is required for uploads")
		}
		// Log content type for debugging Firebase Storage issues
		log.Debug().
			Str("content_type", opts.ContentType).
			Str("key", key).
			Str("method", method).
			Msg("Generating signed URL with content type")
	}

	// Validate image content types for uploads
	if method == "PUT" {
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

		// Enforce 1-minute expiration for uploads for security
		if opts.Expiry > time.Minute {
			opts.Expiry = time.Minute
		}
	}

	gcsOpts := &storage.SignedURLOptions{
		Method:  method,
		Expires: time.Now().Add(opts.Expiry),
	}

	if opts.ContentType != "" {
		gcsOpts.ContentType = opts.ContentType
	}

	url, err := storage.SignedURL(fs.bucketName, key, gcsOpts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %v", err)
	}

	return url, nil
}

// GetPublicURL returns a direct URL for public access without signing
// Uses the Firebase SDK's MediaLink for proper URL generation
func (fs *FirebaseStorage) GetPublicURL(key string) string {
	// Use SDK's GetDownloadURL which properly handles both emulator and production
	ctx := context.Background()
	downloadURL, err := fs.GetDownloadURL(ctx, key)
	if err != nil {
		// Fallback to manual construction only if SDK fails (e.g., object doesn't exist)
		if fs.isEmulator {
			// Use external host if available, otherwise fall back to internal host
			host := fs.emulatorExternalHost
			if host == "" {
				host = fs.emulatorHost
			}
			return fmt.Sprintf("http://%s/v0/b/%s/o/%s?alt=media", host, fs.bucketName, url.PathEscape(key))
		}
		// For production, use standard Firebase Storage public URL format
		return fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", fs.bucketName, url.PathEscape(key))
	}
	return downloadURL
}

// GetDownloadURL returns the SDK-native download URL using MediaLink from object attributes
// This method automatically handles emulator vs production environments without hardcoded logic
func (fs *FirebaseStorage) GetDownloadURL(ctx context.Context, key string) (string, error) {
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return "", fmt.Errorf("object not found: %s", key)
		}
		return "", fmt.Errorf("failed to get object attributes: %v", err)
	}

	// MediaLink provides the correct download URL for both emulator and production
	// The SDK automatically generates the appropriate URL format
	return attrs.MediaLink, nil
}

// Close closes the Firebase Storage client
func (fs *FirebaseStorage) Close() error {
	if fs.client != nil {
		return fs.client.Close()
	}
	return nil
}

// IsAccessible checks if the bucket is accessible (for compatibility with existing interface)
func (fs *FirebaseStorage) IsAccessible(ctx context.Context) (bool, error) {
	if fs.isEmulator {
		return true, nil // Emulator auto-creates buckets
	}

	bucket := fs.client.Bucket(fs.bucketName)
	_, err := bucket.Attrs(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Exists checks if an object exists in the bucket
func (fs *FirebaseStorage) Exists(ctx context.Context, key string) (bool, error) {
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if object exists: %v", err)
	}

	return true, nil
}

// Writer interface to match blob package expectations
type Writer struct {
	writer *storage.Writer
}

// Write implements io.Writer
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

// Close closes the writer and finalizes the upload
func (w *Writer) Close() error {
	return w.writer.Close()
}

// ReadFrom reads data from a reader and writes it to the object
func (w *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(w.writer, r)
}

// NewWriter creates a new writer for uploading data to an object
func (fs *FirebaseStorage) NewWriter(ctx context.Context, key string, opts *WriterOptions) (*Writer, error) {
	bucket := fs.client.Bucket(fs.bucketName)
	obj := bucket.Object(key)

	writer := obj.NewWriter(ctx)

	// Apply options if provided
	if opts != nil {
		if opts.ContentType != "" {
			writer.ContentType = opts.ContentType
		}
		if opts.ContentEncoding != "" {
			writer.ContentEncoding = opts.ContentEncoding
		}
		if opts.CacheControl != "" {
			writer.CacheControl = opts.CacheControl
		}
		if opts.Metadata != nil {
			writer.Metadata = opts.Metadata
		}
	}

	return &Writer{writer: writer}, nil
}

// WriterOptions defines options for creating a writer
type WriterOptions struct {
	ContentType     string
	ContentEncoding string
	CacheControl    string
	Metadata        map[string]string
}

// SignedURLOptions defines options for generating signed URLs
type SignedURLOptions struct {
	Method      string        // HTTP method (GET, PUT, DELETE, etc.)
	Expiry      time.Duration // How long the URL should be valid
	ContentType string        // Content type for uploads
}

// GetFirebaseAuthClient returns a Firebase Auth client (utility function)
func GetFirebaseAuthClient(ctx context.Context, app *firebase.App) (*auth.Client, error) {
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}
	return client, nil
}
