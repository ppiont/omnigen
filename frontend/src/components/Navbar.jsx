import { Link } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import "../styles/navbar.css";

function Navbar() {
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <nav className="navbar" role="navigation" aria-label="Main navigation">
      <div className="navbar-container">
        <Link to="/" className="navbar-logo" aria-label="Omnigen Home">
          {/* SVG VERSION */}
          <svg
            width="420"
            height="120"
            viewBox="0 0 420 120"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            {/* <!-- Background is transparent on purpose --> */}
            <defs>
              {/* <!-- Aurora gradient for prism --> */}
              <linearGradient id="omnigen-aurora" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stopColor="#7cff00" />
                <stop offset="30%" stopColor="#00ffd1" />
                <stop offset="65%" stopColor="#b44cff" />
                <stop offset="100%" stopColor="#ff00ff" />
              </linearGradient>

              {/* <!-- Soft glow --> */}
              <filter
                id="omnigen-glow"
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

              {/* <!-- Gradient for text stroke/fill --> */}
              <linearGradient
                id="omnigen-text-gradient"
                x1="0"
                y1="0"
                x2="1"
                y2="0"
              >
                <stop offset="0%" stopColor="#e8edf5" />
                <stop offset="40%" stopColor="#9ca3b8" />
                <stop offset="100%" stopColor="#b44cff" />
              </linearGradient>
            </defs>

            {/* <!-- Prism / crystal symbol --> */}
            <g transform="translate(0,20)" filter="url(#omnigen-glow)">
              {/* <!-- Faceted hex prism --> */}
              <path
                d="M42 0 L74 18 L74 54 L42 72 L10 54 L10 18 Z"
                fill="url(#omnigen-aurora)"
              />
              {/* <!-- Inner highlight facet --> */}
              <path
                d="M42 8 L66 20 L66 50 L42 62 L18 50 L18 20 Z"
                fill="rgba(10,14,26,0.75)"
              />
              {/* <!-- Light diagonal highlight --> */}
              <path
                d="M18 22 L42 34 L42 60 L18 48 Z"
                fill="rgba(255,255,255,0.06)"
              />
              <path d="M42 34 L66 22 L66 48 L42 60 Z" fill="rgba(0,0,0,0.25)" />
            </g>

            {/* <!-- Wordmark --> */}
            <g transform="translate(100,75)">
              <text
                x="0"
                y="0"
                fontFamily="'Space Grotesk', system-ui, -apple-system, BlinkMacSystemFont, 'SF Pro Text', sans-serif"
                fontSize="44"
                fontWeight="600"
                letterSpacing="0.04em"
                fill="url(#omnigen-text-gradient)"
              >
                OmniGen
              </text>
            </g>
          </svg>
        </Link>

        <div className="navbar-links">
          {isAuthenticated ? (
            <>
              <Link to="/dashboard" className="navbar-link">
                Dashboard
              </Link>
              <span className="navbar-user">{user?.email}</span>
              <button onClick={logout} className="navbar-link navbar-link-secondary">
                Logout
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="navbar-link">
                Login
              </Link>
              <Link to="/signup" className="navbar-link navbar-link-primary">
                Get Started
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}

export default Navbar;
