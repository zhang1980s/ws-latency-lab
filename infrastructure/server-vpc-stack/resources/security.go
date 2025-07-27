package resources

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateSecurityGroups creates security groups for the server VPC
func CreateSecurityGroups(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc) (*ec2.SecurityGroup, error) {
	// Get VPC configuration
	vpcConfig := config.GetServerVpcConfig(cfg)

	// Create security group for the server
	sgName := config.GetSecurityGroupName(vpcConfig.Name, "server")
	sg, err := ec2.NewSecurityGroup(ctx, sgName, &ec2.SecurityGroupArgs{
		VpcId:       vpc.ID(),
		Description: pulumi.String("Security group for WebSocket server"),
		Ingress: ec2.SecurityGroupIngressArray{
			// Allow SSH from anywhere (for management)
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(22),
				ToPort:      pulumi.Int(22),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("SSH access"),
			},
			// Allow WebSocket traffic from transit VPC only
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(config.WebSocketPort),
				ToPort:      pulumi.Int(config.WebSocketPort),
				CidrBlocks:  pulumi.StringArray{pulumi.String(config.TransitVpcCidr)},
				Description: pulumi.String(fmt.Sprintf("WebSocket traffic on port %d from transit VPC", config.WebSocketPort)),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			// Allow all outbound traffic
			&ec2.SecurityGroupEgressArgs{
				Protocol:    pulumi.String("-1"),
				FromPort:    pulumi.Int(0),
				ToPort:      pulumi.Int(0),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("Allow all outbound traffic"),
			},
		},
		Tags: utils.ApplyTags(ctx, sgName, utils.GetNamedTags(sgName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	return sg, nil
}
