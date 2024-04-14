package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	vpcCidr = "10.1.0.0/16"

	firstSubnetCidr  = "10.1.10.0/24"
	secondSubnetCidr = "10.1.11.0/24"
)

func newEC2Client(region string) (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

func VPCCreate(ctx context.Context, region, bucketName string) error {
	client, err := newEC2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating ec2 client: %v", err)
	}

	tags := nameTag(getVpcName(bucketName), types.ResourceTypeVpc)
	_, err = client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock:         aws.String(vpcCidr),
		TagSpecifications: tags,
	})
	if err != nil {
		return fmt.Errorf("Error creating VPC: %v", err)
	}

	return nil
}

func getVpc(ctx context.Context, region, bucketName string) (types.Vpc, error) {
	client, err := newEC2Client(region)
	if err != nil {
		return types.Vpc{}, err
	}
	vpcName := getVpcName(bucketName)
	output, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: nameTagFilter(vpcName),
	})
	if err != nil {
		return types.Vpc{}, err
	}

	if len(output.Vpcs) == 0 {
		return types.Vpc{}, fmt.Errorf("[!(can)-find-vpc-err] %v", vpcName)
	}
	return output.Vpcs[0], nil
}

func getVpcName(bucketName string) string {
	return bucketName + "-storagesynk-vpc"
}

func UpdateVpcAttribute(ctx context.Context, region, vpcID string) error {
	client, err := newEC2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating ec2 client: %v", err)
	}
	_, err = client.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(vpcID),
		EnableDnsSupport: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("Error updating VPC attribute: %v", err)
	}
	_, err = client.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId: aws.String(vpcID),
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("Error updating VPC attribute: %v", err)
	}
	return nil
}

func SubnetCreate(ctx context.Context, region, bucketName string, subnetZones []string) error {
	client, err := newEC2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating ec2 client: %v", err)
	}

	vpc, err := getVpc(ctx, region, bucketName)
	if err != nil {
		return err
	}

	minTwoZones := 2
	if len(subnetZones) < minTwoZones {
		return fmt.Errorf("[!(can)-create-subnets-min-two-zones-required] %+v", subnetZones)
	}

	subnetCidrs := []string{firstSubnetCidr, secondSubnetCidr}
	subnetCidrZonePair := zip(subnetCidrs, subnetZones)

	for i, cidrZonePair := range subnetCidrZonePair {
		_, err = client.CreateSubnet(ctx, &ec2.CreateSubnetInput{
			VpcId:             vpc.VpcId,
			CidrBlock:         &cidrZonePair.First,
			TagSpecifications: nameTag(getSubnetName(bucketName, i+1), types.ResourceTypeSubnet),
			AvailabilityZone:  &cidrZonePair.Second,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func getSubnet(ctx context.Context, region, bucketName string) ([]types.Subnet, error) {
	output := []types.Subnet{}
	client, err := newEC2Client(region)
	if err != nil {
		return []types.Subnet{}, err
	}
	snetName := bucketName + "-storagesynk-snet"
	for i := 1; i <= 2; i++ {
		snet, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
			Filters: nameTagFilter(fmt.Sprintf("%s-%d", snetName, i)),
		})
		if err != nil {
			return []types.Subnet{}, err
		}
		output = append(output, snet.Subnets...)
	}

	if err != nil {
		return []types.Subnet{}, err
	}

	if len(output) < 2 {
		return []types.Subnet{}, fmt.Errorf("[!(can)-find-subnet-err] %v", snetName)
	}
	return output[0:2], nil
}

func VPCEndpointCreate(ctx context.Context, region, bucketName string) (types.VpcEndpoint, error) {
	client, err := newEC2Client(region)
	if err != nil {
		return types.VpcEndpoint{}, fmt.Errorf("Error creating ec2 client: %v", err)
	}
	vpc, err := getVpc(ctx, region, bucketName)
	if err != nil {
		return types.VpcEndpoint{}, err
	}
	subnets, err := getSubnet(ctx, region, bucketName)
	if err != nil {
		return types.VpcEndpoint{}, err
	}

	err = UpdateVpcAttribute(ctx, region, *vpc.VpcId)
	if err != nil {
		return types.VpcEndpoint{}, err
	}

	inp := ec2.CreateVpcEndpointInput{
		ServiceName:     aws.String(getVPCEndpointServiceName(region)),
		VpcEndpointType: types.VpcEndpointTypeInterface,
		VpcId:           vpc.VpcId,
		DnsOptions: &types.DnsOptionsSpecification{
			DnsRecordIpType: types.DnsRecordIpTypeIpv4,
		},
		SubnetIds:         []string{*subnets[0].SubnetId, *subnets[1].SubnetId},
		TagSpecifications: nameTag(getVpcEpName(bucketName), types.ResourceTypeVpcEndpoint),
	}

	output, err := client.CreateVpcEndpoint(ctx, &inp)
	if err != nil {
		return types.VpcEndpoint{}, fmt.Errorf("Error Creating VPC Endpoint: %v", err)
	}
	return *output.VpcEndpoint, nil
}

func getSubnetName(bucketName string, id int) string {
	return fmt.Sprintf("%s-storagesynk-snet-%d", bucketName, id)
}

func nameTag(bucketName string, resourceType types.ResourceType) []types.TagSpecification {
	return tag("Name", bucketName, resourceType)
}

func tag(key, value string, resourceType types.ResourceType) []types.TagSpecification {
	tags := []types.TagSpecification{
		{
			ResourceType: resourceType,
			Tags: []types.Tag{
				{
					Key:   aws.String(key),
					Value: aws.String(value),
				},
			},
		},
	}
	return tags
}

func nameTagFilter(resourceName string) []types.Filter {
	filters := []types.Filter{
		{
			Name:   aws.String("tag:Name"),
			Values: []string{resourceName},
		},
	}
	return filters
}

type Pair[T, U any] struct {
	First  T
	Second U
}

func zip[T, U any](ts []T, us []U) []Pair[T, U] {
	if len(ts) != len(us) {
		return []Pair[T, U]{}
	}
	pairs := make([]Pair[T, U], len(ts))
	for i := 0; i < len(ts); i++ {
		pairs[i] = Pair[T, U]{ts[i], us[i]}
	}
	return pairs
}

func getVPCEndpointServiceName(region string) string {
	return fmt.Sprintf("com.amazonaws.%s.datasync", region)
}

func getVpcEpName(bucketName string) string {
	return fmt.Sprintf("%s-storagesynk-ep", bucketName)
}
