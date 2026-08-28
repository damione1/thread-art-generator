package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Damione1/thread-art-generator/core/util"
)

// DualBucketStorage manages two separate storage buckets - one public and one private
type DualBucketStorage struct {
	publicStorage  *BlobStorage
	privateStorage *BlobStorage
	config         util.StorageConfig
}

func NewDualBucketStorage(ctx context.Context, config util.StorageConfig) (*DualBucketStorage, error) {
	publicStorage, err := NewBlobStorage(ctx, blobConfig(config, config.PublicBucket))
	if err != nil {
		return nil, fmt.Errorf("failed to create public storage: %w", err)
	}

	d := &DualBucketStorage{
		publicStorage:  publicStorage,
		privateStorage: publicStorage,
		config:         config,
	}

	privateBucket := strings.TrimSpace(config.PrivateBucket)
	if privateBucket != "" && privateBucket != config.PublicBucket {
		privateStorage, err := NewBlobStorage(ctx, blobConfig(config, privateBucket))
		if err != nil {
			_ = publicStorage.Close()
			return nil, fmt.Errorf("failed to create private storage: %w", err)
		}
		d.privateStorage = privateStorage
	}

	return d, nil
}

func blobConfig(config util.StorageConfig, bucket string) BlobStorageConfig {
	return BlobStorageConfig{
		Provider:         StorageProvider(config.Provider),
		Bucket:           bucket,
		Region:           config.Region,
		InternalEndpoint: config.InternalEndpoint,
		ExternalEndpoint: config.ExternalEndpoint,
		UseSSL:           config.UseSSL,
		ForceExternalSSL: config.ForceExternalSSL,
		AccessKey:        config.AccessKey,
		SecretKey:        config.SecretKey,
		GCPProjectID:     config.GCPProjectID,
	}
}

func (d *DualBucketStorage) GetPublicStorage() *BlobStorage {
	return d.publicStorage
}

func (d *DualBucketStorage) GetPrivateStorage() *BlobStorage {
	return d.privateStorage
}

func (d *DualBucketStorage) UploadPublic(ctx context.Context, key string, data io.Reader, contentType string) error {
	return d.publicStorage.Upload(ctx, key, data, contentType)
}

func (d *DualBucketStorage) UploadPrivate(ctx context.Context, key string, data io.Reader, contentType string) error {
	return d.privateStorage.Upload(ctx, key, data, contentType)
}

func (d *DualBucketStorage) DeletePublic(ctx context.Context, key string) error {
	return d.publicStorage.Delete(ctx, key)
}

func (d *DualBucketStorage) DeletePrivate(ctx context.Context, key string) error {
	return d.privateStorage.Delete(ctx, key)
}

func (d *DualBucketStorage) Close() error {
	var err1, err2 error
	if d.publicStorage != nil {
		err1 = d.publicStorage.Close()
	}
	if d.privateStorage != nil && d.privateStorage != d.publicStorage {
		err2 = d.privateStorage.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}
