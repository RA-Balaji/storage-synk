package aws

import (
	"context"
	"log"
	"testing"
)

const (
	testRegion     = "us-east-1"
	testBucketName = "balaji-tests"
	testProject    = "947123667364"
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
