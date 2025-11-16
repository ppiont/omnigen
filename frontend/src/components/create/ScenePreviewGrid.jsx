import "../../styles/create.css";

/**
 * ScenePreviewGrid - Displays planned/rendered scene thumbnails
 * Shows during PLANNING, RENDERING, and READY states
 */
function ScenePreviewGrid({ scenes, isVisible }) {
  if (!isVisible || !scenes || scenes.length === 0) {
    return null;
  }

  const getStatusIcon = (status) => {
    switch (status) {
      case "complete":
        return "✅";
      case "rendering":
        return "⏳";
      case "pending":
        return "⏸️";
      case "error":
        return "❌";
      default:
        return "⏸️";
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
              <span className="status-icon">{getStatusIcon(scene.status)}</span>
              <span className="status-text">{getStatusText(scene.status)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default ScenePreviewGrid;
