package storage

import (
	"context"
	"io"

	"github.com/Damione1/thread-art-generator/core/util"
)

// StorageProvider defines the interface for storage operations
// This interface abstracts storage implementations to be environment-agnostic
type StorageProvider interface {
	// Basic CRUD operations
	Upload(ctx context.Context, key string, data io.Reader, contentType string) error
	UploadWithPublicRead(ctx context.Context, key string, data io.Reader, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// URL generation
	SignedURL(ctx context.Context, key string, opts *SignedURLOptions) (string, error)
	GetPublicURL(key string) string
	GetDownloadURL(ctx context.Context, key string) (string, error)

	// Writer interface for streaming uploads
	NewWriter(ctx context.Context, key string, opts *WriterOptions) (*Writer, error)

	// Health and accessibility checks
	IsAccessible(ctx context.Context) (bool, error)
	Close() error
}

// StorageConfig provides configuration for storage providers
// This configuration is separated from util.Config to maintain clear boundaries
type StorageConfig struct {
	Provider             string // e.g., "firebase", "gcs", "s3"
	ProjectID            string
	Bucket               string // Single bucket for all resources
	Region               string
	EmulatorHost         string // For local development
	EmulatorExternalHost string // Host accessible from browsers
	IsEmulatorMode       bool   // Centralized environment detection
}

// ConfigFromUtil converts util.Config to StorageConfig
// This is the centralized bridge between main config and storage config
func ConfigFromUtil(config util.Config) StorageConfig {
	return StorageConfig{
		Provider:             config.Storage.Provider,
		ProjectID:            config.Firebase.ProjectID,
		Bucket:               config.Storage.Bucket,
		Region:               config.Storage.Region,
		EmulatorHost:         config.Firebase.StorageEmulatorHost,
		EmulatorExternalHost: config.Firebase.StorageEmulatorExternalHost,
		IsEmulatorMode:       config.Firebase.StorageEmulatorHost != "", // Storage-specific emulator detection
	}
}

// NewStorageProvider creates a storage provider based on configuration
func NewStorageProvider(ctx context.Context, config StorageConfig) (StorageProvider, error) {
	switch config.Provider {
	case "firebase", "gcs", "":
		// Default to Firebase Storage
		return NewFirebaseStorageFromConfig(ctx, config)
	default:
		return nil, ErrUnsupportedProvider
	}
}

// NewStorageProviderFromUtil creates a storage provider directly from util.Config
// This is a convenient helper that combines ConfigFromUtil and NewStorageProvider
func NewStorageProviderFromUtil(ctx context.Context, config util.Config) (StorageProvider, error) {
	storageConfig := ConfigFromUtil(config)
	return NewStorageProvider(ctx, storageConfig)
}