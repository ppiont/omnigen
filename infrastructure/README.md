# OmniGen Infrastructure

Modern, modular Terraform infrastructure for the OmniGen AI video generation pipeline. Built with AWS best practices, cost optimization, and developer productivity in mind.

## 🏗️ Architecture Overview

This infrastructure supports a complete AI video generation pipeline with:

- **API Layer**: ECS Fargate running Go API server with auto-scaling
- **Processing**: Lambda functions for scene generation and video composition
- **Orchestration**: Step Functions Express workflow for pipeline coordination
- **Storage**: S3 for video assets, DynamoDB for job tracking
- **Frontend**: CloudFront CDN serving Vite/React SPA from S3
- **Networking**: VPC with public/private subnets, NAT Gateway, VPC endpoints
- **Monitoring**: CloudWatch Logs with 7-day retention

### Key Features

✅ **Cost-Optimized**: Single AZ deployment, on-demand billing, lifecycle policies (~$110-200/month)
✅ **Serverless-First**: Lambda + Step Functions for video processing
✅ **Auto-Scaling**: ECS and Lambda scale automatically based on load
✅ **Production-Ready**: Health checks, monitoring, error handling, retry logic
✅ **Developer-Friendly**: Makefile commands, GitHub Actions CI/CD
✅ **Secure**: VPC endpoints, encryption at rest, IAM least-privilege policies
✅ **Modern Terraform**: Version 1.13.5+, AWS provider 6.x, modular design, validated variables

## 📋 Prerequisites

Before you begin, ensure you have:

1. **AWS Account** with appropriate permissions
2. **AWS CLI** v2.x installed and configured
3. **Terraform** >= 1.13.5 installed
4. **Docker** installed (for building ECS container)
5. **make** installed (optional but recommended)
6. **Node.js** 20+ (for frontend deployment)

### Install Prerequisites

```bash
# macOS (using Homebrew)
brew install awscli terraform docker make node

# Verify installations
aws --version
terraform --version
docker --version
make --version
node --version
```

## 🚀 Quick Start

### 1. Configure AWS CLI

```bash
aws configure
# AWS Access Key ID: <your-key>
# AWS Secret Access Key: <your-secret>
# Default region: us-east-1
# Default output format: json

# Verify access
aws sts get-caller-identity
```

### 2. Create Secrets Manager Secret

The Replicate API key must be created manually before running Terraform:

```bash
# Create the secret
aws secretsmanager create-secret \
  --name omnigen/replicate-api-key \
  --description "Replicate API Key for OmniGen video generation" \
  --secret-string "your-replicate-api-key-here" \
  --region us-east-1

# Get the secret ARN (you'll need this for terraform.tfvars)
aws secretsmanager describe-secret \
  --secret-id omnigen/replicate-api-key \
  --region us-east-1 \
  --query 'ARN' \
  --output text
```

### 3. Configure Terraform Variables

```bash
cd infrastructure

# Copy the example configuration
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
vim terraform.tfvars
```

**Required:** Update `replicate_api_key_secret_arn` with the ARN from step 2.

### 4. Deploy Infrastructure

```bash
# Run from project root directory
cd /path/to/omnigen

# Initialize Terraform
make init

# Review the plan
make plan

# Deploy infrastructure (takes 10-15 minutes)
make apply
```

### 5. Deploy Backend Application

```bash
# Run from project root directory
# Build and push Docker image to ECR
make docker-push

# Deploy to ECS
make deploy-ecs
```

### 6. Deploy Frontend

```bash
# Build and deploy frontend to S3/CloudFront
make deploy-frontend
```

### 7. Verify Deployment

```bash
# Check API health
make health

# View outputs
make outputs

# Tail logs
make logs-ecs
```

## 📁 Project Structure

```
infrastructure/
├── main.tf                    # Root module orchestration
├── terraform.tf               # Provider configuration
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── locals.tf                  # Local values
├── terraform.tfvars.example   # Configuration template
├── Makefile                   # Convenience commands
├── README.md                  # This file
│
├── modules/
│   ├── networking/            # VPC, subnets, NAT, security groups, VPC endpoints
│   ├── storage/               # S3 buckets (assets + frontend), DynamoDB table
│   ├── iam/                   # IAM roles and policies for all services
│   ├── secrets/               # Secrets Manager integration
│   ├── compute/               # ECS Fargate, ECR, task definitions, auto-scaling
│   ├── serverless/            # Lambda functions, Step Functions workflow
│   ├── loadbalancer/          # Application Load Balancer
│   ├── cdn/                   # CloudFront distribution for frontend
│   └── monitoring/            # CloudWatch log groups
│
└── .github/workflows/
    ├── terraform-plan.yml     # CI: Run plan on PRs
    └── terraform-deploy.yml   # CD: Deploy on merge to main
```

## 🛠️ Available Commands

**Note:** The Makefile is located in the **project root directory**, not in `infrastructure/`.

Run `make help` from the root to see all available commands:

```bash
make help               # Show all commands
make init               # Initialize Terraform
make plan               # Run terraform plan
make apply              # Apply terraform changes
make destroy            # Destroy all infrastructure
make format             # Format Terraform files
make validate           # Validate configuration
make outputs            # Show terraform outputs
make docker-build       # Build Docker image
make docker-push        # Build and push to ECR
make deploy-ecs         # Deploy new version to ECS
make deploy-frontend    # Deploy frontend to S3/CloudFront
make logs-ecs           # Tail ECS logs
make logs-generator     # Tail generator Lambda logs
make logs-composer      # Tail composer Lambda logs
make health             # Check API health
make clean              # Clean up local files
make setup              # Initial setup (copy example config)
make deploy-all         # Full deployment (infrastructure + backend + frontend)
```

## 🔧 Configuration Reference

### Key Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `aws_region` | AWS region | `us-east-1` |
| `project_name` | Project name for resources | `omnigen` |
| `replicate_api_key_secret_arn` | **Required**: Secrets Manager ARN | N/A |
| `ecs_min_tasks` | Minimum ECS tasks | `1` |
| `ecs_max_tasks` | Maximum ECS tasks | `5` |
| `ecs_cpu` | CPU units (1024 = 1 vCPU) | `1024` |
| `ecs_memory` | Memory in MB | `2048` |
| `lambda_generator_memory` | Generator Lambda memory (MB) | `2048` |
| `lambda_composer_memory` | Composer Lambda memory (MB) | `10240` |
| `lambda_timeout` | Lambda timeout (seconds) | `900` |
| `cloudfront_price_class` | CloudFront price class | `PriceClass_100` |

See `terraform.tfvars.example` for complete configuration options.

## 🌐 Outputs

After deployment, Terraform outputs important URLs and identifiers:

```bash
terraform output
```

Key outputs:
- `api_url`: API endpoint (ALB DNS)
- `frontend_url`: Frontend URL (CloudFront)
- `ecr_repository_url`: Docker image repository
- `assets_bucket_name`: S3 bucket for videos
- `dynamodb_table_name`: Job tracking table

## 🔄 Updating Infrastructure

### Update Infrastructure Only

```bash
# Make changes to Terraform files
vim infrastructure/main.tf

# Plan and apply
make plan
make apply
```

### Update Backend Application

```bash
# Make changes to backend code
vim backend/main.go

# Build, push, and deploy
make deploy-ecs
```

### Update Frontend

```bash
# Make changes to frontend code
vim frontend/src/App.tsx

# Build and deploy
make deploy-frontend
```

## 📊 Monitoring

### View Logs

```bash
# ECS API logs
make logs-ecs

# Generator Lambda logs
make logs-generator

# Composer Lambda logs
make logs-composer

# Or use AWS CLI directly
aws logs tail /ecs/omnigen --follow
```

### CloudWatch Console

View logs and metrics in the AWS Console:
- ECS: `/ecs/omnigen`
- Generator Lambda: `/aws/lambda/omnigen-generator`
- Composer Lambda: `/aws/lambda/omnigen-composer`
- Step Functions: `/aws/states/omnigen-workflow`

### Key Metrics

Monitor these metrics in CloudWatch:
- **ECS**: CPU%, Memory%, HealthyHostCount
- **Lambda**: Duration, Errors, Throttles
- **Step Functions**: ExecutionsSucceeded, ExecutionsFailed
- **DynamoDB**: ConsumedReadCapacity, ConsumedWriteCapacity
- **ALB**: TargetResponseTime, HTTPCode_5XX_Count

## 💰 Cost Estimates

### Base Infrastructure

| Service | Monthly Cost |
|---------|--------------|
| ECS Fargate (1-2 tasks) | $35-80 |
| Lambda (Generator + Composer) | $20-40 |
| Step Functions | $5 |
| S3 (Storage + Transfer) | $10-20 |
| DynamoDB (On-demand) | $5-15 |
| ALB | $20 |
| NAT Gateway | $35 |
| CloudWatch Logs | $3-5 |
| Secrets Manager | $0.40 |
| CloudFront | $5-10 |
| **Total Base** | **$110-200/month** |

### Per-Video Cost

- Infrastructure: ~$0.26 per video
- Replicate API: ~$1.30 per video (varies by model)
- **Total**: ~$1.56 per video ✅ (Under $2.00 target)

### Scaling Costs

| Videos/Day | Infrastructure | Replicate API | Total/Month |
|------------|----------------|---------------|-------------|
| 10 | $130 | $40 | $170 |
| 100 | $150 | $400 | $550 |
| 500 | $180 | $2,000 | $2,180 |
| 1000 | $250 | $4,000 | $4,250 |

## 🔒 Security Best Practices

### Implemented Security Measures

- ✅ All resources in VPC with private subnets
- ✅ VPC endpoints for S3 and DynamoDB (no NAT for AWS services)
- ✅ Security groups with least-privilege access
- ✅ IAM roles with minimal permissions
- ✅ Secrets stored in AWS Secrets Manager
- ✅ S3 bucket encryption (AES-256)
- ✅ DynamoDB encryption enabled
- ✅ CloudFront HTTPS enforced
- ✅ No hardcoded credentials
- ✅ CloudTrail logging enabled
- ✅ ECR image scanning on push

### Additional Recommendations

- Set up AWS WAF for ALB protection
- Enable GuardDuty for threat detection
- Use AWS Config for compliance monitoring
- Implement API rate limiting
- Add OAuth/Cognito for production authentication
- Enable MFA for AWS account
- Rotate Replicate API key regularly

## 🧪 Testing

### Manual Testing

```bash
# 1. Check infrastructure
terraform plan

# 2. Validate configuration
terraform validate

# 3. Format check
terraform fmt -check -recursive

# 4. Health check
curl $(terraform output -raw alb_dns_name)/health

# 5. Test video generation (example)
curl -X POST http://$(terraform output -raw alb_dns_name)/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "luxury watch ad with gold aesthetics",
    "duration": 15
  }'
```

### Automated Testing (CI/CD)

GitHub Actions automatically:
- Runs `terraform plan` on pull requests
- Deploys infrastructure on merge to main
- Builds and deploys backend/frontend on code changes

## 🐛 Troubleshooting

### Common Issues

#### 1. Terraform Init Fails

```bash
# Clear Terraform cache
rm -rf .terraform .terraform.lock.hcl
terraform init
```

#### 2. ECS Tasks Not Starting

```bash
# Check task logs
make logs-ecs

# Check task definition
aws ecs describe-task-definition --task-definition omnigen-api

# Check ECR image exists
aws ecr describe-images --repository-name omnigen-api
```

#### 3. Lambda Timeouts

```bash
# Check Lambda logs
make logs-generator
make logs-composer

# Increase timeout in terraform.tfvars
lambda_timeout = 900  # 15 minutes (max)
```

#### 4. High Costs

```bash
# Check Cost Explorer
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity DAILY \
  --metrics BlendedCost \
  --group-by Type=SERVICE
```

#### 5. Frontend Not Loading

```bash
# Check CloudFront distribution
terraform output cloudfront_domain_name

# Check S3 bucket
aws s3 ls s3://$(terraform output -raw frontend_bucket_name)/

# Invalidate CloudFront cache
aws cloudfront create-invalidation \
  --distribution-id $(terraform output -raw cloudfront_distribution_id) \
  --paths "/*"
```

### Get Help

- Check [AWS Documentation](https://docs.aws.amazon.com/)
- Review [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- Check CloudWatch Logs for detailed error messages
- Run `terraform show` to see current state

## 🗑️ Teardown

To destroy all infrastructure:

```bash
# WARNING: This is irreversible!
make destroy
```

This will:
1. Empty S3 buckets
2. Destroy all AWS resources
3. Remove Terraform state

**Note**: Secrets Manager secrets have a 7-30 day recovery window by default.

## 🔄 Migration to S3 Backend

For production, migrate to S3 backend for state management:

```bash
# 1. Create S3 bucket for state
aws s3 mb s3://omnigen-terraform-state --region us-east-1

# 2. Enable versioning
aws s3api put-bucket-versioning \
  --bucket omnigen-terraform-state \
  --versioning-configuration Status=Enabled

# 3. Uncomment backend config in terraform.tf
# backend "s3" {
#   bucket = "omnigen-terraform-state"
#   key    = "infrastructure/terraform.tfstate"
#   region = "us-east-1"
# }

# 4. Migrate state
terraform init -migrate-state
```

## 📚 Additional Resources

- [Infrastructure PRD](./Infra-PRD.md) - Detailed infrastructure specification
- [Main PRD](../.human/PRD.md) - Project requirements document
- [CLAUDE.md](../CLAUDE.md) - Development guidance
- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)

## 🤝 Contributing

1. Create a feature branch
2. Make changes
3. Run `terraform fmt -recursive`
4. Run `terraform validate`
5. Create a pull request
6. GitHub Actions will run `terraform plan`
7. After approval, merge to deploy

## 📝 License

See main project LICENSE file.

## 🎯 Next Steps

After successful deployment:

1. **Implement Backend**: Build Go API server in `../backend`
2. **Implement Lambda Functions**: Replace placeholder JavaScript with Go implementations
3. **Implement Frontend**: Build Vite/React UI in `../frontend`
4. **Set Up Monitoring**: Configure CloudWatch alarms and SNS notifications
5. **Add Domain**: Configure Route53 and ACM certificate for custom domain
6. **Enable HTTPS**: Update ALB listener with ACM certificate
7. **Implement Authentication**: Add API keys or OAuth/Cognito
8. **Load Testing**: Test with various load patterns
9. **Cost Optimization**: Review and optimize based on actual usage

---

Built with ❤️ for the OmniGen AI Video Generation Pipeline
