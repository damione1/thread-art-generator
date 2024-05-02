package grpcApi

import (
	"context"
	"fmt"
	"log"

	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/token"
	"github.com/Damione1/thread-art-generator/pkg/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Server struct {
	pb.UnimplementedArtGeneratorServiceServer
	config     util.Config
	tokenMaker token.Maker
	bucket     *blob.Bucket
}

func NewServer(config util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token maker. %v", err)
	}

	var bucket *blob.Bucket

	fmt.Println("Environment:", config.Environment)

	switch config.Environment {
	case "production":
		// bucket, err := gcsblob.OpenBucket(context.Background(), config.GCSBucketName)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer bucket.Close()
	case "development":
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String("us-west-2"),             //Placeholders for AWS region
			Endpoint:    aws.String("http://localhost:9000"), // replace with your MinIO server address
			DisableSSL:  aws.Bool(true),
			Credentials: credentials.NewStaticCredentials("local", "locallocal", ""),
		})

		bucket, err := s3blob.OpenBucket(context.Background(), sess, "local", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer bucket.Close()

		//validate the bucket connection
		_, err = bucket.ReadAll(context.Background(), "local")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Successfully connected to the bucket")
	default:
		return nil, fmt.Errorf("unknown environment %s", config.Environment)
	}

	server := &Server{
		config:     config,
		tokenMaker: tokenMaker,
		bucket:     bucket,
	}

	return server, nil
}
