# OmniGen UI Task Plan (Frontend-First)

**Context**: 4-person team focusing solely on the frontend/UI this week. Backend integration, Step Functions wiring, and Lambda work will resume afterwards. The goal is to deliver a fully navigable, high-fidelity UI that mirrors the PRD flow.

---

## Team Roles (Frontend Focus)

1. **Dev A – Experience Lead**
   - Owns landing, auth, navigation polish, and global theme consistency.
2. **Dev B – Creation Flow Owner**
   - Enhances `/create`, prompt UX, state handling, and progress surfaces.
3. **Dev C – Workspace Lead**
   - Builds the new `/workspace/:jobId` experience (player, timeline, chat UI).
4. **Dev D – Library & Shared Components**
   - Owns `/videos`, video cards, shared modals, and global utilities.

Shared responsibilities:
- Daily UI sync (15 min) to keep visual language aligned with Aurora theme.
- Dev C + Dev B coordinate on handing off generation context to workspace.
- Dev D + Dev A ensure cards/widgets share tokens and responsive behavior.

---

## High-Level Flow (UI Only)

1. **Landing → Signup/Login** (Dev A)
2. **Dashboard → Create Page** (Dev B)
3. **Create → Workspace** (Dev B → Dev C)
4. **Videos Library → Workspace** (Dev D → Dev C)
5. **Settings** (Dev A) – password form polish, placeholder for future prefs

Backend calls replaced with mocked data/services where needed.

---

## Detailed Task Breakdown

### Dev A – Experience Lead
1. **Landing Page Refresh**
   - Align hero copy/metrics with PRD.
   - Ensure Aurora animation performance on Safari/Chromium.
   - Add CTA buttons linking to `/signup` and `/create` (if authed).
2. **Auth Pages**
   - Two-column layout per spec.
   - Insert brand stats + testimonials.
   - Add form validation states (client-side only for now).
3. **Global Layout & Navigation**
   - Audit `AppLayout`, `Navbar`, `Sidebar` for consistent spacing/shadows.
   - Implement route-level skeleton loaders for SPA navigation.
   - Document theme tokens in `global.css` (typography, shadows, gradients).
4. **Settings Page (UI)**
   - Styling for password change card.
   - Add placeholders for brand presets + API keys (disabled states).

### Dev B – Create Flow Owner
1. **Prompt Panel**
   - Expand advanced options to match PRD (brand presets, reference image upload UI, future batch toggle – disabled state).
   - Add inline helper text + validation (character counts, missing prompt, etc.).
2. **Preview & Progress**
   - Replace fake progress bar with state machine-ready UI (Pending → Planning → Rendering → Ready) using mocked data.
   - Add scene preview thumbnails placeholder grid.
3. **CTA & Routing**
   - After “Generate” success (simulated), navigate to `/workspace/:jobId` with mocked jobId and pass context via router state.
   - Provide “View workspace” button when generation completes.
4. **Responsive Behavior**
   - Ensure create page works on tablet/mobile (stacked layout, collapsible panels).

### Dev C – Workspace Lead
1. **Route + Layout**
   - Create `/workspace/:jobId` page with AppLayout wrapper.
   - Divide page into three columns: video player, timeline, chat.
2. **Video Player & Timeline**
   - Build player mock with 16:9 & 9:16 toggles.
   - Scene timeline component showing segments, statuses, hover details.
   - Add toolbar actions (Download, Duplicate, Export) – disable pending backend.
3. **Chat Interface**
   - UI for conversation list, composer, quick actions (“Make brighter”, etc.).
   - Display system messages for mock events (Scene regeneration, audio update).
4. **Scene Inspector**
   - Right panel cards per scene with controls (Regenerate, Adjust timing, Replace) – trigger mock modals.
5. **Empty & Error States**
   - No job found, job still processing, job failed.

### Dev D – Library & Shared Components
1. **Video Library Page**
   - Replace static array with mocked data service (job list with statuses, formats, durations).
   - Add filters (status, aspect ratio, date range) – UI only.
   - Wire each card to `/workspace/:jobId` with router navigation.
2. **VideoCard Enhancements**
   - Include status badge, cost, success rate, aspect ratio chips.
   - Hover actions (View, Duplicate, Delete) – UI only.
3. **Shared Modals & Toasts**
   - Build modal framework for confirmations and scene actions.
   - Toast notifications for success/error using context provider.
4. **Utility Components**
   - Loading skeletons for cards/grid.
   - Empty state illustrations (tie into Aurora theme).

---

## Coordination & Deliverables

- **Design Tokens**: Dev A finalizes tokens, others import from `global.css`.
- **Mock Services**: Dev D creates `src/services/mockJobs.js` supplying deterministic data for Create, Workspace, Videos.
- **Routing Contracts**: Dev B & C agree on router state payload shape (jobId, prompt summary, scenes array).
- **Daily Checkpoints**:
  - Day 1: Wireframes ready in code, base layouts committed.
  - Day 2: Interactive states + mock data wired.
  - Day 3: Polish pass, cross-page QA, demo build.

---

## Out of Scope (This Sprint)
- Real backend calls (Step Functions, Replicate, DynamoDB).
- Auth integration beyond existing Cognito wrapper.
- Video upload/download functionality.
- Payment/billing UI.

Backend tasks remain paused until UI prototype is complete.
