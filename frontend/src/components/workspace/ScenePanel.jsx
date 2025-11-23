import { useState } from "react";
import PropTypes from "prop-types";
import { RefreshCw, Play, ChevronDown, ChevronUp, Loader2 } from "lucide-react";
import { jobs } from "../../utils/api.js";
import { showToast } from "../../utils/toast.js";

/**
 * ScenePanel component displays individual scene clips with regeneration capability
 * @param {Object} props - Component props
 */
function ScenePanel({
  jobId,
  scenes = [],
  sceneVideoUrls = [],
  sceneVersions = {},
  onSceneRegenerated,
  disabled = false,
}) {
  const [expandedScene, setExpandedScene] = useState(null);
  const [regeneratingScenes, setRegeneratingScenes] = useState({});
  const [playingScene, setPlayingScene] = useState(null);

  const handleToggleExpand = (sceneNumber) => {
    setExpandedScene(expandedScene === sceneNumber ? null : sceneNumber);
  };

  const handleRegenerate = async (sceneNumber, cascade = false) => {
    if (regeneratingScenes[sceneNumber] || disabled) return;

    setRegeneratingScenes((prev) => ({ ...prev, [sceneNumber]: true }));
    showToast(`Regenerating scene ${sceneNumber}...`, "info");

    try {
      const result = await jobs.regenerateScene(jobId, sceneNumber, { cascade });

      showToast(
        cascade && result.cascade_count > 0
          ? `Regenerated scene ${sceneNumber} and ${result.cascade_count} subsequent scenes`
          : `Scene ${sceneNumber} regenerated (version ${result.new_version})`,
        "success"
      );

      // Notify parent to refresh job data
      if (onSceneRegenerated) {
        onSceneRegenerated(result);
      }
    } catch (error) {
      console.error("Scene regeneration failed:", error);
      showToast(
        error.message || `Failed to regenerate scene ${sceneNumber}`,
        "error"
      );
    } finally {
      setRegeneratingScenes((prev) => ({ ...prev, [sceneNumber]: false }));
    }
  };

  const handlePlayScene = (sceneNumber) => {
    setPlayingScene(playingScene === sceneNumber ? null : sceneNumber);
  };

  if (!sceneVideoUrls || sceneVideoUrls.length === 0) {
    return (
      <div className="scene-panel scene-panel-empty">
        <p>No scenes available</p>
      </div>
    );
  }

  return (
    <div className="scene-panel">
      <div className="scene-panel-header">
        <h3>Scenes ({sceneVideoUrls.length})</h3>
      </div>
      <div className="scene-panel-list">
        {sceneVideoUrls.map((videoUrl, index) => {
          const sceneNumber = index + 1;
          const scene = scenes[index] || {};
          const version = sceneVersions[sceneNumber] || 1;
          const isExpanded = expandedScene === sceneNumber;
          const isRegenerating = regeneratingScenes[sceneNumber];
          const isPlaying = playingScene === sceneNumber;

          return (
            <div
              key={sceneNumber}
              className={`scene-card ${isExpanded ? "scene-card-expanded" : ""}`}
            >
              <div
                className="scene-card-header"
                onClick={() => handleToggleExpand(sceneNumber)}
              >
                <div className="scene-card-title">
                  <span className="scene-number">Scene {sceneNumber}</span>
                  <span className="scene-version">v{version}</span>
                </div>
                <div className="scene-card-actions">
                  {isExpanded ? (
                    <ChevronUp size={18} />
                  ) : (
                    <ChevronDown size={18} />
                  )}
                </div>
              </div>

              {isExpanded && (
                <div className="scene-card-content">
                  <div className="scene-video-container">
                    {isPlaying ? (
                      <video
                        src={videoUrl}
                        controls
                        autoPlay
                        className="scene-video"
                        onEnded={() => setPlayingScene(null)}
                      />
                    ) : (
                      <div
                        className="scene-video-placeholder"
                        onClick={() => handlePlayScene(sceneNumber)}
                      >
                        <Play size={32} />
                        <span>Play Scene</span>
                      </div>
                    )}
                  </div>

                  {scene.location && (
                    <div className="scene-info">
                      <strong>Location:</strong> {scene.location}
                    </div>
                  )}
                  {scene.action && (
                    <div className="scene-info">
                      <strong>Action:</strong> {scene.action}
                    </div>
                  )}
                  {scene.duration && (
                    <div className="scene-info">
                      <strong>Duration:</strong> {scene.duration}s
                    </div>
                  )}

                  <div className="scene-actions">
                    <button
                      className="scene-btn scene-btn-regenerate"
                      onClick={() => handleRegenerate(sceneNumber, false)}
                      disabled={isRegenerating || disabled}
                    >
                      {isRegenerating ? (
                        <>
                          <Loader2 size={16} className="spin" />
                          <span>Regenerating...</span>
                        </>
                      ) : (
                        <>
                          <RefreshCw size={16} />
                          <span>Regenerate</span>
                        </>
                      )}
                    </button>

                    {sceneNumber < sceneVideoUrls.length && (
                      <button
                        className="scene-btn scene-btn-cascade"
                        onClick={() => handleRegenerate(sceneNumber, true)}
                        disabled={isRegenerating || disabled}
                        title="Regenerate this scene and all following scenes"
                      >
                        <RefreshCw size={16} />
                        <span>Regenerate All After</span>
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

ScenePanel.propTypes = {
  jobId: PropTypes.string.isRequired,
  scenes: PropTypes.arrayOf(
    PropTypes.shape({
      scene_number: PropTypes.number,
      location: PropTypes.string,
      action: PropTypes.string,
      duration: PropTypes.number,
    })
  ),
  sceneVideoUrls: PropTypes.arrayOf(PropTypes.string),
  sceneVersions: PropTypes.object,
  onSceneRegenerated: PropTypes.func,
  disabled: PropTypes.bool,
};

export default ScenePanel;
