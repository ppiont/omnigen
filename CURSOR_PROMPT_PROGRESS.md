# Implement Real-Time Video Generation Progress on Create Page

## Overview
Add real-time progress tracking to the Create page using the new backend SSE (Server-Sent Events) progress API at `/api/v1/jobs/:id/progress`.

**Note:** This is different from the existing `/api/v1/jobs` endpoint documented in `backend/JOBS_API_ENHANCEMENT.md`. The SSE progress endpoint provides real-time streaming updates, while the jobs endpoint is for polling/listing jobs.

## Backend API Information

### Endpoint
- **URL**: `GET /api/v1/jobs/:id/progress`
- **Type**: Server-Sent Events (SSE) stream
- **Authentication**: Already handled (mock user in development)

### SSE Events
The endpoint sends three types of events:

1. **`update`** - Sent whenever job stage changes
   - Payload: `ProgressResponse` JSON (see below)
   - Sent every time the stage changes (not spammy)

2. **`done`** - Sent when job completes or fails
   - Payload: `{ "status": "completed" | "failed" }`
   - Stream closes after this event

3. **`error`** - Sent if there's an error fetching job
   - Payload: `{ "error": "error message" }`
   - Stream continues after error

### ProgressResponse Type
```typescript
interface ProgressResponse {
  job_id: string;
  status: string; // "pending" | "processing" | "completed" | "failed"
  progress: number; // 0-100 percentage
  current_stage: string; // Internal stage name (e.g., "scene_2_generating")
  current_stage_display: string; // User-friendly name (e.g., "Generating scene 2")
  stages_completed: StageInfo[];
  stages_pending: StageInfo[];
  estimated_time_remaining: number; // Seconds
  assets?: ProgressAssets; // Optional, contains presigned URLs
}

interface StageInfo {
  name: string;
  display_name: string;
  progress: number; // 0-100
  completed_at?: number; // Unix timestamp (optional)
}

interface ProgressAssets {
  scene_clips: AssetInfo[];
  thumbnails: AssetInfo[];
  audio?: AssetInfo;
  final_video?: AssetInfo;
}

interface AssetInfo {
  url: string;
  scene_number?: number;
  created_at: number; // Unix timestamp
}
```

### Pipeline Stages (in order)
1. **Script Generation** (15% progress)
   - `script_generating` → "Generating script with AI"
   - `script_complete` → "Script ready"

2. **Scene Generation** (15-90% progress, varies by number of scenes)
   - `scene_1_generating` → "Generating scene 1"
   - `scene_1_complete` → "Scene 1 ready"
   - `scene_2_generating` → "Generating scene 2"
   - ... (continues for all scenes)

3. **Audio Generation** (95% progress)
   - `audio_generating` → "Generating background music"
   - `audio_complete` → "Audio ready"

4. **Video Composition** (98% progress)
   - `composing` → "Composing final video"

5. **Complete** (100% progress)
   - `complete` → "Complete"

## Requirements

### 1. Update Create Page to Show Progress

**When user clicks "Generate Video":**
1. Call the existing generate API (`POST /api/v1/generate`)
2. Get the `job_id` from the response
3. Open SSE connection to `/api/v1/jobs/{job_id}/progress`
4. Show a progress modal/overlay with:
   - Overall progress bar (0-100%)
   - Current stage display text
   - Estimated time remaining (formatted as "X min Y sec")
   - List of completed stages (checkmarks)
   - Current stage (spinner/loading indicator)
   - Pending stages (grayed out)

**Progress UI Design:**
- Use a full-screen modal overlay with semi-transparent dark background
- Center the progress card
- Use Aurora design system colors:
  - Primary: `--aurora-purple` (#b44cff)
  - Success: `--aurora-green` (#7cff00)
  - Background: `--bg-elevated` (#0f1420)
  - Text: `--text-primary` (#e8edf5)

**Progress Card Layout:**
```
┌────────────────────────────────────────┐
│  Generating Your Video                 │
│                                        │
│  ████████████░░░░░░░░░░░░  65%        │
│                                        │
│  Generating scene 3                    │
│  Est. 2 min 30 sec remaining           │
│                                        │
│  ✓ Script ready                        │
│  ✓ Scene 1 ready                       │
│  ✓ Scene 2 ready                       │
│  ⟳ Generating scene 3                  │
│  ⋯ Scene 4                             │
│  ⋯ Scene 5                             │
│  ⋯ Background music                    │
│  ⋯ Final composition                   │
│                                        │
│  [Cancel]                              │
└────────────────────────────────────────┘
```

### 2. SSE Connection Management

**Create a new utility:** `frontend/src/utils/sse.js`
```javascript
/**
 * Connect to SSE progress endpoint
 * @param {string} jobId - The job ID to track
 * @param {Object} callbacks - Event callbacks
 * @param {Function} callbacks.onUpdate - Called with ProgressResponse on update
 * @param {Function} callbacks.onDone - Called with status on completion
 * @param {Function} callbacks.onError - Called with error message
 * @returns {Function} cleanup function to close connection
 */
export function connectToProgress(jobId, { onUpdate, onDone, onError }) {
  const eventSource = new EventSource(
    `http://localhost:8080/api/v1/jobs/${jobId}/progress`,
    { withCredentials: true }
  );

  eventSource.addEventListener('update', (event) => {
    const data = JSON.parse(event.data);
    onUpdate(data);
  });

  eventSource.addEventListener('done', (event) => {
    const data = JSON.parse(event.data);
    onDone(data.status);
    eventSource.close();
  });

  eventSource.addEventListener('error', (event) => {
    if (event.data) {
      const data = JSON.parse(event.data);
      onError(data.error);
    } else {
      // Connection error
      onError('Connection lost');
      eventSource.close();
    }
  });

  eventSource.onerror = (error) => {
    console.error('SSE connection error:', error);
    onError('Failed to connect to progress stream');
    eventSource.close();
  };

  // Return cleanup function
  return () => {
    eventSource.close();
  };
}
```

### 3. Create Progress Modal Component

**Create:** `frontend/src/components/create/ProgressModal.jsx`

**Props:**
- `jobId` - The job ID being tracked
- `onComplete` - Callback when video generation completes (receives final job data)
- `onCancel` - Callback to cancel and close modal
- `isOpen` - Boolean to control visibility

**Features:**
- Connect to SSE on mount
- Update UI in real-time as events arrive
- Show success state when `done` event received with status="completed"
- Show error state when `done` event received with status="failed"
- Navigate to workspace or video library on completion
- Clean up SSE connection on unmount

### 4. Update Create.jsx

**Modifications needed:**
1. Import and use `ProgressModal` component
2. Add state for tracking active job: `const [generatingJobId, setGeneratingJobId] = useState(null)`
3. After successful generate API call:
   ```javascript
   const response = await api.generate(payload);
   setGeneratingJobId(response.job_id);
   ```
4. Render ProgressModal when `generatingJobId` is set
5. Handle completion:
   - On success: Navigate to `/workspace/${jobId}` or show success toast
   - On failure: Show error toast and allow user to try again
   - On cancel: Close modal and reset state

### 5. Styling

**Create:** `frontend/src/styles/progress-modal.css`

Use Aurora design system:
- Smooth animations for progress bar
- Fade-in transitions for stage updates
- Pulsing animation for current stage spinner
- Scale animation when stages complete (checkmark appears)

**Progress bar animation:**
```css
.progress-bar {
  transition: width 0.5s ease-in-out;
}

.stage-item {
  transition: all 0.3s ease;
}

.stage-complete {
  animation: checkmark-pop 0.3s ease;
}

@keyframes checkmark-pop {
  0% { transform: scale(0); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}
```

### 6. Error Handling

**Handle these scenarios:**
1. **SSE connection fails**: Show error message, offer retry button
2. **Job fails**: Show error from backend, offer "Try Again" button
3. **Network interruption**: Auto-retry connection (3 attempts with exponential backoff)
4. **User navigates away**: Clean up SSE connection properly

### 7. Time Formatting Utility

**Add to:** `frontend/src/utils/format.js`
```javascript
/**
 * Format seconds into human-readable time
 * @param {number} seconds
 * @returns {string} e.g., "2 min 30 sec", "45 sec", "1 min"
 */
export function formatTimeRemaining(seconds) {
  if (seconds < 60) {
    return `${seconds} sec`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  if (remainingSeconds === 0) {
    return `${minutes} min`;
  }
  return `${minutes} min ${remainingSeconds} sec`;
}
```

## Testing Checklist

- [ ] Progress modal opens after clicking "Generate Video"
- [ ] Progress bar updates smoothly as stages complete
- [ ] Stage list shows completed (✓), current (spinner), and pending (⋯) states
- [ ] Estimated time remaining updates and is formatted correctly
- [ ] Modal closes and navigates to workspace on completion
- [ ] Error state shown if job fails
- [ ] Cancel button closes modal and stops tracking
- [ ] SSE connection cleans up properly on unmount
- [ ] Works with different numbers of scenes (test with 15s, 30s, 60s videos)
- [ ] Progress percentages are accurate and smooth

## Design Reference

Use the existing Aurora design patterns from:
- `frontend/src/styles/aurora.css` - Color variables and effects
- `frontend/src/styles/create.css` - Existing Create page styling
- `frontend/src/components/create/MediaUploadBar.jsx` - Component structure reference

Keep the UI consistent with the rest of the app's dark theme and purple accent colors.

## Additional Notes

- The SSE connection sends updates **only when stage changes**, not continuously, so it's very efficient
- Progress percentages are dynamically calculated based on number of scenes
- The `assets` field in ProgressResponse contains presigned S3 URLs that are valid for 1 hour
- You can optionally show thumbnails/previews of completed scenes using the assets data
- The backend handles all progress calculation - frontend just displays it

## Implementation Order

1. Create SSE utility (`utils/sse.js`)
2. Create time formatting utility (`utils/format.js`)
3. Create ProgressModal component
4. Create progress modal styles
5. Update Create.jsx to use ProgressModal
6. Test with real video generation
7. Polish animations and error handling
