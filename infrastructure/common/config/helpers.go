package config

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// FormatResourceName formats a resource name with the given parts
func FormatResourceName(parts ...string) string {
	return strings.Join(parts, "-")
}

// GetSubnetName returns the name for a subnet
func GetSubnetName(vpcName string, az string, index int) string {
	// Extract the last part of the AZ (e.g., "a" from "ap-east-1a")
	azSuffix := az[len(az)-1:]
	return FormatResourceName(vpcName, "subnet", azSuffix, fmt.Sprintf("%d", index))
}

// GetRouteTableName returns the name for a route table
func GetRouteTableName(vpcName string, suffix string) string {
	return FormatResourceName(vpcName, "rt", suffix)
}

// GetSecurityGroupName returns the name for a security group
func GetSecurityGroupName(vpcName string, suffix string) string {
	return FormatResourceName(vpcName, "sg", suffix)
}

// GetLoadBalancerName returns the name for a load balancer
func GetLoadBalancerName(lbConfig LoadBalancerConfig, suffix string) string {
	return FormatResourceName(lbConfig.Name, suffix)
}

// GetTargetGroupName returns the name for a target group
func GetTargetGroupName(lbName string, suffix string) string {
	return FormatResourceName(lbName, "tg", suffix)
}

// GetEc2Name returns the name for an EC2 instance
func GetEc2Name(ec2Config Ec2Config, suffix string) string {
	return FormatResourceName(ec2Config.Name, suffix)
}

// GetEcrName returns the name for an ECR repository
func GetEcrName(ecrConfig EcrConfig, suffix string) string {
	return FormatResourceName(ecrConfig.Name, suffix)
}

// GetVpcEndpointName returns the name for a VPC endpoint
func GetVpcEndpointName(vpcName string, suffix string) string {
	return FormatResourceName(vpcName, "endpoint", suffix)
}

// GetVpcEndpointServiceName returns the name for a VPC endpoint service
func GetVpcEndpointServiceName(vpcName string, suffix string) string {
	return FormatResourceName(vpcName, "endpoint-service", suffix)
}

// GetVpcPeeringName returns the name for a VPC peering connection
func GetVpcPeeringName(vpcName1 string, vpcName2 string) string {
	return FormatResourceName(vpcName1, vpcName2, "peering")
}

// GetStackOutputs gets outputs from a stack reference
func GetStackOutputs(ctx *pulumi.Context, stackName string, outputNames []string) (map[string]pulumi.Output, error) {
	// Get stack reference
	stackRef, err := GetStackReference(ctx, stackName)
	if err != nil {
		return nil, err
	}

	// Get outputs
	outputs := make(map[string]pulumi.Output)
	for _, name := range outputNames {
		output := stackRef.GetOutput(pulumi.String(name))
		outputs[name] = output
	}

	return outputs, nil
}
