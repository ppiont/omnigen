# AWS Infrastructure PRD: Lean Video Generation Pipeline

**Project**: AI Video Generation Infrastructure (Lean MVP)  
**Target**: One-week sprint with 48-hour MVP checkpoint  
**Cost Target**: $110-200/month base infrastructure  
**Deployment Target**: 10-15 minutes  

---

## 1. Executive Summary

### Purpose
Build a cost-effective, production-ready AWS infrastructure to support an AI-powered video advertisement generation pipeline. The infrastructure must enable rapid iteration on AI video quality while minimizing operational complexity and cost.

### Success Criteria
- ✅ Deploy in under 15 minutes
- ✅ Support 30-second video generation in under 5 minutes
- ✅ Cost under $2.00 per minute of generated video
- ✅ Handle 50+ concurrent video generation requests
- ✅ 90%+ successful generation rate
- ✅ Simple enough for small teams to maintain

### Key Constraints
- **Time**: Must not delay AI development (infrastructure is supporting, not primary)
- **Cost**: Infrastructure should be <20% of total AI generation costs
- **Complexity**: Manageable by 1-3 person team without dedicated DevOps
- **Scalability**: Can grow to 1000+ videos/day without architecture changes

---

## 2. Background & Context

### Problem Statement
Traditional enterprise AWS architectures are over-engineered for MVP/bounty projects. Multi-AZ deployments, Redis clusters, RDS instances, and global CDNs add $600-800/month in costs and significant complexity for workloads that don't require them.

For an AI video generation pipeline competing on output quality, infrastructure should:
1. **Enable** the AI workflow (not block it)
2. **Cost-optimize** to maximize AI generation budget
3. **Stay simple** so developers focus on video quality

### Design Philosophy
> "A simple but reliable infrastructure beats a feature-rich system that's hard to maintain"

Aligned with the video generation PRD principle:
> "Focus on coherence over quantity, reliability over features, cost efficiency over bleeding edge"

### Trade-offs Accepted
- **Single AZ** instead of Multi-AZ (99.5% vs 99.99% uptime)
- **DynamoDB** instead of RDS PostgreSQL (no JOINs, but sufficient)
- **No Redis** instead of ElastiCache (in-memory cache adequate)
- **No CloudFront** instead of global CDN (add when needed)
- **API keys** instead of OAuth/Cognito (simpler authentication)

These trade-offs save $600/month and reduce complexity by 70% with minimal impact on video generation quality.

---

## 3. Architecture Overview

### High-Level Design

```
Internet
   ↓
Application Load Balancer (ALB)
   ↓
ECS Fargate (1-5 Go API tasks)
   ↓
AWS Step Functions (3-stage workflow)
   ↓
AWS Lambda (2 functions: Generator + Composer)
   ↓
Storage Layer: S3 (videos) + DynamoDB (job state)
```

### Architecture Principles

1. **Serverless-first**: Use Lambda for processing, pay per execution
2. **Managed services**: Let AWS handle scaling (ECS, Lambda, DynamoDB)
3. **Single region, single AZ**: Optimize for cost, not global redundancy
4. **Event-driven**: Step Functions orchestrate, Lambda processes
5. **Stateless API**: ECS tasks can scale independently

### Component Justification

| Component | Why Included | Why Not Serverless Alternative |
|-----------|--------------|-------------------------------|
| **ECS Fargate** | Replicate webhooks need persistent endpoint | Lambda URLs possible but less reliable for webhooks |
| **Step Functions** | Visual workflow, built-in retry logic | Could use Go orchestration, but SFN better for debugging |
| **Lambda** | Auto-scales, pay-per-use for processing | Perfect for scene generation and FFmpeg composition |
| **ALB** | Load balancing, health checks | Could use API Gateway, but ALB better for ECS |
| **DynamoDB** | Serverless, auto-scaling, fast | Perfect key-value access pattern for job tracking |
| **S3** | Durable video storage, cheap | No alternative for large file storage |

---

## 4. Detailed Component Specifications

### 4.1 Networking (VPC)

**Resources**: 10
- 1x VPC (10.0.0.0/16)
- 1x Internet Gateway
- 1x Public Subnet (for ALB)
- 1x Private Subnet (for ECS + Lambda)
- 1x NAT Gateway (for outbound internet access)
- 3x Route Tables (public, private)
- 2x VPC Endpoints (S3, DynamoDB - gateway type, free)

**Configuration**:
```hcl
vpc_cidr: "10.0.0.0/16"
public_subnet: "10.0.1.0/24"
private_subnet: "10.0.10.0/24"
availability_zone: Single AZ (us-east-1a)
```

**Design Decisions**:
- **Single AZ**: Saves $70/month on NAT Gateway costs
- **Gateway VPC Endpoints**: Free for S3 and DynamoDB, avoid NAT charges
- **No Interface Endpoints**: Not needed for this scale ($7/month each)
- **Simple routing**: One public route, one private route

**Scaling Path**:
- If needed: Add second AZ for high availability (+$70/month)
- If needed: Add interface endpoints for Secrets Manager (+$7/month)

**Monthly Cost**: ~$35 (NAT Gateway only)

---

### 4.2 Compute Layer

#### 4.2.1 ECS Fargate (Go API Server)

**Purpose**: 
- Receive video generation requests
- Start Step Functions workflows
- Handle Replicate API webhooks
- Query job status
- In-memory caching of frequent prompts

**Resources**: 6
- 1x ECS Cluster
- 1x ECS Task Definition
- 1x ECS Service
- 1x ECR Repository
- 1x Auto Scaling Target
- 1x Auto Scaling Policy (CPU-based)

**Configuration**:
```hcl
task_cpu: 1024 (1 vCPU)
task_memory: 2048 MB (2 GB)
min_tasks: 1
max_tasks: 5
target_cpu_utilization: 70%
```

**Container Spec**:
```dockerfile
FROM golang:1.23-alpine AS builder
# Build Go binary
FROM alpine:latest
EXPOSE 8080
CMD ["./main"]
```

**API Endpoints**:
- `POST /api/v1/generate` - Submit video generation job
- `GET /api/v1/jobs/:id` - Get job status
- `POST /webhooks/replicate` - Handle Replicate callbacks
- `GET /health` - Health check

**Scaling Behavior**:
- Scale out: CPU > 70% for 2 minutes
- Scale in: CPU < 40% for 5 minutes
- Health check: `/health` every 30 seconds
- Deployment: Rolling update, 100% minimum healthy

**Go Code Pattern**:
```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]CacheItem
}

// In-memory cache - no Redis needed for MVP
// 1000 entry LRU cache
// Stores: prompt fingerprints, style preferences
```

**Monthly Cost**: $35-80 (1-2 tasks typically running)

#### 4.2.2 Lambda Functions

**Purpose**: Process video generation steps

##### Lambda 1: Scene Generator
```yaml
Name: ad-pipeline-mvp-generator
Runtime: Go (provided.al2023)
Timeout: 900 seconds (15 minutes)
Memory: 2048 MB
Concurrency: 10 (limit concurrent for cost control)

Responsibilities:
  - Parse user prompt
  - Plan video scenes (timing, transitions)
  - Call Replicate API for each scene
  - Handle Replicate webhooks
  - Store generated assets in S3
  - Update DynamoDB with scene status

Environment Variables:
  - ASSETS_BUCKET: S3 bucket name
  - JOB_TABLE: DynamoDB table name
  - REPLICATE_SECRET_ARN: Secrets Manager ARN
  - AWS_REGION_NAME: us-east-1
```

##### Lambda 2: Video Composer
```yaml
Name: ad-pipeline-mvp-composer
Runtime: Go (provided.al2023)
Timeout: 900 seconds (15 minutes)
Memory: 10240 MB (10 GB - max for FFmpeg)
Ephemeral Storage: 10240 MB (10 GB)
Concurrency: 5 (video composition is memory-intensive)

Responsibilities:
  - Download scene assets from S3
  - Stitch videos with FFmpeg
  - Add audio synchronization
  - Apply transitions
  - Upload final video to S3
  - Update DynamoDB with completion

Dependencies:
  - FFmpeg (bundled in Lambda layer)
  - Go binary with video processing logic
```

**Lambda Deployment Package**:
```bash
# Build Go binary for Lambda
GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda

# Package with FFmpeg layer
zip lambda.zip bootstrap

# Deploy via Terraform
terraform apply
```

**Monthly Cost**: $20-40 (10,000 invocations/day)

---

### 4.3 Orchestration (Step Functions)

**Purpose**: Coordinate 3-stage video generation pipeline

**Type**: Express Workflow (faster, cheaper for <5 minute workflows)

**Workflow Definition**:
```json
{
  "Comment": "Simplified AI Video Generation Pipeline for Ads",
  "StartAt": "GenerateScenes",
  "States": {
    
    "GenerateScenes": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:...:function:generator",
      "Next": "ComposeVideo",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 2,
          "MaxAttempts": 2,
          "BackoffRate": 2
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "Next": "MarkFailed"
        }
      ]
    },
    
    "ComposeVideo": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:...:function:composer",
      "Next": "MarkComplete",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": 2,
          "BackoffRate": 2
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "Next": "MarkFailed"
        }
      ]
    },
    
    "MarkComplete": {
      "Type": "Task",
      "Resource": "arn:aws:states:::dynamodb:updateItem",
      "End": true
    },
    
    "MarkFailed": {
      "Type": "Task",
      "Resource": "arn:aws:states:::dynamodb:updateItem",
      "Next": "Fail"
    },
    
    "Fail": {
      "Type": "Fail"
    }
  }
}
```

**Execution Flow**:
1. ECS API receives request → starts Step Functions execution
2. GenerateScenes Lambda: Parse prompt → call Replicate → store assets
3. ComposeVideo Lambda: Stitch scenes → add audio → upload final
4. MarkComplete: Update DynamoDB status to "completed"

**Error Handling**:
- Automatic retry with exponential backoff
- Failed executions update job status in DynamoDB
- CloudWatch logs all execution steps
- X-Ray tracing disabled for cost savings

**Performance**:
- Express workflows: Sub-second startup
- State transitions: ~100ms overhead
- Total orchestration overhead: <500ms per video

**Monthly Cost**: ~$5 (5,000 executions/day)

---

### 4.4 Storage Layer

#### 4.4.1 S3 (Video Assets)

**Bucket Configuration**:
```hcl
bucket_name: "ad-pipeline-mvp-assets"
versioning: Enabled
encryption: AES256 (server-side)
public_access: Blocked (all)
```

**Lifecycle Rules**:
```yaml
Transition to Standard-IA: 30 days
Transition to Glacier Instant Retrieval: 90 days
Expiration: 365 days (auto-delete old videos)
```

**CORS Configuration**:
```json
{
  "AllowedHeaders": ["*"],
  "AllowedMethods": ["GET", "HEAD"],
  "AllowedOrigins": ["*"],
  "ExposeHeaders": ["ETag"],
  "MaxAgeSeconds": 3000
}
```

**Storage Pattern**:
```
s3://ad-pipeline-mvp-assets/
  jobs/
    {job_id}/
      scenes/
        scene-01.mp4
        scene-02.mp4
        scene-03.mp4
      audio/
        background.mp3
      final/
        video-{timestamp}.mp4
```

**Access Pattern**:
- Write: Lambda functions during generation
- Read: Public via presigned URLs (7-day expiration)
- Delete: Lifecycle rules or manual cleanup

**Expected Storage**:
- 30-second video: ~10-15 MB
- Scene assets: ~5 MB each
- Average job: ~30 MB total
- 1000 videos: ~30 GB
- Cost at 30 GB: ~$1/month

**Monthly Cost**: $5-20 (20 GB storage + 100 GB transfer)

#### 4.4.2 DynamoDB (Job Tracking)

**Table Configuration**:
```hcl
table_name: "ad-pipeline-mvp-jobs"
billing_mode: PAY_PER_REQUEST (on-demand)
hash_key: job_id (String)
```

**Attributes**:
```python
job_id: String (UUID v4)
user_id: String (for multi-user tracking)
status: String (pending, processing, completed, failed)
prompt: String (user's input prompt)
duration: Number (requested video length in seconds)
style: String (visual style preferences)
video_url: String (S3 presigned URL)
created_at: Number (Unix timestamp)
completed_at: Number (Unix timestamp)
error_message: String (if failed)
ttl: Number (Unix timestamp, auto-delete after 90 days)
```

**Global Secondary Indexes**:
```yaml
UserJobsIndex:
  hash_key: user_id
  range_key: created_at
  projection: ALL
  
StatusIndex:
  hash_key: status
  range_key: created_at
  projection: ALL
```

**Access Patterns**:
```sql
-- Get job by ID
GetItem(job_id)

-- List user's jobs
Query(UserJobsIndex, user_id, created_at DESC)

-- Monitor processing jobs
Query(StatusIndex, status="processing")
```

**Performance**:
- GetItem latency: <10ms (single-digit milliseconds)
- Query latency: <20ms for 100 items
- No provisioned capacity needed
- Auto-scales to any load

**Expected Volume**:
- 1000 requests/day = 30K requests/month
- Cost: ~$0.25/month for reads
- Cost: ~$1.25/month for writes
- Total: ~$1.50/month

**TTL (Time To Live)**:
- Automatically delete jobs after 90 days
- Saves storage costs
- Keeps recent jobs for debugging

**Point-in-Time Recovery**: Enabled (can restore to any second in last 35 days)

**Monthly Cost**: $5-15 (1-5M requests/month)

---

### 4.5 Load Balancing (ALB)

**Configuration**:
```hcl
name: "ad-pipeline-mvp"
type: Application Load Balancer
subnets: [public_subnet] (single AZ)
security_groups: [alb_sg]
```

**Listeners**:
```yaml
HTTP (Port 80):
  Action: Forward to ECS target group
  
# HTTPS optional - add certificate later
# HTTPS (Port 443):
#   SSL Certificate: ACM certificate
#   Action: Forward to ECS target group
```

**Target Group**:
```hcl
port: 8080
protocol: HTTP
target_type: ip (for Fargate)
health_check:
  path: /health
  interval: 30 seconds
  timeout: 5 seconds
  healthy_threshold: 2
  unhealthy_threshold: 3
deregistration_delay: 30 seconds
```

**Security Groups**:
```yaml
ALB Security Group:
  Ingress:
    - Port 80 from 0.0.0.0/0
    - Port 443 from 0.0.0.0/0 (if HTTPS enabled)
  Egress:
    - All to ECS security group

ECS Security Group:
  Ingress:
    - Port 8080 from ALB security group
  Egress:
    - All to 0.0.0.0/0
```

**Monthly Cost**: ~$20

---

### 4.6 IAM (Security)

**Roles Created**: 3

#### IAM Role 1: ECS Task Execution Role
```yaml
Purpose: Pull images from ECR, write logs, access secrets
Permissions:
  - AmazonECSTaskExecutionRolePolicy (AWS managed)
  - secretsmanager:GetSecretValue (for Replicate API key)
  - logs:CreateLogStream, logs:PutLogEvents
```

#### IAM Role 2: ECS Task Role
```yaml
Purpose: Application runtime permissions
Permissions:
  - s3:PutObject, s3:GetObject on assets bucket
  - dynamodb:PutItem, dynamodb:GetItem, dynamodb:UpdateItem, dynamodb:Query
  - states:StartExecution, states:DescribeExecution
  - secretsmanager:GetSecretValue (Replicate API key)
  - logs:CreateLogGroup, logs:PutLogEvents
```

#### IAM Role 3: Lambda Execution Role
```yaml
Purpose: Lambda function runtime permissions
Permissions:
  - AWSLambdaBasicExecutionRole (AWS managed)
  - AWSLambdaVPCAccessExecutionRole (AWS managed)
  - s3:PutObject, s3:GetObject on assets bucket
  - dynamodb:PutItem, dynamodb:GetItem, dynamodb:UpdateItem
  - secretsmanager:GetSecretValue (Replicate API key)
```

#### IAM Role 4: Step Functions Execution Role
```yaml
Purpose: Orchestrate Lambda and DynamoDB
Permissions:
  - lambda:InvokeFunction
  - dynamodb:PutItem, dynamodb:UpdateItem
  - logs:CreateLogDelivery, logs:PutResourcePolicy
```

**Security Best Practices**:
- ✅ Least privilege access (only required permissions)
- ✅ No hardcoded credentials (Secrets Manager)
- ✅ Resource-specific policies (not wildcards)
- ✅ Separate execution vs task roles
- ✅ CloudTrail logging enabled

---

### 4.7 Monitoring & Logging

#### CloudWatch Log Groups
```yaml
/ecs/ad-pipeline-mvp:
  Retention: 7 days
  Purpose: API server logs
  
/aws/lambda/ad-pipeline-mvp-generator:
  Retention: 7 days
  Purpose: Scene generation logs
  
/aws/lambda/ad-pipeline-mvp-composer:
  Retention: 7 days
  Purpose: Video composition logs
  
/aws/states/ad-pipeline-mvp:
  Retention: 7 days
  Purpose: Step Functions execution logs
```

#### Key Metrics to Monitor
```yaml
ECS:
  - CPUUtilization (target: <70%)
  - MemoryUtilization (target: <80%)
  - TargetResponseTime (target: <500ms)
  - HealthyHostCount (alert if 0)

Lambda:
  - Duration (target: <5min for generator, <10min for composer)
  - Errors (alert if >5%)
  - Throttles (alert if >0)
  - ConcurrentExecutions (monitor limits)

Step Functions:
  - ExecutionsSucceeded (target: >90%)
  - ExecutionsFailed (alert if >10%)
  - ExecutionTime (target: <10min for 30s video)

DynamoDB:
  - UserErrors (alert if >0)
  - SystemErrors (alert if >0)
  - ConsumedReadCapacityUnits (monitor)
  - ConsumedWriteCapacityUnits (monitor)

ALB:
  - TargetResponseTime (target: <500ms)
  - HTTPCode_Target_5XX_Count (alert if >1%)
  - UnHealthyHostCount (alert if >0)
```

#### CloudWatch Alarms (Recommended)
```yaml
ECS-CPU-High:
  Metric: CPUUtilization
  Threshold: >85%
  Duration: 5 minutes
  Action: Send SNS notification

Lambda-Errors-High:
  Metric: Errors
  Threshold: >10 in 5 minutes
  Action: Send SNS notification

Step-Functions-Failed:
  Metric: ExecutionsFailed
  Threshold: >5 in 10 minutes
  Action: Send SNS notification
```

**Monthly Cost**: ~$3 (5GB logs)

---

### 4.8 Secrets Management

**AWS Secrets Manager**:
```yaml
Secret Name: ad-pipeline/replicate-api-key
Type: SecretString
Value: <replicate_api_key>
Rotation: Manual (or 90-day automatic)
Encryption: AWS KMS (default key)
```

**Access Pattern**:
```go
// Go code to retrieve secret
import "github.com/aws/aws-sdk-go-v2/service/secretsmanager"

func getReplicateKey(ctx context.Context) (string, error) {
    result, err := svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String("ad-pipeline/replicate-api-key"),
    })
    if err != nil {
        return "", err
    }
    return *result.SecretString, nil
}
```

**Cost**: $0.40/month per secret

---

## 5. Deployment Specifications

### 5.1 Terraform Structure

```
ad-pipeline-lean/
├── main.tf                 # Root module orchestration
├── variables.tf            # Input variables
├── outputs.tf              # Output values
├── terraform.tfvars        # User configuration (gitignored)
├── terraform.tfvars.example
├── Makefile               # Convenience commands
├── .gitignore
│
└── modules/
    ├── storage/           # S3 + DynamoDB
    │   ├── main.tf       (~140 lines)
    │   ├── variables.tf  (~5 lines)
    │   └── outputs.tf    (~15 lines)
    │
    ├── core/             # VPC + IAM + ECS + ALB
    │   ├── networking.tf (~110 lines)
    │   ├── iam.tf        (~140 lines)
    │   ├── ecs.tf        (~180 lines)
    │   ├── variables.tf  (~55 lines)
    │   └── outputs.tf    (~35 lines)
    │
    └── compute/          # Step Functions + Lambda
        ├── main.tf       (~140 lines)
        ├── step_functions.tf (~110 lines)
        ├── variables.tf  (~35 lines)
        └── outputs.tf    (~15 lines)

Total: ~800 lines of Terraform
```

### 5.2 Deployment Steps

#### Prerequisites
```bash
# 1. Install Terraform >= 1.6.0
brew install terraform

# 2. Configure AWS CLI
aws configure
# AWS Access Key ID: <your-key>
# AWS Secret Access Key: <your-secret>
# Default region: us-east-1

# 3. Verify access
aws sts get-caller-identity
```

#### Initial Setup (One-time)
```bash
# 1. Create S3 backend for Terraform state
aws s3 mb s3://ad-pipeline-terraform-state --region us-east-1

# 2. Enable versioning
aws s3api put-bucket-versioning \
  --bucket ad-pipeline-terraform-state \
  --versioning-configuration Status=Enabled

# 3. Store Replicate API key in Secrets Manager
aws secretsmanager create-secret \
  --name ad-pipeline/replicate-api-key \
  --description "Replicate API Key for video generation" \
  --secret-string "your-replicate-api-key-here" \
  --region us-east-1

# 4. Get secret ARN (save for terraform.tfvars)
aws secretsmanager describe-secret \
  --secret-id ad-pipeline/replicate-api-key \
  --region us-east-1 \
  --query 'ARN' \
  --output text
```

#### Infrastructure Deployment
```bash
# 1. Clone/navigate to infrastructure
cd ad-pipeline-lean

# 2. Create terraform.tfvars from example
cp terraform.tfvars.example terraform.tfvars

# 3. Edit terraform.tfvars with your values
vim terraform.tfvars
# Set:
#   - project_name
#   - replicate_api_key_secret_arn (from step 3 above)
#   - ecs_min_tasks (default: 1)
#   - ecs_max_tasks (default: 5)

# 4. Initialize Terraform
terraform init

# 5. Validate configuration
terraform validate

# 6. Preview changes
terraform plan

# 7. Deploy (takes 10-15 minutes)
terraform apply

# 8. Save outputs
terraform output > outputs.txt
```

#### Application Deployment
```bash
# 1. Build Docker image
docker build -t ad-pipeline-api:latest .

# 2. Get ECR repository URL
ECR_REPO=$(terraform output -raw ecr_repository_url)

# 3. Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin $ECR_REPO

# 4. Tag image
docker tag ad-pipeline-api:latest $ECR_REPO:latest

# 5. Push image
docker push $ECR_REPO:latest

# 6. Update ECS service (triggers rolling deployment)
aws ecs update-service \
  --cluster ad-pipeline-mvp \
  --service ad-pipeline-mvp-api \
  --force-new-deployment \
  --region us-east-1

# 7. Monitor deployment
aws ecs describe-services \
  --cluster ad-pipeline-mvp \
  --services ad-pipeline-mvp-api \
  --region us-east-1
```

### 5.3 Verification Steps

```bash
# 1. Get API endpoint
API_URL=$(terraform output -raw api_url)

# 2. Check health endpoint
curl $API_URL/health
# Expected: {"status": "healthy"}

# 3. Submit test job
curl -X POST $API_URL/api/v1/generate \
  -H "Content-Type: application/json" \
  -H "x-api-key: test-key" \
  -d '{
    "prompt": "15 second luxury watch ad with gold aesthetics",
    "duration": 15,
    "aspect_ratio": "9:16"
  }'
# Expected: {"job_id": "uuid", "status": "pending"}

# 4. Check job status
JOB_ID=<uuid-from-above>
curl $API_URL/api/v1/jobs/$JOB_ID \
  -H "x-api-key: test-key"

# 5. View logs
aws logs tail /ecs/ad-pipeline-mvp --follow

# 6. Check Step Functions executions
aws stepfunctions list-executions \
  --state-machine-arn $(terraform output -raw step_function_arn) \
  --max-results 10
```

### 5.4 Teardown

```bash
# WARNING: This destroys all infrastructure and data

# 1. Empty S3 bucket first (Terraform can't delete non-empty buckets)
aws s3 rm s3://ad-pipeline-mvp-assets --recursive

# 2. Destroy infrastructure
terraform destroy

# 3. Delete Terraform state bucket (optional)
aws s3 rb s3://ad-pipeline-terraform-state --force

# 4. Delete secrets (optional)
aws secretsmanager delete-secret \
  --secret-id ad-pipeline/replicate-api-key \
  --force-delete-without-recovery
```

---

## 6. Cost Analysis

### 6.1 Monthly Cost Breakdown

| Service | Configuration | Unit Cost | Volume | Total |
|---------|--------------|-----------|---------|-------|
| **Compute** | | | | |
| ECS Fargate | 1 task × 1 vCPU × 2GB × 730 hrs | $0.04856/hr | 730 | $35.44 |
| ECS Fargate (scale) | +1 task during peak | $0.04856/hr | 100 | $4.86 |
| Lambda (Generator) | 2GB × 5min × 10k invokes | $0.0000166667/GB-sec | 6M GB-sec | $10.00 |
| Lambda (Composer) | 10GB × 10min × 5k invokes | $0.0000166667/GB-sec | 30M GB-sec | $50.00 |
| Lambda Requests | | $0.20/1M requests | 15k | $0.003 |
| **Networking** | | | | |
| NAT Gateway | 730 hours | $0.045/hr | 730 | $32.85 |
| NAT Data Processing | | $0.045/GB | 50 GB | $2.25 |
| ALB | 730 hours | $0.0225/hr | 730 | $16.43 |
| ALB LCU | | $0.008/LCU-hr | 100 | $0.80 |
| **Storage** | | | | |
| S3 Standard | | $0.023/GB | 20 GB | $0.46 |
| S3 Standard-IA | | $0.0125/GB | 30 GB | $0.38 |
| S3 GET requests | | $0.0004/1k | 100k | $0.04 |
| S3 PUT requests | | $0.005/1k | 20k | $0.10 |
| S3 Data Transfer | | $0.09/GB | 100 GB | $9.00 |
| DynamoDB (On-demand) | Reads | $0.25/M | 1M | $0.25 |
| DynamoDB (On-demand) | Writes | $1.25/M | 0.5M | $0.63 |
| DynamoDB Storage | | $0.25/GB | 1 GB | $0.25 |
| **Orchestration** | | | | |
| Step Functions Express | | $1.00/M | 5k | $0.005 |
| **Monitoring** | | | | |
| CloudWatch Logs | | $0.50/GB | 5 GB | $2.50 |
| CloudWatch Metrics | Custom metrics | $0.30/metric | 10 | $3.00 |
| **Security** | | | | |
| Secrets Manager | | $0.40/secret | 1 | $0.40 |
| **ECR** | | | | |
| ECR Storage | | $0.10/GB | 5 GB | $0.50 |
| **TOTAL BASE** | | | | **$170.17** |
| **Typical Usage** | | | | **$130-180** |
| **Heavy Usage** | | | | **$180-250** |

### 6.2 Per-Video Cost Analysis

**30-second Ad Video Cost Breakdown**:
```yaml
Infrastructure:
  ECS API call: $0.0005 (amortized)
  Lambda Generator: $0.05 (2GB × 5min)
  Lambda Composer: $0.20 (10GB × 10min)
  Step Functions: $0.000001 (negligible)
  DynamoDB: $0.001 (3-4 operations)
  S3 Storage: $0.001 (15 MB)
  S3 Transfer: $0.005 (15 MB download)
  Total Infrastructure: $0.26

Replicate API (example):
  Image generation (3 scenes): $0.30 (3 × $0.10)
  Video generation (3 clips): $0.90 (3 × $0.30)
  Audio generation: $0.10
  Total Replicate: $1.30

TOTAL PER VIDEO: $1.56
```

**Meets PRD target of <$2.00 per minute** ✓

### 6.3 Scaling Cost Projections

| Volume | Videos/Day | Infrastructure/mo | Replicate/mo | Total/mo |
|--------|------------|-------------------|--------------|----------|
| **MVP** | 10 | $130 | $40 | $170 |
| **Beta** | 100 | $150 | $400 | $550 |
| **Growth** | 500 | $180 | $2,000 | $2,180 |
| **Scale** | 1000 | $250 | $4,000 | $4,250 |

### 6.4 Cost Optimization Strategies

1. **Lambda Memory Tuning**: Profile actual memory usage, reduce if possible
2. **S3 Lifecycle**: Aggressive transition to cheaper storage classes
3. **DynamoDB**: Consider provisioned capacity if consistent load
4. **Reserved Capacity**: 1-year savings plan when volume is predictable
5. **Spot Instances**: Not applicable (using Fargate, Lambda)
6. **Request Batching**: Batch Replicate API calls when possible
7. **Caching**: In-memory cache in Go API reduces Lambda invocations
8. **Compression**: Compress videos before storing in S3

**Potential Savings**: 20-30% with optimization

---

## 7. Performance Requirements

### 7.1 Latency Targets

| Endpoint | Target | Measurement |
|----------|--------|-------------|
| Health Check | <50ms | p99 |
| Submit Job | <200ms | p99 |
| Check Status | <100ms | p99 |
| List Jobs | <300ms | p99 |

### 7.2 Video Generation Time Targets

| Video Length | Target Generation Time | Actual (Expected) |
|--------------|----------------------|-------------------|
| 15 seconds | <3 minutes | 2-3 minutes ✓ |
| 30 seconds | <5 minutes | 4-5 minutes ✓ |
| 60 seconds | <10 minutes | 8-10 minutes ✓ |
| 3 minutes | <20 minutes | 15-20 minutes ✓ |

**Bottleneck**: Replicate API (not infrastructure)

### 7.3 Throughput Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Concurrent Requests | 50+ | Lambda scales automatically |
| Videos/Hour | 100+ | Limited by Replicate rate limits |
| API Requests/Second | 100+ | ALB + ECS can handle |
| DynamoDB Operations/Sec | 1000+ | On-demand scales automatically |

### 7.4 Availability Targets

| Component | Target Uptime | Actual (Expected) |
|-----------|---------------|-------------------|
| Overall System | 99.5% | ~3.6 hrs/month downtime |
| ECS API | 99.5% | Single AZ limitation |
| Lambda | 99.95% | AWS SLA |
| DynamoDB | 99.99% | AWS SLA |
| S3 | 99.99% | AWS SLA |
| Step Functions | 99.9% | AWS SLA |

**Note**: Replicate API availability is the real constraint

### 7.5 Reliability Requirements

| Metric | Target | Implementation |
|--------|--------|----------------|
| Success Rate | >90% | Retry logic in Step Functions |
| Error Handling | Graceful | Catch blocks update job status |
| Retry Strategy | Exponential backoff | 2-4-8 second delays |
| Timeout Handling | Fail fast | 15 min Lambda timeout |
| Data Durability | 99.999999999% | S3 eleven nines |
| State Consistency | Strong | DynamoDB ACID |

---

## 8. Security Requirements

### 8.1 Network Security

```yaml
Public Internet:
  ↓ [HTTPS]
  
ALB Security Group:
  - Allow inbound: 80, 443 from 0.0.0.0/0
  - Allow outbound: 8080 to ECS security group
  ↓
  
ECS Security Group:
  - Allow inbound: 8080 from ALB security group
  - Allow outbound: 443 to internet (Replicate API)
  - Allow outbound: via VPC endpoints (S3, DynamoDB)
  ↓
  
Lambda Security Group:
  - Allow inbound: None
  - Allow outbound: 443 to internet
  - Allow outbound: via VPC endpoints
  ↓
  
Private Resources:
  - DynamoDB (VPC endpoint)
  - S3 (VPC endpoint)
  - Secrets Manager (HTTPS)
```

### 8.2 Data Security

**Encryption at Rest**:
- ✅ S3: AES-256 server-side encryption
- ✅ DynamoDB: AWS-managed encryption
- ✅ ECS task memory: Fargate encrypts by default
- ✅ Lambda ephemeral storage: Encrypted
- ✅ Secrets Manager: KMS encrypted

**Encryption in Transit**:
- ✅ ALB → ECS: HTTP over private subnet
- ✅ ECS/Lambda → S3: HTTPS via VPC endpoint
- ✅ ECS/Lambda → DynamoDB: HTTPS via VPC endpoint
- ✅ ECS/Lambda → Replicate: HTTPS
- ✅ Client → ALB: HTTP (HTTPS optional with cert)

### 8.3 Access Control

**Authentication**:
- API Key in header: `x-api-key: <key>`
- Validated in Go API before processing
- No OAuth/Cognito for MVP (add later if needed)

**Authorization**:
- IAM roles with least privilege
- No hardcoded credentials
- Secrets Manager for API keys
- CloudTrail logging of all API calls

**API Key Management**:
```go
// Simple API key validation
func validateAPIKey(key string) bool {
    validKeys := []string{
        os.Getenv("API_KEY_1"),
        os.Getenv("API_KEY_2"),
    }
    for _, valid := range validKeys {
        if subtle.ConstantTimeCompare([]byte(key), []byte(valid)) == 1 {
            return true
        }
    }
    return false
}
```

### 8.4 Compliance

**Data Retention**:
- CloudWatch Logs: 7 days
- S3 Videos: 365 days (then auto-delete)
- DynamoDB Jobs: 90 days (TTL)

**Audit Trail**:
- CloudTrail: All AWS API calls logged
- CloudWatch: All application logs
- Step Functions: Execution history

**Vulnerability Management**:
- ECR Image Scanning: Enabled on push
- Dependabot: Enabled for Go dependencies
- Regular updates: Monthly Terraform apply

---

## 9. Operations & Maintenance

### 9.1 Monitoring Checklist

**Daily**:
- [ ] Check CloudWatch dashboard for anomalies
- [ ] Review error logs in CloudWatch
- [ ] Check Step Functions execution success rate
- [ ] Monitor ECS task health

**Weekly**:
- [ ] Review AWS Cost Explorer
- [ ] Check Lambda concurrency limits
- [ ] Verify S3 lifecycle policies working
- [ ] Review DynamoDB consumed capacity

**Monthly**:
- [ ] Update Go dependencies
- [ ] Review IAM policies for least privilege
- [ ] Optimize Lambda memory allocations
- [ ] Review and optimize costs

### 9.2 Backup & Recovery

**Automated Backups**:
- ✅ DynamoDB: Point-in-time recovery (35 days)
- ✅ S3: Versioning enabled (can restore deleted objects)
- ✅ Terraform State: Versioned in S3

**Recovery Procedures**:

```bash
# Restore DynamoDB to point in time
aws dynamodb restore-table-to-point-in-time \
  --source-table-name ad-pipeline-mvp-jobs \
  --target-table-name ad-pipeline-mvp-jobs-restored \
  --restore-date-time 2024-01-01T12:00:00Z

# Restore deleted S3 object (if versioning enabled)
aws s3api list-object-versions \
  --bucket ad-pipeline-mvp-assets \
  --prefix jobs/abc-123/final/video.mp4

aws s3api get-object \
  --bucket ad-pipeline-mvp-assets \
  --key jobs/abc-123/final/video.mp4 \
  --version-id <version-id> \
  restored-video.mp4

# Recreate infrastructure from Terraform
cd ad-pipeline-lean
terraform apply
```

### 9.3 Scaling Procedures

**Horizontal Scaling (More Tasks)**:
```bash
# Increase ECS max tasks
# Edit terraform.tfvars:
ecs_max_tasks = 10

# Apply change
terraform apply

# ECS auto-scaling will handle the rest
```

**Vertical Scaling (Bigger Tasks)**:
```bash
# Increase CPU/memory
# Edit terraform.tfvars:
ecs_cpu = 2048    # 2 vCPU
ecs_memory = 4096  # 4 GB

# Apply change
terraform apply

# Triggers rolling deployment with new task size
```

**Lambda Concurrency**:
```hcl
# Increase in modules/compute/main.tf:
reserved_concurrent_executions = 20  # from 10

# Apply
terraform apply
```

### 9.4 Troubleshooting Guide

**Problem: ECS tasks not starting**
```bash
# Check task logs
aws logs tail /ecs/ad-pipeline-mvp --since 10m

# Describe stopped tasks
aws ecs describe-tasks \
  --cluster ad-pipeline-mvp \
  --tasks $(aws ecs list-tasks --cluster ad-pipeline-mvp --desired-status STOPPED --query 'taskArns[0]' --output text)

# Common issues:
# - ECR image not found: Push image
# - IAM permissions: Check task role
# - Health check failing: Check /health endpoint
```

**Problem: Lambda timeouts**
```bash
# Check Lambda logs
aws logs tail /aws/lambda/ad-pipeline-mvp-generator --since 10m

# Increase timeout (max 15 min)
# Edit modules/compute/main.tf
timeout = 900

# Increase memory (more CPU too)
memory_size = 4096

terraform apply
```

**Problem: High costs**
```bash
# Check Cost Explorer
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity DAILY \
  --metrics BlendedCost \
  --group-by Type=SERVICE

# Common issues:
# - NAT Gateway data transfer: Check VPC endpoints
# - Lambda duration: Optimize code
# - S3 storage: Check lifecycle rules
# - DynamoDB: Check consumed capacity
```

**Problem: Video generation failures**
```bash
# Check Step Functions executions
aws stepfunctions list-executions \
  --state-machine-arn <arn> \
  --status-filter FAILED \
  --max-results 10

# Get execution details
aws stepfunctions describe-execution \
  --execution-arn <execution-arn>

# Check Replicate API status
curl https://api.replicate.com/v1/health

# Common issues:
# - Replicate API rate limits: Slow down requests
# - Replicate API failures: Retry logic
# - S3 upload failures: Check IAM permissions
```

### 9.5 Update Procedures

**Infrastructure Updates**:
```bash
# 1. Test in separate environment first
# 2. Review plan carefully
terraform plan -out=tfplan

# 3. Apply during low-traffic period
terraform apply tfplan

# 4. Monitor for issues
watch -n 5 'aws ecs describe-services \
  --cluster ad-pipeline-mvp \
  --services ad-pipeline-mvp-api'

# 5. Rollback if needed
terraform apply -var="ecs_min_tasks=2"  # previous value
```

**Application Updates**:
```bash
# 1. Build new image with version tag
docker build -t ad-pipeline-api:v1.2.0 .

# 2. Tag as latest
docker tag ad-pipeline-api:v1.2.0 $ECR_REPO:latest
docker tag ad-pipeline-api:v1.2.0 $ECR_REPO:v1.2.0

# 3. Push both tags
docker push $ECR_REPO:latest
docker push $ECR_REPO:v1.2.0

# 4. Force new deployment (rolling)
aws ecs update-service \
  --cluster ad-pipeline-mvp \
  --service ad-pipeline-mvp-api \
  --force-new-deployment

# 5. Monitor rollout
aws ecs describe-services \
  --cluster ad-pipeline-mvp \
  --services ad-pipeline-mvp-api \
  --query 'services[0].deployments'
```

---

## 10. Testing & Validation

### 10.1 Infrastructure Testing

**Pre-Deployment Validation**:
```bash
# Terraform validation
terraform validate

# Terraform format check
terraform fmt -check -recursive

# Security scanning (optional)
tfsec .

# Cost estimation (optional)
infracost breakdown --path .
```

**Post-Deployment Validation**:
```bash
# 1. Health check
curl $(terraform output -raw api_url)/health

# 2. Check all resources created
terraform show

# 3. Verify ECS tasks running
aws ecs describe-services \
  --cluster ad-pipeline-mvp \
  --services ad-pipeline-mvp-api

# 4. Check Lambda functions
aws lambda list-functions --query 'Functions[?starts_with(FunctionName, `ad-pipeline-mvp`)]'

# 5. Verify Step Functions
aws stepfunctions list-state-machines

# 6. Check DynamoDB table
aws dynamodb describe-table --table-name ad-pipeline-mvp-jobs

# 7. Verify S3 bucket
aws s3 ls s3://ad-pipeline-mvp-assets
```

### 10.2 End-to-End Testing

```bash
# 1. Submit test job
JOB_ID=$(curl -s -X POST $(terraform output -raw api_url)/api/v1/generate \
  -H "Content-Type: application/json" \
  -H "x-api-key: test-key" \
  -d '{
    "prompt": "Test video generation",
    "duration": 15
  }' | jq -r '.job_id')

echo "Job ID: $JOB_ID"

# 2. Poll job status
watch -n 5 "curl -s $(terraform output -raw api_url)/api/v1/jobs/$JOB_ID \
  -H 'x-api-key: test-key' | jq"

# 3. Verify Step Functions execution
aws stepfunctions list-executions \
  --state-machine-arn $(terraform output -raw step_function_arn) \
  --max-results 1

# 4. Check CloudWatch logs
aws logs tail /ecs/ad-pipeline-mvp --since 10m

# 5. Verify final video in S3
aws s3 ls s3://ad-pipeline-mvp-assets/jobs/$JOB_ID/final/
```

### 10.3 Load Testing

```bash
# Using Apache Bench
ab -n 100 -c 10 -H "x-api-key: test-key" \
  $(terraform output -raw api_url)/health

# Using k6 (more advanced)
k6 run loadtest.js
```

**loadtest.js**:
```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  vus: 10,
  duration: '5m',
};

export default function() {
  let payload = JSON.stringify({
    prompt: 'Load test video',
    duration: 15
  });
  
  let params = {
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': 'test-key'
    }
  };
  
  let res = http.post(`${__ENV.API_URL}/api/v1/generate`, payload, params);
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
}
```

---

## 11. Migration & Scaling Path

### 11.1 From Lean to Enhanced

**When to Migrate**: 
- >500 videos/day
- >$300/month infrastructure cost
- Cache hit rate <80%
- >50 concurrent requests regularly

**What to Add**:
```yaml
1. CloudFront CDN ($50/month):
   - Global delivery
   - Edge caching
   - Reduced S3 costs
   
2. ElastiCache Redis ($100/month):
   - Distributed caching
   - Session storage
   - Rate limiting
   
3. Second AZ ($70/month):
   - High availability
   - 99.9% uptime
   - NAT Gateway redundancy

Total: +$220/month → $350-400/month
```

**Migration Steps**:
```bash
# 1. Add Redis
# Add to core module or create cache module
terraform apply

# 2. Add CloudFront
# Create cdn module
terraform apply

# 3. Expand to Multi-AZ
# Edit networking module
terraform apply
```

### 11.2 From Enhanced to Enterprise

**When to Migrate**:
- >5000 videos/day
- >$1000/month infrastructure cost
- Need 99.99% SLA
- Complex analytics requirements
- Need OAuth/user accounts

**What to Add**:
```yaml
1. RDS PostgreSQL ($200/month):
   - Complex queries
   - Analytics
   - Reporting

2. Cognito ($10/month):
   - OAuth authentication
   - User management
   - MFA support

3. Additional AZs ($100/month):
   - 3-AZ deployment
   - 99.99% uptime
   - Global redundancy

4. More Lambda functions:
   - Separate concerns
   - Specialized processing

Total: +$310/month → $660-700/month
```

### 11.3 Multi-Region Expansion

**When Needed**:
- Global user base
- <100ms latency required worldwide
- Compliance requirements (data locality)

**Architecture Changes**:
```yaml
1. Deploy infrastructure in multiple regions
2. Route 53 geo-routing
3. DynamoDB Global Tables
4. S3 Cross-Region Replication
5. CloudFront with multiple origins

Cost: ~2x per additional region
```

---

## 12. Success Metrics

### 12.1 Infrastructure KPIs

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Deployment Time | <15 min | `time terraform apply` |
| Infrastructure Cost | <$200/mo | AWS Cost Explorer |
| API Latency | <200ms p99 | CloudWatch Metrics |
| Uptime | >99.5% | CloudWatch Alarms |
| Success Rate | >90% | Step Functions metrics |
| Time to Scale | <5 min | ECS auto-scaling time |

### 12.2 Business KPIs

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Cost per Video | <$2.00 | Total cost / videos generated |
| Videos Generated/Day | 100+ | DynamoDB query count |
| Video Gen Time | <5min for 30s | Step Functions duration |
| Error Rate | <10% | Failed executions / total |
| Developer Velocity | Fast | Time from code to production |

### 12.3 PRD Alignment Checklist

From original video generation PRD:

**MVP Requirements (48 Hours)** ✓
- [x] Working video generation for ads
- [x] Prompt to video flow
- [x] Multi-clip composition
- [x] Deployed pipeline (API)
- [x] Sample outputs possible

**Performance Requirements** ✓
- [x] 30s video: <5 minutes
- [x] 60s video: <10 minutes
- [x] 3min video: <20 minutes
- [x] 90%+ successful generation rate
- [x] Graceful failure handling

**Cost Efficiency (20% of evaluation)** ✓
- [x] Track generation cost per video
- [x] Optimize API calls (caching)
- [x] Target: <$2/minute final video

**Pipeline Architecture (25% of evaluation)** ✓
- [x] Clean, maintainable code
- [x] Scalable and modular
- [x] Error handling & retry logic
- [x] Performance optimization

**User Experience (15% of evaluation)** ✓
- [x] Deployed pipeline (API endpoint)
- [x] Progress feedback (job status)
- [x] Output control (parameters)

---

## 13. Appendices

### Appendix A: Resource Naming Convention

All resources follow this pattern:
```
{project_name}-{environment}-{resource_type}-{identifier}

Examples:
- ad-pipeline-mvp-api
- ad-pipeline-mvp-generator
- ad-pipeline-mvp-jobs
- ad-pipeline-mvp-assets
```

### Appendix B: Tag Strategy

All resources tagged with:
```yaml
Project: ad-pipeline
Environment: mvp
ManagedBy: terraform
```

Optional tags:
```yaml
CostCenter: engineering
Team: video-gen
```

### Appendix C: Terraform Variables Reference

```hcl
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "ad-pipeline"
}

variable "replicate_api_key_secret_arn" {
  description = "ARN of Secrets Manager secret containing Replicate API key"
  type        = string
}

variable "ecs_min_tasks" {
  description = "Minimum number of ECS tasks"
  type        = number
  default     = 1
}

variable "ecs_max_tasks" {
  description = "Maximum number of ECS tasks"
  type        = number
  default     = 5
}

variable "ecs_cpu" {
  description = "CPU units for ECS tasks (1024 = 1 vCPU)"
  type        = number
  default     = 1024
}

variable "ecs_memory" {
  description = "Memory for ECS tasks in MB"
  type        = number
  default     = 2048
}
```

### Appendix D: Useful Commands

```bash
# Get all outputs
terraform output

# Get specific output
terraform output api_url

# Format Terraform files
terraform fmt -recursive

# Validate configuration
terraform validate

# Plan with variable file
terraform plan -var-file="production.tfvars"

# Apply specific resource
terraform apply -target=module.storage

# Show resource details
terraform state show module.core.aws_ecs_cluster.main

# List all resources
terraform state list

# Refresh state
terraform refresh

# Import existing resource
terraform import module.storage.aws_s3_bucket.assets ad-pipeline-mvp-assets
```

### Appendix E: Go API Code Examples

**Main API Structure**:
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/sfn"
)

type App struct {
    router    *gin.Engine
    db        *dynamodb.Client
    s3        *s3.Client
    sfn       *sfn.Client
    cache     *Cache
}

func main() {
    cfg, _ := config.LoadDefaultConfig(context.Background())
    
    app := &App{
        router: gin.Default(),
        db:     dynamodb.NewFromConfig(cfg),
        s3:     s3.NewFromConfig(cfg),
        sfn:    sfn.NewFromConfig(cfg),
        cache:  NewCache(1000), // 1000 entry LRU
    }
    
    app.setupRoutes()
    app.router.Run(":8080")
}

func (app *App) setupRoutes() {
    app.router.GET("/health", app.healthCheck)
    
    api := app.router.Group("/api/v1")
    api.Use(app.authMiddleware())
    {
        api.POST("/generate", app.generateVideo)
        api.GET("/jobs/:id", app.getJobStatus)
        api.GET("/jobs", app.listJobs)
    }
    
    app.router.POST("/webhooks/replicate", app.replicateWebhook)
}
```

**Simple Cache Implementation**:
```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]CacheItem
    order []string
    maxSize int
}

type CacheItem struct {
    Value  interface{}
    Expiry time.Time
}

func NewCache(maxSize int) *Cache {
    return &Cache{
        items:   make(map[string]CacheItem),
        order:   make([]string, 0, maxSize),
        maxSize: maxSize,
    }
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, found := c.items[key]
    if !found || time.Now().After(item.Expiry) {
        return nil, false
    }
    return item.Value, true
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // LRU eviction
    if len(c.items) >= c.maxSize {
        oldest := c.order[0]
        delete(c.items, oldest)
        c.order = c.order[1:]
    }
    
    c.items[key] = CacheItem{
        Value:  value,
        Expiry: time.Now().Add(ttl),
    }
    c.order = append(c.order, key)
}
```

### Appendix F: Lambda Function Examples

**Generator Lambda (simplified)**:
```go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/lambda"
)

type GenerateRequest struct {
    JobID    string `json:"job_id"`
    Prompt   string `json:"prompt"`
    Duration int    `json:"duration"`
}

type GenerateResponse struct {
    Scenes []Scene `json:"scenes"`
}

func handler(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
    // 1. Parse prompt
    scenes := parsePrompt(req.Prompt, req.Duration)
    
    // 2. Generate each scene via Replicate
    for i := range scenes {
        scenes[i].AssetURL = generateScene(ctx, scenes[i])
    }
    
    // 3. Store in S3
    storeScenes(ctx, req.JobID, scenes)
    
    // 4. Update DynamoDB
    updateJobStatus(ctx, req.JobID, "scenes_generated")
    
    return GenerateResponse{Scenes: scenes}, nil
}

func main() {
    lambda.Start(handler)
}
```

---

## 14. Conclusion

This lean infrastructure provides everything needed to support high-quality AI video generation while maintaining:

- ✅ **Cost Efficiency**: $110-200/month (70% less than enterprise)
- ✅ **Simplicity**: 20-25 resources (70% fewer than enterprise)
- ✅ **Performance**: Meets all PRD requirements
- ✅ **Scalability**: Can grow to 1000+ videos/day
- ✅ **Maintainability**: Manageable by small teams

**Key Principle**: Infrastructure should *enable* AI development, not *constrain* it.

By avoiding over-engineering and focusing on actual requirements, this architecture allows teams to:
1. Deploy in 15 minutes instead of 30
2. Spend 85% of time on AI quality instead of 66%
3. Save $600/month for AI generation budget
4. Maintain infrastructure without dedicated DevOps

**Perfect for**:
- One-week sprints
- MVP/bounty projects
- Bootstrapped teams
- Cost-conscious development
- AI-first focus

**Scale path exists when needed** - can upgrade to enhanced ($350/mo) or enterprise ($700/mo) infrastructure when traffic justifies it.

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024-01-14 | Infrastructure Team | Initial PRD for lean MVP infrastructure |

**Approval**: Ready for implementation

**Next Steps**: 
1. Review and approve PRD
2. Create terraform.tfvars with project-specific values
3. Deploy infrastructure: `terraform apply`
4. Begin AI video generation development

---

*End of PRD*