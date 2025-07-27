package config

// VpcConfig represents the configuration for a VPC
type VpcConfig struct {
	Name              string
	CidrBlock         string
	SubnetCidrBlocks  []string
	AvailabilityZones []string
}

// GetClientVpcConfig returns the configuration for the client VPC
func GetClientVpcConfig() VpcConfig {
	return VpcConfig{
		Name:              ClientVpcPrefix,
		CidrBlock:         ClientVpcCidr,
		SubnetCidrBlocks:  []string{ClientSubnetCidr},
		AvailabilityZones: []string{AvailabilityZone},
	}
}

// GetTransitVpcConfig returns the configuration for the transit VPC
func GetTransitVpcConfig(cfg *Config) VpcConfig {
	return VpcConfig{
		Name:              TransitVpcPrefix,
		CidrBlock:         TransitVpcCidr,
		SubnetCidrBlocks:  []string{TransitSubnetCidr, TransitSubnetCidr2},
		AvailabilityZones: []string{AvailabilityZone, AvailabilityZone2},
	}
}

// GetServerVpcConfig returns the configuration for the server VPC
func GetServerVpcConfig(cfg *Config) VpcConfig {
	return VpcConfig{
		Name:              ServerVpcPrefix,
		CidrBlock:         ServerVpcCidr,
		SubnetCidrBlocks:  []string{ServerSubnetCidr},
		AvailabilityZones: []string{AvailabilityZone2},
	}
}

// GetNetworkNlbConfig returns the configuration for the Network Load Balancer
func GetNetworkNlbConfig() LoadBalancerConfig {
	return LoadBalancerConfig{
		Name:            "ws-nlb",
		Type:            "network",
		Port:            NlbPort,
		Protocol:        "TCP",
		Internal:        false,
		HealthCheckPath: "/health",
	}
}

// GetNetworkAlbConfig returns the configuration for the Application Load Balancer
func GetNetworkAlbConfig() LoadBalancerConfig {
	return LoadBalancerConfig{
		Name:            "ws-alb",
		Type:            "application",
		Port:            AlbPort,
		Protocol:        "HTTPS",
		Internal:        false, // Changed from true to false to make ALB public
		HealthCheckPath: "/health",
	}
}

// Ec2Config represents the configuration for an EC2 instance
type Ec2Config struct {
	Name           string
	InstanceType   string
	KeyName        string
	UserDataScript string
}

// GetClientEc2Config returns the configuration for the client EC2 instance
func GetClientEc2Config() Ec2Config {
	return Ec2Config{
		Name:         "ws-client",
		InstanceType: ClientInstanceType,
		KeyName:      KeyPairName,
		UserDataScript: `#!/bin/bash
echo "Setting up WebSocket client..."
hostnamectl set-hostname ws-client
echo "127.0.0.1 ws-client" >> /etc/hosts
`,
	}
}

// GetServerEc2Config returns the configuration for the server EC2 instance
func GetServerEc2Config() Ec2Config {
	return Ec2Config{
		Name:         "ws-server",
		InstanceType: ServerInstanceType,
		KeyName:      KeyPairName,
		UserDataScript: `#!/bin/bash
echo "Setting up WebSocket server..."
hostnamectl set-hostname ws-server
echo "127.0.0.1 ws-server" >> /etc/hosts
`,
	}
}

// EcrConfig represents the configuration for an ECR repository
// Deprecated: Kept for backward compatibility, application will be run manually
type EcrConfig struct {
	Name string
}

// GetEcrConfig returns the configuration for the ECR repository
// Deprecated: Kept for backward compatibility, application will be run manually
func GetEcrConfig() EcrConfig {
	return EcrConfig{
		Name: "ws-server",
	}
}
