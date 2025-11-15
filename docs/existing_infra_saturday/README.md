# OmniGen Infrastructure Documentation

> Comprehensive technical documentation for the OmniGen AI video generation pipeline

## Overview

OmniGen is a production-grade, serverless-first AI video generation pipeline built on AWS. This documentation provides detailed architectural diagrams, data flows, and technical specifications derived directly from the infrastructure code.

**Architecture Type:** Hybrid serverless (ECS Fargate + AWS Lambda)
**Cloud Provider:** AWS (us-east-1)
**Infrastructure as Code:** Terraform 1.13.5+
**Deployment:** GitHub Actions CI/CD with OIDC

## Quick Facts

- **Monthly Base Cost:** ~$100 (idle infrastructure)
- **Per-Video Cost:** ~$1.32 (30-second video)
- **Tech Stack:** Go 1.25 (API), Node.js 20 (Lambdas), React 18 (Frontend)
- **Deployment Time:** Infrastructure → Backend → Frontend in 5-15 minutes
- **High Availability:** Multi-AZ ALB, ECS auto-recovery, Lambda retry logic

---

## Documentation Index

### 🏗️ Architecture

1. **[Architecture Overview](./architecture-overview.md)**
   High-level system architecture using C4 diagrams. Shows all AWS services, their relationships, and data flow at the system context level.

2. **[Infrastructure Modules](./infrastructure-modules.md)**
   Terraform module structure and relationships. Understand how the 9 infrastructure modules work together.

### 🌐 Networking

3. **[Network Topology](./network-topology.md)**
   Detailed VPC architecture, subnets, security groups, routing tables, NAT Gateway, and VPC endpoints.

### 🔄 Data & Request Flow

4. **[Data Flow Diagrams](./data-flow.md)**
   End-to-end request/response flows from user browser through CloudFront, ALB, ECS, and backend services.

5. **[Video Generation Workflow](./video-workflow.md)**
   Step Functions state machine orchestrating Lambda functions for AI video generation.

### 🔐 Security & Authentication

6. **[Authentication Flow](./authentication-flow.md)**
   Cognito OAuth2/OIDC flow, JWT validation, httpOnly cookies, rate limiting, and quota enforcement.

### 🚀 CI/CD & Deployment

7. **[CI/CD Pipeline](./cicd-pipeline.md)**
   GitHub Actions workflows for infrastructure, backend, and frontend deployments with OIDC authentication.

### 💻 Application Architecture

8. **[Backend Architecture](./backend-architecture.md)**
   Go API internal structure, middleware stack, handlers, and AWS SDK integrations.

---

## Architecture Decision Records

### Key Technical Decisions

| Decision | Rationale | Trade-offs |
|----------|-----------|------------|
| **Hybrid ECS + Lambda** | ECS for always-on API, Lambda for bursty video processing | Complexity vs cost optimization |
| **Step Functions Express** | <5min video generation fits within Express limits | 50% cheaper, no long-term history |
| **DynamoDB On-Demand** | Unpredictable traffic for new product | Higher per-request cost, zero capacity planning |
| **Single-AZ Private Subnet** | Cost savings ($32/month NAT avoided) | Lower HA, acceptable for MVP |
| **CloudFront + ALB** | Unified origin for frontend + API (no CORS) | Dual edge/origin architecture |
| **Cognito for Auth** | Managed OAuth2/OIDC, SOC2 compliant | Vendor lock-in vs dev velocity |
| **GitHub OIDC** | No long-lived AWS credentials | Modern security vs complexity |

### Upgrade Path to Production Scale

```
MVP (Current)              Production Scale
├─ Single AZ              ├─ Multi-AZ (2-3 AZs)
├─ HTTP ALB               ├─ HTTPS with ACM certificate
├─ No WAF                 ├─ AWS WAF + Shield
├─ Manual secrets         ├─ Automated secret rotation
├─ 7-day logs             ├─ 90-day log retention
├─ No tracing             ├─ AWS X-Ray distributed tracing
└─ Deletion protection off └─ Deletion protection on
```

---

## Infrastructure Components

### Terraform Modules (9)

| Module | Purpose | Key Resources |
|--------|---------|---------------|
| `networking` | VPC, subnets, security groups, NAT, VPC endpoints | VPC, 2 public subnets, 1 private subnet, IGW, NAT, 4 SGs, 4 VPC endpoints |
| `compute` | Container orchestration | ECS cluster, ECR repository, task definition, service, auto-scaling |
| `serverless` | Video processing functions | 2 Lambdas (generator, composer), Step Functions state machine |
| `storage` | Data persistence | 2 S3 buckets (assets, frontend), 2 DynamoDB tables (jobs, usage) |
| `loadbalancer` | Traffic routing | ALB, target group, HTTP listener |
| `cdn` | Edge distribution | CloudFront distribution, OAC, cache behaviors |
| `auth` | User authentication | Cognito user pool, app client, hosted UI |
| `iam` | Access control | 4 IAM roles, 3 policies (ECS, Lambda, Step Functions) |
| `monitoring` | Observability | CloudWatch log groups, Container Insights |

### AWS Services Used (25)

<details>
<summary>Expand to see all AWS services</summary>

**Compute:**
- ECS Fargate
- Lambda
- Step Functions

**Networking:**
- VPC
- Application Load Balancer
- CloudFront
- Route 53 (future)

**Storage:**
- S3
- DynamoDB
- ECR

**Security:**
- IAM
- Cognito
- Secrets Manager
- WAF (future)

**Monitoring:**
- CloudWatch Logs
- CloudWatch Metrics
- Container Insights
- X-Ray (future)

**Developer Tools:**
- CodeBuild (future)
- CodePipeline (future)

</details>

---

## Cost Breakdown

### Monthly Operating Costs

| Service | Configuration | Monthly Cost |
|---------|--------------|--------------|
| **ECS Fargate** | 1 task, 1 vCPU, 2 GB, 24/7 | $35.04 |
| **NAT Gateway** | 1 NAT, us-east-1a | $32.40 |
| **ALB** | Internet-facing | $16.20 |
| **VPC Endpoints** | 2 interface endpoints (ECR) | $14.40 |
| **S3** | 10 GB storage | $0.23 |
| **CloudWatch** | 1 GB logs/month | $0.53 |
| **CloudFront** | 10 GB transfer | $0.85 |
| **Secrets Manager** | 1 secret | $0.40 |
| **Cognito** | <50K MAU | $0.00 |
| **Lambda** | 0 invocations idle | $0.00 |
| **DynamoDB** | On-demand, 0 requests | $0.00 |
| **TOTAL (Idle)** | | **~$100/month** |

### Per-Video Generation Cost

**30-second video:**
- Infrastructure: $0.02 (Lambda, Step Functions, DynamoDB, S3)
- Replicate API: $1.30 (AI model calls)
- **Total: $1.32/video** (well under $2.00 target)

---

## Quick Start

### Prerequisites
- AWS Account with admin access
- GitHub repository
- Terraform 1.13.5+
- Go 1.25+ (for local backend development)
- Bun (for local frontend development)

### Deployment Steps

1. **Setup AWS OIDC** (one-time)
   ```bash
   cd infrastructure/github-oidc
   terraform init
   terraform apply
   ```

2. **Configure GitHub Secrets**
   - `AWS_ROLE_ARN`: Output from OIDC setup
   - `REPLICATE_API_KEY_SECRET_ARN`: ARN of secret in Secrets Manager

3. **Deploy Infrastructure**
   ```bash
   git push origin master  # Triggers infrastructure.yml workflow
   ```

4. **Deploy Backend**
   ```bash
   # Auto-deploys after infrastructure via workflow_run trigger
   ```

5. **Deploy Frontend**
   ```bash
   # Auto-deploys after infrastructure via workflow_run trigger
   ```

---

## Monitoring & Operations

### Health Checks

- **Frontend:** https://{cloudfront-domain}
- **API:** https://{cloudfront-domain}/api/v1/health
- **ECS Service:** AWS Console → ECS → omnigen cluster

### Key Metrics

- **ECS:** CPU, memory, running task count
- **Lambda:** Invocations, duration, errors, throttles
- **Step Functions:** Executions succeeded/failed
- **DynamoDB:** Consumed capacity, throttles
- **ALB:** Request count, 5xx errors, target response time

### Logs

- **API Logs:** CloudWatch → /ecs/omnigen
- **Generator Lambda:** CloudWatch → /aws/lambda/omnigen-generator
- **Composer Lambda:** CloudWatch → /aws/lambda/omnigen-composer
- **Step Functions:** CloudWatch → /aws/states/omnigen-workflow

---

## Support & Maintenance

### Common Operations

- **Scale ECS:** Update `ecs_task_count` variable in terraform.tfvars
- **Update Docker Image:** Push to master, GitHub Actions auto-deploys
- **Invalidate CloudFront:** AWS Console or CLI
- **View Logs:** CloudWatch Logs Insights

### Troubleshooting

See individual documentation pages for detailed troubleshooting guides.

---

## Version Information

- **Terraform:** 1.13.5
- **AWS Provider:** 6.21
- **Go:** 1.25.4
- **Node.js:** 20.x
- **Bun:** latest
- **Mermaid (diagrams):** 11.0.0

---

**Last Updated:** December 2024
**Maintained By:** OmniGen Team
**Infrastructure Code:** `/infrastructure`
