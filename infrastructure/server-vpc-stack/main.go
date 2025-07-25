package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/server-vpc-stack/resources"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// Get ECR repository reference
		ecrStackRef, err := pulumi.NewStackReference(ctx, config.EcrStackName, nil)
		if err != nil {
			return err
		}

		repositoryUrl := ecrStackRef.GetOutput(pulumi.String("repositoryUrl"))

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

		// Create EC2 instance
		instance, logGroup, err := resources.CreateEc2Resources(ctx, cfg, vpc, subnets[0], sg, repositoryUrl)
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
		ctx.Export("instanceId", instance.ID())
		ctx.Export("instancePublicIp", instance.PublicIp)
		ctx.Export("instancePrivateIp", instance.PrivateIp)
		ctx.Export("logGroupName", logGroup.Name)

		// Export route table ID for transit VPC stack to use
		ctx.Export("routeTableId", routeTable.ID())

		return nil
	})
}
