# ECR Repository Migration Summary

**Date:** February 15, 2026
**Status:** ✅ Complete

## Overview

Successfully migrated CI/CD pipeline to use existing ECR repositories:
- **Private ECR:** `apipark/apinto`
- **Public ECR:** `apinto`

The pipeline now builds and pushes images to both private and public ECR registries simultaneously.

---

## Changes Made

### 1. CI/CD Workflow Updates

**File:** `.github/workflows/ecr-build-push.yaml`

**Key Changes:**
- Updated environment variables to use both private and public ECR repositories
- Added login to Public ECR
- Modified build-and-push step to push to both registries
- Tagged images with both version-specific and `latest` tags
- Disabled manifest creation job (already handled by Docker Buildx)

**New Environment Variables:**
```yaml
ECR_PRIVATE_REPOSITORY: apipark/apinto
ECR_PUBLIC_REPOSITORY: apinto
PUBLIC_REGISTRY_ALIAS: e5v3y2z9
```

**Image Tags Created:**
- Private ECR:
  - `{VERSION}-amd64` (e.g., `0.1.0-amd64`)
  - `{VERSION}-arm64` (e.g., `0.1.0-arm64`)
  - `latest-amd64`
  - `latest-arm64`

- Public ECR: (same tags as private)

### 2. Kubernetes Deployment Updates

**File:** `k8s/deployment.yaml`

**Updated Image References:**
```yaml
# Private ECR (default - requires authentication)
image: 865783518572.dkr.ecr.us-east-1.amazonaws.com/apipark/apinto:latest-amd64

# Public ECR (alternative - no authentication needed)
image: public.ecr.aws/e5v3y2z9/apinto:latest-amd64
```

**Benefits:**
- Private repository for internal/production use
- Public repository for easy testing and distribution

### 3. Script Updates

#### Setup Script (`k8s/setup-ecr-auth.sh`)
- Now displays both private and public image options
- Automatically detects architecture and recommends appropriate image
- Includes note about public ECR not requiring authentication

#### Verification Script (`scripts/verify-ecr-build.sh`)
- Checks both private and public ECR repositories
- Displays images in private repository
- Shows public repository URI if available

#### Test Deployment Script (`scripts/test-deployment.sh`)
- Updated to use private repository by default
- Changed default tag to `latest-amd64`

---

## Repository Information

### Private ECR Repository

**Name:** `apipark/apinto`
**URI:** `865783518572.dkr.ecr.us-east-1.amazonaws.com/apipark/apinto`
**Region:** `us-east-1`
**Access:** Requires AWS credentials / IAM roles

**Use Cases:**
- Production deployments
- Internal development
- Secure, private image storage

### Public ECR Repository

**Name:** `apinto`
**URI:** `public.ecr.aws/e5v3y2z9/apinto`
**Region:** N/A (public registry)
**Access:** Public - no authentication required

**Use Cases:**
- Public distribution
- Quick testing
- External developers
- CI/CD without AWS credentials

---

## Image Availability

Both repositories will contain:

| Tag | Architecture | Description |
|-----|-------------|-------------|
| `latest-amd64` | linux/amd64 | Latest stable build for Intel/AMD |
| `latest-arm64` | linux/arm64 | Latest stable build for ARM |
| `{VERSION}-amd64` | linux/amd64 | Specific version for Intel/AMD |
| `{VERSION}-arm64` | linux/arm64 | Specific version for ARM |

**Standard Architecture:** `amd64` (for most Linux systems)

---

## Deployment Options

### Option 1: Private ECR (Recommended for Production)

**Requires:** ECR authentication setup

```bash
# Setup authentication
cd k8s
./setup-ecr-auth.sh

# Deploy
kubectl apply -f deployment.yaml
```

**Advantages:**
- Full control
- Private and secure
- Integrated with AWS IAM

### Option 2: Public ECR (Easy for Testing)

**Requires:** No authentication

```yaml
# Update deployment.yaml
image: public.ecr.aws/e5v3y2z9/apinto:latest-amd64
# Remove imagePullSecrets section
```

**Advantages:**
- No authentication needed
- Faster setup
- Great for demos and testing

---

## CI/CD Pipeline Flow

### Build Trigger
Push to `main` branch with conventional commit message

### Pipeline Steps

1. **Semantic Versioning**
   - Analyzes commits
   - Determines next version
   - Creates version tags

2. **Build Images**
   - Builds for amd64 and arm64 (parallel)
   - Creates tarballs from Go source
   - Builds Docker images with Dockerfile

3. **Push to Registries**
   - Login to Private ECR
   - Login to Public ECR
   - Push images to both registries
   - Tags with version and latest

4. **Create Release**
   - Creates GitHub release
   - Generates changelog
   - Attaches artifacts

---

## Breaking Changes

### ⚠️ Image URLs Changed

**Old:**
```
865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:0.1.0-amd64
```

**New:**
```
# Private
865783518572.dkr.ecr.us-east-1.amazonaws.com/apipark/apinto:latest-amd64

# Public
public.ecr.aws/e5v3y2z9/apinto:latest-amd64
```

### Migration Steps for Existing Deployments

1. **Update Image References:**
   ```bash
   kubectl set image deployment/apinto-gateway \
       apinto=865783518572.dkr.ecr.us-east-1.amazonaws.com/apipark/apinto:latest-amd64 \
       -n apinto-gateway
   ```

2. **Or use Public ECR:**
   ```bash
   kubectl set image deployment/apinto-gateway \
       apinto=public.ecr.aws/e5v3y2z9/apinto:latest-amd64 \
       -n apinto-gateway
   ```

3. **Watch Rollout:**
   ```bash
   kubectl rollout status deployment/apinto-gateway -n apinto-gateway
   ```

---

## Testing the Pipeline

### Trigger a Build

```bash
# Make a change and commit with conventional format
git commit -m "feat: update to use existing ECR repositories

- Migrate to apipark/apinto (private)
- Add public.ecr.aws/e5v3y2z9/apinto (public)
- Support dual-registry push"

git push origin main
```

### Watch Build Progress

```bash
# Via GitHub Actions UI
https://github.com/sagarkolli249/apinto/actions

# Via CLI (if gh installed)
gh run watch
```

### Verify Images

```bash
# Check private ECR
./scripts/verify-ecr-build.sh us-east-1 apipark/apinto

# Check public ECR
aws ecr-public describe-images \
    --repository-name apinto \
    --region us-east-1
```

### Test Deployment

```bash
# Test private ECR image
./scripts/test-deployment.sh us-east-1 apipark/apinto latest-amd64

# Or test public ECR (no auth needed)
docker pull public.ecr.aws/e5v3y2z9/apinto:latest-amd64
docker run -p 8080:8080 public.ecr.aws/e5v3y2z9/apinto:latest-amd64
```

---

## Troubleshooting

### Issue: Cannot pull from private ECR

**Error:** `unauthorized: authentication required`

**Solution:**
```bash
# For Kubernetes
cd k8s
./setup-ecr-auth.sh

# For Docker
aws ecr get-login-password --region us-east-1 | \
    docker login --username AWS --password-stdin \
    865783518572.dkr.ecr.us-east-1.amazonaws.com
```

### Issue: Public ECR image not found

**Check if image was pushed:**
```bash
aws ecr-public describe-images \
    --repository-name apinto \
    --region us-east-1 \
    --query 'images[*].imageTags' \
    --output table
```

**If missing:** Wait for CI/CD pipeline to complete or check GitHub Actions logs

### Issue: Wrong architecture error

**Error:** `exec format error`

**Solution:** Use correct architecture tag:
- For Intel/AMD: `latest-amd64`
- For ARM: `latest-arm64`

---

## Benefits of This Setup

### 1. Dual-Registry Strategy
- **Private:** Secure production images
- **Public:** Easy distribution and testing

### 2. Automatic Versioning
- Semantic versions from commit messages
- No manual version management
- Consistent tagging across registries

### 3. Multi-Architecture Support
- Single pipeline builds both amd64 and arm64
- Optimized for different cloud instances
- Cost savings with ARM instances (Graviton)

### 4. Simplified Deployment
- Latest tags always point to newest version
- Version-specific tags for rollbacks
- Choose private or public based on needs

---

## Next Steps

1. ✅ CI/CD pipeline updated
2. ✅ Kubernetes manifests updated
3. ✅ Scripts updated
4. ⏳ **Trigger first build** with new configuration
5. ⏳ **Verify images** in both registries
6. ⏳ **Test deployment** with new image URLs
7. 📝 Update existing deployments to use new URLs

---

## Quick Reference

### Image URLs

**Private ECR (amd64):**
```
865783518572.dkr.ecr.us-east-1.amazonaws.com/apipark/apinto:latest-amd64
```

**Public ECR (amd64):**
```
public.ecr.aws/e5v3y2z9/apinto:latest-amd64
```

### Common Commands

```bash
# Setup ECR auth for Kubernetes
./k8s/setup-ecr-auth.sh

# Verify images in ECR
./scripts/verify-ecr-build.sh

# Test deployment
./scripts/test-deployment.sh us-east-1 apipark/apinto latest-amd64

# Deploy to Kubernetes
kubectl apply -f k8s/deployment.yaml

# Watch GitHub Actions
gh run watch

# Pull public image (no auth)
docker pull public.ecr.aws/e5v3y2z9/apinto:latest-amd64
```

---

**Migration Complete!** 🎉

The CI/CD pipeline is now configured to use your existing ECR repositories with full support for both private and public registries.
