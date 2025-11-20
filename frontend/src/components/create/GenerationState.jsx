import "../../styles/create.css";
import { getStageIcon, shouldDance } from "../../utils/stageIcons";

/**
 * GenerationState - State machine UI component
 * Displays different UI based on generation state:
 * IDLE, RENDERING (with real-time progress), COMPLETED, ERROR
 */
function GenerationState({
  state,
  jobProgress,
  error,
  videoPreview,
  onRetry,
  onViewVideo,
  aspectRatio,
}) {
  // IDLE State - Before generation starts
  if (state === "idle") {
    return (
      <div className="preview-placeholder">
        <div className="preview-aspect-ratio">
          <span className="preview-aspect-text">{aspectRatio}</span>
        </div>
        <p className="preview-placeholder-text">Your video will appear here</p>
      </div>
    );
  }

  // RENDERING State - Real-time progress with dancing icon
  if (state === "rendering" && jobProgress) {
    const { percentage, currentStage, currentStageRaw } = jobProgress;
    const stageInfo = getStageIcon(currentStageRaw);
    const IconComponent = stageInfo.icon;
    const isDancing = shouldDance(currentStageRaw);

    return (
      <div className="preview-state rendering">
        <div className={`generation-progress ${isDancing ? 'dancing-icon' : ''}`}>
          <IconComponent
            size={48}
            color={stageInfo.color}
            strokeWidth={2}
          />
        </div>

        <h4 className="state-title">{currentStage}</h4>

        <div className="state-progress-bar">
          <div
            className="state-progress-fill"
            style={{ width: `${percentage}%` }}
          ></div>
        </div>

        <p className="state-detail">{percentage}%</p>
      </div>
    );
  }

  // COMPLETED State - Video ready with manual navigation
  if (state === "completed") {
    return (
      <div className="preview-state ready">
        <div className="state-icon">✅</div>
        <h4 className="state-title">Video Ready!</h4>
        <p className="state-description">
          Your video has been generated successfully
        </p>

        {videoPreview && (
          <div className="video-thumbnail-preview">
            <img src={videoPreview} alt="Generated video thumbnail" />
            <div className="play-overlay">
              <div className="play-icon">▶</div>
            </div>
          </div>
        )}

        <div className="ready-actions">
          <button className="btn-view-workspace" onClick={onViewVideo}>
            View Video
          </button>
        </div>
      </div>
    );
  }

  // ERROR State - Generation failed
  if (state === "error") {
    return (
      <div className="preview-state error">
        <div className="state-icon">❌</div>
        <h4 className="state-title">Generation Failed</h4>
        <p className="error-message">
          {error || "An unexpected error occurred during generation"}
        </p>
        <button className="btn-retry" onClick={onRetry}>
          Retry Generation
        </button>
      </div>
    );
  }

  // Fallback
  return null;
}

export default GenerationState;
