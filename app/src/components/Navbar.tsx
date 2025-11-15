import { Link } from "react-router-dom";
import Logo from "./Logo";
import "../styles/navbar.css";

export default function Navbar() {
  return (
    <nav className="navbar" role="navigation" aria-label="Main navigation">
      <div className="navbar-container">
        <Link to="/" className="navbar-logo" aria-label="Omnigen Home" style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <Logo size={48} />
          <span style={{ 
            fontFamily: "'Space Grotesk', system-ui, sans-serif",
            fontSize: "1.5rem",
            fontWeight: "600",
            letterSpacing: "0.04em",
            background: "linear-gradient(to right, #e8edf5 0%, #9ca3b8 40%, #b44cff 100%)",
            WebkitBackgroundClip: "text",
            WebkitTextFillColor: "transparent",
            backgroundClip: "text"
          }}>
            OmniGen
          </span>
        </Link>

        <div className="navbar-links">
          <Link to="/login" className="navbar-link">
            Login
          </Link>
          <Link to="/signup" className="navbar-link navbar-link-primary">
            Get Started
          </Link>
        </div>
      </div>
    </nav>
  );
}

