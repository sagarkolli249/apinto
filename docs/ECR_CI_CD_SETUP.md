# ECR CI/CD Pipeline Setup Guide

This guide explains how to set up and use the automated CI/CD pipeline for building and pushing Docker images to AWS Elastic Container Registry (ECR) with automatic semantic versioning.

## Overview

The CI/CD pipeline automatically:
- Determines the next semantic version based on commit messages
- Builds multi-architecture Docker images (amd64 and arm64)
- Pushes images to AWS ECR
- Creates multi-arch manifests
- Generates GitHub releases with changelogs

## Prerequisites

1. **AWS Account** with ECR repository created
2. **GitHub Repository** with appropriate permissions
3. **AWS Credentials** with ECR push permissions

## Setup Instructions

### 1. Create ECR Repository

```bash
# Login to AWS Console or use AWS CLI
aws ecr create-repository \
    --repository-name apinto-gateway \
    --region us-east-1 \
    --image-scanning-configuration scanOnPush=true
```

### 2. Configure GitHub Secrets

Add the following secrets to your GitHub repository:
- Go to: Settings → Secrets and variables → Actions → New repository secret

Required secrets:
- `AWS_ACCESS_KEY_ID`: Your AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key

**Creating AWS IAM User for GitHub Actions:**

```bash
# Create IAM policy (save as ecr-push-policy.json)
{
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
}

# Create IAM user and attach policy
aws iam create-user --user-name github-actions-ecr
aws iam create-policy --policy-name ECRPushPolicy --policy-document file://ecr-push-policy.json
aws iam attach-user-policy --user-name github-actions-ecr --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/ECRPushPolicy
aws iam create-access-key --user-name github-actions-ecr
```

### 3. Update Workflow Configuration

Edit `.github/workflows/ecr-build-push.yaml` and update these variables:

```yaml
env:
  AWS_REGION: us-east-1              # Change to your AWS region
  ECR_REPOSITORY: apinto-gateway     # Change to your ECR repository name
```

### 4. Verify Build Scripts Permissions

Ensure build scripts have execute permissions:

```bash
chmod +x build/cmd/*.sh
git add build/cmd/*.sh
git commit -m "chore: make build scripts executable"
```

## Semantic Versioning

### Commit Message Format

The pipeline uses **Conventional Commits** to determine version bumps:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Examples:**

```bash
# Patch release (0.1.0 → 0.1.1)
git commit -m "fix: resolve authentication issue in bedrock provider"

# Minor release (0.1.0 → 0.2.0)
git commit -m "feat: add support for AWS Bedrock tools API"

# Major release (0.1.0 → 1.0.0)
git commit -m "feat!: redesign strategy actuator interface

BREAKING CHANGE: Strategy actuators now require implementing new interface methods"
```

### Commit Types and Version Bumps

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor (0.1.0 → 0.2.0) |
| `fix` | Bug fix | Patch (0.1.0 → 0.1.1) |
| `perf` | Performance improvement | Patch |
| `refactor` | Code refactoring | Patch |
| `docs` | Documentation only | None |
| `style` | Code style changes | None |
| `test` | Test changes | None |
| `chore` | Maintenance tasks | None |
| `ci` | CI/CD changes | None |
| `BREAKING CHANGE` | Breaking changes | Major (0.1.0 → 1.0.0) |

### Manual Version Override

To manually create a release with a specific version:

```bash
# Create and push a tag
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3

# Or use semantic-release CLI locally
npx semantic-release --no-ci
```

## Pipeline Workflow

### Trigger Conditions

The pipeline runs on:
- **Push to `main` branch**: Full build, push to ECR, and release
- **Push to `develop` branch**: Build with beta pre-release version
- **Pull requests to `main`**: Version calculation only (no push)

### Pipeline Stages

1. **semantic-version**
   - Analyzes commit history
   - Determines next version using conventional commits
   - Outputs version number for subsequent jobs

2. **build-and-push** (Matrix Strategy)
   - Builds for amd64 and arm64 architectures in parallel
   - Compiles Go binary
   - Creates Docker images
   - Pushes to ECR with version tag

3. **create-manifest**
   - Creates multi-arch manifest combining amd64 and arm64 images
   - Tags with version number (e.g., `1.2.3`)
   - Tags with `latest`

4. **release**
   - Creates GitHub release
   - Generates changelog from commits
   - Attaches release artifacts

## Using the Images

### Pull from ECR

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | \
    docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Pull specific version
docker pull YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:1.2.3

# Pull latest
docker pull YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:latest

# Pull specific architecture
docker pull YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:1.2.3-amd64
```

### Run Container

```bash
docker run -d \
    -p 8080:8080 \
    -v /var/lib/apinto:/var/lib/apinto \
    YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:latest
```

## Monitoring and Debugging

### View Workflow Runs

1. Go to: GitHub repository → Actions tab
2. Select "Build and Push to ECR with Semantic Versioning"
3. Click on a specific run to see logs

### Common Issues

#### 1. AWS Authentication Failed

```
Error: Unable to locate credentials
```

**Solution:** Verify GitHub secrets are set correctly:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

#### 2. ECR Repository Not Found

```
Error: repository not found
```

**Solution:** Ensure ECR repository exists:
```bash
aws ecr describe-repositories --repository-names apinto-gateway --region us-east-1
```

#### 3. No Version Bump Triggered

**Solution:** Check commit messages follow conventional commit format:
```bash
# Bad
git commit -m "updated code"

# Good
git commit -m "feat: add new authentication method"
```

#### 4. Build Script Permission Denied

```
Error: Permission denied: ./build/cmd/package.sh
```

**Solution:**
```bash
chmod +x build/cmd/*.sh
git add build/cmd/*.sh
git commit -m "chore: fix build script permissions"
git push
```

## Best Practices

### 1. Commit Messages

- Use clear, descriptive commit messages
- Follow conventional commit format
- Include scope when applicable: `feat(auth): add OAuth2 support`
- Document breaking changes in commit body

### 2. Branch Strategy

- `main`: Production-ready code, triggers releases
- `develop`: Pre-release code, creates beta versions
- Feature branches: Create PRs to `main` or `develop`

### 3. Testing

- Test locally before pushing:
  ```bash
  # Build locally
  ./build/cmd/package.sh 0.0.0-test amd64

  # Test Docker build
  ./build/cmd/docker_build.sh 0.0.0-test username amd64
  ```

### 4. Version Management

- Don't manually edit version in `package.json`
- Let semantic-release handle versioning
- Use tags for hotfixes: `git tag v1.2.4`

## Advanced Configuration

### Custom Release Rules

Edit `.releaserc.json` to customize versioning rules:

```json
{
  "releaseRules": [
    { "type": "feat", "release": "minor" },
    { "type": "hotfix", "release": "patch" },
    { "scope": "breaking", "release": "major" }
  ]
}
```

### Multi-Region ECR

To push to multiple ECR regions, duplicate the `build-and-push` job with different regions:

```yaml
strategy:
  matrix:
    arch: [amd64, arm64]
    region: [us-east-1, eu-west-1]
```

### Pre-release Versions

For beta/alpha releases from `develop`:

```bash
git checkout develop
git commit -m "feat: experimental feature"
git push origin develop

# Creates version like: 1.2.3-beta.1
```

## Rollback Strategy

### Rolling Back to Previous Version

```bash
# List available versions
aws ecr describe-images \
    --repository-name apinto-gateway \
    --query 'sort_by(imageDetails,& imagePushedAt)[*].[imageTags[0],imagePushedAt]' \
    --output table

# Update deployment to use previous version
kubectl set image deployment/apinto apinto=YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:1.2.2
```

## Maintenance

### Clean Up Old Images

```bash
# Delete images older than 30 days
aws ecr list-images \
    --repository-name apinto-gateway \
    --filter tagStatus=UNTAGGED \
    --query 'imageIds[*]' \
    --output json | \
jq -r '.[] | .imageDigest' | \
while read digest; do
    aws ecr batch-delete-image \
        --repository-name apinto-gateway \
        --image-ids imageDigest=$digest
done
```

### Lifecycle Policy

Create an ECR lifecycle policy to automatically clean old images:

```json
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Keep last 10 images",
      "selection": {
        "tagStatus": "any",
        "countType": "imageCountMoreThan",
        "countNumber": 10
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
```

Apply policy:
```bash
aws ecr put-lifecycle-policy \
    --repository-name apinto-gateway \
    --lifecycle-policy-text file://lifecycle-policy.json
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/sagarkolli249/apinto/issues
- Workflow Logs: Check Actions tab in GitHub repository

## References

- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [AWS ECR Documentation](https://docs.aws.amazon.com/ecr/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
