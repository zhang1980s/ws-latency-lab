package resources

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateVpc creates a VPC with a single subnet in ap-east-1b for the server
func CreateVpc(ctx *pulumi.Context, cfg *config.Config) (*ec2.Vpc, []pulumi.IDOutput, *ec2.RouteTable, error) {
	// Get VPC configuration
	vpcConfig := config.GetServerVpcConfig()

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

	// Create a single subnet in ap-east-1b
	var subnets []pulumi.IDOutput
	subnetName := config.FormatResourceName(vpcConfig.Name, "subnet", "ap-east-1b")

	subnet, err := ec2.NewSubnet(ctx, subnetName, &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String(config.ServerSubnetCidr),
		AvailabilityZone: pulumi.String(config.AvailabilityZone), // ap-east-1b
		Tags:             utils.ApplyTags(ctx, subnetName, utils.GetNamedTags(subnetName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	subnets = append(subnets, subnet.ID())

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

	// Associate route table with subnet
	_, err = ec2.NewRouteTableAssociation(ctx, config.FormatResourceName(vpcConfig.Name, "rt-assoc", "public"), &ec2.RouteTableAssociationArgs{
		SubnetId:     subnet.ID(),
		RouteTableId: rt.ID(),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return vpc, subnets, rt, nil
}

// AddRouteToTransitVpc adds a route from the server VPC to the transit VPC via the VPC peering connection
func AddRouteToTransitVpc(ctx *pulumi.Context, cfg *config.Config, routeTable *ec2.RouteTable, vpcPeeringId pulumi.StringInput) error {
	// Get VPC configurations
	serverVpcConfig := config.GetServerVpcConfig()
	transitVpcConfig := config.GetTransitVpcConfig()

	// Create route from server VPC to transit VPC
	_, err := ec2.NewRoute(ctx, config.FormatResourceName(serverVpcConfig.Name, "route", "to-transit"), &ec2.RouteArgs{
		RouteTableId:           routeTable.ID(),
		DestinationCidrBlock:   pulumi.String(transitVpcConfig.CidrBlock),
		VpcPeeringConnectionId: vpcPeeringId,
	})
	if err != nil {
		return err
	}

	return nil
}
