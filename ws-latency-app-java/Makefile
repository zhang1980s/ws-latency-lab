# Makefile for WebSocket Latency Testing Application

# Variables
APP_NAME := ws-latency-app
VERSION := 1.0.0
AWS_REGION := ap-east-1
AWS_ACCOUNT_ID := $(shell aws sts get-caller-identity --query Account --output text)
ECR_REPO_NAME := ws-server
ECR_REPO_URL := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(ECR_REPO_NAME)
JAVA_VERSION := 17

# Default target
.PHONY: all
all: build docker-build ecr-push

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning up..."
	cd ws-latency-app-java && mvn clean
	rm -rf target/

# Build Java application
.PHONY: build
build:
	@echo "Building Java application..."
	cd ws-latency-app-java && mvn clean package
	@echo "Build complete."

# Setup Docker BuildX for cross-platform builds
.PHONY: setup-buildx
setup-buildx:
	@echo "Setting up Docker BuildX..."
	docker buildx create --name multiarch-builder --use || true
	docker buildx inspect --bootstrap
	@echo "BuildX setup complete."

# Build Docker image
.PHONY: docker-build
docker-build: build setup-buildx
	@echo "Building Docker image for x86_64 architecture using BuildX..."
	cd ws-latency-app-java && docker buildx build \
		--platform linux/amd64 \
		--load \
		-t $(APP_NAME):$(VERSION) \
		-t $(APP_NAME):latest .
	@echo "Docker image built successfully."

# Login to ECR
.PHONY: ecr-login
ecr-login:
	@echo "Logging in to ECR..."
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
	@echo "Login successful."

# Tag Docker image for ECR
.PHONY: ecr-tag
ecr-tag: docker-build
	@echo "Tagging Docker image for ECR..."
	docker tag $(APP_NAME):$(VERSION) $(ECR_REPO_URL):$(VERSION)
	docker tag $(APP_NAME):$(VERSION) $(ECR_REPO_URL):latest
	@echo "Image tagged successfully."

# Push Docker image to ECR
.PHONY: ecr-push
ecr-push: ecr-login ecr-tag
	@echo "Pushing Docker image to ECR..."
	docker push $(ECR_REPO_URL):$(VERSION)
	docker push $(ECR_REPO_URL):latest
	@echo "Image pushed successfully to $(ECR_REPO_URL)."

# Help target
.PHONY: help
help:
	@echo "WebSocket Latency Testing Application Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all          : Build Java app, Docker image, and push to ECR (default)"
	@echo "  clean        : Clean up build artifacts"
	@echo "  build        : Build Java application"
	@echo "  setup-buildx : Setup Docker BuildX for cross-platform builds"
	@echo "  docker-build : Build Docker image"
	@echo "  ecr-login    : Login to ECR"
	@echo "  ecr-tag      : Tag Docker image for ECR"
	@echo "  ecr-push     : Push Docker image to ECR"
	@echo "  help         : Show this help message"
	@echo ""
	@echo "Example usage:"
	@echo "  make build           # Build Java application"
	@echo "  make docker-build    # Build Docker image"
	@echo "  make ecr-push        # Push Docker image to ECR"
	@echo "  make                 # Build and push to ECR (all of the above)"