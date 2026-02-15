#!/bin/bash

# Script to verify ECR build and images
# Usage: ./scripts/verify-ecr-build.sh [AWS_REGION] [ECR_REPOSITORY]

set -e

# Configuration
AWS_REGION="${1:-us-east-1}"
ECR_REPOSITORY="${2:-apinto-gateway}"

echo "========================================"
echo "ECR Build Verification Script"
echo "========================================"
echo "AWS Region: $AWS_REGION"
echo "ECR Repository: $ECR_REPOSITORY"
echo ""

# Check AWS CLI installation
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed"
    echo "Install: https://aws.amazon.com/cli/"
    exit 1
fi

echo "✅ AWS CLI is installed"

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials are not configured"
    echo "Run: aws configure"
    exit 1
fi

echo "✅ AWS credentials are configured"
aws sts get-caller-identity

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo ""
echo "AWS Account ID: $ACCOUNT_ID"

# Check if ECR repository exists
echo ""
echo "Checking if ECR repository exists..."
if aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" &> /dev/null; then
    echo "✅ ECR repository '$ECR_REPOSITORY' exists"
else
    echo "❌ ECR repository '$ECR_REPOSITORY' does not exist"
    echo ""
    echo "Create it with:"
    echo "  aws ecr create-repository --repository-name $ECR_REPOSITORY --region $AWS_REGION"
    exit 1
fi

# List images in ECR
echo ""
echo "========================================"
echo "Images in ECR Repository"
echo "========================================"

IMAGES=$(aws ecr describe-images \
    --repository-name "$ECR_REPOSITORY" \
    --region "$AWS_REGION" \
    --query 'sort_by(imageDetails,& imagePushedAt)[-10:].[imageTags[0],imagePushedAt,imageSizeInBytes]' \
    --output table 2>/dev/null || echo "")

if [ -z "$IMAGES" ]; then
    echo "⚠️  No images found in repository"
    echo ""
    echo "Possible reasons:"
    echo "1. GitHub Actions workflow hasn't run yet"
    echo "2. Build failed (check GitHub Actions logs)"
    echo "3. AWS credentials not set in GitHub Secrets"
    echo ""
    echo "Check workflow status at:"
    echo "  https://github.com/sagarkolli249/apinto/actions"
else
    echo "$IMAGES"

    # Get the latest image tag
    LATEST_TAG=$(aws ecr describe-images \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags[0]' \
        --output text 2>/dev/null || echo "")

    if [ -n "$LATEST_TAG" ] && [ "$LATEST_TAG" != "None" ]; then
        echo ""
        echo "✅ Latest image tag: $LATEST_TAG"

        # Check for architecture-specific tags
        echo ""
        echo "Checking for multi-arch support..."
        for arch in amd64 arm64; do
            if aws ecr describe-images \
                --repository-name "$ECR_REPOSITORY" \
                --region "$AWS_REGION" \
                --image-ids imageTag="${LATEST_TAG}-${arch}" \
                &> /dev/null; then
                echo "  ✅ ${arch} image found"
            else
                echo "  ⚠️  ${arch} image not found"
            fi
        done

        # Full image URI
        echo ""
        echo "========================================"
        echo "Image URIs"
        echo "========================================"
        IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}"
        echo "Latest: ${IMAGE_URI}:latest"
        echo "Version: ${IMAGE_URI}:${LATEST_TAG}"
        echo "AMD64: ${IMAGE_URI}:${LATEST_TAG}-amd64"
        echo "ARM64: ${IMAGE_URI}:${LATEST_TAG}-arm64"
    fi
fi

echo ""
echo "========================================"
echo "Next Steps"
echo "========================================"
echo "1. Test deployment:"
echo "   ./scripts/test-deployment.sh $AWS_REGION $ECR_REPOSITORY"
echo ""
echo "2. Pull image locally:"
echo "   aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
echo "   docker pull ${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}:latest"
echo ""
echo "3. Check GitHub Actions:"
echo "   https://github.com/sagarkolli249/apinto/actions"
