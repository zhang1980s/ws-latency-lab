# WebSocket Latency Testing Infrastructure

This directory contains the Pulumi infrastructure code for the WebSocket latency testing application. The infrastructure is split into multiple stacks following best practices for better organization, separation of concerns, and easier management.

## Architecture

The infrastructure is divided into four separate stacks:

1. **ECR Stack**: Contains the ECR repository for storing Docker images
2. **Server VPC Stack**: Contains the VPC, subnets, security groups, and EC2 resources for the WebSocket server
3. **Transit VPC Stack**: Contains the VPC, load balancers, and VPC endpoint service for the transit tier
4. **Client VPC Stack**: Contains the VPC, VPC endpoints, and ECS resources for the WebSocket clients

## Directory Structure

```
infrastructure/
├── common/                  # Shared code used across stacks
│   ├── config/              # Configuration code
│   └── utils/               # Utility functions
├── ecr-stack/               # ECR repository stack
├── server-vpc-stack/        # Server VPC stack
├── transit-vpc-stack/       # Transit VPC stack
├── client-vpc-stack/        # Client VPC stack
├── deploy.sh                # Script to deploy all stacks
└── cleanup.sh               # Script to destroy all stacks
```

## Stack Dependencies

The stacks have the following dependencies:

- Client VPC Stack depends on Transit VPC Stack
- Transit VPC Stack depends on Server VPC Stack
- Server VPC Stack depends on ECR Stack

## Deployment

### Using the Deploy Script

The deploy.sh script handles the entire deployment process:

1. Builds the Java application using Maven
2. Creates a Docker image
3. Pushes the Docker image to ECR
4. Deploys all Pulumi stacks in the correct order

To deploy the infrastructure, run the following command:

```bash
./deploy.sh
```

This will deploy the stacks in the correct order:

1. ECR Stack
2. Server VPC Stack
3. Transit VPC Stack
4. Client VPC Stack

### Manual Deployment

If you prefer to build and deploy separately:

```bash
# First, build and push the Docker image
cd ..
make

# Then deploy the infrastructure
cd infrastructure
./deploy.sh
```

## Cleanup

To destroy the infrastructure, run the following command:

```bash
./cleanup.sh
```

This will destroy the stacks in the reverse order:

1. Client VPC Stack
2. Transit VPC Stack
3. Server VPC Stack
4. ECR Stack

## Configuration

Each stack has its own configuration files:

- `Pulumi.yaml`: Contains the stack name and description
- `Pulumi.<stack-name>.yaml`: Contains the stack-specific configuration


## Network Architecture

The infrastructure creates three VPCs:

1. **Server VPC**: Contains the WebSocket server running on EC2
2. **Transit VPC**: Contains the load balancers and VPC endpoint service
3. **Client VPC**: Contains the WebSocket clients running in ECS

The VPCs are connected as follows:

- Server VPC and Transit VPC are connected via VPC peering
- Transit VPC exposes a VPC endpoint service
- Client VPC connects to the Transit VPC via a VPC endpoint

This architecture allows for testing WebSocket latency across different network boundaries.

## EC2 Instance Configuration

The WebSocket server runs on an EC2 instance with the following configuration:

- Instance Type: m6i.8xlarge
- AMI: Amazon Linux 2023 minimal with kernel-6.12
- Region: ap-east-1
- Availability Zone: ap-east-1b
- Key Pair: keypair-sandbox0-hkg

The AMI is dynamically selected using AWS SSM Parameter Store, which ensures that the latest version of the AMI is always used. The parameter path is:

```
/aws/service/ami-amazon-linux-latest/al2023-ami-minimal-kernel-6.12-x86_64
```

## Docker Container Deployment

The EC2 instance is configured to run the WebSocket server in a Docker container. The user data script:

1. Installs Docker
2. Installs the CloudWatch agent for logging
3. Creates a systemd service for the WebSocket server
4. Pulls the WebSocket server image from ECR
5. Runs the container with the appropriate environment variables and port mappings
6. Sets up a cron job to check and restart the container if it's not running

This ensures that the WebSocket server is automatically started when the EC2 instance boots up and is restarted if it crashes.

## Load Balancer Configuration

The infrastructure includes both Application Load Balancer (ALB) and Network Load Balancer (NLB):

- ALB: Used for HTTP/HTTPS traffic, internal to the transit VPC
- NLB: Used for TLS traffic, exposed to the client VPC via VPC endpoint service

## Common Code Structure

The `common` directory contains shared code used across all stacks:

1. **Configuration**: Constants, environment variables, and configuration structures
2. **Utilities**: Helper functions for resource naming, tagging, and other common operations

This structure promotes code reusability and maintainability across the stacks.

## Building and Pushing Docker Images

A Makefile is provided in the root directory to simplify building the Java application, creating Docker images, and pushing them to the ECR repository:

```bash
# Build the Java application, Docker image, and push to ECR
make

# Build the Java application
make build

# Build the Docker image
make docker-build

# Push the Docker image to ECR
make ecr-push

# For more options
make help
```

The Docker image is tagged with both the version number and `latest` tag, and the ECR repository is configured with a lifecycle policy to keep only the last 10 images.

The deploy.sh script automatically uses the Makefile to build and push the Docker image before deploying the infrastructure.