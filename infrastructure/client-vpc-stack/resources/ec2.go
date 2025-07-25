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

// CreateEc2Resources creates an EC2 instance for the WebSocket client
func CreateEc2Resources(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnet pulumi.IDOutput, sg *ec2.SecurityGroup, endpoint *ec2.VpcEndpoint) (*ec2.Instance, *cloudwatch.LogGroup, error) {
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

	// Get ECR repository reference
	ecrStackRef, err := pulumi.NewStackReference(ctx, config.EcrStackName, nil)
	if err != nil {
		return nil, nil, err
	}

	ecrRepoUrl := ecrStackRef.GetOutput(pulumi.String("repositoryUrl"))

	// Get the latest Amazon Linux 2023 minimal with kernel-6.12 AMI from SSM Parameter Store
	ssmParam, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
		Name: config.AmazonLinux2023SsmParameter,
	})
	if err != nil {
		return nil, nil, err
	}

	// Use the endpoint DNS entries to configure the WebSocket server URL
	endpointDns := endpoint.DnsEntries.Index(pulumi.Int(0)).DnsName().ApplyT(func(dnsName *string) string {
		return *dnsName
	}).(pulumi.StringOutput)

	// Create user data script to set up the WebSocket client
	userData := pulumi.All(ecrRepoUrl, logGroup.Name, endpointDns).ApplyT(func(args []interface{}) string {
		repoUrl := args[0].(string)
		logGroupName := args[1].(string)
		dnsName := args[2].(string)

		return fmt.Sprintf(`#!/bin/bash
# Set hostname to ws-client
hostnamectl set-hostname ws-client
echo "127.0.0.1 ws-client" >> /etc/hosts

# Install Docker on Amazon Linux 2023
dnf install -y docker awscli
systemctl start docker
systemctl enable docker

# Add ec2-user to docker group and apply changes immediately
usermod -a -G docker ec2-user
# Create a script to refresh group membership without requiring logout/login
cat > /home/ec2-user/refresh-docker-group.sh << 'EOF'
#!/bin/bash
# This script refreshes group membership for the current user
newgrp docker << EONG
# Now you can run docker commands without sudo
echo "Docker group membership refreshed. You can now run docker commands without sudo."
EONG
EOF
chmod +x /home/ec2-user/refresh-docker-group.sh
chown ec2-user:ec2-user /home/ec2-user/refresh-docker-group.sh

# Install CloudWatch agent
dnf install -y amazon-cloudwatch-agent
cat > /opt/aws/amazon-cloudwatch-agent/bin/config.json << 'EOF'
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/ws-client.log",
            "log_group_name": "%s",
            "log_stream_name": "{instance_id}"
          }
        ]
      }
    }
  }
}
EOF
systemctl start amazon-cloudwatch-agent
systemctl enable amazon-cloudwatch-agent

# Configure Docker to start on boot
systemctl enable docker

# Extract AWS region and account ID from the ECR repository URL
ECR_URL="%s"
AWS_REGION=$(echo $ECR_URL | cut -d'.' -f4)
AWS_ACCOUNT_ID=$(echo $ECR_URL | cut -d'.' -f1)

# Create a script to authenticate with ECR and pull the container image
cat > /usr/local/bin/pull-ws-client.sh << 'EOF'
#!/bin/bash

# Get ECR login token and authenticate Docker
aws ecr get-login-password --region %s | docker login --username AWS --password-stdin %s

# Pull the latest image
docker pull %s:latest

echo "Docker image pulled successfully. To run the client, use the following command:"
echo "docker run -d --name ws-client --restart always -p %d:%d -e MODE=client -e WS_SERVER_URL=wss://%s:%d -e CLIENT_MODE=true %s:latest"
EOF

chmod +x /usr/local/bin/pull-ws-client.sh

# Run the pull script to download the image
/usr/local/bin/pull-ws-client.sh > /var/log/ws-client.log 2>&1

# Create a README file with instructions for the user
cat > /home/ec2-user/README.md << 'EOFREADME'
# WebSocket Latency Test Client

The Docker image has been pulled and is ready to use.

To run the client, use one of the following methods:

## Method 1: Using environment variables (recommended)
    docker run -d \
      --name ws-client \
      --restart always \
      -p PORT:PORT \
      -e MODE=client \
      -e WS_SERVER_URL=wss://ENDPOINT_DNS:NLB_PORT \
      -e CLIENT_MODE=true \
      REPO_URL:latest

## Method 2: Using command-line arguments
    docker run -d \
      --name ws-client \
      --restart always \
      -p PORT:PORT \
      REPO_URL:latest \
      -m client \
      -s wss://ENDPOINT_DNS:NLB_PORT \
      -d 30

Note: Do not use -m=client format. Use -m client instead (with a space, not an equals sign).

To view logs:

    docker logs -f ws-client

To stop the client:

    docker stop ws-client

To remove the container:

    docker rm ws-client
EOFREADME

# Replace placeholders with actual values
sed -i "s|PORT|%d|g" /home/ec2-user/README.md
sed -i "s|ENDPOINT_DNS|%s|g" /home/ec2-user/README.md
sed -i "s|NLB_PORT|%d|g" /home/ec2-user/README.md
sed -i "s|REPO_URL|%s|g" /home/ec2-user/README.md
EOF

chown ec2-user:ec2-user /home/ec2-user/README.md
`,
			logGroupName,
			repoUrl,
			cfg.Region,
			repoUrl,
			repoUrl,
			config.WebSocketPort, config.WebSocketPort,
			dnsName, config.NlbPort,
			repoUrl,
			config.WebSocketPort,
			dnsName, config.NlbPort,
			repoUrl)
	}).(pulumi.StringOutput)

	// Create EC2 instance with a larger root volume
	instanceName := "ws-client-instance"
	instance, err := ec2.NewInstance(ctx, instanceName, &ec2.InstanceArgs{
		Ami:                      pulumi.String(ssmParam.Value),
		InstanceType:             pulumi.String("m6i.8xlarge"),          // Use m6i.8xlarge as requested
		KeyName:                  pulumi.String("keypair-sandbox0-hkg"), // Use keypair-sandbox0-hkg as requested
		SubnetId:                 subnet,
		VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
		IamInstanceProfile:       instanceProfile.Name,
		AssociatePublicIpAddress: pulumi.Bool(true),
		UserData:                 userData,
		// Don't specify AvailabilityZone, let AWS use the subnet's AZ
		Tags: utils.ApplyTags(ctx, instanceName, utils.GetNamedTags(instanceName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
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
	roleName := "ws-client-role"
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
	_, err = iam.NewRolePolicyAttachment(ctx, "ws-client-ssm-policy", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "ws-client-cloudwatch-policy", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"),
	})
	if err != nil {
		return nil, nil, err
	}

	// Create policy for ECR access
	_, err = iam.NewRolePolicy(ctx, "ws-client-ecr-policy", &iam.RolePolicyArgs{
		Role: role.Name,
		Policy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Action": [
						"ecr:GetDownloadUrlForLayer",
						"ecr:BatchGetImage",
						"ecr:BatchCheckLayerAvailability",
						"ecr:GetAuthorizationToken"
					],
					"Resource": "*"
				}
			]
		}`),
	})
	if err != nil {
		return nil, nil, err
	}

	// Create instance profile
	profileName := "ws-client-profile"
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
	logGroupName := "ws-client-logs"
	logGroup, err := cloudwatch.NewLogGroup(ctx, logGroupName, &cloudwatch.LogGroupArgs{
		Name:            pulumi.String("/ec2/ws-client"),
		RetentionInDays: pulumi.Int(7),
		Tags:            utils.ApplyTags(ctx, logGroupName, utils.GetNamedTags(logGroupName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	return logGroup, nil
}
