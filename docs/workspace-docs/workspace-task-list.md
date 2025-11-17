# Workspace Page - Task List

> **Project:** OmniGen AI Video Generation Platform  
> **Feature:** Video Workspace / Editor Page  
> **Version:** 1.0  
> **Last Updated:** November 16, 2025  
> **Status:** Ready for Implementation

---

## Table of Contents

1. [Overview](#overview)
2. [Tasks](#tasks)
3. [Task Dependencies](#task-dependencies)
4. [Estimation Summary](#estimation-summary)

---

## Overview

### Project Scope

**What We're Building:**

- ✅ Workspace.jsx (main page)
- ✅ 4 new components: VideoPlayer, ChatInterface, VideoMetadata, ActionsToolbar
- ✅ Workspace-specific styling
- ✅ API integration with backend

**What We're NOT Building:**

- ❌ Timeline component (out of scope)
- ❌ Backend API changes
- ❌ Changes to existing pages (Create, Library, Dashboard)
- ❌ Changes to shared components (Navbar, Sidebar, AppLayout)

### Team

- **Frontend Developer:** 1 person
- **Timeline:** 2 weeks (10 working days)
- **Sprint:** Single sprint, MVP focus

---

## Tasks

### Task #1: Install Dependencies

**Priority:** 🔴 Critical  
**Estimate:** 0.5 hours  
**Dependencies:** None

**Description:**  
Install required npm packages for the Workspace page.

**Subtasks:**

- [ ] Install `date-fns` package: `npm install date-fns`
- [ ] Verify package is added to `package.json`
- [ ] Run `npm install` to update lock file
- [ ] Test import in a sample file

**Acceptance Criteria:**

- ✅ `date-fns` installed and importable
- ✅ No version conflicts with existing packages

---

### Task #2: Create Component Folder Structure

**Priority:** 🔴 Critical  
**Estimate:** 0.25 hours  
**Dependencies:** None

**Description:**  
Create the folder structure for workspace components.

**Subtasks:**

- [ ] Create `/frontend/src/components/workspace/` folder
- [ ] Create placeholder files:
  - [ ] `VideoPlayer.jsx`
  - [ ] `ChatInterface.jsx`
  - [ ] `VideoMetadata.jsx`
  - [ ] `ActionsToolbar.jsx`
- [ ] Add basic component boilerplate to each file
- [ ] Verify imports work from `Workspace.jsx`

**Acceptance Criteria:**

- ✅ Folder structure created
- ✅ All 4 component files exist with basic exports
- ✅ No import errors

---

### Task #3: Extend API Utility

**Priority:** 🔴 Critical  
**Estimate:** 1 hour  
**Dependencies:** None

**Description:**  
Add job-related API functions to `utils/api.js`.

**Subtasks:**

- [ ] Open `/frontend/src/utils/api.js`
- [ ] Add `jobs.get(id)` function
- [ ] Add `jobs.list(params)` function
- [ ] Add JSDoc comments for both functions
- [ ] Test API calls with sample job ID
- [ ] Handle error responses (401, 404, 429)

**Code to Add:**

```javascript
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
```

**Acceptance Criteria:**

- ✅ `jobs.get()` makes correct API call
- ✅ `jobs.list()` supports pagination and filtering
- ✅ Error handling works for all status codes
- ✅ Returns properly typed responses

---

### Task #4: Build Workspace.jsx (Main Container)

**Priority:** 🔴 Critical  
**Estimate:** 4 hours  
**Dependencies:** Task #2, Task #3

**Description:**  
Build the main Workspace page component with state management and API integration.

**Subtasks:**

**Part 1: Basic Structure (1 hour)**

- [ ] Import required dependencies (React, Router, components)
- [ ] Extract `videoId` from URL params using `useParams()`
- [ ] Add UUID validation logic (disabled by default)
- [ ] Create basic layout with breadcrumbs
- [ ] Add navigation links (Library, Video Editor)

**Part 2: State Management (1.5 hours)**

- [ ] Add state: `jobData` (job object from API)
- [ ] Add state: `loading` (boolean for loading state)
- [ ] Add state: `error` (string for error messages)
- [ ] Add state: `pollingInterval` (interval ID for cleanup)
- [ ] Create `fetchJob(videoId)` function
- [ ] Handle all error cases (404, 401, 429, network errors)

**Part 3: Polling Logic (1.5 hours)**

- [ ] Create `useEffect` hook for initial job fetch
- [ ] Implement polling logic for "processing" / "pending" status
- [ ] Poll every 5 seconds when status is not final
- [ ] Stop polling when status is "completed" or "failed"
- [ ] Add 10-minute timeout for polling
- [ ] Cleanup interval on component unmount

**Code Structure:**

```javascript
function Workspace() {
  const { videoId } = useParams();
  const navigate = useNavigate();

  // State
  const [jobData, setJobData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [pollingInterval, setPollingInterval] = useState(null);

  // Fetch job on mount
  useEffect(() => {
    fetchJob(videoId);
  }, [videoId]);

  // Polling logic
  useEffect(() => {
    if (jobData?.status === "processing" || jobData?.status === "pending") {
      const interval = setInterval(() => {
        fetchJob(videoId);
      }, 5000);

      setPollingInterval(interval);

      // Timeout after 10 minutes
      setTimeout(() => clearInterval(interval), 600000);

      return () => clearInterval(interval);
    }
  }, [jobData?.status, videoId]);

  // Render components
  return (
    <div className="workspace-page">
      {/* Breadcrumbs, ActionsToolbar, VideoPlayer, VideoMetadata, ChatInterface */}
    </div>
  );
}
```

**Acceptance Criteria:**

- ✅ Extracts videoId from URL
- ✅ Fetches job data on mount
- ✅ Polls for updates when processing
- ✅ Stops polling when completed/failed
- ✅ Handles all error scenarios
- ✅ Cleans up intervals on unmount

---

### Task #5: Build VideoPlayer Component

**Priority:** 🔴 Critical  
**Estimate:** 3 hours  
**Dependencies:** Task #4

**Description:**  
Build the video player component with HTML5 video controls.

**Subtasks:**

**Part 1: Basic Video Player (1 hour)**

- [ ] Create component structure
- [ ] Add HTML5 `<video>` element
- [ ] Accept props: `videoUrl`, `status`, `aspectRatio`
- [ ] Add loading state when `videoUrl` is null
- [ ] Add error state for video load failures
- [ ] Display video with correct aspect ratio (16:9, 9:16, 1:1)

**Part 2: Custom Controls (1.5 hours)**

- [ ] Add Play/Pause button
- [ ] Add Seek to Beginning button
- [ ] Add Seek to End button
- [ ] Add progress bar (scrubber)
- [ ] Add current time / duration display (e.g., "0:15 / 0:30")
- [ ] Add volume control
- [ ] Add fullscreen toggle

**Part 3: Keyboard Accessibility (0.5 hours)**

- [ ] Add Space key for play/pause
- [ ] Add Arrow Left/Right for seek
- [ ] Add Arrow Up/Down for volume
- [ ] Add F key for fullscreen
- [ ] Add focus indicators for controls

**Component Props:**

```javascript
VideoPlayer.propTypes = {
  videoUrl: PropTypes.string, // S3 presigned URL
  status: PropTypes.string.isRequired, // "pending" | "processing" | "completed" | "failed"
  aspectRatio: PropTypes.string, // "16:9" | "9:16" | "1:1"
  onError: PropTypes.func, // Error callback
};
```

**Acceptance Criteria:**

- ✅ Video loads from S3 presigned URL
- ✅ All controls work correctly
- ✅ Keyboard navigation works
- ✅ Shows loading state when processing
- ✅ Shows error state on load failure
- ✅ Maintains aspect ratio on all screen sizes
- ✅ Progress bar syncs with video playback

---

### Task #6: Build VideoMetadata Component

**Priority:** 🟡 High  
**Estimate:** 2.5 hours  
**Dependencies:** Task #4

**Description:**  
Build the video metadata panel displaying job details.

**Subtasks:**

**Part 1: Display Backend Fields (1 hour)**

- [ ] Create component structure
- [ ] Accept props: `jobData`
- [ ] Display duration (e.g., "30 seconds")
- [ ] Display aspect ratio (e.g., "16:9")
- [ ] Display style (e.g., "luxury, minimal")
- [ ] Display status badge (Completed/Processing/Failed)
- [ ] Display original prompt (multiline text)
- [ ] Display created timestamp (relative time using `date-fns`)

**Part 2: Client-Side Calculated Fields (1 hour)**

- [ ] Calculate cost: `duration × 0.07` (e.g., "$2.10")
- [ ] Display model: "Kling v2.5 Turbo Pro" (hardcoded)
- [ ] Display resolution: "1080p" (assumed)
- [ ] Format relative time: "Generated 2 hours ago"
- [ ] Handle missing fields (show "Not specified")

**Part 3: Editable Title (0.5 hours)**

- [ ] Add editable title field (contentEditable or input)
- [ ] Save title to localStorage with key: `video-title-${jobId}`
- [ ] Load title from localStorage on mount
- [ ] Add purple underline on hover (Aurora theme)
- [ ] Handle empty title (show default: "Untitled Video")

**Component Props:**

```javascript
VideoMetadata.propTypes = {
  jobData: PropTypes.shape({
    job_id: PropTypes.string.isRequired,
    status: PropTypes.string.isRequired,
    prompt: PropTypes.string,
    duration: PropTypes.number,
    style: PropTypes.string,
    aspect_ratio: PropTypes.string,
    created_at: PropTypes.number,
    completed_at: PropTypes.number,
  }).isRequired,
};
```

**Acceptance Criteria:**

- ✅ Displays all backend-provided fields
- ✅ Calculates cost correctly
- ✅ Shows relative time ("2 hours ago")
- ✅ Title is editable and persists to localStorage
- ✅ Status badge uses correct colors (green/orange/red)
- ✅ Missing fields show "Not specified"

---

### Task #7: Build ChatInterface Component (Placeholder)

**Priority:** 🟡 High  
**Estimate:** 2 hours  
**Dependencies:** Task #4

**Description:**  
Build the AI chat interface as a placeholder (no actual edit functionality in MVP).

**Subtasks:**

**Part 1: Chat UI Structure (1 hour)**

- [ ] Create component structure
- [ ] Accept props: `jobData`
- [ ] Add message history container
- [ ] Add input field (multiline textarea)
- [ ] Add submit button (disabled state)
- [ ] Add placeholder text: "Describe the changes you'd like to make..."
- [ ] Style user messages (right-aligned, green accent)
- [ ] Style system messages (left-aligned, purple accent)

**Part 2: Placeholder Functionality (1 hour)**

- [ ] Add initial system message:

  ```
  "✨ Want to try a different version?

  Edit functionality is coming soon! For now, you can:

  → Create a new generation with your changes

  [Create New Video with These Changes]"
  ```

- [ ] Add "Create New Video" button
- [ ] Button navigates to `/create` with pre-filled prompt
- [ ] Store messages in localStorage (optional)
- [ ] Add tooltip to submit button: "Edit functionality coming in Phase 2"
- [ ] Handle Shift+Enter for new line (Enter to submit - disabled)

**Component Props:**

```javascript
ChatInterface.propTypes = {
  jobData: PropTypes.shape({
    job_id: PropTypes.string.isRequired,
    prompt: PropTypes.string,
  }).isRequired,
};
```

**Acceptance Criteria:**

- ✅ Chat UI renders correctly
- ✅ System message displays on mount
- ✅ Input field is disabled but shows placeholder
- ✅ "Create New Video" button works
- ✅ Navigates to Create page with pre-filled prompt
- ✅ Submit button shows tooltip explaining it's disabled

---

### Task #8: Build ActionsToolbar Component

**Priority:** 🟡 High  
**Estimate:** 2 hours  
**Dependencies:** Task #4

**Description:**  
Build the actions toolbar with Download and Delete buttons.

**Subtasks:**

**Part 1: Download Functionality (1 hour)**

- [ ] Create component structure
- [ ] Accept props: `jobData`, `onDownload`, `onDelete`
- [ ] Add Download button with icon
- [ ] Implement download logic:
  - [ ] Create temporary `<a>` element
  - [ ] Set `href` to `jobData.video_url`
  - [ ] Set `download` attribute to filename (e.g., `video-${jobId}.mp4`)
  - [ ] Trigger programmatic click
  - [ ] Clean up element
- [ ] Add tooltip: "Download video"
- [ ] Disable button when status is not "completed"

**Part 2: Delete Functionality (1 hour)**

- [ ] Add Delete button with icon
- [ ] Create confirmation modal:
  - [ ] Title: "Delete Video?"
  - [ ] Message: "Are you sure? This cannot be undone."
  - [ ] Buttons: "Cancel" and "Delete"
- [ ] Show modal on Delete click
- [ ] Add tooltip: "Delete video"
- [ ] Disable button when status is "processing"
- [ ] Navigate to `/library` after deletion (future - no API yet)

**Component Props:**

```javascript
ActionsToolbar.propTypes = {
  jobData: PropTypes.shape({
    job_id: PropTypes.string.isRequired,
    status: PropTypes.string.isRequired,
    video_url: PropTypes.string,
  }).isRequired,
  onDownload: PropTypes.func,
  onDelete: PropTypes.func,
};
```

**Acceptance Criteria:**

- ✅ Download button downloads video file
- ✅ Delete button shows confirmation modal
- ✅ Buttons disabled during processing
- ✅ Tooltips display on hover
- ✅ Download uses presigned S3 URL
- ✅ Confirmation modal blocks accidental deletions

---

### Task #9: Integrate Components in Workspace.jsx

**Priority:** 🔴 Critical  
**Estimate:** 2 hours  
**Dependencies:** Task #4, Task #5, Task #6, Task #7, Task #8

**Description:**  
Wire up all components in the main Workspace page.

**Subtasks:**

- [ ] Import all 4 components
- [ ] Pass `jobData` to VideoPlayer
- [ ] Pass `jobData` to VideoMetadata
- [ ] Pass `jobData` to ChatInterface
- [ ] Pass `jobData` to ActionsToolbar
- [ ] Create `handleDownload()` function
- [ ] Create `handleDelete()` function (placeholder)
- [ ] Add loading state UI
- [ ] Add error state UI
- [ ] Test component communication

**Acceptance Criteria:**

- ✅ All components render correctly
- ✅ Props passed correctly to children
- ✅ State updates propagate to all components
- ✅ Loading/error states display properly

---

### Task #10: Implement Error Handling

**Priority:** 🔴 Critical  
**Estimate:** 2 hours  
**Dependencies:** Task #4

**Description:**  
Implement comprehensive error handling for all error scenarios.

**Subtasks:**

**HTTP Error Handling**

- [ ] Handle 404 (Job Not Found):
  - [ ] Display: "This video doesn't exist or has been deleted."
  - [ ] Show "Return to Library" button
  - [ ] Navigate to `/library` on click
- [ ] Handle 401 (Session Expired):
  - [ ] Display: "Your session has expired. Please log in again."
  - [ ] Redirect to `/login` automatically
- [ ] Handle 429 (Rate Limited):
  - [ ] Display: "Too many requests. Please wait X seconds."
  - [ ] Show countdown timer
  - [ ] Retry after delay using `Retry-After` header
- [ ] Handle 403 (URL Expired - from S3):
  - [ ] Display: "Video link expired. Refreshing..."
  - [ ] Fetch new presigned URL
  - [ ] Retry video load
- [ ] Handle 500+ (Server Error):
  - [ ] Display: "Server error. Please try again."
  - [ ] Show "Retry" button
  - [ ] Log error to console

**Network Error Handling**

- [ ] Handle fetch failures:
  - [ ] Display: "Connection lost. Check your internet."
  - [ ] Auto-retry after 3 seconds
  - [ ] Show manual "Retry" button

**Job Status Error Handling**

- [ ] Handle status = "failed":
  - [ ] Display: "Video generation failed: {error_message}"
  - [ ] Show error from backend
  - [ ] Add "Try Again" button (navigate to Create page)

**Video Load Error Handling**

- [ ] Handle video error event:
  - [ ] Display: "Unable to load video. Please try again."
  - [ ] Show "Refresh" button
  - [ ] Call `GET /api/v1/jobs/{jobId}` to get new URL

**Acceptance Criteria:**

- ✅ All error scenarios handled
- ✅ User-friendly error messages displayed
- ✅ Recovery actions work correctly
- ✅ Errors logged appropriately

---

### Task #11: Implement Status Polling

**Priority:** 🔴 Critical  
**Estimate:** 1.5 hours  
**Dependencies:** Task #4

**Description:**  
Implement status polling for in-progress videos.

**Subtasks:**

- [ ] Check job status on initial load
- [ ] Start polling if status is "processing" or "pending"
- [ ] Poll every 5 seconds
- [ ] Update UI with latest status
- [ ] Stop polling when status is "completed" or "failed"
- [ ] Add 10-minute timeout
- [ ] Display timeout message: "Generation taking longer than expected..."
- [ ] Cleanup interval on component unmount
- [ ] Handle network errors during polling (don't stop polling)

**Polling Logic:**

```javascript
useEffect(() => {
  if (jobData?.status === "processing" || jobData?.status === "pending") {
    const interval = setInterval(async () => {
      const updatedJob = await jobs.get(videoId);
      setJobData(updatedJob);

      if (updatedJob.status === "completed" || updatedJob.status === "failed") {
        clearInterval(interval);
      }
    }, 5000);

    setPollingInterval(interval);

    const timeout = setTimeout(() => {
      clearInterval(interval);
      setError("Generation taking longer than expected...");
    }, 600000);

    return () => {
      clearInterval(interval);
      clearTimeout(timeout);
    };
  }
}, [jobData?.status, videoId]);
```

**Acceptance Criteria:**

- ✅ Polls every 5 seconds when processing
- ✅ Stops polling when completed/failed
- ✅ Timeout after 10 minutes
- ✅ Cleanup on unmount
- ✅ UI updates during polling

---

### Task #12: Create Workspace CSS File

**Priority:** 🟡 High  
**Estimate:** 4 hours  
**Dependencies:** Task #4, Task #5, Task #6, Task #7, Task #8

**Description:**  
Create comprehensive styling for the Workspace page.

**Subtasks:**

**Part 1: Layout Styling (1.5 hours)**

- [ ] Create `/frontend/src/styles/workspace.css`
- [ ] Add page layout (flexbox/grid)
- [ ] Style breadcrumbs navigation
- [ ] Style top bar (breadcrumbs + actions)
- [ ] Style 2-column grid (Metadata 30% + Chat 70%)
- [ ] Add responsive breakpoints (desktop, tablet, mobile)

**Part 2: Component Styling (1.5 hours)**

- [ ] Style VideoPlayer container
- [ ] Style video controls (buttons, progress bar)
- [ ] Style VideoMetadata panel
- [ ] Style ChatInterface container
- [ ] Style ActionsToolbar buttons
- [ ] Add hover states for all interactive elements

**Part 3: Aurora Theme Integration (1 hour)**

- [ ] Use Aurora color palette:
  - [ ] `--aurora-green` (#7CFF00) for play button, active states
  - [ ] `--aurora-purple` (#B44CFF) for secondary actions
  - [ ] `--aurora-teal` (#00FFD1) for info messages
  - [ ] `--aurora-orange` (#FFA500) for processing states
  - [ ] `--error` (#FF4D6A) for error states
- [ ] Use correct typography (Space Grotesk)
- [ ] Use correct surface colors (`--bg`, `--bg-elevated`, `--bg-highlight`)
- [ ] Add glow effects on hover

**CSS Structure:**

```css
/* Layout */
.workspace-page {
  /* Main container */
}
.workspace-breadcrumbs {
  /* Navigation */
}
.workspace-main {
  /* Content area */
}
.workspace-grid {
  /* 2-column grid */
}

/* VideoPlayer */
.video-player-container {
  /* Player wrapper */
}
.video-controls {
  /* Control bar */
}
.video-progress {
  /* Progress bar */
}

/* VideoMetadata */
.video-metadata-panel {
  /* Metadata container */
}
.metadata-field {
  /* Individual field */
}
.status-badge {
  /* Status badge */
}
.editable-title {
  /* Title input */
}

/* ChatInterface */
.chat-interface {
  /* Chat container */
}
.chat-messages {
  /* Message history */
}
.chat-input {
  /* Input field */
}

/* ActionsToolbar */
.actions-toolbar {
  /* Toolbar container */
}
.action-btn {
  /* Action buttons */
}
```

**Acceptance Criteria:**

- ✅ Layout matches wireframe
- ✅ Aurora theme colors used correctly
- ✅ Responsive on all screen sizes
- ✅ Hover states work
- ✅ Typography matches design spec

---

### Task #13: Mobile Responsive Design

**Priority:** 🟡 High  
**Estimate:** 2 hours  
**Dependencies:** Task #13

**Description:**  
Ensure Workspace page is fully responsive on mobile devices.

**Subtasks:**

**Mobile Layout (Tablet: 768px - 1279px)**

- [ ] Stack VideoMetadata and ChatInterface vertically
- [ ] Keep VideoPlayer full width
- [ ] Adjust font sizes for smaller screens
- [ ] Test on iPad resolution

**Mobile Layout (Phone: 320px - 767px)**

- [ ] Full-width VideoPlayer
- [ ] Collapsible VideoMetadata (accordion)
- [ ] Full-width ChatInterface at bottom
- [ ] Simplified controls (hide some buttons)
- [ ] Test on iPhone and Android resolutions

**Touch Interactions**

- [ ] Increase button sizes for touch targets (min 44px)
- [ ] Add touch-friendly controls
- [ ] Test video scrubbing on touch devices

**Acceptance Criteria:**

- ✅ Layout adapts to all screen sizes
- ✅ All controls accessible on mobile
- ✅ Touch interactions work smoothly
- ✅ No horizontal scrolling

---

### Task #14: Loading & Error States Styling

**Priority:** 🟡 High  
**Estimate:** 1 hour  
**Dependencies:** Task #13

**Description:**  
Style loading and error states.

**Subtasks:**

- [ ] Create loading spinner (aurora theme)
- [ ] Style loading message: "Loading video..."
- [ ] Style processing message: "Processing... 45%"
- [ ] Style error messages (red accent)
- [ ] Style "Return to Library" button
- [ ] Style "Retry" buttons
- [ ] Add smooth transitions between states

**Acceptance Criteria:**

- ✅ Loading states look polished
- ✅ Error states are clear and actionable
- ✅ Transitions are smooth

---

### Task #15: Code Review & Cleanup

**Priority:** 🔴 Critical  
**Estimate:** 2 hours  
**Dependencies:** All component, integration, styling, and testing tasks

**Description:**  
Prepare code for deployment.

**Subtasks:**

- [ ] Remove all console.log statements
- [ ] Remove commented-out code
- [ ] Remove unused imports
- [ ] Run ESLint and fix warnings
- [ ] Add PropTypes to all components
- [ ] Add JSDoc comments to functions
- [ ] Format code (Prettier)

**Acceptance Criteria:**

- ✅ No console logs in production code
- ✅ No linter warnings
- ✅ All components have PropTypes
- ✅ Code is well-documented

---

### Task #16: Build & Test Production Bundle

**Priority:** 🔴 Critical  
**Estimate:** 1 hour  
**Dependencies:** Task #20

**Description:**  
Build production bundle and test.

**Subtasks:**

- [ ] Run `npm run build`
- [ ] Verify build completes without errors
- [ ] Test production build locally
- [ ] Check bundle size (should not be significantly larger)
- [ ] Verify all assets load correctly
- [ ] Test with production API endpoints

**Acceptance Criteria:**

- ✅ Production build succeeds
- ✅ Bundle size is reasonable
- ✅ All features work in production mode
- ✅ No console errors

---

## Task Dependencies

### Dependency Chart

```
Task #1 (Install Dependencies)
    │
    ├─► Task #3 (Extend API)
    │
    └─► Task #2 (Folder Structure)
            │
            └─► Task #4 (Workspace.jsx)
                    │
                    ├─► Task #5 (VideoPlayer)
                    ├─► Task #6 (VideoMetadata)
                    ├─► Task #7 (ChatInterface)
                    ├─► Task #8 (ActionsToolbar)
                    │
                    └─► Task #9 (Integration)
                            │
                            ├─► Task #10 (Error Handling)
                            ├─► Task #11 (Polling)
                            ├─► Task #12 (UUID Validation)
                            │
                            └─► Task #13 (CSS)
                                    │
                                    ├─► Task #14 (Responsive)
                                    ├─► Task #15 (States)
                                    │
                                    └─► Task #16 (Manual Testing)
                                            │
                                            ├─► Task #17 (Cross-Browser)
                                            ├─► Task #18 (Accessibility)
                                            ├─► Task #19 (Performance)
                                            │
                                            └─► Task #20 (Cleanup)
                                                    │
                                                    ├─► Task #21 (Build)
                                                    ├─► Task #22 (Docs)
                                                    │
                                                    └─► Task #23 (Deploy)
```

---

## Estimation Summary

### By Phase

| Phase           | Tasks        | Total Hours     | Percentage |
| --------------- | ------------ | --------------- | ---------- |
| **Setup**       | 3 tasks      | 1.75 hours      | 2%         |
| **Components**  | 5 tasks      | 13.5 hours      | 17%        |
| **Integration** | 4 tasks      | 6 hours         | 8%         |
| **Styling**     | 3 tasks      | 7 hours         | 9%         |
| **Testing**     | 4 tasks      | 7 hours         | 9%         |
| **Deployment**  | 4 tasks      | 5 hours         | 6%         |
| **TOTAL**       | **23 tasks** | **40.25 hours** | **100%**   |

### By Priority

| Priority    | Tasks    | Hours       |
| ----------- | -------- | ----------- |
| 🔴 Critical | 10 tasks | 21.75 hours |
| 🟡 High     | 10 tasks | 15 hours    |
| 🟢 Medium   | 3 tasks  | 3.5 hours   |

### By Developer

| Developer      | Tasks    | Hours       | Days (8h/day) |
| -------------- | -------- | ----------- | ------------- |
| Frontend Dev 1 | 23 tasks | 40.25 hours | **~5 days**   |

**Note:** Original estimate was 2 weeks (10 days). With focused effort, MVP can be completed in **5-6 working days**.

---

## Sprint Planning

### Week 1 (Days 1-3)

**Focus:** Setup + Core Components

- **Day 1:** Setup tasks + Start Workspace.jsx
  - Task #1, Task #2, Task #3
  - Task #4 (Part 1 & 2)
- **Day 2:** Finish Workspace.jsx + Start VideoPlayer
  - Task #4 (Part 3)
  - Task #5 (All parts)
- **Day 3:** Build VideoMetadata + ChatInterface
  - Task #6
  - Task #7

### Week 1 (Days 4-5)

**Focus:** Integration + Styling

- **Day 4:** ActionsToolbar + Integration
  - Task #8
  - Task #9, Task #10
- **Day 5:** Polling + Styling
  - Task #11, Task #12
  - Task #13, Task #14

### Week 2 (Days 6-8)

**Focus:** Testing + Polish

- **Day 6:** Testing
  - Task #16 (Manual)
  - Task #17 (Cross-browser)
- **Day 7:** Accessibility + Performance
  - Task #18, Task #19
  - Task #15 (Loading states)
- **Day 8:** Bug fixes from testing

### Week 2 (Days 9-10)

**Focus:** Deployment

- **Day 9:** Code review + Cleanup
  - Task #20
  - Task #21
  - Task #22
- **Day 10:** Deployment + Buffer
  - Task #23
  - Monitor production
  - Fix any critical issues

---

## Definition of Done

A task is considered **DONE** when:

- ✅ All subtasks completed
- ✅ Code reviewed (self-review minimum)
- ✅ Manually tested and working
- ✅ No console errors
- ✅ Styled according to Aurora theme
- ✅ Responsive (if UI component)
- ✅ Accessible (if UI component)
- ✅ PropTypes added (if component)
- ✅ Comments added for complex logic
- ✅ Committed to Git with descriptive message

---

## Risk Mitigation

### Identified Risks

1. **Presigned URL Expiration (7 days)**

   - **Mitigation:** Implement auto-refresh on 403 errors (Task #10)
   - **Test:** Manually test with expired URL

2. **Polling Performance**

   - **Mitigation:** Cleanup intervals on unmount (Task #11)
   - **Test:** Monitor with React DevTools Profiler (Task #19)

3. **Large Video Files (up to 500MB)**

   - **Mitigation:** Use progressive loading (browser handles this)
   - **Test:** Test with 500MB video file (Task #19)

4. **Backend API Changes**

   - **Mitigation:** Verify API contracts before starting (Task #3)
   - **Test:** Test with actual backend early

5. **Cross-Browser Video Support**
   - **Mitigation:** Use standard MP4 format (backend handles this)
   - **Test:** Test on all browsers (Task #17)

---

## Notes

### Development Tips

- **Start with structure:** Build component structure before styling
- **Test early:** Test API integration early to catch issues
- **Use console.log:** Debug polling logic with console logs (remove later)
- **Mobile-first:** Consider mobile layout from the start
- **Incremental commits:** Commit after each completed task

### Common Pitfalls

- ❌ Forgetting to cleanup intervals on unmount → memory leaks
- ❌ Not handling all error cases → bad UX
- ❌ Hardcoding values → inflexible code
- ❌ Skipping accessibility → excludes users
- ❌ Not testing on mobile → broken mobile experience

---

## Appendix

### Reference Documents

- [Workspace PRD](./workspace-prd.md)
- [Workspace Architecture](./workspace-architecture.md)
- [OmniGen Design Spec](../existing_infra_saturday/omnigen_design_spec.md)
- [Aurora Color Palette](../existing_infra_saturday/omnigen_color_palette.md)

### External Resources

- [React Hooks Documentation](https://react.dev/reference/react)
- [HTML5 Video API](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/video)
- [date-fns Documentation](https://date-fns.org/docs/Getting-Started)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)

---

**End of Task List**
