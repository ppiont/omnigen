# OmniGen MVP - Quick Reference

## 🎯 Competition Goals

**Deadline**: Nov 24, 2025 10:59 PM CT  
**Prize**: $5,000  
**MVP Checkpoint**: 48 hours (Nov 17, 2025)  
**Category**: Ad Creative Pipeline

---

## ✅ MVP Requirements (48 Hours)

1. Working video generation for Ad Creative ✓
2. Basic prompt to video flow ✓
3. Audio-visual sync (music + SFX) 🔨
4. Multi-clip composition (3-4 scenes) ✓
5. Consistent visual style across clips 🔨
6. Deployed pipeline ✓
7. Sample outputs (2+ videos) 🔨

---

## 🏗️ Architecture Overview

```
User → CloudFront → ALB → ECS (Go API)
                              ↓
                      Step Functions
                       ├─→ Lambda Generator (Replicate)
                       └─→ Lambda Composer (FFmpeg)
                              ↓
                         S3 + DynamoDB
```

---

## 📋 Development Priority

### Day 1-2 (MVP - 48 hours)

**Critical Path**:
1. ✅ Prompt Parser (existing)
2. ✅ Scene Planner (existing)
3. 🔨 Lambda Generator (NEW)
   - Replicate API integration
   - Character consistency (reference image)
   - Scene generation
4. 🔨 Lambda Composer (NEW)
   - FFmpeg video stitching
   - Audio track integration
   - Basic sound effects
5. 🔨 Generate 2 sample videos

**Status**: 60% complete (backend services ready, need Lambda functions)

---

### Day 3-5 (Refinement)

**Workspace Page** (NEW):
- `/workspace/{job_id}` route
- Video player with scene timeline
- Chat interface for iterative refinement
- Scene regeneration controls

**Enhanced Features**:
- Character consistency engine
- Sound effect generation
- Beat detection + sync

---

### Day 6-7 (Polish)

- Batch generation
- Multi-format support
- Error handling improvements
- Performance optimization

---

### Day 8-9 (Submission)

- Demo videos
- Demo video recording (5-7 min)
- Technical deep dive
- Documentation

---

## 🎬 Pipeline Stages

### 1. Prompt Parsing ✅
**Service**: `PromptParser`  
**Input**: Text prompt  
**Output**: Product type, styles, colors, scenes

### 2. Scene Planning ✅
**Service**: `ScenePlanner`  
**Output**: 3-4 scenes with timing

### 3. Scene Generation 🔨
**Service**: Lambda Generator  
**Process**:
- Generate reference image
- Call Replicate for each scene
- Maintain character consistency
- Store clips in S3

### 4. Video Composition 🔨
**Service**: Lambda Composer  
**Process**:
- Download scenes
- Apply transitions
- Add audio + SFX
- Render with FFmpeg
- Upload final video

---

## 💰 Cost Target

**30-second ad**: $1.80 (✅ under $2.00 target)

**Breakdown**:
- Scene generation: $1.50
- Audio: $0.15
- Composition: $0.15

---

## 🚀 API Endpoints

### Existing ✅
```
POST   /api/v1/generate          - Create job
GET    /api/v1/jobs/{job_id}     - Get status
GET    /api/v1/jobs              - List jobs
```

### New 🔨
```
GET    /api/v1/workspace/{job_id}    - Workspace data
POST   /api/v1/jobs/{job_id}/refine  - Iterative refinement
POST   /api/v1/jobs/{job_id}/scene   - Regenerate scene
GET    /api/v1/download/{job_id}     - Download video
```

---

## 📱 Pages

### Existing ✅
- Landing `/` 
- Login `/login`
- Signup `/signup`
- Dashboard `/dashboard`
- Create `/create`
- Videos `/videos`
- Settings `/settings`

### New 🔨
- **Workspace `/workspace/{job_id}`**
  - Video player + timeline
  - Chat interface
  - Scene editor
  - Export controls

---

## 🎨 User Flow

```
Landing → Signup → Login → Dashboard
                              ↓
                         Create Page
                              ↓
                     [Generate Video]
                              ↓
                    Workspace (new job)
                    ├─ Chat: "Make brighter"
                    ├─ Regenerate Scene 2
                    └─ Export video
                              ↓
                         Videos Library
                              ↓
                    Workspace (existing job)
```

---

## 🔧 Tech Stack

**Backend**: Go, Gin, AWS SDK  
**Frontend**: React, Vite, React Router  
**Infrastructure**: ECS, Lambda, Step Functions, DynamoDB, S3  
**AI**: Replicate (FLUX, Stable Video Diffusion)  
**Audio**: ElevenLabs, Epidemic Sound  
**Video**: FFmpeg

---

## 📊 Success Metrics

**MVP (48h)**:
- 1 working ad generation ✅
- 2 sample videos 🔨
- <$2 per video ✅

**Competition (9d)**:
- Output Quality: 40% (visual coherence, sync, polish)
- Architecture: 25% (code quality, scalability)
- Cost Efficiency: 20% (<$2/video)
- User Experience: 15% (intuitive, flexible)

---

## ⚠️ Critical Risks

1. **Replicate API limits** → Retry logic + multiple keys
2. **Video generation failures** → Step Functions retry
3. **FFmpeg timeout** → 15min Lambda timeout
4. **Cost overrun** → Track costs, use cheaper models for testing
5. **Character consistency** → Reference image + seed approach

---

## 🎯 Competitive Advantages

1. End-to-end automation
2. Iterative refinement (chat interface)
3. Character consistency
4. Auto sound effects
5. <$2 per video
6. Beautiful Aurora UI
7. Scalable AWS architecture

---

## 📝 Next Actions

### Immediate (Today)
1. Implement Lambda Generator
2. Replicate API integration
3. Test scene generation

### Tomorrow
1. Implement Lambda Composer
2. FFmpeg pipeline
3. Generate 2 sample videos
4. Deploy + submit MVP

---

## 📞 Key Commands

```bash
# Deploy backend
cd backend && docker build -t omnigen-api .
docker push {ecr-url}:latest
aws ecs update-service --cluster omnigen --service api --force-new-deployment

# Deploy Lambda
cd infrastructure && terraform apply

# Deploy frontend
cd frontend && npm run build
aws s3 sync dist/ s3://omnigen-frontend --delete

# Test generation
curl -X POST https://api.omnigen.com/api/v1/generate \
  -H "Authorization: Bearer $JWT" \
  -d '{"prompt":"Luxury watch ad","duration":30,"aspect_ratio":"9:16"}'
```

---

**Status**: Ready to build Lambda functions  
**Next Milestone**: MVP in 48 hours  
**Let's ship! 🚀**

