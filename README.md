# OmniGen

AI Video Generation Pipeline - Create professional-quality video content from high-level prompts

## Overview

OmniGen is an end-to-end pipeline that takes text prompts and outputs publication-ready video content with minimal human intervention. It supports music videos, ad creatives, and educational/explainer videos with synchronized audio, coherent visuals, and professional polish.

## Repository Structure

```
omnigen/
├── backend/         # Go API - Pipeline logic and AI orchestration
├── frontend/        # React UI - Web interface
├── infrastructure/  # Terraform IaC - AWS resources
└── .github/         # GitHub Actions CI/CD workflows
```

## Quick Start

### Prerequisites

- AWS CLI configured with credentials
- Terraform >= 1.13.5
- Node.js 20.x (for frontend)
- Go 1.25.4 (for backend)
- Bun or npm (for frontend package management)

### Local Development

1. **Backend (Go API)**
   ```bash
   cd backend
   go run cmd/api/main.go
   ```

2. **Frontend (React)**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

3. **Infrastructure**
   ```bash
   cd infrastructure
   terraform init
   terraform plan
   terraform apply
   ```

### Deployment

Deploy all components:
```bash
make deploy-all
```

Deploy individually:
```bash
make apply           # Infrastructure
make deploy-ecs      # Backend API
make deploy-frontend # Frontend
```

See `Makefile` for all available commands.

## CI/CD Pipeline

This project uses GitHub Actions with AWS OIDC for secure, automated deployments.

### Workflows

**Infrastructure** (`.github/workflows/infrastructure.yml`)
- Trigger: Changes to `infrastructure/**`
- PR: Terraform plan (posted as comment)
- Merge: Terraform apply + export outputs

**Backend** (`.github/workflows/backend.yml`)
- Trigger: Changes to `backend/**`
- PR: Build verification
- Merge: Build Docker image → Push to ECR → Deploy to ECS

**Frontend** (`.github/workflows/frontend.yml`)
- Trigger: Changes to `frontend/**`
- PR: Build verification
- Merge: Build → Deploy to S3 → Invalidate CloudFront

### Setup CI/CD

For detailed setup instructions, see **[CICD_SETUP.md](./CICD_SETUP.md)**

Quick summary:
1. Create S3 bucket and DynamoDB table for Terraform state
2. Migrate Terraform state to S3
3. Deploy OIDC infrastructure (`infrastructure/github-oidc/`)
4. Add GitHub secrets (`AWS_ROLE_ARN`, `REPLICATE_API_KEY_SECRET_ARN`)
5. Test workflows with a PR

## Architecture

**Infrastructure:**
- VPC with public/private subnets
- ECS Fargate for backend API
- S3 + CloudFront for frontend
- Lambda + Step Functions for video generation
- DynamoDB for state management
- AWS Cognito for authentication

**Backend:**
- Go REST API with Gin framework
- JWT authentication with Cognito
- Swagger documentation at `/swagger/`
- Health check at `/health`

**Frontend:**
- React with Vite
- Material-UI components
- AWS Cognito authentication
- Hosted UI integration

## API Documentation

Once deployed, access Swagger UI at:
```
https://<cloudfront-domain>/swagger/index.html
```

## Authentication

The application uses AWS Cognito for user authentication:
- User pool for user management
- Hosted UI for sign-up/sign-in
- JWT tokens for API authorization
- HTTP-only cookies for session management

## Environment Variables

**Backend:**
- `PORT` - API port (default: 8080)
- `AWS_REGION` - AWS region
- `ASSETS_BUCKET` - S3 bucket for generated assets
- `JOB_TABLE` - DynamoDB table for jobs
- `USAGE_TABLE` - DynamoDB table for usage tracking
- `STEP_FUNCTIONS_ARN` - Step Functions state machine ARN
- `REPLICATE_SECRET_ARN` - Secrets Manager ARN for Replicate API key
- `COGNITO_USER_POOL_ID` - Cognito user pool ID
- `COGNITO_CLIENT_ID` - Cognito app client ID
- `JWT_ISSUER` - JWT token issuer URL
- `COGNITO_DOMAIN` - Cognito hosted UI domain
- `CLOUDFRONT_DOMAIN` - CloudFront distribution domain

**Frontend:**
- `VITE_API_URL` - Backend API URL (CloudFront domain)
- `VITE_COGNITO_USER_POOL_ID` - Cognito user pool ID
- `VITE_COGNITO_CLIENT_ID` - Cognito app client ID
- `VITE_COGNITO_DOMAIN` - Cognito hosted UI domain

## Deployment Environments

**Production:**
- Frontend: CloudFront distribution
- Backend: ECS Fargate behind ALB (proxied through CloudFront)
- State: Terraform state in S3 with DynamoDB locking

## Monitoring

**Logs:**
```bash
make logs-ecs        # ECS API logs
make logs-generator  # Lambda generator logs
make logs-composer   # Lambda composer logs
```

**Health Check:**
```bash
make health          # Check API health
```

## Security

- HTTPS-only CloudFront distribution
- VPC with private subnets for ECS tasks
- IAM roles with least-privilege policies
- Secrets stored in AWS Secrets Manager
- No long-lived AWS credentials (OIDC)
- JWT authentication for API

## Cost Optimization

- S3 lifecycle policies for assets
- DynamoDB on-demand pricing
- ECS auto-scaling based on CPU
- CloudFront caching for static assets
- Lambda concurrency limits

## Contributing

1. Create a feature branch
2. Make changes
3. Create PR (workflows run automatically)
4. Merge to master (auto-deploy)

## Troubleshooting

**CI/CD Issues:**
- See [CICD_SETUP.md](./CICD_SETUP.md) troubleshooting section

**Infrastructure Issues:**
- See `infrastructure/README.md`

**OIDC Setup:**
- See `infrastructure/github-oidc/README.md`

## Project Documentation

- **Product Requirements:** `.human/PRD.md`
- **Infrastructure Guide:** `infrastructure/README.md`
- **CI/CD Setup:** `CICD_SETUP.md`
- **OIDC Setup:** `infrastructure/github-oidc/README.md`
- **Development Guidelines:** `CLAUDE.md`

## License

See LICENSE file for details.