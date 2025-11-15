# CI/CD Pipeline

> GitHub Actions workflows for infrastructure, backend, and frontend deployments with OIDC authentication

## Overview

OmniGen uses **GitHub Actions** for continuous deployment with three independent workflows:
1. **Infrastructure** (`infrastructure.yml`) - Terraform deployment of AWS resources
2. **Backend** (`backend.yml`) - Docker build + ECS deployment
3. **Frontend** (`frontend.yml`) - Bun build + S3/CloudFront deployment

**Key Features:**
- **OIDC Authentication:** No long-lived AWS credentials stored in GitHub
- **Workflow Dependencies:** Backend + Frontend wait for infrastructure completion
- **Concurrency Control:** Prevents simultaneous deployments to same environment
- **Path-Based Triggers:** Only deploy what changed
- **Manual Override:** `workflow_dispatch` for on-demand deployments

**Security:**
- AWS credentials valid for 1 hour (STS AssumeRoleWithWebIdentity)
- Terraform state locked via DynamoDB (prevents concurrent modifications)
- No secrets in code (everything via GitHub Secrets)

---

## Complete CI/CD Architecture

High-level view of all workflows and their interactions.

```mermaid
flowchart TB
    subgraph GitHub[\"GitHub Repository\"]
        Code[Source Code<br/>infrastructure/, backend/, frontend/]
        Actions[GitHub Actions<br/>Workflow Runner]
        Secrets[GitHub Secrets<br/>AWS_ROLE_ARN, etc.]
    end

    subgraph OIDC[\"AWS OIDC Provider\"]
        Provider[OIDC Provider<br/>token.actions.githubusercontent.com]
        Role[IAM Role<br/>GitHubActionsRole]
        Trust[Trust Policy<br/>Sub: repo:owner/repo:ref:refs/heads/master]
    end

    subgraph Workflows[\"GitHub Actions Workflows\"]
        Infra[Infrastructure Workflow<br/>infrastructure.yml]
        Backend[Backend Workflow<br/>backend.yml]
        Frontend[Frontend Workflow<br/>frontend.yml]
    end

    subgraph AWS[\"AWS Cloud (us-east-1)\"]
        S3State[S3 Bucket<br/>Terraform State]
        DDBLock[DynamoDB Table<br/>State Lock]

        ECR[ECR Repository<br/>Docker Images]
        ECS[ECS Fargate<br/>API Service]

        S3Frontend[S3 Bucket<br/>Frontend Assets]
        CloudFront[CloudFront<br/>Distribution]

        VPC[VPC + Networking]
        Lambda[Lambda Functions]
        SFN[Step Functions]
        Cognito[Cognito User Pool]
    end

    Code -->|Push to master| Actions
    Actions -->|Trigger| Workflows

    Infra -->|1. Authenticate| Provider
    Backend -->|1. Authenticate| Provider
    Frontend -->|1. Authenticate| Provider

    Provider -->|Verify JWT| Trust
    Trust -->|AssumeRole| Role
    Role -->|Temporary Creds| Infra
    Role -->|Temporary Creds| Backend
    Role -->|Temporary Creds| Frontend

    Infra -->|2. Terraform Init| S3State
    Infra -->|3. Acquire Lock| DDBLock
    Infra -->|4. Terraform Apply| VPC
    Infra -->|Deploy| Lambda
    Infra -->|Deploy| SFN
    Infra -->|Deploy| Cognito

    Backend -->|2. Build Docker| ECR
    Backend -->|3. Deploy| ECS

    Frontend -->|2. Bun Build| S3Frontend
    Frontend -->|3. Invalidate Cache| CloudFront

    Infra -.->|workflow_run trigger| Backend
    Infra -.->|workflow_run trigger| Frontend

    style Workflows fill:#e1f5ff,stroke:#0288d1,stroke-width:2px
    style OIDC fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    style AWS fill:#fff3e0,stroke:#f57c00,stroke-width:2px
```

---

## Infrastructure Workflow

Terraform deployment of all AWS resources.

```mermaid
sequenceDiagram
    actor Dev as Developer
    participant GH as GitHub Actions<br/>infrastructure.yml
    participant OIDC as AWS OIDC Provider
    participant STS as AWS STS
    participant S3 as S3 Terraform State
    participant DDB as DynamoDB Lock
    participant TF as Terraform
    participant AWS as AWS APIs<br/>(VPC, ECS, etc.)

    Dev->>GH: git push origin master<br/>(changes in infrastructure/)

    Note over GH: Trigger: push.paths includes<br/>'infrastructure/**'

    GH->>GH: Checkout code<br/>actions/checkout@v5

    GH->>GH: Setup Terraform<br/>hashicorp/setup-terraform@v3<br/>version: 1.13.5

    Note over GH: OIDC Authentication

    GH->>OIDC: Request JWT token<br/>Audience: sts.amazonaws.com<br/>Subject: repo:owner/omnigen:ref:refs/heads/master

    OIDC-->>GH: JWT token (5 min TTL)

    GH->>STS: AssumeRoleWithWebIdentity<br/>RoleArn: arn:aws:iam::123456789012:role/GitHubActionsRole<br/>WebIdentityToken: {JWT}<br/>RoleSessionName: github-actions-deploy

    STS->>STS: Validate JWT signature<br/>Check trust policy<br/>Verify repo + branch match

    alt Trust policy allows
        STS-->>GH: Temporary credentials<br/>AccessKeyId, SecretAccessKey, SessionToken<br/>Valid for: 1 hour
    else Trust policy denies
        STS-->>GH: Error: Not authorized
        GH-->>Dev: Workflow failed: OIDC auth failed
    end

    GH->>GH: Configure AWS credentials<br/>aws-actions/configure-aws-credentials@v5<br/>role-to-assume: {RoleArn}

    Note over GH: Terraform Init

    GH->>TF: terraform init<br/>-backend-config="bucket=omnigen-terraform-state"<br/>-backend-config="key=terraform.tfstate"<br/>-backend-config="region=us-east-1"

    TF->>S3: Check if state exists<br/>GetObject: s3://omnigen-terraform-state/terraform.tfstate

    alt First deployment
        S3-->>TF: 404 Not Found
        TF->>TF: Initialize empty state
    else State exists
        S3-->>TF: State file (JSON)
        TF->>TF: Load existing state
    end

    TF->>DDB: Acquire state lock<br/>PutItem: terraform-state-lock<br/>LockID: omnigen-terraform-state/terraform.tfstate

    alt Lock acquired
        DDB-->>TF: Success
    else Lock held by another process
        DDB-->>TF: Error: Lock already held
        TF-->>GH: Error: State locked
        GH-->>Dev: Workflow failed: Concurrent deployment detected
    end

    Note over GH: Terraform Plan

    GH->>TF: terraform plan -out=tfplan

    TF->>AWS: DescribeVpcs, DescribeSubnets, etc.<br/>(Read current AWS state)

    AWS-->>TF: Current resource state

    TF->>TF: Compare desired state (HCL)<br/>vs current state (AWS)<br/>Generate execution plan

    TF-->>GH: Plan output:<br/>+ 15 to create<br/>~ 3 to modify<br/>- 0 to destroy

    Note over GH: Terraform Apply

    GH->>TF: terraform apply -auto-approve tfplan

    loop For each resource change
        TF->>AWS: Create/Update/Delete resource<br/>Example: CreateVpc, CreateSubnet, CreateSecurityGroup

        AWS-->>TF: Resource created/updated<br/>ResourceId, ARN, etc.

        TF->>TF: Update state file<br/>Add/Modify resource in state
    end

    TF->>S3: PutObject: Updated state file<br/>s3://omnigen-terraform-state/terraform.tfstate

    S3-->>TF: State saved

    TF->>DDB: Release lock<br/>DeleteItem: terraform-state-lock

    DDB-->>TF: Lock released

    TF-->>GH: Apply complete!<br/>Resources: 15 added, 3 changed, 0 destroyed

    Note over GH: Output Terraform Values

    GH->>TF: terraform output -json

    TF-->>GH: JSON output:<br/>{<br/>  "ecs_cluster_name": "omnigen",<br/>  "ecr_repository_url": "123456789012.dkr.ecr.us-east-1.amazonaws.com/omnigen-api",<br/>  "cloudfront_domain": "d1234567890.cloudfront.net"<br/>}

    GH->>GH: Upload outputs as artifact<br/>actions/upload-artifact@v5<br/>name: terraform-outputs

    GH-->>Dev: Workflow succeeded<br/>Duration: 3-5 minutes

    Note over GH,Dev: Backend + Frontend workflows triggered<br/>via workflow_run event
```

**Workflow File:** `.github/workflows/infrastructure.yml`

**Key Steps:**
1. **Checkout:** `actions/checkout@v5`
2. **Setup Terraform:** `hashicorp/setup-terraform@v3` (v1.13.5)
3. **Configure AWS:** `aws-actions/configure-aws-credentials@v5` (OIDC)
4. **Terraform Init:** Backend S3 + DynamoDB lock
5. **Terraform Plan:** Generate execution plan
6. **Terraform Apply:** Deploy resources (auto-approve)
7. **Upload Outputs:** Save as artifact for dependent workflows

**Triggers:**
```yaml
on:
  push:
    branches: [master]
    paths:
      - 'infrastructure/**'
      - '.github/workflows/infrastructure.yml'
  workflow_dispatch:  # Manual trigger
```

**Concurrency:**
```yaml
concurrency:
  group: infrastructure-deploy
  cancel-in-progress: false  # Don't cancel running deployments
```

---

## Backend Workflow

Docker build and ECS deployment.

```mermaid
sequenceDiagram
    actor Dev as Developer
    participant GH as GitHub Actions<br/>backend.yml
    participant AWS as AWS Configure
    participant ECR as ECR Repository
    participant Docker as Docker Build
    participant ECS as ECS Fargate
    participant ALB as Application LB

    Dev->>GH: git push origin master<br/>(changes in backend/)

    alt Infrastructure changed
        Note over GH: Wait for infrastructure.yml<br/>to complete (workflow_run trigger)
        GH->>GH: Download terraform-outputs artifact<br/>actions/download-artifact@v6
    else Infrastructure unchanged
        Note over GH: Run immediately on push
    end

    GH->>GH: Checkout code<br/>actions/checkout@v5

    GH->>AWS: Configure AWS credentials<br/>OIDC authentication

    Note over GH: Login to ECR

    GH->>ECR: aws ecr get-login-password | docker login

    ECR-->>GH: Login succeeded

    Note over GH: Build Docker Image

    GH->>Docker: docker build -f backend/Dockerfile<br/>-t omnigen-api:$GITHUB_SHA<br/>--target production<br/>backend/

    Note over Docker: Multi-stage build:<br/>1. Build stage (Go compile)<br/>2. Production stage (scratch)

    Docker-->>GH: Image built: omnigen-api:abc1234

    Note over GH: Tag and Push to ECR

    GH->>GH: docker tag omnigen-api:$GITHUB_SHA<br/>123456789012.dkr.ecr.us-east-1.amazonaws.com/omnigen-api:$GITHUB_SHA

    GH->>GH: docker tag omnigen-api:$GITHUB_SHA<br/>123456789012.dkr.ecr.us-east-1.amazonaws.com/omnigen-api:latest

    GH->>ECR: docker push (SHA tag)
    GH->>ECR: docker push (latest tag)

    ECR-->>GH: Push complete (100 MB image)

    Note over GH: Update ECS Task Definition

    GH->>ECS: aws ecs describe-task-definition<br/>--task-definition omnigen-api

    ECS-->>GH: Current task definition JSON

    GH->>GH: Update image in JSON:<br/>image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/omnigen-api:$GITHUB_SHA

    GH->>ECS: aws ecs register-task-definition<br/>--cli-input-json {updatedDefinition}

    ECS-->>GH: New task definition revision: 42

    Note over GH: Deploy to ECS Service

    GH->>ECS: aws ecs update-service<br/>--cluster omnigen<br/>--service omnigen-api<br/>--task-definition omnigen-api:42<br/>--force-new-deployment

    ECS->>ECS: Start new task with revision 42

    Note over ECS: Rolling deployment:<br/>1. Start new task (v42)<br/>2. Wait for health check<br/>3. Stop old task (v41)

    ECS->>ALB: Register new task target<br/>IP: 10.0.10.5, Port: 8080

    ALB->>ECS: Health check GET /api/v1/health

    loop Health check (every 30s, timeout 5s)
        ECS-->>ALB: 200 OK {"status": "healthy"}
    end

    alt Health check passes
        ALB->>ALB: Mark target healthy<br/>Begin routing traffic

        ALB->>ECS: Deregister old task<br/>IP: 10.0.10.4

        ECS->>ECS: Stop old task (graceful shutdown)

        ECS-->>GH: Service updated successfully<br/>Desired: 1, Running: 1, Pending: 0
    else Health check fails (2 consecutive)
        ALB-->>ECS: Target unhealthy

        ECS->>ECS: Stop new task (rollback)

        ECS->>ECS: Keep old task running

        ECS-->>GH: Deployment failed: Circuit breaker triggered
        GH-->>Dev: Workflow failed: ECS health check failed
    end

    GH->>GH: Wait for deployment stable<br/>aws ecs wait services-stable<br/>--cluster omnigen --services omnigen-api

    ECS-->>GH: Service stable (timeout: 10 min)

    GH-->>Dev: Workflow succeeded<br/>Duration: 5-8 minutes
```

**Workflow File:** `.github/workflows/backend.yml`

**Key Steps:**
1. **Wait for Infrastructure:** `workflow_run` trigger (if infra changed)
2. **Checkout:** `actions/checkout@v5`
3. **Configure AWS:** OIDC authentication
4. **ECR Login:** `docker login` via AWS CLI
5. **Build Docker Image:** Multi-stage build (Go compile → scratch)
6. **Push to ECR:** Tag with `$GITHUB_SHA` and `latest`
7. **Update Task Definition:** New revision with updated image
8. **Deploy to ECS:** Rolling update with circuit breaker
9. **Wait for Stable:** Health checks must pass

**Triggers:**
```yaml
on:
  push:
    branches: [master]
    paths:
      - 'backend/**'
      - '.github/workflows/backend.yml'
  workflow_run:
    workflows: ["Infrastructure"]
    types: [completed]
    branches: [master]
  workflow_dispatch:
```

**Health Check Configuration:**
```yaml
Health Check Path: /api/v1/health
Health Check Interval: 30 seconds
Health Check Timeout: 5 seconds
Healthy Threshold: 2 consecutive successes
Unhealthy Threshold: 2 consecutive failures
```

**Circuit Breaker:**
- **Enabled:** Prevents bad deployments from taking down service
- **Rollback Trigger:** 2 failed health checks
- **Behavior:** Stops new task, keeps old task running

---

## Frontend Workflow

Bun build and S3/CloudFront deployment.

```mermaid
sequenceDiagram
    actor Dev as Developer
    participant GH as GitHub Actions<br/>frontend.yml
    participant Bun as Bun Build
    participant S3 as S3 Frontend Bucket
    participant CF as CloudFront

    Dev->>GH: git push origin master<br/>(changes in frontend/)

    alt Infrastructure changed
        Note over GH: Wait for infrastructure.yml<br/>to complete (workflow_run trigger)
        GH->>GH: Download terraform-outputs artifact<br/>Get CloudFront domain, API URL
    else Infrastructure unchanged
        Note over GH: Run immediately
    end

    GH->>GH: Checkout code<br/>actions/checkout@v5

    GH->>GH: Setup Bun<br/>oven-sh/setup-bun@v2<br/>bun-version: latest

    Note over GH: Install Dependencies

    GH->>Bun: bun install --frozen-lockfile

    Bun->>Bun: Read bun.lockb<br/>Install exact versions

    Bun-->>GH: Installed 200 packages (2s)

    Note over GH: Build Frontend

    GH->>GH: Set environment variables:<br/>VITE_API_URL={terraform_output.api_url}<br/>VITE_COGNITO_DOMAIN={terraform_output.cognito_domain}<br/>VITE_COGNITO_CLIENT_ID={terraform_output.cognito_client_id}

    GH->>Bun: bun run build

    Bun->>Bun: Vite build pipeline:<br/>1. Load React components<br/>2. Bundle JavaScript (esbuild)<br/>3. Minify CSS<br/>4. Copy static assets<br/>5. Generate index.html

    Bun-->>GH: Build complete<br/>Output: frontend/dist/ (5 MB)

    Note over GH: Deploy to S3

    GH->>GH: Configure AWS credentials<br/>OIDC authentication

    GH->>S3: aws s3 sync frontend/dist/<br/>s3://omnigen-frontend<br/>--delete<br/>--cache-control "public, max-age=31536000"

    Note over S3: --delete removes old files<br/>Ensures no stale content

    S3-->>GH: Sync complete<br/>Uploaded: 25 files, Deleted: 3 files

    Note over GH: Invalidate CloudFront Cache

    GH->>CF: aws cloudfront create-invalidation<br/>--distribution-id E1234567890ABC<br/>--paths "/*"

    CF->>CF: Invalidate all edge locations<br/>(global cache clear)

    CF-->>GH: Invalidation created<br/>ID: I1ABCDEFGHIJK<br/>Status: InProgress

    GH->>GH: Wait for invalidation complete<br/>aws cloudfront wait invalidation-completed<br/>--id I1ABCDEFGHIJK

    loop Poll status (every 20s)
        CF-->>GH: Status: InProgress
    end

    CF-->>GH: Status: Completed (60-90s)

    GH-->>Dev: Workflow succeeded<br/>Frontend live at: https://d1234567890.cloudfront.net<br/>Duration: 3-5 minutes
```

**Workflow File:** `.github/workflows/frontend.yml`

**Key Steps:**
1. **Wait for Infrastructure:** `workflow_run` trigger
2. **Checkout:** `actions/checkout@v5`
3. **Setup Bun:** `oven-sh/setup-bun@v2`
4. **Install Dependencies:** `bun install --frozen-lockfile`
5. **Build:** `bun run build` (Vite + React)
6. **Configure AWS:** OIDC authentication
7. **Sync to S3:** `aws s3 sync` with `--delete` flag
8. **Invalidate CloudFront:** Clear global edge cache
9. **Wait for Invalidation:** Ensure fresh content served

**Triggers:**
```yaml
on:
  push:
    branches: [master]
    paths:
      - 'frontend/**'
      - '.github/workflows/frontend.yml'
  workflow_run:
    workflows: ["Infrastructure"]
    types: [completed]
    branches: [master]
  workflow_dispatch:
```

**S3 Sync Configuration:**
```bash
--delete              # Remove files not in source
--cache-control "public, max-age=31536000"  # 1 year cache for assets
--exclude "*.html"    # Don't cache HTML (needs fresh content)
```

**CloudFront Invalidation:**
- **Paths:** `/*` (invalidate everything)
- **Cost:** First 1,000 invalidations/month free, then $0.005 per path
- **Duration:** 60-90 seconds (global edge cache clear)

---

## OIDC Authentication Flow

Detailed view of GitHub Actions OIDC authentication with AWS.

```mermaid
sequenceDiagram
    participant GH as GitHub Actions<br/>Runner
    participant GHOIDC as GitHub OIDC<br/>Provider
    participant AWS_OIDC as AWS OIDC<br/>Provider
    participant STS as AWS STS
    participant IAM as IAM Role

    Note over GH: Workflow starts

    GH->>GHOIDC: Request JWT token<br/>POST https://token.actions.githubusercontent.com/<br/>Body: {<br/>  "audience": "sts.amazonaws.com",<br/>  "repository": "owner/omnigen",<br/>  "ref": "refs/heads/master",<br/>  "sha": "abc1234..."<br/>}

    GHOIDC->>GHOIDC: Generate JWT with claims:<br/>{<br/>  "iss": "https://token.actions.githubusercontent.com",<br/>  "sub": "repo:owner/omnigen:ref:refs/heads/master",<br/>  "aud": "sts.amazonaws.com",<br/>  "repository": "owner/omnigen",<br/>  "ref": "refs/heads/master",<br/>  "sha": "abc1234...",<br/>  "workflow": "Infrastructure",<br/>  "exp": now + 5 minutes<br/>}

    GHOIDC-->>GH: JWT token (5 min TTL)

    GH->>STS: AssumeRoleWithWebIdentity<br/>POST https://sts.amazonaws.com/<br/>?Action=AssumeRoleWithWebIdentity<br/>&RoleArn=arn:aws:iam::123456789012:role/GitHubActionsRole<br/>&WebIdentityToken={JWT}<br/>&RoleSessionName=github-actions-deploy<br/>&DurationSeconds=3600

    STS->>AWS_OIDC: Fetch OIDC provider keys<br/>GET https://token.actions.githubusercontent.com/.well-known/jwks

    AWS_OIDC-->>STS: Public keys (RSA)

    STS->>STS: Verify JWT signature<br/>using GitHub public key

    alt Signature invalid
        STS-->>GH: Error: Invalid token signature
    else Signature valid
        STS->>IAM: Get role trust policy<br/>arn:aws:iam::123456789012:role/GitHubActionsRole

        IAM-->>STS: Trust policy:<br/>{<br/>  "Statement": [{<br/>    "Effect": "Allow",<br/>    "Principal": {<br/>      "Federated": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"<br/>    },<br/>    "Action": "sts:AssumeRoleWithWebIdentity",<br/>    "Condition": {<br/>      "StringEquals": {<br/>        "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"<br/>      },<br/>      "StringLike": {<br/>        "token.actions.githubusercontent.com:sub": "repo:owner/omnigen:ref:refs/heads/master"<br/>      }<br/>    }<br/>  }]<br/>}

        STS->>STS: Validate JWT claims:<br/>1. "aud" = "sts.amazonaws.com" ✅<br/>2. "sub" matches pattern ✅<br/>3. "exp" > now ✅

        alt Trust policy allows
            STS->>STS: Generate temporary credentials<br/>Valid for: 1 hour (3600s)

            STS-->>GH: Credentials:<br/>{<br/>  "AccessKeyId": "ASIAXYZ...",<br/>  "SecretAccessKey": "...",<br/>  "SessionToken": "...",<br/>  "Expiration": "2024-12-15T11:30:00Z"<br/>}

            GH->>GH: Set environment variables:<br/>AWS_ACCESS_KEY_ID={AccessKeyId}<br/>AWS_SECRET_ACCESS_KEY={SecretAccessKey}<br/>AWS_SESSION_TOKEN={SessionToken}

            Note over GH: Credentials valid for 1 hour<br/>All AWS CLI/SDK calls use these creds
        else Trust policy denies
            STS-->>GH: Error: Not authorized to assume role
        end
    end
```

**OIDC Provider Configuration (Terraform):**
```hcl
# infrastructure/github-oidc/main.tf
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",  # GitHub OIDC thumbprint
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"   # Backup thumbprint
  ]
}

resource "aws_iam_role" "github_actions" {
  name = "GitHubActionsRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = aws_iam_openid_connect_provider.github.arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
        }
        StringLike = {
          "token.actions.githubusercontent.com:sub" = "repo:owner/omnigen:ref:refs/heads/master"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "github_actions_admin" {
  role       = aws_iam_role.github_actions.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"  # MVP only
}
```

**Security Benefits:**
- No long-lived AWS credentials in GitHub Secrets
- Credentials auto-rotate every deployment
- Credentials expire after 1 hour
- Fine-grained access control via IAM role
- Audit trail via CloudTrail (all STS AssumeRole calls logged)

---

## Deployment Timeline

Gantt chart showing all workflows and their dependencies.

```mermaid
gantt
    title Complete CI/CD Pipeline Timeline
    dateFormat mm:ss
    axisFormat %M:%S

    section Git Push
    Developer pushes to master    :done, push, 00:00, 5s

    section Infrastructure
    Checkout code                 :done, infra_checkout, after push, 10s
    Setup Terraform               :done, infra_setup, after infra_checkout, 5s
    OIDC Authentication           :done, infra_oidc, after infra_setup, 10s
    Terraform Init                :done, infra_init, after infra_oidc, 30s
    Terraform Plan                :done, infra_plan, after infra_init, 45s
    Terraform Apply               :crit, infra_apply, after infra_plan, 180s
    Upload Outputs                :done, infra_upload, after infra_apply, 5s

    section Backend (Parallel after Infra)
    Wait for Infrastructure       :done, backend_wait, after infra_upload, 5s
    Download Outputs              :done, backend_download, after backend_wait, 5s
    Checkout code                 :done, backend_checkout, after backend_download, 10s
    OIDC Authentication           :done, backend_oidc, after backend_checkout, 10s
    Docker Build                  :active, backend_build, after backend_oidc, 120s
    Push to ECR                   :active, backend_push, after backend_build, 60s
    Update ECS                    :active, backend_deploy, after backend_push, 180s
    Wait for Stable               :active, backend_wait_stable, after backend_deploy, 60s

    section Frontend (Parallel after Infra)
    Wait for Infrastructure       :done, frontend_wait, after infra_upload, 5s
    Download Outputs              :done, frontend_download, after frontend_wait, 5s
    Checkout code                 :done, frontend_checkout, after frontend_download, 10s
    Setup Bun                     :done, frontend_setup, after frontend_checkout, 5s
    Bun Install                   :active, frontend_install, after frontend_setup, 30s
    Bun Build                     :active, frontend_build, after frontend_install, 60s
    OIDC Authentication           :done, frontend_oidc, after frontend_build, 10s
    S3 Sync                       :active, frontend_sync, after frontend_oidc, 30s
    CloudFront Invalidation       :active, frontend_invalidate, after frontend_sync, 90s

    section Complete
    All Workflows Done            :milestone, after backend_wait_stable, 0s
```

**Total Deployment Time:**
- **Infrastructure Only:** ~5 minutes
- **Backend After Infra:** ~8 minutes
- **Frontend After Infra:** ~4 minutes
- **Complete Pipeline (all changed):** ~13 minutes (infra sequential, then backend + frontend parallel)

---

## Concurrency Control

How workflows prevent simultaneous deployments.

```yaml
# infrastructure.yml
concurrency:
  group: infrastructure-deploy
  cancel-in-progress: false  # Don't cancel running deployments

# backend.yml
concurrency:
  group: backend-deploy-${{ github.ref }}
  cancel-in-progress: true  # Cancel older deployments (safe for stateless app)

# frontend.yml
concurrency:
  group: frontend-deploy-${{ github.ref }}
  cancel-in-progress: true  # Cancel older deployments (safe for static assets)
```

**Behavior:**
- **Infrastructure:** Never cancel (could leave Terraform state locked)
- **Backend/Frontend:** Cancel older runs (latest code wins)
- **Per-Branch:** Each branch has independent concurrency group

**Example Scenario:**
1. Developer pushes commit A (starts workflow A)
2. 2 minutes later, pushes commit B (starts workflow B)
3. **Infrastructure:** Workflow A completes, then workflow B runs
4. **Backend:** Workflow A canceled, workflow B runs immediately
5. **Frontend:** Workflow A canceled, workflow B runs immediately

---

## Monitoring and Notifications

### GitHub Actions Status

**Workflow Status Page:**
- View at: `https://github.com/owner/omnigen/actions`
- Shows: All workflow runs, duration, logs, artifacts
- Filters: By workflow, branch, event type

**Logs:**
- **Persistent:** 90 days retention
- **Downloadable:** Full logs as ZIP
- **Real-time:** Live tail during execution

### Notifications

**GitHub Notifications:**
- Email on workflow failure (default)
- Slack integration (via GitHub Apps)
- Discord webhook (custom)

**AWS Notifications (Future):**
- SNS topic for deployment events
- CloudWatch alarms for ECS service unhealthy

---

## Troubleshooting

### Common Issues

**1. OIDC Authentication Failed**
```
Error: Not authorized to assume role
Causes:
- Trust policy doesn't match repo/branch
- OIDC provider thumbprint outdated
- IAM role deleted
Solution: Verify trust policy in infrastructure/github-oidc/main.tf
```

**2. Terraform State Locked**
```
Error: Error acquiring the state lock
Causes:
- Previous workflow failed mid-apply
- Concurrent deployment attempted
Solution: Manually release lock:
  aws dynamodb delete-item \
    --table-name terraform-state-lock \
    --key '{"LockID": {"S": "omnigen-terraform-state/terraform.tfstate"}}'
```

**3. ECS Health Check Failed**
```
Error: Circuit breaker triggered, rolling back
Causes:
- Backend code crashes on startup
- Health endpoint not responding
- Security group blocking ALB → ECS
Solution: Check ECS logs in CloudWatch, verify health endpoint
```

**4. CloudFront Cache Not Invalidated**
```
Error: Old content still served
Causes:
- Invalidation not waited for
- Cache-Control headers too aggressive
Solution: Wait for invalidation complete, adjust cache headers
```

---

## Cost Analysis

### GitHub Actions

**Free Tier:**
- 2,000 minutes/month (public repos unlimited)
- 500 MB storage (artifacts, logs, caches)

**Usage:**
- **Infrastructure:** ~5 min/run
- **Backend:** ~8 min/run
- **Frontend:** ~4 min/run
- **Total per deployment:** ~17 minutes

**Estimated Monthly Cost:**
- 20 deployments/month x 17 min = 340 minutes
- Well under free tier (2,000 min)
- **Cost: $0/month**

### AWS (CI/CD Related)

| Service | Usage | Cost |
|---------|-------|------|
| **S3 (Terraform State)** | <1 GB storage, 100 requests/month | $0.03 |
| **DynamoDB (State Lock)** | On-demand, ~100 writes/month | $0.001 |
| **CloudFront (Invalidations)** | 20/month x 1 path = 20 | $0.00 (first 1,000 free) |
| **ECR (Docker Images)** | 5 GB storage, 100 pushes/month | $0.50 |
| **TOTAL** | | **$0.53/month** |

---

**Related Documentation:**
- [Architecture Overview](./architecture-overview.md) - System design
- [Backend Architecture](./backend-architecture.md) - Docker build details
- [Infrastructure Modules](./infrastructure-modules.md) - Terraform structure
