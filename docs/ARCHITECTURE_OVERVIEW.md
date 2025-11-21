# OmniGen AI Video Generation Pipeline - Architecture Overview

## 🎬 System Overview

OmniGen is an AI-powered pharmaceutical advertisement video generation platform that transforms text prompts into professionally composed video ads with music and regulatory compliance. Built on a React/Go/AWS stack, the system orchestrates multiple AI models through a Replicate-based pipeline. The generation pipeline flows from text to script (GPT-4o) → scene-by-scene videos (Kling v2.5 Turbo Pro) → background music (Minimax Music 1.5) → final multi-track composition with real-time progress streaming.

---

## 🤖 AI Models Integration

| Model | Purpose | Key Features | Cost/Performance |
|-------|---------|--------------|------------------|
| **GPT-4o** | Script generation & image analysis | Scene-by-scene cinematography, pharma compliance (side effects, dosage), JSON-structured output | ~60s generation time |
| **Kling v2.5 Turbo Pro** | Video scene generation | 5-10s clips, visual continuity via last-frame seeding, professional cinematography | ~$0.07/second, 90-180s/scene |
| **Minimax Music 1.5** | Background music synthesis | Mood-based instrumental composition, pharma-appropriate tone | ~60s generation time |

---

## ⚡ Data Flow Architecture

```mermaid
graph TB
    A[User Input<br/>Prompt + Duration + Style + Side Effects] --> B[POST /api/v1/generate]
    B --> C[Create Job in DynamoDB<br/>Status: queued<br/>jobs table PK: userID#jobID]
    C --> D[Return job_id immediately<br/>HTTP 202 Accepted]
    D --> E[Frontend: Start SSE Connection<br/>GET /jobs/{jobID}/progress]
    C --> F[Async Goroutine Pipeline<br/>Non-blocking execution]
    
    F --> G[GPT-4o Script Generation<br/>Status: generating_script<br/>⏱️ 30-120s<br/>S3: script.json]
    G --> H[Scene 1: Kling Video<br/>+ Product Image Analysis<br/>⏱️ 90-180s<br/>S3: clips/scene_1.mp4]
    H --> I[Extract Last Frame<br/>Visual Continuity Seed<br/>⏱️ 2-5s]
    I --> J[Scene 2: Kling Video<br/>+ Last Frame Seed<br/>⏱️ 90-180s<br/>S3: clips/scene_2.mp4]
    J --> K[Scene N: Kling Video<br/>Parallel Processing<br/>S3: clips/scene_N.mp4]
    
    G --> L[Minimax Music Generation<br/>Status: generating_music<br/>⏱️ 30-90s<br/>S3: audio/background.mp3]
    
    K --> M[FFmpeg Composition<br/>Status: composing<br/>⏱️ 10-30s<br/>Merge video + audio tracks]
    L --> M
    
    M --> N[Upload Final to S3<br/>final/output.mp4<br/>Generate Presigned URL 1hr]
    N --> O[DynamoDB Update<br/>Status: completed<br/>finalVideoURL + metadata]
    
    E --> P[SSE Progress Updates<br/>Push every 1s<br/>JSON: status, stage, percent]
    P --> Q[Frontend Progress Display<br/>Real-time UI updates<br/>Auto-reconnect on disconnect]
    
    G -.Update DynamoDB.-> P
    H -.Update DynamoDB.-> P
    J -.Update DynamoDB.-> P
    L -.Update DynamoDB.-> P
    M -.Update DynamoDB.-> P
    O -.Update DynamoDB.-> P
    
    style G fill:#4CAF50,stroke:#2E7D32,stroke-width:2px
    style H fill:#2196F3,stroke:#1565C0,stroke-width:2px
    style J fill:#2196F3,stroke:#1565C0,stroke-width:2px
    style L fill:#FF9800,stroke:#E65100,stroke-width:2px
    style M fill:#9C27B0,stroke:#6A1B9A,stroke-width:2px
    style O fill:#4CAF50,stroke:#2E7D32,stroke-width:3px
    style P fill:#e1f5ff,stroke:#0066cc,stroke-width:2px
```

**S3 Structure**: `s3://omnigen-assets/users/{userID}/jobs/{jobID}/` with subdirs: `clips/`, `audio/`, `final/`, `script.json`  
**DynamoDB Updates**: Each pipeline stage updates job status → triggers SSE push → frontend receives update within 1s

---

## 🏗️ Key Components

### Backend (Go)
- **ParserService**: Extracts scenes/assets from GPT-4o JSON responses, handles duration/continuity metadata
- **AssetService**: Manages S3 uploads/downloads, presigned URL generation (1hr expiration)
- **Adapters**: `GPT4oAdapter` (script gen), `KlingAdapter` (video gen with last-frame seeding), `MinimaxAdapter` (music gen)
- **SSE Progress Handler**: Real-time job status streaming to frontend (`/jobs/{id}/progress`)
- **Repository Layer**: DynamoDB operations (jobs, users, usage tracking), S3 asset management
- **Concurrency**: Goroutine-based parallel scene generation with semaphore limiting

### Frontend (React)
- **Create Page**: Prompt input, preset templates (30s/60s ads), job submission
- **Workspace Editor**: Multi-track timeline (video/music/text tracks), scene reordering, asset preview
- **Progress Dashboard**: SSE-powered real-time updates with auto-reconnect (exponential backoff)
- **Video Library**: Job history, asset browsing, download management
- **Auth Flow**: Cognito-based JWT authentication with protected routes

### Infrastructure (AWS)
- **DynamoDB**: Jobs table (job metadata, status, progress), Users table, Usage tracking
- **S3**: Asset storage (videos, music, scripts), presigned URL delivery
- **ECS Fargate**: Containerized Go backend with auto-scaling
- **Cognito**: User authentication, JWT token management
- **CloudFront**: CDN for frontend delivery

---

## 🚀 Production Features

<table>
<tr>
<td valign="top" width="50%">

**AI & Media Processing**
- Visual continuity via last-frame seeding between scenes
- Pharmaceutical compliance: side effects, dosage, regulatory language
- Scene-by-scene cinematography control (camera angles, mood)
- JSON-structured script parsing with error recovery

</td>
<td valign="top" width="50%">

**Infrastructure & UX**
- Presigned URL management (1hr expiration, auto-refresh)
- Real-time SSE progress streaming (8 stages: queued → completed)
- Multi-track timeline editor with drag-drop reordering
- JWT authentication with refresh tokens
- Concurrent goroutine processing (non-blocking pipeline)
- Auto-reconnect SSE with exponential backoff

</td>
</tr>
</table>

---

## 📊 Performance Metrics

| Metric | Value | Details |
|--------|-------|---------|
| **Total Generation Time** | 3-6 minutes | For 30-second pharmaceutical ad (3-5 scenes) |
| **Script Generation** | ~60 seconds | GPT-4o processing time |
| **Per-Scene Video** | 90-180 seconds | Kling v2.5 Turbo Pro, parallel processing |
| **Music Generation** | ~60 seconds | Minimax Music 1.5, processed in parallel with scenes |
| **Concurrency** | Parallel goroutines | All scenes generated simultaneously (non-blocking) |
| **SSE Reconnect** | Exponential backoff | 1s → 2s → 4s → 8s (max 30s) |
| **Asset URL Expiration** | 1 hour | Presigned S3 URLs, auto-refresh on frontend |
| **Job Stages** | 8 tracked states | `queued` → `generating_script` → `generating_videos` → `generating_music` → `composing` → `completed` → `failed` → `cancelled` |

---

## 📋 Quick Reference

**Tech Stack**: React 18 • Go 1.21+ • DynamoDB • S3 • ECS Fargate • Cognito • Replicate API  

**Models**: GPT-4o (`openai/gpt-4o`) • Kling v2.5 Turbo Pro • Minimax Music 1.5  

**API**: Gorilla Mux framework • SSE streaming • JWT auth • Presigned URLs (1hr expiration)  

**Pipeline**: ~3-6 min total • Async goroutines • Real-time progress • Visual continuity via last-frame seeding  

**Storage**: S3 structure: `users/{userID}/jobs/{jobID}/clips/|audio/|final/` • DynamoDB: jobs + users + usage tables  

**Features**: Pharma compliance (side effects, dosage) • Multi-track timeline editor • Auto URL refresh • Dark mode UI • Scene reordering • Concurrent processing

