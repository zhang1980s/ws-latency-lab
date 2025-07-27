package resources

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateTransitEc2Instance creates an EC2 instance in the transit VPC
func CreateTransitEc2Instance(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnets []pulumi.IDOutput, sg *ec2.SecurityGroup) (*ec2.Instance, error) {
	// Create IAM role for EC2
	_, instanceProfile, err := createIamRole(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Get the latest Amazon Linux 2023 minimal with kernel-6.12 AMI from SSM Parameter Store
	ssmParam, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
		Name: config.AmazonLinux2023SsmParameter,
	})
	if err != nil {
		return nil, err
	}

	// Create user data script for the transit EC2 instance
	userData := pulumi.String(fmt.Sprintf(`#!/bin/bash
# Set hostname
hostnamectl set-hostname transit-client-instance
echo "127.0.0.1 transit-client-instance" >> /etc/hosts

# Basic setup
yum update -y
yum install -y amazon-cloudwatch-agent
`))

	// Get VPC configuration to determine which subnet to use
	vpcConfig := config.GetTransitVpcConfig(cfg)

	// Find the subnet in ap-east-1b (which is AvailabilityZone2 in constants.go)
	// The subnets are created in order of the availability zones in the VPC config
	// So we need to find the index of ap-east-1b in the availability zones
	var subnetIndex int
	for i, az := range vpcConfig.AvailabilityZones {
		if az == config.AvailabilityZone2 {
			subnetIndex = i
			break
		}
	}

	// Make sure we don't go out of bounds
	if subnetIndex >= len(subnets) {
		subnetIndex = 0
	}

	// Create EC2 instance
	instanceName := "transit-client-instance"
	instance, err := ec2.NewInstance(ctx, instanceName, &ec2.InstanceArgs{
		Ami:                      pulumi.String(ssmParam.Value),
		InstanceType:             pulumi.String(cfg.TransitInstanceType),
		KeyName:                  pulumi.String(cfg.KeyPairName),
		SubnetId:                 subnets[subnetIndex],
		VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
		IamInstanceProfile:       instanceProfile.Name,
		AssociatePublicIpAddress: pulumi.Bool(true), // Ensure it gets a public IP
		UserData:                 userData,
		AvailabilityZone:         pulumi.String(config.AvailabilityZone2),
		Tags:                     utils.ApplyTags(ctx, instanceName, utils.GetNamedTags(instanceName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(20), // 20 GB root volume
			VolumeType: pulumi.String("gp3"),
		},
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// createIamRole creates an IAM role for the EC2 instance
func createIamRole(ctx *pulumi.Context, cfg *config.Config) (*iam.Role, *iam.InstanceProfile, error) {
	// Create IAM role
	roleName := "transit-client-instance-role"
	role, err := iam.NewRole(ctx, roleName, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "ec2.amazonaws.com"
				},
				"Effect": "Allow",
				"Sid": ""
			}]
		}`),
		Tags: utils.ApplyTags(ctx, roleName, utils.GetNamedTags(roleName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}

	// Attach policies to the role
	_, err = iam.NewRolePolicyAttachment(ctx, "transit-client-instance-ssm-policy", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "transit-client-instance-cloudwatch-policy", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"),
	})
	if err != nil {
		return nil, nil, err
	}

	// Create instance profile
	profileName := "transit-client-instance-profile"
	instanceProfile, err := iam.NewInstanceProfile(ctx, profileName, &iam.InstanceProfileArgs{
		Role: role.Name,
		Tags: utils.ApplyTags(ctx, profileName, utils.GetNamedTags(profileName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}

	return role, instanceProfile, nil
}
