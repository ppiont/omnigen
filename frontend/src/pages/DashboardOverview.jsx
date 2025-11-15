import { useNavigate } from "react-router-dom";
import AppLayout from "../components/AppLayout.jsx";

function DashboardOverview() {
  const navigate = useNavigate();

  // Mock analytics data
  const analytics = {
    totalVideos: 6,
    videosThisWeek: 3,
    totalCost: "$9.80",
    avgDuration: "35s",
  };

  return (
    <AppLayout>
      <section className="dashboard-overview">
        <div className="welcome-section">
          <h1 className="welcome-title">Welcome to OMNIGEN</h1>
          <p className="welcome-subtitle">
            Your AI-powered video generation platform. Create stunning video ads in minutes.
          </p>
        </div>

        <div className="analytics-grid">
          <div className="analytics-card">
            <div className="analytics-icon">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="4" y="4" width="16" height="16" rx="2" />
                <path d="M9 8l5 4-5 4V8z" fill="currentColor" stroke="none" />
              </svg>
            </div>
            <div className="analytics-content">
              <p className="analytics-label">Total Videos</p>
              <p className="analytics-value">{analytics.totalVideos}</p>
            </div>
          </div>

          <div className="analytics-card">
            <div className="analytics-icon">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </div>
            <div className="analytics-content">
              <p className="analytics-label">Total Spent</p>
              <p className="analytics-value">{analytics.totalCost}</p>
            </div>
          </div>

          <div className="analytics-card">
            <div className="analytics-icon">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </div>
            <div className="analytics-content">
              <p className="analytics-label">Avg Duration</p>
              <p className="analytics-value">{analytics.avgDuration}</p>
            </div>
          </div>

          <div className="analytics-card">
            <div className="analytics-icon">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </div>
            <div className="analytics-content">
              <p className="analytics-label">This Week</p>
              <p className="analytics-value">{analytics.videosThisWeek}</p>
            </div>
          </div>
        </div>

        <div className="quick-actions">
          <h2 className="section-title">Quick Actions</h2>
          <div className="quick-actions-grid">
            <button
              className="quick-action-card"
              onClick={() => navigate("/create")}
            >
              <div className="quick-action-icon">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M12 3.5l1.4 4.1 4.1 1.4-4.1 1.4L12 14.5l-1.4-4.1-4.1-1.4 4.1-1.4z" strokeLinejoin="round" />
                  <circle cx="12" cy="12" r="2.2" />
                </svg>
              </div>
              <h3 className="quick-action-title">Create Video</h3>
              <p className="quick-action-description">Generate a new video ad with AI</p>
            </button>

            <button
              className="quick-action-card"
              onClick={() => navigate("/videos")}
            >
              <div className="quick-action-icon">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <rect x="4" y="4" width="16" height="16" rx="2" />
                  <path d="M9 8l5 4-5 4V8z" fill="currentColor" stroke="none" />
                </svg>
              </div>
              <h3 className="quick-action-title">Browse Videos</h3>
              <p className="quick-action-description">View your video library</p>
            </button>

            <button
              className="quick-action-card"
              onClick={() => navigate("/settings")}
            >
              <div className="quick-action-icon">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M12 8.25A3.75 3.75 0 1112 15.75 3.75 3.75 0 0112 8.25z" />
                  <path d="M4.5 12.75V11.25l2.1-.7a5.36 5.36 0 01.9-1.54l-.48-2.2 1.06-1.06 2.2.48a5.36 5.36 0 011.54-.9l.7-2.1h1.5l.7 2.1a5.36 5.36 0 011.54.9l2.2-.48 1.06 1.06-.48 2.2a5.36 5.36 0 01.9 1.54l2.1.7v1.5l-2.1.7a5.36 5.36 0 01-.9 1.54l.48 2.2-1.06 1.06-2.2-.48a5.36 5.36 0 01-1.54.9l-.7 2.1h-1.5l-.7-2.1a5.36 5.36 0 01-1.54-.9l-2.2.48-1.06-1.06.48-2.2a5.36 5.36 0 01-.9-1.54z" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </div>
              <h3 className="quick-action-title">Settings</h3>
              <p className="quick-action-description">Manage your account</p>
            </button>
          </div>
        </div>
      </section>
    </AppLayout>
  );
}

export default DashboardOverview;
