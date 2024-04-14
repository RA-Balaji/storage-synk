package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	dataSyncAmi = "/aws/service/datasync/ami"
)

func newSsmClient(region string) (*ssm.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return ssm.NewFromConfig(cfg), nil
}

func SsmParameterGet(ctx context.Context, region string) (ssm.GetParameterOutput, error) {
	client, err := newSsmClient(region)
	if err != nil {
		return ssm.GetParameterOutput{}, fmt.Errorf("Error creating ssm client: %v", err)
	}
	res, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(dataSyncAmi),
	})
	if err != nil {
		return ssm.GetParameterOutput{}, fmt.Errorf("Error getting SSM Parameter: %v", err)
	}
	return *res, nil
}

func LaunchEc2ForDatasync(ctx context.Context, region string) error {
	ssmParam, err := SsmParameterGet(ctx, region)
	if err != nil {
		return err
	}

	client, err := newEC2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating ec2 client: %v", err)
	}

	inp := ec2.RunInstancesInput{
		MaxCount:     aws.Int32(1),
		MinCount:     aws.Int32(1),
		ImageId:      ssmParam.Parameter.Value,
		InstanceType: types.InstanceTypeM1Small, // allow user to pass this
	}
	_, err = client.RunInstances(ctx, &inp)
	if err != nil {
		return fmt.Errorf("Error Launching ec2 Instance: %v", err)
	}

	return nil
}
