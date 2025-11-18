// import { Navigate } from "react-router-dom";
// import { useAuth } from "../contexts/useAuth.js";

/**
 * ProtectedRoute component that redirects to login if user is not authenticated
 * 
 * AUTH DISABLED FOR TESTING - Always allows access
 */
export default function ProtectedRoute({ children }) {
  // AUTH DISABLED: Always render children without checking authentication
  return children;
  
  // Original auth check (disabled):
  // const { isAuthenticated, loading } = useAuth();
  // if (loading) {
  //   return (
  //     <div className="auth-loading">
  //       <div className="loading-spinner"></div>
  //       <p>Loading...</p>
  //     </div>
  //   );
  // }
  // if (!isAuthenticated) {
  //   return <Navigate to="/login" replace />;
  // }
  // return children;
}
