import "../../styles/create.css";

/**
 * GenerationState - State machine UI component
 * Displays different UI based on generation state:
 * IDLE, PENDING, PLANNING, RENDERING, STITCHING, READY, ERROR
 */
function GenerationState({
  state,
  progress,
  error,
  sceneCount,
  currentScene,
  videoPreview,
  onRetry,
  onViewWorkspace,
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

  // PENDING State - Job submitted, waiting
  if (state === "pending") {
    return (
      <div className="preview-state pending">
        <div className="state-icon">⏳</div>
        <h4 className="state-title">Job Submitted</h4>
        <p className="state-description">Waiting for processing slot...</p>
        <div className="state-spinner"></div>
      </div>
    );
  }

  // PLANNING State - AI analyzing prompt
  if (state === "planning") {
    return (
      <div className="preview-state planning">
        <div className="state-icon">🧠</div>
        <h4 className="state-title">Planning Scenes</h4>
        <p className="state-description">
          AI is analyzing your prompt and creating a narrative structure...
        </p>
        <div className="state-progress-bar">
          <div
            className="state-progress-fill"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
        {sceneCount > 0 && (
          <p className="state-detail">Scene breakdown: {sceneCount} clips planned</p>
        )}
      </div>
    );
  }

  // RENDERING State - Generating video clips
  if (state === "rendering") {
    return (
      <div className="preview-state rendering">
        <div className="state-icon">🎬</div>
        <h4 className="state-title">Rendering Video</h4>
        <p className="state-description">
          Generating clips with AI video models...
        </p>
        <div className="state-progress-bar">
          <div
            className="state-progress-fill"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
        {currentScene && sceneCount && (
          <p className="state-detail">
            Clip {currentScene} of {sceneCount} complete
          </p>
        )}
      </div>
    );
  }

  // STITCHING State - Post-generation processing
  if (state === "stitching") {
    return (
      <div className="preview-state stitching">
        <div className="state-icon">✂️</div>
        <h4 className="state-title">Finalizing Video</h4>
        <p className="state-description">
          Stitching clips together and adding final touches...
        </p>
        <div className="state-progress-bar">
          <div
            className="state-progress-fill"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
        {sceneCount > 1 && (
          <p className="state-detail">
            Combining {sceneCount} clips into final video
          </p>
        )}
      </div>
    );
  }

  // READY State - Video complete
  if (state === "ready") {
    return (
      <div className="preview-state ready">
        <div className="state-icon">✅</div>
        <h4 className="state-title">Video Ready!</h4>
        <p className="state-description">Your video has been generated successfully</p>

        {videoPreview && (
          <div className="video-thumbnail-preview">
            <img src={videoPreview} alt="Generated video thumbnail" />
            <div className="play-overlay">
              <div className="play-icon">▶</div>
            </div>
          </div>
        )}

        <div className="ready-actions">
          <button
            className="btn-view-workspace"
            onClick={onViewWorkspace}
          >
            View in Workspace
          </button>
          <button
            className="btn-generate-another"
            onClick={onRetry}
          >
            Generate Another
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
