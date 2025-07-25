package resources

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// CreateVpc creates a VPC with public and private subnets for the client
func CreateVpc(ctx *pulumi.Context, cfg *config.Config) (*ec2.Vpc, []pulumi.IDOutput, error) {
	// Create VPC
	vpcName := config.FormatResourceName(cfg.Project, "client-vpc")
	vpc, err := ec2.NewVpc(ctx, vpcName, &ec2.VpcArgs{
		CidrBlock:          pulumi.String("10.3.0.0/16"), // Different CIDR from server and middle VPCs
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags:               utils.ApplyTags(ctx, vpcName, utils.GetNamedTags(vpcName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}

	// Create an internet gateway
	igwName := config.FormatResourceName(vpcName, "igw")
	igw, err := ec2.NewInternetGateway(ctx, igwName, &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags:  utils.ApplyTags(ctx, igwName, utils.GetNamedTags(igwName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}

	// Create only one public subnet in ap-east-1b for the client instance
	var subnets []pulumi.IDOutput
	var publicSubnets []*ec2.Subnet

	// Create public subnet in ap-east-1b
	publicSubnetName := config.FormatResourceName(vpcName, "public", "b")
	publicSubnet, err := ec2.NewSubnet(ctx, publicSubnetName, &ec2.SubnetArgs{
		VpcId:               vpc.ID(),
		CidrBlock:           pulumi.String("10.3.0.0/24"),
		AvailabilityZone:    pulumi.Sprintf("%s%s", cfg.Region, "b"), // Explicitly use ap-east-1b
		MapPublicIpOnLaunch: pulumi.Bool(true),
		Tags:                utils.ApplyTags(ctx, publicSubnetName, utils.GetNamedTags(publicSubnetName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}
	subnets = append(subnets, publicSubnet.ID())
	publicSubnets = append(publicSubnets, publicSubnet)

	// Create route table for public subnets
	publicRtName := config.FormatResourceName(vpcName, "public-rt")
	publicRt, err := ec2.NewRouteTable(ctx, publicRtName, &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags:  utils.ApplyTags(ctx, publicRtName, utils.GetNamedTags(publicRtName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, nil, err
	}

	// Create route to internet gateway
	_, err = ec2.NewRoute(ctx, config.FormatResourceName(publicRtName, "igw"), &ec2.RouteArgs{
		RouteTableId:         publicRt.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            igw.ID(),
	})
	if err != nil {
		return nil, nil, err
	}

	// Associate public subnet with public route table
	_, err = ec2.NewRouteTableAssociation(ctx, config.FormatResourceName(publicRtName, "assoc", "b"), &ec2.RouteTableAssociationArgs{
		SubnetId:     publicSubnet.ID(),
		RouteTableId: publicRt.ID(),
	})
	if err != nil {
		return nil, nil, err
	}

	return vpc, subnets, nil
}
