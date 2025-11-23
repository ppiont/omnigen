# PharmaGen Video Generation Pipeline

## Visual Overview

```
POST /api/v1/generate
        │
        ▼
┌───────────────────┐
│   Create Job      │──────────────▶ HTTP 202 + job_id
│   (DynamoDB)      │                (instant response)
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Acquire Semaphore │
│   (max 10 jobs)   │
└───────────────────┘
        │
        ▼
┌───────────────────┐     ┌─────────────────────────────────────┐
│ 1. SCRIPT GEN     │     │  GPT-4o via Replicate               │
│                   │────▶│  Generates structured scene scripts │
└───────────────────┘     └─────────────────────────────────────┘
        │
        ▼
┌───────────────────┐     ┌─────────────────────────────────────┐
│ 2. NARRATOR TTS   │     │  OpenAI tts-1                       │
│    (optional)     │────▶│  Variable speed for side effects    │
└───────────────────┘     └─────────────────────────────────────┘
        │
        ▼
┌───────────────────┐     ┌─────────────────────────────────────┐
│ 3. VIDEO CLIPS    │     │  Veo 3.1 via Replicate              │
│                   │────▶│  Sequential: last frame → next clip │
└───────────────────┘     └─────────────────────────────────────┘
        │
        ▼
┌───────────────────┐     ┌─────────────────────────────────────┐
│ 4. AUDIO GEN      │     │  Minimax via Replicate              │
│                   │────▶│  Background music                   │
└───────────────────┘     └─────────────────────────────────────┘
        │
        ▼
┌───────────────────┐     ┌─────────────────────────────────────┐
│ 5. COMPOSITION    │     │  ffmpeg: concat + text overlay      │
│                   │────▶│  Upload to S3                       │
└───────────────────┘     └─────────────────────────────────────┘
        │
        ▼
┌───────────────────┐
│ 6. COMPLETE       │──────────────▶ video_url available
│   (DynamoDB)      │
└───────────────────┘
```

## Infrastructure

```
                        Internet
                            │
                            ▼
                    ┌──────────────┐
                    │  CloudFront  │
                    └──────────────┘
                      /          \
                     /            \
                    ▼              ▼
            ┌──────────┐    ┌──────────┐
            │    S3    │    │   ALB    │
            │ Frontend │    └──────────┘
            └──────────┘          │
                                  ▼
                      ┌────────────────────┐
                      │    ECS Fargate     │
                      │    (1-10 tasks)    │
                      │   4 vCPU / 16 GB   │
                      └────────────────────┘
                        /      │       \
                       /       │        \
                      ▼        ▼         ▼
               ┌────────┐ ┌─────────┐ ┌─────────┐
               │   S3   │ │DynamoDB │ │ Secrets │
               │ Assets │ │  Jobs   │ │ Manager │
               └────────┘ └─────────┘ └─────────┘
```

## Cost (30s video)

```
Script (GPT-4o)     $0.20
Video (Veo 3.1)     $2.80  ← 4 clips × 8s × $0.07/s
Audio (Minimax)     $0.50
TTS (OpenAI)        $0.10
                   ──────
Total              ~$3.60
```

---

## Summary

PharmaGen uses a fully asynchronous video generation pipeline deployed on AWS ECS Fargate. The pipeline processes user prompts through 6 sequential stages: script generation (GPT-4o), optional narrator voiceover (OpenAI TTS), sequential video clip generation (Veo 3.1), audio generation (Minimax), and final composition (ffmpeg). The infrastructure uses Terraform modules with auto-scaling (1-10 tasks), CloudFront CDN, S3 storage with lifecycle policies, and DynamoDB for state management.

---

## 1. Pipeline Flow

### Request Entry Point
- **Endpoint**: `POST /api/v1/generate`
- **Handler**: `backend/internal/api/handlers/generate.go:110`
- **Response**: HTTP 202 Accepted with `job_id` (instant, <100ms)

### 6-Stage Pipeline (`generate_async.go:218-572`)

| Stage | Description | External API | Timeout |
|-------|-------------|--------------|---------|
| 1. Script Generation | GPT-4o generates structured scene scripts | Replicate | 5 min |
| 2. Narrator Voiceover | OpenAI TTS (pharmaceutical ads only) | OpenAI | 60s + retries |
| 3. Video Clips | Sequential per-scene video generation | Replicate (Veo 3.1) | 10 min/clip |
| 4. Audio Generation | Background music generation | Replicate (Minimax) | 5 min |
| 5. Composition | ffmpeg concatenation + text overlay | Local | Variable |
| 6. Completion | DynamoDB status update | - | - |

### Visual Coherence Strategy
- Last frame of Scene N becomes start image for Scene N+1
- Style reference images analyzed by GPT-4o Vision and appended to all scene prompts
- Consistent visual parameters (lighting, color grade, mood) enforced via structured prompts

---

## 2. Concurrency Model

### Job-Level Concurrency
- **Semaphore**: Limits to 10 concurrent jobs (`MaxConcurrentGenerations`)
- **Implementation**: Channel-based semaphore (`internal/concurrency/semaphore.go`)
- **Goroutines**: Each job runs in its own goroutine

### Scene-Level Processing
- **Sequential**: Scenes processed one-at-a-time within a job
- **Reason**: Enables visual continuity via last-frame passing

### External API Polling
- All Replicate APIs use synchronous polling loops
- Poll interval: 5 seconds
- Context-aware cancellation supported

---

## 3. Infrastructure Overview

### Compute (ECS Fargate)

| Property | Value |
|----------|-------|
| CPU | 4096 (4 vCPU) |
| Memory | 16384 MB (16 GB) |
| Architecture | ARM64 (Graviton) |
| Min Tasks | 1 |
| Max Tasks | 10 |
| Container Port | 8080 |

### Auto-Scaling
- **CPU Target**: 70% utilization
- **Memory Target**: 80% utilization
- **Scale-out Cooldown**: 60 seconds
- **Scale-in Cooldown**: 300 seconds

### Networking
- **VPC**: 10.0.0.0/16
- **Public Subnets**: 2 (multi-AZ for ALB)
- **Private Subnet**: 1 (ECS tasks)
- **NAT Gateway**: Yes (for outbound internet)
- **VPC Endpoints**: S3, DynamoDB (gateway), ECR API/Docker (interface)

### Storage

**S3 Assets Bucket**:
- Lifecycle: Standard → IA (30d) → Glacier (90d) → Delete (365d)
- Versioning: Enabled
- Encryption: AES256 SSE

**DynamoDB Tables**:
- `omnigen-jobs`: Job state, GSI on user_id+created_at
- `omnigen-usage`: Usage tracking per user/period
- Billing: PAY_PER_REQUEST (on-demand)

### CDN (CloudFront)
- Frontend SPA served from S3
- API routes (`/api/*`) forwarded to ALB
- Custom error pages for SPA routing (403/404 → index.html)

---

## 4. External API Integrations

| Service | Provider | Model | Cost |
|---------|----------|-------|------|
| Script Generation | Replicate | `openai/gpt-4o` | ~$0.10-0.30/script |
| Video Generation | Replicate | `google/veo-3.1` | ~$0.07/second |
| Audio Generation | Replicate | `minimax/music-1.5` | ~$0.50/track |
| Text-to-Speech | OpenAI Direct | `tts-1` | Standard OpenAI pricing |

**Total Cost**: ~$3.70-4.80 per 30-second video

---

## 5. Secrets Management

### Flow
```
AWS Secrets Manager → Terraform (ARN reference) → ECS Task Definition (env var) → Application (SecretsService)
```

### Secrets
| Secret | Purpose |
|--------|---------|
| `omnigen/replicate-api-key` | Replicate API (GPT-4o, Veo, Minimax) |
| `omnigen/openai-api-key` | OpenAI TTS API |

### Local Development
- `REPLICATE_API_KEY` env var bypasses Secrets Manager
- `OPENAI_API_KEY` env var bypasses Secrets Manager

---

## 6. Error Handling

### Failure Recovery
- Stage-specific error messages for user feedback
- S3 asset cleanup on job failure
- Panic recovery with job status update
- Retry logic: TTS adapter (3 attempts with exponential backoff)

### Monitoring
- CloudWatch Logs: 7-day retention
- Container Insights: Enabled
- Health check: `/health` endpoint

---

## 7. S3 Asset Structure

```
users/{userID}/jobs/{jobID}/
├── clips/
│   └── scene-001.mp4, scene-002.mp4, ...
├── thumbnails/
│   └── scene-001.jpg, job-thumbnail.jpg, ...
├── audio/
│   └── background-music.mp3, narrator-voiceover.mp3
└── final/
    └── video.mp4
```

---

## Code References

### Backend
- `backend/internal/api/handlers/generate.go:110` - Request handler
- `backend/internal/api/handlers/generate_async.go:218` - Pipeline orchestrator
- `backend/internal/adapters/gpt4o_adapter.go:72` - Script generation
- `backend/internal/adapters/kling_adapter.go:58` - Video generation (Veo)
- `backend/internal/adapters/minimax_adapter.go:70` - Audio generation
- `backend/internal/adapters/tts_adapter.go:74` - TTS generation
- `backend/internal/service/secrets.go:70` - Secrets retrieval

### Infrastructure
- `infrastructure/modules/compute/task-definition.tf` - ECS task config
- `infrastructure/modules/compute/autoscaling.tf` - Scaling policies
- `infrastructure/modules/iam/ecs-roles.tf` - IAM permissions
- `infrastructure/modules/storage/dynamodb.tf` - Database tables
- `infrastructure/modules/cdn/main.tf` - CloudFront distribution

---

## Architecture Documentation

### Design Patterns
- **Adapter Pattern**: Video/audio adapters abstract external AI APIs
- **Factory Pattern**: AdapterFactory creates appropriate video adapter
- **Semaphore Pattern**: Controls concurrent job execution
