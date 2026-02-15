#!/bin/bash

# Setup ECR Authentication for Kubernetes
# This script helps configure your Kubernetes cluster to pull images from ECR

set -e

# Configuration
ECR_REGISTRY="865783518572.dkr.ecr.us-east-1.amazonaws.com"
ECR_REPOSITORY="apinto-gateway"
AWS_REGION="us-east-1"
NAMESPACE="apinto-gateway"
SECRET_NAME="ecr-secret"

echo "========================================"
echo "ECR Authentication Setup for Kubernetes"
echo "========================================"
echo "Registry: $ECR_REGISTRY"
echo "Region: $AWS_REGION"
echo "Namespace: $NAMESPACE"
echo ""

# Check prerequisites
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed"
    exit 1
fi
echo "✅ kubectl is installed"

if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed"
    exit 1
fi
echo "✅ AWS CLI is installed"

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials not configured"
    echo "Run: aws configure"
    exit 1
fi
echo "✅ AWS credentials configured"

# Check kubectl connection
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster"
    exit 1
fi
echo "✅ Connected to Kubernetes cluster"

# Detect cluster type
CLUSTER_INFO=$(kubectl cluster-info 2>&1 || echo "unknown")
if echo "$CLUSTER_INFO" | grep -q "eks.amazonaws.com"; then
    CLUSTER_TYPE="EKS"
    echo "✅ Detected AWS EKS cluster"
elif echo "$CLUSTER_INFO" | grep -q "gke"; then
    CLUSTER_TYPE="GKE"
    echo "✅ Detected GCP GKE cluster"
else
    CLUSTER_TYPE="OTHER"
    echo "✅ Detected Kubernetes cluster (non-cloud)"
fi

# Check node architecture
echo ""
echo "Checking node architecture..."
ARCH=$(kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.architecture}')
echo "Architecture: $ARCH"

if [ "$ARCH" = "amd64" ] || [ "$ARCH" = "x86_64" ]; then
    echo "✅ Nodes are amd64 - use 0.1.0-amd64 image"
    RECOMMENDED_IMAGE="${ECR_REGISTRY}/${ECR_REPOSITORY}:0.1.0-amd64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    echo "✅ Nodes are arm64 - use 0.1.0-arm64 image"
    RECOMMENDED_IMAGE="${ECR_REGISTRY}/${ECR_REPOSITORY}:0.1.0-arm64"
else
    echo "⚠️  Unknown architecture: $ARCH"
    RECOMMENDED_IMAGE="${ECR_REGISTRY}/${ECR_REPOSITORY}:0.1.0-amd64"
fi

echo "Recommended image: $RECOMMENDED_IMAGE"

# Create namespace if it doesn't exist
echo ""
if kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo "✅ Namespace '$NAMESPACE' exists"
else
    echo "Creating namespace '$NAMESPACE'..."
    kubectl create namespace "$NAMESPACE"
    echo "✅ Namespace created"
fi

# Setup authentication based on cluster type
echo ""
echo "========================================"
echo "Setting up ECR Authentication"
echo "========================================"

if [ "$CLUSTER_TYPE" = "EKS" ]; then
    echo ""
    echo "For EKS clusters, we recommend using IAM roles for service accounts (IRSA)."
    echo ""
    read -p "Do you want to use IAM roles (recommended) or image pull secrets? (iam/secret): " -r
    echo

    if [[ $REPLY =~ ^[Ii]am$ ]] || [[ $REPLY =~ ^[Ii]$ ]]; then
        echo "Setting up IAM role for ECR access..."
        echo ""
        echo "Manual steps required:"
        echo "1. Create IAM policy (if not exists):"
        echo ""
        cat <<'EOF'
aws iam create-policy \
    --policy-name ApintoECRReadOnly \
    --policy-document '{
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "ecr:GetAuthorizationToken",
            "ecr:BatchCheckLayerAvailability",
            "ecr:GetDownloadUrlForLayer",
            "ecr:BatchGetImage"
          ],
          "Resource": "*"
        }
      ]
    }'
EOF
        echo ""
        echo "2. Attach policy to your EKS node IAM role:"
        echo ""
        echo "aws iam attach-role-policy \\"
        echo "    --role-name <your-eks-node-role> \\"
        echo "    --policy-arn arn:aws:iam::<account-id>:policy/ApintoECRReadOnly"
        echo ""
        echo "3. Your nodes can now pull from ECR without image pull secrets!"
        echo ""
        exit 0
    fi
fi

# Create image pull secret
echo ""
echo "Creating image pull secret..."

# Delete existing secret if it exists
if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
    echo "Deleting existing secret..."
    kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE"
fi

# Get ECR token
echo "Getting ECR authentication token..."
ECR_TOKEN=$(aws ecr get-login-password --region "$AWS_REGION")

# Create secret
kubectl create secret docker-registry "$SECRET_NAME" \
    --docker-server="$ECR_REGISTRY" \
    --docker-username=AWS \
    --docker-password="$ECR_TOKEN" \
    --namespace="$NAMESPACE"

echo "✅ Image pull secret created: $SECRET_NAME"

# Verify secret
echo ""
echo "Verifying secret..."
if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
    echo "✅ Secret verified"
else
    echo "❌ Secret creation failed"
    exit 1
fi

# Test pulling image
echo ""
echo "========================================"
echo "Testing Image Pull"
echo "========================================"
echo ""
echo "Creating test pod to verify ECR access..."

# Create test pod
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ecr-test
  namespace: $NAMESPACE
spec:
  imagePullSecrets:
  - name: $SECRET_NAME
  containers:
  - name: test
    image: $RECOMMENDED_IMAGE
    command: ["/bin/sh", "-c", "echo 'ECR pull successful!' && sleep 10"]
  restartPolicy: Never
EOF

echo "Waiting for pod to start..."
sleep 5

# Check pod status
POD_STATUS=$(kubectl get pod ecr-test -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "Failed")

if [ "$POD_STATUS" = "Running" ] || [ "$POD_STATUS" = "Succeeded" ]; then
    echo "✅ Image pulled successfully from ECR!"
    kubectl logs ecr-test -n "$NAMESPACE" 2>/dev/null || true
else
    echo "⚠️  Pod status: $POD_STATUS"
    echo "Checking pod details..."
    kubectl describe pod ecr-test -n "$NAMESPACE"
fi

# Cleanup test pod
echo ""
echo "Cleaning up test pod..."
kubectl delete pod ecr-test -n "$NAMESPACE" --ignore-not-found=true

echo ""
echo "========================================"
echo "Setup Complete!"
echo "========================================"
echo ""
echo "Image pull secret created: $SECRET_NAME"
echo "Namespace: $NAMESPACE"
echo "Recommended image: $RECOMMENDED_IMAGE"
echo ""
echo "⚠️  IMPORTANT: ECR tokens expire after 12 hours!"
echo ""
echo "For production, set up automatic token refresh:"
echo "1. Use IAM roles (recommended for EKS)"
echo "2. Use a CronJob to refresh the secret periodically"
echo ""
echo "Next steps:"
echo "1. Update k8s/deployment.yaml with:"
echo "   - image: $RECOMMENDED_IMAGE"
echo "   - imagePullSecrets: [name: $SECRET_NAME]"
echo ""
echo "2. Deploy:"
echo "   kubectl apply -f k8s/deployment.yaml"
echo ""
echo "3. Verify:"
echo "   kubectl get pods -n $NAMESPACE"
echo ""
