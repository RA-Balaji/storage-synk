package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		return fmt.Errorf("Error initializing s3client: %v", err)
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

func S3BucketGet(ctx context.Context, region, bucketName string) (*types.Bucket, error) {
	client, err := newS3Client(region)
	if err != nil {
		return nil, fmt.Errorf("Error initializing s3client: %v", err)
	}

	bucketList, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("Error Listing S3 buckets: %v", err)
	}

	var res *types.Bucket
	for _, bucket := range bucketList.Buckets {
		if *bucket.Name == bucketName {
			res = &bucket
			break
		}
	}
	if res == nil {
		return nil, fmt.Errorf("Bucket NotFound: bucket: [%s]", bucketName)
	}

	return res, nil
}
