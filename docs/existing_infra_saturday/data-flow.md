# Data Flow Diagrams

> End-to-end request/response flows through the OmniGen platform

## Overview

This document maps complete data flows from user browser through CloudFront, ALB, ECS, and backend services. All diagrams are derived from actual infrastructure code and backend implementation.

**Key Flows Documented:**
1. API Request Flow (REST endpoints)
2. Video Generation Flow (Step Functions orchestration)
3. Job Status Polling (WebSocket alternative)
4. Video Playback Flow (CloudFront + S3)
5. Middleware Stack (Go API request pipeline)

---

## API Request Flow

This sequence shows a standard authenticated API request from frontend to backend.

```mermaid
sequenceDiagram
    actor User
    participant CF as CloudFront<br/>CDN
    participant ALB as Application<br/>Load Balancer
    participant ECS as ECS Fargate<br/>Go API
    participant Cognito as Cognito<br/>User Pool
    participant DDB as DynamoDB<br/>Jobs/Usage
    participant S3 as S3<br/>Assets Bucket

    User->>CF: GET /api/v1/jobs<br/>Authorization: Bearer {JWT}
    CF->>ALB: Proxy /api/* to ALB origin
    ALB->>ECS: HTTP GET :8080/api/v1/jobs<br/>+ Headers

    Note over ECS: Middleware Stack:<br/>Logger → CORS → JWT → RateLimit → Quota

    ECS->>Cognito: Validate JWT via JWKS<br/>https://cognito-idp.{region}.amazonaws.com/{pool}/.well-known/jwks.json
    Cognito-->>ECS: JWT Valid + Claims (sub, email)

    ECS->>DDB: Query omnigen-usage<br/>userId = {sub}
    DDB-->>ECS: Usage data (quota remaining)

    alt Quota Exceeded
        ECS-->>ALB: 429 Too Many Requests<br/>{"error": "Monthly quota exceeded"}
        ALB-->>CF: 429 Response
        CF-->>User: 429 Too Many Requests
    else Quota OK
        ECS->>DDB: Query omnigen-jobs<br/>userId = {sub}
        DDB-->>ECS: Job records

        ECS->>S3: GetObject (thumbnails)<br/>omnigen-assets/{jobId}/thumbnail.jpg
        S3-->>ECS: Presigned URL (1 hour TTL)

        ECS-->>ALB: 200 OK<br/>{"jobs": [...], "thumbnails": [...]}
        ALB-->>CF: 200 Response + Cache-Control
        CF-->>User: 200 OK (JSON payload)
    end

    Note over CF: Cache TTL: 0s for API responses<br/>(no caching for dynamic content)
```

**Key Points:**
- **CloudFront Routing:** `/api/*` paths bypass CloudFront cache and proxy to ALB
- **JWT Validation:** JWKS fetched on startup, cached in-memory for 1 hour
- **Rate Limiting:** 100 requests per user per minute (in-memory counter)
- **Quota Enforcement:** DynamoDB query on every request (on-demand billing)
- **S3 URLs:** Presigned URLs with 1-hour expiration for security

---

## Video Generation Flow

This sequence shows the complete video generation pipeline from API request to final video output.

```mermaid
sequenceDiagram
    actor User
    participant ECS as ECS Fargate<br/>Go API
    participant DDB_Jobs as DynamoDB<br/>omnigen-jobs
    participant DDB_Usage as DynamoDB<br/>omnigen-usage
    participant SFN as Step Functions<br/>Express Workflow
    participant L1 as Lambda Generator<br/>Node.js 20 (2 GB)
    participant Secrets as Secrets Manager<br/>Replicate Key
    participant Replicate as Replicate AI<br/>External API
    participant S3 as S3 Assets<br/>Video Storage
    participant L2 as Lambda Composer<br/>Node.js 20 (10 GB)

    User->>ECS: POST /api/v1/jobs<br/>{"prompt": "...", "category": "music-video"}

    Note over ECS: Auth Middleware:<br/>JWT → RateLimit → Quota

    ECS->>DDB_Usage: Check quota<br/>userId = {sub}
    DDB_Usage-->>ECS: Remaining: 8 videos

    ECS->>DDB_Jobs: PutItem<br/>{jobId, userId, status: "pending", prompt}
    DDB_Jobs-->>ECS: Item created

    ECS->>SFN: StartExecution<br/>input: {jobId, prompt, category}
    SFN-->>ECS: executionArn

    ECS-->>User: 202 Accepted<br/>{"jobId": "...", "status": "pending"}

    Note over SFN,L1: Step Functions State Machine<br/>Express Workflow (5 min max)

    SFN->>L1: Invoke (Scene Generator)<br/>Payload: {jobId, prompt, sceneIndex: 1}

    L1->>Secrets: GetSecretValue<br/>omnigen/replicate-api-key
    Secrets-->>L1: API Key (cached 1 hour)

    L1->>Replicate: POST /v1/predictions<br/>Model: stable-diffusion-xl<br/>Prompt: "{scene description}"
    Replicate-->>L1: {"id": "...", "status": "processing"}

    loop Poll for completion (max 60s)
        L1->>Replicate: GET /v1/predictions/{id}
        Replicate-->>L1: {"status": "succeeded", "output": "https://..."}
    end

    L1->>S3: PutObject<br/>omnigen-assets/{jobId}/scene-001.mp4
    S3-->>L1: Upload complete

    L1->>DDB_Jobs: UpdateItem<br/>{jobId, sceneStatus: "scene-001: complete"}
    DDB_Jobs-->>L1: Updated

    L1-->>SFN: Return {"sceneUrl": "s3://..."}

    Note over SFN: Parallel execution:<br/>Scenes 2, 3, 4, 5 (Map state)

    SFN->>L2: Invoke (Video Composer)<br/>Payload: {jobId, sceneUrls: [...]}

    L2->>S3: GetObject (all scenes)<br/>Download to /tmp/
    S3-->>L2: Scene files (5 x ~50 MB)

    Note over L2: FFmpeg Processing:<br/>1. Concatenate scenes<br/>2. Add transitions<br/>3. Sync audio<br/>4. Encode to H.264

    L2->>L2: ffmpeg -i concat.txt<br/>-vf "fade,scale=1920:1080"<br/>-c:v libx264 -preset fast<br/>output.mp4

    L2->>S3: PutObject<br/>omnigen-assets/{jobId}/final.mp4
    S3-->>L2: Upload complete (200 MB)

    L2->>DDB_Jobs: UpdateItem<br/>{jobId, status: "completed", videoUrl: "s3://..."}
    DDB_Jobs-->>L2: Updated

    L2->>DDB_Usage: UpdateItem<br/>Increment videosGenerated counter
    DDB_Usage-->>L2: Updated

    L2-->>SFN: Return {"status": "success"}

    SFN-->>DDB_Jobs: Final status update (optional)

    Note over User: Frontend polls GET /api/v1/jobs/{jobId}<br/>every 5 seconds until status = "completed"
```

**Performance Metrics:**
- **Scene Generation:** 30-60s per scene (Replicate API latency)
- **Parallel Scenes:** 5 scenes in ~60s (Map state parallelism)
- **Video Composition:** 30-90s (FFmpeg encoding)
- **Total Time:** 2-3 minutes for 30-second video

**Cost Breakdown (30-second video):**
- Step Functions Express: $0.000025 per execution = $0.000025
- Lambda Generator: 5 invocations x 60s x 2048 MB = $0.01
- Lambda Composer: 1 invocation x 90s x 10240 MB = $0.01
- DynamoDB: 10 write requests = $0.0000125
- S3: 200 MB storage + PUT = $0.005
- Replicate API: $1.30 (5 scenes x $0.26)
- **Total: $1.32/video**

---

## Job Status Polling

Lightweight polling mechanism for frontend to track job progress (WebSocket alternative).

```mermaid
sequenceDiagram
    actor User
    participant React as React Frontend
    participant CF as CloudFront
    participant ECS as ECS API
    participant DDB as DynamoDB<br/>omnigen-jobs

    User->>React: Click "Generate Video"
    React->>CF: POST /api/v1/jobs<br/>{prompt, category}
    CF->>ECS: Proxy to backend
    ECS-->>CF: 202 Accepted<br/>{"jobId": "abc123", "status": "pending"}
    CF-->>React: Response

    Note over React: Store jobId in state<br/>Start polling loop

    loop Poll every 5 seconds
        React->>CF: GET /api/v1/jobs/abc123
        CF->>ECS: Proxy
        ECS->>DDB: GetItem (jobId = abc123)
        DDB-->>ECS: {status: "processing", progress: 60%}
        ECS-->>CF: 200 OK<br/>{"status": "processing", "progress": 60%}
        CF-->>React: Response

        Note over React: Update UI progress bar

        alt Status = "completed"
            React->>React: Stop polling
            React->>CF: GET /api/v1/jobs/abc123/video
            CF->>ECS: Proxy
            ECS-->>CF: 302 Redirect<br/>Location: https://cloudfront.net/{jobId}/final.mp4
            CF-->>React: Redirect to video
            React->>React: Show video player
        else Status = "failed"
            React->>React: Stop polling
            React->>React: Show error message
        end
    end
```

**Polling Configuration:**
- **Interval:** 5 seconds (balance between UX and API load)
- **Timeout:** 5 minutes (Step Functions max execution time)
- **Retry Logic:** Exponential backoff on 5xx errors
- **Cache Headers:** `Cache-Control: no-store, must-revalidate`

**Alternative (Future):** WebSocket via API Gateway for real-time updates

---

## Video Playback Flow

This sequence shows how generated videos are delivered to end users via CloudFront.

```mermaid
sequenceDiagram
    actor User
    participant React as React Frontend
    participant CF as CloudFront<br/>CDN
    participant S3 as S3 Assets<br/>Origin

    User->>React: Click video thumbnail
    React->>CF: GET /assets/{jobId}/final.mp4

    alt CloudFront Cache HIT
        Note over CF: Video cached at edge location<br/>(24-hour TTL)
        CF-->>React: 200 OK (from cache)<br/>Content-Type: video/mp4<br/>Range: bytes 0-1048575
        React->>React: Video player starts buffering
    else CloudFront Cache MISS
        CF->>S3: GetObject<br/>omnigen-assets/{jobId}/final.mp4
        S3-->>CF: 200 OK (200 MB video file)
        Note over CF: Cache video at edge<br/>TTL: 86400s (24 hours)
        CF-->>React: 200 OK<br/>Content-Type: video/mp4<br/>Range: bytes 0-1048575
    end

    Note over User,React: HTML5 Video Player:<br/>Adaptive streaming with range requests

    User->>React: Seek to 0:45
    React->>CF: GET /assets/{jobId}/final.mp4<br/>Range: bytes 15728640-16777215
    CF-->>React: 206 Partial Content<br/>(byte range from cache)

    User->>React: Click "Download"
    React->>CF: GET /assets/{jobId}/final.mp4<br/>Full file download
    CF-->>React: 200 OK (entire file)<br/>Content-Disposition: attachment
```

**CloudFront Configuration:**
- **Cache Behavior:** `/assets/*` → S3 origin
- **TTL:** 24 hours for videos (immutable content)
- **Compression:** Gzip disabled for video files (already compressed)
- **Range Requests:** Enabled for video seeking
- **CORS:** Allowed for cross-origin requests

**Performance:**
- **First Byte:** <100ms (CloudFront edge)
- **Throughput:** 10-100 Mbps (depends on user location)
- **Availability:** 99.9% (CloudFront SLA)

---

## Middleware Stack Flow

This flowchart shows the Go API middleware pipeline that every request passes through.

```mermaid
flowchart TB
    Start([Incoming HTTP Request])

    Start --> Logger[Logger Middleware<br/>Log request method, path, IP]

    Logger --> CORS{CORS Middleware<br/>Preflight?}
    CORS -->|OPTIONS| CORSHeaders[Set CORS Headers<br/>Access-Control-Allow-Origin: *<br/>Access-Control-Allow-Methods: GET,POST,PUT,DELETE]
    CORSHeaders --> Return200([Return 200 OK])

    CORS -->|GET/POST/PUT/DELETE| JWT{JWT Auth Middleware<br/>Bearer token present?}

    JWT -->|No token| Return401A([Return 401 Unauthorized<br/>WWW-Authenticate: Bearer])

    JWT -->|Token present| ValidateJWT[Validate JWT<br/>1. Fetch JWKS from Cognito<br/>2. Verify signature<br/>3. Check expiration<br/>4. Extract claims]

    ValidateJWT -->|Invalid/Expired| Return401B([Return 401 Unauthorized<br/>Error: Invalid token])

    ValidateJWT -->|Valid| SetContext[Set Gin Context<br/>c.Set'userId', claims.Sub<br/>c.Set'email', claims.Email]

    SetContext --> RateLimit{Rate Limit Middleware<br/>Check in-memory counter<br/>Key: userId}

    RateLimit -->|> 100 req/min| Return429A([Return 429 Too Many Requests<br/>Retry-After: 60])

    RateLimit -->|<= 100 req/min| IncrCounter[Increment counter<br/>Sliding window: 1 minute]

    IncrCounter --> Quota{Quota Middleware<br/>Check DynamoDB<br/>omnigen-usage table}

    Quota -->|Query Error| Return500([Return 500 Internal Server Error<br/>Error: Database unavailable])

    Quota -->|Quota Exceeded| Return429B([Return 429 Too Many Requests<br/>Error: Monthly quota exceeded])

    Quota -->|Quota OK| Handler[Route Handler<br/>jobs.GetJobs<br/>jobs.CreateJob<br/>jobs.GetJobStatus]

    Handler -->|Success| ResponseLog[Logger Middleware<br/>Log status code, duration]
    ResponseLog --> Return200B([Return 200/201/202<br/>JSON Response])

    Handler -->|Business Logic Error| Return400([Return 400/404/500<br/>JSON Error])

    style Start fill:#e1f5ff,stroke:#0288d1,stroke-width:2px
    style Return200 fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style Return200B fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style Return401A fill:#ffcdd2,stroke:#c62828,stroke-width:2px
    style Return401B fill:#ffcdd2,stroke:#c62828,stroke-width:2px
    style Return429A fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style Return429B fill:#fff9c4,stroke:#f9a825,stroke-width:2px
    style Return400 fill:#ffcdd2,stroke:#c62828,stroke-width:2px
    style Return500 fill:#ffcdd2,stroke:#c62828,stroke-width:2px
```

**Middleware Execution Order:**

1. **Logger (Pre-Handler)**
   - Logs: `[GIN] 2024-12-15 10:30:45 | 200 | 45ms | 192.168.1.1 | GET /api/v1/jobs`
   - Writes to: CloudWatch Logs `/ecs/omnigen`

2. **CORS (Preflight Handler)**
   - Handles: `OPTIONS` requests from browser
   - Headers: `Access-Control-Allow-Origin: *` (MVP - should restrict to CloudFront domain in prod)

3. **JWT Auth (Security)**
   - JWKS URL: `https://cognito-idp.us-east-1.amazonaws.com/{userPoolId}/.well-known/jwks.json`
   - Cache: JWKS cached in-memory for 1 hour
   - Claims: `sub` (userId), `email`, `cognito:username`

4. **Rate Limit (Abuse Prevention)**
   - Algorithm: Sliding window counter (in-memory map)
   - Limit: 100 requests per user per minute
   - Reset: Automatic after 1 minute
   - Storage: In-memory (resets on ECS task restart)

5. **Quota Enforcement (Billing)**
   - Storage: DynamoDB `omnigen-usage` table
   - Quota: 10 videos per user per month (MVP)
   - Counter: Incremented on successful video completion
   - Reset: Monthly cron job (Lambda function)

6. **Handler (Business Logic)**
   - Routes: Gin router with groups
   - Validation: Request body validation with `binding:"required"`
   - Errors: Standardized JSON error responses

7. **Logger (Post-Handler)**
   - Logs: Response status, duration, error messages
   - Metrics: CloudWatch custom metrics (request count, latency)

---

## Error Handling

### Error Response Format

All API errors follow this JSON structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE_CONSTANT",
  "details": {
    "field": "Additional context"
  },
  "timestamp": "2024-12-15T10:30:45Z"
}
```

### Common Error Scenarios

```mermaid
flowchart LR
    subgraph Client[\"Client Errors (4xx)\"]
        E400[400 Bad Request<br/>Invalid JSON, missing fields]
        E401[401 Unauthorized<br/>Missing/invalid JWT]
        E403[403 Forbidden<br/>Valid JWT, insufficient permissions]
        E404[404 Not Found<br/>Job ID not found]
        E429[429 Too Many Requests<br/>Rate limit or quota exceeded]
    end

    subgraph Server[\"Server Errors (5xx)\"]
        E500[500 Internal Server Error<br/>Unhandled exception]
        E502[502 Bad Gateway<br/>ECS task unhealthy]
        E503[503 Service Unavailable<br/>DynamoDB throttling]
        E504[504 Gateway Timeout<br/>Lambda timeout > 900s]
    end

    style Client fill:#fff9c4,stroke:#f9a825
    style Server fill:#ffcdd2,stroke:#c62828
```

---

## Data Flow Summary

| Flow Type | Latency | Cost (per request) | Caching |
|-----------|---------|-------------------|---------|
| **API Request** | 50-200ms | $0.0001 | No (dynamic) |
| **Video Generation** | 2-3 min | $1.32 | N/A |
| **Status Polling** | 50-100ms | $0.00005 | No |
| **Video Playback** | <100ms | $0.01/GB | Yes (24h) |
| **Middleware** | 5-10ms | $0 (included) | JWKS (1h) |

**Bottlenecks:**
1. **Replicate API:** 30-60s per scene (external dependency)
2. **Lambda Cold Start:** 1-3s for first invocation (mitigated with provisioned concurrency)
3. **FFmpeg Encoding:** 30-90s (CPU-bound, cannot optimize further without hardware encoding)

**Optimizations Applied:**
- JWKS caching (1 hour) to reduce Cognito API calls
- Presigned S3 URLs (1 hour TTL) to offload traffic from API
- CloudFront caching for static assets and videos
- DynamoDB on-demand billing to handle bursty traffic
- Step Functions Express for cost-effective orchestration

---

**Related Documentation:**
- [Architecture Overview](./architecture-overview.md) - High-level system design
- [Network Topology](./network-topology.md) - VPC and routing details
- [Authentication Flow](./authentication-flow.md) - Cognito OAuth2/JWT deep dive
- [Video Workflow](./video-workflow.md) - Step Functions state machine
