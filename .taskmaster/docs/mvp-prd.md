# OmniGen AI Video Generation Pipeline - MVP PRD

**Project**: OmniGen AI Video Ad Generation Platform  
**Competition**: $5,000 Bounty - AI Video Generation Pipeline  
**Timeline**: 9 days (MVP at 48 hours)  
**Category Focus**: Ad Creative Pipeline (Primary), Music Video (Stretch)  
**Team**: Solo/Small Team  
**Version**: 1.0 - MVP Focus

---

## Executive Summary

OmniGen is an AI-powered video generation platform that transforms text prompts into professional ad creatives in minutes. Built on AWS serverless architecture with a focus on cost efficiency ($2/video), visual coherence, and production quality.

**Core Value Proposition**: Generate professional video ads at scale with AI-powered automation, complete with scene planning, visual consistency, audio-visual sync, and iterative refinement.

---

## Competition Alignment

### MVP Requirements (48 Hour Checkpoint) ✅

1. ✅ **Working video generation for Ad Creative category**
2. ✅ **Basic prompt to video flow** (text input → video output)
3. ✅ **Audio visual sync** (background music + sound effects)
4. ✅ **Multi-clip composition** (3-4 scenes stitched together)
5. ✅ **Consistent visual style across clips** (LoRA/style transfer)
6. ✅ **Deployed pipeline** (AWS ECS + CloudFront)
7. ✅ **Sample outputs** (2+ generated ad videos)

### Core Pipeline Requirements

**Example MVP Architecture Alignment**:
1. **Prompt Parser** → `PromptParser` service (regex + Claude API)
2. **Content Planner** → `ScenePlanner` service (3-4 scene structure)
3. **Generation Engine** → Lambda functions via Step Functions
4. **Composition Layer** → Composer Lambda (FFmpeg stitching)
5. **Output Handler** → S3 storage + CloudFront delivery

---

## Technical Architecture

### Current Infrastructure (Existing)

```
Frontend (React + Vite)
    ↓
CloudFront CDN
    ↓
ALB (Application Load Balancer)
    ↓
ECS Fargate (Go API - 1-5 tasks)
    ↓
Step Functions Express Workflow
    ├─→ Lambda: Scene Generator (Replicate API)
    └─→ Lambda: Video Composer (FFmpeg)
    ↓
Storage Layer
    ├─→ S3: Video assets + outputs
    └─→ DynamoDB: Job tracking
```

### User Flow Architecture

```
Landing Page → Sign Up → Login
    ↓
Dashboard (Overview + Stats)
    ↓
    ├─→ Create Page (/create)
    │       ├─→ Prompt input + Advanced options
    │       └─→ Generate button → Workspace (/workspace/{job_id})
    │
    ├─→ Videos Library (/videos)
    │       └─→ Select video → Workspace (/workspace/{job_id})
    │
    ├─→ Workspace (/workspace/{job_id})
    │       ├─→ Video player (scene timeline)
    │       ├─→ Chat interface (iterative refinement)
    │       ├─→ Scene editor (regenerate/adjust)
    │       └─→ Export controls
    │
    └─→ Settings (/settings)
            └─→ Password change + basic preferences
```

---

## MVP Features (48 Hours)

### 1. Ad Creative Pipeline ⭐ PRIMARY FOCUS

**Input**: Product description + brand guidelines + ad specs  
**Output**: 15-60 second video advertisement

**Example Prompts**:
- "Create a 30 second Instagram ad for luxury watches with elegant gold aesthetics"
- "Generate 3 variations of a TikTok ad for energy drinks with extreme sports footage"
- "Make a product showcase video for minimalist skincare brand with clean white backgrounds"

#### Scene Structure

**Short Ads (15-30s)**: 3 scenes
- Scene 1 (30%): Product intro/reveal
- Scene 2 (40%): Product showcase/features
- Scene 3 (30%): Brand CTA

**Long Ads (30-60s)**: 4 scenes
- Scene 1 (20%): Hook/attention grabber
- Scene 2 (25%): Product reveal
- Scene 3 (30%): Product in use/lifestyle
- Scene 4 (25%): CTA/Brand identity

#### Implementation Status

**EXISTING** ✅:
- `PromptParser`: Extracts product type, visual style, colors, text overlays
- `ScenePlanner`: Creates 3-4 scene structure with timing
- `GeneratorService`: Orchestrates pipeline via Step Functions
- Backend handlers with JWT auth + rate limiting

**TO BUILD** 🔨:
- Lambda Generator: Call Replicate models for scene generation
- Lambda Composer: FFmpeg video stitching + transitions
- Character consistency engine (LoRA/reference images)
- Sound effect generation integration
- Workspace page with iterative refinement

---

## Core Pipeline Stages

### Stage 1: Prompt Parsing & Planning

**Service**: `PromptParser` + `ScenePlanner`

**Input**:
```json
{
  "prompt": "Luxury watch ad with gold aesthetics, elegant reveal",
  "duration": 30,
  "aspect_ratio": "9:16",
  "style": "cinematic, luxury, minimal"
}
```

**Output**:
```json
{
  "product_type": "luxury watch",
  "visual_style": ["luxury", "elegant", "cinematic"],
  "color_palette": ["gold", "black"],
  "scenes": [
    {
      "number": 1,
      "duration": 9.0,
      "prompt": "luxury watch emerging from elegant background, luxury, elegant style, gold, black colors, dramatic reveal",
      "style": "luxury, elegant",
      "transition": "fade"
    },
    // ... 2 more scenes
  ]
}
```

**Implementation**:
- ✅ Regex-based parsing (MVP)
- 🔨 Claude API integration (post-MVP)
- ✅ Scene timing algorithm
- 🔨 Brand preset integration

---

### Stage 2: Scene Generation

**Service**: Lambda Generator Function

**Process**:
1. Receive scene prompts from Step Functions
2. Generate reference image (style consistency)
3. Call Replicate video model for each scene
4. Monitor generation status (polling/webhooks)
5. Download generated clips to S3
6. Return clip metadata to workflow

**Models** (via Replicate):

**MVP Models** (Fast + Cheap):
- **Image**: SDXL Turbo, Playground v2.5
- **Video**: Zeroscope, AnimateDiff
- **Cost Target**: $0.50-1.00 per scene

**Production Models** (High Quality):
- **Image**: FLUX.1 Pro, Midjourney
- **Video**: Runway Gen-3, Kling AI, Pika
- **Cost Target**: $1.50-2.50 per scene

**Character Consistency Strategy**:
```
1. Generate reference image for product/character (scene 1)
2. Use image-to-video with reference for subsequent scenes
3. Maintain style seed across generations
4. Apply LoRA fine-tuning for brand consistency (advanced)
```

**Replicate API Call**:
```javascript
// Scene generation with reference
const output = await replicate.run(
  "stability-ai/stable-video-diffusion",
  {
    input: {
      video_length: scene.duration,
      image: referenceImageUrl, // Character consistency
      prompt: scene.prompt,
      style_preset: "cinematic",
      motion_bucket_id: 127
    }
  }
);
```

---

### Stage 3: Audio & Sound Effects

**NEW FEATURE** 🎵

**Components**:
1. **Background Music**: Suno/Udio API or royalty-free library
2. **Sound Effects**: ElevenLabs sound effects or generated
3. **Audio Sync**: Match scene transitions to music beats

**Implementation**:

```javascript
// Generate background music for ad
const music = await generateBackgroundMusic({
  duration: totalDuration,
  mood: parsed.visualStyle,
  genre: "cinematic",
  bpm: 120
});

// Generate sound effects for key moments
const soundEffects = {
  scene1: "whoosh_reveal.mp3",    // Product reveal
  scene2: "subtle_click.mp3",     // Feature highlight
  scene3: "powerful_impact.mp3"   // CTA emphasis
};

// Sync to scene transitions
const audioTimeline = [
  { time: 0, audio: music, volume: 0.7 },
  { time: 9.0, audio: soundEffects.scene1, volume: 1.0 },
  { time: 21.0, audio: soundEffects.scene3, volume: 0.9 }
];
```

**Audio Libraries**:
- **Music**: Epidemic Sound API, Artlist, Suno AI
- **SFX**: Freesound.org, ElevenLabs Sound Effects
- **Voiceover**: ElevenLabs, Play.ht (optional)

---

### Stage 4: Video Composition

**Service**: Lambda Composer Function

**Process**:
1. Download all scene clips from S3
2. Apply transitions (fade, dissolve, cut)
3. Add audio track (music + sound effects)
4. Apply text overlays (product name, CTA)
5. Add brand watermark/logo
6. Render final video with FFmpeg
7. Upload to S3 + update DynamoDB

**FFmpeg Pipeline**:

```bash
# Scene stitching with transitions
ffmpeg -i scene1.mp4 -i scene2.mp4 -i scene3.mp4 \
  -i background_music.mp3 -i sfx_reveal.wav \
  -filter_complex "
    [0:v]fade=t=out:st=8.5:d=0.5[v0];
    [1:v]fade=t=in:st=0:d=0.5,fade=t=out:st=11.5:d=0.5[v1];
    [2:v]fade=t=in:st=0:d=0.5[v2];
    [v0][v1]concat[vtemp];
    [vtemp][v2]concat[vout];
    [3:a][4:a]amix=inputs=2:duration=first[aout]
  " \
  -map "[vout]" -map "[aout]" \
  -c:v libx264 -preset medium -crf 23 \
  -c:a aac -b:a 192k \
  output.mp4
```

**Quality Settings**:
- Resolution: 1080p minimum
- Frame Rate: 30 FPS
- Bitrate: 5000k (video), 192k (audio)
- Codec: H.264 (max compatibility)

---

## Advanced Features (Post-MVP)

### 1. Character Consistency Engine 🎭

**Purpose**: Maintain consistent product/character appearance across all scenes

**Approach**:

**MVP** (Week 1):
- Generate reference image in Scene 1
- Use image-to-video with reference for Scene 2-4
- Maintain style prompts + seeds

**Advanced** (Post-Competition):
- Train custom LoRA on brand/product images
- Face/object detection + keypoint tracking
- Style transfer from reference library

**Implementation**:

```javascript
// Generate reference image
const referenceImage = await replicate.run("flux-schnell", {
  prompt: `product photography, ${productType}, ${visualStyle}, studio lighting`,
  aspect_ratio: aspectRatio,
  seed: brandSeed
});

// Use reference for all scenes
const scenes = await Promise.all(
  scenePlan.map(scene => 
    generateSceneWithReference({
      prompt: scene.prompt,
      reference: referenceImage,
      duration: scene.duration,
      seed: brandSeed
    })
  )
);
```

---

### 2. Iterative Refinement Workspace 💬

**NEW PAGE**: `/workspace/{job_id}`

**Purpose**: Chat-based interface for editing and refining generated videos

**UI Components**:

```
┌─────────────────────────────────────────────────┐
│  Video Player (Scene Timeline)                  │
│  ┌──────────┬──────────┬──────────┬──────────┐ │
│  │ Scene 1  │ Scene 2  │ Scene 3  │ Scene 4  │ │
│  │  9.0s    │  7.5s    │  9.0s    │  7.5s    │ │
│  └──────────┴──────────┴──────────┴──────────┘ │
├─────────────────────────────────────────────────┤
│  Chat Interface                                 │
│  User: "Make scene 2 brighter"                  │
│  AI: "Regenerating scene 2 with increased       │
│       brightness... ✨"                         │
│  [Regenerate] [Apply] [Undo]                    │
├─────────────────────────────────────────────────┤
│  Scene Controls                                 │
│  Selected: Scene 2                              │
│  - Regenerate with new prompt                   │
│  - Adjust duration                              │
│  - Change transition                            │
│  - Replace with different style                 │
└─────────────────────────────────────────────────┘
```

**Chat Commands**:

```typescript
// Example chat interactions
const chatCommands = {
  "Make scene 2 brighter": {
    action: "regenerate_scene",
    scene: 2,
    modifications: { brightness: +20 }
  },
  "Add more motion to the chorus": {
    action: "regenerate_scene",
    scene: 3,
    modifications: { motion_intensity: "high" }
  },
  "Change the color palette to warmer tones": {
    action: "regenerate_all",
    modifications: { color_palette: ["orange", "red", "yellow"] }
  },
  "Swap scene 1 and 2": {
    action: "reorder_scenes",
    order: [2, 1, 3, 4]
  }
};
```

**Backend API**:

```go
// POST /api/v1/jobs/{job_id}/refine
type RefineRequest struct {
    Message     string            `json:"message"`      // Chat message
    Scene       *int              `json:"scene"`        // Optional scene number
    Action      string            `json:"action"`       // regenerate, adjust, replace
    Modifications map[string]any  `json:"modifications"` // Style changes
}

type RefineResponse struct {
    Status      string   `json:"status"`
    NewJobID    *string  `json:"new_job_id"`    // For full regeneration
    UpdatedScene *int    `json:"updated_scene"` // For single scene
    Message     string   `json:"message"`
}
```

**State Management**:

```javascript
// Workspace state
const workspace = {
  jobId: "abc-123",
  version: 2, // Iteration version
  scenes: [
    { number: 1, status: "completed", s3Key: "scene1_v2.mp4" },
    { number: 2, status: "regenerating", s3Key: null },
    { number: 3, status: "completed", s3Key: "scene3_v1.mp4" },
    { number: 4, status: "completed", s3Key: "scene4_v1.mp4" }
  ],
  chatHistory: [
    { role: "user", message: "Make scene 2 brighter" },
    { role: "assistant", message: "Regenerating scene 2..." }
  ]
};
```

---

### 3. Sound Effect Generation 🔊

**Purpose**: Auto-generate sound effects that match visual events

**Approach**:

**MVP**:
- Pre-mapped sound effects library
- Rule-based timing (scene transitions, key moments)

**Advanced**:
- Audio generation models (ElevenLabs, AudioCraft)
- Visual analysis → sound matching
- Beat detection for music sync

**Implementation**:

```javascript
// Analyze scenes for sound effect opportunities
const soundEffectMap = {
  "product reveal": "whoosh_reveal.wav",
  "feature highlight": "subtle_click.wav",
  "dramatic moment": "powerful_impact.wav",
  "transition": "smooth_transition.wav"
};

// Auto-detect moments needing SFX
const sfxTimeline = analyzeScenes(scenes).map(event => ({
  time: event.timestamp,
  effect: soundEffectMap[event.type],
  volume: event.intensity * 0.8
}));

// Apply to composition
await composerService.addSoundEffects(videoId, sfxTimeline);
```

---

### 4. Batch Generation 🚀

**Purpose**: Generate multiple ad variations simultaneously

**Use Cases**:
- A/B testing (3-5 variations)
- Multi-platform (16:9, 9:16, 1:1)
- Style variations (luxury, playful, minimal)

**API**:

```go
// POST /api/v1/generate/batch
type BatchGenerateRequest struct {
    BasePrompt   string            `json:"base_prompt"`
    Variations   []PromptVariation `json:"variations"`
    AspectRatios []string          `json:"aspect_ratios"`
}

type PromptVariation struct {
    Name   string `json:"name"`
    Style  string `json:"style"`
    Prompt string `json:"prompt"`
}

// Response: array of job IDs
type BatchGenerateResponse struct {
    Jobs []Job `json:"jobs"`
    BatchID string `json:"batch_id"`
}
```

---

## Technical Requirements

### 1. Generation Quality ⭐

**Visual Coherence**:
- ✅ Consistent art style across all clips (reference image + seed)
- ✅ Smooth transitions (fade, dissolve, cut)
- ✅ No jarring style shifts (LoRA fine-tuning)
- ✅ Professional color grading (LUT application)

**Audio-Visual Sync**:
- ✅ Beat-matched transitions (music videos)
- ✅ Sound effects aligned with visuals
- ✅ No audio-video drift (precise timestamps)

**Output Quality**:
- ✅ 1080p resolution minimum
- ✅ 30 FPS
- ✅ Clean audio (192kbps AAC)
- ✅ Optimized file size (<50MB for 30s)

### 2. Pipeline Performance ⚡

**Speed Targets**:
- 30s video: <5 minutes ✅
- 60s video: <10 minutes ✅
- 3min video: <20 minutes (stretch)

**Cost Efficiency**:
- Target: <$2.00 per minute of video ✅
- Track costs per model call
- Cache repeated elements
- Use cheaper models for MVP iteration

**Reliability**:
- 90%+ success rate ✅
- Automatic retry (Step Functions)
- Graceful failure handling
- Error logging + alerting

### 3. User Experience 🎨

**Input Flexibility**:
- Natural language prompts ✅
- Advanced options (style, duration, aspect ratio) ✅
- Reference image uploads (character consistency)
- Brand guideline presets (post-MVP)

**Output Control**:
- Real-time generation progress ✅
- Scene-by-scene preview
- Regenerate specific scenes (workspace)
- Export multiple formats

**Feedback Loop**:
- Chat-based refinement ✅
- Visual timeline editor
- Version history
- Undo/redo support

---

## Database Schema

### DynamoDB: Jobs Table

```json
{
  "job_id": "uuid",
  "user_id": "auth0|12345",
  "status": "pending|processing|completed|failed",
  "prompt": "Luxury watch ad...",
  "duration": 30,
  "aspect_ratio": "9:16",
  "style": "luxury, cinematic",
  "video_key": "videos/abc-123/final.mp4",
  "scenes": [
    {
      "number": 1,
      "s3_key": "scenes/abc-123/scene1.mp4",
      "status": "completed",
      "duration": 9.0
    }
  ],
  "audio": {
    "music_key": "audio/abc-123/background.mp3",
    "sfx_keys": ["audio/abc-123/sfx1.wav"]
  },
  "version": 1,
  "parent_job_id": null,
  "created_at": 1699999999,
  "completed_at": 1700000100,
  "cost": 1.85,
  "ttl": 1707691199
}
```

### DynamoDB: Users Table (Future)

```json
{
  "user_id": "auth0|12345",
  "email": "user@example.com",
  "quota_used": 45,
  "quota_limit": 100,
  "created_at": 1699999999,
  "subscription_tier": "free|pro|enterprise"
}
```

---

## API Endpoints

### Core Generation

```
POST   /api/v1/generate           - Create video generation job
GET    /api/v1/jobs/{job_id}      - Get job status + metadata
GET    /api/v1/jobs                - List user's jobs
DELETE /api/v1/jobs/{job_id}      - Delete job
```

### Workspace & Refinement (NEW)

```
GET    /api/v1/workspace/{job_id}     - Get workspace data
POST   /api/v1/jobs/{job_id}/refine   - Iterative refinement
POST   /api/v1/jobs/{job_id}/scene    - Regenerate single scene
PATCH  /api/v1/jobs/{job_id}/timeline - Reorder/adjust scenes
```

### Batch Operations (NEW)

```
POST   /api/v1/generate/batch     - Batch generation
GET    /api/v1/batch/{batch_id}   - Batch status
```

### Assets

```
GET    /api/v1/download/{job_id}  - Download final video
GET    /api/v1/scene/{scene_id}   - Download individual scene
POST   /api/v1/upload              - Upload reference image
```

---

## Frontend Pages

### 1. Landing Page `/` ✅ EXISTING
- Hero with aurora animation
- Feature cards (multi-format, brand consistency, A/B test, cost)
- Pipeline steps (Brief → Generate → Export)
- CTA: Sign up

### 2. Sign Up `/signup` ✅ EXISTING
- AWS Cognito integration
- Email + password
- Free tier: 5 videos

### 3. Login `/login` ✅ EXISTING
- JWT authentication
- Password reset flow

### 4. Dashboard `/dashboard` ✅ EXISTING
- Stats cards (videos generated, avg time, cost, success rate)
- Recent videos grid
- Quick actions

### 5. Create Page `/create` ✅ EXISTING (Enhance)

**Current**:
- Prompt textarea
- Advanced options (category, style, duration, aspect ratio)
- Generate button

**Enhancements** 🔨:
- Real-time cost estimate
- Character reference upload
- Brand preset selector
- Multi-format batch toggle

### 6. Videos Library `/videos` ✅ EXISTING
- Grid of generated videos
- Filter by status, format, date
- Click → Workspace

### 7. Workspace `/workspace/{job_id}` 🔨 NEW

**Layout**:

```
┌─────────────────────────────────────────────────┐
│  Header: Video Title | Status | Export          │
├─────────────────────────────────────────────────┤
│                                                 │
│  Video Player (Scene Timeline)                  │
│  [Scene 1][Scene 2][Scene 3][Scene 4]          │
│                                                 │
├─────────────────────────────────────────────────┤
│  Scene Editor    │  Chat Interface              │
│  - Regenerate    │  User: "Make brighter"       │
│  - Adjust timing │  AI: Regenerating...         │
│  - Change style  │  [Send]                      │
└─────────────────────────────────────────────────┘
```

**Features**:
- Video player with scene markers
- Timeline scrubbing
- Chat-based editing
- Scene controls (regenerate, adjust, replace)
- Version history
- Export options (format, quality)

**Routes**:
- Accessed from `/create` after generation
- Accessed from `/videos` by clicking video card

### 8. Settings `/settings` ✅ EXISTING
- Password change
- Profile info
- (Future: Brand presets, API keys, billing)

---

## Cost Analysis

### Generation Cost Breakdown

**30-second ad (3 scenes)**:

```
Scene 1 (9s):
  - Image generation (reference): $0.05
  - Video generation: $0.40
  - Audio (BGM segment): $0.10
  - SFX: $0.05

Scene 2 (12s):
  - Video generation (w/ reference): $0.50
  - SFX: $0.05

Scene 3 (9s):
  - Video generation (w/ reference): $0.40
  - SFX: $0.05

Composition:
  - FFmpeg processing: $0.10 (Lambda)
  - Storage (S3): $0.01
  - Data transfer: $0.05

Total: ~$1.76 per 30s video ✅ Under $2 target
```

**Optimization Strategies**:
- Use cheaper models for MVP testing
- Cache reference images
- Reuse audio tracks
- Batch processing discounts
- Smart model selection (quality vs speed)

---

## Deployment & Infrastructure

### Current Setup ✅

**Frontend**:
- Vite build → S3 bucket
- CloudFront CDN
- Custom domain (optional)

**Backend**:
- Docker image → ECR
- ECS Fargate (auto-scaling 1-5 tasks)
- ALB with health checks

**Serverless**:
- Lambda Generator (Node.js, 3GB RAM, 15min timeout)
- Lambda Composer (Node.js + FFmpeg layer, 5GB RAM, 15min timeout)
- Step Functions Express workflow

**Storage**:
- S3 bucket (videos + assets)
- DynamoDB table (jobs)
- CloudWatch logs

### Configuration

**Environment Variables**:
```bash
# Backend (ECS)
PORT=8080
AWS_REGION=us-east-1
ASSETS_BUCKET=omnigen-assets-{account_id}
JOB_TABLE=omnigen-jobs
STEP_FUNCTIONS_ARN=arn:aws:states:...
REPLICATE_API_KEY=r8_...
COGNITO_USER_POOL_ID=us-east-1_...
JWT_SECRET=...

# Lambda
REPLICATE_API_KEY=r8_...
S3_BUCKET=omnigen-assets-{account_id}
DYNAMODB_TABLE=omnigen-jobs
```

### Deployment Commands

```bash
# Frontend
cd frontend
npm run build
aws s3 sync dist/ s3://omnigen-frontend --delete
aws cloudfront create-invalidation --distribution-id E123 --paths "/*"

# Backend
cd backend
docker build -t omnigen-api .
docker tag omnigen-api:latest {ecr-url}:latest
docker push {ecr-url}:latest
aws ecs update-service --cluster omnigen --service api --force-new-deployment

# Lambda (via Terraform)
cd infrastructure
terraform apply
```

---

## Testing Plan

### Unit Tests

```go
// Backend services
func TestPromptParser(t *testing.T) { /* ... */ }
func TestScenePlanner(t *testing.T) { /* ... */ }
func TestGeneratorService(t *testing.T) { /* ... */ }
```

```javascript
// Lambda functions
test('Generator Lambda: generates scenes', async () => { /* ... */ });
test('Composer Lambda: stitches video', async () => { /* ... */ });
```

### Integration Tests

```bash
# End-to-end generation test
POST /api/v1/generate
{
  "prompt": "Test ad for luxury watch",
  "duration": 15,
  "aspect_ratio": "16:9"
}

# Poll status until complete
GET /api/v1/jobs/{job_id}

# Download and validate video
GET /api/v1/download/{job_id}
```

### Evaluation Scenarios (From Competition)

**Music Videos**:
- ❌ Skip for MVP (focus on Ad Creative)

**Ad Creatives** ✅:
- "Create 30s Instagram ad for luxury watches with elegant gold aesthetics"
- "Generate 3 variations of 15s TikTok ad for energy drinks"
- "Make product showcase video for minimalist skincare with clean white backgrounds"

**Stress Tests**:
- 5 concurrent generation requests
- 60-second video (4 scenes)
- Complex multi-part prompt

---

## Success Metrics

### MVP (48 Hours)

- [ ] 1 working ad creative generation end-to-end
- [ ] 3-4 scene composition with transitions
- [ ] Visual style consistency across scenes
- [ ] Audio track + basic sound effects
- [ ] Deployed API + frontend
- [ ] 2 sample videos submitted

### Competition (9 Days)

**Output Quality** (40%):
- Visual coherence: Professional, consistent style
- Audio-visual sync: Transitions aligned
- Creative execution: Matches prompt intent
- Technical polish: 1080p, 30 FPS, clean audio

**Pipeline Architecture** (25%):
- Clean, documented code
- Scalable design (Step Functions + Lambda)
- Robust error handling
- Performance optimized

**Cost Effectiveness** (20%):
- <$2.00 per minute of video
- Smart caching + optimization
- Efficient API usage

**User Experience** (15%):
- Intuitive interface (React + Aurora theme)
- Flexible prompts
- Real-time progress feedback
- Iterative refinement (workspace)

### KPIs

- Generation success rate: >90%
- Average generation time (30s video): <5 minutes
- Average cost per video: <$1.80
- User satisfaction: Coherent, professional output

---

## Timeline

### Day 1-2: MVP Foundation (48 Hours) ⏰

**Day 1**:
- [ ] Lambda Generator: Replicate integration
- [ ] Scene generation with reference image
- [ ] Character consistency (basic)
- [ ] Test 3-scene ad generation

**Day 2**:
- [ ] Lambda Composer: FFmpeg pipeline
- [ ] Audio track integration
- [ ] Basic sound effects
- [ ] Deploy + generate 2 sample videos

### Day 3-5: Refinement

**Day 3**:
- [ ] Workspace page UI
- [ ] Scene timeline display
- [ ] Export controls

**Day 4**:
- [ ] Chat interface backend
- [ ] Scene regeneration API
- [ ] Version management

**Day 5**:
- [ ] Sound effect generation
- [ ] Beat detection + sync
- [ ] Quality improvements

### Day 6-7: Polish

**Day 6**:
- [ ] Batch generation API
- [ ] Multi-format support
- [ ] Cost tracking

**Day 7**:
- [ ] Error handling improvements
- [ ] Performance optimization
- [ ] Testing + bug fixes

### Day 8-9: Submission

**Day 8**:
- [ ] Generate demo videos
- [ ] Record demo video (5-7 min)
- [ ] Technical deep dive doc

**Day 9**:
- [ ] Final testing
- [ ] Documentation polish
- [ ] Submit by 10:59 PM

---

## Risk Mitigation

### High Priority Risks

**Risk**: Replicate API rate limits  
**Mitigation**: Implement exponential backoff, queue system, multiple API keys

**Risk**: Video generation failures  
**Mitigation**: Step Functions retry logic, fallback to simpler models

**Risk**: FFmpeg processing timeout  
**Mitigation**: Increase Lambda timeout to 15min, optimize compression

**Risk**: Cost overrun  
**Mitigation**: Track costs per generation, alert at thresholds, use cheaper models for testing

**Risk**: Character consistency issues  
**Mitigation**: Use reference image + seed, LoRA fine-tuning, manual fallback

---

## Competitive Advantages

1. **End-to-End Pipeline** ✅: Prompt → Video with zero manual intervention
2. **Iterative Refinement** 🆕: Chat-based editing after generation
3. **Character Consistency** 🆕: Reference image + LoRA approach
4. **Sound Design** 🆕: Auto-generated sound effects matching visuals
5. **Cost Efficiency**: <$2/video with smart model selection
6. **Production Quality**: 1080p, professional transitions, clean audio
7. **Scalable Architecture**: AWS serverless, auto-scaling
8. **Beautiful UI**: Aurora-themed, modern, intuitive

---

## Post-Competition Roadmap

### Phase 1: Production (Weeks 2-4)
- Custom LoRA training for brands
- Advanced beat detection + music sync
- Voiceover integration (ElevenLabs)
- Brand preset library

### Phase 2: Scale (Months 2-3)
- Music video pipeline
- Educational/explainer pipeline
- Batch generation UI
- API for developers

### Phase 3: Monetization (Month 4+)
- Free tier: 5 videos/month
- Pro tier: $49/month (50 videos)
- Enterprise: Custom pricing
- Affiliate program

---

## References

### Competition Materials
- AI Video Generation Pipeline Brief (attached)
- MVP requirements (48-hour checkpoint)
- Evaluation criteria (Output Quality 40%, Architecture 25%, Cost 20%, UX 15%)

### Existing Codebase
- Backend: `/backend` (Go + Gin)
- Frontend: `/frontend` (React + Vite)
- Infrastructure: `/infrastructure` (Terraform)
- Design: `/docs/omnigen_design_spec.md`

### External Resources
- Replicate API: https://replicate.com/docs
- FFmpeg: https://ffmpeg.org/documentation.html
- Step Functions: https://docs.aws.amazon.com/step-functions/
- ElevenLabs: https://elevenlabs.io/docs

---

## Appendix

### A. Example Prompts

**Luxury Watch Ad (30s, 9:16)**:
```
"Create a 30 second Instagram ad for luxury watches with elegant gold aesthetics. Show the watch in dramatic lighting, highlight the craftsmanship, end with brand logo and 'Timeless Elegance' tagline."
```

**Energy Drink TikTok (15s, 9:16)**:
```
"Generate a 15 second TikTok ad for energy drinks with extreme sports footage. Show skateboarding, BMX, parkour. High energy, fast cuts, bold colors. CTA: 'Fuel Your Limits'"
```

**Skincare Product Showcase (60s, 16:9)**:
```
"Make a 60 second product showcase video for minimalist skincare brand with clean white backgrounds. Show product shots, texture close-ups, before/after, and application demo. Calm, clean aesthetic."
```

### B. Cost Breakdown Examples

**15s Short Ad** (3 scenes):
- Generation: $0.90
- Composition: $0.15
- Total: **$1.05** ✅

**30s Standard Ad** (4 scenes):
- Generation: $1.60
- Composition: $0.20
- Total: **$1.80** ✅

**60s Long Ad** (5 scenes):
- Generation: $2.80
- Composition: $0.30
- Total: **$3.10** ⚠️ (over $2/min target)

### C. Tech Stack Summary

**Frontend**:
- React 18 + Vite
- React Router v6
- AWS Amplify (Auth)
- Tailwind CSS (Aurora theme)

**Backend**:
- Go 1.25.4
- Gin web framework
- AWS SDK v2
- Zap logging

**Infrastructure**:
- ECS Fargate (API)
- Lambda (Generators + Composer)
- Step Functions (Orchestration)
- DynamoDB (Jobs)
- S3 (Videos)
- CloudFront (CDN)

**AI Models**:
- Replicate API (Video generation)
- FLUX.1 / SDXL (Image reference)
- Stable Video Diffusion (Video)
- ElevenLabs (Audio/SFX)

---

**Document Version**: 1.0  
**Last Updated**: Nov 15, 2025  
**Status**: MVP in progress  
**Competition Deadline**: Nov 24, 2025 10:59 PM CT

**Focus**: Ship MVP in 48 hours, iterate for 7 days, win $5,000 bounty 🚀

