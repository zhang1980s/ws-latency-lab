package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/transit-vpc-stack/resources"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// Get server VPC reference
		serverVpcRef, err := pulumi.NewStackReference(ctx, config.ServerVpcStackName, nil)
		if err != nil {
			return err
		}

		serverVpcId := serverVpcRef.GetOutput(pulumi.String("vpcId"))
		serverSubnetIds := serverVpcRef.GetOutput(pulumi.String("subnetIds"))
		serverSecurityGroupId := serverVpcRef.GetOutput(pulumi.String("securityGroupId"))
		serverRouteTableId := serverVpcRef.GetOutput(pulumi.String("routeTableId"))

		// Create VPC
		vpc, subnets, routeTable, err := resources.CreateVpc(ctx, cfg)
		if err != nil {
			return err
		}

		// Create security groups
		sg, err := resources.CreateSecurityGroups(ctx, cfg, vpc)
		if err != nil {
			return err
		}

		// Create VPC peering with server VPC and add routes in both directions
		peering, err := resources.CreateVpcPeering(ctx, cfg, vpc, routeTable, serverVpcId, serverRouteTableId)
		if err != nil {
			return err
		}

		// Create load balancers
		alb, nlb, err := resources.CreateLoadBalancers(ctx, cfg, vpc, subnets, sg, serverVpcRef, serverVpcId, serverSubnetIds, serverSecurityGroupId)
		if err != nil {
			return err
		}

		// Create VPC endpoint service
		endpointService, err := resources.CreateVpcEndpointService(ctx, cfg, nlb)
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
		ctx.Export("securityGroupId", sg.ID())
		ctx.Export("albDnsName", alb.DnsName)
		ctx.Export("nlbDnsName", nlb.DnsName)
		ctx.Export("vpcPeeringId", peering.ID())
		ctx.Export("endpointServiceName", endpointService.ServiceName)

		return nil
	})
}
