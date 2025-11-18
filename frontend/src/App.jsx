import { Routes, Route, useLocation } from "react-router-dom";
import { ThemeProvider } from "./context/ThemeContext.jsx";
import Navbar from "./components/Navbar.jsx";
import AppLayout from "./components/AppLayout.jsx";
import ProtectedRoute from "./components/ProtectedRoute.jsx";
import Landing from "./pages/Landing.jsx";
import Login from "./pages/Login.jsx";
import Signup from "./pages/Signup.jsx";
import Verify from "./pages/Verify.jsx";
import ForgotPassword from "./pages/ForgotPassword.jsx";
import Dashboard from "./pages/Dashboard.jsx";
import Create from "./pages/Create.jsx";
import VideoLibrary from "./pages/VideoLibrary.jsx";
import Workspace from "./pages/Workspace.jsx";
import Settings from "./pages/Settings.jsx";
import NotFound from "./pages/NotFound.jsx";
import Onboarding from "./pages/Onboarding.jsx";

function App() {
  const location = useLocation();

  // Navbar only on public pages
  const publicRoutes = ["/", "/login", "/signup", "/verify", "/forgot-password"];
  const showNavbar = publicRoutes.includes(location.pathname);

  return (
    <ThemeProvider>
      <div className="app-shell">
        {showNavbar && <Navbar />}
        <Routes>
          {/* Public Routes */}
          <Route path="/" element={<Landing />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          <Route path="/verify" element={<Verify />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />

          {/* Protected Routes with AppLayout (Sidebar) */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <Dashboard />
                </AppLayout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/create"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <Create />
                </AppLayout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/library"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <VideoLibrary />
                </AppLayout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/workspace/:videoId"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <Workspace />
                </AppLayout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/settings"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <Settings />
                </AppLayout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/onboarding"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <Onboarding />
                </AppLayout>
              </ProtectedRoute>
            }
          />

          {/* 404 Catch-all */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </div>
    </ThemeProvider>
  );
}

export default App;
