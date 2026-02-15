# Kubernetes Deployment Quick Start

## For Linux Machines: Use amd64 (Standard)

**Bottom Line:** 99% of Linux machines in Kubernetes use **amd64** architecture.

### Quick Answer

**Use this image:**
```
865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.1.0-amd64
```

**Standard architecture:** amd64 (Intel/AMD processors)

## Step-by-Step Deployment

### 1. Check Your Architecture (Verify First)

```bash
# Quick check
kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.architecture}'

# Expected output: amd64 (or x86_64)
```

### 2. Setup ECR Authentication

**Automated Setup:**
```bash
cd k8s
./setup-ecr-auth.sh
```

This script will:
- Detect your cluster type (EKS, GKE, or other)
- Create the namespace
- Set up ECR authentication
- Test image pulling
- Tell you exactly which image to use

**Manual Setup (Alternative):**
```bash
# Create namespace
kubectl create namespace apinto-gateway

# Create ECR secret
ECR_TOKEN=$(aws ecr get-login-password --region us-east-1)
kubectl create secret docker-registry ecr-secret \
    --docker-server=865783518572.dkr.ecr.us-east-1.amazonaws.com \
    --docker-username=AWS \
    --docker-password="${ECR_TOKEN}" \
    --namespace=apinto-gateway
```

### 3. Deploy Apinto Gateway

**Quick Deploy (Standard amd64):**
```bash
# Deploy everything
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get pods -n apinto-gateway

# Expected output:
# NAME                              READY   STATUS    RESTARTS   AGE
# apinto-gateway-xxxxxxxxxx-xxxxx   1/1     Running   0          30s
# apinto-gateway-xxxxxxxxxx-xxxxx   1/1     Running   0          30s
# apinto-gateway-xxxxxxxxxx-xxxxx   1/1     Running   0          30s
```

### 4. Verify Deployment

```bash
# Check all resources
kubectl get all -n apinto-gateway

# View logs
kubectl logs -n apinto-gateway -l app=apinto-gateway --tail=50

# Port forward for testing
kubectl port-forward -n apinto-gateway svc/apinto-gateway 8080:8080

# Test (in another terminal)
curl http://localhost:8080
```

## Architecture Reference

### Standard Configuration (amd64)

**For Linux machines with Intel/AMD processors:**
- EC2 instances: t3, m5, c5, r5 series
- GCP: n1, n2, e2 standard instances
- Azure: Standard VMs
- Most on-premise servers
- 99% of Kubernetes clusters

**Image to use:**
```yaml
image: 865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.1.0-amd64
nodeSelector:
  kubernetes.io/arch: amd64
```

### Alternative: ARM Architecture (arm64)

**Only if you have:**
- AWS Graviton instances (t4g, m6g, c6g)
- Raspberry Pi clusters
- ARM-based servers

**Image to use:**
```yaml
image: 865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.1.0-arm64
nodeSelector:
  kubernetes.io/arch: arm64
```

## Deployment Manifest Overview

The `k8s/deployment.yaml` includes:

1. **Namespace:** `apinto-gateway`
2. **Deployment:** 3 replicas, amd64 configured by default
3. **Service (ClusterIP):** Internal access on port 8080
4. **Service (LoadBalancer):** Optional external access
5. **Ingress:** Optional domain-based routing
6. **HPA:** Auto-scaling based on CPU/memory
7. **ConfigMap:** Configuration management

## Common Scenarios

### Scenario 1: Internal Gateway (ClusterIP)

Default configuration - gateway accessible only within cluster.

**No changes needed!** Just deploy:
```bash
kubectl apply -f k8s/deployment.yaml
```

### Scenario 2: External Access (LoadBalancer)

Exposes gateway with external IP.

**Already included!** Check external IP:
```bash
kubectl get svc -n apinto-gateway apinto-gateway-external
```

### Scenario 3: Domain-based Routing (Ingress)

Access via domain name like `api.yourdomain.com`.

**Edit `k8s/deployment.yaml`:**
```yaml
spec:
  rules:
  - host: api.yourdomain.com  # Change this
```

Then apply:
```bash
kubectl apply -f k8s/deployment.yaml
```

## Troubleshooting

### Problem: ImagePullBackOff

**Check:**
```bash
kubectl describe pod -n apinto-gateway <pod-name>
```

**Common causes:**
1. ECR secret not created → Run `./setup-ecr-auth.sh`
2. ECR secret expired (12 hours) → Recreate secret
3. Wrong image tag → Verify image exists in ECR

**Fix:**
```bash
# Recreate ECR secret
kubectl delete secret ecr-secret -n apinto-gateway
ECR_TOKEN=$(aws ecr get-login-password --region us-east-1)
kubectl create secret docker-registry ecr-secret \
    --docker-server=865783518572.dkr.ecr.us-east-1.amazonaws.com \
    --docker-username=AWS \
    --docker-password="${ECR_TOKEN}" \
    --namespace=apinto-gateway
```

### Problem: exec format error

**Cause:** Wrong architecture image

**Check node architecture:**
```bash
kubectl get nodes -o jsonpath='{.items[*].status.nodeInfo.architecture}'
```

**Fix:** Use correct image in `deployment.yaml`
- amd64 nodes → use `0.1.0-amd64` image
- arm64 nodes → use `0.1.0-arm64` image

### Problem: Pods not ready

**Check logs:**
```bash
kubectl logs -n apinto-gateway <pod-name>
```

**Common issues:**
1. Configuration missing
2. Port conflicts
3. Health check failing

## Updating to New Versions

When new version is released (e.g., 0.2.0):

```bash
# Method 1: Edit deployment.yaml
# Change: 0.1.0-amd64 → 0.2.0-amd64
kubectl apply -f k8s/deployment.yaml

# Method 2: Direct update
kubectl set image deployment/apinto-gateway \
    apinto=865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.2.0-amd64 \
    -n apinto-gateway

# Watch rollout
kubectl rollout status deployment/apinto-gateway -n apinto-gateway
```

## Production Recommendations

### 1. ECR Authentication

**For EKS:**
- Use IAM roles for service accounts (IRSA)
- No secrets needed, tokens don't expire

**For other Kubernetes:**
- Use CronJob to refresh ECR secret every 6 hours
- Or use ECR credential helper on nodes

### 2. Resource Limits

Adjust based on your traffic:
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 3. Scaling

Configure HPA thresholds:
```yaml
minReplicas: 3
maxReplicas: 10
targetCPUUtilizationPercentage: 70
```

### 4. Monitoring

Add monitoring labels:
```yaml
metadata:
  labels:
    app: apinto-gateway
    version: v0.1.0
    environment: production
```

## Quick Commands Reference

```bash
# Deploy
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get all -n apinto-gateway

# View logs
kubectl logs -n apinto-gateway -l app=apinto-gateway -f

# Scale
kubectl scale deployment apinto-gateway --replicas=5 -n apinto-gateway

# Port forward
kubectl port-forward -n apinto-gateway svc/apinto-gateway 8080:8080

# Restart deployment
kubectl rollout restart deployment/apinto-gateway -n apinto-gateway

# Delete
kubectl delete -f k8s/deployment.yaml
```

## Summary

**Default Setup (Most Common):**
1. ✅ Linux machines = amd64 architecture
2. ✅ Use image: `0.1.0-amd64`
3. ✅ Run setup script: `./k8s/setup-ecr-auth.sh`
4. ✅ Deploy: `kubectl apply -f k8s/deployment.yaml`
5. ✅ Verify: `kubectl get pods -n apinto-gateway`

**That's it!** Your Apinto gateway will be running on Kubernetes.

---

**Need help?**
- Full guide: `k8s/README.md`
- Troubleshooting: Check pod logs
- Architecture issues: Verify with `kubectl get nodes -o wide`
