import { Link } from "react-router-dom";
import { useAuth } from "../contexts/useAuth.js";
import "../styles/not-found.css";

function NotFound() {
  const { isAuthenticated } = useAuth();

  return (
    <div className="not-found-page">
      <div className="not-found-content">
        <h1 className="not-found-code">404</h1>
        <h2 className="not-found-title">Page Not Found</h2>
        <p className="not-found-description">
          The page you're looking for doesn't exist.
        </p>
        {isAuthenticated ? (
          <Link to="/dashboard" className="btn btn-primary">
            Go to Dashboard
          </Link>
        ) : (
          <Link to="/" className="btn btn-primary">
            Go to Home
          </Link>
        )}
      </div>
    </div>
  );
}

export default NotFound;
