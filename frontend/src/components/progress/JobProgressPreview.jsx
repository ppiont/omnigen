import PropTypes from "prop-types";
import { useState } from "react";
import ScenePreview from "./ScenePreview.jsx";
import AudioPreview from "./AudioPreview.jsx";
import JobControls from "../intervention/JobControls.jsx";
import SceneEditor from "../intervention/SceneEditor.jsx";
import ParameterEditor from "../intervention/ParameterEditor.jsx";
import { useJobIntervention } from "../../hooks/useJobIntervention.js";

const EMPTY_ARRAY = [];
const DEFAULT_STAGE_ORDER = [
  { key: "script", label: "Script Generation" },
  { key: "audio", label: "Audio Mix" },
  { key: "scenes", label: "Scene Rendering" },
  { key: "compose", label: "Final Composition" },
];

const formatDuration = (seconds) => {
  if (!seconds && seconds !== 0) return "Calculating…";
  if (seconds < 60) return `${Math.max(1, Math.round(seconds))}s remaining`;
  const minutes = Math.floor(seconds / 60);
  const secs = Math.round(seconds % 60);
  return `${minutes}m ${secs}s remaining`;
};

function ProgressStage({ label, status, description }) {
  return (
    <div className={`progress-stage progress-stage-${status}`}>
      <div className="progress-stage-indicator" aria-hidden="true" />
      <div>
        <p className="progress-stage-label">{label}</p>
        {description ? <p className="progress-stage-description">{description}</p> : null}
      </div>
    </div>
  );
}

ProgressStage.propTypes = {
  label: PropTypes.string.isRequired,
  status: PropTypes.oneOf(["pending", "active", "completed"]).isRequired,
  description: PropTypes.string,
};

function ScriptPreview({ scenes }) {
  if (!scenes.length) {
    return (
      <div className="script-preview-empty">
        <p>Script is warming up. You&apos;ll see the detailed storyboard as soon as it is ready.</p>
      </div>
    );
  }

  return (
    <ol className="script-preview-list">
      {scenes.map((scene) => (
        <li key={scene.scene_number || scene.SceneNumber} className="script-preview-item">
          <div className="script-preview-item-header">
            <span className="script-preview-scene-number">
              Scene {scene.scene_number || scene.SceneNumber}
            </span>
            {scene.duration ? (
              <span className="script-preview-duration">
                {(scene.duration || 0).toFixed(1)}s
              </span>
            ) : null}
          </div>
          <p className="script-preview-location">
            {scene.location || scene.Location || "Scene"}
          </p>
          <p className="script-preview-action">
            {scene.action || scene.Action || scene.generation_prompt || "Generating visuals…"}
          </p>
        </li>
      ))}
    </ol>
  );
}

ScriptPreview.propTypes = {
  scenes: PropTypes.arrayOf(PropTypes.object).isRequired,
};

function FinalPreview({ finalVideoUrl, thumbnailUrl, status }) {
  return (
    <section className="job-progress-section">
      <div className="section-heading">
        <h3>Final Composition</h3>
        <p>Your finished video will appear here once all scenes are assembled.</p>
      </div>
      <div className="final-preview-card">
        {finalVideoUrl ? (
          <video
            controls
            poster={thumbnailUrl || undefined}
            className="final-preview-video"
          >
            <source src={finalVideoUrl} type="video/mp4" />
            Your browser does not support videos.
          </video>
        ) : (
          <div className="final-preview-placeholder">
            <span>{status === "completed" ? "Preparing download…" : "Composing final cut…"}</span>
          </div>
        )}
      </div>
    </section>
  );
}

FinalPreview.propTypes = {
  finalVideoUrl: PropTypes.string,
  thumbnailUrl: PropTypes.string,
  status: PropTypes.string,
};

function JobProgressPreview({
  job,
  progress,
  percentage,
  stageTimeline,
  estimatedTimeRemaining,
  onRefresh // Added to trigger manual refreshes after intervention
}) {
  const [editingScene, setEditingScene] = useState(null);
  const [isParameterEditorOpen, setIsParameterEditorOpen] = useState(false);
  const { 
    isUpdating, 
    pauseJob, 
    resumeJob, 
    modifyScene, 
    skipScene, 
    updateParameters 
  } = useJobIntervention(job?.job_id, onRefresh);

  const prompt = job?.prompt || progress?.prompt || "Untitled video";
  const status = progress?.status || job?.status || "processing";
  const isPaused = job?.is_paused || status === "paused";
  
  const scenesCompleted =
    progress?.scenes_completed ??
    job?.scenes_completed ??
    progress?.assets?.scenes?.filter((scene) => scene.url)?.length ??
    0;

  const jobScenes =
    progress?.assets?.script?.scenes ||
    job?.scenes ||
    progress?.scenes ||
    EMPTY_ARRAY;

  const totalScenes =
    jobScenes.length ||
    progress?.scene_total ||
    (job?.scene_video_urls ? job.scene_video_urls.length : 0);

  const narratorAudioUrl =
    progress?.assets?.narrator_audio?.url ||
    progress?.assets?.narrator?.url ||
    job?.narrator_audio_url ||
    job?.narrator_audio;

  const backgroundAudioUrl =
    progress?.assets?.background_music?.url ||
    progress?.assets?.audio?.url ||
    job?.audio_url;

  const finalVideoUrl =
    progress?.assets?.final_video?.url ||
    job?.video_url;

  const thumbnailUrl =
    progress?.assets?.thumbnail?.url ||
    job?.thumbnail_url;

  const sceneAssetMap = new Map();
  (progress?.assets?.scenes || []).forEach((sceneAsset) => {
    if (sceneAsset?.scene_number) {
      sceneAssetMap.set(sceneAsset.scene_number, {
        url: sceneAsset.url,
        thumbnail: sceneAsset.thumbnail_url || sceneAsset.thumbnail,
      });
    }
  });

  const gridScenes = jobScenes.length ? jobScenes : (job?.scenes || []).map((scene, index) => ({
    ...scene,
    scene_number: scene.scene_number || scene.SceneNumber || index + 1,
  }));

  const handleEditScene = (scene) => {
    setEditingScene(scene);
  };

  const handleSaveScene = async (data) => {
    if (editingScene) {
      await modifyScene(editingScene.scene_number, data);
      setEditingScene(null);
    }
  };

  const handleSkipScene = async (sceneNumber) => {
    await skipScene(sceneNumber);
    setEditingScene(null); // Close editor if skipping currently edited scene
  };

  const handleSaveParameters = async (params) => {
    await updateParameters(params);
    setIsParameterEditorOpen(false);
  };

  const isJobActive = status === "processing" || status === "paused";

  return (
    <div className="job-progress-preview">
      <header className="job-progress-header">
        <div>
          <p className="eyebrow-text">Job status</p>
          <h2>{prompt}</h2>
          <span className={`job-status-badge status-${status.toLowerCase()}`}>
            {status.replace(/_/g, " ")}
          </span>
        </div>
        
        {/* Intervention Controls */}
        {isJobActive && (
          <JobControls 
            isPaused={isPaused}
            onPause={pauseJob}
            onResume={resumeJob}
            onEditParameters={() => setIsParameterEditorOpen(true)}
            disabled={isUpdating}
          />
        )}

        <div className="progress-metrics">
          <div>
            <p className="metric-label">Overall progress</p>
            <p className="metric-value">{Math.round(percentage)}%</p>
          </div>
          <div>
            <p className="metric-label">Scenes</p>
            <p className="metric-value">
              {scenesCompleted}/{totalScenes || "—"}
            </p>
          </div>
          <div>
            <p className="metric-label">ETA</p>
            <p className="metric-value">
              {estimatedTimeRemaining != null
                ? formatDuration(estimatedTimeRemaining)
                : "Estimating…"}
            </p>
          </div>
        </div>
      </header>

      <div className="progress-bar-wrapper" aria-label="Overall progress">
        <div className="progress-bar-track">
          <div
            className="progress-bar-fill"
            style={{ 
              width: `${Math.min(100, Math.max(0, percentage))}%`,
              opacity: isPaused ? 0.6 : 1
            }}
          />
        </div>
      </div>

      <section className="job-progress-section">
        <div className="section-heading">
          <h3>Production Timeline</h3>
          <p>Track each stage of the pipeline in real time.</p>
        </div>
        <div className="progress-stage-list">
          {stageTimeline.map((stage) => (
            <ProgressStage
              key={stage.key}
              label={stage.label}
              status={stage.status}
              description={
                stage.status === "completed"
                  ? "Done"
                  : stage.status === "active"
                  ? (isPaused ? "Paused" : "In progress")
                  : "Queued"
              }
            />
          ))}
        </div>
      </section>

      <section className="job-progress-section">
        <div className="section-heading">
          <h3>Storyboard</h3>
          <p>Review the scripted narration and visual guidance powering your video.</p>
        </div>
        <ScriptPreview scenes={gridScenes} />
      </section>

      <AudioPreview
        narratorUrl={narratorAudioUrl}
        backgroundUrl={backgroundAudioUrl}
        isGeneratingNarrator={!narratorAudioUrl && status !== "completed"}
        isGeneratingMusic={!backgroundAudioUrl && status !== "completed"}
      />

      <section className="job-progress-section">
        <div className="section-heading">
          <h3>Scene Previews</h3>
          <p>Scenes become playable the moment each render finishes.</p>
        </div>
        
        {editingScene && (
          <SceneEditor 
            scene={editingScene}
            onSave={handleSaveScene}
            onCancel={() => setEditingScene(null)}
            onSkip={handleSkipScene}
            isSaving={isUpdating}
          />
        )}

        <div className="scene-preview-grid">
          {gridScenes.length ? (
            gridScenes.map((scene, index) => {
              const sceneNumber = scene.scene_number || scene.SceneNumber || index + 1;
              const clipInfo =
                sceneAssetMap.get(sceneNumber) ||
                (job?.scene_video_urls?.[sceneNumber - 1]
                  ? { url: job.scene_video_urls[sceneNumber - 1] }
                  : null);

              let statusForScene = "pending";
              if (sceneNumber <= scenesCompleted) {
                statusForScene = clipInfo?.url ? "completed" : "active";
              } else if (sceneNumber === scenesCompleted + 1) {
                statusForScene = "active";
              }

              // Determine if scene is editable (only upcoming or active scenes when paused)
              const isEditable = isJobActive && !clipInfo?.url && (statusForScene === "pending" || (statusForScene === "active" && isPaused));

              return (
                <div key={sceneNumber} className="scene-preview-wrapper">
                  <ScenePreview
                    sceneNumber={sceneNumber}
                    title={scene.location || scene.Location || scene.title}
                    description={scene.action || scene.Action || scene.generation_prompt}
                    videoUrl={clipInfo?.url}
                    thumbnailUrl={clipInfo?.thumbnail}
                    duration={scene.duration || scene.Duration}
                    status={statusForScene}
                  />
                  {isEditable && !editingScene && (
                    <button 
                      className="btn-secondary btn-sm scene-edit-btn"
                      onClick={() => handleEditScene(scene)}
                      disabled={isUpdating}
                    >
                      Edit Scene
                    </button>
                  )}
                </div>
              );
            })
          ) : (
            <div className="scene-preview-empty">
              <p>Scenes will appear here once the script is finalized.</p>
            </div>
          )}
        </div>
      </section>

      <FinalPreview
        finalVideoUrl={finalVideoUrl}
        thumbnailUrl={thumbnailUrl}
        status={status}
      />

      {isParameterEditorOpen && (
        <ParameterEditor 
          currentParams={{
            aspectRatio: job?.aspect_ratio,
            style: job?.style
          }}
          onSave={handleSaveParameters}
          onCancel={() => setIsParameterEditorOpen(false)}
          isSaving={isUpdating}
        />
      )}
    </div>
  );
}

JobProgressPreview.propTypes = {
  job: PropTypes.object,
  progress: PropTypes.object,
  percentage: PropTypes.number,
  stageTimeline: PropTypes.arrayOf(
    PropTypes.shape({
      key: PropTypes.string,
      label: PropTypes.string,
      status: PropTypes.oneOf(["pending", "active", "completed"]),
    })
  ),
  estimatedTimeRemaining: PropTypes.number,
  onRefresh: PropTypes.func,
};

JobProgressPreview.defaultProps = {
  job: null,
  progress: null,
  percentage: 0,
  stageTimeline: DEFAULT_STAGE_ORDER.map((stage) => ({ ...stage, status: "pending" })),
  estimatedTimeRemaining: null,
  onRefresh: () => {},
};

export default JobProgressPreview;
