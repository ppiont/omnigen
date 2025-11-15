# GitHub OIDC Setup for AWS

This directory contains Terraform configuration for setting up OpenID Connect (OIDC) authentication between GitHub Actions and AWS. This allows GitHub Actions workflows to authenticate to AWS without storing long-lived credentials.

## What This Creates

1. **OIDC Provider**: Establishes trust between GitHub Actions and AWS
2. **IAM Role**: `omnigen-github-actions` role that workflows will assume
3. **IAM Policies**: Three policies attached to the role:
   - **Terraform Operations**: Full infrastructure management permissions
   - **Backend Deployment**: ECR push and ECS deployment permissions
   - **Frontend Deployment**: S3 sync and CloudFront invalidation permissions

## Prerequisites

Before deploying this OIDC setup, you must:

1. Have AWS CLI configured with admin credentials
2. Have Terraform >= 1.13.5 installed
3. Have created the remote backend (S3 + DynamoDB) for Terraform state

## Deployment Steps

### 1. Navigate to this directory

```bash
cd infrastructure/github-oidc
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Review the plan

```bash
terraform plan
```

You should see it will create:
- 1 OIDC provider
- 1 IAM role
- 3 IAM policies
- 3 policy attachments

### 4. Apply the configuration

```bash
terraform apply
```

Type `yes` when prompted.

### 5. Save the role ARN

After successful apply, note the output:

```bash
terraform output github_actions_role_arn
```

Example output:
```
arn:aws:iam::123456789012:role/omnigen-github-actions
```

### 6. Add GitHub Secrets

Go to your GitHub repository settings and add these secrets:

**Repository Secrets:**
1. Navigate to: `Settings` → `Secrets and variables` → `Actions` → `New repository secret`

2. Add `AWS_ROLE_ARN`:
   - Name: `AWS_ROLE_ARN`
   - Value: The ARN from step 5 (e.g., `arn:aws:iam::123456789012:role/omnigen-github-actions`)

3. Add `REPLICATE_API_KEY_SECRET_ARN`:
   - Name: `REPLICATE_API_KEY_SECRET_ARN`
   - Value: Your existing Secrets Manager ARN for Replicate API key

**Repository Variables:**
1. Navigate to: `Settings` → `Secrets and variables` → `Actions` → `Variables` → `New repository variable`

2. Add `AWS_REGION`:
   - Name: `AWS_REGION`
   - Value: `us-east-1` (or your preferred region)

## Customization

### Change GitHub Repository

If deploying for a different repository, update `variables.tf`:

```hcl
variable "github_org" {
  default = "your-org-or-username"
}

variable "github_repo" {
  default = "your-repo-name"
}
```

### Adjust Permissions

To modify IAM permissions, edit the policy blocks in `main.tf`:
- `aws_iam_policy.terraform_operations` - Infrastructure management
- `aws_iam_policy.backend_deployment` - Backend deployment
- `aws_iam_policy.frontend_deployment` - Frontend deployment

## Security Considerations

1. **Scoped Trust Policy**: The OIDC trust policy is scoped to your specific repository using:
   ```
   "repo:ppiont/omnigen:*"
   ```

2. **No Long-Lived Credentials**: GitHub Actions receives temporary credentials (valid for 1 hour) via STS AssumeRoleWithWebIdentity

3. **Least Privilege**: Policies are scoped to specific resources where possible

4. **Audit Trail**: All AWS API calls are logged in CloudTrail with the GitHub Actions role identity

## Troubleshooting

### Error: "Not authorized to perform sts:AssumeRoleWithWebIdentity"

**Cause**: The trust policy doesn't match your repository or branch.

**Fix**: Verify `github_org` and `github_repo` in `variables.tf` match your GitHub repository exactly.

### Error: "Access Denied" during workflow execution

**Cause**: Missing permissions in one of the IAM policies.

**Fix**: Check CloudTrail logs to see which specific action was denied, then add it to the appropriate policy in `main.tf`.

### OIDC thumbprint warnings

**Cause**: GitHub may have updated their certificate thumbprints.

**Fix**: GitHub's current thumbprints are included in `main.tf`. If you see warnings, check the [official documentation](https://github.blog/changelog/2023-06-27-github-actions-update-on-oidc-integration-with-aws/) for the latest values.

## Cleanup

To remove the OIDC setup:

```bash
terraform destroy
```

**Warning**: This will break GitHub Actions workflows until recreated.

## Next Steps

After setting up OIDC:

1. Configure the remote Terraform backend (S3 + DynamoDB) - see `../README.md`
2. Migrate Terraform state to S3 backend
3. Test GitHub Actions workflows with a test PR

## References

- [AWS OIDC Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html)
- [GitHub Actions OIDC with AWS](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [aws-actions/configure-aws-credentials](https://github.com/aws-actions/configure-aws-credentials)
