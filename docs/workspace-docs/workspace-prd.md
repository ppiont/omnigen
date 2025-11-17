# Workspace Page - Product Requirements Document (PRD)

> **Project:** OmniGen AI Video Generation Platform  
> **Feature:** Video Workspace / Editor Page  
> **Version:** 1.0  
> **Last Updated:** November 16, 2025  
> **Status:** Draft

---

## Executive Summary

The Workspace page is the core editing and iteration interface where users interact with their generated videos. Users access this page after generating a new video from the Create page or selecting an existing video from the Library page. This page enables users to view their video with full playback controls and use an AI chat interface to request edits and regenerations.

**Key Goal:** Provide an intuitive, production-quality video editing workspace that allows users to iteratively refine their AI-generated videos through natural language prompts.

---

## 🎯 Scope Boundaries

**IN SCOPE - What We're Building:**

- ✅ `/frontend/src/pages/Workspace.jsx` - Main workspace page component
- ✅ New workspace-specific components (VideoPlayer, ChatInterface, VideoMetadata, ActionsToolbar)
- ✅ `/frontend/src/styles/workspace.css` - Workspace page styling
- ✅ Workspace page logic and state management

**OUT OF SCOPE - What We're NOT Touching:**

- ❌ Create page (`Create.jsx`) - Already exists
- ❌ Library page (`VideoLibrary.jsx`) - Already exists
- ❌ Dashboard page (`Dashboard.jsx`) - Already exists
- ❌ Shared components (Navbar, Sidebar, AppLayout, etc.) - Already exists
- ❌ Authentication flow or context (`AuthContext.jsx`) - Already exists
- ❌ Backend API implementation - Separate work stream
- ❌ Routing changes in `App.jsx` - Already configured

**Development Mode:**

- 🔧 UUID validation is **temporarily disabled** in `Workspace.jsx` for testing
- 🔧 Can access workspace via `/workspace/test-id` during development
- ✅ Will **re-enable UUID validation** before production deployment

---

## Problem Statement

After generating a video, users need a dedicated workspace to:

1. **Preview** their generated video with full playback controls
2. **Iterate** on the video through AI-powered edits via conversational prompts
3. **Understand** the generation status and parameters used

Currently, users can only generate videos but have no dedicated interface to review, iterate, and refine their content.

---

## User Journey

```
┌─────────────┐       ┌──────────────┐       ┌─────────────────┐
│   Create    │──────▶│  Generating  │──────▶│   Workspace     │
│    Page     │       │   (Status)   │       │  (This Page)    │
└─────────────┘       └──────────────┘       └─────────────────┘
       │                                              │
       │                                              │
       └──────────────────────────────────────────────┘
              User requests edit via AI chat
```

### Entry Points

1. **From Create Page:** User submits generation request → Backend creates job and returns `videoId` → Redirected to Workspace with `videoId`
2. **From Library Page:** User clicks on existing video card → Opens Workspace with `videoId`

**Note:** Both entry points use `videoId` (which is the same as `jobId` in the backend). The Create page waits for the backend to return the `videoId` before redirecting to the Workspace page.

### Exit Points

1. **Back to Library:** User clicks breadcrumb or back button
2. **New Generation:** User navigates to Create page for new video

---

## Core Features

### 1. Video Player (Priority: P0 - Critical)

**Description:** Full-featured video player with standard playback controls.

**Requirements:**

- ✅ Display generated video in 16:9 aspect ratio (default)
- ✅ Support other aspect ratios (9:16, 1:1) based on video metadata
- ✅ Playback controls:
  - Play/Pause toggle
  - Seek to beginning (restart)
  - Seek to end (skip to end)
  - Scrubber for precise timestamp navigation
  - Volume control
  - Fullscreen toggle
- ✅ Display current timestamp / total duration (e.g., "0:05 / 0:30")
- ✅ Loading state while video is fetching from S3
- ✅ Error state if video fails to load

**Acceptance Criteria:**

- Video plays smoothly without stuttering
- All controls are keyboard accessible (Space = play/pause, Arrow keys = seek)
- Video maintains aspect ratio on all screen sizes
- Scrubber shows preview thumbnail on hover (Future enhancement)

**Design Notes:**

- Player background: `--bg-elevated` (#0f1420)
- Controls use aurora accent colors (green for active states)
- Smooth transitions on hover (0.3s ease)

---

### 2. AI Chat Interface (Priority: P0 - Critical)

**Description:** Conversational interface for requesting video edits via natural language prompts.

⚠️ **MVP Limitation:** The backend does NOT have an edit endpoint. The chat interface will be a **placeholder** that guides users to create a new generation.

**Requirements:**

- ✅ Chat input box positioned below video player (right side or bottom panel)
- ✅ Message history showing user prompts and system responses
- ✅ Input field with placeholder: "Describe the changes you'd like to make..."
- ✅ Submit button (disabled in MVP, labeled "Edit (Coming Soon)")
- ⚠️ **MVP:** Input-only interface, no actual edit functionality
- 🔮 **Future:** When backend implements `/jobs/{id}/remix`, enable edit requests

**Acceptance Criteria:**

- User can type multi-line prompts (Shift+Enter for new line, Enter to submit)
- Chat shows informational message about creating new generation
- Submit button has tooltip: "Edit functionality coming in Phase 2"
- Link to "Create New Video" page pre-fills with modified prompt

**MVP User Flow:**

```
User: "Make the background more blue and add slow motion"
System: "✨ Want to try a different version?

        Edit functionality is coming soon! For now, you can:

        → Create a new generation with your changes

        [Create New Video with These Changes]"

[Button redirects to Create page with pre-filled prompt]
```

**Design Notes:**

- Chat container: `--bg-elevated` with `--border-glow`
- User messages: align right, `--aurora-green` accent
- System messages: align left, `--aurora-purple` accent
- Input box: `--bg-surface` with teal focus glow (disabled state)
- Info banner: aurora teal background with helpful guidance

**Phase 2 Enhancement:**

When backend implements remix endpoint:

```javascript
// POST /api/v1/jobs/{jobId}/remix
{
  "edit_prompt": "Make background more blue",
  "preserve_settings": true  // Keep original duration, style
}
```

---

### 3. Video Metadata Panel (Priority: P2 - Medium)

**Description:** Display video generation details and parameters.

**Requirements:**

- ✅ Video title (editable - client-side only in MVP)
- ✅ Generation timestamp ("Generated 2 hours ago")
- ✅ Video specifications:
  - Duration: "30 seconds" (from backend)
  - Aspect ratio: "16:9" (from backend)
  - Resolution: "1080p" (assumed/calculated)
  - Cost: "$2.10" (calculated: duration × $0.07)
- ✅ Original prompt used for generation (from backend)
- ✅ Generation model: "Kling v2.5 Turbo Pro" (hardcoded)
- ✅ Status badge: "Completed" / "Processing" / "Failed" (from backend)
- 🔮 **Future:** Tags, categories, custom metadata

**Acceptance Criteria:**

- All available metadata is fetched from backend API (`GET /api/v1/jobs/{jobId}`)
- Title is editable inline (localStorage only in MVP, no backend save)
- Status badge updates if polling detects status change
- Cost is calculated client-side: `duration × 0.07`
- Missing fields show placeholder or "N/A"

**Backend-Provided Fields:**

```json
{
  "job_id": "uuid",
  "status": "completed",
  "prompt": "Original prompt text",
  "duration": 30,
  "style": "luxury, minimal",
  "aspect_ratio": "16:9",
  "created_at": 1731744000,
  "completed_at": 1731747600
}
```

⚠️ **Note:** Values shown are examples. Actual API response contains real job data from the backend.

**Client-Side Calculated Fields:**

```javascript
// Cost estimation
const cost = (job.duration * 0.07).toFixed(2);

// Model (hardcoded)
const model = "Kling v2.5 Turbo Pro";

// Resolution (assumed)
const resolution = "1080p";

// Relative time
import { formatDistanceToNow } from "date-fns";
const timeAgo = formatDistanceToNow(job.created_at * 1000, { addSuffix: true });
```

**Design Notes:**

- Panel positioned on left sidebar or top bar
- Metadata uses `--text-secondary` color
- Status badges use functional colors (green=completed, orange=processing, red=failed)
- Editable title has subtle purple underline on hover
- Missing fields (style, aspect_ratio) show "Not specified"

---

### 4. Video Actions Toolbar (Priority: P1 - High)

**Description:** Quick actions for video management.

**Requirements:**

- ✅ Download button (downloads video file)
- ✅ Share button (generates shareable link - Future)
- ✅ Duplicate button (creates new generation with same prompt)
- ✅ Delete button (with confirmation modal)
- ⚠️ **MVP:** Download and Delete only
- 🔮 **Future:** Share, Export to different formats, Add to collection

**Acceptance Criteria:**

- Download triggers browser download of video file
- Delete shows confirmation modal: "Are you sure? This cannot be undone."
- All buttons have tooltips explaining their function
- Buttons disabled during video generation

**Design Notes:**

- Toolbar positioned in top-right corner
- Icon buttons with hover tooltips
- Primary action (Download): green accent
- Destructive action (Delete): red accent
- Secondary actions: purple/teal accents

---

## API Integration Points

### OmniGen Backend API (Actual Implementation)

**Base URL:** `/api/v1`  
**Authentication:** JWT Bearer tokens from AWS Cognito  
**Rate Limit:** 100 requests/minute per user

⚠️ **Important:** The backend uses a **job-centric** model where `jobId` and `videoId` are the same thing. There is no separate `/videos` endpoint.

---

#### Get Job Details (Video Status)

```http
GET /api/v1/jobs/{jobId}
Authorization: Bearer {jwt_token}
```

**Response:**

```json
{
  "job_id": "uuid",
  "status": "pending | processing | completed | failed",
  "prompt": "15 second luxury watch ad with gold aesthetics",
  "duration": 30,
  "style": "luxury, minimal, elegant",
  "video_url": "https://s3-presigned-url.com/...",
  "created_at": 1731744000,
  "completed_at": 1731747600,
  "error_message": null
}
```

⚠️ **Note:** The values above are **example/documentation only**. The actual API response will contain real data:

- `job_id` will be an actual UUID (e.g., `"abc-123-def-456"`)
- `prompt` will be the user's actual input text
- `video_url` will be a real S3 presigned URL
- Timestamps will be actual Unix timestamps from when the job was created/completed

**Field Details:**

- `video_url` - S3 presigned URL with **7-day expiration** (only present when status = "completed")
- `created_at` / `completed_at` - Unix timestamps (seconds)
- `error_message` - Human-readable error if status = "failed"

**Note:** Presigned URLs expire after 7 days. To refresh, call this endpoint again to get a new URL.

---

#### Create New Video Generation

```http
POST /api/v1/generate
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "prompt": "15 second luxury watch ad with gold aesthetics",
  "duration": 15,
  "aspect_ratio": "16:9",
  "style": "luxury, minimal, elegant"
}
```

**Request Validation:**

- `prompt` - Required, 10-1000 characters
- `duration` - Required, 15-180 seconds (integer)
- `aspect_ratio` - Optional, one of: "16:9", "9:16", "1:1" (defaults to "16:9")
- `style` - Optional, max 200 characters

**Response (201 Created):**

```json
{
  "job_id": "uuid",
  "status": "pending",
  "created_at": 1731744000,
  "estimated_completion": 150
}
```

⚠️ **Note:** Values shown are examples. Actual response contains real job IDs, timestamps, and completion estimates.

**Field Details:**

- `estimated_completion` - Estimated seconds until completion (duration × 10)
  - 15s video → ~150s (2.5 min)
  - 30s video → ~300s (5 min)
  - 60s video → ~600s (10 min)

---

#### List User Jobs

```http
GET /api/v1/jobs?page=1&page_size=20&status=completed
Authorization: Bearer {jwt_token}
```

**Query Parameters:**

- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20)
- `status` - Filter by status (optional)

**Response:**

```json
{
  "jobs": [
    {
      "job_id": "uuid",
      "status": "completed",
      "prompt": "...",
      "duration": 30,
      "video_url": "https://...",
      "created_at": 1731744000
    }
  ],
  "total_count": 42,
  "page": 1,
  "page_size": 20
}
```

⚠️ **Note:** Values shown are examples. Actual response contains real job data, counts, and pagination values.

**Implementation Note:** This endpoint is partially implemented (MVP stub). Pagination and filtering work, but may return empty results.

---

### Backend Implementation Details

#### Scene Planning (Automatic)

The backend **automatically** generates scene plans based on video duration:

**Short Videos (15-30s):** 3 scenes

- Scene 1 (30%): Product intro/reveal
- Scene 2 (40%): Product showcase
- Scene 3 (30%): Brand/CTA

**Long Videos (30-180s):** 4 scenes

- Scene 1 (20%): Hook/opening
- Scene 2 (25%): Product reveal
- Scene 3 (30%): Product showcase
- Scene 4 (25%): CTA/brand

⚠️ **Important:** Scene details are used internally by Step Functions but are **NOT returned** in the API response. The Workspace page does not display scene information in MVP.

---

#### Status Polling Strategy

Since the backend doesn't support WebSockets, use polling:

```javascript
// Poll every 5 seconds until status changes
const pollJobStatus = async (jobId) => {
  const interval = setInterval(async () => {
    const job = await fetch(`/api/v1/jobs/${jobId}`, {
      credentials: "include",
    }).then((r) => r.json());

    if (job.status === "completed" || job.status === "failed") {
      clearInterval(interval);
      // Update UI with final status
    }
  }, 5000); // 5 second interval

  // Stop polling after 10 minutes (timeout)
  setTimeout(() => clearInterval(interval), 600000);
};
```

---

#### Video URL Security

- **Method:** S3 presigned URLs
- **Expiration:** 7 days (604,800 seconds)
- **CORS:** Configured for `localhost:5173` (dev) and CloudFront domain (prod)
- **Refresh:** If URL expires, call `GET /api/v1/jobs/{jobId}` to get new presigned URL

**Example URL:**

```
https://omnigen-assets-123456789.s3.us-east-1.amazonaws.com/abc-123/final.mp4?
X-Amz-Algorithm=AWS4-HMAC-SHA256&
X-Amz-Credential=...&
X-Amz-Date=20251116T100000Z&
X-Amz-Expires=604800&
X-Amz-Signature=...
```

---

### Kling v2.5 Turbo Pro API (External - Backend Handles This)

**Model:** `kwaivgi/kling-v2.5-turbo-pro`  
**Provider:** Replicate  
**Cost:** $0.07 per second of output video

⚠️ **Note:** The Workspace page does **NOT** call Replicate directly. The backend handles all Replicate API calls via Step Functions. This information is for reference only.

**Pricing:** https://replicate.com/kwaivgi/kling-v2.5-turbo-pro/api

---

## User Stories

### Epic: Video Review and Iteration

**Story 1: View Generated Video**

- **As a** user
- **I want to** view my generated video with full playback controls
- **So that I can** review the AI-generated content before deciding on edits

**Story 2: Request AI-Powered Edits**

- **As a** user
- **I want to** describe changes in natural language via chat
- **So that I can** iterate on my video without learning complex editing tools

**Story 3: Download Final Video**

- **As a** user
- **I want to** download my video in high quality
- **So that I can** use it in my marketing campaigns

**Story 4: Track Generation Status**

- **As a** user
- **I want to** see real-time status updates during regeneration
- **So that I** know how long to wait and what's currently processing

**Story 5: Understand Video Metadata**

- **As a** user
- **I want to** see the original prompt, cost, and generation parameters
- **So that I can** understand what went into creating this video

---

## Success Metrics

### Primary Metrics (MVP)

- **User Engagement:** 80%+ of users who generate a video visit the Workspace page
- **Iteration Rate:** 40%+ of users request at least one AI edit
- **Session Duration:** Average 5+ minutes per Workspace session
- **Download Rate:** 60%+ of completed videos are downloaded

### Secondary Metrics (Future)

- **Edit Success Rate:** 85%+ of edit requests succeed without errors
- **Time to Edit:** Average 2-3 minutes per iteration cycle
- **Satisfaction Score:** 4+ stars on "How satisfied are you with the editing experience?"

---

## Technical Constraints

### Performance Requirements

- Video player loads within 2 seconds
- AI chat responds with status update within 5 seconds
- Page supports videos up to 500MB

### Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

### Responsive Breakpoints

- Desktop: 1280px+ (optimal experience)
- Tablet: 768px - 1279px (stacked layout)
- Mobile: 320px - 767px (simplified controls)

### Accessibility

- WCAG 2.1 AA compliance
- Keyboard navigation for all controls
- Screen reader support for video controls
- High contrast mode support

---

## Out of Scope (MVP)

The following features are explicitly **NOT** included in the initial release:

❌ **Timeline Component**

- Video timeline visualization
- Scene markers and segments
- Drag-and-drop scene reordering
- Trim/cut tools
- Transition effect selection
- Audio waveform visualization

❌ **Video Editing / Remix Functionality**

- AI-powered video editing
- Modify existing video via chat prompts
- Scene-specific regeneration
- Parameter adjustment (color, speed, etc.)
- **Reason:** Backend endpoint `/jobs/{id}/remix` not implemented yet

❌ **Real-time Collaboration**

- Multi-user editing sessions
- Comments and annotations
- Version history with rollback

❌ **Export Options**

- Custom resolution/bitrate export
- Format conversion (MOV, AVI, etc.)
- Platform-specific optimizations (TikTok, Instagram)

❌ **AI Voice Integration**

- AI voiceover generation (planned for Phase 2)
- Voice cloning
- Script-to-speech synchronization

❌ **Advanced Effects**

- Text overlays and captions
- Color grading presets
- Filters and LUTs
- Motion graphics templates

❌ **Video Thumbnails**

- Thumbnail generation
- Custom thumbnail selection
- **Reason:** Backend doesn't generate thumbnails yet

---

## Wireframe / Layout Description

```
┌────────────────────────────────────────────────────────────────────┐
│  [← Library] / Video Editor                    [Download] [Delete] │
├────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │                                                               │ │
│  │                    VIDEO PLAYER                               │ │
│  │                   (16:9 aspect ratio)                         │ │
│  │                                                               │ │
│  │                 [1920x1080 video preview]                     │ │
│  │                                                               │ │
│  │  [Play] [⏮] [⏭] ━━━━━━●━━━━━━━━━━━━━━ 0:15 / 0:30 [🔊] [⛶]   │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌─────────────────────────┐  ┌──────────────────────────────────┐ │
│  │  VIDEO METADATA         │  │  AI CHAT INTERFACE               │ │
│  │                         │  │                                  │ │
│  │  Title: [Editable]      │  │  ┌────────────────────────────┐ │ │
│  │  Duration: 30s          │  │  │ User: Make background blue │ │ │
│  │  Aspect: 16:9           │  │  └────────────────────────────┘ │ │
│  │  Model: Kling v2.5      │  │                                  │ │
│  │  Cost: $1.32            │  │  ┌────────────────────────────┐ │ │
│  │  Status: ✓ Completed    │  │  │ System: Edit coming soon... │ │ │
│  │                         │  │  └────────────────────────────┘ │ │
│  │  Original Prompt:       │  │                                  │ │
│  │  "A woman dancing..."   │  │  [Type your edit request...  ] │ │
│  │                         │  │              [Edit (Coming Soon)] │ │
│  └─────────────────────────┘  └──────────────────────────────────┘ │
│                                                                     │
└────────────────────────────────────────────────────────────────────┘
```

### Layout Grid (Desktop)

- **Top Bar:** Breadcrumbs (left), Actions (right) - Height: 64px
- **Video Player:** Full width, centered - Height: Auto (16:9 ratio)
- **Bottom Section:** 2-column grid
  - Left: Metadata Panel (30% width)
  - Right: AI Chat Interface (70% width)

### Responsive Behavior (Mobile)

- Video player: Full width
- Metadata: Collapsible accordion
- Chat: Full width, positioned at bottom

---

## Component File Structure

```
frontend/src/
├── pages/
│   └── Workspace.jsx                    ← MODIFY (main page)
├── components/
│   └── workspace/                       ← NEW FOLDER
│       ├── VideoPlayer.jsx             ← NEW (video playback)
│       ├── ChatInterface.jsx           ← NEW (AI chat)
│       ├── VideoMetadata.jsx           ← NEW (video info panel)
│       └── ActionsToolbar.jsx          ← NEW (download/delete)
├── styles/
│   └── workspace.css                    ← UPDATE (workspace styling)
└── utils/
    └── api.js                           ← EXISTING (reuse for API calls)

UNCHANGED (Do Not Touch):
├── pages/
│   ├── Create.jsx
│   ├── VideoLibrary.jsx
│   ├── Dashboard.jsx
│   ├── Login.jsx
│   ├── Signup.jsx
│   └── Settings.jsx
├── components/
│   ├── AppLayout.jsx
│   ├── Navbar.jsx
│   ├── Sidebar.jsx
│   ├── ProtectedRoute.jsx
│   └── ... (all other existing components)
└── contexts/
    └── AuthContext.jsx
```

---

## Design System References

### Color Palette (Aurora Theme)

- **Primary Action (Generate/Play):** `--aurora-green` (#7CFF00)
- **Secondary Action (Edit/Modify):** `--aurora-purple` (#B44CFF)
- **Info/Neutral (Status):** `--aurora-teal` (#00FFD1)
- **Warning (Processing):** `--aurora-orange` (#FFA500)
- **Error (Failed):** `--error` (#FF4D6A)

### Typography

- **Page Title:** 32px, weight 700, `--text-primary`
- **Section Headers:** 20px, weight 600, `--text-primary`
- **Body Text:** 16px, weight 400, `--text-secondary`
- **Metadata Labels:** 14px, weight 500, `--text-muted`

### Component Surfaces

- **Background:** `--bg` (#0a0e1a)
- **Elevated Panels:** `--bg-elevated` (#0f1420)
- **Hover State:** `--bg-highlight` (#1a1f33)
- **Border:** 1px solid `--bg-highlight`, glow on hover

---

## Dependencies

### Frontend Dependencies (Workspace Page Only)

- React 18 (existing)
- React Router (existing) - Route already configured in `App.jsx`
- Native HTML5 `<video>` element (no external video library needed for MVP)
- Fetch API for backend communication (already in `utils/api.js`)
- localStorage for title editing persistence (MVP)
- `date-fns` - For relative time formatting ("2 hours ago")
- WebSocket support for real-time status updates (Future)

### Backend Dependencies (Assumed Available)

- Go API with `/api/v1/jobs/*` endpoints ✅ **IMPLEMENTED**
- JWT authentication via Cognito ✅ **IMPLEMENTED**
- S3 presigned URLs (7-day expiration) ✅ **IMPLEMENTED**
- DynamoDB for job metadata ✅ **IMPLEMENTED**
- Rate limiting and quota enforcement ✅ **IMPLEMENTED**

### Extending `api.js`

The workspace page needs to add job-related functions to the existing `utils/api.js`:

```javascript
// Add to frontend/src/utils/api.js

/**
 * Jobs API endpoints
 */
export const jobs = {
  /**
   * Get a specific job by ID
   * @param {string} id - Job ID
   * @returns {Promise<Object>}
   */
  get: (id) => apiRequest(`/api/v1/jobs/${id}`),

  /**
   * List all jobs for the current user
   * @param {Object} params - Query parameters
   * @param {number} params.page - Page number
   * @param {number} params.page_size - Items per page
   * @param {string} params.status - Filter by status
   * @returns {Promise<Object>}
   */
  list: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return apiRequest(`/api/v1/jobs?${query}`);
  },
};

// Note: generate.create already exists for creating new jobs
```

**Note:** Workspace page will make API calls to backend endpoints, but does NOT implement backend logic.

---

## Risk Assessment

### High Risk

- ⚠️ **Video URL Expiration:** Presigned URLs expire after 7 days
  - **Mitigation:** Detect 403 errors, fetch new presigned URL from `GET /api/v1/jobs/{jobId}`, retry video load
- ⚠️ **Video Streaming Performance:** Large video files may be slow to load
  - **Mitigation:** Use CloudFront CDN, implement progressive loading, show buffering state

### Medium Risk

- ⚠️ **Backend API Availability:** If backend is down, workspace page is unusable
  - **Mitigation:** Show error page with retry button, detect 5xx errors, implement exponential backoff
- ⚠️ **State Synchronization:** Video status updates may be delayed during polling
  - **Mitigation:** Implement polling (every 5s), show loading indicators, handle stale data

### Low Risk

- ⚠️ **Browser Compatibility:** Older browsers may not support video formats
  - **Mitigation:** Detect browser capabilities, show warning for unsupported browsers

---

## Error Handling Matrix

Based on actual backend implementation, here are the error scenarios to handle:

| Error Type              | HTTP Status         | User Message                                     | Recovery Action                              |
| ----------------------- | ------------------- | ------------------------------------------------ | -------------------------------------------- |
| **Job Not Found**       | 404                 | "This video doesn't exist or has been deleted."  | "Return to Library" button                   |
| **Video URL Expired**   | 403 (from S3)       | "Video link expired. Refreshing..."              | Auto-fetch new presigned URL, retry          |
| **Video Load Failure**  | Video error event   | "Unable to load video. Please try again."        | "Refresh" button calls `GET /jobs/{id}`      |
| **Network Error**       | 0 (fetch fails)     | "Connection lost. Check your internet."          | Auto-retry after 3s, manual retry button     |
| **Session Expired**     | 401                 | "Your session has expired. Please log in again." | Redirect to login page                       |
| **Rate Limit Exceeded** | 429                 | "Too many requests. Please wait a moment."       | Show countdown timer, retry after delay      |
| **Generation Failed**   | 200 (status=failed) | "Video generation failed: {error_message}"       | Show error from backend, "Try Again" button  |
| **Processing Timeout**  | N/A (polling)       | "Generation is taking longer than expected..."   | Continue polling, show extended wait message |

### Error Response Format (from backend)

```json
{
  "error": {
    "message": "Human-readable error message",
    "code": "ERROR_CODE_CONSTANT",
    "details": {
      "field": "Additional context"
    }
  }
}
```

### Implementation Example

```javascript
async function fetchVideo(jobId) {
  try {
    const response = await fetch(`/api/v1/jobs/${jobId}`, {
      credentials: "include",
    });

    if (!response.ok) {
      if (response.status === 404) {
        showError("This video doesn't exist or has been deleted.");
        navigateToLibrary();
      } else if (response.status === 401) {
        showError("Your session has expired. Please log in again.");
        redirectToLogin();
      } else if (response.status === 429) {
        const retryAfter = response.headers.get("Retry-After") || 60;
        showError(`Too many requests. Please wait ${retryAfter} seconds.`);
        setTimeout(() => fetchVideo(jobId), retryAfter * 1000);
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
      return;
    }

    const job = await response.json();

    // Handle job-level errors
    if (job.status === "failed") {
      showError(
        `Video generation failed: ${job.error_message || "Unknown error"}`
      );
    }

    return job;
  } catch (error) {
    // Network error
    showError("Connection lost. Check your internet.");
    // Auto-retry after 3s
    setTimeout(() => fetchVideo(jobId), 3000);
  }
}
```

---

## Open Questions

1. **Chat History Persistence:** Should chat history persist across sessions or only during current session?

   - **Recommendation:** Store in localStorage for MVP, move to backend in Phase 2

2. **Concurrent Edit Requests:** Can user submit multiple edit requests simultaneously?

   - **Recommendation:** MVP allows one active job at a time, queue additional requests

3. **Cost Transparency:** Should we show estimated cost before submitting edit?

   - **Recommendation:** Yes, always show estimate with confirmation modal for >$5 requests

4. **Video Versioning:** Should we keep previous versions when user requests edit?

   - **Recommendation:** MVP creates new video, Phase 2 implements version history

5. **Partial Regeneration:** Can we regenerate specific scenes without regenerating entire video?
   - **Recommendation:** Phase 2 feature, MVP regenerates full video

---

## Implementation Phases

### Phase 1: MVP (Current - Workspace Page Only)

**Files to Create/Modify:**

- ✅ `/frontend/src/pages/Workspace.jsx` - Main page component (modify existing)
- ✅ `/frontend/src/components/workspace/VideoPlayer.jsx` - NEW component
- ✅ `/frontend/src/components/workspace/ChatInterface.jsx` - NEW component (placeholder)
- ✅ `/frontend/src/components/workspace/VideoMetadata.jsx` - NEW component
- ✅ `/frontend/src/components/workspace/ActionsToolbar.jsx` - NEW component
- ✅ `/frontend/src/styles/workspace.css` - Update existing styles
- ✅ `/frontend/src/utils/api.js` - Add `jobs.get()` and `jobs.list()` functions

**What We're Building:**

- Basic video player with HTML5 controls
- AI chat interface placeholder (informational messages, no edit functionality)
- Video metadata display panel (with client-side calculated fields)
- Download/Delete action buttons
- Status polling for in-progress videos
- UUID validation (re-enabled before deployment)

**Backend Integration:**

- Call `GET /api/v1/jobs/{jobId}` to fetch video data
- Use presigned `video_url` from response (7-day expiration)
- Poll job status every 5s if status is "processing"
- Handle "pending", "processing", "completed", "failed" states

**What We're NOT Building in Phase 1:**

- Timeline component (removed from MVP scope)
- Real video editing (backend endpoint not available)
- Video thumbnails (backend doesn't generate them)
- Persistent title storage (localStorage only)
- Chat history persistence across sessions

**Timeline:** 2 weeks  
**Team:** 1 Frontend Developer (working only on Workspace page)

---

### Phase 2: Enhanced Editing (Q1 2026)

**Backend Requirements (must be implemented first):**

- `POST /api/v1/jobs/{jobId}/remix` - Create variation of existing video
- Scene data in job API response
- Thumbnail generation
- Title/metadata update endpoint

**Frontend Enhancements:**

- Scene-specific regeneration via chat
- Timeline component with scene markers
- Video versioning UI
- Streaming AI chat responses
- Export options dialog

**Timeline:** 4 weeks  
**Team:** 1 Frontend + 1 Backend Developer

---

### Phase 3: Advanced Features (Q2 2026)

- Drag-and-drop timeline editing
- Text overlays and effects
- Real-time collaboration
- Advanced export settings
- Video thumbnail management

**Timeline:** 6 weeks  
**Team:** 2 Frontend + 1 Backend Developer

---

## Approval & Sign-off

| Role             | Name   | Date   | Status     |
| ---------------- | ------ | ------ | ---------- |
| Product Owner    | [Name] | [Date] | ⏳ Pending |
| Engineering Lead | [Name] | [Date] | ⏳ Pending |
| Design Lead      | [Name] | [Date] | ⏳ Pending |

---

## Appendix

### Glossary

- **Job:** Background task for video generation (tracked via Step Functions)
- **Scene:** Individual video segment (5-10 seconds) within a multi-scene video
- **Edit:** User-requested modification to existing video via AI chat

### References

- [OmniGen Design Spec](../existing_infra_saturday/omnigen_design_spec.md)
- [Aurora Color Palette](../existing_infra_saturday/omnigen_color_palette.md)
- [Video Generation Workflow](../existing_infra_saturday/video-workflow.md)
- [Kling v2.5 Turbo Pro API](https://replicate.com/kwaivgi/kling-v2.5-turbo-pro)

---

**End of PRD**
