package resources

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateVpc creates a VPC with subnets for the transit VPC (load balancers)
func CreateVpc(ctx *pulumi.Context, cfg *config.Config) (*ec2.Vpc, []pulumi.IDOutput, *ec2.RouteTable, error) {
	// Get VPC configuration
	vpcConfig := config.GetTransitVpcConfig(cfg)

	// Create VPC
	vpc, err := ec2.NewVpc(ctx, vpcConfig.Name, &ec2.VpcArgs{
		CidrBlock:          pulumi.String(vpcConfig.CidrBlock),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags:               utils.ApplyTags(ctx, vpcConfig.Name, utils.GetNamedTags(vpcConfig.Name, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create internet gateway
	igw, err := ec2.NewInternetGateway(ctx, config.FormatResourceName(vpcConfig.Name, "igw"), &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags:  utils.ApplyTags(ctx, config.FormatResourceName(vpcConfig.Name, "igw"), utils.GetNamedTags(config.FormatResourceName(vpcConfig.Name, "igw"), cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create subnets
	var subnets []pulumi.IDOutput
	for i, cidr := range vpcConfig.SubnetCidrBlocks {
		az := vpcConfig.AvailabilityZones[i%len(vpcConfig.AvailabilityZones)]
		subnetName := config.GetSubnetName(vpcConfig.Name, az, i)

		subnet, err := ec2.NewSubnet(ctx, subnetName, &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String(cidr),
			AvailabilityZone:    pulumi.String(az),
			MapPublicIpOnLaunch: pulumi.Bool(true), // Enable auto-assign public IP for instances in this subnet
			Tags:                utils.ApplyTags(ctx, subnetName, utils.GetNamedTags(subnetName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
		})
		if err != nil {
			return nil, nil, nil, err
		}

		subnets = append(subnets, subnet.ID())
	}

	// Create route table
	rt, err := ec2.NewRouteTable(ctx, config.GetRouteTableName(vpcConfig.Name, "public"), &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags:  utils.ApplyTags(ctx, config.GetRouteTableName(vpcConfig.Name, "public"), utils.GetNamedTags(config.GetRouteTableName(vpcConfig.Name, "public"), cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create route to internet gateway
	_, err = ec2.NewRoute(ctx, config.FormatResourceName(vpcConfig.Name, "route", "igw"), &ec2.RouteArgs{
		RouteTableId:         rt.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            igw.ID(),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Associate route table with subnets
	for i, subnetId := range subnets {
		// Use string concatenation instead of pulumi.Sprintf
		rtAssocName := config.FormatResourceName(vpcConfig.Name, "rt-assoc", "public", fmt.Sprintf("%d", i))
		_, err = ec2.NewRouteTableAssociation(ctx, rtAssocName, &ec2.RouteTableAssociationArgs{
			SubnetId:     subnetId,
			RouteTableId: rt.ID(),
		})
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return vpc, subnets, rt, nil
}

// CreateVpcPeering creates a VPC peering connection between the transit VPC and server VPC
func CreateVpcPeering(ctx *pulumi.Context, cfg *config.Config, vpc *ec2.Vpc, transitRouteTable *ec2.RouteTable, serverVpcId pulumi.Output, serverRouteTableId pulumi.Output) (*ec2.VpcPeeringConnection, error) {
	// Get VPC configurations
	transitVpcConfig := config.GetTransitVpcConfig(cfg)
	serverVpcConfig := config.GetServerVpcConfig(cfg)

	// Create VPC peering connection
	peeringName := config.GetVpcPeeringName(transitVpcConfig.Name, serverVpcConfig.Name)

	// Convert serverVpcId from pulumi.Output to pulumi.StringInput
	serverVpcIdString := serverVpcId.ApplyT(func(id interface{}) string {
		return id.(string)
	}).(pulumi.StringOutput)

	peering, err := ec2.NewVpcPeeringConnection(ctx, peeringName, &ec2.VpcPeeringConnectionArgs{
		VpcId:      vpc.ID(),
		PeerVpcId:  serverVpcIdString,
		AutoAccept: pulumi.Bool(true),
		Tags:       utils.ApplyTags(ctx, peeringName, utils.GetNamedTags(peeringName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	// Create route from transit VPC to server VPC in the main route table
	_, err = ec2.NewRoute(ctx, config.FormatResourceName(transitVpcConfig.Name, "route", "to-server", "main"), &ec2.RouteArgs{
		RouteTableId:           vpc.MainRouteTableId,
		DestinationCidrBlock:   pulumi.String(config.ServerVpcCidr),
		VpcPeeringConnectionId: peering.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{peering}))
	if err != nil {
		return nil, err
	}

	// Create route from transit VPC to server VPC in the transit route table
	_, err = ec2.NewRoute(ctx, config.FormatResourceName(transitVpcConfig.Name, "route", "to-server", "transit"), &ec2.RouteArgs{
		RouteTableId:           transitRouteTable.ID(),
		DestinationCidrBlock:   pulumi.String(config.ServerVpcCidr),
		VpcPeeringConnectionId: peering.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{peering}))
	if err != nil {
		return nil, err
	}

	// Since we're using a single route table for all subnets in the transit VPC,
	// and we've already added a route to the server VPC in that route table,
	// we don't need to add additional routes for each subnet.
	// The route table is already associated with all subnets in the CreateVpc function.

	// Convert serverRouteTableId from pulumi.Output to pulumi.StringInput
	serverRouteTableIdString := serverRouteTableId.ApplyT(func(id interface{}) string {
		if id == nil {
			ctx.Log.Error("Server route table ID is nil", nil)
			return ""
		}
		rtId, ok := id.(string)
		if !ok {
			ctx.Log.Error(fmt.Sprintf("Expected server route table ID to be string, got %T", id), nil)
			return ""
		}
		return rtId
	}).(pulumi.StringOutput)

	// Create route from server VPC to transit VPC only if serverRouteTableId is not empty
	serverRouteTableIdString.ApplyT(func(rtId string) error {
		if rtId == "" {
			ctx.Log.Warn("Skipping creation of route from server VPC to transit VPC because server route table ID is empty", nil)
			return nil
		}

		_, err := ec2.NewRoute(ctx, config.FormatResourceName(serverVpcConfig.Name, "route", "to-transit"), &ec2.RouteArgs{
			RouteTableId:           pulumi.String(rtId),
			DestinationCidrBlock:   pulumi.String(config.TransitVpcCidr),
			VpcPeeringConnectionId: peering.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{peering}))
		if err != nil {
			return err
		}
		return nil
	})

	return peering, nil
}
