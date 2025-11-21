# OmniGen Backend

AI-powered video generation pipeline with intelligent script generation and multi-adapter support.

## Video Generation Adapters

OmniGen supports multiple video generation adapters. Select your adapter using the `VIDEO_ADAPTER_TYPE` environment variable.

### Google Veo 3.1 (Default)

- **Duration**: 4s, 6s, or 8s clips
- **Audio**: Optional native context-aware audio generation (disabled by default)
- **Resolution**: 720p or 1080p
- **Advanced Features**: Reference images (R2V), last frame interpolation
- **Cost**: $0.15/second with audio, $0.10/second without
- **Best for**: High-quality video generation with optional audio

```bash
# Veo is the default - no environment variable needed
# Or explicitly set it:
export VIDEO_ADAPTER_TYPE=veo

# Enable audio generation (optional):
export VEO_GENERATE_AUDIO=true
```

**Note:** By default, Veo runs **without audio generation** and uses Minimax for separate audio. This gives you the same pipeline as Kling but with higher quality video output.

### Kling v2.5 Turbo Pro

- **Duration**: 5s or 10s clips
- **Audio**: No native audio (uses Minimax for separate audio generation)
- **Cost**: $0.07/second
- **Best for**: Lower-cost video generation

```bash
export VIDEO_ADAPTER_TYPE=kling
```

## Configuration

### Environment Variables

See `.env.example` for a complete list of configuration options.

#### Adapter Selection

```bash
# Use Veo 3.1 (default - no setting needed)
# VIDEO_ADAPTER_TYPE=veo

# Use Kling
VIDEO_ADAPTER_TYPE=kling
```

#### Veo-Specific Options (optional)

```bash
VEO_RESOLUTION=1080p         # Options: 720p, 1080p
VEO_GENERATE_AUDIO=true      # Enable native audio generation (disabled by default)
```

**Audio Generation Behavior:**
- **Veo without `VEO_GENERATE_AUDIO=true`**: Uses Minimax for separate audio (same as Kling)
- **Veo with `VEO_GENERATE_AUDIO=true`**: Uses native Veo audio generation (skips Minimax)
- **Kling**: Always uses Minimax for audio (no native audio support)

## Development

### Prerequisites

- Go 1.21+
- FFmpeg (required for video composition)
- AWS credentials configured
- Replicate API key (stored in AWS Secrets Manager)

### Building

```bash
# Build all packages
go build ./...

# Build the main API binary
go build ./cmd/api

# Run the API server
./api
```

### Running Locally

```bash
# Set environment variables
export ENVIRONMENT=development
export PORT=8080
export AWS_REGION=us-east-1
# ... other required env vars from .env.example

# Run the server
go run ./cmd/api
```

## Architecture

### Pipeline Flow

1. **Script Generation**: GPT-4o analyzes the user prompt and generates a structured script with scenes
2. **Video Generation**: Selected adapter (Kling or Veo) generates video clips for each scene
3. **Audio Generation**:
   - **Veo**: Native audio embedded in video clips
   - **Kling**: Separate Minimax audio generation
4. **Composition**: FFmpeg stitches clips and audio into final video
5. **Storage**: Final video uploaded to S3 with presigned URLs

### Adapter Pattern

The video generation system uses an adapter pattern to support multiple AI video models:

```go
type VideoGeneratorAdapter interface {
    GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error)
    GetStatus(ctx context.Context, predictionID string) (*VideoGenerationResult, error)
    GetModelName() string
    GetCostPerSecond() float64
}
```

Adding new adapters is straightforward:
1. Implement the `VideoGeneratorAdapter` interface
2. Add to the `AdapterFactory`
3. Add to environment configuration

## Future Enhancements

Veo 3.1 supports advanced features not yet implemented:

- **Reference Images (R2V)**: Subject-consistent generation across clips
- **Last Frame Interpolation**: Smooth transitions between scenes
- **Dynamic Resolution**: Per-request resolution control
- **Cost Optimization**: Selective audio generation based on requirements

These features can be added as needed without breaking changes.

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/adapters
go test ./internal/api/handlers

# Run with coverage
go test -cover ./...
```

## Deployment

See `/infrastructure/README.md` for deployment instructions and infrastructure setup.

## Cost Analysis

### Kling Pipeline (30s video)
- Video: 3 clips × 10s × $0.07/s = $2.10
- Audio: Minimax ~$0.50
- **Total: ~$2.60**

### Veo Pipeline (30s video)
- Video with Audio: 4 clips × 8s × $0.15/s = $4.80
- Audio: Included in video generation
- **Total: ~$4.80**

*Note: Veo is more expensive per second but includes audio and may provide better quality.*

## Troubleshooting

### Video Generation Fails

1. Check adapter logs for specific error messages
2. Verify Replicate API key is valid
3. Check AWS Secrets Manager configuration
4. Review rate limits for selected adapter

### FFmpeg Errors

1. Ensure FFmpeg is installed: `ffmpeg -version`
2. Check temporary directory permissions
3. Review FFmpeg logs in application output

### Audio Not Playing

- **Kling**: Check Minimax API status and logs
- **Veo**: Verify `VEO_GENERATE_AUDIO=true` is set
- Check browser console for playback errors

## License

Proprietary - OmniGen Project
