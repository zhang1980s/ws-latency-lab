#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting WebSocket Latency Testing Infrastructure Cleanup${NC}"

# Function to destroy a stack
destroy_stack() {
    local stack_dir=$1
    local stack_name=$2
    
    echo -e "${YELLOW}Destroying $stack_name stack...${NC}"
    
    cd $stack_dir
    
    # Check if the stack exists
    if [ -d ".pulumi" ]; then
        # Ask for confirmation before destroying
        read -p "Destroy $stack_name stack? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}Destroying $stack_name stack...${NC}"
            pulumi destroy --yes
            echo -e "${GREEN}$stack_name stack destroyed successfully!${NC}"
        else
            echo -e "${RED}Skipping $stack_name stack destruction.${NC}"
        fi
    else
        echo -e "${RED}$stack_name stack does not exist. Skipping.${NC}"
    fi
    
    cd ..
}

# Destroy stacks in reverse order
# Destroy Client VPC stack
destroy_stack "client-vpc-stack" "Client VPC"

# Destroy Transit VPC stack
destroy_stack "transit-vpc-stack" "Transit VPC"

# Destroy Server VPC stack
destroy_stack "server-vpc-stack" "Server VPC"

# Destroy ECR stack
destroy_stack "ecr-stack" "ECR"

echo -e "${GREEN}WebSocket Latency Testing Infrastructure Cleanup Complete!${NC}"