import { useEffect, useRef, useState } from "react";
import PropTypes from "prop-types";
import { jobs } from "../../utils/api";

/**
 * RefinementModal displays progress for video refinement generation.
 *
 * @param {{isOpen: boolean, jobId: string, onClose: Function}} props - Component props
 * @returns {JSX.Element|null} Modal component or null if not open
 */
function RefinementModal({ isOpen, jobId, onClose }) {
  const [jobStatus, setJobStatus] = useState(null);
  const [progress, setProgress] = useState(0);
  const [currentStage, setCurrentStage] = useState("");
  const [error, setError] = useState(null);
  const pollingTimeoutRef = useRef(null);

  useEffect(() => {
    if (!isOpen || !jobId) {
      return;
    }

    console.log("[REFINEMENT_MODAL] 🎬 Starting refinement tracking for job:", jobId);
    setProgress(0);
    setCurrentStage("Initializing...");
    setError(null);

    let pollCount = 0;
    const maxPollAttempts = 300; // ~35 minutes max
    const progressPollInterval = 7000; // 7 seconds between polls

    const pollProgress = async () => {
      pollCount++;

      // Check max polling attempts
      if (pollCount > maxPollAttempts) {
        console.error("[REFINEMENT_MODAL] ⚠️ Max polling attempts reached");
        setError("Video generation is taking longer than expected. Please check back later.");
        if (pollingTimeoutRef.current) clearTimeout(pollingTimeoutRef.current);
        return;
      }

      try {
        console.log(`[REFINEMENT_MODAL] 🔄 Polling job status (poll #${pollCount})...`);

        const job = await jobs.get(jobId);

        console.log("[REFINEMENT_MODAL] 📊 Job status update:", {
          status: job.status,
          stage: job.stage,
          progress_percent: job.progress_percent,
        });

        setJobStatus(job.status);
        setProgress(job.progress_percent || 0);

        // Update stage description
        if (job.stage) {
          const stageDescriptions = {
            script_generating: "Generating script...",
            script_complete: "Script complete",
            scene_1_generating: "Rendering scene 1...",
            scene_1_complete: "Scene 1 complete",
            scene_2_generating: "Rendering scene 2...",
            scene_2_complete: "Scene 2 complete",
            scene_3_generating: "Rendering scene 3...",
            scene_3_complete: "Scene 3 complete",
            audio_generating: "Generating audio...",
            audio_complete: "Audio complete",
            composing: "Composing final video...",
            complete: "Complete",
            failed: "Failed",
          };
          setCurrentStage(stageDescriptions[job.stage] || job.stage);
        }

        if (job.status === "completed") {
          console.log("[REFINEMENT_MODAL] ✅ Refinement completed!");
          setProgress(100);
          setCurrentStage("Complete!");
          
          // Close modal after a brief delay to show completion
          setTimeout(() => {
            onClose();
          }, 1500);
        } else if (job.status === "failed") {
          console.error("[REFINEMENT_MODAL] ❌ Refinement failed");
          setError(job.error_message || "Video generation failed");
          if (pollingTimeoutRef.current) clearTimeout(pollingTimeoutRef.current);
        } else {
          // Continue polling
          pollingTimeoutRef.current = setTimeout(pollProgress, progressPollInterval);
        }
      } catch (error) {
        // Handle rate limit errors
        if (error.status === 429) {
          const retryAfter = error.details?.reset_in || 60;
          console.warn(
            `[REFINEMENT_MODAL] ⚠️ Rate limit hit. Waiting ${retryAfter} seconds...`
          );
          pollingTimeoutRef.current = setTimeout(
            pollProgress,
            retryAfter * 1000
          );
          return;
        }

        console.error(`[REFINEMENT_MODAL] ⚠️ Error polling progress:`, error);
        // Continue polling on other errors after delay
        pollingTimeoutRef.current = setTimeout(pollProgress, progressPollInterval);
      }
    };

    // Start polling
    pollingTimeoutRef.current = setTimeout(pollProgress, progressPollInterval);

    return () => {
      if (pollingTimeoutRef.current) {
        clearTimeout(pollingTimeoutRef.current);
      }
    };
  }, [isOpen, jobId, onClose]);

  if (!isOpen) {
    return null;
  }

  return (
    <div className="refinement-modal-overlay" onClick={onClose}>
      <div
        className="refinement-modal-content"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="refinement-modal-header">
          <h2>Refining Your Video</h2>
          {error && (
            <button
              type="button"
              className="refinement-modal-close"
              onClick={onClose}
              aria-label="Close modal"
            >
              ×
            </button>
          )}
        </div>

        <div className="refinement-modal-body">
          {error ? (
            <div className="refinement-error">
              <p className="refinement-error-title">Error</p>
              <p className="refinement-error-message">{error}</p>
              <button
                type="button"
                className="btn-retry"
                onClick={onClose}
              >
                Close
              </button>
            </div>
          ) : (
            <>
              <div className="refinement-progress-container">
                <div className="refinement-progress-bar">
                  <div
                    className="refinement-progress-fill"
                    style={{ width: `${Math.min(progress, 100)}%` }}
                  />
                </div>
                <p className="refinement-progress-text">
                  {Math.round(progress)}%
                </p>
              </div>

              <p className="refinement-stage-text">{currentStage}</p>

              <p className="refinement-info-text">
                This usually takes 2-5 minutes. You can continue working while
                your video is being generated.
              </p>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

RefinementModal.propTypes = {
  isOpen: PropTypes.bool.isRequired,
  jobId: PropTypes.string,
  onClose: PropTypes.func.isRequired,
};

export default RefinementModal;

