# Network Topology

> Detailed VPC architecture, subnets, security groups, routing, and VPC endpoints

## VPC Architecture Overview

```mermaid
%%{init: {'theme':'base', 'themeVariables': {'fontSize':'13px'}}}%%
flowchart TB
    Internet([Internet])

    subgraph AWS["AWS Cloud - us-east-1"]
        IGW[Internet Gateway]

        subgraph VPC["VPC: omnigen-vpc (10.0.0.0/16)"]

            subgraph AZ1["Availability Zone: us-east-1a"]
                subgraph PubSub1["Public Subnet (10.0.1.0/24)"]
                    ALB1[Application Load Balancer<br/>Primary]
                    NAT[NAT Gateway<br/>+ Elastic IP]
                end

                subgraph PrivSub["Private Subnet (10.0.10.0/24)"]
                    ECS[ECS Fargate Tasks<br/>Go API]
                    Lambda1[Lambda: Generator<br/>2048 MB]
                    Lambda2[Lambda: Composer<br/>10240 MB]
                end
            end

            subgraph AZ2["Availability Zone: us-east-1b"]
                subgraph PubSub2["Public Subnet (10.0.2.0/24)"]
                    ALB2[Application Load Balancer<br/>Secondary]
                end
            end

            subgraph Endpoints["VPC Endpoints (Private)"]
                VPCE_S3[S3 Gateway Endpoint<br/>Free]
                VPCE_DDB[DynamoDB Gateway Endpoint<br/>Free]
                VPCE_ECR_API[ECR API Interface Endpoint<br/>$0.01/hr]
                VPCE_ECR_DKR[ECR Docker Interface Endpoint<br/>$0.01/hr]
            end

            subgraph RouteTables["Route Tables"]
                RT_Pub[Public Route Table<br/>0.0.0.0/0 → IGW]
                RT_Priv[Private Route Table<br/>0.0.0.0/0 → NAT<br/>S3/DDB → VPC Endpoints]
            end
        end

        subgraph Services["AWS Services (Public)"]
            S3[S3 Buckets<br/>Assets + Frontend]
            DDB[DynamoDB Tables<br/>Jobs + Usage]
            ECR[ECR Registry<br/>Docker Images]
        end
    end

    Internet <-->|HTTPS| IGW
    IGW <--> PubSub1
    IGW <--> PubSub2

    PubSub1 <-->|HTTP:8080| PrivSub
    PubSub2 <-->|HTTP:8080| PrivSub

    NAT -->|Outbound Only| IGW
    PrivSub -->|Outbound| NAT

    PrivSub -.->|Private Access| VPCE_S3
    PrivSub -.->|Private Access| VPCE_DDB
    PrivSub -.->|Private Access| VPCE_ECR_API
    PrivSub -.->|Private Access| VPCE_ECR_DKR

    VPCE_S3 -.-> S3
    VPCE_DDB -.-> DDB
    VPCE_ECR_API -.-> ECR
    VPCE_ECR_DKR -.-> ECR

    PubSub1 & PubSub2 -.->|Associated| RT_Pub
    PrivSub -.->|Associated| RT_Priv

    style VPC fill:#e3f2fd,stroke:#1976d2,stroke-width:3px
    style AZ1 fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    style AZ2 fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    style PubSub1 fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style PubSub2 fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style PrivSub fill:#ffcdd2,stroke:#d32f2f,stroke-width:2px
    style Endpoints fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
```

---

## Security Groups Architecture

```mermaid
%%{init: {'theme':'base', 'themeVariables': {'fontSize':'14px'}}}%%
flowchart LR
    Internet([Internet<br/>0.0.0.0/0])

    subgraph VPC["VPC Security Groups"]
        subgraph SG_ALB["ALB Security Group"]
            ALB[Application Load Balancer]
            ALB_In["Inbound:<br/>0.0.0.0/0:80<br/>0.0.0.0/0:443"]
            ALB_Out["Outbound:<br/>ECS-SG:8080"]
        end

        subgraph SG_ECS["ECS Security Group"]
            ECS[ECS Fargate Tasks]
            ECS_In["Inbound:<br/>ALB-SG:8080"]
            ECS_Out["Outbound:<br/>0.0.0.0/0:*<br/>(Replicate API, etc.)"]
        end

        subgraph SG_Lambda["Lambda Security Group"]
            Lambda[Lambda Functions]
            Lambda_In["Inbound:<br/>None"]
            Lambda_Out["Outbound:<br/>0.0.0.0/0:*"]
        end

        subgraph SG_VPCE["VPC Endpoints SG"]
            VPCE[VPC Endpoints<br/>ECR, S3, DDB]
            VPCE_In["Inbound:<br/>ECS-SG:443<br/>Lambda-SG:443"]
            VPCE_Out["Outbound:<br/>0.0.0.0/0:*"]
        end
    end

    Replicate[Replicate AI API<br/>External]
    AWS_Services[AWS Services<br/>S3, DynamoDB, Secrets Manager]

    Internet -->|80, 443| ALB_In
    ALB_Out -->|8080| ECS_In
    ECS_Out -->|HTTPS| Replicate
    ECS_Out -->|443| VPCE_In
    Lambda_Out -->|HTTPS| Replicate
    Lambda_Out -->|443| VPCE_In
    VPCE_Out -.-> AWS_Services

    style SG_ALB fill:#bbdefb,stroke:#1976d2,stroke-width:2px
    style SG_ECS fill:#c5e1a5,stroke:#388e3c,stroke-width:2px
    style SG_Lambda fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style SG_VPCE fill:#f8bbd0,stroke:#c2185b,stroke-width:2px
```

---

## Traffic Flow Patterns

### Inbound User Traffic

```mermaid
%%{init: {'theme':'base'}}%%
sequenceDiagram
    participant User
    participant CloudFront
    participant IGW as Internet Gateway
    participant ALB as ALB (Public Subnet)
    participant ECS as ECS Task (Private Subnet)

    User->>CloudFront: HTTPS Request
    CloudFront->>IGW: Forward to ALB origin
    IGW->>ALB: Route to public subnet
    ALB->>ECS: HTTP:8080 (private IP)
    ECS-->>ALB: Response
    ALB-->>IGW: Response
    IGW-->>CloudFront: Response
    CloudFront-->>User: HTTPS Response

    Note over ALB,ECS: ALB → ECS allowed by security groups<br/>ALB-SG:8080 → ECS-SG
```

### Outbound Internet Traffic (from Private Subnet)

```mermaid
%%{init: {'theme':'base'}}%%
sequenceDiagram
    participant ECS as ECS/Lambda<br/>(Private Subnet)
    participant NAT as NAT Gateway<br/>(Public Subnet)
    participant IGW as Internet Gateway
    participant Replicate as Replicate API

    ECS->>NAT: Outbound request (private IP)
    Note over ECS,NAT: Source: 10.0.10.x<br/>Destination: Replicate IP
    NAT->>IGW: SNAT to Elastic IP
    Note over NAT,IGW: Source: EIP (public)<br/>Destination: Replicate IP
    IGW->>Replicate: Routed to internet
    Replicate-->>IGW: Response to EIP
    IGW-->>NAT: Response
    NAT-->>ECS: DNAT to private IP

    Note over ECS,Replicate: NAT provides outbound internet<br/>for Replicate AI API calls
```

### VPC Endpoint Traffic (Cost-Optimized)

```mermaid
%%{init: {'theme':'base'}}%%
sequenceDiagram
    participant ECS as ECS/Lambda<br/>(Private Subnet)
    participant VPCE as VPC Endpoints<br/>(Gateway/Interface)
    participant S3 as S3 Buckets
    participant DDB as DynamoDB
    participant ECR as ECR Registry

    Note over ECS,VPCE: Gateway Endpoints (Free)
    ECS->>VPCE: Request to S3
    VPCE->>S3: Direct AWS network route
    S3-->>VPCE: Response
    VPCE-->>ECS: Response

    ECS->>VPCE: Request to DynamoDB
    VPCE->>DDB: Direct AWS network route
    DDB-->>VPCE: Response
    VPCE-->>ECS: Response

    Note over ECS,VPCE: Interface Endpoints ($0.01/hr each)
    ECS->>VPCE: Request to ECR
    VPCE->>ECR: Private connection
    ECR-->>VPCE: Docker image layers
    VPCE-->>ECS: Response

    Note over ECS,ECR: No NAT Gateway charges for<br/>S3, DynamoDB, ECR traffic
```

---

## IP Address Allocation

| Subnet | CIDR | Available IPs | Reserved | Usable | Purpose |
|--------|------|---------------|----------|--------|---------|
| **VPC** | 10.0.0.0/16 | 65,536 | 5 (AWS) | 65,531 | Total address space |
| **Public (us-east-1a)** | 10.0.1.0/24 | 256 | 5 | 251 | ALB, NAT Gateway |
| **Public (us-east-1b)** | 10.0.2.0/24 | 256 | 5 | 251 | ALB (multi-AZ) |
| **Private (us-east-1a)** | 10.0.10.0/24 | 256 | 5 | 251 | ECS tasks, Lambdas |

**Reserved IPs per Subnet (AWS):**
- `.0` - Network address
- `.1` - VPC router
- `.2` - DNS server
- `.3` - Future use
- `.255` - Broadcast address

---

## Routing Configuration

### Public Route Table

| Destination | Target | Purpose |
|-------------|--------|---------|
| 10.0.0.0/16 | local | VPC internal traffic |
| 0.0.0.0/0 | igw-xxx | Internet access |

**Associated Subnets:**
- 10.0.1.0/24 (us-east-1a public)
- 10.0.2.0/24 (us-east-1b public)

---

### Private Route Table

| Destination | Target | Purpose |
|-------------|--------|---------|
| 10.0.0.0/16 | local | VPC internal traffic |
| 0.0.0.0/0 | nat-xxx | Outbound internet via NAT |
| s3.us-east-1 | vpce-s3-xxx | S3 via gateway endpoint |
| dynamodb.us-east-1 | vpce-ddb-xxx | DynamoDB via gateway endpoint |

**Associated Subnets:**
- 10.0.10.0/24 (us-east-1a private)

---

## VPC Endpoints Details

### Gateway Endpoints (No Hourly Cost)

```mermaid
flowchart LR
    subgraph Private["Private Subnet"]
        ECS[ECS Tasks]
        Lambda[Lambdas]
    end

    subgraph Endpoints["Gateway Endpoints"]
        S3_EP[S3 Endpoint<br/>vpce-s3-xxx<br/>FREE]
        DDB_EP[DynamoDB Endpoint<br/>vpce-ddb-xxx<br/>FREE]
    end

    subgraph Services["AWS Services"]
        S3[S3 Buckets]
        DDB[DynamoDB Tables]
    end

    ECS -->|Private| S3_EP
    Lambda -->|Private| S3_EP
    ECS -->|Private| DDB_EP
    Lambda -->|Private| DDB_EP

    S3_EP -.->|AWS Network| S3
    DDB_EP -.->|AWS Network| DDB

    style Endpoints fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style Private fill:#ffcdd2,stroke:#d32f2f,stroke-width:2px
    style Services fill:#fff9c4,stroke:#f9a825,stroke-width:2px
```

**S3 Gateway Endpoint:**
- Prefix List: pl-63a5400a (us-east-1)
- Route added automatically to route table
- No data transfer charges within same region

**DynamoDB Gateway Endpoint:**
- Prefix List: pl-02cd2c6b (us-east-1)
- Route added automatically to route table
- No data transfer charges

---

### Interface Endpoints ($0.01/hour each)

```mermaid
flowchart LR
    subgraph Private["Private Subnet"]
        ECS[ECS Tasks]
    end

    subgraph Endpoints["Interface Endpoints<br/>(ENI in Private Subnet)"]
        ECR_API[ECR API Endpoint<br/>vpce-ecr-api-xxx<br/>$0.01/hr]
        ECR_DKR[ECR Docker Endpoint<br/>vpce-ecr-dkr-xxx<br/>$0.01/hr]
    end

    subgraph Services["AWS Services"]
        ECR[ECR Registry]
    end

    ECS -->|Private DNS| ECR_API
    ECS -->|Private DNS| ECR_DKR

    ECR_API -.->|PrivateLink| ECR
    ECR_DKR -.->|PrivateLink| ECR

    style Endpoints fill:#f8bbd0,stroke:#c2185b,stroke-width:2px
    style Private fill:#ffcdd2,stroke:#d32f2f,stroke-width:2px
    style Services fill:#fff9c4,stroke:#f9a825,stroke-width:2px
```

**ECR API Endpoint (com.amazonaws.us-east-1.ecr.api):**
- Private DNS: Yes
- Cost: $0.01/hour = $7.20/month
- Data transfer: $0.01/GB
- Purpose: ECR API calls (list images, describe repositories)

**ECR Docker Endpoint (com.amazonaws.us-east-1.ecr.dkr):**
- Private DNS: Yes
- Cost: $0.01/hour = $7.20/month
- Data transfer: $0.01/GB
- Purpose: Docker pull/push operations

**Total Interface Endpoint Cost:** $14.40/month

---

## Network Cost Optimization

### Current Architecture Costs

| Component | Monthly Cost | Annual Cost |
|-----------|-------------|-------------|
| NAT Gateway (730 hrs) | $32.85 | $394.20 |
| NAT Data Processing (10 GB) | $0.45 | $5.40 |
| Interface Endpoints (2 × 730 hrs) | $14.60 | $175.20 |
| **Total Networking** | **$47.90** | **$574.80** |

---

### Cost Savings from VPC Endpoints

**Without VPC Endpoints:**
- S3 traffic via NAT: 100 GB/month × $0.045/GB = $4.50
- DynamoDB traffic via NAT: 50 GB/month × $0.045/GB = $2.25
- **Total NAT savings:** $6.75/month

**With Gateway Endpoints (Free):**
- S3 traffic: $0
- DynamoDB traffic: $0
- **Savings:** $6.75/month = $81/year

**Interface Endpoint ROI:**
- ECR data transfer savings: ~5 GB/month × $0.045 = $0.23
- Cost: $14.40/month
- **Net cost:** -$14.17/month (worth it for security & performance)

---

## Network Performance

### Latency Characteristics

| Path | Typical Latency | Notes |
|------|----------------|-------|
| User → CloudFront | 10-50ms | Edge location proximity |
| CloudFront → ALB | 5-15ms | AWS network |
| ALB → ECS | <1ms | Same VPC |
| ECS → DynamoDB (VPC Endpoint) | 1-3ms | AWS PrivateLink |
| ECS → S3 (VPC Endpoint) | 1-3ms | AWS PrivateLink |
| ECS → Replicate (via NAT) | 50-200ms | Internet latency |

---

### Throughput Limits

| Component | Limit | Notes |
|-----------|-------|-------|
| NAT Gateway | 45 Gbps | Scales automatically |
| VPC Endpoints (Gateway) | High throughput | No hard limit |
| VPC Endpoints (Interface) | 10 Gbps | Per ENI |
| Internet Gateway | No limit | AWS managed |

---

## High Availability Considerations

### Current Setup

**Multi-AZ:**
- ✅ ALB spans 2 AZs (us-east-1a, us-east-1b)
- ✅ Public subnets in 2 AZs

**Single-AZ (Cost Optimization):**
- ⚠️ Private subnet in 1 AZ (us-east-1a)
- ⚠️ NAT Gateway in 1 AZ
- ⚠️ ECS tasks in 1 AZ

**Impact of AZ Failure:**
- us-east-1a failure: ECS tasks down, NAT unavailable, ~2-5 min recovery
- us-east-1b failure: ALB continues on us-east-1a, no impact

---

### Production Upgrade Path

To achieve full high availability:

1. **Add Second Private Subnet**
   - Create 10.0.11.0/24 in us-east-1b
   - Cost: $0 (subnet is free)

2. **Add Second NAT Gateway**
   - Deploy NAT in us-east-1b public subnet
   - Cost: +$32.85/month
   - Benefit: AZ-independent outbound internet

3. **Update ECS Service**
   - Change subnet configuration to include both AZs
   - ECS auto-distributes tasks across AZs
   - Cost: $0 (same number of tasks)

**Total HA Upgrade Cost:** +$32.85/month for second NAT

---

## Security Best Practices

### Implemented

✅ Private subnets for all compute (ECS, Lambda)
✅ Security groups with least-privilege rules
✅ VPC endpoints for AWS service access (no internet exposure)
✅ NAT Gateway for controlled outbound access
✅ No public IPs on ECS tasks or Lambdas
✅ CloudFront HTTPS enforcement

### Recommended Additions

- [ ] Enable VPC Flow Logs (CloudWatch Logs)
- [ ] Implement AWS WAF on ALB/CloudFront
- [ ] Add Network ACLs for defense-in-depth
- [ ] Enable GuardDuty for threat detection
- [ ] Implement AWS Shield Standard (free) or Advanced

---

**Related Documentation:**
- [Architecture Overview](./architecture-overview.md) - High-level system design
- [Data Flow](./data-flow.md) - Request/response sequences
- [Infrastructure Modules](./infrastructure-modules.md) - Terraform structure
