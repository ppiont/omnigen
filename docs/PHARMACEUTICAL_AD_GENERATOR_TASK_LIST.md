# Pharmaceutical Ad Video Generator - Task List

## Overview

This task list breaks down the implementation of the Pharmaceutical Ad Video Generator into focused, manageable tasks. Each task targets specific files and implements one feature at a time.

**Based on**: `PHARMACEUTICAL_AD_GENERATOR_PRD.md`  
**Status**: Ready for implementation  
**Estimated Total Time**: 3-4 weeks (1 developer)

---

## Phase 1: Backend Foundation (Week 1)

### Task 1.1: Create TTS Adapter Interface and Implementation

**Description**: Create the Text-to-Speech adapter to generate narrator voiceovers using OpenAI TTS API.

**Acceptance Criteria**:

- [ ] TTS adapter interface defined
- [ ] OpenAI TTS implementation complete
- [ ] Voice mapping (male → "onyx", female → "nova")
- [ ] Retry logic: 3 attempts, exponential backoff
- [ ] Error handling and logging

**Files to Create**:

- `backend/internal/adapters/tts_adapter.go`

**Files to Modify**:

- None

**Dependencies**: None

**Estimated Time**: 4 hours

---

### Task 1.2: Add Voice and Side Effects Fields to Domain Models

**Description**: Update Job and Script domain models to support voice selection and side effects text.

**Acceptance Criteria**:

- [ ] `Voice` field added to Job struct
- [ ] `SideEffects` field added to Job struct
- [ ] `NarratorAudioURL` field added to Job struct
- [ ] `SideEffectsText` field added to Job struct
- [ ] `SideEffectsStartTime` field added to Job struct
- [ ] `NarratorScript` field added to AudioSpec
- [ ] `SideEffectsText` field added to AudioSpec
- [ ] `SideEffectsStartTime` field added to AudioSpec
- [ ] DynamoDB tags updated with correct JSON/DynamoDB annotations

**Files to Modify**:

- `backend/internal/domain/job.go`
- `backend/internal/domain/script.go`

**Dependencies**: None

**Estimated Time**: 2 hours

---

### Task 1.3: Update Generate Request Handler with Validation

**Description**: Add voice and side effects parameters to the generate endpoint with comprehensive validation.

**Acceptance Criteria**:

- [ ] `Voice` field added to GenerateRequest (required, enum: male/female)
- [ ] `SideEffects` field added to GenerateRequest (required, 10-500 chars)
- [ ] Validation error messages implemented (user-friendly)
- [ ] Voice and side effects stored in Job entity
- [ ] Product image validation (required, max 10MB)
- [ ] Duration validation (10-60s, divisible by 5)

**Validation Error Messages**:

- Voice missing: `"Please select a narrator voice (male or female)"`
- Voice invalid: `"Invalid voice selection. Choose 'male' or 'female'"`
- Side effects missing: `"Side effects disclosure is required for pharmaceutical ads"`
- Side effects too short: `"Side effects text must be at least 10 characters"`
- Side effects too long: `"Side effects text cannot exceed 500 characters (currently: {count})"`
- Product image missing: `"Product image is required for pharmaceutical ads"`
- Product image too large: `"Product image must be under 10MB (currently: {size}MB)"`
- Duration invalid: `"Duration must be between 10-60 seconds and divisible by 5"`

**Files to Modify**:

- `backend/internal/api/handlers/generate.go`

**Dependencies**: Task 1.2

**Estimated Time**: 3 hours

---

### Task 1.4: Update GPT-4o Script Generation Prompts

**Description**: Enhance GPT-4o prompts to generate narrator scripts, handle side effects text, and enforce scene duration constraints.

**Acceptance Criteria**:

- [ ] System prompt updated with narrator script instructions
- [ ] Word count scaling by duration (10s: 20-25 words, 30s: 60-80 words, 60s: 120-160 words)
- [ ] Scene duration constraint enforced (5s or 10s only)
- [ ] Side effects handling: Use user's text VERBATIM
- [ ] JSON schema updated with `narrator_script`, `side_effects_text`, `side_effects_start_time`
- [ ] Time allocation: 80% main content, 20% side effects
- [ ] Scene count calculation logic based on duration

**Files to Modify**:

- `backend/internal/prompts/ad_script_prompt.go`

**Dependencies**: None

**Estimated Time**: 3 hours

---

### Task 1.5: Initialize TTS Adapter in Main

**Description**: Set up TTS adapter initialization and dependency injection.

**Acceptance Criteria**:

- [ ] TTS adapter imported
- [ ] TTS API key retrieved from Secrets Manager or env
- [ ] TTS adapter initialized with configuration
- [ ] TTS adapter passed to GenerateHandler
- [ ] ServerConfig updated to include TTS adapter
- [ ] Environment variables documented

**Environment Variables**:

```bash
TTS_API_KEY=sk-...
TTS_PROVIDER=openai
```

**Files to Modify**:

- `backend/cmd/api/main.go`

**Dependencies**: Task 1.1

**Estimated Time**: 2 hours

---

## Phase 2: Audio Generation Pipeline (Week 1-2)

### Task 2.1: Implement Narrator Voiceover Generation

**Description**: Add narrator audio generation with variable speed for side effects segment.

**Acceptance Criteria**:

- [ ] `generateNarratorVoiceover()` method created
- [ ] Call TTS adapter with narrator script from GPT-4o
- [ ] Generate full script at 1.0x speed
- [ ] Apply variable speed using ffmpeg:
  - Main content (0 to side_effects_start_time): 1.0x speed
  - Side effects (side_effects_start_time to end): 1.4x speed
- [ ] Upload narrator audio to S3: `users/{userID}/jobs/{jobID}/audio/narrator-voiceover.mp3`
- [ ] Update job with narrator_audio_url
- [ ] Add stages: `narrator_generating`, `narrator_complete`
- [ ] Error handling: Fail job if TTS fails

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: Task 1.1, Task 1.2, Task 1.4

**Estimated Time**: 6 hours

---

### Task 2.2: Update Async Generation Pipeline Flow

**Description**: Integrate narrator generation into the video generation pipeline.

**Acceptance Criteria**:

- [ ] Pipeline flow updated:
  1. Script generation (GPT-4o)
  2. Narrator audio generation (TTS) ← NEW
  3. Video clips generation (Kling AI)
  4. Background music generation (Minimax)
  5. Video composition (ffmpeg)
- [ ] Progress stages include narrator generation
- [ ] Error propagation works correctly
- [ ] Full pipeline tested end-to-end

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: Task 2.1

**Estimated Time**: 3 hours

---

## Phase 3: Video Composition with Text Overlay (Week 2)

### Task 3.1: Update Video Composition Function Signature

**Description**: Modify composeVideo() to accept side effects parameters. All audio tracks remain separate.

**Acceptance Criteria**:

- [ ] Function signature updated with `sideEffectsText` parameter (from user input)
- [ ] Function signature updated with `sideEffectsStartTime` parameter (from GPT-4o)
- [ ] Remove `backgroundMusicURL` parameter (music is separate track, not mixed)
- [ ] Remove `narratorAudioURL` parameter (narrator is separate track, not mixed)
- [ ] All callers updated with new parameters
- [ ] Validation: Fail if side effects text is missing

**New Signature**:

```go
func (h *GenerateHandler) composeVideo(
    ctx context.Context,
    userID string,
    jobID string,
    clips []ClipVideo,
    sideEffectsText string,
    sideEffectsStartTime float64,
) (videoURL string, error) // Returns video only (no audio), all audio tracks are separate
```

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: Task 2.2

**Estimated Time**: 2 hours

---

### Task 3.2: Remove Audio Mixing from Video Composition

**Description**: Remove background music mixing from video composition. All tracks remain separate.

**Acceptance Criteria**:

- [ ] Remove background music mixing from `composeVideo()` function
- [ ] Video composition now only concatenates clips and adds text overlay
- [ ] Background music remains separate file (from Minimax, already generated)
- [ ] Narrator audio remains separate file (from OpenAI TTS, already generated)
- [ ] Update comments to reflect 4-track architecture
- [ ] Video output has no audio (`-an` flag in ffmpeg)

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: Task 3.1

**Estimated Time**: 2 hours

---

### Task 3.3: Implement Text Overlay with ffmpeg

**Description**: Add side effects text overlay with proper escaping, word wrapping, and font fallback.

**Acceptance Criteria**:

- [ ] `escapeFfmpegText()` helper function implemented
- [ ] Font fallback strategy implemented (DejaVu → Liberation → default)
- [ ] Text properties configured (36px, white, black stroke, bottom 20%)
- [ ] Word wrapping at 80% video width
- [ ] Text visible only during side effects segment
- [ ] Special characters escaped correctly
- [ ] Font size fallback if text exceeds 4 lines (36px → 32px)
- [ ] Tested on all aspect ratios (16:9, 9:16, 1:1)

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: Task 3.2

**Note**: All audio tracks (music and narrator) are separate files. Frontend handles mixing and playback synchronization.

**Estimated Time**: 6 hours

---

### Task 3.4: Add S3 Key Helper Functions

**Description**: Create helper functions for consistent S3 key naming.

**Acceptance Criteria**:

- [ ] `buildNarratorAudioKey(userID, jobID)` function created
- [ ] S3 folder structure documented in code comments
- [ ] Consistent key naming across codebase
- [ ] Unit tests for key generation

**S3 Structure**:

```
users/{userID}/jobs/{jobID}/
  ├── clips/scene-001.mp4
  ├── audio/background-music.mp3
  ├── audio/narrator-voiceover.mp3
  ├── thumbnails/scene-001.jpg
  └── final/video.mp4
```

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: Task 3.3

**Estimated Time**: 2 hours

---

### Task 3.5: Integrate Product Image for Last Scene Only

**Description**: Use product image exclusively for the last scene (side effects segment).

**Acceptance Criteria**:

- [ ] Scene generation loop updated
- [ ] Last scene uses `req.StartImage` (product image)
- [ ] Other scenes use `lastFrameURL` (continuity)
- [ ] Image-to-video quality verified on all aspect ratios
- [ ] Product image appears correctly in final video

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`

**Dependencies**: None (can be parallel with Task 3.1-3.4)

**Estimated Time**: 2 hours

---

## Phase 4: Frontend Integration (Week 2-3)

### Task 4.1: Update Create Page with Voice and Side Effects

**Description**: Enable voice selection and side effects input in the Create form.

**Acceptance Criteria**:

- [ ] Remove TODO comments for voice and side effects
- [ ] Add `voice` to API request parameters
- [ ] Add `side_effects` to API request parameters
- [ ] Validation enforces voice selection (required)
- [ ] Validation enforces side effects text (required, 10-500 chars)
- [ ] Character counter for side effects input
- [ ] Product image upload flow implemented (presigned S3 URL)
- [ ] Console logging includes voice and side effects

**Files to Modify**:

- `frontend/src/pages/Create.jsx`

**Dependencies**: Task 1.3

**Estimated Time**: 4 hours

---

### Task 4.2: Update Timeline Component for Four Tracks

**Description**: Display four separate tracks: Video, Music, Audio (narrator), and Text.

**Acceptance Criteria**:

- [ ] `backgroundMusicUrl` prop added to Timeline
- [ ] `narratorAudioUrl` prop added to Timeline
- [ ] `sideEffectsText` prop added to Timeline
- [ ] `sideEffectsStartTime` prop added to Timeline
- [ ] Timeline displays four separate tracks:
  - **Video Track**: Scene segments (no audio)
  - **Music Track**: Background music (continuous, 0-duration)
  - **Audio Track**: Narrator voiceover (continuous, 0-duration)
  - **Text Track**: Side effects text (last 20% only)
- [ ] Visual distinction between tracks (icons: video, music, microphone, text)
- [ ] PropTypes updated
- [ ] Timeline tested with all four tracks

**Files to Modify**:

- `frontend/src/components/workspace/Timeline.jsx`

**Dependencies**: None

**Estimated Time**: 3 hours

---

### Task 4.3: Update Workspace Page with Generation State Handling

**Description**: Handle workspace behavior during generation, completion, and failure states with four-track support.

**Acceptance Criteria**:

- [ ] Extract `video_url` from job data (video track)
- [ ] Extract `audio_url` from job data (background music track)
- [ ] Extract `narrator_audio_url` from job data (narrator track)
- [ ] Extract `side_effects_text` from job data (text track)
- [ ] Extract `side_effects_start_time` from job data
- [ ] Show progress indicator when `status === "processing"`
- [ ] Hide video player/timeline during generation
- [ ] Poll job status every 2 seconds
- [ ] Display error message with retry button on failure
- [ ] Show video player + timeline when `status === "complete"`
- [ ] Pass all four track data to Timeline component

**Files to Modify**:

- `frontend/src/pages/Workspace.jsx`

**Dependencies**: Task 4.2

**Estimated Time**: 4 hours

---

### Task 4.4: Implement Synchronized Four-Track Playback

**Description**: Synchronize video, music, and narrator tracks playback with frontend volume control.

**Acceptance Criteria**:

- [ ] `<video>` element for video track (no audio)
- [ ] `<audio>` element for background music track (30% volume)
- [ ] `<audio>` element for narrator track (100% volume)
- [ ] Play/pause/seek events synchronized across all three media elements
- [ ] Volume control implemented:
  - Background music: 30% volume (0.3)
  - Narrator: 100% volume (1.0)
- [ ] Drift detection: Check every 500ms for all tracks
- [ ] Auto-resync if drift > 200ms
- [ ] Warning toast if drift > 500ms
- [ ] Edge cases handled:
  - Audio loading delays (music or narrator)
  - Duration mismatches between tracks
  - Network buffering on any track
- [ ] Synchronization accuracy within ±100ms

**Files to Modify**:

- `frontend/src/pages/Workspace.jsx`

**Dependencies**: Task 4.3

**Estimated Time**: 6 hours

---

### Task 4.5: Update Progress Tracking for Narrator Stage

**Description**: Add narrator generation stage to progress calculation and display.

**Acceptance Criteria**:

- [ ] `calculateDynamicProgress()` updated with narrator stages
- [ ] `formatStageName()` handles "narrator_generating", "narrator_complete"
- [ ] Progress allocation:
  - Script: 5% (0-5%)
  - Narrator: 3% (5-8%)
  - Scenes: 72% (8-80%, divided by scene count)
  - Music: 5% (80-85%)
  - Composition: 15% (85-100%)
- [ ] `buildStagesCompleted()` includes narrator stage
- [ ] `buildStagesPending()` includes narrator stage
- [ ] Scene-by-scene progress calculation working
- [ ] Progress display tested during generation

**Files to Modify**:

- `backend/internal/api/handlers/progress.go`
- `backend/internal/api/handlers/jobs.go`

**Dependencies**: Task 2.1

**Estimated Time**: 3 hours

---

## Phase 5: Testing & Refinement (Week 3)

### Task 5.1: End-to-End Pipeline Testing

**Description**: Comprehensive testing of the entire video generation pipeline.

**Acceptance Criteria**:

- [ ] Test with male voice
- [ ] Test with female voice
- [ ] Test durations: 10s, 20s, 30s, 60s
- [ ] Test aspect ratios: 16:9, 9:16, 1:1
- [ ] Verify side effects text appears on last frame
- [ ] Verify side effects audio plays at end (1.4x speed)
- [ ] Verify narrator audio plays during main content (1.0x speed)
- [ ] Verify background music audible at 30% volume (separate track, controlled by frontend)
- [ ] Verify product image appears in last scene only
- [ ] Test various side effects text lengths (10-500 chars)
- [ ] All test cases documented

**Files to Create**:

- `backend/test/integration/pharmaceutical_ad_test.go`

**Files to Modify**:

- None

**Dependencies**: All Phase 1-4 tasks

**Estimated Time**: 8 hours

---

### Task 5.2: Audio Quality and Synchronization Testing

**Description**: Verify audio mixing quality and synchronization accuracy.

**Acceptance Criteria**:

- [ ] Background music level verified (30%, separate track, controlled by frontend)
- [ ] Narrator audio level verified (100%, separate track, controlled by frontend)
- [ ] Side effects speed verified (1.4x, fast but clear)
- [ ] No audio clipping or distortion
- [ ] Synchronization tested (video + narrator within ±100ms)
- [ ] Tested on devices: desktop, mobile, headphones, speakers
- [ ] Audio balance confirmed: narrator clear over music

**Files to Modify**:

- None (testing only)

**Dependencies**: Task 4.4, Task 5.1

**Estimated Time**: 4 hours

---

### Task 5.3: Text Overlay Testing

**Description**: Verify text overlay readability and proper rendering on all configurations.

**Acceptance Criteria**:

- [ ] Test on 16:9 videos (1920x1080)
- [ ] Test on 9:16 videos (1080x1920)
- [ ] Test on 1:1 videos (1080x1080)
- [ ] Short text (50 chars): Single line, centered
- [ ] Medium text (200 chars): 2-3 lines, wrapped correctly
- [ ] Long text (500 chars): 4+ lines, readable
- [ ] Text position verified (bottom 20%, centered)
- [ ] Text stroke verified (2px black outline visible)
- [ ] Special characters tested (apostrophes, quotes, colons)
- [ ] Word wrapping at 80% width verified

**Files to Modify**:

- None (testing only)

**Dependencies**: Task 3.3, Task 5.1

**Estimated Time**: 3 hours

---

### Task 5.4: Error Handling and Validation Testing

**Description**: Test all error scenarios and validation rules.

**Acceptance Criteria**:

- [ ] TTS API failure → Job fails with clear message
- [ ] GPT-4o failure → Job fails during script generation
- [ ] Kling API failure → Job fails during scene generation
- [ ] Invalid product image → Validation error before job creation
- [ ] Side effects text too long (>500 chars) → Validation error
- [ ] Missing voice parameter → Validation error
- [ ] Missing side effects → Validation error
- [ ] Duration not divisible by 5 → Validation error
- [ ] All errors surfaced to user in Create page
- [ ] No partial videos created on failure
- [ ] Error messages user-friendly and actionable

**Files to Modify**:

- None (testing only)

**Dependencies**: Task 1.3, Task 5.1

**Estimated Time**: 4 hours

---

## Phase 6: Documentation & Deployment (Week 4)

### Task 6.1: Update API Documentation

**Description**: Update Swagger documentation with new API fields and examples.

**Acceptance Criteria**:

- [ ] `voice` parameter documented (required, enum: male/female)
- [ ] `side_effects` parameter documented (required, 10-500 chars)
- [ ] `narrator_audio_url` response field documented
- [ ] `side_effects_text` response field documented
- [ ] `side_effects_start_time` response field documented
- [ ] Example requests with voice and side effects
- [ ] Error codes for validation failures documented
- [ ] Swagger UI tested

**Files to Modify**:

- `backend/docs/swagger.yaml`
- `backend/docs/swagger.json`

**Dependencies**: Task 1.3

**Estimated Time**: 3 hours

---

### Task 6.2: Add Code Documentation and Comments

**Description**: Add comprehensive inline documentation for new features.

**Acceptance Criteria**:

- [ ] TTS adapter documented (voice mapping, API calls)
- [ ] Video composition documented (text overlay, no audio mixing - all tracks separate)
- [ ] Text overlay documented (ffmpeg filters, escaping)
- [ ] Narrator script generation documented
- [ ] Font fallback strategy documented
- [ ] S3 key structure documented
- [ ] README updated with pharmaceutical ad features

**Files to Modify**:

- `backend/internal/adapters/tts_adapter.go`
- `backend/internal/api/handlers/generate_async.go`
- `backend/internal/prompts/ad_script_prompt.go`
- `README.md`

**Dependencies**: All implementation tasks

**Estimated Time**: 4 hours

---

### Task 6.3: Environment Configuration and Deployment Setup

**Description**: Configure environment variables and deployment for production.

**Acceptance Criteria**:

- [ ] TTS API key added to Secrets Manager
- [ ] TTS provider configuration documented
- [ ] Environment variables documented:
  - `TTS_API_KEY` or Secrets Manager ARN
  - `TTS_PROVIDER` (default: "openai")
- [ ] Deployment docs updated with new dependencies
- [ ] ffmpeg font requirements documented
- [ ] Docker image includes DejaVu Sans font
- [ ] Font validation added to deployment script

**Files to Modify**:

- `backend/Dockerfile`
- `infrastructure/` (if using Terraform)
- `README.md`

**Files to Create**:

- `backend/scripts/validate-fonts.sh`

**Dependencies**: All implementation tasks

**Estimated Time**: 3 hours

---

### Task 6.4: Implement Structured Logging

**Description**: Add comprehensive structured logging throughout the pipeline.

**Acceptance Criteria**:

- [ ] JSON logs at each pipeline stage:
  - Script generation start/complete
  - Narrator generation start/complete
  - Each scene generation start/complete
  - Music generation start/complete
  - Composition start/complete
- [ ] Log format: `{job_id, user_id, stage, duration_ms, status, error}`
- [ ] CloudWatch Logs integration configured
- [ ] CloudWatch alarms set up:
  - TTS timeout (narrator_generating > 60s)
  - Kling API errors (scene failures)
  - Composition failures
- [ ] CloudWatch dashboard created:
  - Job completion rate
  - Average stage timings
  - Error breakdown

**Files to Modify**:

- `backend/internal/api/handlers/generate_async.go`
- `backend/internal/api/handlers/generate.go`
- `backend/pkg/logger/logger.go`

**Files to Create**:

- `infrastructure/modules/monitoring/cloudwatch-dashboard.tf` (if using Terraform)

**Dependencies**: All implementation tasks

**Estimated Time**: 5 hours

---

### Task 6.5: Implement S3 Presigned URL Strategy

**Description**: Implement secure presigned URLs for asset access with 7-day expiration.

**Acceptance Criteria**:

- [ ] Presigned URL generation in GetJob handler
- [ ] 7-day expiration configured
- [ ] Regenerate URLs on each API request
- [ ] All asset URLs use presigned URLs:
  - video_url
  - narrator_audio_url
  - audio_url (background music track, separate)
  - narrator_audio_url (narrator track, separate)
  - thumbnail URLs
- [ ] No public bucket policy needed
- [ ] Performance tested (~5ms overhead per URL)

**Files to Modify**:

- `backend/internal/api/handlers/jobs.go`
- `backend/internal/repository/s3.go`

**Dependencies**: None (can be parallel)

**Estimated Time**: 3 hours

---

## Summary

### Total Tasks: 30

### By Phase:

- **Phase 1** (Backend Foundation): 5 tasks, ~14 hours
- **Phase 2** (Audio Pipeline): 2 tasks, ~9 hours
- **Phase 3** (Video Composition): 5 tasks, ~16 hours
- **Phase 4** (Frontend Integration): 5 tasks, ~20 hours
- **Phase 5** (Testing): 4 tasks, ~19 hours
- **Phase 6** (Documentation & Deployment): 5 tasks, ~18 hours

### Total Estimated Time: ~96 hours (12-13 days, 1 developer)

### Critical Path:

1. Task 1.1 → Task 1.5 → Task 2.1 → Task 2.2 → Task 3.1 → Task 3.2 → Task 3.3 → Task 4.4 → Task 5.1

### Parallel Opportunities:

- Task 3.5 (Product Image) can run parallel with Task 3.1-3.4
- Task 4.1-4.2 (Frontend) can start after Task 1.3 completes
- Task 6.5 (S3 URLs) can be done anytime
- Task 6.1-6.2 (Documentation) can be done in parallel with testing

---

## Getting Started

### Prerequisites

1. ✅ PRD reviewed and approved
2. ✅ Development environment set up
3. ✅ AWS credentials configured
4. ✅ OpenAI API key obtained

### Recommended Workflow

1. **Week 1**: Complete Phase 1 (Backend Foundation)
2. **Week 2**: Complete Phase 2 (Audio Pipeline) + Start Phase 3 (Video Composition)
3. **Week 3**: Complete Phase 3 + Phase 4 (Frontend Integration) + Start Phase 5 (Testing)
4. **Week 4**: Complete Phase 5 (Testing) + Phase 6 (Documentation & Deployment)

### Daily Standups

- Review completed tasks
- Identify blockers
- Adjust estimates as needed
- Update task status

---

## Task Status Legend

- [ ] Not Started
- [🔄] In Progress
- [✅] Completed
- [⚠️] Blocked
- [❌] Cancelled

---

**Document Version**: 1.0  
**Last Updated**: 2024-01-15  
**Next Review**: After Phase 1 completion
