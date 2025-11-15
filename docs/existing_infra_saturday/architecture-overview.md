# Architecture Overview

> High-level system architecture for OmniGen AI video generation pipeline

## System Context (C4 Level 1)

This diagram shows the OmniGen system from a bird's-eye view, including external actors and systems.

```mermaid
C4Context
    title System Context - OmniGen AI Video Generation Platform

    Person(user, "End User", "Creates AI-generated videos via web interface")

    System_Boundary(omnigen, "OmniGen Platform") {
        System(webapp, "OmniGen Web App", "React SPA hosted on CloudFront/S3")
        System(api, "OmniGen API", "Go API on ECS Fargate")
        System(pipeline, "Video Pipeline", "Step Functions + Lambda")
    }

    System_Ext(cognito, "AWS Cognito", "OAuth2/OIDC authentication")
    System_Ext(replicate, "Replicate AI", "AI model API (Stable Diffusion, Runway, etc.)")
    System_Ext(github, "GitHub Actions", "CI/CD via OIDC")

    Rel(user, webapp, "Uses", "HTTPS")
    Rel(webapp, api, "Calls API", "HTTPS/JSON")
    Rel(webapp, cognito, "Authenticates", "OAuth2")
    Rel(api, cognito, "Validates JWT", "JWKS")
    Rel(api, pipeline, "Starts workflow", "Step Functions API")
    Rel(pipeline, replicate, "Generates media", "HTTPS/REST")
    Rel(github, webapp, "Deploys frontend", "S3 sync")
    Rel(github, api, "Deploys backend", "ECR push + ECS deploy")

    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

---

## Container Diagram (C4 Level 2)

This diagram shows the major containers (applications/services) within the OmniGen system.

```mermaid
C4Container
    title Container Diagram - OmniGen Platform Components

    Person(user, "End User", "Video creator")
    System_Ext(replicate, "Replicate AI", "AI model inference")

    Container_Boundary(aws, "AWS Cloud (us-east-1)") {

        Container(cdn, "CloudFront CDN", "AWS CloudFront", "Global edge distribution, HTTPS termination, routing")
        Container(spa, "React SPA", "Vite + React 18", "Single-page application, OAuth2 login")
        Container(alb, "Application Load Balancer", "AWS ALB", "HTTP routing to ECS tasks")
        Container(ecs, "Go API Server", "Gin + Go 1.25", "REST API, JWT validation, business logic")

        ContainerDb(s3_fe, "Frontend Bucket", "S3", "Static assets (HTML/JS/CSS)")
        ContainerDb(s3_assets, "Assets Bucket", "S3", "Generated videos, media files")
        ContainerDb(ecr, "Container Registry", "ECR", "Docker images")
        ContainerDb(jobs_db, "Jobs Table", "DynamoDB", "Job metadata, status tracking")
        ContainerDb(usage_db, "Usage Table", "DynamoDB", "User quotas, rate limits")

        Container(sfn, "Workflow Orchestrator", "Step Functions Express", "Video generation state machine")
        Container(generator, "Scene Generator", "Lambda Node.js 20", "Calls Replicate for scenes")
        Container(composer, "Video Composer", "Lambda Node.js 20", "FFmpeg video stitching")

        Container(cognito, "User Pool", "Cognito", "OAuth2 provider, user management")
        ContainerDb(secrets, "Secrets Store", "Secrets Manager", "Replicate API key")
    }

    Rel(user, cdn, "HTTPS requests", "443")
    Rel(cdn, spa, "GET /", "Static files")
    Rel(cdn, alb, "POST /api/*", "Proxy to backend")
    Rel(spa, cognito, "OAuth2 login", "Hosted UI")
    Rel(alb, ecs, "HTTP", "Port 8080")

    Rel(ecs, jobs_db, "Read/Write", "DynamoDB SDK")
    Rel(ecs, usage_db, "Read/Write", "DynamoDB SDK")
    Rel(ecs, s3_assets, "Upload/Download", "S3 SDK")
    Rel(ecs, sfn, "StartExecution", "Step Functions SDK")
    Rel(ecs, cognito, "Validate JWT", "JWKS endpoint")

    Rel(sfn, generator, "Invoke", "Lambda SDK")
    Rel(sfn, composer, "Invoke", "Lambda SDK")
    Rel(generator, replicate, "Generate scenes", "HTTPS/REST")
    Rel(generator, s3_assets, "Upload media", "S3 SDK")
    Rel(generator, secrets, "Get API key", "Secrets SDK")
    Rel(composer, s3_assets, "Read/Write video", "S3 SDK")

    Rel(ecr, ecs, "Pull image", "Docker")
    Rel(s3_fe, cdn, "Origin", "CloudFront OAC")

    UpdateRelStyle(user, cdn, $offsetY="-30")
    UpdateRelStyle(cdn, alb, $offsetX="80")
    UpdateLayoutConfig($c4ShapeInRow="4", $c4BoundaryInRow="1")
```

---

## AWS Services Architecture

This flowchart shows all AWS services and their relationships with color-coding by service category.

```mermaid
%%{init: {'theme':'base', 'themeVariables': { 'fontSize':'14px'}}}%%
flowchart TB
    User([End User])

    subgraph Internet["🌐 Internet"]
        GH[GitHub Actions<br/>CI/CD]
    end

    subgraph CDN["📡 Edge Layer"]
        CF[CloudFront<br/>Distribution]
    end

    subgraph Compute["💻 Compute Layer"]
        ALB[Application<br/>Load Balancer]
        ECS[ECS Fargate<br/>1 vCPU, 2 GB]
        L1[Lambda Generator<br/>2048 MB, 900s]
        L2[Lambda Composer<br/>10240 MB, 900s]
        SFN[Step Functions<br/>Express]
    end

    subgraph Storage["💾 Storage Layer"]
        S3FE[S3 Frontend<br/>Static Site]
        S3Assets[S3 Assets<br/>Video Files]
        DDB1[DynamoDB Jobs<br/>On-Demand]
        DDB2[DynamoDB Usage<br/>On-Demand]
        ECR[ECR Repository<br/>Docker Images]
    end

    subgraph Security["🔐 Security Layer"]
        Cognito[Cognito User Pool<br/>OAuth2 + JWT]
        Secrets[Secrets Manager<br/>API Keys]
        IAM[IAM Roles<br/>4 roles]
    end

    subgraph Network["🌐 Network Layer"]
        VPC[VPC 10.0.0.0/16]
        PubSub1[Public Subnet<br/>10.0.1.0/24]
        PubSub2[Public Subnet<br/>10.0.2.0/24]
        PrivSub[Private Subnet<br/>10.0.10.0/24]
        IGW[Internet Gateway]
        NAT[NAT Gateway]
        VPCE[VPC Endpoints<br/>S3, DDB, ECR]
    end

    subgraph Monitor["📊 Monitoring"]
        CW[CloudWatch Logs<br/>4 log groups]
        Insights[Container Insights]
    end

    External[Replicate AI API]

    User -->|HTTPS| CF
    GH -->|Deploy| S3FE
    GH -->|Push| ECR
    GH -->|Update| ECS

    CF -->|Origin| S3FE
    CF -->|/api/* proxy| ALB

    ALB --> ECS
    ECS --> SFN
    SFN --> L1
    SFN --> L2

    ECS --> DDB1
    ECS --> DDB2
    ECS --> S3Assets
    L1 --> S3Assets
    L2 --> S3Assets

    ECS -.->|Validate| Cognito
    User -.->|Login| Cognito

    L1 --> Secrets
    L1 -->|Generate| External

    ECS --> CW
    L1 --> CW
    L2 --> CW
    ECS -.-> Insights

    ALB --- PubSub1
    ALB --- PubSub2
    ECS --- PrivSub
    L1 --- PrivSub
    L2 --- PrivSub

    PubSub1 & PubSub2 --> IGW
    PrivSub --> NAT
    NAT --> IGW
    PrivSub -.->|Private Access| VPCE

    PubSub1 & PubSub2 & PrivSub --- VPC

    ECS -.->|Assume Role| IAM
    L1 -.->|Assume Role| IAM
    L2 -.->|Assume Role| IAM
    SFN -.->|Assume Role| IAM

    style Compute fill:#e1f5ff,stroke:#0288d1,stroke-width:2px
    style Storage fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    style Security fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    style Network fill:#e8f5e9,stroke:#388e3c,stroke-width:2px
    style Monitor fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style CDN fill:#fce4ec,stroke:#c2185b,stroke-width:2px
```

---

## Technology Stack

### Frontend Stack

```mermaid
flowchart LR
    subgraph Build["Build Tools"]
        Bun[Bun<br/>Package Manager]
        Vite[Vite 5<br/>Build Tool]
    end

    subgraph Framework["Framework"]
        React[React 18<br/>UI Library]
        Router[React Router<br/>Client Routing]
    end

    subgraph State["State Management"]
        Context[React Context<br/>Theme, Auth]
        Hooks[React Hooks<br/>useState, useEffect]
    end

    subgraph Auth["Authentication"]
        CognitoSDK[AWS Cognito SDK<br/>OAuth2 Flow]
        JWT[JWT Storage<br/>localStorage]
    end

    subgraph Deploy["Deployment"]
        S3[S3 Bucket<br/>Static Hosting]
        CloudFront[CloudFront<br/>CDN]
    end

    Bun --> Vite
    Vite --> React
    React --> Router
    React --> Context
    Context --> Hooks
    React --> CognitoSDK
    CognitoSDK --> JWT
    Vite -->|Build Output| S3
    S3 -->|Origin| CloudFront

    style Build fill:#f0f4c3
    style Framework fill:#b3e5fc
    style State fill:#c5e1a5
    style Auth fill:#f8bbd0
    style Deploy fill:#ffe0b2
```

### Backend Stack

```mermaid
flowchart LR
    subgraph Runtime["Runtime"]
        Go[Go 1.25.4<br/>Compiled Binary]
        Docker[Docker<br/>Multi-stage Build]
    end

    subgraph Framework_B["Web Framework"]
        Gin[Gin<br/>HTTP Router]
        Middleware[Middleware Stack<br/>Logger, CORS, Auth]
    end

    subgraph AWS_SDK["AWS SDKs"]
        DynamoSDK[DynamoDB SDK<br/>Jobs, Usage]
        S3SDK[S3 SDK<br/>Assets]
        SFNSDK[Step Functions SDK<br/>Orchestration]
        SecretsSDK[Secrets SDK<br/>API Keys]
    end

    subgraph Auth_B["Authentication"]
        JWT_B[JWT Validation<br/>JWKS Fetching]
        RateLimit[Rate Limiter<br/>In-Memory]
        Quota[Quota Enforcement<br/>DynamoDB]
    end

    subgraph Deploy_B["Deployment"]
        ECR_B[ECR Repository<br/>Docker Images]
        ECS_B[ECS Fargate<br/>Container Runtime]
    end

    Go --> Docker
    Docker --> Gin
    Gin --> Middleware
    Middleware --> JWT_B
    Middleware --> RateLimit
    Middleware --> Quota

    Gin --> DynamoSDK
    Gin --> S3SDK
    Gin --> SFNSDK
    Gin --> SecretsSDK

    Docker -->|Push| ECR_B
    ECR_B -->|Pull| ECS_B

    style Runtime fill:#b3e5fc
    style Framework_B fill:#c5e1a5
    style AWS_SDK fill:#ffe0b2
    style Auth_B fill:#f8bbd0
    style Deploy_B fill:#f0f4c3
```

### Serverless Stack (Lambdas)

```mermaid
flowchart LR
    subgraph Runtime_L["Runtime"]
        Node[Node.js 20.x<br/>Lambda Runtime]
    end

    subgraph Generator["Generator Lambda"]
        GenLogic[Scene Generation<br/>Logic]
        ReplicateClient[Replicate Client<br/>HTTP Calls]
    end

    subgraph Composer["Composer Lambda"]
        FFmpeg[FFmpeg Binary<br/>Video Processing]
        ComposerLogic[Stitching Logic<br/>Transitions]
    end

    subgraph Storage_L["Storage"]
        EphemeralGen[Ephemeral /tmp<br/>512 MB]
        EphemeralComp[Ephemeral /tmp<br/>10 GB]
    end

    subgraph Orchestrator["Orchestration"]
        StepFunc[Step Functions<br/>State Machine]
    end

    Node --> GenLogic
    Node --> FFmpeg
    GenLogic --> ReplicateClient
    FFmpeg --> ComposerLogic

    GenLogic --> EphemeralGen
    ComposerLogic --> EphemeralComp

    StepFunc -->|Invoke| GenLogic
    StepFunc -->|Invoke| ComposerLogic

    style Runtime_L fill:#b3e5fc
    style Generator fill:#c5e1a5
    style Composer fill:#f8bbd0
    style Storage_L fill:#ffe0b2
    style Orchestrator fill:#f0f4c3
```

---

## High Availability & Disaster Recovery

```mermaid
flowchart TB
    subgraph MultiAZ["Multi-AZ Components"]
        ALB_HA[ALB<br/>2 AZs: us-east-1a, us-east-1b]
        PubSub_HA[Public Subnets<br/>2 AZs for redundancy]
    end

    subgraph SingleAZ["Single-AZ Components"]
        PrivSub_SA[Private Subnet<br/>1 AZ: us-east-1a]
        NAT_SA[NAT Gateway<br/>us-east-1a only]
        ECS_SA[ECS Tasks<br/>Can restart in other AZ]
    end

    subgraph Serverless_HA["Serverless (Built-in HA)"]
        Lambda_HA[Lambda Functions<br/>Multi-AZ by default]
        DDB_HA[DynamoDB<br/>Multi-AZ replication]
        S3_HA[S3<br/>11 9s durability]
        CF_HA[CloudFront<br/>Global edge network]
    end

    subgraph Recovery["Recovery Mechanisms"]
        AutoScale[ECS Auto-Scaling<br/>CPU/Memory targets]
        HealthCheck[Health Checks<br/>ALB + ECS]
        Retry[Lambda Retries<br/>Exponential backoff]
        CircuitBreaker[ECS Circuit Breaker<br/>Rollback on failure]
    end

    MultiAZ -.->|High Availability| Recovery
    SingleAZ -.->|Quick Recovery| Recovery
    Serverless_HA -.->|Built-in Resilience| Recovery

    style MultiAZ fill:#c8e6c9,stroke:#388e3c,stroke-width:3px
    style SingleAZ fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style Serverless_HA fill:#b3e5fc,stroke:#0288d1,stroke-width:3px
    style Recovery fill:#f8bbd0,stroke:#c2185b,stroke-width:2px
```

---

## Key Architecture Decisions

| Decision | Why | Trade-off |
|----------|-----|-----------|
| **Hybrid ECS + Lambda** | ECS for always-on API (predictable load), Lambda for video processing (bursty, event-driven) | Increased complexity vs optimal cost |
| **Step Functions Express** | Video generation < 5 min, 50% cheaper than Standard | No long-term execution history (90 days max) |
| **Single-AZ Private Subnet** | Cost savings: 1 NAT vs 2 NATs = $32/month saved | Lower HA, but ECS can restart quickly |
| **DynamoDB On-Demand** | Unpredictable traffic for new product, zero capacity planning | Higher per-request cost vs provisioned |
| **CloudFront + ALB** | Single domain for frontend + API (no CORS complexity) | Dual routing architecture |
| **Cognito for Auth** | Managed OAuth2/OIDC, SOC2 compliant, < 50K MAU free | AWS vendor lock-in |
| **GitHub OIDC** | No long-lived AWS credentials, auto-rotation | One-time setup complexity |

---

**Related Documentation:**
- [Network Topology](./network-topology.md) - Detailed VPC architecture
- [Data Flow](./data-flow.md) - Request/response sequences
- [Infrastructure Modules](./infrastructure-modules.md) - Terraform structure
