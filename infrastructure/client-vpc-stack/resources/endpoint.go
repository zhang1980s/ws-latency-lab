package resources

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateVpcEndpoint creates a VPC endpoint to connect to the endpoint service
func CreateVpcEndpoint(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, subnets []pulumi.IDOutput, sg *ec2.SecurityGroup, endpointServiceName pulumi.Output) (*ec2.VpcEndpoint, error) {
	// Create VPC endpoint
	endpointName := config.FormatResourceName(cfg.Project, "client-endpoint")

	// Use the only subnet we have (public subnet in ap-east-1b)
	// The VPC endpoint service should be compatible with ap-east-1b
	subnetIDs := pulumi.StringArray{}
	if len(subnets) > 0 {
		subnetIDs = append(subnetIDs, subnets[0].ToStringOutput())
	}

	// Create the VPC endpoint
	endpoint, err := ec2.NewVpcEndpoint(ctx, endpointName, &ec2.VpcEndpointArgs{
		VpcId: vpc.ID(),
		// Convert endpointServiceName from pulumi.Output to pulumi.StringInput
		ServiceName: endpointServiceName.ApplyT(func(name interface{}) string {
			return name.(string)
		}).(pulumi.StringOutput),
		VpcEndpointType:   pulumi.String("Interface"),
		SubnetIds:         subnetIDs,
		SecurityGroupIds:  pulumi.StringArray{sg.ID()},
		PrivateDnsEnabled: pulumi.Bool(false), // Set to false as the endpoint service doesn't provide a private DNS name
		Tags:              utils.ApplyTags(ctx, endpointName, utils.GetNamedTags(endpointName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}
