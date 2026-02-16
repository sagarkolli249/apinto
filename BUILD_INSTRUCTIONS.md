# Apinto Local Build Instructions

**Last Successful Build:** Apinto 0.1.1 (Feb 15, 2026)
**Build Machine:** Mac ARM64 (darwin/arm64)
**Target Platform:** Linux AMD64
**Go Version:** 1.25.1

---

## Prerequisites

### Required Tools
```bash
# Verify Go version
go version
# Expected: go version go1.25.1 darwin/arm64

# Verify Docker Buildx
docker buildx version

# Verify AWS CLI and ECR access
aws --version
aws ecr-public get-login-password --region us-east-1 --profile fintinc
```

### Directory Structure
```
/Users/satishsomaraju/work/NS/apipark/EXTERNAL_REPOS/apinto/
├── build/
│   ├── cmd/
│   │   ├── build.sh          ← Main build script
│   │   ├── package.sh        ← Packaging script
│   │   ├── common.sh         ← Shared functions
│   │   └── Dockerfile        ← Container build file
│   └── resources/            ← Docker build context
├── drivers/
│   └── ai-provider/
│       ├── bedrock/          ← Bedrock implementation (reference)
│       └── mistralai/        ← Mistral implementation
├── out/                      ← Build artifacts directory
└── go.mod                    ← Go dependencies
```

---

## Build Process (Step-by-Step)

### Step 1: Navigate to Apinto Repository
```bash
cd /Users/satishsomaraju/work/NS/apipark/EXTERNAL_REPOS/apinto
```

### Step 2: Verify Code Changes (Optional)
```bash
# Check what's modified
git status --short

# View Mistral changes
git diff drivers/ai-provider/mistralai/mistralai.go

# Verify compilation
go build -o /dev/null ./drivers/ai-provider/mistralai/
```

### Step 3: Build Binary for Linux AMD64
```bash
# Syntax: ./build/cmd/build.sh <VERSION> <ARCH>
./build/cmd/build.sh 0.1.1 amd64
```

**What This Does:**
- Cross-compiles Go code: `GOOS=linux GOARCH=amd64 go build`
- Creates binary at: `out/apinto-0.1.1-amd64/apinto`
- Includes all scripts: `auto-start.sh`, `config.yml.tpl`, etc.
- Embeds version: `0.1.1`

**Expected Output:**
```
Building apinto version 0.1.1 for amd64...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s ..." -o out/apinto-0.1.1-amd64/apinto
Build complete: out/apinto-0.1.1-amd64/
```

**Verify Build:**
```bash
ls -lh out/apinto-0.1.1-amd64/
# Should see: apinto binary, auto-start.sh, config.yml.tpl, etc.

file out/apinto-0.1.1-amd64/apinto
# Expected: ELF 64-bit LSB executable, x86-64, statically linked
```

### Step 4: Package Binary as Tarball
```bash
cd out

# Create proper directory structure (must be named 'apinto')
cp -r apinto-0.1.1-amd64 apinto

# Create tarball (specific name required by Dockerfile)
tar -czf apinto.linux.x64.tar.gz apinto

# Verify tarball
tar -tzf apinto.linux.x64.tar.gz | head -10
```

**Expected Tarball Contents:**
```
apinto/
apinto/config.yml.tpl
apinto/Dockerfile
apinto/install.sh
apinto/auto-start.sh
apinto/apinto.yml.tpl
apinto/version
apinto/apinto          ← The binary
...
```

### Step 5: Prepare Docker Build Context
```bash
# Copy tarball to Docker build resources directory
cp apinto.linux.x64.tar.gz ../build/resources/

# Verify
ls -lh ../build/resources/apinto.linux.x64.tar.gz
```

**Critical:** The Dockerfile expects exactly `apinto.linux.x64.tar.gz` at this path.

### Step 6: Build Docker Image (Multi-Platform)
```bash
cd ../build/resources

# Build for linux/amd64 (required for EKS)
docker buildx build \
  --platform linux/amd64 \
  -t public.ecr.aws/e5v3y2z9/apinto:0.1.1 \
  -t public.ecr.aws/e5v3y2z9/apinto:latest \
  --load \
  -f ../cmd/Dockerfile \
  .
```

**Build Parameters Explained:**
- `--platform linux/amd64` - Target architecture (must match binary)
- `-t public.ecr.aws/e5v3y2z9/apinto:0.1.1` - Version tag
- `-t public.ecr.aws/e5v3y2z9/apinto:latest` - Latest tag
- `--load` - Load image into local Docker (not push yet)
- `-f ../cmd/Dockerfile` - Path to Dockerfile
- `.` - Build context (current directory with tarball)

**Expected Output:**
```
[+] Building 45.2s (15/15) FINISHED
 => [internal] load build definition from Dockerfile
 => => transferring dockerfile: 1.23kB
 => [internal] load .dockerignore
 => [internal] load metadata for docker.io/library/alpine:latest
 => [1/7] FROM docker.io/library/alpine:latest@sha256...
 => [2/7] RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.aliyun.com/alpine|g' ...
 => [3/7] COPY ./apinto.linux.x64.tar.gz /
 => [4/7] RUN tar -zxvf apinto.linux.x64.tar.gz && rm -rf ../apinto.linux.x64.tar.gz
 => [5/7] RUN mkdir -p /etc/apinto
 => [6/7] RUN cp /apinto/apinto.yml.tpl /etc/apinto/apinto.yml ...
 => [7/7] RUN chmod 755 /apinto/*.sh
 => exporting to image
 => => exporting layers
 => => writing image sha256:78eeaa5232c628f37160b27382f7edaba67659fed2ec94df72e60dc02b848ee3
 => => naming to public.ecr.aws/e5v3y2z9/apinto:0.1.1
 => => naming to public.ecr.aws/e5v3y2z9/apinto:latest
```

**Verify Image:**
```bash
docker images public.ecr.aws/e5v3y2z9/apinto:0.1.1
# Check size, creation date

docker inspect public.ecr.aws/e5v3y2z9/apinto:0.1.1 | grep -A 3 "Architecture"
# Expected: "Architecture": "amd64"
```

### Step 7: Authenticate to AWS ECR Public
```bash
aws ecr-public get-login-password \
  --region us-east-1 \
  --profile fintinc | docker login \
  --username AWS \
  --password-stdin public.ecr.aws
```

**Expected Output:**
```
Login Succeeded
```

### Step 8: Push Image to ECR Public
```bash
# Push version tag
docker push public.ecr.aws/e5v3y2z9/apinto:0.1.1

# Push latest tag
docker push public.ecr.aws/e5v3y2z9/apinto:latest
```

**Expected Output:**
```
The push refers to repository [public.ecr.aws/e5v3y2z9/apinto]
78eeaa5232c6: Pushed
0.1.1: digest: sha256:78eeaa5232c628f37160b27382f7edaba67659fed2ec94df72e60dc02b848ee3 size: 2417
```

**Verify Push:**
```bash
aws ecr-public describe-images \
  --repository-name apinto \
  --region us-east-1 \
  --profile fintinc \
  --image-ids imageTag=0.1.1
```

---

## Deployment to Kubernetes

### Option 1: Manual Update (Fast)
```bash
# Update deployment image
kubectl set image deployment/apipark-apinto \
  apinto=public.ecr.aws/e5v3y2z9/apinto:0.1.1 \
  -n apipark

# Force pod restart (if using 'latest' tag)
kubectl rollout restart deployment/apipark-apinto -n apipark

# Watch rollout
kubectl rollout status deployment/apipark-apinto -n apipark

# IMPORTANT: Delete old pods to prevent PVC race condition
kubectl get pods -n apipark | grep apinto
kubectl delete pod -n apipark -l app=apipark-apinto,pod-template-hash=<OLD_HASH>
```

### Option 2: Helm Chart Update (Proper)
```bash
cd /Users/satishsomaraju/work/NS/apipark/EXTERNAL_REPOS/apipark

# Edit helm-chart/charts/apinto/values.yaml
# Change: tag: 0.1.1

# Deploy via Helm
helm upgrade apipark ./helm-chart \
  --namespace apipark \
  --reuse-values \
  --set apinto.tag=0.1.1

# Watch deployment
kubectl get pods -n apipark -w
```

---

## Verification

### 1. Check Deployed Image
```bash
kubectl get deployment/apipark-apinto -n apipark \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
# Expected: public.ecr.aws/e5v3y2z9/apinto:0.1.1
```

### 2. Check Pod Status
```bash
kubectl get pods -n apipark -l app=apipark-apinto
# Expected: Running 1/1
```

### 3. Check Logs for Mistral
```bash
kubectl logs -n apipark deployment/apipark-apinto --tail=50 | grep -i mistral
# Expected (when tool call is made):
# Mistral: Request converted. Model=mistral-large-latest, Messages=1, Tools=1
```

### 4. Test Mistral Endpoint
```bash
curl -X POST https://apinto.aigfintinc.com/4af6371c/mistral \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "mistral-large-latest",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

---

## Troubleshooting

### Build Fails
```bash
# Clean and retry
rm -rf out/apinto-0.1.1-amd64 out/apinto.linux.x64.tar.gz
./build/cmd/build.sh 0.1.1 amd64

# Check Go dependencies
go mod tidy
go mod verify
```

### Docker Build Fails
```bash
# Check tarball exists
ls -lh build/resources/apinto.linux.x64.tar.gz

# Rebuild with verbose output
docker buildx build --progress=plain \
  --platform linux/amd64 \
  -t public.ecr.aws/e5v3y2z9/apinto:0.1.1 \
  -f build/cmd/Dockerfile \
  build/resources/
```

### Push Fails (Auth)
```bash
# Re-authenticate
aws ecr-public get-login-password \
  --region us-east-1 \
  --profile fintinc | docker login \
  --username AWS \
  --password-stdin public.ecr.aws

# Verify credentials
aws sts get-caller-identity --profile fintinc
```

### Pod Fails to Start
```bash
# Check events
kubectl describe pod -n apipark -l app=apipark-apinto

# Check image pull
kubectl get events -n apipark --sort-by='.lastTimestamp' | grep apinto

# Force image pull
kubectl delete pod -n apipark -l app=apipark-apinto
```

### PVC Race Condition
```bash
# Delete ALL old pods before new one starts
kubectl get pods -n apipark -l app=apipark-apinto -o wide
kubectl delete pod <OLD_POD_NAME_1> <OLD_POD_NAME_2> -n apipark

# Or delete by ReplicaSet hash
kubectl delete pod -n apipark -l app=apipark-apinto,pod-template-hash=<OLD_HASH>
```

---

## Build History Reference

### Successful Build: 0.1.1 (Feb 15, 2026)
```
Code Modified:     Feb 15, 19:17 (Mistral tool calling)
Binary Built:      Feb 15, 20:27 (1h 10m after code)
Docker Created:    Feb 15, 20:54 (27m after binary)
ECR Pushed:        Feb 15, 20:59 (5m after docker)
Deployed to K8s:   Feb 15, 21:05 (6m after push)

Status:           ✅ Success
Pod Status:       Running 1/1
Image Digest:     sha256:78eeaa5232c628f37160b27382f7edaba67659fed2ec94df72e60dc02b848ee3
```

**Changes Included:**
- Mistral tool calling support (drivers/ai-provider/mistralai/mistralai.go)
- Custom RequestConvert/ResponseConvert methods
- OpenAI-compatible tool/tool_calls passthrough
- Enhanced logging for debugging

---

## Quick Reference Commands

### Full Build (0.1.2 Example)
```bash
cd /Users/satishsomaraju/work/NS/apipark/EXTERNAL_REPOS/apinto

# Build
./build/cmd/build.sh 0.1.2 amd64

# Package
cd out
cp -r apinto-0.1.2-amd64 apinto
tar -czf apinto.linux.x64.tar.gz apinto
cp apinto.linux.x64.tar.gz ../build/resources/

# Docker
cd ../build/resources
docker buildx build --platform linux/amd64 \
  -t public.ecr.aws/e5v3y2z9/apinto:0.1.2 \
  -t public.ecr.aws/e5v3y2z9/apinto:latest \
  --load -f ../cmd/Dockerfile .

# Push
aws ecr-public get-login-password --region us-east-1 --profile fintinc | \
  docker login --username AWS --password-stdin public.ecr.aws
docker push public.ecr.aws/e5v3y2z9/apinto:0.1.2
docker push public.ecr.aws/e5v3y2z9/apinto:latest

# Deploy
kubectl set image deployment/apipark-apinto \
  apinto=public.ecr.aws/e5v3y2z9/apinto:0.1.2 -n apipark
kubectl delete pod -n apipark -l app=apipark-apinto,pod-template-hash=<OLD>
```

---

## Notes

- **Always use version tags** (0.1.1, 0.1.2) for tracking
- **Always tag 'latest'** for convenience (but use versions in production)
- **Delete old pods** when updating to prevent PVC conflicts
- **Check logs immediately** after deployment for errors
- **Backup code** before making changes (`cp mistralai.go mistralai.go.backup`)
- **Test locally first** if possible (though cross-arch limits this)

---

**Last Updated:** 2026-02-16
**Maintained By:** Satish Somaraju
**Build System:** Local Mac ARM64 → Linux AMD64 cross-compile
