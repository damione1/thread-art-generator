package storage

import (
	"context"
	"time"

	"gocloud.dev/blob"
)

// ImageURLGenerator provides methods for generating different types of URLs for stored images
type ImageURLGenerator interface {
	// GetPublicURL returns a direct public URL for the given key
	// This URL can be cached by CDNs and accessed without authentication
	GetPublicURL(key string) string

	// GetSignedURL returns a signed URL with expiration for secure access
	// This is used for admin operations or when public access is not desired
	GetSignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error)

	// GetUploadURL returns a signed URL for uploading content
	// This includes security constraints and content type validation
	GetUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error)
}

// PublicImageURLGenerator generates URLs for publicly accessible images (CDN-cacheable)
type PublicImageURLGenerator struct {
	storage *BlobStorage
}

// PrivateImageURLGenerator generates URLs for private images (signed URLs only)
type PrivateImageURLGenerator struct {
	storage *BlobStorage
}

// URLGenerationOptions provides options for URL generation
type URLGenerationOptions struct {
	// UsePublicURL determines whether to use public URLs for image access
	UsePublicURL bool

	// FallbackToSigned determines whether to fallback to signed URLs if public URLs are not available
	FallbackToSigned bool

	// SignedURLExpiry sets the expiry time for signed URLs (only used when UsePublicURL is false)
	SignedURLExpiry time.Duration
}

// DefaultURLOptions returns the default URL generation options
func DefaultURLOptions() *URLGenerationOptions {
	return &URLGenerationOptions{
		UsePublicURL:     true,  // Use public URLs by default for CDN caching
		FallbackToSigned: true,  // Fallback to signed URLs if public URLs are not available
		SignedURLExpiry:  15 * time.Minute,
	}
}

// AdminURLOptions returns URL options for admin operations (always signed)
func AdminURLOptions() *URLGenerationOptions {
	return &URLGenerationOptions{
		UsePublicURL:     false, // Admin operations should use signed URLs
		FallbackToSigned: true,
		SignedURLExpiry:  5 * time.Minute, // Shorter expiry for admin operations
	}
}

// URLGenerator implements the ImageURLGenerator interface using BlobStorage
type URLGenerator struct {
	storage *BlobStorage
}

// NewURLGenerator creates a new URL generator with the given storage
func NewURLGenerator(storage *BlobStorage) ImageURLGenerator {
	return &URLGenerator{
		storage: storage,
	}
}

// NewPublicURLGenerator creates a URL generator for public bucket content
func NewPublicURLGenerator(storage *BlobStorage) ImageURLGenerator {
	return &PublicImageURLGenerator{
		storage: storage,
	}
}

// NewPrivateURLGenerator creates a URL generator for private bucket content
func NewPrivateURLGenerator(storage *BlobStorage) ImageURLGenerator {
	return &PrivateImageURLGenerator{
		storage: storage,
	}
}

// GetPublicURL returns a direct public URL for the given key
func (g *URLGenerator) GetPublicURL(key string) string {
	return g.storage.GetPublicURL(key)
}

// GetSignedURL returns a signed URL with expiration for secure access
func (g *URLGenerator) GetSignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	return g.storage.SignedURL(ctx, key, opts)
}

// GetUploadURL returns a signed URL for uploading content with security constraints
func (g *URLGenerator) GetUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	opts := &blob.SignedURLOptions{
		Expiry:      expiry,
		Method:      "PUT",
		ContentType: contentType,
	}
	return g.storage.SignedURL(ctx, key, opts)
}

// GenerateImageURL generates an image URL based on the given options
// This is a helper function that chooses between public and signed URLs
func GenerateImageURL(ctx context.Context, generator ImageURLGenerator, key string, opts *URLGenerationOptions) string {
	if opts == nil {
		opts = DefaultURLOptions()
	}

	// Try public URL first if enabled
	if opts.UsePublicURL {
		publicURL := generator.GetPublicURL(key)
		if publicURL != "" {
			return publicURL
		}

		// Fall back to signed URL if public URL is not available and fallback is enabled
		if !opts.FallbackToSigned {
			return ""
		}
	}

	// Generate signed URL
	signedOpts := &blob.SignedURLOptions{
		Expiry: opts.SignedURLExpiry,
		Method: "GET",
	}

	signedURL, err := generator.GetSignedURL(ctx, key, signedOpts)
	if err != nil {
		return "" // Return empty string on error
	}

	return signedURL
}

// PublicImageURLGenerator implementations
func (g *PublicImageURLGenerator) GetPublicURL(key string) string {
	return g.storage.GetPublicURL(key)
}

func (g *PublicImageURLGenerator) GetSignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	return g.storage.SignedURL(ctx, key, opts)
}

func (g *PublicImageURLGenerator) GetUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	opts := &blob.SignedURLOptions{
		Expiry:      expiry,
		Method:      "PUT",
		ContentType: contentType,
	}
	return g.storage.SignedURL(ctx, key, opts)
}

// PrivateImageURLGenerator implementations
func (g *PrivateImageURLGenerator) GetPublicURL(key string) string {
	// Private images don't have public URLs, return empty string
	return ""
}

func (g *PrivateImageURLGenerator) GetSignedURL(ctx context.Context, key string, opts *blob.SignedURLOptions) (string, error) {
	return g.storage.SignedURL(ctx, key, opts)
}

func (g *PrivateImageURLGenerator) GetUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	opts := &blob.SignedURLOptions{
		Expiry:      expiry,
		Method:      "PUT",
		ContentType: contentType,
	}
	return g.storage.SignedURL(ctx, key, opts)
}