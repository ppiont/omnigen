import { Play, Download, Trash2 } from "lucide-react";

/**
 * VideoCard component - displays a single video with metadata and actions
 * @param {Object} video - Video object with id, title, status, createdAt, duration, aspectRatios, thumbnailUrl
 * @param {Function} onDownload - Callback when download button is clicked
 * @param {Function} onDelete - Callback when delete button is clicked
 */
function VideoCard({ video, onDownload, onDelete }) {
  const isCompleted = video.status?.toLowerCase() === "completed";
  const statusClass = video.status ? video.status.toLowerCase() : "processing";

  const thumbnailStyle = video.thumbnailUrl
    ? {
        backgroundImage: `linear-gradient(120deg, rgba(10, 14, 26, 0.65), rgba(20, 25, 38, 0.65)), url(${video.thumbnailUrl})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
      }
    : undefined;

  return (
    <article className="video-card">
      <div className="video-thumbnail" style={thumbnailStyle}>
        <span className={`status-badge ${statusClass}`}>{video.status}</span>
        <div className="thumbnail-overlay" aria-hidden="true">
          <Play size={20} />
          <span>Preview</span>
        </div>
      </div>

      <div className="video-card-body">
        <h3 className="video-card-title" title={video.title}>
          {video.title}
        </h3>
        <div className="video-card-meta">
          <span className="video-date">{video.createdAt}</span>
          <span className="video-duration">
            <Play size={14} />
            {video.duration}
          </span>
        </div>
        <div className="aspect-badges" aria-label="Aspect ratios">
          {video.aspectRatios.map((ratio) => (
            <span className="aspect-badge" key={`${video.id}-${ratio}`}>
              {ratio}
            </span>
          ))}
        </div>
      </div>

      <div className="video-card-actions">
        <button
          type="button"
          onClick={() => onDownload?.(video)}
          disabled={!isCompleted}
          aria-label={
            isCompleted ? "Download video" : "Video must be completed to download"
          }
          title={isCompleted ? "Download" : "Video must be completed"}
        >
          <Download size={16} />
          <span>Download</span>
        </button>
        <button
          type="button"
          onClick={() => onDelete?.(video)}
          disabled={!isCompleted}
          aria-label={isCompleted ? "Delete video" : "Video must be completed to delete"}
          title={isCompleted ? "Delete" : "Video must be completed"}
        >
          <Trash2 size={16} />
        </button>
      </div>
    </article>
  );
}

export default VideoCard;
