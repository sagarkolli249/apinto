# Deployment Configuration

The apinto service is deployed as a StatefulSet with automated CI/CD pipeline.

## EKS Deployment
- Cluster: ai-eks
- Namespace: default
- Image: public.ecr.aws/e5v3y2z9/apinto



Note: IAM policy updated to allow EKS cluster access

Added RBAC permissions for GitHub Actions IAM user

## Deployment Details
- Cluster: ai-eks
- Namespace: apipark (automated deployment)
- StatefulSet with volumeClaimTemplates for PVC management
- Non-root user security context enabled

