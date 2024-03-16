package aws

import (
	"context"
	"log"
	"testing"
)

const (
	testRegion     = "us-east-1"
	testBucketName = "storage-synk-test"
)

func TestBucketCreate(t *testing.T) {
	err := S3BucketCreate(context.Background(), testRegion, testBucketName)
	if err != nil {
		log.Printf("Test Failed with err: %v", err)
		t.Fatal()
	}
}

func TestBucketGet(t *testing.T) {
	bucket, err := S3BucketGet(context.Background(), testRegion, testBucketName)
	if err != nil {
		log.Printf("Test failed with the err: %v", err)
		t.Fatal()
	}
	log.Printf("Succesfully fetched bucket: %s", *bucket.Name)
}
