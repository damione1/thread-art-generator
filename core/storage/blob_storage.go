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

func NewMinioBlobStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, publicURL string) (*blob.Bucket, error) {
	// Create a session with public URL if specified, for generating signed URLs
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!useSSL),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String("us-east-1"),
	}

	// If public URL is provided, configure it for signed URLs
	if publicURL != "" {
		s3Config.Endpoint = aws.String(publicURL)
	}

	s3Session, err := session.NewSession(s3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create a session with the internal endpoint for actual operations
	if publicURL != "" {
		// Only create a second session if we're using different endpoints
		internalConfig := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Endpoint:         aws.String(endpoint),
			DisableSSL:       aws.Bool(!useSSL),
			S3ForcePathStyle: aws.Bool(true),
			Region:           aws.String("us-east-1"),
		}

		internalSession, err := session.NewSession(internalConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create internal AWS session: %v", err)
		}

		// Initialize minio client object for actual operations
		blob, err := s3blob.OpenBucket(context.Background(), internalSession, bucket, nil)
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

	// If no public URL was provided, just use the single session
	blob, err := s3blob.OpenBucket(context.Background(), s3Session, bucket, nil)
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
