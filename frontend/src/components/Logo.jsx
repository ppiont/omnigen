function Logo({ size = "default" }) {
  const dimensions = 
    size === "small" ? { width: 40, height: 40 } : 
    size === "medium" ? { width: 56, height: 56 } : 
    size === "large" ? { width: 64, height: 64 } :
    { width: 84, height: 72 };

  return (
    <svg
      width={dimensions.width}
      height={dimensions.height}
      viewBox="0 0 84 92"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      style={{
        willChange: 'transform',
        transform: 'translateZ(0)', // GPU acceleration
      }}
    >
      <defs>
        {/* Mint green and baby blue gradient for prism - matches Navbar */}
        <linearGradient id="logo-aurora" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#A8E6CF" />
          <stop offset="50%" stopColor="#7FD4B0" />
          <stop offset="100%" stopColor="#B3E5FC" />
        </linearGradient>

        {/* Soft glow - matches Navbar */}
        <filter
          id="logo-glow"
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
      </defs>

      <g
        transform="translate(0,10)"
        filter="url(#logo-glow)"
        style={{ willChange: 'filter' }}
      >
        {/* Faceted hex prism */}
        <path
          d="M42 0 L74 18 L74 54 L42 72 L10 54 L10 18 Z"
          fill="url(#logo-aurora)"
        />
        {/* Inner highlight facet - matches Navbar */}
        <path
          d="M42 8 L66 20 L66 50 L42 62 L18 50 L18 20 Z"
          fill="rgba(255,255,255,0.3)"
        />
        {/* Light diagonal highlight */}
        <path
          d="M18 22 L42 34 L42 60 L18 48 Z"
          fill="rgba(255,255,255,0.06)"
        />
        <path d="M42 34 L66 22 L66 48 L42 60 Z" fill="rgba(0,0,0,0.25)" />
      </g>
    </svg>
  );
}

export default Logo;
