# Architecture Documentation

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER INPUT                               │
│            Text Prompt + Configuration Options                   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        CLI INTERFACE                             │
│                         (cli.js)                                 │
│  • Argument parsing                                              │
│  • Environment validation                                        │
│  • Workflow orchestration                                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CONTENT PLANNING LAYER                        │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              PromptParser (src/promptParser.js)            │ │
│  │                                                            │ │
│  │  • GPT-4o API Integration                                 │ │
│  │  • Story Arc Generation                                   │ │
│  │  • Scene Structure Planning                               │ │
│  │  • Style Guide Application                                │ │
│  │  • Anti-pattern Detection                                 │ │
│  │                                                            │ │
│  │  Input: User prompt + creative options                    │ │
│  │  Output: Structured scenes array                          │ │
│  └────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    ┌────────┴────────┐
                    │                 │
         ┌──────────▼──────────┐     │
         │  With --keyframes?  │     │
         └──────────┬──────────┘     │
                    │                 │
              YES   │   NO            │
                    │                 │
         ┌──────────▼──────────┐     │
         │                     │     │
┌────────┴─────────────────────▼─────▼────────────────────────────┐
│                    GENERATION LAYER                              │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │         KeyframeGenerator (src/keyframeGenerator.js)       │ │
│  │  [OPTIONAL - Only if --keyframes flag set]                │ │
│  │                                                            │ │
│  │  • Style Reference Generation (SDXL)                      │ │
│  │  • Per-Scene Keyframe Creation                            │ │
│  │  • Style Consistency Enforcement                          │ │
│  │                                                            │ │
│  │  Input: Scenes + optional style reference image           │ │
│  │  Output: Keyframe images for each scene                   │ │
│  └────────────────────────────────────────────────────────────┘ │
│                             │                                    │
│                             ▼                                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │           VideoGenerator (src/videoGenerator.js)           │ │
│  │                                                            │ │
│  │  • Replicate API Integration                              │ │
│  │  • Model Selection & Configuration                        │ │
│  │  • Sequential Clip Generation                             │ │
│  │  • Progress Tracking & Polling                            │ │
│  │                                                            │ │
│  │  Sub-Component: FrameExtractor                            │ │
│  │  • FFmpeg-based last frame extraction                     │ │
│  │  • Frame-to-frame continuity                              │ │
│  │                                                            │ │
│  │  For each scene:                                          │ │
│  │    1. Select input image (keyframe OR last frame)         │ │
│  │    2. Add previous scene context to prompt                │ │
│  │    3. Generate video via Replicate                        │ │
│  │    4. Extract last frame for next clip                    │ │
│  │                                                            │ │
│  │  Input: Scenes + keyframes/images + model config          │ │
│  │  Output: Video URLs for each scene                        │ │
│  └────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    ┌────────┴────────┐
                    │                 │
         ┌──────────▼──────────┐     │
         │  With --voiceover?  │     │
         └──────────┬──────────┘     │
                    │                 │
              YES   │   NO            │
                    │                 │
         ┌──────────▼──────────┐     │
         │                     │     │
┌────────┴─────────────────────▼─────▼────────────────────────────┐
│                       AUDIO LAYER                                │
│                  [OPTIONAL - Only if --voiceover]                │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │           AudioGenerator (src/audioGenerator.js)           │ │
│  │                                                            │ │
│  │  For each video clip:                                     │ │
│  │                                                            │ │
│  │  Step 1: Generate Ad Copy                                 │ │
│  │  • GPT-4o-mini generates "show don't tell" script         │ │
│  │  • Extracts emotion from scene mood                       │ │
│  │  • Creates compelling ad copy (20-25 words)               │ │
│  │                                                            │ │
│  │  Step 2: Synthesize Speech                                │ │
│  │  • Minimax Speech 02 HD API                               │ │
│  │  • Voice selection + emotion matching                     │ │
│  │  • Download audio file                                    │ │
│  │                                                            │ │
│  │  Step 3: Merge Audio with Video                           │ │
│  │  • lucataco/video-audio-merge model                       │ │
│  │  • Sync audio with video timeline                         │ │
│  │  • Download final merged clip                             │ │
│  │                                                            │ │
│  │  Input: Video URLs + scenes                               │ │
│  │  Output: Videos with voiceover                            │ │
│  └────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       OUTPUT LAYER                               │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │          VideoDownloader (src/downloader.js)               │ │
│  │  • Download all video clips from URLs                     │ │
│  │  • Generate summary.json with metadata                    │ │
│  └────────────────────────────────────────────────────────────┘ │
│                             │                                    │
│                             ▼                                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │         VideoStitcher (src/videoStitcher.js)               │ │
│  │  [Only for multi-clip sequences]                          │ │
│  │                                                            │ │
│  │  • FFmpeg concat demuxer (lossless)                       │ │
│  │  • Stitch all clips into single MP4                       │ │
│  │  • Fallback to re-encode if needed                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                             │                                    │
│                             ▼                                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │            XMLExporter (src/xmlExporter.js)                │ │
│  │  • Export scene metadata                                  │ │
│  │  • Structure for post-production                          │ │
│  └────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        FINAL OUTPUT                              │
│                                                                  │
│  • final_stitched_TIMESTAMP.mp4  (Combined video)               │
│  • clip_N_with_audio_TIMESTAMP.mp4  (Individual clips)          │
│  • scene_structure_TIMESTAMP.xml  (Metadata)                    │
│  • summary.json  (Generation details)                           │
└─────────────────────────────────────────────────────────────────┘
```

## External API Dependencies

```
┌─────────────────────┐
│    OpenAI API       │
│    (GPT-4o/Mini)    │
│                     │
│  • Content Planning │
│  • Ad Copy Gen      │
└─────────────────────┘

┌─────────────────────┐
│   Replicate API     │
│                     │
│  Models Used:       │
│  • Minimax (Video)  │
│  • PixVerse (Video) │
│  • SDXL (Keyframes) │
│  • Minimax Speech   │
│  • Video-Audio Merge│
└─────────────────────┘

┌─────────────────────┐
│      FFmpeg         │
│   (Local Binary)    │
│                     │
│  • Frame Extract    │
│  • Video Stitch     │
└─────────────────────┘
```

## Data Flow Diagram

```
┌──────────────┐
│ User Prompt  │  "athlete drinking energy drink"
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ GPT-4o Content Planning                      │
│ Output: Scenes Array                         │
│ [                                            │
│   {                                          │
│     description: "Athlete tired after game", │
│     prompt: "Close-up of athlete...",        │
│     structure: { mood: "exhausted",... }     │
│   },                                         │
│   { ... scene 2 ... },                       │
│   { ... scene 3 ... }                        │
│ ]                                            │
└──────┬───────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ Keyframe Generation (if --keyframes)         │
│ • Generate style reference (SDXL)            │
│ • Create keyframe per scene                  │
│ Output: keyframes/keyframe_1.jpg, etc.       │
└──────┬───────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ Video Generation Loop (per scene)            │
│                                              │
│ Scene 1:                                     │
│   Input: keyframe_1.jpg OR user image       │
│   Prompt: Enhanced with style guide          │
│   → Replicate Minimax Video                  │
│   Output: video_1_url                        │
│   → Extract last frame → frame_1.jpg         │
│                                              │
│ Scene 2:                                     │
│   Input: keyframe_2.jpg OR frame_1.jpg       │
│   Prompt: Enhanced + "Continuing from..."   │
│   → Replicate Minimax Video                  │
│   Output: video_2_url                        │
│   → Extract last frame → frame_2.jpg         │
│                                              │
│ Scene 3: (repeat)                            │
│                                              │
│ Output: [video_1_url, video_2_url, ...]      │
└──────┬───────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ Voiceover Generation (if --voiceover)        │
│                                              │
│ For each scene:                              │
│   GPT-4o-mini:                               │
│     Input: scene.description + mood          │
│     Output: "Every champion needs fuel"      │
│                                              │
│   Minimax Speech:                            │
│     Input: ad copy + voice + emotion         │
│     Output: audio_1.mp3                      │
│                                              │
│   Video-Audio Merge:                         │
│     Input: video_1_url + audio_1.mp3         │
│     Output: final_video_1_url                │
│                                              │
│ Output: [final_video_1_url, ...]             │
└──────┬───────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ Download All Videos                          │
│ • Download video_1 → clip_1_with_audio.mp4   │
│ • Download video_2 → clip_2_with_audio.mp4   │
│ • Download video_3 → clip_3_with_audio.mp4   │
└──────┬───────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ FFmpeg Concatenation                         │
│ Input: [clip_1.mp4, clip_2.mp4, clip_3.mp4]  │
│ Command: ffmpeg -f concat -i list.txt output │
│ Output: final_stitched_TIMESTAMP.mp4         │
└──────┬───────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│ Export Metadata                              │
│ • scene_structure.xml                        │
│ • summary.json                               │
└──────────────────────────────────────────────┘
```

## Component Interactions

### Video Generation Flow

```
┌────────────────┐
│ VideoGenerator │
└────────┬───────┘
         │
         │ generateSequence(model, scenes, image, keyframes)
         │
         ▼
    ┌────────────────────────────────────────┐
    │ For each scene in scenes:              │
    │                                        │
    │  1. Determine input image              │
    │     Priority:                          │
    │     a) keyframes[i] (if provided)      │
    │     b) user image (if scene 0)         │
    │     c) last frame of previous video    │
    │                                        │
    │  2. Enhance prompt                     │
    │     • Add previous scene context       │
    │     • Convert image to data URI        │
    │                                        │
    │  3. Call Replicate API                 │
    │     • Create prediction                │
    │     • Poll for completion              │
    │     • Extract video URL                │
    │                                        │
    │  4. Extract last frame (if needed)     │
    │     FrameExtractor.getLastFrameFromURL │
    │     • Download video temporarily       │
    │     • FFmpeg extract frame             │
    │     • Save to temp/frame_N.jpg         │
    │                                        │
    └────────────────────────────────────────┘
```

### Voiceover Flow

```
┌────────────────┐
│ AudioGenerator │
└────────┬───────┘
         │
         │ addVoiceoversToSequence(results, scenes, voice)
         │
         ▼
    ┌────────────────────────────────────────┐
    │ For each result in results:            │
    │                                        │
    │  1. Generate Script (GPT-4o-mini)      │
    │     Input: scene.description + mood    │
    │     System: "Show don't tell" rules    │
    │     Output: 20-25 word ad copy         │
    │                                        │
    │  2. Extract Emotion                    │
    │     Map scene.structure.mood to:       │
    │     calm, happy, excited, etc.         │
    │                                        │
    │  3. Generate Audio (Minimax Speech)    │
    │     Input: script + voice + emotion    │
    │     Output: temp/voiceover_N.mp3       │
    │                                        │
    │  4. Merge Audio + Video                │
    │     Input: video URL + audio file      │
    │     Model: lucataco/video-audio-merge  │
    │     Output: final video URL            │
    │                                        │
    │  5. Download Final Video               │
    │     Save to output/clip_N_with_audio   │
    │                                        │
    └────────────────────────────────────────┘
```

## Configuration & Environment

```
.env
├── REPLICATE_API_TOKEN     [Required]  Video/audio generation
└── OPENAI_API_KEY          [Optional]  Content planning & ad copy

models.js
├── minimax: Hailuo 2.3     $0.28-0.56/clip, supports image
├── pixverse: PixVerse v5   ~$0.30/clip, supports image
├── seedance: Seedance Pro  Budget, supports image
├── veo: Google Veo 3.1     Premium, supports image
└── wan: Wan 2.5            Cheap, text-only

audioGenerator.js - Available Voices
├── Deep_Voice_Man          Male, authoritative
├── Calm_Woman              Female, soothing
├── Friendly_Person         Neutral, warm
└── [10+ more English voices]
```

## Error Handling & Retries

```
Video Generation Error Handling:
┌────────────────────────────────────┐
│ Prediction fails?                  │
│ → Log error                        │
│ → Continue to next clip            │
│ → Mark result with error flag      │
│ → Final output shows failed clips  │
└────────────────────────────────────┘

Frame Extraction Error Handling:
┌────────────────────────────────────┐
│ FFmpeg fails?                      │
│ → Log warning                      │
│ → Continue without image           │
│ → Next clip uses text-only         │
└────────────────────────────────────┘

Voiceover Error Handling:
┌────────────────────────────────────┐
│ TTS or merge fails?                │
│ → Log error                        │
│ → Return original silent video     │
│ → Mark with audioError flag        │
└────────────────────────────────────┘
```

## Performance Optimizations

1. **Parallel Processing**: Independent operations run concurrently
2. **Caching**: Keyframes cached for style consistency
3. **Streaming**: Progress updates streamed to user
4. **Cleanup**: Temporary files deleted after use (unless --keep-frames)
5. **Lossless Stitching**: FFmpeg concat demuxer (no re-encoding)

## Security Considerations

- API keys in `.env` (git-ignored)
- No hardcoded credentials
- Input sanitization for file paths
- Temporary file cleanup
- Rate limiting recommended for backend integration

---

**For detailed implementation, see source files in `src/` directory**
