package config

// Network constants
const (
	// VPC CIDR blocks
	ClientVpcCidr  = "10.3.0.0/16"
	TransitVpcCidr = "10.2.0.0/16"
	ServerVpcCidr  = "10.1.0.0/16"

	// Subnet CIDR blocks
	ClientSubnetCidr   = "10.3.2.0/24"
	TransitSubnetCidr  = "10.2.1.0/24"
	TransitSubnetCidr2 = "10.2.2.0/24"
	ServerSubnetCidr   = "10.1.2.0/24"

	// Availability zones
	AvailabilityZone  = "ap-east-1a"
	AvailabilityZone2 = "ap-east-1b"

	// Port configurations
	WebSocketPort   = 10443
	HealthCheckPort = 10443 // Using the same port as WebSocketPort since health API and WebSocket API are shared
	NlbPort         = 8443
	AlbPort         = 8443
)

// Resource name prefixes
const (
	ClientVpcPrefix  = "ws-client-vpc"
	TransitVpcPrefix = "ws-transit-vpc"
	ServerVpcPrefix  = "ws-server-vpc"
	// EcrPrefix removed - application will be run manually
)

// EC2 instance configurations
const (
	ClientInstanceType  = "m7i.8xlarge"
	TransitInstanceType = "m7i.8xlarge"
	ServerInstanceType  = "m7i.8xlarge"
	KeyPairName         = "keypair-root-hkg"

	// SSM parameter name for Amazon Linux 2023 minimal with kernel-6.12 AMI
	AmazonLinux2023SsmParameter = "/aws/service/ami-amazon-linux-latest/al2023-ami-minimal-kernel-6.12-x86_64"
)

// Stack names
const (
	ClientVpcStackName  = "zhang1980s/ws-latency-client-vpc/dev"
	TransitVpcStackName = "zhang1980s/ws-latency-transit-vpc/dev"
	ServerVpcStackName  = "zhang1980s/ws-server-vpc-stack/dev"
)
