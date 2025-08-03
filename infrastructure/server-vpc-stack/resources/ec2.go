package resources

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateEc2Resources creates an EC2 instance for the WebSocket server
func CreateEc2Resources(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnet pulumi.IDOutput, sg *ec2.SecurityGroup) (*ec2.Instance, *cloudwatch.LogGroup, error) {
	// Create IAM role for EC2
	_, instanceProfile, err := createIamRole(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	// Create CloudWatch log group
	logGroup, err := createEc2LogGroup(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	// Get the latest Amazon Linux 2023 minimal with kernel-6.12 AMI from SSM Parameter Store
	ssmParam, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
		Name: config.AmazonLinux2023SsmParameter,
	})
	if err != nil {
		return nil, nil, err
	}

	// Create user data script to set up the WebSocket server
	userData := pulumi.All(logGroup.Name).ApplyT(func(args []interface{}) string {
		logGroupName := args[0].(string)

		return fmt.Sprintf(`#!/bin/bash
# Set hostname to ws-server
hostnamectl set-hostname ws-server
echo "127.0.0.1 ws-client" >> /etc/hosts

# Install tuned and realtime profile
dnf install -y tuned tuned-profiles-realtime
systemctl enable --now tuned
tuned-adm profile realtime
`,
			logGroupName)
	}).(pulumi.StringOutput)

	// Create EC2 instance with a larger root volume
	instanceName := "ws-server-instance"
	instance, err := ec2.NewInstance(ctx, instanceName, &ec2.InstanceArgs{
		Ami:                      pulumi.String(ssmParam.Value),
		InstanceType:             pulumi.String(cfg.InstanceType),
		KeyName:                  pulumi.String(cfg.KeyPairName),
		SubnetId:                 subnet,
		VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
		IamInstanceProfile:       instanceProfile.Name,
		AssociatePublicIpAddress: pulumi.Bool(true),
		UserData:                 userData,
		AvailabilityZone:         pulumi.String(config.AvailabilityZone2),
		Tags:                     utils.ApplyTags(ctx, instanceName, utils.GetNamedTags(instanceName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize: pulumi.Int(30), // Increase root volume size to 30 GB
			VolumeType: pulumi.String("gp3"),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return instance, logGroup, nil
}

// createIamRole creates an IAM role for the EC2 instance
func createIamRole(ctx *pulumi.Context, cfg *config.Config) (*iam.Role, *iam.InstanceProfile, error) {
	// Create IAM role
	roleName := "ws-server-role"
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
	_, err = iam.NewRolePolicyAttachment(ctx, "ws-server-ssm-policy", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "ws-server-cloudwatch-policy", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"),
	})
	if err != nil {
		return nil, nil, err
	}

	// ECR policy removed - application will be run manually

	// Create instance profile
	profileName := "ws-server-profile"
	instanceProfile, err := iam.NewInstanceProfile(ctx, profileName, &iam.InstanceProfileArgs{
		Role: role.Name,
		Tags: utils.ApplyTags(ctx, profileName, utils.GetNamedTags(profileName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}

	return role, instanceProfile, nil
}

// createEc2LogGroup creates a CloudWatch log group for the EC2 instance
func createEc2LogGroup(ctx *pulumi.Context, cfg *config.Config) (*cloudwatch.LogGroup, error) {
	logGroupName := "ws-server-logs"
	logGroup, err := cloudwatch.NewLogGroup(ctx, logGroupName, &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("/ec2/ws-server"),
		RetentionInDays: pulumi.Int(7),
		Tags:            utils.ApplyTags(ctx, logGroupName, utils.GetNamedTags(logGroupName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	return logGroup, nil
}
