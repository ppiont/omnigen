import { Link } from "react-router-dom";
import Logo from "./Logo";
import "../styles/navbar.css";

export default function Navbar() {
  return (
    <nav className="navbar" role="navigation" aria-label="Main navigation">
      <div className="navbar-container">
        <Link to="/" className="navbar-logo" aria-label="Omnigen Home">
          <Logo size={40} />
          <span className="navbar-logo-text">OmniGen</span>
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

