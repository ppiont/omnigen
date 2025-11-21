import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Video, TrendingUp, CheckCircle, Plus } from "lucide-react";
import { jobs as jobsAPI } from "../utils/api";
import StatCard from "../components/StatCard";
import VideoCard from "../components/VideoCard";
import { useAuth } from "../contexts/useAuth.js";
import { getRecentlyOpenedVideos, removeRecentlyOpenedVideo } from "../utils/recentVideos";
import OnboardingSection from "../components/OnboardingSection";
import "../styles/dashboard.css";

function Dashboard() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [jobs, setJobs] = useState([]);
  const [recentVideos, setRecentVideos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch all jobs for stats
      const jobsResponse = await jobsAPI.list({ page: 1, page_size: 50 });
      const allJobs = jobsResponse.jobs || [];
      setJobs(allJobs);

      // Get recently opened videos from localStorage
      const recentlyOpened = getRecentlyOpenedVideos();
      
      // Fetch job data for recently opened videos
      const recentVideoPromises = recentlyOpened
        .slice(0, 4) // Only show top 4
        .map(async (recent) => {
          try {
            const job = await jobsAPI.get(recent.jobId);
            return {
              ...job,
              openedAt: recent.openedAt,
            };
          } catch (err) {
            // If job not found (404), remove it from localStorage to prevent future errors
            if (err.status === 404) {
              removeRecentlyOpenedVideo(recent.jobId);
            } else {
              // Only log non-404 errors to reduce console noise
              console.warn(`Failed to load job ${recent.jobId}:`, err);
            }
            return null;
          }
        });

      const recentVideoResults = await Promise.all(recentVideoPromises);
      const validRecentVideos = recentVideoResults.filter((job) => job !== null);
      
      // Sort by openedAt (most recently opened first)
      validRecentVideos.sort((a, b) => (b.openedAt || 0) - (a.openedAt || 0));
      
      setRecentVideos(validRecentVideos);
    } catch (err) {
      console.error("Failed to load dashboard data:", err);
      setError(err.message || "Failed to load dashboard data");
    } finally {
      setLoading(false);
    }
  };

  // Calculate statistics
  const stats = {
    totalVideos: jobs.length,
    completedVideos: jobs.filter((j) => j.status === "completed").length,
    successRate: jobs.length > 0
      ? Math.round((jobs.filter((j) => j.status === "completed").length / jobs.length) * 100)
      : 0,
  };

  const formatTimeAgo = (timestamp) => {
    if (!timestamp) return "Unknown";
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 60) return "Just now";
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  const handleVideoClick = (jobId) => {
    navigate(`/workspace/${jobId}`);
  };

  return (
    <div className="dashboard-content-wrapper">
      {/* Header */}
      <div className="dashboard-header">
        <div>
          <h1 className="page-title">Dashboard</h1>
          <p className="dashboard-subtitle">
            Welcome back, {user?.name?.split(" ")[0] || "there"}!
          </p>
        </div>
        <button
          className="btn-primary"
          onClick={() => navigate("/create")}
        >
          <Plus size={18} />
          New Video
        </button>
      </div>

      <section className="dashboard-section">
        <OnboardingSection />
      </section>

      {/* Stats Cards */}
      <div className="dashboard-stats">
        <StatCard
          label="Total Videos"
          value={stats.totalVideos}
          icon={<Video size={20} />}
        />
        <StatCard
          label="This Month"
          value={stats.completedVideos}
          icon={<TrendingUp size={20} />}
        />
        <StatCard
          label="Success Rate"
          value={`${stats.successRate}%`}
          icon={<CheckCircle size={20} />}
        />
      </div>

      {/* Recent Videos */}
      <section className="dashboard-section">
        <div className="section-header">
          <h2 className="section-title">Recent Videos</h2>
          <button
            className="btn-text"
            onClick={() => navigate("/library")}
          >
            View all
          </button>
        </div>

        {loading ? (
          <div className="loading-state">Loading your videos...</div>
        ) : error ? (
          <div className="error-state">
            <p>Failed to load videos</p>
            <button onClick={loadDashboardData} className="btn-secondary">
              Try again
            </button>
          </div>
        ) : recentVideos.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              No recently opened videos. Open a video from your library to see it here!
            </p>
            <button
              className="btn-primary"
              onClick={() => navigate("/library")}
            >
              <Plus size={18} />
              View Library
            </button>
          </div>
        ) : (
          <div className="videos-grid-compact">
            {recentVideos.map((job) => (
              <div
                key={job.job_id}
                onClick={() => handleVideoClick(job.job_id)}
                style={{ cursor: "pointer" }}
              >
                <VideoCard
                  video={{
                    id: job.job_id,
                    title: job.prompt || job.title || "Untitled Video",
                    status: job.status === "completed" ? "Completed" : 
                           job.status === "processing" ? "Processing" :
                           job.status === "failed" ? "Failed" : "Pending",
                    createdAt: formatTimeAgo(job.openedAt),
                    duration: job.duration ? `${job.duration}s` : "N/A",
                    aspectRatios: job.aspect_ratio ? [job.aspect_ratio] : ["16:9"],
                    thumbnailUrl: job.metadata?.thumbnail_url || job.thumbnail_url,
                  }}
                  onDownload={() => console.log("Download", job.job_id)}
                  onDelete={() => console.log("Delete", job.job_id)}
                />
              </div>
            ))}
          </div>
        )}
      </section>

    </div>
  );
}

export default Dashboard;
