# Jobs API Enhancement

## Overview

Enhanced the Jobs API with status filtering and structured progress tracking for better user experience and frontend integration.

## Changes Made

### 1. Domain Model Updates (`internal/domain/job.go`)

Added structured progress fields to the `Job` struct:

```go
// Progress fields (NEW - structured for better API responses)
ThumbnailURL    string   `dynamodbav:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`       // Preview thumbnail from first scene
AudioURL        string   `dynamodbav:"audio_url,omitempty" json:"audio_url,omitempty"`               // Generated audio track URL
ScenesCompleted int      `dynamodbav:"scenes_completed,omitempty" json:"scenes_completed,omitempty"` // Number of completed scenes
SceneVideoURLs  []string `dynamodbav:"scene_video_urls,omitempty" json:"scene_video_urls,omitempty"` // Individual scene video URLs
```

**Benefits:**
- Frontend can display thumbnails while video is processing
- Show individual scene completion status
- Provide preview of audio before final composite
- Better progress visualization for users

### 2. Repository Layer Updates

**File:** `internal/repository/interfaces.go`
- Updated `GetJobsByUser` signature to accept optional `status` parameter

**File:** `internal/repository/dynamodb.go`
- Implemented status filtering using DynamoDB FilterExpression
- Added conditional logic to only apply filter when status is provided

```go
// Add status filter if provided
if status != "" {
    input.FilterExpression = aws.String("#status = :status")
    input.ExpressionAttributeNames = map[string]string{
        "#status": "status",
    }
    input.ExpressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: status}
}
```

### 3. Handler Layer Updates

**File:** `internal/api/handlers/jobs.go`

#### Updated `JobResponse` struct:
```go
type JobResponse struct {
    // ... existing fields ...

    // Progress fields (NEW)
    ThumbnailURL    string   `json:"thumbnail_url,omitempty"`
    AudioURL        string   `json:"audio_url,omitempty"`
    ScenesCompleted int      `json:"scenes_completed,omitempty"`
    SceneVideoURLs  []string `json:"scene_video_urls,omitempty"`
}
```

#### Updated `ListJobs` handler:
- Added `status` query parameter support
- Updated logging to include status filter
- Pass status filter to repository

#### Updated response mapping:
- Both `GetJob` and `ListJobs` now include progress fields in responses
- Added `ProgressPercent` calculation to `ListJobs`

## API Usage

### List All Jobs
```bash
GET /api/v1/jobs?page_size=20
```

### Filter by Status
```bash
GET /api/v1/jobs?status=completed
GET /api/v1/jobs?status=processing
GET /api/v1/jobs?status=failed
GET /api/v1/jobs?status=pending
```

### Pagination
```bash
GET /api/v1/jobs?page_size=50&status=completed
```

## Response Example

```json
{
  "jobs": [
    {
      "job_id": "job_123",
      "status": "processing",
      "stage": "scene_2_complete",
      "progress_percent": 66,
      "thumbnail_url": "https://s3.../thumbnail.jpg",
      "audio_url": "https://s3.../audio.mp3",
      "scenes_completed": 2,
      "scene_video_urls": [
        "https://s3.../scene_1.mp4",
        "https://s3.../scene_2.mp4"
      ],
      "prompt": "Create a 30s ad for...",
      "duration": 30,
      "created_at": 1700000000,
      "updated_at": 1700000123
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 20
}
```

## Frontend Integration

The enhanced API enables:

1. **Real-time Progress Visualization**
   - Show thumbnail preview immediately when first scene completes
   - Display scene-by-scene completion status
   - Calculate and display progress percentage

2. **Better UX During Processing**
   - Users can preview individual scenes as they complete
   - Audio preview available before final composite
   - Clear indication of what stage video generation is in

3. **Status-Based Views**
   - "Completed Videos" tab: `?status=completed`
   - "Processing" tab: `?status=processing`
   - "All Jobs" view: no status filter

4. **Pagination Support**
   - Control page size (1-100 items)
   - Default: 20 items per page

## DynamoDB Considerations

**Query Pattern:**
- Primary query: `UserJobsIndex` (user_id + created_at)
- Filter applied client-side by DynamoDB after query
- No additional GSI required (status is a filter, not part of key)

**Performance:**
- Filter is applied AFTER reading items from index
- May read more items than `limit` if many are filtered out
- For production with high volume, consider dedicated GSI for status queries
- Current implementation is optimal for MVP (<1000 jobs per user)

**Cost:**
- No additional read capacity units for filtering
- Same RCU consumption as non-filtered query

## Testing

Build verification:
```bash
✓ go build -v ./...
```

## Future Enhancements

1. **Cursor-based Pagination**
   - Add `LastEvaluatedKey` support for large result sets

2. **Multiple Status Filters**
   - Support `?status=completed,processing` (comma-separated)

3. **Date Range Filtering**
   - Add `created_after` and `created_before` query params

4. **Search/Filter by Prompt**
   - Add text search capability (requires ElasticSearch/CloudSearch)

5. **Sort Options**
   - Allow sorting by different fields (currently: created_at DESC only)
