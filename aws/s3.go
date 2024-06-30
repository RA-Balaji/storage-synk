package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	//"github.com/aws/aws-sdk-go/aws/awserr"
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

func S3BucketDelete(ctx context.Context, region, bucketName string) error {
	client, err := newS3Client(region)
	if err != nil {
		return fmt.Errorf("Error initializing s3client: %v", err)
	}

	delReq := &s3.DeleteBucketInput{
		Bucket: &bucketName,
	}

	_, err = client.DeleteBucket(ctx, delReq)
	if err != nil {
		return fmt.Errorf("Error creating S3 bucket: %v", err)
	}

	return nil
}

// func isBucketAlreadyExistsError(err error) bool {
// 	if awsErr, ok := err.(awserr.Error); ok {
// 		return awsErr.Code() == s3. || awsErr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou
// 	}
// 	return false
// }

func S3FileUpload(ctx context.Context, region, bucketName, fileName, key string) error {
	client, err := newS3Client(region)
	if err != nil {
		return fmt.Errorf("Error initializing s3client: %v", err)
	}

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Error opening file %s: %v", fileName, err)
	}
	defer file.Close()

	// Get the file size and read the file content into a buffer
	// TODO: Use Multi-part upload for files > 100 Mb
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Upload the file to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf(
			"Error Uploading file to S3 bucket [%s], File [%s]: %v",
			bucketName, fileName, err)
	}
	return nil
}

func S3FolderUpload(
	ctx context.Context,
	region, bucketName, folderName string,
	wg *sync.WaitGroup, sem chan struct{}) error {

	err := filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it's a directory, just continue
		if info.IsDir() {
			return nil
		}

		sem <- struct{}{} // Acquire semaphore
		wg.Add(1)

		go func(path string) error {
			defer func() {
				wg.Done()
				<-sem
			}()

			key := path[len(folderName)+1:]
			if err := S3FileUpload(ctx, region, bucketName, key, path); err != nil {
				return err
			}
			return nil
		}(path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("Error Uploading folder [%s]", folderName)
	}

	return nil
}
