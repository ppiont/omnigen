import { Navigate } from "react-router-dom";
import { useAuth } from "../contexts/useAuth.js";

/**
 * ProtectedRoute component that redirects to login if user is not authenticated
 * 
 * TEMPORARY: Set DISABLE_AUTH to true to bypass authentication for local testing
 */
const DISABLE_AUTH = false; // Set to false to re-enable authentication

export default function ProtectedRoute({ children }) {
  // Bypass authentication check if disabled
  if (DISABLE_AUTH) {
    return children;
  }

  const { isAuthenticated, loading } = useAuth();
  
  if (loading) {
    return (
      <div className="auth-loading">
        <div className="loading-spinner"></div>
        <p>Loading...</p>
      </div>
    );
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return children;
}
