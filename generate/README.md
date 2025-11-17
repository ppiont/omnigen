# AI Video Advertisement Generator

**Production-ready CLI tool for generating professional video advertisements using AI.** Create multi-clip video sequences with narrative continuity, AI voiceovers, and automatic stitching.

---

## 🎯 Overview

Transforms simple text prompts into professional video advertisements:
- **Content Planning** with GPT-4o for narrative story arcs
- **Video Generation** using AI models (Minimax, PixVerse, Seedance, Veo)
- **AI Voiceovers** with XTTS-v2 voice cloning and "show don't tell" ad copy
- **Visual Continuity** through frame extraction and keyframes
- **Automatic Stitching** into final MP4

**Use Cases:** Product ads, brand stories, social media content, marketing campaigns

---

## 🚀 Quick Start

### Prerequisites
- Node.js 18+
- FFmpeg
- Replicate API token
- OpenAI API key (optional but recommended)

### Installation
```bash
npm install
cp .env.example .env
# Add API keys to .env
```

### Usage
```bash
# Single clip
node cli.js generate "luxury watch reveal" --model minimax

# Multi-clip with voiceover (requires speaker audio)
node cli.js generate "athlete energy drink" --clips 3 --voiceover --speaker ./my_voice.wav

# Custom creative style with voice cloning
node cli.js generate "kids cereal" --clips 3 --style playful --tone humorous --voiceover --speaker ./voice.wav
```

---

## 📋 Features

### 1. Intelligent Content Planning (GPT-4o)
- Story arc generation (setup → development → resolution)
- Camera progression (wide → medium → close)
- Mood evolution across clips
- Anti-pattern detection (prevents repetitive scenes)

**Creative Controls:**
- `--style`: cinematic, documentary, energetic, minimal, dramatic, playful
- `--tone`: premium, friendly, edgy, inspiring, humorous
- `--tempo`: slow, medium, fast
- `--creative`: Boost AI creativity

### 2. Video Generation (Replicate)
| Model | Duration | Cost | Image Input |
|-------|----------|------|-------------|
| Minimax 2.3 | 6-10s | $0.28-0.56 | ✅ |
| PixVerse v5 | 5-8s | ~$0.30 | ✅ |
| Veo 3.1 | 8s | Premium | ✅ |
| Wan 2.5 | 5-8s | Cheap | ❌ |

### 3. Visual Continuity
**A. Last-Frame Chaining (Default)**
```bash
node cli.js generate "story" --clips 3 --model minimax
```

**B. Keyframe-First Generation**
```bash
node cli.js generate "story" --clips 3 --keyframes
node cli.js generate "story" --clips 3 --keyframes --style-ref ./brand.jpg
```

### 4. AI Voiceover with Narrative Arc (XTTS-v2 Voice Cloning)
- GPT-4o-mini generates ad copy with **narrative progression** (not scene description)
- **Arc-aware**: Each clip advances the story (hook → develop → resolve → CTA)
- **No repetition**: AI understands full story context to avoid saying the same thing 4 times
- **Voice cloning** - clone any voice from audio sample (6+ seconds)
- 16 supported languages (en, es, fr, de, it, pt, pl, tr, ru, nl, cs, ar, zh, hu, ko, hi)
- Auto-sync with video

**Narrative Progression Example (4 clips):**
- Clip 1 (Hook): "Pros don't crash when it counts"
- Clip 2 (Problem): "But cheap fuel fails when pressure peaks"
- Clip 3 (Solution): "CrackFuel. Zero sugar. No crash. Pure focus"
- Clip 4 (CTA): "Grind unstoppable. Twenty percent off now"

**"Show Don't Tell" Example:**
❌ Bad: "Kobe enters locker room, sits down"
✅ Good: "Every champion knows: greatness starts with the right fuel"

```bash
# With custom voice
node cli.js generate "coffee brewing" --voiceover --speaker ./my_voice.wav --language en

# Using default voice (place default_speaker file in ./temp/)
# Supports: default_speaker.wav, default_speaker.mp3, default_speaker.m4a, etc.
node cli.js generate "coffee brewing" --voiceover

node cli.js voices  # Show voice cloning info
```

### 5. Automatic Stitching with Smooth Transitions
Multi-clip sequences automatically combine into single MP4 with optional transition effects

**A. Basic Stitching (Fast, No Transitions)**
```bash
# Simple concatenation
node cli.js generate "product story" --clips 3 --voiceover
```

**B. Frame Interpolation (Smooth Motion)**
```bash
# Uses FFmpeg minterpolate to create intermediate frames
# Reduces jarring cuts between clips
node cli.js generate "product story" --clips 3 --voiceover --smooth-transitions

# Custom FPS for interpolation (default: 60)
node cli.js generate "product story" --clips 3 --voiceover --smooth-transitions --interpolation-fps 120
```

**C. Crossfade Transitions (Professional Look)**
```bash
# Fade duration in seconds (default: 1.0)
node cli.js generate "product story" --clips 3 --voiceover --crossfade 1.5
```

### 6. Structured Output
- **XML metadata**: Scene structure for post-production
- **JSON summary**: Generation details
- **Individual clips**: Editable components

---

## 📖 CLI Reference

### Commands

**generate** - Generate video advertisement
```bash
node cli.js generate <prompt> [options]
```

**Options:**
| Flag | Description | Default |
|------|-------------|---------|
| `-m, --model` | minimax/pixverse/seedance/veo/wan | minimax |
| `-c, --clips` | Number of clips | 1 |
| `-i, --image` | Input image path | none |
| `-o, --output` | Output directory | ./output |
| `--voiceover` | Add AI voiceover (XTTS-v2) | off |
| `--speaker` | Speaker audio file for cloning | none |
| `--language` | Voiceover language (en/es/fr/etc) | en |
| `--style` | Visual style | cinematic |
| `--tone` | Ad tone | premium |
| `--tempo` | Pacing | medium |
| `--creative` | Boost creativity | off |
| `--keyframes` | Style consistency | off |
| `--style-ref` | Reference image | none |
| `--smooth-transitions` | Frame interpolation | off |
| `--crossfade <duration>` | Crossfade transitions (seconds) | none |
| `--interpolation-fps <fps>` | Target FPS for interpolation | 60 |

**models** - List available models
```bash
node cli.js models
```

**voices** - Show XTTS-v2 voice cloning info
```bash
node cli.js voices
```

---

## 🏗️ Architecture

See `ARCHITECTURE.md` for detailed diagrams.

**High-Level Flow:**
```
User Prompt
    ↓
GPT-4o Planning (story arc, scenes)
    ↓
Keyframe Gen (optional, SDXL)
    ↓
Video Gen (Replicate, per scene)
    ↓
Voiceover (GPT-4o-mini + XTTS-v2 Voice Cloning)
    ↓
Download & Stitch (FFmpeg)
    ↓
Final MP4 + Metadata
```

**File Structure:**
```
my-replicate-app/
├── cli.js                 # Entry point
├── src/
│   ├── promptParser.js    # GPT-4o planning
│   ├── videoGenerator.js  # Replicate video
│   ├── audioGenerator.js  # TTS + merge
│   ├── videoStitcher.js   # FFmpeg concat
│   └── ...
├── output/                # Generated videos
└── .env                   # API keys
```

---

## 🔌 Backend Integration

### API Wrapper
```javascript
import { spawn } from 'child_process';

app.post('/api/generate', (req, res) => {
  const { prompt, clips, model, voiceover } = req.body;
  
  const process = spawn('node', [
    'cli.js', 'generate', prompt,
    '--clips', clips,
    '--model', model,
    voiceover ? '--voiceover' : ''
  ]);
  
  // Stream progress, return video URL
});
```

### Direct Import
```javascript
import { PromptParser } from './src/promptParser.js';
import { VideoGenerator } from './src/videoGenerator.js';

export class VideoService {
  async generate(prompt, options) {
    const scenes = await this.parser.planContent(prompt, options.clips);
    const results = await this.generator.generateSequence(options.model, scenes);
    return results;
  }
}
```

### Database Schema
```sql
CREATE TABLE video_jobs (
  id UUID PRIMARY KEY,
  user_id UUID,
  prompt TEXT,
  options JSONB,
  status VARCHAR(50),
  video_url TEXT,
  created_at TIMESTAMP
);

CREATE TABLE video_clips (
  id UUID PRIMARY KEY,
  job_id UUID REFERENCES video_jobs(id),
  clip_number INT,
  video_url TEXT,
  voiceover_script TEXT
);
```

### Frontend Integration
```javascript
// React example
async function generateVideo(prompt, options) {
  const response = await fetch('/api/video/generate', {
    method: 'POST',
    body: JSON.stringify({ prompt, options })
  });
  return response.json();
}

// WebSocket progress
socket.on('video:progress', (msg) => setProgress(msg));
socket.on('video:complete', (data) => setVideoUrl(data.url));
```

---

## 💰 Cost Estimation

**Per-Clip (Minimax 768p, 10s):**
- Video: $0.56
- Voiceover (XTTS-v2): $0.0023 (~2s generation)
- Merge: $0.01
- **Total: ~$0.57/clip**

**3-clip ad:** ~$1.71
**5-clip story:** ~$2.85

**Optimization:**
- Use Wan model (cheaper, text-only): $0.54/clip
- Skip keyframes: save $0.02/clip
- No voiceover: save $0.01/clip

---

## 🐛 Troubleshooting

**FFmpeg not found:**
```bash
brew install ffmpeg  # macOS
sudo apt-get install ffmpeg  # Ubuntu
```

**Module errors:**
```bash
rm -rf node_modules && npm install
```

**API errors:**
```bash
# Check .env file
cat .env
# Should show: REPLICATE_API_TOKEN=r8_...
```

**Debug mode:**
```bash
DEBUG=1 node cli.js generate "test" --clips 3
```

---

## 📊 Performance

| Clips | Video Gen | Voiceover | Stitch | Total |
|-------|-----------|-----------|--------|-------|
| 1 | ~60s | +15s | - | ~75s |
| 3 | ~180s | +45s | +5s | ~230s |
| 5 | ~300s | +75s | +10s | ~385s |

---

## 🚢 Deployment

**Docker:**
```dockerfile
FROM node:18-alpine
RUN apk add --no-cache ffmpeg
COPY . /app
WORKDIR /app
RUN npm ci
CMD ["node", "cli.js"]
```

**Queue System (Bull):**
```javascript
import Queue from 'bull';

const videoQueue = new Queue('video-gen');
videoQueue.process(async (job) => {
  return await generateVideo(job.data);
});

app.post('/api/generate', async (req, res) => {
  const job = await videoQueue.add(req.body);
  res.json({ jobId: job.id });
});
```

---

## 📄 License

MIT License

---

**Built for professional AI-generated video advertisements** 🎬
