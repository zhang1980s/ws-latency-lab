package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/client-vpc-stack/resources"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// Get transit VPC reference
		transitVpcRef, err := pulumi.NewStackReference(ctx, config.TransitVpcStackName, nil)
		if err != nil {
			return err
		}

		// Get endpoint service name from transit VPC stack
		endpointServiceName := transitVpcRef.GetOutput(pulumi.String("endpointServiceName"))

		// Create VPC
		vpc, subnets, err := resources.CreateVpc(ctx, cfg)
		if err != nil {
			return err
		}

		// Create security groups
		sgClient, sgEndpoint, err := resources.CreateSecurityGroups(ctx, cfg, vpc)
		if err != nil {
			return err
		}

		// Create VPC endpoint to connect to the endpoint service
		endpoint, err := resources.CreateVpcEndpoint(ctx, cfg, vpc, subnets, sgEndpoint, endpointServiceName)
		if err != nil {
			return err
		}

		// Create EC2 instance for client
		instance, logGroup, err := resources.CreateEc2Resources(ctx, cfg, vpc, subnets[0], sgClient, endpoint)
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("vpcId", vpc.ID())

		// Convert subnet IDs to a pulumi.Array for export
		var subnetArray pulumi.Array
		for _, subnet := range subnets {
			subnetArray = append(subnetArray, subnet)
		}
		ctx.Export("subnetIds", subnetArray)
		ctx.Export("clientSecurityGroupId", sgClient.ID())
		ctx.Export("endpointSecurityGroupId", sgEndpoint.ID())
		ctx.Export("endpointId", endpoint.ID())
		ctx.Export("endpointDnsEntries", endpoint.DnsEntries)
		ctx.Export("instanceId", instance.ID())
		ctx.Export("instancePublicIp", instance.PublicIp)
		ctx.Export("instancePrivateIp", instance.PrivateIp)
		ctx.Export("logGroupName", logGroup.Name)

		return nil
	})
}
