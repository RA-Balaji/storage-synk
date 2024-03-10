package aws

import (
	"fmt"
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func newS3Client(region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}


func S3BucketCreate(ctx context.Context, region, bucketName string) error {
	client, err := newS3Client(region)
	if err != nil {

	}

	createReq := &s3.CreateBucketInput{
		Bucket: &bucketName,
	}

	_, err = client.CreateBucket(ctx, createReq)
	if err != nil {
		return fmt.Errorf("Error creating S3 bucket: %v", err)
	}

	return nil
}