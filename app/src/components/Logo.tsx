type LogoProps = {
  size?: number
  showText?: boolean
  className?: string
}

export default function Logo({ size = 72, showText = false, className }: LogoProps) {
  const uniqueId = `logo-${Math.random().toString(36).substr(2, 9)}`
  
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 84 84"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      style={{ display: 'block' }}
    >
      <defs>
        {/* Aurora gradient for prism */}
        <linearGradient id={`omnigen-aurora-${uniqueId}`} x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#7cff00" />
          <stop offset="30%" stopColor="#00ffd1" />
          <stop offset="65%" stopColor="#b44cff" />
          <stop offset="100%" stopColor="#ff00ff" />
        </linearGradient>

        {/* Soft glow */}
        <filter
          id={`omnigen-glow-${uniqueId}`}
          x="-40%"
          y="-40%"
          width="180%"
          height="180%"
        >
          <feGaussianBlur stdDeviation="10" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        {/* Gradient for text stroke/fill */}
        {showText && (
          <linearGradient
            id={`omnigen-text-gradient-${uniqueId}`}
            x1="0"
            y1="0"
            x2="1"
            y2="0"
          >
            <stop offset="0%" stopColor="#e8edf5" />
            <stop offset="40%" stopColor="#9ca3b8" />
            <stop offset="100%" stopColor="#b44cff" />
          </linearGradient>
        )}
      </defs>

      {/* Prism / crystal symbol - centered in viewBox */}
      <g transform="translate(10, 6)" filter={`url(#omnigen-glow-${uniqueId})`}>
        {/* Faceted hex prism */}
        <path
          d="M42 0 L74 18 L74 54 L42 72 L10 54 L10 18 Z"
          fill={`url(#omnigen-aurora-${uniqueId})`}
        />
        {/* Inner highlight facet */}
        <path
          d="M42 8 L66 20 L66 50 L42 62 L18 50 L18 20 Z"
          fill="rgba(10,14,26,0.75)"
        />
        {/* Light diagonal highlight */}
        <path
          d="M18 22 L42 34 L42 60 L18 48 Z"
          fill="rgba(255,255,255,0.06)"
        />
        <path d="M42 34 L66 22 L66 48 L42 60 Z" fill="rgba(0,0,0,0.25)" />
      </g>

      {/* Wordmark (optional) */}
      {showText && (
        <g transform="translate(90, 50)">
          <text
            x="0"
            y="0"
            fontFamily="'Space Grotesk', system-ui, -apple-system, BlinkMacSystemFont, 'SF Pro Text', sans-serif"
            fontSize="44"
            fontWeight="600"
            letterSpacing="0.04em"
            fill={`url(#omnigen-text-gradient-${uniqueId})`}
          >
            OmniGen
          </text>
        </g>
      )}
    </svg>
  )
}

