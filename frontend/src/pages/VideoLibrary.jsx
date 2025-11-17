import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import VideoCard from "../components/VideoCard.jsx";
import { jobs } from "../utils/api.js";
import "../styles/dashboard.css";

const LOCAL_DEV_MODE = import.meta.env.VITE_LOCAL_DEV_MODE === "true";

const MOCK_VIDEOS = [
  {
    job_id: "job-11111111-2222-3333-4444-555555555555",
    prompt: "Product Showcase - Tech Headphones",
    status: "completed",
    created_at: Math.floor(Date.now() / 1000) - 7200,
    duration: 30,
    aspect_ratio: "16:9",
  },
  {
    job_id: "job-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    prompt: "Social Ad - Fashion Brand",
    status: "completed",
    created_at: Math.floor(Date.now() / 1000) - 14400,
    duration: 15,
    aspect_ratio: "9:16",
  },
  {
    job_id: "job-99999999-8888-7777-6666-555555555555",
    prompt: "Explainer Video - SaaS Platform",
    status: "processing",
    created_at: Math.floor(Date.now() / 1000) - 86400,
    duration: 60,
    aspect_ratio: "16:9",
  },
];

const USE_SAMPLE_VIDEOS_ON_AUTH_ERROR = LOCAL_DEV_MODE;

const mapJobToVideoCard = (job) => ({
  id: job.job_id,
  title: job.prompt || "Untitled Video",
  status: job.status
    ? job.status.charAt(0).toUpperCase() + job.status.slice(1)
    : "Processing",
  createdAt: job.created_at
    ? new Date(job.created_at * 1000).toLocaleString()
    : "Unknown",
  duration: job.duration ? `${job.duration}s` : "Unknown",
  aspectRatios: job.aspect_ratio ? [job.aspect_ratio] : ["16:9"],
  thumbnailUrl: job.thumbnailUrl || null,
});

function VideoLibrary() {
  const navigate = useNavigate();
  const [videos, setVideos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [usingMockData, setUsingMockData] = useState(false);

  const handleVideoClick = (videoId) => {
    navigate(`/workspace/${videoId}`);
  };

  // Fetch videos from API
  useEffect(() => {
    const fetchVideos = async () => {
      try {
        setLoading(true);
        const response = await jobs.list({ page_size: 50 });
        const transformedVideos = (response?.jobs || []).map(mapJobToVideoCard);

        setVideos(transformedVideos);
        setUsingMockData(false);
        setError(null);
      } catch (err) {
        console.error("Failed to fetch videos:", err);

        // Use mock data for local development/offline backend scenarios
        if (!navigator.onLine || err.status === 0) {
          setVideos(MOCK_VIDEOS.map(mapJobToVideoCard));
          setUsingMockData(true);
          setError(null);
        } else if (err.status === 401 || err.status === 403) {
          if (USE_SAMPLE_VIDEOS_ON_AUTH_ERROR) {
            setVideos(MOCK_VIDEOS.map(mapJobToVideoCard));
            setUsingMockData(true);
            setError(null);
          } else {
            setError("You need to sign in again to view your videos.");
          }
        } else {
          setError("Failed to load videos. Please try again.");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchVideos();
  }, []);

  if (loading) {
    return (
      <div className="video-library-page">
        <h1 className="page-title">Video Library</h1>
        <div className="loading-state">
          <p>Loading your videos...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="video-library-page">
        <h1 className="page-title">Video Library</h1>
        <div className="error-state">
          <p>{error}</p>
          <button onClick={() => window.location.reload()}>
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="video-library-page">
      <h1 className="page-title">Video Library</h1>
      <section className="videos-section">
        {usingMockData && (
          <div className="info-banner">
            Showing sample videos because the API is unavailable in this environment.
          </div>
        )}
        <h2 className="section-title">Your videos</h2>
        {videos.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              Your generated videos will appear here
            </p>
          </div>
        ) : (
          <div className="videos-grid">
            {videos.map((video) => (
              <div
                key={video.id}
                onClick={() => handleVideoClick(video.id)}
                style={{ cursor: "pointer" }}
              >
                <VideoCard
                  video={video}
                  onDownload={() => console.log('Download', video.id)}
                  onDelete={() => console.log('Delete', video.id)}
                />
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

export default VideoLibrary;
