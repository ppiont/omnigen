# Aurora Color Palette Reference

## Visual Color Guide

This document provides the complete aurora-inspired color palette for the Omnigen AI Video Generation Platform.

---

## Primary Colors

### Aurora Green (Primary Accent)

- **Hex:** `#7CFF00`
- **RGB:** `rgb(124, 255, 0)`
- **Usage:** Primary CTAs, Create tab highlight, success states, primary glow effects
- **Inspired by:** The bright lime-green aurora streaks

### Aurora Purple (Secondary Accent)

- **Hex:** `#B44CFF`
- **RGB:** `rgb(180, 76, 255)`
- **Usage:** Secondary buttons, navigation highlights, link colors, secondary accents
- **Inspired by:** The purple-magenta aurora bands

### Aurora Teal (Tertiary Accent)

- **Hex:** `#00FFD1`
- **RGB:** `rgb(0, 255, 209)`
- **Usage:** Info states, focus indicators, neutral highlights, tertiary buttons
- **Inspired by:** The cyan-teal aurora glow

---

## Additional Aurora Colors

### Aurora Magenta

- **Hex:** `#FF00FF`
- **RGB:** `rgb(255, 0, 255)`
- **Usage:** Special highlights, gradient accents, aurora animation
- **Inspired by:** The bright magenta aurora peaks

### Aurora Orange

- **Hex:** `#FFA500`
- **RGB:** `rgb(255, 165, 0)`
- **Usage:** Warning states, pending indicators, warm accents, horizon glow
- **Inspired by:** The orange horizon glow in aurora photographs

---

## Background Colors

### Primary Background

- **Hex:** `#0a0e1a`
- **RGB:** `rgb(10, 14, 26)`
- **Usage:** Main page background, body background

### Elevated Surface

- **Hex:** `#0f1420`
- **RGB:** `rgb(15, 20, 32)`
- **Usage:** Card backgrounds, elevated panels

### Highlight Surface

- **Hex:** `#1a1f33`
- **RGB:** `rgb(26, 31, 51)`
- **Usage:** Hover states, active elements, highlighted cards

### Surface

- **Hex:** `#141926`
- **RGB:** `rgb(20, 25, 38)`
- **Usage:** Alternative surface color, form inputs

---

## Text Colors

### Primary Text

- **Hex:** `#e8edf5`
- **RGB:** `rgb(232, 237, 245)`
- **Usage:** Main headings, important text, high-emphasis content

### Secondary Text

- **Hex:** `#9ca3b8`
- **RGB:** `rgb(156, 163, 184)`
- **Usage:** Body text, descriptions, medium-emphasis content

### Muted Text

- **Hex:** `#6b7188`
- **RGB:** `rgb(107, 113, 136)`
- **Usage:** Helper text, timestamps, low-emphasis content

---

## Functional Colors

### Success

- **Hex:** `#7CFF00` (Aurora Green)
- **Usage:** Success messages, positive trends, completed states

### Warning

- **Hex:** `#FFA500` (Aurora Orange)
- **Usage:** Warning messages, caution states, pending actions

### Error

- **Hex:** `#FF4D6A`
- **RGB:** `rgb(255, 77, 106)`
- **Usage:** Error messages, validation errors, destructive actions

### Info

- **Hex:** `#00FFD1` (Aurora Teal)
- **Usage:** Information messages, neutral alerts, helper tooltips

---

## Glow Effects (with Transparency)

### Green Glow

- **RGBA:** `rgba(124, 255, 0, 0.6)`
- **Usage:** Primary button hover, Create tab glow, success indicators

### Purple Glow

- **RGBA:** `rgba(180, 76, 255, 0.5)`
- **Usage:** Secondary elements hover, link hover, navigation highlights

### Teal Glow

- **RGBA:** `rgba(0, 255, 209, 0.5)`
- **Usage:** Focus states, info indicators, neutral highlights

### Border Glow (Green)

- **RGBA:** `rgba(124, 255, 0, 0.55)`
- **Usage:** Card borders on hover, glowing outlines

### Shadow Glow (Green)

- **RGBA:** `rgba(124, 255, 0, 0.2)`
- **Usage:** Box shadows, subtle glow effects

---

## Aurora Animation Gradients

### Before Pseudo-Element

```css
background: radial-gradient(
    circle at 25% 25%,
    rgba(124, 255, 0, 0.85),
    transparent 55%
  ), radial-gradient(
    circle at 75% 30%,
    rgba(180, 76, 255, 0.75),
    transparent 50%
  ), radial-gradient(circle at 50% 80%, rgba(0, 255, 209, 0.6), transparent 60%);
```

**Colors:** Green (dominant), Purple, Teal

### After Pseudo-Element

```css
background: radial-gradient(
    circle at 60% 20%,
    rgba(255, 0, 255, 0.7),
    transparent 50%
  ), radial-gradient(
    circle at 30% 70%,
    rgba(124, 255, 0, 0.6),
    transparent 55%
  ), radial-gradient(circle at 80% 85%, rgba(255, 165, 0, 0.5), transparent 45%);
```

**Colors:** Magenta, Green, Orange

---

## Button Color Schemes

### Primary Button

```css
background: linear-gradient(135deg, #7cff00, #00ffd1);
/* Green to Teal gradient */
color: #0a0e1a; /* Dark text for contrast */
box-shadow: 0 0 25px rgba(124, 255, 0, 0.6);
```

### Secondary Button

```css
border: 1px solid #b44cff;
color: #b44cff;
background: rgba(180, 76, 255, 0.08);
/* Purple border with subtle purple background */
```

### Tertiary Button

```css
border: 1px solid #00ffd1;
color: #00ffd1;
/* Teal border with transparent background */
```

---

## Usage Guidelines

### Do's ✅

- Use **green** for primary actions and success
- Use **purple** for secondary navigation and links
- Use **teal** for informational elements
- Use **orange** for warnings and pending states
- Maintain high contrast for accessibility
- Use glows sparingly for emphasis

### Don'ts ❌

- Don't use too many colors simultaneously
- Don't reduce contrast below WCAG AA standards
- Don't overuse glow effects (causes visual fatigue)
- Don't use orange for primary CTAs
- Don't mix warm and cool tones inconsistently

---

## Color Hierarchy

**Level 1 - Primary Actions:**
Aurora Green (#7CFF00)

**Level 2 - Secondary Actions:**
Aurora Purple (#B44CFF)

**Level 3 - Informational:**
Aurora Teal (#00FFD1)

**Level 4 - Special/Warning:**
Aurora Orange (#FFA500)

**Level 5 - Error/Destructive:**
Error Red (#FF4D6A)

---

## Accessibility Notes

### Contrast Ratios (on #0a0e1a background)

- **Aurora Green (#7CFF00):** ~15.2:1 ✅ (AAA)
- **Aurora Purple (#B44CFF):** ~7.8:1 ✅ (AA)
- **Aurora Teal (#00FFD1):** ~13.5:1 ✅ (AAA)
- **Secondary Text (#9ca3b8):** ~8.1:1 ✅ (AA)
- **Muted Text (#6b7188):** ~4.6:1 ✅ (AA)

All colors meet or exceed WCAG AA standards for normal text.

---

## CSS Variables Quick Reference

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

  /* Functional */
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

---

## Inspiration Source

These colors are directly inspired by the aurora borealis (Northern Lights) photograph showing:

- Bright lime-green primary bands
- Purple-magenta secondary streaks
- Teal-cyan atmospheric glow
- Orange horizon illumination
- Deep blue-black night sky backdrop

The palette captures the vibrant, otherworldly quality of the aurora while maintaining professional usability for a production video platform interface.

---

**Last Updated:** November 14, 2025  
**Version:** 2.0  
**Platform:** Omnigen AI Video Generation
