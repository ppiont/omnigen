import { Navigate } from "react-router-dom";
import { useAuth } from "../contexts/useAuth.js";

const LOCAL_DEV_MODE = import.meta.env.VITE_LOCAL_DEV_MODE === "true";

/**
 * ProtectedRoute component that redirects to login if user is not authenticated
 */
export default function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();

  if (LOCAL_DEV_MODE) {
    return children;
  }

  // Show loading state while checking authentication
  if (loading) {
    return (
      <div className="auth-loading">
        <div className="loading-spinner"></div>
        <p>Loading...</p>
      </div>
    );
  }

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // User is authenticated, render children
  return children;
}
