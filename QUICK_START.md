# Quick Start: ECR CI/CD Pipeline

## Prerequisites Setup (IMPORTANT!)

Before the GitHub Actions workflow can push to ECR, you need to set up AWS credentials in GitHub Secrets.

### Option 1: Automated Setup (Recommended)

Run the setup script:
```bash
./scripts/setup-github-secrets.sh
```

This will:
- Create IAM user for GitHub Actions
- Create and attach ECR policy
- Generate access keys
- Optionally set GitHub secrets (requires gh CLI)

### Option 2: Manual Setup

1. **Create IAM User:**
   ```bash
   aws iam create-user --user-name github-actions-ecr
   ```

2. **Create IAM Policy:**
   Save this as `ecr-policy.json`:
   ```json
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
   ```

   Create policy:
   ```bash
   aws iam create-policy \
       --policy-name ECRPushPolicy \
       --policy-document file://ecr-policy.json
   ```

3. **Attach Policy to User:**
   ```bash
   ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
   aws iam attach-user-policy \
       --user-name github-actions-ecr \
       --policy-arn arn:aws:iam::${ACCOUNT_ID}:policy/ECRPushPolicy
   ```

4. **Create Access Key:**
   ```bash
   aws iam create-access-key --user-name github-actions-ecr
   ```

   Save the `AccessKeyId` and `SecretAccessKey` from the output!

5. **Add to GitHub Secrets:**
   - Go to: https://github.com/sagarkolli249/apinto/settings/secrets/actions
   - Click "New repository secret"
   - Add:
     - Name: `AWS_ACCESS_KEY_ID`, Value: [Your Access Key ID]
     - Name: `AWS_SECRET_ACCESS_KEY`, Value: [Your Secret Access Key]

## Current Status

✅ **ECR Repository Created:**
- Repository: `apinto-gateway`
- Region: `us-east-1`
- URI: `865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway`

✅ **Workflow Files Ready:**
- `.github/workflows/ecr-build-push.yaml` - Main CI/CD workflow
- `.releaserc.json` - Semantic versioning config

⚠️ **Action Required:**
- Set up GitHub Secrets (AWS credentials) - See above

## Trigger a Build

Once GitHub Secrets are set, push any commit with conventional commit format:

```bash
# This will trigger a minor version bump (0.x.0 -> 0.y.0)
git commit -m "feat: add new feature"
git push origin main

# This will trigger a patch version bump (0.0.x -> 0.0.y)
git commit -m "fix: resolve bug"
git push origin main
```

## Monitor the Build

1. **GitHub Actions:**
   https://github.com/sagarkolli249/apinto/actions

2. **Check workflow run:**
   - Look for "Build and Push to ECR with Semantic Versioning"
   - Click on the latest run to see logs

## Verify Images in ECR

After build completes, verify images:

```bash
./scripts/verify-ecr-build.sh
```

This will:
- Check if ECR repository exists
- List all images with tags
- Show multi-arch support status
- Display image URIs

## Test Deployment

Deploy and test the image locally:

```bash
./scripts/test-deployment.sh
```

This will:
- Login to ECR
- Pull the latest image
- Start a container
- Run health checks
- Show logs and status

## Troubleshooting

### Build Not Starting

**Problem:** Workflow doesn't trigger after push

**Solutions:**
1. Check commit message format (must use conventional commits)
2. Verify push is to `main` or `develop` branch
3. Check GitHub Actions tab for any errors

### Build Fails at AWS Authentication

**Problem:** Error: "Unable to locate credentials"

**Solution:** GitHub Secrets not set. Follow setup steps above.

### Build Fails at ECR Push

**Problem:** Error: "repository not found"

**Solution:** ECR repository should auto-exist now. If not:
```bash
aws ecr create-repository --repository-name apinto-gateway --region us-east-1
```

### No Version Bump

**Problem:** Workflow runs but no release created

**Solution:**
- Check commit message follows conventional commits
- Types that trigger releases: `feat`, `fix`, `perf`, `refactor`
- Types that don't: `docs`, `style`, `test`, `chore`, `ci`

## Next Steps

1. ✅ ECR repository created
2. ⚠️ Set up GitHub Secrets (AWS credentials)
3. 🚀 Push commit to trigger build
4. 👀 Monitor workflow in GitHub Actions
5. ✅ Verify images with `./scripts/verify-ecr-build.sh`
6. 🧪 Test deployment with `./scripts/test-deployment.sh`

## Useful Commands

```bash
# Check if secrets are set (requires gh CLI)
gh secret list

# View latest workflow runs (requires gh CLI)
gh run list --limit 5

# View workflow logs (requires gh CLI)
gh run view [RUN_ID] --log

# List ECR images
aws ecr describe-images \
    --repository-name apinto-gateway \
    --region us-east-1 \
    --query 'sort_by(imageDetails,& imagePushedAt)[-5:].[imageTags[0],imagePushedAt]' \
    --output table

# Pull image from ECR
aws ecr get-login-password --region us-east-1 | \
    docker login --username AWS --password-stdin 865783518572.dkr.ecr.us-east-1.amazonaws.com
docker pull 865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:latest

# Run container
docker run -d \
    --name apinto \
    -p 8080:8080 \
    865783518572.dkr.ecr.us-east-1.amazonaws.com/apinto-gateway:latest
```

## Documentation

- [ECR CI/CD Setup Guide](docs/ECR_CI_CD_SETUP.md) - Comprehensive setup guide
- [Commit Conventions](docs/COMMIT_CONVENTIONS.md) - Commit message format
- [Contributing Guide](CONTRIBUTING.md) - Development guidelines

## Support

If you encounter issues:
1. Check workflow logs in GitHub Actions
2. Verify AWS credentials: `aws sts get-caller-identity`
3. Check ECR repository: `aws ecr describe-repositories`
4. Review documentation in `docs/` folder
