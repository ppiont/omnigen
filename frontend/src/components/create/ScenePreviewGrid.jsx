import "../../styles/create.css";
import { Film, CheckCircle, XCircle, Clock } from "lucide-react";

/**
 * ScenePreviewGrid - Displays planned/rendered scene thumbnails
 * Shows during RENDERING and COMPLETED states with real-time updates
 */
function ScenePreviewGrid({ scenes, isVisible, jobProgress }) {
  if (!isVisible || !scenes || scenes.length === 0) {
    return null;
  }

  // Enhanced status icon with lucide icons
  const getStatusIcon = (status) => {
    switch (status) {
      case "complete":
        return <CheckCircle size={16} color="#10b981" />;
      case "rendering":
        return <Film size={16} color="#8b5cf6" className="dancing-icon" />;
      case "pending":
        return <Clock size={16} color="#6b7280" />;
      case "error":
        return <XCircle size={16} color="#ef4444" />;
      default:
        return <Clock size={16} color="#6b7280" />;
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case "complete":
        return "Ready";
      case "rendering":
        return "Rendering...";
      case "pending":
        return "Pending";
      case "error":
        return "Failed";
      default:
        return "Pending";
    }
  };

  return (
    <div className="scene-previews-container">
      <h5 className="scene-previews-title">Planned Scenes</h5>
      <div className="scenes-grid">
        {scenes.map((scene, idx) => (
          <div
            key={scene.id || idx}
            className={`scene-card scene-card-${scene.status}`}
          >
            <div className="scene-number">Scene {idx + 1}</div>

            <div className="scene-thumbnail">
              {scene.thumbnailUrl ? (
                <img
                  src={scene.thumbnailUrl}
                  alt={`Scene ${idx + 1}`}
                  className="scene-thumbnail-img"
                />
              ) : (
                <div className="scene-placeholder">
                  <div className="scene-placeholder-icon">🎬</div>
                </div>
              )}
            </div>

            <div className="scene-info">
              <p className="scene-description">{scene.description}</p>
              {scene.duration && (
                <span className="scene-duration">{scene.duration}</span>
              )}
            </div>

            <div className={`scene-status scene-status-${scene.status}`}>
              <div className="status-icon">{getStatusIcon(scene.status)}</div>
              <span className="status-text">{getStatusText(scene.status)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default ScenePreviewGrid;
