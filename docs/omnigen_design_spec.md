# Omnigen Frontend Design Specification v2

## AI Video Generation Platform • Ad Creative Pipeline • Aurora Theme

# 1. Design Goals

The UI represents a professional AI video generation platform focused on creating ad creatives at scale. The experience must feel:

- **Fast and efficient** – Video production pipeline aesthetic
- **Visually striking** – Aurora-inspired color palette
- **Production-ready** – Professional creative tool interface
- **Cost-transparent** – Clear metrics and performance indicators

Visual identity:

- Dark mode with aurora-inspired accents (vibrant greens, purples, teals)
- Video production workflow aesthetic
- Real-time generation progress indicators
- Clear cost and performance metrics
- Signature aurora (Northern Lights) animation in hero section

---

# 2. Global Aesthetic System

## 2.1 Color Tokens (Aurora Palette)

```css
:root {
  /* Backgrounds */
  --bg: #0a0e1a;
  --bg-elevated: #0f1420;
  --bg-highlight: #1a1f33;
  --bg-surface: #141926;

  /* Text */
  --text-primary: #e8edf5;
  --text-secondary: #9ca3b8;
  --text-muted: #6b7188;

  /* Aurora Accents */
  --aurora-green: #7cff00;
  --aurora-purple: #b44cff;
  --aurora-teal: #00ffd1;
  --aurora-magenta: #ff00ff;
  --aurora-orange: #ffa500;

  /* Primary Palette */
  --accent-primary: var(--aurora-green);
  --accent-secondary: var(--aurora-purple);
  --accent-tertiary: var(--aurora-teal);

  /* Functional Colors */
  --success: #7cff00;
  --warning: #ffa500;
  --error: #ff4d6a;
  --info: #00ffd1;

  /* Effects */
  --glow-green: rgba(124, 255, 0, 0.6);
  --glow-purple: rgba(180, 76, 255, 0.5);
  --glow-teal: rgba(0, 255, 209, 0.5);
  --border-glow: rgba(124, 255, 0, 0.55);
  --shadow-strong: 0 12px 50px rgba(0, 0, 0, 0.6);
  --shadow-glow: 0 0 30px rgba(124, 255, 0, 0.2);
}
```

## 2.2 Typography

Font: **Space Grotesk** (tech/modern aesthetic)

Scale:

- Hero: 64–80px, weight 700–800
- Section headers: 32–40px, weight 700
- Subheaders: 20–24px, weight 600
- Body: 15–17px, weight 400–500
- Buttons/Labels: 16–18px, weight 600

## 2.3 Aurora Animation

Updated aurora with vibrant greens, purples, and teals:

```css
.aurora-field::before {
  background: radial-gradient(
      circle at 25% 25%,
      rgba(124, 255, 0, 0.85),
      transparent 55%
    ), radial-gradient(
      circle at 75% 30%,
      rgba(180, 76, 255, 0.75),
      transparent 50%
    ), radial-gradient(circle at 50% 80%, rgba(0, 255, 209, 0.6), transparent
        60%);
  animation: auroraDrift 40s ease-in-out infinite;
}

.aurora-field::after {
  background: radial-gradient(
      circle at 60% 20%,
      rgba(255, 0, 255, 0.7),
      transparent 50%
    ), radial-gradient(
      circle at 30% 70%,
      rgba(124, 255, 0, 0.6),
      transparent 55%
    ), radial-gradient(circle at 80% 85%, rgba(255, 165, 0, 0.5), transparent
        45%);
  animation: auroraDrift 45s ease-in-out infinite reverse;
}
```

## 2.4 Component Surfaces

```css
/* Card/Panel Base */
background: var(--bg-elevated);
border: 1px solid var(--bg-highlight);
border-radius: 14px;
padding: 24px;
box-shadow: var(--shadow-strong);
transition: all 0.3s ease;

/* Hover State */
background: var(--bg-highlight);
border-color: var(--aurora-green);
box-shadow: 0 0 24px var(--glow-green);
transform: translateY(-3px);
```

---

# 3. Content Strategy – Ad Creative Pipeline

## 3.1 Core Messaging

**Value Proposition:**
"Generate professional video ads at scale with AI-powered automation"

**Key Features:**

1. **Multi-Format Generation** – Create ads in 16:9, 9:16, 1:1 simultaneously
2. **Brand Consistency** – Maintain visual identity across all variations
3. **A/B Testing Ready** – Generate multiple creative variations instantly
4. **Cost Efficient** – Under $2 per minute of video content

**Target Users:**

- Marketing teams scaling ad production
- E-commerce brands needing product videos
- Agencies managing multiple clients
- Solo creators building ad portfolios

## 3.2 Video Categories Focus

**Primary:** Ad Creative Pipeline

- Product showcase videos (15-60s)
- Social media ads (TikTok, Instagram, YouTube)
- Brand videos with text overlays
- Multi-format export support

---

# 4. Page Specifications

---

# Page 1 — Landing Page

## 4.1 Objective

Convert visitors by showcasing the power of AI video ad generation. Clear value prop, instant understanding.

## 4.2 Layout

### Hero Section

**Left Column:**

- Eyebrow: "AI Video Generation Platform"
- Hero title: "Create Video Ads at Scale"
- Subtitle: "Generate professional product videos and ad creatives in minutes, not days. Multi-format output, brand-consistent results, production-ready quality."
- CTA Primary: "Start Generating"
- CTA Secondary: "View Demo"

**Right Column:**

- Video preview card showing sample ad generation
- Floating metrics: "2.4s avg render" / "1080p output" / "$1.20/video"

### Feature Strip (4 cards)

**Feature 1: Multi-Format Export**

- Icon: 📱
- Title: "Multi-Format Export"
- Description: "Generate videos in 16:9, 9:16, and 1:1 aspect ratios simultaneously for all platforms."

**Feature 2: Brand Consistency**

- Icon: 🎨
- Title: "Brand Consistency"
- Description: "Apply your brand colors, fonts, and style guidelines across all video variations automatically."

**Feature 3: A/B Test Ready**

- Icon: ⚡
- Title: "A/B Test Ready"
- Description: "Create multiple creative variations instantly to test what resonates with your audience."

**Feature 4: Cost Efficient**

- Icon: 💰
- Title: "Cost Efficient"
- Description: "Generate professional quality videos at under $2 per minute with optimized AI pipelines."

### Pipeline Steps (3 steps)

**Step 01: Brief**

- Title: "Define Your Ad"
- Description: "Describe your product, upload assets, set brand guidelines and creative direction."

**Step 02: Generate**

- Title: "AI Creates Variations"
- Description: "Our pipeline generates multiple ad variations with synced audio, text overlays, and brand styling."

**Step 03: Export**

- Title: "Download & Deploy"
- Description: "Get production-ready videos in all formats, ready to upload to any advertising platform."

### Social Proof / Stats Section

Three stat cards:

- "10,000+ videos generated"
- "98.4% success rate"
- "$1.20 avg cost per video"

### Final CTA

- Title: "Ready to scale your ad production?"
- Subtitle: "Join teams using AI to create hundreds of ad variations in the time it used to take to produce one."
- CTA: "Get Started Free"

## 4.3 Aurora Animation

Applied to hero section with vibrant aurora colors:

- Bright lime green (#7CFF00)
- Purple/magenta (#B44CFF, #FF00FF)
- Teal cyan (#00FFD1)
- Orange horizon glow (#FFA500)
- Slow 40-45s drift animation
- Higher opacity than v1 (0.3-0.4) for more vibrant feel

## 4.4 Responsive Design

- Hero stacks vertically on mobile
- Features become vertical list
- Stats remain 3-column but smaller
- CTAs stack full-width

---

# Page 2 — Login Page

## 5.1 Objective

Quick, frictionless login for returning users.

## 5.2 Layout

### Desktop Two-Column:

**Left Column:**

- Logo: "Omnigen"
- Tagline: "AI Video Generation Platform"
- Value statement: "Create professional video ads in minutes. Join thousands of marketers scaling their creative production with AI."
- Mini stats: "10K+ videos generated" / "98% success rate"

**Right Column:**

- Auth card with elevated surface
- Email field
- Password field
- "Sign In" button (aurora green accent)
- Link to signup

### Visual Updates:

- Aurora gradient background on left column
- Green accent for primary button
- Purple accent for links
- Teal for focus states

## 5.3 Styling

- Input fields with aurora-accent focus glow
- Validation errors in error red
- Loading state with green shimmer

---

# Page 3 — Signup Page

## 6.1 Objective

Convert new users with streamlined onboarding.

## 6.2 Layout

Same two-column layout as login.

**Left Column:**

- Updated value prop: "Start generating video ads in under 60 seconds"
- Benefits list:
  - "Free trial with 5 video generations"
  - "No credit card required"
  - "Export in all formats"

**Right Column:**

- Name field
- Email field
- Password field
- Confirm password field
- "Create Account" button (aurora green)
- Link to login

## 6.3 Styling

- Same aurora-inspired accent system
- Progress indicator showing "Step 1 of 1"
- Success state with green glow animation

---

# Page 4 — Dashboard

## 7.1 Objective

Command center for video ad generation with clear metrics and quick access to creation tools.

## 7.2 Layout

### Sidebar (left)

**Navigation Tabs:**

1. **Create** (highlighted with green glow)
   - Icon: ✨ Sparkle/Star
   - Description: "New video ad"
   - Visual: Green accent gradient, strong glow
2. **Videos**
   - Icon: 🎬 Film
   - Description: "Browse library"
3. **Analytics**
   - Icon: 📊 Chart
   - Description: "Performance data"
4. **Settings**
   - Icon: ⚙️ Gear
   - Description: "Brand & billing"

**Active State:**

- Left accent bar (aurora green)
- Background glow
- Elevated surface

### Main Content Area

**Top Bar:**

- Left: Page title "Ad Creative Studio"
- Right: Search field + Profile avatar

**Stats Row (4 cards):**

**Stat 1: Videos Generated**

- Value: "247"
- Helper: "Last 30 days"
- Trend: "+45% vs last month" (green up arrow)

**Stat 2: Avg Generation Time**

- Value: "2.4s"
- Helper: "Per video render"
- Trend: "-0.3s improvement" (green)

**Stat 3: Cost Efficiency**

- Value: "$1.18"
- Helper: "Per video average"
- Trend: "-$0.12 vs target" (green)

**Stat 4: Success Rate**

- Value: "98.4%"
- Helper: "Successful renders"
- Trend: "±0% stable" (info color)

**Main Panel Grid:**

**Primary Panel: Recent Videos**

- List of recently generated videos
- Thumbnail previews
- Format badges (16:9, 9:16, 1:1)
- Status indicators (Rendering / Complete / Failed)
- Quick actions: Download, Edit, Duplicate

**Secondary Panel: Quick Tips**

- "Use brand presets for consistency"
- "Generate 3+ variations for A/B testing"
- "Export in all formats to maximize reach"

**Activity Feed:**

- "Product showcase video generated - 16:9, 9:16, 1:1"
- "Brand preset 'Tech Minimal' created"
- "Batch generation completed: 12 variations"

### Dashboard Colors

- Green accents for Create tab and success metrics
- Purple accents for secondary actions
- Teal for info/neutral states
- Orange for warnings/pending states

## 7.3 Motion

- Sidebar tabs have subtle hover glow (green)
- Stat cards slide in with stagger
- Video thumbnails have hover zoom
- Create tab pulses gently with green glow

## 7.4 Mobile

- Sidebar collapses to hamburger menu
- Stats stack 2x2 grid
- Panels stack vertically
- Quick action buttons in floating toolbar

---

# 5. Component Specifications

## 5.1 Button Variants

```css
/* Primary Button (Green) */
.btn-primary {
  background: linear-gradient(135deg, var(--aurora-green), var(--aurora-teal));
  color: #0a0e1a;
  box-shadow: 0 0 25px var(--glow-green);
}

/* Secondary Button (Purple) */
.btn-secondary {
  border: 1px solid var(--aurora-purple);
  color: var(--aurora-purple);
  background: rgba(180, 76, 255, 0.08);
}

/* Tertiary Button (Teal) */
.btn-tertiary {
  border: 1px solid var(--aurora-teal);
  color: var(--aurora-teal);
}
```

## 5.2 Video Card Component

New component for displaying generated videos:

```jsx
<VideoCard>
  <Thumbnail src={video.thumbnail} />
  <Badges>
    <Badge>16:9</Badge>
    <Badge>1080p</Badge>
  </Badges>
  <Title>{video.title}</Title>
  <Meta>
    <Time>Generated 2h ago</Time>
    <Cost>$1.20</Cost>
  </Meta>
  <Actions>
    <Button icon="download">Download</Button>
    <Button icon="edit">Edit</Button>
  </Actions>
</VideoCard>
```

## 5.3 Stat Card Updates

Enhanced with color-coded trends:

- Green trend: Positive improvement
- Orange trend: Warning/attention needed
- Red trend: Decline/issue
- Teal trend: Neutral/stable

---

# 6. Animation System

## 6.1 Aurora Drift

- 40-45s duration
- Smooth ease-in-out
- Multi-layer radial gradients
- Higher opacity for visibility (0.3-0.4)

## 6.2 Page Load Sequence

1. Aurora fades in (0.8s)
2. Hero content rises with stagger (1.2s total)
3. Feature cards cascade in (0.15s delay each)

## 6.3 Interaction States

- Hover: 0.25s ease with glow effect
- Click: 0.15s scale and brightness
- Loading: Green shimmer animation
- Success: Green pulse expand

---

# 7. Icon System

Use simple line icons for consistency:

**Navigation:**

- Create: ✨ (sparkle/star)
- Videos: 🎬 (film clapper)
- Analytics: 📊 (bar chart)
- Settings: ⚙️ (gear)

**Features:**

- Multi-format: 📱 (device)
- Brand: 🎨 (palette)
- A/B Test: ⚡ (lightning)
- Cost: 💰 (money)

**Actions:**

- Download: ⬇️ (arrow down)
- Edit: ✏️ (pencil)
- Delete: 🗑️ (trash)
- Duplicate: 📋 (copy)

---

# 8. Accessibility

- WCAG AA contrast compliance (text against dark bg)
- Focus indicators using aurora green
- ARIA labels on all interactive elements
- Keyboard navigation support
- Screen reader announcements for generation status

---

# 9. Component File Structure

```
frontend/
  src/
    components/
      Sidebar.jsx
      FeatureCard.jsx
      StatCard.jsx
      VideoCard.jsx (NEW)
      VideoGrid.jsx (NEW)
      GenerationStatus.jsx (NEW)
    pages/
      Landing.jsx
      Login.jsx
      Signup.jsx
      Dashboard.jsx
    styles/
      global.css (UPDATE colors)
      aurora.css (UPDATE with new gradient)
      landing.css (UPDATE content)
      auth.css (UPDATE colors)
      dashboard.css (UPDATE colors)
      video-card.css (NEW)
```

---

# 10. Data Integration Points

For future backend integration:

**Landing Page:**

- Real-time stats API (videos generated, avg cost)

**Dashboard:**

- User video library API
- Generation queue status
- Cost tracking metrics
- Usage analytics

**Video Generation:**

- Generation request API
- Status polling endpoint
- Download/export URLs
- Format selection

---

# 11. Required Deliverables

1. **Updated Color System** – Aurora palette across all pages
2. **Updated Content** – Ad Creative Pipeline messaging
3. **Updated Components** – VideoCard, enhanced StatCard
4. **Updated Aurora Animation** – Vibrant green/purple/teal
5. **Updated Landing Page** – Video ad generation focus
6. **Updated Dashboard** – Video library and metrics
7. **Updated Auth Pages** – Ad generation value props
8. **Responsive Design** – All pages mobile-optimized

---

# 12. Design Principles

**Clarity Over Complexity**

- Clear CTAs and navigation
- Obvious next steps
- Simple mental model

**Performance Feedback**

- Show generation progress
- Display cost estimates
- Indicate success/failure clearly

**Professional Polish**

- Production-quality UI
- Smooth animations
- Consistent spacing and typography

**Aurora Identity**

- Vibrant but not overwhelming
- Green = primary/create actions
- Purple = secondary/navigation
- Teal = info/neutral
- Orange = warnings/pending

---

# End of Design Specification v2
