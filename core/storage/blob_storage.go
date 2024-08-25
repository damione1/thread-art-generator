package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

func NewMinioBlobStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*blob.Bucket, error) {
	session, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String("us-east-1"),
	})

	// Initialize minio client object.
	blob, err := s3blob.OpenBucket(context.Background(), session, bucket, nil)
	if err != nil {
		return nil, err
	}

	if ok, err := blob.IsAccessible(context.Background()); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("bucket %s is not accessible", bucket)
	}

	return blob, nil
}
