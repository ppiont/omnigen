# Video Generation Pipeline - Process Overview

## What This Does
Transforms text prompts into professional multi-clip video advertisements with AI-powered narrative planning, voiceovers, and automatic stitching.

---

## Process Flow: User Prompt → Final Video

### Stage 1: Content Planning (`promptParser.js`)
**Purpose:** Structure user prompt into narrative scenes

**Input:** Text prompt + creative options (style, tone, tempo)

**Process:**
- GPT-4o analyzes prompt and generates scene breakdowns
- Creates narrative arc (setup → development → resolution)
- Anti-pattern detection prevents repetitive scenes
- Each scene gets: prompt, description, duration, structure (camera, lighting, mood, etc.)

**Output:** Structured scenes array

**Example:**
```
Prompt: "athlete drinking energy drink"
→ Scene 1: Exhausted athlete in locker room
→ Scene 2: Drinking energy drink, energized
→ Scene 3: Back on court performing
```

---

### Stage 2: Keyframe Generation (Optional, `keyframeGenerator.js`)
**Purpose:** Prevent style drift across clips

**Process:**
- Generates master style reference image (SDXL)
- Creates consistent keyframe for each scene
- Saves to `./keyframes/`

**When:** Use `--keyframes` flag

---

### Stage 3: Video Generation (`videoGenerator.js`)
**Purpose:** Generate video clips with visual continuity

**For each scene:**

1. **Determine Input Image:**
   - Priority 1: Pre-generated keyframe (if using `--keyframes`)
   - Priority 2: User-provided image (first scene only)
   - Priority 3: Last frame from previous clip

2. **Generate Video:**
   - Call Replicate API (Minimax/PixVerse/Google Veo/etc.)
   - Poll for completion (~60-180s per clip)
   - Returns video URL

3. **Extract Last Frame** (`frameExtractor.js`):
   - Download video temporarily
   - FFmpeg extracts last frame: `ffmpeg -sseof -1 -i video.mp4 -frames:v 1 frame.jpg`
   - Save to `./temp/last_frame_N.jpg`
   - Used as input for next clip

**Continuity Methods:**
- **Last-Frame Chaining (default):** Each clip uses previous clip's last frame
- **Keyframe-First:** Each clip uses pre-generated keyframe

---

### Stage 4: Voiceover Generation (Optional, `audioGenerator.js`)
**Purpose:** Add professional narration to videos

**For each video clip:**

1. **Script Generation:**
   - GPT-4o-mini creates "show don't tell" ad copy (20-25 words)
   - Focuses on emotion/aspiration, not scene description
   - Example: "Every champion needs fuel that matches their dedication"

2. **Text-to-Speech:**
   - Minimax Speech 02 HD synthesis
   - Extracts emotion from scene mood (calm, happy, energetic, etc.)
   - 10+ voice options available
   - Saves to `./temp/voiceover_N.mp3`

3. **Video-Audio Merge:**
   - Replicate video-audio-merge model
   - Combines video URL + audio file
   - Downloads final video to `./output/clip_N_with_audio_*.mp4`

**When:** Use `--voiceover` flag

---

### Stage 5: Download & Output (`downloader.js`)
**Purpose:** Save all generated content locally

**Process:**
- Downloads all videos from Replicate URLs
- Saves to `./output/` directory
- Creates `summary_*.json` with metadata

---

### Stage 6: Video Stitching (`videoStitcher.js`)
**Purpose:** Combine multiple clips into single video

**Process:**
- Creates FFmpeg concat list file
- Lossless concatenation: `ffmpeg -f concat -safe 0 -i list.txt -c copy output.mp4`
- Saves to `./output/final_stitched_*.mp4`

**When:** Only if multiple clips generated

---

### Stage 7: Metadata Export (`xmlExporter.js`)
**Purpose:** Export scene structure for post-production

**Process:**
- Creates XML with scene descriptions, prompts, camera/lighting/mood info
- Saves to `./output/scene_structure_*.xml`

---

## Flow Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                       USER INPUT                              │
│  Text prompt + Options (--clips, --voiceover, --style, etc.) │
└───────────────────────────┬──────────────────────────────────┘
                            │
                            ▼
                  ┌─────────────────────┐
                  │  STAGE 1: PLANNING  │
                  │   (PromptParser)    │
                  │   • GPT-4o analysis │
                  │   • Narrative arc   │
                  │   • Scene structure │
                  └──────────┬──────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
   ┌──────────────────────┐    ┌──────────────────────┐
   │  STAGE 2: KEYFRAMES  │    │   SKIP KEYFRAMES     │
   │  (KeyframeGenerator) │    │   (Use last-frame    │
   │  • Style reference   │    │    continuity)       │
   │  • SDXL generation   │    │                      │
   └──────────┬───────────┘    └──────────┬───────────┘
              │                           │
              └─────────────┬─────────────┘
                            ▼
                  ┌─────────────────────────┐
                  │ STAGE 3: VIDEO GEN      │
                  │  (VideoGenerator)       │
                  │  For each scene:        │
                  │  1. Get input image     │
                  │  2. Replicate API call  │
                  │  3. Extract last frame  │
                  └──────────┬──────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
   ┌──────────────────────┐    ┌──────────────────────┐
   │  STAGE 4: VOICEOVER  │    │  SKIP VOICEOVER      │
   │  (AudioGenerator)    │    │                      │
   │  • Script (GPT-4o)   │    │                      │
   │  • TTS (Minimax)     │    │                      │
   │  • Merge video+audio │    │                      │
   └──────────┬───────────┘    └──────────┬───────────┘
              │                           │
              └─────────────┬─────────────┘
                            ▼
                  ┌─────────────────────────┐
                  │   STAGE 5: DOWNLOAD     │
                  │   (VideoDownloader)     │
                  │   • Save all videos     │
                  │   • Create summary JSON │
                  └──────────┬──────────────┘
                            │
                            ▼
                  ┌─────────────────────────┐
                  │   STAGE 6: STITCHING    │
                  │   (VideoStitcher)       │
                  │   • FFmpeg concat       │
                  │   • Lossless merge      │
                  └──────────┬──────────────┘
                            │
                            ▼
                  ┌─────────────────────────┐
                  │  STAGE 7: XML EXPORT    │
                  │   (XMLExporter)         │
                  │   • Scene metadata      │
                  └──────────┬──────────────┘
                            │
                            ▼
             ┌──────────────────────────────┐
             │       FINAL OUTPUTS          │
             │  • final_stitched_*.mp4      │
             │  • clip_*_with_audio_*.mp4   │
             │  • scene_structure_*.xml     │
             │  • summary_*.json            │
             └──────────────────────────────┘
```

---

## Data Flow Through Pipeline

```
User Prompt
    ↓
[GPT-4o] → Scenes Array: [{prompt, description, structure}]
    ↓
[Optional: SDXL] → Keyframes: [image1.jpg, image2.jpg, ...]
    ↓
[Replicate Video API] → Video URLs + Last Frames
    ↓
[Optional: GPT-4o-mini + Minimax TTS] → Videos with Voiceovers
    ↓
[Download] → Local MP4 files in ./output/
    ↓
[FFmpeg Concat] → final_stitched.mp4
    ↓
[XML Export] → scene_structure.xml
```

---

## Key Components

| Component | File | Purpose |
|-----------|------|---------|
| **PromptParser** | `src/promptParser.js` | Parse prompt into narrative scenes via GPT-4o |
| **KeyframeGenerator** | `src/keyframeGenerator.js` | Generate style-consistent keyframes via SDXL |
| **VideoGenerator** | `src/videoGenerator.js` | Generate video clips via Replicate (Minimax/PixVerse/etc.) |
| **FrameExtractor** | `src/frameExtractor.js` | Extract last frame from videos for continuity |
| **AudioGenerator** | `src/audioGenerator.js` | Generate scripts (GPT-4o-mini) + voiceover (TTS) + merge |
| **VideoDownloader** | `src/downloader.js` | Download videos from URLs to local storage |
| **VideoStitcher** | `src/videoStitcher.js` | Concatenate clips with FFmpeg |
| **XMLExporter** | `src/xmlExporter.js` | Export scene metadata to XML |

---

## Models & APIs Used

### Video Generation (Replicate)
- **Minimax Hailuo 2.3** - 6-10s clips, $0.28-0.56/clip
- **PixVerse v5** - 5-8s clips, ~$0.30/clip
- **Seedance Pro** - 5-10s clips, budget-friendly
- **Google Veo 3.1** - 8s clips, premium
- **Wan 2.5** - 5-8s clips, text-only (no image input)

### AI Services
- **GPT-4o** (OpenAI) - Content planning, narrative structure
- **GPT-4o-mini** (OpenAI) - Voiceover script generation
- **Minimax Speech 02 HD** (Replicate) - Text-to-speech synthesis
- **SDXL** (Replicate/Stability AI) - Keyframe image generation

### Local Tools
- **FFmpeg** - Frame extraction, video concatenation

---

## CLI Usage

```bash
# Single clip with voiceover
node cli.js generate "athlete drinking energy drink" \
  --model minimax \
  --voiceover \
  --voice Deep_Voice_Man

# Multi-clip sequence with style consistency
node cli.js generate "product launch journey" \
  --model minimax \
  --clips 3 \
  --keyframes \
  --style cinematic \
  --tone premium \
  --tempo medium

# Full creative control
node cli.js generate "tech startup story" \
  --model pixverse \
  --clips 5 \
  --voiceover \
  --voice Friendly_Person \
  --style documentary \
  --tone inspiring \
  --tempo fast \
  --creative-boost
```

---

## Output Structure

```
./output/
├── final_stitched_20250115_123456.mp4    # Combined video
├── clip_1_with_audio_20250115_123456.mp4 # Individual clips
├── clip_2_with_audio_20250115_123456.mp4
├── clip_3_with_audio_20250115_123456.mp4
├── scene_structure_20250115_123456.xml   # Metadata
└── summary_20250115_123456.json          # Generation details

./keyframes/ (if --keyframes used)
├── style_reference.jpg
├── keyframe_0.jpg
├── keyframe_1.jpg
└── keyframe_2.jpg

./temp/ (cleaned unless --keep-frames)
├── last_frame_0.jpg
├── last_frame_1.jpg
└── voiceover_*.mp3
```

---

## Performance Characteristics

### Timing (Approximate)
- **1 clip:** ~75s (60s video gen + 15s voiceover)
- **3 clips:** ~230s (180s video gen + 45s voiceover + 5s stitch)
- **5 clips:** ~385s (300s video gen + 75s voiceover + 10s stitch)

### Costs (Per Clip)
- **Video generation:** $0.28-0.56
- **Voiceover:** ~$0.05
- **Audio merge:** ~$0.01
- **Total:** ~$0.62/clip

---

## Creative Options

### Style
`cinematic`, `documentary`, `energetic`, `minimal`, `dramatic`, `playful`

### Tone
`premium`, `friendly`, `edgy`, `inspiring`, `humorous`

### Tempo
`slow`, `medium`, `fast`

### Voices
**English:** Deep_Voice_Man, Friendly_Person, Calm_Woman, Casual_Guy, Wise_Woman, Inspirational_girl, Determined_Man, Lively_Girl, Sweet_Girl_2, Elegant_Man

**Chinese:** male-qn-jingying, male-qn-qingse, female-shaonv, female-yujie, female-chengshu, male-qn-daxuesheng

---

## Environment Setup

Required:
```bash
REPLICATE_API_TOKEN=<your_token>
```

Optional (but recommended):
```bash
OPENAI_API_KEY=<your_key>  # For content planning and scripts
```

Without OpenAI, uses fallback simple parsing (no narrative structure).

---

## Key Features

1. **Narrative Intelligence** - GPT-4o creates story arcs with anti-pattern detection
2. **Visual Continuity** - Last-frame extraction chains clips seamlessly
3. **Style Consistency** - Optional keyframe generation prevents drift
4. **Professional Audio** - "Show don't tell" ad copy with emotional TTS
5. **Automatic Assembly** - FFmpeg stitches final product
6. **Structured Output** - XML metadata for post-production workflows

---

## Entry Point

**File:** `cli.js`

**Commands:**
- `generate <prompt>` - Main video generation pipeline
- `models` - List available video generation models
- `voices` - List available voiceover voices
