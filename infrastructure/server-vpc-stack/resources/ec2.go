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
echo "127.0.0.1 ws-server" >> /etc/hosts

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
            "file_path": "/var/log/ws-server.log",
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

# Create a script to run the WebSocket server
cat > /usr/local/bin/start-ws-server.sh << 'EOF'
#!/bin/bash

# Note: This script assumes the WebSocket server application is available locally
# Since we're not using ECR anymore, you'll need to manually transfer and run the application

# Example of how to run the application directly (adjust as needed)
# java -jar /path/to/ws-latency-app.jar -mode=server -port=%d -rate=10 > /var/log/ws-server.log 2>&1

# Example of how to run with Docker if you have a local image
# docker run -d \
#   --name ws-server \
#   --restart always \
#   -p %d:%d \
#   -e MODE=server \
#   -e PORT=%d \
#   -e EVENTS_PER_SECOND=10 \
#   your-local-image:latest > /var/log/ws-server.log 2>&1

echo "WebSocket server script placeholder - manual setup required" > /var/log/ws-server.log
EOF

chmod +x /usr/local/bin/start-ws-server.sh

# Create a systemd service for the WebSocket server
cat > /etc/systemd/system/ws-server.service << EOF
[Unit]
Description=WebSocket Latency Testing Server
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/bin/start-ws-server.sh
ExecStop=/usr/bin/docker stop ws-server
ExecStopPost=/usr/bin/docker rm ws-server

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the WebSocket server service
systemctl enable ws-server
systemctl start ws-server

# Run the start script immediately to start the server
/usr/local/bin/start-ws-server.sh

# Set up a cron job to check and restart the container if it's not running
cat > /etc/cron.d/ws-server-check << EOF
*/5 * * * * root docker ps | grep ws-server || /usr/local/bin/start-ws-server.sh
EOF
chmod 644 /etc/cron.d/ws-server-check
`,
			logGroupName,
			config.WebSocketPort,
			config.WebSocketPort, config.WebSocketPort,
			config.WebSocketPort)
	}).(pulumi.StringOutput)

	// Create EC2 instance with a larger root volume
	instanceName := "ws-server-instance"
	instance, err := ec2.NewInstance(ctx, instanceName, &ec2.InstanceArgs{
		Ami:                      pulumi.String(ssmParam.Value),
		InstanceType:             pulumi.String(config.ServerInstanceType),
		KeyName:                  pulumi.String(cfg.KeyPairName),
		SubnetId:                 subnet,
		VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
		IamInstanceProfile:       instanceProfile.Name,
		AssociatePublicIpAddress: pulumi.Bool(true),
		UserData:                 userData,
		AvailabilityZone:         pulumi.String(config.AvailabilityZone),
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

	// Create policy for ECR access
	_, err = iam.NewRolePolicy(ctx, "ws-server-ecr-policy", &iam.RolePolicyArgs{
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
