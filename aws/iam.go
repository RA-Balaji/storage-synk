package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

const dataSyncRole = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "datasync.amazonaws.com"
            },
            "Action": "sts:AssumeRole",
            "Condition": {
                "StringEquals": {
                    "aws:SourceAccount": "%s"
                },
                "StringLike": {
                    "aws:SourceArn": "arn:aws:datasync:%s:%s:*"
                }
            }
        }
    ]
}`

func newIAMClient(region string) (*iam.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return iam.NewFromConfig(cfg), nil
}

func IAMRoleCreate(ctx context.Context, project, region, bucketName string) error {
	iamClient, err := newIAMClient(region)
	if err != nil {
		return fmt.Errorf("Error initializing iam client: %v", err)
	}

	input := iam.CreateRoleInput{
		RoleName:                 aws.String(fmt.Sprintf("storage-synk-%s", bucketName)),
		AssumeRolePolicyDocument: aws.String(fmt.Sprintf(dataSyncRole, project, region, project)),
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(bucketName),
			},
		},
	}
	_, err = iamClient.CreateRole(ctx, &input)
	if err != nil {
		return fmt.Errorf("Error creating IAM Role: %v", err)
	}

	return nil
}
