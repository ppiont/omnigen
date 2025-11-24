# Video Generation Pipeline - 1 Pager

## Pipeline Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        VIDEO GENERATION PIPELINE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  USER INPUT ──▶ SCRIPT GEN ──▶ VIDEO GEN ──▶ AUDIO GEN ──▶ COMPOSITION     │
│   (prompt)       (GPT-4o)      (per scene)    (parallel)     (FFmpeg)       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Pipeline Stages

| Stage | Service | Function | Output |
|-------|---------|----------|--------|
| **1. Script Generation** | GPT-4o (Replicate) | Transforms user prompt → structured script | `domain.Script` with scenes, audio spec |
| **2. Video Generation** | Veo/Kling (Replicate) | Generates video clips per scene (sequential) | Scene clips in S3 |
| **3. Audio Generation** | Minimax + OpenAI TTS | Background music + narrator voiceover (parallel) | Audio files in S3 |
| **4. Composition** | FFmpeg (local) | Concatenate clips + mux audio + add fades/end card | Final MP4/WebM |

## Core File Locations

| Component | File |
|-----------|------|
| **Entry Point** | `backend/internal/api/handlers/generate.go` |
| **Pipeline Orchestration** | `backend/internal/api/handlers/generate_async.go` |
| **Script Generation** | `backend/internal/service/parser.go` → `backend/internal/adapters/gpt4o_adapter.go` |
| **Video Adapters** | `backend/internal/adapters/veo_adapter.go`, `kling_adapter.go` |
| **Audio Adapters** | `backend/internal/adapters/minimax_adapter.go`, `tts_adapter.go` |
| **Prompts** | `backend/internal/prompts/ad_script_prompt.go` |
| **Domain Models** | `backend/internal/domain/job.go`, `script.go` |

## Job Status Flow

```
pending → processing → completed
              ↓
           failed / paused
```

**Stage Progression:**
`script_generating` → `script_complete` → `scene_N_generating` → `scene_N_complete` → `audio_generating` → `audio_complete` → `composing` → `complete`

## External Service Integrations

| Service | Provider | Purpose |
|---------|----------|---------|
| **GPT-4o** | Replicate | Script generation, style reference analysis |
| **Veo 3.1** | Replicate (Google) | Video clip generation (default) |
| **Kling v2.5** | Replicate | Video clip generation (alternative) |
| **Minimax Music 1.5** | Replicate | Background music generation |
| **OpenAI TTS** | OpenAI Direct | Narrator voiceover generation |
| **DynamoDB** | AWS | Job state persistence |
| **S3** | AWS | Asset storage (clips, audio, final video) |

## Key Configuration

| Constant | Value | Location |
|----------|-------|----------|
| Video Generation Timeout | 15 min | `constants.go:8` |
| Video Poll Attempts | 240 (5s intervals) | `constants.go:13` |
| Audio Poll Attempts | 60 (5s intervals) | `constants.go:16` |
| Max Concurrent Jobs | 10 | `constants.go:25` |
| End Card Duration | 2.5s | `generate_async.go:1888` |
| Cross-Dissolve | 1.5s | `generate_async.go:1889` |

## S3 Asset Structure

```
users/{userID}/jobs/{jobID}/
├── clips/
│   ├── scene-001.mp4
│   └── scene-00N.mp4
├── thumbnails/
│   ├── scene-001.jpg
│   └── job-thumbnail.jpg
├── audio/
│   ├── background-music.mp3
│   └── narrator-voiceover.mp3
└── final/
    ├── video.mp4
    └── video.webm
```

## FFmpeg Post-Processing Chain

```
1. Concatenate clips (stream copy)
2. Apply FPS interpolation (if < 30fps)
3. Add fade in/out (1.5s / 2.0s)
4. Add text overlay (pharma disclaimer)
5. Generate end card (product image + CTA)
6. Cross-dissolve to end card
7. Mux audio (music @ 30% + narrator @ 100%)
8. Transcode to WebM (VP9/Opus)
```

## API Endpoints

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/generate` | `Generate()` - Start video generation |
| GET | `/api/v1/jobs/:id/progress` | `Progress()` - SSE stream for job progress |
| POST | `/api/v1/jobs/:id/scenes/:n/regenerate` | `Regenerate()` - Regenerate specific scene |
| GET | `/api/v1/jobs/:id` | `GetJob()` - Get job details |
| POST | `/api/v1/jobs/:id/pause` | `PauseJob()` - Pause job |
| POST | `/api/v1/jobs/:id/resume` | `ResumeJob()` - Resume job |
