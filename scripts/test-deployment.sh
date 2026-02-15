#!/bin/bash

# Script to test deployment of Apinto from ECR
# Usage: ./scripts/test-deployment.sh [AWS_REGION] [ECR_REPOSITORY] [TAG]

set -e

# Configuration
AWS_REGION="${1:-us-east-1}"
ECR_PRIVATE_REPOSITORY="${2:-apipark/apinto}"
TAG="${3:-latest}"

echo "========================================"
echo "Apinto Test Deployment Script"
echo "========================================"
echo "AWS Region: $AWS_REGION"
echo "ECR Repository: $ECR_PRIVATE_REPOSITORY"
echo "Tag: $TAG"
echo ""

# Check Docker installation
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    echo "Install: https://docs.docker.com/get-docker/"
    exit 1
fi

echo "✅ Docker is installed"

# Check AWS CLI installation
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed"
    exit 1
fi

echo "✅ AWS CLI is installed"

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_PRIVATE_REPOSITORY}:${TAG}"

echo ""
echo "Image URI: $IMAGE_URI"

# Login to ECR
echo ""
echo "Logging in to ECR..."
aws ecr get-login-password --region "$AWS_REGION" | \
    docker login --username AWS --password-stdin \
    "${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

if [ $? -eq 0 ]; then
    echo "✅ Successfully logged in to ECR"
else
    echo "❌ Failed to login to ECR"
    exit 1
fi

# Pull the image
echo ""
echo "Pulling image from ECR..."
docker pull "$IMAGE_URI"

if [ $? -eq 0 ]; then
    echo "✅ Successfully pulled image"
else
    echo "❌ Failed to pull image"
    exit 1
fi

# Inspect the image
echo ""
echo "========================================"
echo "Image Information"
echo "========================================"
docker inspect "$IMAGE_URI" --format='
Image: {{.RepoTags}}
Architecture: {{.Architecture}}
OS: {{.Os}}
Created: {{.Created}}
Size: {{.Size}} bytes
' || true

# Stop existing container if running
CONTAINER_NAME="apinto-test"
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo ""
    echo "Stopping existing container..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
fi

# Create test directory for volumes
TEST_DIR="/tmp/apinto-test"
mkdir -p "$TEST_DIR"

echo ""
echo "========================================"
echo "Starting Container"
echo "========================================"
echo "Container name: $CONTAINER_NAME"
echo "Data directory: $TEST_DIR"
echo ""

# Run the container
docker run -d \
    --name "$CONTAINER_NAME" \
    -p 8080:8080 \
    -v "$TEST_DIR:/var/lib/apinto" \
    "$IMAGE_URI"

if [ $? -eq 0 ]; then
    echo "✅ Container started successfully"
else
    echo "❌ Failed to start container"
    exit 1
fi

# Wait for container to be ready
echo ""
echo "Waiting for container to be ready..."
sleep 5

# Check container status
echo ""
echo "========================================"
echo "Container Status"
echo "========================================"
docker ps --filter "name=$CONTAINER_NAME" --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"

# Check container logs
echo ""
echo "========================================"
echo "Container Logs (last 20 lines)"
echo "========================================"
docker logs --tail 20 "$CONTAINER_NAME"

# Test if container is healthy
echo ""
echo "========================================"
echo "Health Check"
echo "========================================"

# Wait a bit more for the service to fully start
sleep 5

# Check if port is listening
if docker exec "$CONTAINER_NAME" sh -c "netcat -z localhost 8080" 2>/dev/null; then
    echo "✅ Service is listening on port 8080"
else
    echo "⚠️  Service might not be ready yet or port 8080 is not listening"
fi

# Try to access the service
echo ""
echo "Testing HTTP endpoint..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 2>/dev/null || echo "000")

if [ "$HTTP_CODE" != "000" ]; then
    echo "✅ HTTP endpoint responded with status: $HTTP_CODE"
else
    echo "⚠️  Could not connect to HTTP endpoint"
    echo "   This might be normal if the gateway needs additional configuration"
fi

# Show running processes in container
echo ""
echo "========================================"
echo "Container Processes"
echo "========================================"
docker exec "$CONTAINER_NAME" ps aux 2>/dev/null || echo "Could not list processes"

echo ""
echo "========================================"
echo "Test Commands"
echo "========================================"
echo "View logs:"
echo "  docker logs -f $CONTAINER_NAME"
echo ""
echo "Execute shell in container:"
echo "  docker exec -it $CONTAINER_NAME /bin/bash"
echo ""
echo "Stop container:"
echo "  docker stop $CONTAINER_NAME"
echo ""
echo "Remove container:"
echo "  docker rm $CONTAINER_NAME"
echo ""
echo "Test API endpoint:"
echo "  curl http://localhost:8080"
echo ""
echo "========================================"
echo "Deployment Test Complete!"
echo "========================================"
