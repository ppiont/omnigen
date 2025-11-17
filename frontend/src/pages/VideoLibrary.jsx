import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import VideoCard from "../components/VideoCard.jsx";
import { jobs } from "../utils/api.js";
import { showToast } from "../utils/toast.js";
import "../styles/dashboard.css";
import "../styles/videos.css";

function VideoLibrary() {
  const navigate = useNavigate();
  const [videos, setVideos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [videoToDelete, setVideoToDelete] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [statusFilter, setStatusFilter] = useState("all");

  // Format time ago
  const formatTimeAgo = (timestamp) => {
    if (!timestamp) return "Unknown";
    const seconds = Math.floor((Date.now() / 1000 - timestamp));
    if (seconds < 60) return "Just now";
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  // Fetch videos from API
  const loadVideos = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await jobs.list();
      const jobsList = response.jobs || [];
      
      // Transform jobs to video format
      const transformedVideos = jobsList.map((job) => ({
        id: job.job_id,
        title: job.prompt || "Untitled Video",
        status: job.status || "processing",
        createdAt: formatTimeAgo(job.created_at),
        duration: job.duration ? `${job.duration}s` : "0s",
        aspectRatios: job.aspect_ratios || ["16:9"],
        thumbnailUrl: job.thumbnail_url,
        videoUrl: job.video_url,
        jobData: job, // Keep full job data for download
      }));
      
      setVideos(transformedVideos);
    } catch (err) {
      console.error("Failed to load videos:", err);
      setError("Failed to load videos. Please try again.");
      showToast("Failed to load videos", "error");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadVideos();
  }, [loadVideos]);

  // Filter videos based on selected status
  const filteredVideos = videos.filter((video) => {
    if (statusFilter === "all") return true;
    return video.status?.toLowerCase() === statusFilter.toLowerCase();
  });

  const handleVideoClick = (videoId) => {
    navigate(`/workspace/${videoId}`);
  };

  const handleStatusFilterChange = (e) => {
    setStatusFilter(e.target.value);
  };

  const handleDeleteClick = (video) => {
    setVideoToDelete(video);
    setDeleteModalOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!videoToDelete || isDeleting) return;

    setIsDeleting(true);
    try {
      await jobs.delete(videoToDelete.id);
      showToast("Video deleted successfully", "success");
      setVideos((prev) => prev.filter((v) => v.id !== videoToDelete.id));
      setDeleteModalOpen(false);
      setVideoToDelete(null);
    } catch (err) {
      console.error("Failed to delete video:", err);
      showToast(
        err.message || "Failed to delete video. Please try again.",
        "error"
      );
    } finally {
      setIsDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteModalOpen(false);
    setVideoToDelete(null);
  };

  const handleDownloadClick = async (video) => {
    if (!video.videoUrl) {
      showToast("Video URL not available", "error");
      return;
    }

    try {
      // Create a temporary anchor element to trigger download
      const link = document.createElement("a");
      link.href = video.videoUrl;
      link.download = `${video.title || "video"}.mp4`;
      link.target = "_blank";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      showToast("Download started", "info");
    } catch (err) {
      console.error("Failed to download video:", err);
      showToast("Failed to start download. Please try again.", "error");
    }
  };

  const handleModalBackdropClick = (e) => {
    if (e.target === e.currentTarget) {
      handleDeleteCancel();
    }
  };

  useEffect(() => {
    if (!deleteModalOpen) return;

    const handleKeyDown = (e) => {
      if (e.key === "Escape") {
        handleDeleteCancel();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [deleteModalOpen]);

  return (
    <div className="video-library-page">
      <h1 className="page-title">Video Library</h1>
      <section className="videos-section">
        <h2 className="section-title">Your videos</h2>
        
        {/* Filters Section */}
        <div className="videos-filters">
          <div className="filter-select">
            <label htmlFor="status-filter">Status</label>
            <select
              id="status-filter"
              value={statusFilter}
              onChange={handleStatusFilterChange}
            >
              <option value="all">All Videos</option>
              <option value="completed">Completed</option>
              <option value="processing">Processing</option>
              <option value="failed">Failed</option>
            </select>
          </div>
        </div>

        {loading ? (
          <div className="empty-state">
            <p className="empty-state-text">Loading videos...</p>
          </div>
        ) : error ? (
          <div className="empty-state">
            <p className="empty-state-text">{error}</p>
            <button onClick={loadVideos} className="btn-primary">
              Try again
            </button>
          </div>
        ) : filteredVideos.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              {videos.length === 0
                ? "Your generated videos will appear here"
                : `No ${statusFilter === "all" ? "" : statusFilter} videos found`}
            </p>
          </div>
        ) : (
          <div className="videos-grid">
            {filteredVideos.map((video) => (
              <div
                key={video.id}
                onClick={() => handleVideoClick(video.id)}
                style={{ cursor: "pointer" }}
              >
                <VideoCard
                  video={video}
                  onDownload={handleDownloadClick}
                  onDelete={handleDeleteClick}
                />
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Delete Confirmation Modal */}
      {deleteModalOpen && (
        <div
          className="modal-backdrop"
          onClick={handleModalBackdropClick}
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-modal-title"
        >
          <div className="modal-card">
            <h2 id="delete-modal-title" className="modal-title">
              Delete Video?
            </h2>
            <p>Are you sure you want to delete this video?</p>
            <div className="modal-actions">
              <button
                type="button"
                className="modal-cancel"
                onClick={handleDeleteCancel}
                disabled={isDeleting}
              >
                Cancel
              </button>
              <button
                type="button"
                className="modal-delete"
                onClick={handleDeleteConfirm}
                disabled={isDeleting}
              >
                {isDeleting ? "Deleting..." : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default VideoLibrary;
