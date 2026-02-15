#!/bin/bash

# Script to help setup GitHub Secrets for ECR CI/CD
# This script provides instructions and validates AWS setup

set -e

echo "========================================"
echo "GitHub Secrets Setup Guide"
echo "========================================"
echo ""

# Check if gh CLI is installed
if command -v gh &> /dev/null; then
    echo "✅ GitHub CLI (gh) is installed"
    GH_INSTALLED=true
else
    echo "⚠️  GitHub CLI (gh) is not installed"
    echo "   Install from: https://cli.github.com/"
    GH_INSTALLED=false
fi

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed"
    echo "   Install from: https://aws.amazon.com/cli/"
    exit 1
fi

echo "✅ AWS CLI is installed"

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials are not configured"
    echo ""
    echo "Configure AWS credentials:"
    echo "  aws configure"
    exit 1
fi

echo "✅ AWS credentials are configured"

# Get AWS account info
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
AWS_USER=$(aws sts get-caller-identity --query Arn --output text)

echo ""
echo "Current AWS User: $AWS_USER"
echo "AWS Account ID: $ACCOUNT_ID"

# Create IAM user for GitHub Actions (if needed)
echo ""
echo "========================================"
echo "Step 1: Create IAM User for GitHub Actions"
echo "========================================"
echo ""

IAM_USER="github-actions-ecr"
echo "Checking if IAM user '$IAM_USER' exists..."

if aws iam get-user --user-name "$IAM_USER" &> /dev/null; then
    echo "✅ IAM user '$IAM_USER' already exists"
else
    echo "⚠️  IAM user '$IAM_USER' does not exist"
    echo ""
    read -p "Do you want to create it? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        aws iam create-user --user-name "$IAM_USER"
        echo "✅ Created IAM user: $IAM_USER"
    fi
fi

# Create ECR policy
POLICY_NAME="ECRPushPolicy"
POLICY_DOC='{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": "*"
    }
  ]
}'

echo ""
echo "========================================"
echo "Step 2: Create/Attach ECR Policy"
echo "========================================"

# Check if policy exists
POLICY_ARN="arn:aws:iam::${ACCOUNT_ID}:policy/${POLICY_NAME}"
if aws iam get-policy --policy-arn "$POLICY_ARN" &> /dev/null; then
    echo "✅ Policy '$POLICY_NAME' already exists"
else
    echo "⚠️  Policy '$POLICY_NAME' does not exist"
    echo ""
    read -p "Do you want to create it? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "$POLICY_DOC" > /tmp/ecr-policy.json
        aws iam create-policy \
            --policy-name "$POLICY_NAME" \
            --policy-document file:///tmp/ecr-policy.json
        rm /tmp/ecr-policy.json
        echo "✅ Created policy: $POLICY_NAME"
    fi
fi

# Attach policy to user
echo ""
echo "Checking if policy is attached to user..."
if aws iam list-attached-user-policies --user-name "$IAM_USER" 2>/dev/null | grep -q "$POLICY_NAME"; then
    echo "✅ Policy is already attached to user"
else
    echo "⚠️  Policy is not attached to user"
    echo ""
    read -p "Do you want to attach it? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        aws iam attach-user-policy \
            --user-name "$IAM_USER" \
            --policy-arn "$POLICY_ARN"
        echo "✅ Attached policy to user"
    fi
fi

# Create access key
echo ""
echo "========================================"
echo "Step 3: Create Access Key"
echo "========================================"
echo ""
echo "⚠️  WARNING: The access key will be displayed only once!"
echo ""
read -p "Do you want to create a new access key for '$IAM_USER'? (y/n): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    ACCESS_KEY_OUTPUT=$(aws iam create-access-key --user-name "$IAM_USER")
    ACCESS_KEY_ID=$(echo "$ACCESS_KEY_OUTPUT" | grep -o '"AccessKeyId": "[^"]*' | cut -d'"' -f4)
    SECRET_ACCESS_KEY=$(echo "$ACCESS_KEY_OUTPUT" | grep -o '"SecretAccessKey": "[^"]*' | cut -d'"' -f4)

    echo ""
    echo "✅ Access key created successfully!"
    echo ""
    echo "========================================"
    echo "⚠️  SAVE THESE CREDENTIALS NOW!"
    echo "========================================"
    echo "AWS_ACCESS_KEY_ID: $ACCESS_KEY_ID"
    echo "AWS_SECRET_ACCESS_KEY: $SECRET_ACCESS_KEY"
    echo "========================================"
    echo ""

    # Optionally set secrets using gh CLI
    if [ "$GH_INSTALLED" = true ]; then
        echo ""
        read -p "Do you want to set these as GitHub secrets now? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            gh auth status > /dev/null 2>&1 || gh auth login

            echo "$ACCESS_KEY_ID" | gh secret set AWS_ACCESS_KEY_ID
            echo "$SECRET_ACCESS_KEY" | gh secret set AWS_SECRET_ACCESS_KEY

            echo "✅ GitHub secrets set successfully!"
        fi
    fi
fi

# Manual instructions
echo ""
echo "========================================"
echo "Step 4: Set GitHub Secrets Manually"
echo "========================================"
echo ""
echo "If you haven't set the secrets automatically, do it manually:"
echo ""
echo "1. Go to: https://github.com/sagarkolli249/apinto/settings/secrets/actions"
echo "2. Click 'New repository secret'"
echo "3. Add the following secrets:"
echo "   - Name: AWS_ACCESS_KEY_ID"
echo "     Value: [Your Access Key ID]"
echo "   - Name: AWS_SECRET_ACCESS_KEY"
echo "     Value: [Your Secret Access Key]"
echo ""
echo "========================================"
echo "Step 5: Verify Secrets"
echo "========================================"
echo ""
if [ "$GH_INSTALLED" = true ]; then
    echo "Checking GitHub secrets..."
    if gh secret list | grep -q "AWS_ACCESS_KEY_ID"; then
        echo "✅ AWS_ACCESS_KEY_ID is set"
    else
        echo "⚠️  AWS_ACCESS_KEY_ID is not set"
    fi

    if gh secret list | grep -q "AWS_SECRET_ACCESS_KEY"; then
        echo "✅ AWS_SECRET_ACCESS_KEY is set"
    else
        echo "⚠️  AWS_SECRET_ACCESS_KEY is not set"
    fi
else
    echo "Install GitHub CLI to verify secrets automatically"
    echo "Or check manually at: https://github.com/sagarkolli249/apinto/settings/secrets/actions"
fi

echo ""
echo "========================================"
echo "Setup Complete!"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Ensure ECR repository exists (or will be created)"
echo "2. Push a commit with conventional commit format"
echo "3. Check GitHub Actions: https://github.com/sagarkolli249/apinto/actions"
echo "4. Verify build with: ./scripts/verify-ecr-build.sh"
