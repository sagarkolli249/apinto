# APIPark Deployment Guide - Step by Step

This guide will help you deploy APIPark on Amazon EKS from scratch. No prior knowledge assumed - just follow the steps!

## What You'll Deploy

- **APIPark**: API management platform
- **Apinto**: API Gateway
- **Grafana**: Monitoring dashboard
- **Loki**: Log management
- **InfluxDB**: Metrics database
- **NSQ**: Message queue

## What You Need Before Starting

### 1. Tools on Your Computer

Install these tools first:

```bash
# Check if you have these installed:
kubectl version --client
helm version
aws --version

# If not installed, install them:
# kubectl: https://kubernetes.io/docs/tasks/tools/
# helm: https://helm.sh/docs/intro/install/
# aws cli: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
```

### 2. AWS Resources You Must Have

- **EKS Cluster**: A running Kubernetes cluster on AWS
- **RDS MySQL Database**: External database (we don't create this in Helm)
- **ElastiCache Redis**: External cache (optional, or you can enable internal Redis)
- **Your AWS CLI configured**: Run `aws sts get-caller-identity` to verify

### 3. Storage Class

Check if you have a working storage class:

```bash
# Check existing storage classes
kubectl get storageclass

# You should see 'gp2' or 'gp3' in the list
# If you see gp2, you're good to go!
```

### 4. Access Information

You'll need these details ready:

- Your EKS cluster name
- Your AWS account ID
- Your AWS region
- MySQL database endpoint
- MySQL username and password
- Redis endpoint (if using external Redis)
- Domain names for your services

---

## Step-by-Step Setup

### STEP 1: Set Up Your Environment Variables

Copy and paste this, replacing with YOUR actual values:

```bash
# Set these to YOUR values
export CLUSTER_NAME=ai-eks                    # Your EKS cluster name
export AWS_REGION=us-east-1                   # Your AWS region
export ACCOUNT_ID=865783518572                # Your AWS account ID (12 digits)
export MYSQL_ENDPOINT=apipark.c0bysoi2yx9s.us-east-1.rds.amazonaws.com     # Your RDS endpoint
export MYSQL_USERNAME=admin                   # Your MySQL username
export MYSQL_PASSWORD=YourSecurePassword      # Your MySQL password
export REDIS_ENDPOINT=apipark-redis.cm24bo.ng.0001.use1.cache.amazonaws.com:6379   # Your Redis endpoint

# Verify they're set correctly
echo "Cluster: $CLUSTER_NAME"
echo "Region: $AWS_REGION"
echo "Account: $ACCOUNT_ID"
```

### STEP 2: Connect to Your EKS Cluster

```bash
# Update your kubectl config to connect to the cluster
aws eks update-kubeconfig \
  --name $CLUSTER_NAME \
  --region $AWS_REGION

# Test the connection
kubectl get nodes

# You should see a list of nodes. If you get an error, your cluster isn't accessible.
```

### STEP 3: Install EKS Pod Identity Agent Add-on

This is **REQUIRED** - it allows pods to access AWS services securely:

```bash
# Install EKS Pod Identity Agent
echo "Installing Pod Identity Agent..."
aws eks create-addon \
  --cluster-name $CLUSTER_NAME \
  --addon-name eks-pod-identity-agent \
  --region $AWS_REGION

# Wait 2 minutes for it to install
echo "Waiting for installation to complete..."
sleep 120

# Check the status (should show "ACTIVE")
aws eks describe-addon \
  --cluster-name $CLUSTER_NAME \
  --addon-name eks-pod-identity-agent \
  --region $AWS_REGION \
  --query 'addon.status' \
  --output text

# Verify it's running on all nodes
kubectl get daemonset eks-pod-identity-agent -n kube-system

# You should see all pods ready, like: DESIRED=3, CURRENT=3, READY=3
echo "✓ Pod Identity Agent installed successfully"
```

### STEP 4: (OPTIONAL) Install EBS CSI Driver

**Skip this if you already have gp2 storage class working!**

This is only needed if:
- You're on a very new EKS cluster (1.23+) that doesn't have in-tree volume support
- You want to use gp3 or other new volume types
- You see errors about storage when deploying

```bash
# Check if you need this
kubectl get storageclass gp2

# If the above command shows 'gp2' storage class, you can SKIP this step!

# Only run this if you don't have gp2 or if you want to use gp3:
aws eks create-addon \
  --cluster-name $CLUSTER_NAME \
  --addon-name aws-ebs-csi-driver \
  --region $AWS_REGION

echo "Note: This step is OPTIONAL if gp2 storage already works"
```

### STEP 5: Install NGINX Ingress Controller

This handles incoming traffic to your applications:

```bash
# Add Helm repository
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Install NGINX Ingress
echo "Installing NGINX Ingress Controller..."
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer \
  --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-type"="nlb"

# Wait for it to get a load balancer (takes 2-3 minutes)
echo "Waiting for load balancer to be ready..."
kubectl wait --for=condition=available \
  --timeout=300s \
  deployment/ingress-nginx-controller \
  -n ingress-nginx

# Get the load balancer URL (SAVE THIS - you'll need it for DNS)
echo ""
echo "================================================"
echo "Load Balancer URL (save this for DNS setup):"
kubectl get service ingress-nginx-controller \
  -n ingress-nginx \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
echo ""
echo "================================================"
```

**Important**: Copy the load balancer hostname from above. You'll use it to set up DNS.

### STEP 6: Create IAM Role for Pod Identity

This allows your pods to securely access AWS Systems Manager Parameter Store:

```bash
# Create trust policy file
cat > /tmp/pod-identity-trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "pods.eks.amazonaws.com"
      },
      "Action": [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }
  ]
}
EOF

# Create IAM role
echo "Creating IAM role..."
aws iam create-role \
  --role-name apipark-pod-identity-role \
  --assume-role-policy-document file:///tmp/pod-identity-trust-policy.json

# Create permission policy file
cat > /tmp/ssm-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath"
      ],
      "Resource": "arn:aws:ssm:$AWS_REGION:$ACCOUNT_ID:parameter/apipark/*"
    }
  ]
}
EOF

# Create the policy
echo "Creating IAM policy..."
aws iam create-policy \
  --policy-name apipark-ssm-access \
  --policy-document file:///tmp/ssm-policy.json

# Attach policy to role
echo "Attaching policy to role..."
aws iam attach-role-policy \
  --role-name apipark-pod-identity-role \
  --policy-arn arn:aws:iam::$ACCOUNT_ID:policy/apipark-ssm-access

echo "✓ IAM role created successfully"
```

### STEP 7: Store Your Secrets in AWS Parameter Store

We'll store sensitive information in AWS Systems Manager Parameter Store:

```bash
# Create all required parameters
echo "Storing secrets in AWS Parameter Store..."

# MySQL credentials
aws ssm put-parameter \
  --name /apipark/mysql_username \
  --value "$MYSQL_USERNAME" \
  --type String \
  --region $AWS_REGION \
  --overwrite

aws ssm put-parameter \
  --name /apipark/mysql_password \
  --value "$MYSQL_PASSWORD" \
  --type SecureString \
  --region $AWS_REGION \
  --overwrite

aws ssm put-parameter \
  --name /apipark/mysql_database \
  --value "apipark" \
  --type String \
  --region $AWS_REGION \
  --overwrite

aws ssm put-parameter \
  --name /apipark/mysql_host \
  --value "$MYSQL_ENDPOINT" \
  --type String \
  --region $AWS_REGION \
  --overwrite

aws ssm put-parameter \
  --name /apipark/mysql_port \
  --value "3306" \
  --type String \
  --region $AWS_REGION \
  --overwrite

# Redis configuration
aws ssm put-parameter \
  --name /apipark/redis_addr \
  --value "$REDIS_ENDPOINT" \
  --type String \
  --region $AWS_REGION \
  --overwrite

aws ssm put-parameter \
  --name /apipark/redis_password \
  --value "" \
  --type String \
  --region $AWS_REGION \
  --overwrite

# InfluxDB token (generate a random token)
INFLUX_TOKEN=$(openssl rand -base64 32)
aws ssm put-parameter \
  --name /apipark/influx_token \
  --value "$INFLUX_TOKEN" \
  --type SecureString \
  --region $AWS_REGION \
  --overwrite

# APIPark admin credentials (change after first login!)
aws ssm put-parameter \
  --name /apipark/apipark_admin_user \
  --value "admin" \
  --type String \
  --region $AWS_REGION \
  --overwrite

aws ssm put-parameter \
  --name /apipark/apipark_admin_password \
  --value "ChangeMe!" \
  --type String \
  --region $AWS_REGION \
  --overwrite

# Verify all parameters were created
echo "Verifying parameters..."
aws ssm get-parameters-by-path \
  --path /apipark \
  --region $AWS_REGION \
  --query 'Parameters[].Name' \
  --output table

echo "✓ All secrets stored successfully"
```

### STEP 8: Install External Secrets Operator

This syncs secrets from AWS to Kubernetes:

```bash
# Add Helm repository
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

# Create namespace
kubectl create namespace external-secrets --dry-run=client -o yaml | kubectl apply -f -

# Install External Secrets
echo "Installing External Secrets Operator..."
helm install external-secrets \
  external-secrets/external-secrets \
  --namespace external-secrets \
  --set installCRDs=true

# Wait for it to be ready
kubectl wait --for=condition=available \
  --timeout=300s \
  deployment/external-secrets \
  -n external-secrets

echo "✓ External Secrets Operator installed"
```

### STEP 9: Create Pod Identity Associations

Link the IAM role to your Kubernetes service accounts:

```bash
# Create apipark namespace
kubectl create namespace apipark --dry-run=client -o yaml | kubectl apply -f -

# Create Pod Identity for apipark service account
echo "Creating Pod Identity associations..."
aws eks create-pod-identity-association \
  --cluster-name $CLUSTER_NAME \
  --namespace apipark \
  --service-account apipark-sa \
  --role-arn arn:aws:iam::$ACCOUNT_ID:role/apipark-pod-identity-role \
  --region $AWS_REGION

# Create Pod Identity for external-secrets service account
aws eks create-pod-identity-association \
  --cluster-name $CLUSTER_NAME \
  --namespace external-secrets \
  --service-account external-secrets \
  --role-arn arn:aws:iam::$ACCOUNT_ID:role/apipark-pod-identity-role \
  --region $AWS_REGION

# Restart External Secrets to pick up the identity
kubectl rollout restart deployment external-secrets -n external-secrets

# Wait for restart
sleep 20

echo "✓ Pod Identity associations created"
```

### STEP 10: Prepare TLS Certificates

You need TLS certificates for HTTPS. Create Kubernetes secrets:

```bash
# Replace the paths below with your actual certificate files
# Example if you have certificates:

kubectl create secret tls aig-secret-tls \
  --cert=/path/to/aig-cert.crt \
  --key=/path/to/aig-key.key \
  --namespace apipark

kubectl create secret tls apinto-tls \
  --cert=/path/to/apinto-cert.crt \
  --key=/path/to/apinto-key.key \
  --namespace apipark

# If you don't have certificates yet, you can:
# 1. Use AWS Certificate Manager (ACM) with ALB ingress
# 2. Use Let's Encrypt with cert-manager
# 3. Get certificates from your organization

echo "⚠️  Make sure TLS secrets are created before proceeding!"
```

### STEP 11: Update values.yaml Configuration

Navigate to your helm chart directory and verify/update configuration:

```bash
# Navigate to helm chart directory
cd /Users/sagarkolli/Downloads/helm-chart

# The values.yaml should already have correct configuration from earlier setup
# But verify these key values match your environment:

cat values.yaml | grep -A 2 "DBHOST:\|REDISADDR:"

# Should show:
# DBHOST: your-mysql-endpoint
# REDISADDR: your-redis-endpoint

# If you need to edit, use your preferred editor:
# nano values.yaml
# or
# vi values.yaml
```

Key sections to verify in `values.yaml`:

```yaml
# MySQL - using external RDS
mysql:
  enabled: false
DBHOST: apipark.c0bysoi2yx9s.us-east-1.rds.amazonaws.com

# Redis - using external ElastiCache
redis:
  enabled: false
REDISADDR: apipark-redis.cm24bo.ng.0001.use1.cache.amazonaws.com:6379

# Service Account - MUST be enabled
serviceAccount:
  create: true
  name: apipark-sa

# External Secrets - MUST be enabled
externalSecrets:
  enabled: true
  region: us-east-1
  ssmPrefix: /apipark

# Ingress - update with your domains
ingress:
  enabled: true
  hosts:
    apipark:
      host: aig.aigfintinc.com
    apinto:
      host: apinto.aigfintinc.com
    grafana:
      host: grafana.aigfintinc.com
  tls:
  - hosts:
    - aig.aigfintinc.com
    secretName: aig-secret-tls
  - hosts:
    - apinto.aigfintinc.com
    secretName: apinto-tls
```

### STEP 12: Deploy APIPark

Now deploy everything!

```bash
# Make sure you're in the helm chart directory
cd /Users/sagarkolli/Downloads/helm-chart

# Validate the Helm chart first
echo "Validating Helm chart..."
helm lint .

# Do a dry-run to see what will be created (optional but recommended)
echo "Running dry-run..."
helm install apipark . \
  --namespace apipark \
  --dry-run \
  --debug > /tmp/apipark-dry-run.yaml

echo "✓ Dry-run successful. Review /tmp/apipark-dry-run.yaml if needed"

# Actually install it
echo "Installing APIPark..."
helm install apipark . \
  --namespace apipark \
  --timeout 10m

echo "✓ APIPark installation started!"
```

### STEP 13: Monitor the Deployment

Watch the pods come up (this takes 5-10 minutes):

```bash
# Watch pods starting in real-time (press Ctrl+C to exit)
kubectl get pods -n apipark -w

# Or check status periodically:
kubectl get pods -n apipark

# Expected final state (all should show 1/1 Running):
# NAME                               READY   STATUS    RESTARTS   AGE
# apipark-apipark-xxx                1/1     Running   0          5m
# apipark-apinto-xxx                 1/1     Running   0          5m
# apipark-grafana-xxx                1/1     Running   0          5m
# apipark-influxdb-0                 1/1     Running   0          5m
# apipark-loki-xxx                   1/1     Running   0          5m
# apipark-nsq-xxx                    1/1     Running   0          5m
```

### STEP 14: Verify Everything is Healthy

```bash
# Check all resources
kubectl get all -n apipark

# Check External Secrets are syncing properly
kubectl get secretstore -n apipark
# Should show: STATUS=Valid, READY=True

kubectl get externalsecret -n apipark
# Should show: STATUS=SecretSynced, READY=True

# Check the synced secret was created
kubectl get secret apipark-app-secrets -n apipark
# Should show the secret exists with DATA=10

# Check ingress
kubectl get ingress -n apipark

# Check deployments
kubectl get deployments -n apipark
# All should show READY=1/1

echo "✓ All components deployed successfully!"
```

### STEP 15: Set Up DNS

Point your domains to the load balancer:

```bash
# Get the load balancer hostname
LB_HOSTNAME=$(kubectl get ingress apipark-apipark-ingress -n apipark \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo "================================================"
echo "Load Balancer Hostname: $LB_HOSTNAME"
echo "================================================"
echo ""
echo "Create these DNS CNAME records in your DNS provider:"
echo ""
echo "  aig.aigfintinc.com       -> $LB_HOSTNAME"
echo "  apinto.aigfintinc.com    -> $LB_HOSTNAME"
echo "  grafana.aigfintinc.com   -> $LB_HOSTNAME"
echo ""
echo "Wait 5-10 minutes for DNS to propagate, then proceed to testing."
```

### STEP 16: Test Your Deployment

After DNS propagates, test the endpoints:

```bash
# Test APIPark (main application)
curl -I https://aig.aigfintinc.com

# Test Apinto (API Gateway)
curl -I https://apinto.aigfintinc.com

# Test Grafana (monitoring)
curl -I https://grafana.aigfintinc.com

# All should return HTTP 200 or show the service is responding
```

### STEP 17: First Login

Access APIPark in your browser:

```bash
echo "================================================"
echo "🎉 APIPark is now accessible at:"
echo ""
echo "  Main App:  https://aig.aigfintinc.com"
echo "  Gateway:   https://apinto.aigfintinc.com"
echo "  Grafana:   https://grafana.aigfintinc.com"
echo ""
echo "Default credentials:"
echo "  Username: admin"
echo "  Password: ChangeMe!"
echo ""
echo "⚠️  IMPORTANT: Change the password after first login!"
echo "================================================"
```

---

## Common Issues and Solutions

### Issue: Pods stuck in "Pending" state

**Symptom**: Pods show "Pending" status for more than 5 minutes

**Solution**: Check if nodes have enough resources or if persistent volumes can be created

```bash
# Check pod details
kubectl describe pod <pod-name> -n apipark

# Look for messages like:
# - "Insufficient cpu" or "Insufficient memory" -> Need larger nodes
# - "FailedScheduling" -> Check node selectors/taints

# Check node resources
kubectl top nodes

# Check if storage class exists for InfluxDB
kubectl get storageclass
# Should show 'gp2' or 'gp3'
```

### Issue: "CreateContainerConfigError" status

**Symptom**: Pods show this error in status

**Solution**: The secret isn't ready yet - External Secrets may still be syncing

```bash
# Check External Secrets status
kubectl get externalsecret -n apipark
# Should show STATUS=SecretSynced, READY=True

# If not synced, check details:
kubectl describe externalsecret apipark-secrets -n apipark

# Check if parameters exist in SSM
aws ssm get-parameters-by-path \
  --path /apipark \
  --region $AWS_REGION \
  --query 'Parameters[].Name' \
  --output table

# Check External Secrets logs
kubectl logs -n external-secrets deployment/external-secrets --tail=50
```

### Issue: "ImagePullBackOff" error

**Symptom**: Cannot pull container image

**Solution**: Check image registry access

```bash
# Check the exact error
kubectl describe pod <pod-name> -n apipark | grep -A 5 "Events:"

# If using private ECR, verify:
# 1. The image URL is correct in values.yaml
# 2. Your nodes have ECR pull permissions (via instance profile)
# 3. The image exists in ECR

# Check current image
kubectl get deployment apipark-apipark -n apipark \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### Issue: Pods crash with "invalid password" error

**Symptom**: Pods start but crash, logs show password errors

**Solution**: Password mismatch between SSM and application expectations

```bash
# Check pod logs
kubectl logs -n apipark deployment/apipark-apipark --tail=50

# If you see "invalid password", update SSM parameter
aws ssm put-parameter \
  --name /apipark/apipark_admin_password \
  --value "ChangeMe!" \
  --type String \
  --region $AWS_REGION \
  --overwrite

# Wait 60 seconds for External Secrets to refresh
sleep 60

# Restart the deployment
kubectl rollout restart deployment apipark-apipark -n apipark

# Monitor restart
kubectl rollout status deployment apipark-apipark -n apipark
```

### Issue: Cannot access via HTTPS

**Symptom**: curl fails or browser shows certificate errors

**Solutions**:

```bash
# 1. Check TLS secret exists
kubectl get secret aig-secret-tls -n apipark
kubectl get secret apinto-tls -n apipark

# 2. Check DNS is resolving correctly
nslookup aig.aigfintinc.com

# 3. Check ingress configuration
kubectl describe ingress apipark-apipark-ingress -n apipark

# 4. Check NGINX ingress controller logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller --tail=100

# 5. Verify the load balancer is working
kubectl get service -n ingress-nginx ingress-nginx-controller
```

### Issue: External Secrets not syncing

**Symptom**: ExternalSecret shows "SecretSyncedError"

**Solution**: Check Pod Identity and IAM permissions

```bash
# Check Pod Identity associations exist
aws eks list-pod-identity-associations \
  --cluster-name $CLUSTER_NAME \
  --region $AWS_REGION

# Should show associations for:
# - namespace: apipark, serviceAccount: apipark-sa
# - namespace: external-secrets, serviceAccount: external-secrets

# Check External Secrets controller logs
kubectl logs -n external-secrets deployment/external-secrets --tail=100

# Look for errors like:
# - "timeout" -> Pod Identity Agent may not be installed
# - "AccessDenied" -> IAM permissions issue
# - "parameter not found" -> SSM parameter doesn't exist

# Verify IAM role permissions
aws iam list-attached-role-policies \
  --role-name apipark-pod-identity-role

# Test SSM access manually
aws ssm get-parameter \
  --name /apipark/mysql_username \
  --region $AWS_REGION
```

### Issue: Database connection errors

**Symptom**: Pods crash with MySQL connection errors

**Solution**: Check database connectivity and credentials

```bash
# Test MySQL connection from within the cluster
kubectl run mysql-test --rm -it --image=mysql:8.0 -n apipark -- \
  mysql -h $MYSQL_ENDPOINT -u $MYSQL_USERNAME -p$MYSQL_PASSWORD -e "SHOW DATABASES;"

# Check security groups allow EKS -> RDS traffic
# - RDS security group should allow 3306 from EKS node security group

# Verify SSM has correct MySQL credentials
aws ssm get-parameter \
  --name /apipark/mysql_host \
  --region $AWS_REGION \
  --query 'Parameter.Value' \
  --output text

aws ssm get-parameter \
  --name /apipark/mysql_username \
  --region $AWS_REGION \
  --query 'Parameter.Value' \
  --output text
```

### Issue: InfluxDB pod won't start - storage issue

**Symptom**: InfluxDB StatefulSet pod shows volume mount errors

**Solution**: Check storage class and PVC

```bash
# Check if PVC was created
kubectl get pvc -n apipark

# Check PVC details
kubectl describe pvc data-apipark-influxdb-0 -n apipark

# Check storage class exists
kubectl get storageclass

# If using gp2, make sure it's set as default or specified in values.yaml
# In values.yaml, InfluxDB should have:
# storageClassName: "gp2"
```

---

## Upgrading APIPark

To upgrade after making changes to values.yaml or when a new version is available:

```bash
# Navigate to helm chart directory
cd /Users/sagarkolli/Downloads/helm-chart

# Pull latest changes (if using git)
# git pull

# Update values.yaml with your changes

# Upgrade the release
helm upgrade apipark . \
  --namespace apipark \
  --timeout 10m

# Monitor the rollout
kubectl rollout status deployment/apipark-apipark -n apipark

# Check all pods are running
kubectl get pods -n apipark
```

---

## Uninstalling Everything

To completely remove the deployment:

```bash
# Uninstall Helm release
helm uninstall apipark --namespace apipark

# Delete namespace
kubectl delete namespace apipark

# Delete External Secrets
helm uninstall external-secrets --namespace external-secrets
kubectl delete namespace external-secrets

# Delete NGINX Ingress
helm uninstall ingress-nginx --namespace ingress-nginx
kubectl delete namespace ingress-nginx

# Get and delete Pod Identity associations
ASSOCIATION_IDS=$(aws eks list-pod-identity-associations \
  --cluster-name $CLUSTER_NAME \
  --region $AWS_REGION \
  --query 'associations[].associationId' \
  --output text)

for ASSOC_ID in $ASSOCIATION_IDS; do
  aws eks delete-pod-identity-association \
    --cluster-name $CLUSTER_NAME \
    --association-id $ASSOC_ID \
    --region $AWS_REGION
done

# Remove IAM resources
aws iam detach-role-policy \
  --role-name apipark-pod-identity-role \
  --policy-arn arn:aws:iam::$ACCOUNT_ID:policy/apipark-ssm-access

aws iam delete-policy \
  --policy-arn arn:aws:iam::$ACCOUNT_ID:policy/apipark-ssm-access

aws iam delete-role --role-name apipark-pod-identity-role

# Delete SSM parameters (optional - removes your secrets!)
aws ssm get-parameters-by-path \
  --path /apipark \
  --region $AWS_REGION \
  --query 'Parameters[].Name' \
  --output text | tr '\t' '\n' | while read param; do
  aws ssm delete-parameter --name "$param" --region $AWS_REGION
done

echo "✓ Complete uninstall finished"
```

---

## Quick Reference Commands

```bash
# Check everything is running
kubectl get all -n apipark

# View logs from main application
kubectl logs -n apipark deployment/apipark-apipark --tail=100 -f

# Restart a deployment
kubectl rollout restart deployment/apipark-apipark -n apipark

# Check secret sync status
kubectl get externalsecret -n apipark
kubectl get secretstore -n apipark

# Check ingress status
kubectl get ingress -n apipark

# Get load balancer URL
kubectl get ingress apipark-apipark-ingress -n apipark \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'

# View all secrets (names only, not values)
kubectl get secrets -n apipark

# Check Pod Identity associations
aws eks list-pod-identity-associations \
  --cluster-name $CLUSTER_NAME \
  --region $AWS_REGION

# Check SSM parameters
aws ssm get-parameters-by-path \
  --path /apipark \
  --region $AWS_REGION \
  --query 'Parameters[].Name'

# Check events for troubleshooting
kubectl get events -n apipark --sort-by='.lastTimestamp'

# Exec into a pod for debugging
kubectl exec -it <pod-name> -n apipark -- /bin/sh
```

---

## What Each Component Does

- **apipark-apipark**: Main APIPark application - manages your APIs
- **apipark-apinto**: API Gateway - handles API traffic routing
- **apipark-grafana**: Monitoring dashboard - view metrics and logs
- **apipark-loki**: Log aggregation - collects logs from all services
- **apipark-influxdb**: Time-series database - stores metrics data
- **apipark-nsq**: Message queue - handles async messaging between services

### External Dependencies:
- **MySQL (RDS)**: Main application database (external)
- **Redis (ElastiCache)**: Caching layer (external)
- **External Secrets Operator**: Syncs secrets from AWS SSM to Kubernetes
- **EKS Pod Identity Agent**: Provides AWS credentials to pods
- **NGINX Ingress Controller**: Routes external HTTPS traffic to services

---

## Architecture Overview

```
Internet
   ↓
Route 53 (DNS)
   ↓
Application Load Balancer (from Ingress)
   ↓
NGINX Ingress Controller
   ↓
┌─────────────────────────────────────────┐
│  APIPark Services (in EKS)              │
│  ├─ apipark-apipark (main app)          │
│  ├─ apipark-apinto (gateway)            │
│  ├─ apipark-grafana (monitoring)        │
│  ├─ apipark-loki (logs)                 │
│  ├─ apipark-influxdb (metrics)          │
│  └─ apipark-nsq (queue)                 │
└─────────────────────────────────────────┘
         ↓                    ↓
    RDS MySQL          ElastiCache Redis
    (external)            (external)
         ↓
   AWS Systems Manager
   Parameter Store
   (secrets storage)
```

---

## Security Best Practices

### After Installation Checklist:

- [ ] Change default admin password immediately after first login
- [ ] Review and restrict IAM policy to minimum required permissions
- [ ] Enable encryption at rest for RDS and ElastiCache
- [ ] Use SecureString type for all sensitive SSM parameters
- [ ] Rotate database credentials regularly
- [ ] Enable audit logging on RDS
- [ ] Review and tighten security groups (RDS, ElastiCache, EKS nodes)
- [ ] Set up monitoring alerts in CloudWatch
- [ ] Enable pod security standards in EKS
- [ ] Use separate namespaces for dev/staging/prod environments
- [ ] Implement network policies to restrict pod-to-pod communication
- [ ] Regularly update EKS cluster and add-ons to latest versions

### Recommended IAM Policy Restrictions:

The IAM policy we created allows access to all parameters under `/apipark/*`. For production, consider:
- Using different paths for dev/staging/prod (e.g., `/apipark/prod/*`)
- Restricting by environment in the IAM policy
- Using separate IAM roles for different namespaces

---

## Required AWS Resources Summary

### What This Guide Installs:
✅ EKS Pod Identity Agent add-on
✅ NGINX Ingress Controller
✅ External Secrets Operator
✅ IAM Role and Policy for Pod Identity
✅ APIPark and all its components

### What You Must Provide:
❗ EKS Cluster (existing)
❗ RDS MySQL Database (external)
❗ ElastiCache Redis (external, or enable internal redis)
❗ TLS Certificates
❗ DNS configuration
❗ Storage class (gp2 or gp3)

### Optional:
⚪ EBS CSI Driver (only if gp2 doesn't work)

---

## Getting Help

If you run into issues:

1. **Check pod status**: `kubectl get pods -n apipark`
2. **Check pod logs**: `kubectl logs <pod-name> -n apipark --tail=100`
3. **Check pod events**: `kubectl describe pod <pod-name> -n apipark`
4. **Check External Secrets**: `kubectl describe externalsecret apipark-secrets -n apipark`
5. **Check ingress**: `kubectl describe ingress apipark-apipark-ingress -n apipark`
6. **Check all events**: `kubectl get events -n apipark --sort-by='.lastTimestamp'`
7. **Review this README** troubleshooting section

---

**You're all set! 🎉**

Your APIPark installation should now be running and accessible via your configured domain names.

For questions or issues not covered here, check the pod logs and events - they usually contain helpful error messages that point to the problem.
