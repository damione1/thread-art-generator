package storage

import (
	"context"

	"github.com/Damione1/thread-art-generator/core/util"
)

func NewBucket(ctx context.Context, cfg Config) (Bucket, error) {
	return NewS3Bucket(ctx, cfg)
}

func BucketConfigFromUtil(c util.Config) Config {
	return Config{
		Endpoint:       c.Storage.Endpoint,
		Region:         c.Storage.Region,
		Bucket:         c.Storage.Bucket,
		AccessKey:      c.Storage.AccessKey,
		SecretKey:      c.Storage.SecretKey,
		PublicBaseURL:  c.Storage.PublicBaseURL,
		ForcePathStyle: c.Storage.ForcePathStyle,
		UseTLS:         c.Storage.UseTLS,
	}
}
