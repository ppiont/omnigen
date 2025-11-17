import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Video, TrendingUp, CheckCircle, Plus } from "lucide-react";
import { jobs as jobsAPI, presets as presetsAPI } from "../utils/api";
import StatCard from "../components/StatCard";
import PresetCard from "../components/PresetCard";
import VideoCard from "../components/VideoCard";
import { useAuth } from "../contexts/useAuth.js";
import "../styles/dashboard.css";

// Mock data for dev mode when API is not available
const MOCK_JOBS = [
  {
    job_id: "1",
    prompt: "Product Showcase - Tech Headphones",
    status: "completed",
    created_at: Date.now() - 7200000,
    duration: 30,
    video_url: "https://example.com/video1.mp4",
    thumbnailUrl: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=600&h=400&fit=crop",
  },
  {
    job_id: "2",
    prompt: "Social Ad - Fashion Brand",
    status: "completed",
    created_at: Date.now() - 14400000,
    duration: 15,
    video_url: "https://example.com/video2.mp4",
    thumbnailUrl: "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=600&h=400&fit=crop",
  },
];

const MOCK_PRESETS = [
  {
    id: "1",
    name: "Modern Minimalist",
    description: "Clean, elegant design with smooth transitions",
    style: "Minimalist",
    color_palette: ["#000000", "#FFFFFF", "#7CFF00", "#00E5FF"],
    music_mood: "Calm",
  },
  {
    id: "2",
    name: "Bold & Dynamic",
    description: "High-energy visuals with vibrant colors",
    style: "Bold",
    color_palette: ["#FF006E", "#8338EC", "#3A86FF", "#FFBE0B"],
    music_mood: "Energetic",
  },
  {
    id: "3",
    name: "Corporate Professional",
    description: "Professional, trustworthy brand presence",
    style: "Modern",
    color_palette: ["#1E3A8A", "#3B82F6", "#60A5FA", "#DBEAFE"],
    music_mood: "Professional",
  },
  {
    id: "4",
    name: "Warm & Organic",
    description: "Natural, earth-toned aesthetic",
    style: "Playful",
    color_palette: ["#92400E", "#D97706", "#F59E0B", "#FDE68A"],
    music_mood: "Warm",
  },
];

function Dashboard() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [jobs, setJobs] = useState([]);
  const [presets, setPresets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [useMockData, setUseMockData] = useState(false);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Try to fetch from API
      const [jobsResponse, presetsResponse] = await Promise.all([
        jobsAPI.list({ page: 1, page_size: 20 }),
        presetsAPI.list(),
      ]);

      setJobs(jobsResponse.jobs || []);
      setPresets(presetsResponse.presets || []);
      setUseMockData(false);
    } catch (err) {
      console.log("API not available, using mock data:", err);
      // Use mock data if API is not available
      setJobs(MOCK_JOBS);
      setPresets(MOCK_PRESETS);
      setUseMockData(true);
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

  // Get recent completed jobs
  const recentJobs = jobs
    .filter((j) => j.status === "completed")
    .sort((a, b) => b.created_at - a.created_at)
    .slice(0, 4);

  const formatTimeAgo = (timestamp) => {
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 60) return "Just now";
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  const handleVideoClick = (jobId) => {
    navigate(`/workspace/${jobId}`);
  };

  const handlePresetClick = (preset) => {
    // Navigate to create page with preset pre-selected
    navigate("/create", { state: { preset } });
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

      {/* Quick Start Templates */}
      {presets.length > 0 && (
        <section className="dashboard-section">
          <h2 className="section-title" style={{ marginBottom: '32px' }}>Quick Start Templates</h2>
          <div className="presets-grid">
            {presets.slice(0, 4).map((preset) => (
              <PresetCard
                key={preset.id}
                preset={preset}
                onClick={handlePresetClick}
              />
            ))}
          </div>
        </section>
      )}

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
            <button onClick={loadJobs} className="btn-secondary">
              Try again
            </button>
          </div>
        ) : recentJobs.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              No videos yet. Create your first video to get started!
            </p>
            <button
              className="btn-primary"
              onClick={() => navigate("/create")}
            >
              <Plus size={18} />
              Create Video
            </button>
          </div>
        ) : (
          <div className="videos-grid-compact">
            {recentJobs.map((job) => (
              <div
                key={job.job_id}
                onClick={() => handleVideoClick(job.job_id)}
                style={{ cursor: "pointer" }}
              >
                <VideoCard
                  video={{
                    id: job.job_id,
                    title: job.prompt,
                    status: "Completed",
                    createdAt: formatTimeAgo(job.created_at),
                    duration: `${job.duration}s`,
                    aspectRatios: ["16:9"],
                    thumbnailUrl: job.thumbnailUrl,
                  }}
                  onDownload={() => console.log("Download", job.job_id)}
                  onDelete={() => console.log("Delete", job.job_id)}
                />
              </div>
            ))}
          </div>
        )}
      </section>

      {useMockData && (
        <div className="dev-notice">
          Using mock data - API not available
        </div>
      )}
    </div>
  );
}

export default Dashboard;
