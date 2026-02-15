# Kubernetes Deployment Guide for Apinto Gateway

## Prerequisites

### 1. Determine Your Node Architecture

**Quick check:**
```bash
# Method 1: Check node architecture
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.nodeInfo.architecture}{"\n"}{end}'

# Method 2: SSH to a node and check
ssh your-node
uname -m
```

**Results:**
- `amd64` or `x86_64` → Use **amd64** image (most common)
- `arm64` or `aarch64` → Use **arm64** image

### 2. Configure ECR Authentication

Kubernetes needs credentials to pull images from your private ECR repository.

#### Option A: Using AWS IAM Roles (EKS Recommended)

**For AWS EKS clusters:**
```bash
# 1. Create IAM policy for ECR access
cat > ecr-policy.json <<EOF
{
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
}
EOF

# 2. Create policy
aws iam create-policy \
    --policy-name ApintoECRReadOnly \
    --policy-document file://ecr-policy.json

# 3. Attach to node IAM role
aws iam attach-role-policy \
    --role-name <your-eks-node-role> \
    --policy-arn arn:aws:iam::<account-id>:policy/ApintoECRReadOnly
```

#### Option B: Using Image Pull Secrets (Any Kubernetes)

```bash
# 1. Get ECR login token
ECR_TOKEN=$(aws ecr get-login-password --region us-east-1)

# 2. Create Docker config
kubectl create secret docker-registry ecr-secret \
    --docker-server=865783518572.dkr.ecr.us-east-1.amazonaws.com \
    --docker-username=AWS \
    --docker-password="${ECR_TOKEN}" \
    --namespace=apinto-gateway

# Note: ECR tokens expire after 12 hours!
# For production, use a cronjob to refresh the secret
```

#### Option C: Using ECR Credential Helper (Recommended for Non-EKS)

Install the ECR credential helper on all nodes:
```bash
# On each node
curl -LO https://amazon-ecr-credential-helper-releases.s3.us-east-2.amazonaws.com/0.7.1/linux-amd64/docker-credential-ecr-login
chmod +x docker-credential-ecr-login
sudo mv docker-credential-ecr-login /usr/local/bin/

# Configure Docker
cat > /etc/docker/daemon.json <<EOF
{
  "credHelpers": {
    "865783518572.dkr.ecr.us-east-1.amazonaws.com": "ecr-login"
  }
}
EOF

sudo systemctl restart docker
```

## Deployment Steps

### Step 1: Update Image Architecture

Edit `deployment.yaml` and choose the correct image:

**For amd64 (most common):**
```yaml
containers:
- name: apinto
  image: 865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.1.0-amd64
  nodeSelector:
    kubernetes.io/arch: amd64
```

**For arm64:**
```yaml
containers:
- name: apinto
  image: 865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.1.0-arm64
  nodeSelector:
    kubernetes.io/arch: arm64
```

### Step 2: Add Image Pull Secret (if using Option B)

Add to `deployment.yaml` under `spec.template.spec`:
```yaml
spec:
  template:
    spec:
      imagePullSecrets:
      - name: ecr-secret
```

### Step 3: Deploy to Kubernetes

```bash
# Apply the deployment
kubectl apply -f k8s/deployment.yaml

# Verify deployment
kubectl get pods -n apinto-gateway

# Check pod status
kubectl describe pod -n apinto-gateway <pod-name>

# View logs
kubectl logs -n apinto-gateway -l app=apinto-gateway --tail=50 -f
```

### Step 4: Verify Service

```bash
# Check service
kubectl get svc -n apinto-gateway

# Test from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
    curl http://apinto-gateway.apinto-gateway.svc.cluster.local:8080

# Port forward for local testing
kubectl port-forward -n apinto-gateway svc/apinto-gateway 8080:8080

# Then test locally
curl http://localhost:8080
```

## Deployment Options

### 1. ClusterIP (Internal Only)
Default in `deployment.yaml` - only accessible within cluster.

### 2. LoadBalancer (External Access)
Uncomment the `apinto-gateway-external` service:
```bash
# After deployment
kubectl get svc -n apinto-gateway apinto-gateway-external

# Get external IP
EXTERNAL_IP=$(kubectl get svc -n apinto-gateway apinto-gateway-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl http://${EXTERNAL_IP}
```

### 3. Ingress (Domain-based Routing)
Update the Ingress section in `deployment.yaml`:
```yaml
spec:
  ingressClassName: nginx  # or your ingress controller
  rules:
  - host: api.yourdomain.com  # Your domain
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: apinto-gateway
            port:
              number: 8080
```

Then apply:
```bash
kubectl apply -f k8s/deployment.yaml

# Check ingress
kubectl get ingress -n apinto-gateway
```

## Scaling

### Manual Scaling
```bash
# Scale to 5 replicas
kubectl scale deployment apinto-gateway -n apinto-gateway --replicas=5

# Check status
kubectl get pods -n apinto-gateway
```

### Auto Scaling
The HPA is included in `deployment.yaml`. Monitor it:
```bash
# Check HPA status
kubectl get hpa -n apinto-gateway

# Describe HPA
kubectl describe hpa apinto-gateway -n apinto-gateway
```

## Updating Image Version

When a new version is released:

```bash
# Method 1: Update deployment.yaml and apply
# Edit image tag to new version
kubectl apply -f k8s/deployment.yaml

# Method 2: Direct update
kubectl set image deployment/apinto-gateway \
    apinto=865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.2.0-amd64 \
    -n apinto-gateway

# Watch rollout
kubectl rollout status deployment/apinto-gateway -n apinto-gateway

# Rollback if needed
kubectl rollout undo deployment/apinto-gateway -n apinto-gateway
```

## Troubleshooting

### Issue: Image Pull Error

**Symptom:**
```
Failed to pull image: Error response from daemon: pull access denied
```

**Solution:**
```bash
# Check image pull secret
kubectl get secret ecr-secret -n apinto-gateway

# Verify secret is in pod spec
kubectl get deployment apinto-gateway -n apinto-gateway -o yaml | grep imagePullSecrets

# Recreate secret with fresh token
kubectl delete secret ecr-secret -n apinto-gateway
ECR_TOKEN=$(aws ecr get-login-password --region us-east-1)
kubectl create secret docker-registry ecr-secret \
    --docker-server=865783518572.dkr.ecr.us-east-1.amazonaws.com \
    --docker-username=AWS \
    --docker-password="${ECR_TOKEN}" \
    --namespace=apinto-gateway
```

### Issue: Wrong Architecture

**Symptom:**
```
exec format error
```

**Solution:**
Check node architecture and use correct image:
```bash
# Check what architecture your nodes use
kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.architecture}'

# Update deployment with correct image
# amd64 or arm64
```

### Issue: Pod CrashLoopBackOff

**Check logs:**
```bash
kubectl logs -n apinto-gateway <pod-name>
kubectl describe pod -n apinto-gateway <pod-name>
```

**Common causes:**
1. Missing configuration
2. Port conflicts
3. Resource limits too low

### Issue: Service Not Accessible

**Check:**
```bash
# Verify pods are running
kubectl get pods -n apinto-gateway

# Check service endpoints
kubectl get endpoints -n apinto-gateway apinto-gateway

# Test from another pod
kubectl run -it --rm debug --image=busybox --restart=Never -- \
    wget -O- http://apinto-gateway.apinto-gateway.svc.cluster.local:8080
```

## Monitoring

### View Logs
```bash
# All pods
kubectl logs -n apinto-gateway -l app=apinto-gateway --tail=100 -f

# Specific pod
kubectl logs -n apinto-gateway <pod-name> -f

# Previous instance (if crashed)
kubectl logs -n apinto-gateway <pod-name> --previous
```

### Check Resources
```bash
# Resource usage
kubectl top pods -n apinto-gateway

# Describe deployment
kubectl describe deployment apinto-gateway -n apinto-gateway

# Events
kubectl get events -n apinto-gateway --sort-by='.lastTimestamp'
```

## Production Checklist

- [ ] Architecture verified (amd64 or arm64)
- [ ] ECR authentication configured
- [ ] Image pull secret created (if not using IAM roles)
- [ ] Resource limits set appropriately
- [ ] Health checks configured
- [ ] Persistent volumes configured (if needed)
- [ ] Ingress/LoadBalancer configured
- [ ] TLS certificates configured (for HTTPS)
- [ ] HPA configured for auto-scaling
- [ ] Monitoring and logging set up
- [ ] Backup strategy defined

## Architecture-Specific Notes

### For amd64 Clusters (Intel/AMD)
- Most common, broad compatibility
- Use: `0.1.0-amd64` images
- No special configuration needed

### For arm64 Clusters (AWS Graviton, Raspberry Pi)
- Better price/performance on AWS Graviton
- Use: `0.1.0-arm64` images
- 20-40% cost savings on AWS

### For Mixed Architecture Clusters
Use node affinity to schedule pods on appropriate nodes:
```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/arch
          operator: In
          values:
          - amd64  # or arm64
```

## Quick Reference

```bash
# Check node architecture
kubectl get nodes -o jsonpath='{.items[*].status.nodeInfo.architecture}'

# Deploy
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get all -n apinto-gateway

# View logs
kubectl logs -n apinto-gateway -l app=apinto-gateway -f

# Scale
kubectl scale deployment apinto-gateway -n apinto-gateway --replicas=5

# Update image
kubectl set image deployment/apinto-gateway apinto=IMAGE:TAG -n apinto-gateway

# Port forward
kubectl port-forward -n apinto-gateway svc/apinto-gateway 8080:8080

# Delete
kubectl delete -f k8s/deployment.yaml
```

## Support

For issues:
- Check logs: `kubectl logs -n apinto-gateway <pod-name>`
- Check events: `kubectl get events -n apinto-gateway`
- Describe pod: `kubectl describe pod -n apinto-gateway <pod-name>`
- GitHub: https://github.com/sagarkolli249/apinto/issues
