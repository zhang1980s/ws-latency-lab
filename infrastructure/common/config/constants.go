package config

// Network constants
const (
	// VPC CIDR blocks
	ClientVpcCidr  = "10.1.0.0/16"
	TransitVpcCidr = "10.2.0.0/16"
	ServerVpcCidr  = "10.3.0.0/16"

	// Subnet CIDR blocks
	ClientSubnetCidr   = "10.1.1.0/24"
	TransitSubnetCidr  = "10.2.1.0/24"
	TransitSubnetCidr2 = "10.2.2.0/24"
	ServerSubnetCidr   = "10.3.1.0/24"

	// Availability zones
	AvailabilityZone  = "ap-east-1b"
	AvailabilityZone2 = "ap-east-1c"

	// Port configurations
	WebSocketPort   = 10443
	HealthCheckPort = 10443 // Using the same port as WebSocketPort since health API and WebSocket API are shared
	MetricsPort     = 9091  // Kept for reference but no longer used
	NlbPort         = 8443
	AlbPort         = 8443
)

// Resource name prefixes
const (
	ClientVpcPrefix  = "ws-client-vpc"
	TransitVpcPrefix = "ws-transit-vpc"
	ServerVpcPrefix  = "ws-server-vpc"
	EcrPrefix        = "ws-server-ecr"
)

// EC2 instance configurations
const (
	ClientInstanceType = "t3.micro"
	ServerInstanceType = "m6i.8xlarge"
	KeyPairName        = "keypair-sandbox0-hkg"

	// SSM parameter name for Amazon Linux 2023 minimal with kernel-6.12 AMI
	AmazonLinux2023SsmParameter = "/aws/service/ami-amazon-linux-latest/al2023-ami-minimal-kernel-6.12-x86_64"
)

// Docker image configurations
const (
	DockerImageTag = "latest"
)

// Stack names
const (
	EcrStackName        = "zhang1980s/ws-ecr-stack/dev"
	ClientVpcStackName  = "zhang1980s/ws-latency-client-vpc/dev"
	TransitVpcStackName = "zhang1980s/ws-latency-transit-vpc/dev"
	ServerVpcStackName  = "zhang1980s/ws-server-vpc-stack/dev"
)
