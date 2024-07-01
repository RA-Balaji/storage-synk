package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	// TODO: Use Multi-part upload for files > 100 Mb
	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, file); err != nil {
		fmt.Errorf("Error reading file: [%v]", err)
	}

	// Upload the file to S3
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buffer.Bytes()),
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
			// To convert '\' to '/'
			if isWindowsOS() {
				key = convKeyToS3Format(key)
			}
			if err := S3FileUpload(ctx, region, bucketName, path, key); err != nil {
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

func isWindowsOS() bool {
	os := runtime.GOOS
	if os == "windows" {
		return true
	}
	return false
}

func convKeyToS3Format(key string) string {
	parts := strings.Split(key, "\\")
	return strings.Join(parts, "/")
}
