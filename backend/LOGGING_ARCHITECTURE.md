# Logging Architecture

## Correlation Strategy

### job_id as Primary Correlation ID

All video generation operations are correlated using **job_id**:

```
POST /api/v1/generate
  ↓
job_id: "job-{uuid}" created
  ↓
All subsequent logs include job_id:
  - Script generation (GPT-4o)
  - Scene generation (Kling AI, parallel)
  - Audio generation (Minimax)
  - Video composition (ffmpeg)
  - Job completion
```

### Log Correlation Flow

**Example trace for a single video generation:**

```
INFO  Request started                  method=POST path=/api/v1/generate
INFO  Starting fully async generation  job_id=job-abc123 user_id=user-xyz
INFO  Generating script with GPT-4o    job_id=job-abc123
INFO  Script generated successfully    job_id=job-abc123 script_id=script-def456
INFO  Scene details                    job_id=job-abc123 scene_number=1 duration=5
INFO  Scene details                    job_id=job-abc123 scene_number=2 duration=5
INFO  Calling Kling adapter            job_id=job-abc123 scene=1
INFO  Calling Kling adapter            job_id=job-abc123 scene=2
INFO  Scene 1 generation started       job_id=job-abc123 prediction_id=kling-1
INFO  Scene 2 generation started       job_id=job-abc123 prediction_id=kling-2
DEBUG Kling still processing           job_id=job-abc123 scene=1 attempt=12
DEBUG Kling still processing           job_id=job-abc123 scene=2 attempt=12
INFO  Scene 1 video generated          job_id=job-abc123 video_url=s3://...
INFO  Scene 2 video generated          job_id=job-abc123 video_url=s3://...
INFO  Generating audio                 job_id=job-abc123
INFO  Audio generated                  job_id=job-abc123 audio_url=s3://...
INFO  Composing final video            job_id=job-abc123
INFO  Video generation complete        job_id=job-abc123 video_key=videos/final.mp4
```

### What We Don't Log

**Filtered out to reduce noise:**
- ❌ `GET /health` - ECS/ALB health checks (every 30s)
- ❌ `GET /api/v1/jobs/{id}` - Job status polling (every 5-10s)
- ❌ "User authenticated" - Auth middleware success (every request)

**Only critical auth events are logged:**
- ✅ Invalid/expired JWT tokens
- ✅ Missing authorization headers
- ✅ Token verification failures

### HTTP Request Logging

**Logged for all non-filtered endpoints:**

```go
// Request start
INFO Request started
  method=POST
  path=/api/v1/generate
  client_ip=192.168.1.100

// Request completion
INFO Request completed
  method=POST
  path=/api/v1/generate
  status=202
  latency=15ms
  response_size=124
```

**No trace_id** - we use job_id for video generation correlation instead.

## Log Levels

### INFO
- Request start/completion
- Job lifecycle events (created, processing, completed, failed)
- Major pipeline stages (script generation, video generation, composition)
- External API calls (Kling, Minimax, GPT-4o)
- S3 uploads/downloads

### DEBUG
- Polling loops (reduced to every 60s instead of every 5s)
- Detailed scene parameters
- ffmpeg operations

### WARN
- Retries
- Timeouts (non-fatal)
- Scene timing mismatches

### ERROR
- Job failures
- API errors
- Database errors
- File operation failures
- Panic recovery

## Log Structure

All logs use **structured logging** (zap):

```go
logger.Info("Video generation complete",
    zap.String("job_id", jobID),
    zap.String("video_key", videoKey),
    zap.Int("num_scenes", len(scenes)),
    zap.Duration("total_time", elapsed),
)
```

**Benefits:**
- Machine-parseable (JSON format in production)
- Easy filtering by job_id, user_id, etc.
- CloudWatch Logs Insights queries
- No string concatenation

## CloudWatch Logs Insights Queries

### Trace entire job lifecycle:
```
fields @timestamp, @message, job_id, stage, status
| filter job_id = "job-abc123"
| sort @timestamp asc
```

### Find failed jobs:
```
fields @timestamp, job_id, error_message
| filter status = "failed"
| sort @timestamp desc
| limit 20
```

### Video generation performance:
```
fields job_id, duration, total_time
| filter @message like /Video generation complete/
| stats avg(total_time) as avg_time, max(total_time) as max_time by duration
```

### External API errors:
```
fields @timestamp, job_id, @message
| filter @message like /API error/ or @message like /failed/
| filter job_id != ""
| sort @timestamp desc
```

## Related Files

- **Middleware:** `internal/api/middleware/logging.go` - HTTP request logging
- **Auth Middleware:** `internal/auth/middleware.go` - JWT validation (no success logs)
- **Generate Handler:** `internal/api/handlers/generate_async.go` - All video generation logs with job_id
- **Adapters:** `internal/adapters/*_adapter.go` - External API calls (no job_id context needed - wrapped by handler)
- **Services:** `internal/service/*.go` - Business logic (minimal logging, handler logs boundaries)

## Best Practices

1. **Always log job_id** for video generation operations
2. **Use structured fields** instead of string formatting
3. **Log at boundaries** (before/after external calls, major stages)
4. **Don't log in tight loops** (reduced polling logs to 1/12th frequency)
5. **Filter noisy endpoints** at middleware level
6. **Use appropriate log levels** (DEBUG for details, INFO for flow, ERROR for failures)
7. **No sensitive data** in logs (user emails OK, passwords/tokens NOT OK)
