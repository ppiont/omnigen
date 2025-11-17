import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import VideoCard from "../components/VideoCard.jsx";
import { jobs } from "../utils/api.js";
import "../styles/dashboard.css";

// Helper function to format relative time
const formatRelativeTime = (timestamp) => {
  if (!timestamp) return "Unknown";
  
  const now = Math.floor(Date.now() / 1000);
  const diff = now - timestamp;
  
  if (diff < 60) return "Just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 604800) return `${Math.floor(diff / 86400)}d ago`;
  return `${Math.floor(diff / 604800)}w ago`;
};

// Helper function to truncate prompt for title
const truncatePrompt = (text, maxLength) => {
  if (!text) return "Untitled Video";
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + "...";
};

function VideoLibrary() {
  const navigate = useNavigate();
  const [videos, setVideos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchVideos = async () => {
      try {
        console.log("[VIDEO_LIBRARY] 📥 Fetching jobs from API...");
        setLoading(true);
        setError(null);
        
        const response = await jobs.list({ page: 1, page_size: 50 });
        console.log("[VIDEO_LIBRARY] ✅ Jobs received:", response);
        
        // Transform backend job format to VideoCard format
        const transformedVideos = (response.jobs || []).map((job) => {
          // Format created_at timestamp to relative time
          const createdAt = job.created_at 
            ? formatRelativeTime(job.created_at)
            : "Unknown";
          
          // Get thumbnail from metadata or use a placeholder
          const thumbnailUrl = job.metadata?.thumbnail_url || 
            (job.video_url ? `${job.video_url}?thumbnail=true` : null);
          
          return {
            id: job.job_id,
            title: job.prompt ? truncatePrompt(job.prompt, 50) : "Untitled Video",
            status: job.status === "completed" ? "Completed" : 
                   job.status === "processing" ? "Processing" :
                   job.status === "failed" ? "Failed" : "Pending",
            createdAt,
            duration: job.duration ? `${job.duration}s` : "N/A",
            aspectRatios: job.aspect_ratio ? [job.aspect_ratio] : ["16:9"],
            thumbnailUrl,
            videoUrl: job.video_url,
            jobData: job, // Store full job data for potential use
          };
        });
        
        console.log("[VIDEO_LIBRARY] 📊 Transformed videos:", transformedVideos);
        setVideos(transformedVideos);
      } catch (err) {
        console.error("[VIDEO_LIBRARY] ❌ Error fetching videos:", err);
        setError(err.message || "Failed to load videos");
      } finally {
        setLoading(false);
      }
    };

    fetchVideos();
  }, []);

  const handleVideoClick = (videoId) => {
    console.log("[VIDEO_LIBRARY] 🖱️ Video clicked:", videoId);
    navigate(`/workspace/${videoId}`);
  };

  if (loading) {
    return (
      <div className="video-library-page">
        <h1 className="page-title">Video Library</h1>
        <div className="empty-state">
          <p className="empty-state-text">Loading videos...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="video-library-page">
        <h1 className="page-title">Video Library</h1>
        <div className="empty-state">
          <p className="empty-state-text">Error: {error}</p>
          <button 
            onClick={() => window.location.reload()} 
            className="btn-retry"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="video-library-page">
      <h1 className="page-title">Video Library</h1>
      <section className="videos-section">
        <h2 className="section-title">Your videos</h2>
        {videos.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              Your generated videos will appear here
            </p>
            <p className="empty-state-subtext">
              Create your first video from the Create page
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
