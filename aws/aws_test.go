package aws

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testRegion     = "us-east-1"
	testBucketName = "balaji-tests"
	testProject    = "947123667364"
	testDirectory  = "test-dir"
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

func TestIAMRoleCreate(t *testing.T) {
	err := IAMRoleCreate(context.Background(), testProject, testRegion, testBucketName)
	if err != nil {
		log.Printf("Error creating IAM Role: %v", err)
		t.Fatal()
	}
}

func TestVpcCreate(t *testing.T) {
	err := VPCCreate(context.Background(), testRegion, testBucketName)
	if err != nil {
		log.Printf("Error creating VPC: %v", err)
		t.Fatal()
	}
}

func TestSubnetCreate(t *testing.T) {
	testSnetZones := []string{"us-east-1a", "us-east-1b"}
	err := SubnetCreate(context.Background(), testRegion, testBucketName, testSnetZones)
	if err != nil {
		log.Printf("Error creating Subnet(s): %v", err)
		t.Fatal()
	}
}

func TestVpcEndpointCreate(t *testing.T) {
	_, err := VPCEndpointCreate(context.Background(), testRegion, testBucketName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSSMParameterGet(t *testing.T) {
	ssm, err := SsmParameterGet(context.Background(), testRegion)
	if err != nil {
		t.Fatal(err)
	}
	if *ssm.Parameter.Name != dataSyncAmi {
		t.Fatalf("Parameter: %s", *ssm.Parameter.Name)
	}
}

func TestS3FolderUpload(t *testing.T) {
	ctx := context.Background()

	// Create bucket if it doesn't exist
	err := S3BucketCreate(ctx, testRegion, "balaji-tests-2")

	// Create a temporary directory and some files
	tmpDir, err := os.MkdirTemp("", "testdir")
	assert.NoError(t, err)
	log.Println("tmpDir:", tmpDir)

	filePaths := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "subdir", "file2.txt"),
		filepath.Join(tmpDir, "subdir", "file3.txt"),
	}

	// Create file and lead content
	for _, path := range filePaths {
		os.MkdirAll(filepath.Dir(path), 0755)
		f, err := os.Create(path)
		assert.NoError(t, err)
		f.WriteString("test content")
		f.Close()
	}

	// Initialize wait group, and semaphore
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit to 10 concurrent uploads

	// Perform the upload
	err = S3FolderUpload(ctx, testRegion, "balaji-tests-2", tmpDir, &wg, sem)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
