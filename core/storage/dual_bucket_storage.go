package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/Damione1/thread-art-generator/core/util"
)

// DualBucketStorage manages two separate storage buckets - one public and one private
type DualBucketStorage struct {
	publicStorage  *BlobStorage // For CDN-cacheable content
	privateStorage *BlobStorage // For signed-URL-only content
	config         util.StorageConfig
}

// NewDualBucketStorage creates a new dual bucket storage system
func NewDualBucketStorage(ctx context.Context, config util.StorageConfig) (*DualBucketStorage, error) {
	// Create public bucket storage
	publicStorage, err := NewBlobStorage(ctx, BlobStorageConfig{
		Provider:         StorageProvider(config.Provider),
		Bucket:           config.PublicBucket,
		Region:           config.Region,
		InternalEndpoint: config.InternalEndpoint,
		ExternalEndpoint: config.ExternalEndpoint,
		UseSSL:           config.UseSSL,
		ForceExternalSSL: config.ForceExternalSSL,
		AccessKey:        config.AccessKey,
		SecretKey:        config.SecretKey,
		GCPProjectID:     config.GCPProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create public storage: %w", err)
	}

	// Create private bucket storage
	privateStorage, err := NewBlobStorage(ctx, BlobStorageConfig{
		Provider:         StorageProvider(config.Provider),
		Bucket:           config.PrivateBucket,
		Region:           config.Region,
		InternalEndpoint: config.InternalEndpoint,
		ExternalEndpoint: config.ExternalEndpoint,
		UseSSL:           config.UseSSL,
		ForceExternalSSL: config.ForceExternalSSL,
		AccessKey:        config.AccessKey,
		SecretKey:        config.SecretKey,
		GCPProjectID:     config.GCPProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create private storage: %w", err)
	}

	return &DualBucketStorage{
		publicStorage:  publicStorage,
		privateStorage: privateStorage,
		config:         config,
	}, nil
}

// GetPublicStorage returns the public bucket storage for CDN-cacheable content
func (d *DualBucketStorage) GetPublicStorage() *BlobStorage {
	return d.publicStorage
}

// GetPrivateStorage returns the private bucket storage for signed-URL-only content
func (d *DualBucketStorage) GetPrivateStorage() *BlobStorage {
	return d.privateStorage
}

// UploadPublic uploads content to the public bucket (CDN-cacheable)
func (d *DualBucketStorage) UploadPublic(ctx context.Context, key string, data io.Reader, contentType string) error {
	return d.publicStorage.Upload(ctx, key, data, contentType)
}

// UploadPrivate uploads content to the private bucket (signed URLs only)
func (d *DualBucketStorage) UploadPrivate(ctx context.Context, key string, data io.Reader, contentType string) error {
	return d.privateStorage.Upload(ctx, key, data, contentType)
}

// DeletePublic deletes content from the public bucket
func (d *DualBucketStorage) DeletePublic(ctx context.Context, key string) error {
	return d.publicStorage.Delete(ctx, key)
}

// DeletePrivate deletes content from the private bucket
func (d *DualBucketStorage) DeletePrivate(ctx context.Context, key string) error {
	return d.privateStorage.Delete(ctx, key)
}

// Close closes both storage connections
func (d *DualBucketStorage) Close() error {
	var err1, err2 error
	if d.publicStorage != nil {
		err1 = d.publicStorage.Close()
	}
	if d.privateStorage != nil {
		err2 = d.privateStorage.Close()
	}
	
	if err1 != nil {
		return err1
	}
	return err2
}