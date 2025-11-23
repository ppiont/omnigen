import PropTypes from "prop-types";

function ScenePreview({
  sceneNumber,
  title,
  description,
  videoUrl,
  thumbnailUrl,
  status,
  duration,
}) {
  return (
    <div className="scene-preview-card">
      <div className="scene-preview-media">
        {videoUrl ? (
          <video
            controls
            preload="metadata"
            poster={thumbnailUrl || undefined}
            aria-label={`Scene ${sceneNumber} preview`}
          >
            <source src={videoUrl} type="video/mp4" />
            Your browser does not support embedded videos.
          </video>
        ) : (
          <div className="scene-preview-placeholder">
            <span>{status === "completed" ? "Processing clip…" : "Awaiting render"}</span>
          </div>
        )}
        <span className={`scene-status scene-status-${status}`}>
          {status === "completed" ? "Completed" : status === "active" ? "Rendering" : "Pending"}
        </span>
      </div>
      <div className="scene-preview-body">
        <div className="scene-preview-header">
          <span className="scene-number">Scene {sceneNumber}</span>
          {duration ? <span className="scene-duration">{duration.toFixed(1)}s</span> : null}
        </div>
        <p className="scene-title">{title || "Untitled scene"}</p>
        {description ? <p className="scene-description">{description}</p> : null}
      </div>
    </div>
  );
}

ScenePreview.propTypes = {
  sceneNumber: PropTypes.number.isRequired,
  title: PropTypes.string,
  description: PropTypes.string,
  videoUrl: PropTypes.string,
  thumbnailUrl: PropTypes.string,
  status: PropTypes.oneOf(["pending", "active", "completed"]).isRequired,
  duration: PropTypes.number,
};

export default ScenePreview;

