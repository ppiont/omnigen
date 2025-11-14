# Color Palette Migration Guide

## Teammate's Color Palette (frontend/)

### Aurora Theme Colors
- **Primary (Green)**: `#7cff00` (aurora-green)
- **Secondary (Purple)**: `#b44cff` (aurora-purple)  
- **Tertiary (Teal)**: `#00ffd1` (aurora-teal)
- **Magenta**: `#ff00ff`
- **Orange**: `#ffa500`

### Background Colors
- **Base**: `#0a0e1a` (--bg)
- **Elevated**: `#0f1420` (--bg-elevated)
- **Highlight**: `#1a1f33` (--bg-highlight)
- **Surface**: `#141926` (--bg-surface)

### Text Colors
- **Primary**: `#e8edf5` (--text-primary)
- **Secondary**: `#9ca3b8` (--text-secondary)
- **Muted**: `#6b7188` (--text-muted)

### Status Colors
- **Success**: `#7cff00`
- **Warning**: `#ffa500`
- **Error**: `#ff4d6a`
- **Info**: `#00ffd1`

## My Current Color Palette (app/)

### Dark Mode
- **Primary**: `#8b5cf6` (Purple)
- **Secondary**: `#10b981` (Green)
- **Background**: `#0a0a0a`
- **Surface**: `#1a1a1a`

### Light Mode
- **Background**: `#f0fdf4` (Very light green)
- **Surface**: `#ffffff`
- **Border**: `#d1fae5`
- **Text**: `#1e293b`
- **Text Secondary**: `#64748b`

## Migration Strategy

1. Replace my purple/green scheme with teammate's aurora green/purple/teal
2. Update all background colors to match their darker theme
3. Update text colors to their palette
4. Keep light mode support but adjust to match their aesthetic if they have one
5. Preserve all dashboard functionality

## Files to Update

### Core Config
- `app/tailwind.config.js` - Update color definitions

### Components
- `app/src/components/SurfaceCard.tsx`
- `app/src/components/PrimaryButton.tsx`
- `app/src/components/layout/DashboardLayout.tsx`
- `app/src/components/ui/*` - All UI components

### Pages
- `app/src/pages/Dashboard.tsx`

### Context
- `app/src/context/ThemeContext.tsx` - May need light mode adjustments

## Files NOT to Touch
- `frontend/` - Teammate's login/signup pages (leave untouched)

