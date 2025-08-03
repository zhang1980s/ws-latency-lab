package resources

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateSecurityGroups creates security groups for the client VPC
func CreateSecurityGroups(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc) (*ec2.SecurityGroup, error) {
	// Create security group for client containers
	sgName := config.FormatResourceName(cfg.Project, "client-sg")
	sg, err := ec2.NewSecurityGroup(ctx, sgName, &ec2.SecurityGroupArgs{
		VpcId:       vpc.ID(),
		Description: pulumi.String("Security group for WebSocket latency test client"),
		Ingress: ec2.SecurityGroupIngressArray{
			// Allow SSH access from anywhere (for testing purposes only)
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(22),
				ToPort:      pulumi.Int(22),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("SSH access"),
			},
			// Removed WebSocket port 10443 as it's not needed
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
