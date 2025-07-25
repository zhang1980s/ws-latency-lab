#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting WebSocket Latency Testing Infrastructure Deployment${NC}"

# Build and push Docker image to ECR
echo -e "${YELLOW}Building and pushing Docker image to ECR...${NC}"
cd ..
make
cd infrastructure

# Function to deploy a stack
deploy_stack() {
    local stack_dir=$1
    local stack_name=$2
    
    echo -e "${YELLOW}Deploying $stack_name stack...${NC}"
    
    cd $stack_dir
    
    # Initialize the stack if needed
    if [ ! -d ".pulumi" ]; then
        echo -e "${YELLOW}Initializing $stack_name stack...${NC}"
        pulumi stack init dev
    fi
    
    # Preview the changes
    echo -e "${YELLOW}Previewing changes for $stack_name stack...${NC}"
    pulumi preview
    
    # Ask for confirmation before deploying
    read -p "Deploy $stack_name stack? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Deploying $stack_name stack...${NC}"
        pulumi up --yes
        echo -e "${GREEN}$stack_name stack deployed successfully!${NC}"
    else
        echo -e "${RED}Skipping $stack_name stack deployment.${NC}"
    fi
    
    cd ..
}

# Deploy ECR stack
deploy_stack "ecr-stack" "ECR"

# Deploy Server VPC stack
deploy_stack "server-vpc-stack" "Server VPC"

# Deploy Transit VPC stack
deploy_stack "transit-vpc-stack" "Transit VPC"

# Deploy Client VPC stack
deploy_stack "client-vpc-stack" "Client VPC"

echo -e "${GREEN}WebSocket Latency Testing Infrastructure Deployment Complete!${NC}"