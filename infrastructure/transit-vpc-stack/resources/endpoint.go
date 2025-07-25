package resources

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateVpcEndpointService creates a VPC endpoint service for the NLB
func CreateVpcEndpointService(ctx *pulumi.Context, cfg *config.Config, nlb *lb.LoadBalancer) (*ec2.VpcEndpointService, error) {
	// Create VPC endpoint service
	endpointServiceName := config.FormatResourceName(cfg.Project, "endpoint-service")
	endpointService, err := ec2.NewVpcEndpointService(ctx, endpointServiceName, &ec2.VpcEndpointServiceArgs{
		AcceptanceRequired:      pulumi.Bool(false),   // Auto-accept connections for testing purposes
		AllowedPrincipals:       pulumi.StringArray{}, // In production, you would specify AWS principals allowed to connect
		NetworkLoadBalancerArns: pulumi.StringArray{nlb.Arn},
		Tags:                    utils.ApplyTags(ctx, endpointServiceName, utils.GetNamedTags(endpointServiceName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	// Note: VPC endpoint service permissions are now managed through the AllowedPrincipals property
	// in the VpcEndpointServiceArgs above. The separate permission resource is no longer needed.

	return endpointService, nil
}
