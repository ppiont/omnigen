/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // Aurora Theme - Matching teammate's palette
        // Backgrounds
        background: '#0a0e1a', // --bg
        'bg-elevated': '#0f1420', // --bg-elevated
        'bg-highlight': '#1a1f33', // --bg-highlight
        surface: '#141926', // --bg-surface

        // Aurora Accents
        primary: '#7cff00', // --aurora-green (primary accent)
        secondary: '#b44cff', // --aurora-purple (secondary accent)
        'aurora-teal': '#00ffd1', // --aurora-teal (tertiary)
        'aurora-magenta': '#ff00ff',
        'aurora-orange': '#ffa500',

        // Text Colors
        foreground: {
          DEFAULT: '#e8edf5', // --text-primary
          secondary: '#9ca3b8', // --text-secondary
          muted: '#6b7188', // --text-muted
        },

        // Status Colors
        success: '#7cff00',
        warning: '#ffa500',
        error: '#ff4d6a',
        info: '#00ffd1',

        // Light mode colors (original light theme)
        'light-bg': '#f0fdf4', // Very light green
        'light-surface': '#ffffff', // White
        'light-border': '#d1fae5', // Light green border
        'light-text': '#1e293b', // Dark text
        'light-text-secondary': '#64748b', // Medium gray
        'light-accent': '#dcfce7', // Light green accent
      },
    },
  },
  plugins: [],
}

