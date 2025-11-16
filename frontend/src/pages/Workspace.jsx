import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import VideoPlayer from "../components/workspace/VideoPlayer";
import ChatInterface from "../components/workspace/ChatInterface";
import VideoMetadata from "../components/workspace/VideoMetadata";
import ActionsToolbar from "../components/workspace/ActionsToolbar";
import { jobs } from "../utils/api";
import "../styles/workspace.css";

/**
 * Workspace orchestrates fetching, polling, and rendering for the video
 * workspace experience.
 *
 * @returns {JSX.Element} Workspace page content
 */
function Workspace() {
  const { videoId } = useParams();
  const navigate = useNavigate();

  const [jobData, setJobData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [errorState, setErrorState] = useState(null);
  const [rateLimitCountdown, setRateLimitCountdown] = useState(null);
  const pollingIntervalRef = useRef(null);
  const pollingTimeoutRef = useRef(null);
  const retryTimeoutRef = useRef(null);
  const linkRefreshAttemptsRef = useRef(0);

  /**
   * Flag used to disable UUID validation during early development. Set to
   * false when enforcing strict routing requirements.
   */
  const SKIP_UUID_VALIDATION = false;
  const LOCAL_DEV_MODE = import.meta.env.VITE_LOCAL_DEV_MODE === "true";
  const USE_MOCK_WORKSPACE_ON_AUTH_ERROR = LOCAL_DEV_MODE;
  const MOCK_WORKSPACE_JOB = {
    job_id: "job-local-dev-placeholder",
    title: "Sample Video Workspace",
    prompt: "Product showcase video for wireless headphones with modern aesthetic",
    status: "completed",
    duration: 30,
    aspect_ratio: "16:9",
    created_at: Math.floor(Date.now() / 1000),
    completed_at: Math.floor(Date.now() / 1000),
    video_url: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4",
    style: "Cinematic",
    error_message: null,
    progress_percentage: 100,
  };

  /**
   * Validates whether the provided ID string is a valid job ID (job-<uuid> format).
   *
   * @param {string} id - Identifier to validate
   * @returns {boolean} True when the id matches job ID formatting
   */
  const isValidJobID = (id) => {
    const jobIDRegex =
      /^job-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return jobIDRegex.test(id);
  };

  /**
   * Clears active polling intervals/timeouts.
   */
  const clearPolling = () => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
      pollingIntervalRef.current = null;
    }
    if (pollingTimeoutRef.current) {
      clearTimeout(pollingTimeoutRef.current);
      pollingTimeoutRef.current = null;
    }
  };

  /**
   * Clears the queued retry timeout, if present.
   */
  const clearRetryTimeout = () => {
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }
  };

  /**
   * Determines whether the given API error signals an expired video URL.
   *
   * @param {import("../utils/api").APIError|Error} apiError - Error details
   * @returns {boolean} True if the error points to an expired URL
   */
  const isUrlExpiredError = (apiError) => {
    if (!apiError) {
      return false;
    }
    const code = (apiError.code || "").toString().toLowerCase();
    const message = (apiError.message || "").toLowerCase();
    const detailsReason = apiError.details?.reason?.toLowerCase();
    return (
      code.includes("expired") ||
      message.includes("expired") ||
      detailsReason === "url_expired"
    );
  };

  /**
   * Fetches job data and handles all associated error states.
   *
   * @param {string} id - Job identifier
   * @param {{showLoader?: boolean}} [options] - Loader configuration
   * @returns {Promise<Object|undefined>} Job data when successful
   */
  const fetchJob = useCallback(
    async (id, { showLoader = true } = {}) => {
      if (!id) {
        console.warn("[WORKSPACE] ⚠️ fetchJob called without ID");
        return;
      }

      console.log(`[WORKSPACE] 📥 Fetching job: ${id} (showLoader: ${showLoader})`);
      clearRetryTimeout();

      try {
        if (showLoader) {
          setLoading(true);
        }

        console.log("[WORKSPACE] 📡 Calling GET /api/v1/jobs/" + id);
        const data = await jobs.get(id);
        console.log("[WORKSPACE] ✅ Job data received:", {
          job_id: data.job_id,
          status: data.status,
          has_video_url: !!data.video_url,
        });
        setJobData(data);
        setRateLimitCountdown(null);
        linkRefreshAttemptsRef.current = 0;

        if (data.status === "failed") {
          console.error("[WORKSPACE] ❌ Job failed:", data.error_message);
          setErrorState({
            type: "job_failed",
            title: "Video Generation Failed",
            message: data.error_message
              ? `Video generation failed: ${data.error_message}`
              : "Video generation failed. Please try again from the Create page.",
            action: "try_again",
          });
        } else {
          console.log("[WORKSPACE] ✅ Job status:", data.status);
          setErrorState(null);
        }

        return data;
      } catch (err) {
        console.error("[WORKSPACE] ❌ Error fetching job:", {
          status: err.status,
          code: err.code,
          message: err.message,
        });
        
        if (err.status === 404) {
          console.error("[WORKSPACE] ❌ Job not found (404)");
          setErrorState({
            type: "not_found",
            title: "Video Not Found",
            message: "This video doesn't exist or has been deleted.",
            action: "return_to_library",
          });
          return;
        }

        if (err.status === 401) {
          if (USE_MOCK_WORKSPACE_ON_AUTH_ERROR) {
            console.warn("[WORKSPACE] ⚠️ Auth error. Loading mock workspace job for local dev.");
            setJobData(MOCK_WORKSPACE_JOB);
            setErrorState(null);
            setLoading(false);
            return MOCK_WORKSPACE_JOB;
          }

          setErrorState({
            type: "unauthorized",
            title: "Session Expired",
            message: "Your session has expired. Please log in again.",
            action: "redirect_to_login",
          });
          setTimeout(() => navigate("/login"), 1500);
          return;
        }

        if (err.status === 429) {
          const retryAfter =
            Number(err.details?.retry_after) ||
            Number(err.details?.["Retry-After"]) ||
            60;
          setRateLimitCountdown(retryAfter);
          setErrorState({
            type: "rate_limited",
            title: "Rate Limit Reached",
            message: `Too many requests. Please wait ${retryAfter} seconds.`,
            action: "countdown",
            retryAfter,
          });
          retryTimeoutRef.current = setTimeout(
            () => fetchJob(id, { showLoader: false }),
            retryAfter * 1000
          );
          return;
        }

        if (err.status === 403) {
          if (isUrlExpiredError(err)) {
            setErrorState({
              type: "url_expired",
              title: "Video Link Expired",
              message: "Video link expired. Refreshing...",
              action: "refreshing",
            });

            if (linkRefreshAttemptsRef.current < 3) {
              linkRefreshAttemptsRef.current += 1;
              retryTimeoutRef.current = setTimeout(
                () => fetchJob(id, { showLoader: false }),
                1500
              );
            } else {
              setErrorState({
                type: "forbidden",
                title: "Access Restricted",
                message:
                  "We couldn't refresh the video link. Please try again later.",
                action: "retry",
              });
            }
          } else {
            setErrorState({
              type: "forbidden",
              title: "Access Restricted",
              message: "You do not have permission to view this video.",
              action: "return_to_library",
            });
          }
          return;
        }

        if (err.status === 0) {
          setErrorState({
            type: "network_error",
            title: "Connection Lost",
            message: "Connection lost. Check your internet.",
            action: "auto_retry",
          });
          retryTimeoutRef.current = setTimeout(
            () => fetchJob(id, { showLoader: false }),
            3000
          );
          return;
        }

        setErrorState({
          type: "server_error",
          title: "Server Error",
          message: "Server error. Please try again.",
          action: "retry",
        });
      } finally {
        if (showLoader) {
          setLoading(false);
        }
      }
    },
    [navigate]
  );

  /**
   * Begins polling for job status updates until the job finishes.
   */
  const beginPolling = useCallback(() => {
    if (!videoId) {
      console.warn("[WORKSPACE] ⚠️ Cannot start polling: No video ID");
      return;
    }
    
    console.log("[WORKSPACE] 🔄 Starting polling for job:", videoId);
    clearPolling();

    let pollCount = 0;
    const pollOnce = async () => {
      pollCount++;
      try {
        console.log(`[WORKSPACE] 🔄 Polling job status (poll #${pollCount})...`);
        const updatedJob = await fetchJob(videoId, { showLoader: false });
        if (
          updatedJob?.status === "completed" ||
          updatedJob?.status === "failed"
        ) {
          console.log(`[WORKSPACE] ✅ Polling complete. Final status: ${updatedJob.status}`);
          clearPolling();
        }
      } catch (error) {
        console.warn(`[WORKSPACE] ⚠️ Poll error (poll #${pollCount}):`, error);
        // Swallow transient errors; fetchJob manages error UI state
      }
    };

    console.log("[WORKSPACE] ⏰ Setting up polling interval (every 5 seconds)");
    pollingIntervalRef.current = setInterval(pollOnce, 5000);

    console.log("[WORKSPACE] ⏰ Setting polling timeout (10 minutes)");
    pollingTimeoutRef.current = setTimeout(() => {
      console.warn("[WORKSPACE] ⏰ Polling timeout reached (10 minutes)");
      clearPolling();
      setErrorState({
        type: "timeout",
        title: "Still Processing",
        message: "Generation is taking longer than expected...",
        action: "retry",
      });
    }, 600000);
  }, [fetchJob, videoId]);

  /**
   * Triggers a browser download for the provided job's video file.
   *
   * @param {Object} [job=jobData] - Job data containing video_url
   */
  const handleDownload = (job = jobData) => {
    console.log("[WORKSPACE] ⬇️ Download requested for job:", job?.job_id);
    
    if (!job?.video_url) {
      console.warn("[WORKSPACE] ⚠️ Cannot download: No video URL");
      setErrorState({
        type: "download_unavailable",
        title: "Video Not Ready",
        message: "Video is not ready for download yet.",
        action: "retry",
      });
      return;
    }

    console.log("[WORKSPACE] ⬇️ Starting download:", job.video_url);
    const link = document.createElement("a");
    link.href = job.video_url;
    link.download = `video-${job.job_id}.mp4`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    console.log("[WORKSPACE] ✅ Download initiated");
  };

  /**
   * Placeholder delete handler that currently navigates back to the library.
   *
   * @param {Object} [job=jobData] - Job data targeted for deletion
   */
  const handleDelete = async (job = jobData) => {
    if (!job) {
      return;
    }
    // Placeholder for future delete API integration
    navigate("/library");
  };

  /**
   * Refreshes job data without showing the loader (used for video refreshes).
   */
  const handleVideoRefresh = () => {
    console.log("[WORKSPACE] 🔄 Manual refresh requested");
    fetchJob(videoId, { showLoader: false });
  };

  useEffect(() => {
    console.log("=".repeat(60));
    console.log("[WORKSPACE] 🏗️ Workspace component mounted/updated");
    console.log("[WORKSPACE] Video ID:", videoId);
    
    if (videoId) {
      console.log("[WORKSPACE] 📥 Fetching job data for video:", videoId);
      fetchJob(videoId);
    } else {
      console.warn("[WORKSPACE] ⚠️ No video ID provided");
    }

    return () => {
      console.log("[WORKSPACE] 🧹 Cleaning up polling and timeouts");
      clearPolling();
      clearRetryTimeout();
    };
  }, [fetchJob, videoId]);

  useEffect(() => {
    if (jobData?.status === "processing" || jobData?.status === "pending") {
      console.log(`[WORKSPACE] 🔄 Job is ${jobData.status}, starting polling`);
      beginPolling();
      return () => {
        console.log("[WORKSPACE] 🧹 Stopping polling (job status changed)");
        clearPolling();
      };
    }

    if (jobData?.status) {
      console.log(`[WORKSPACE] ✅ Job status: ${jobData.status} (not polling)`);
    }
    clearPolling();
  }, [beginPolling, jobData?.status]);

  useEffect(() => {
    let countdownInterval;
    if (rateLimitCountdown !== null) {
      countdownInterval = setInterval(() => {
        setRateLimitCountdown((prev) => {
          if (prev === null || prev <= 1) {
            clearInterval(countdownInterval);
            return null;
          }
          return prev - 1;
        });
      }, 1000);
    }

    return () => {
      if (countdownInterval) {
        clearInterval(countdownInterval);
      }
    };
  }, [rateLimitCountdown]);

  /**
   * Routes the user back to the library page.
   */
  const goToLibrary = () => navigate("/library");

  /**
   * Retries fetching job data immediately.
   */
  const handleRetry = () => {
    if (!videoId) {
      console.warn("[WORKSPACE] ⚠️ Cannot retry: No video ID");
      return;
    }
    console.log("[WORKSPACE] 🔄 Retry requested");
    fetchJob(videoId);
  };

  /**
   * Navigates to the Create page, pre-filling the last prompt when available.
   */
  const handleTryAgain = () => {
    if (jobData?.prompt) {
      navigate("/create", { state: { prompt: jobData.prompt } });
      return;
    }
    navigate("/create");
  };

  /**
   * Renders a fully-styled error view for the workspace.
   *
   * @param {Object|null} overrideState - Optional override for error state
   * @returns {JSX.Element|null} Error UI block, if state is present
   */
  const renderErrorView = (overrideState = null) => {
    const state = overrideState || errorState;
    if (!state) {
      return null;
    }

    const countdownValue =
      state.type === "rate_limited" && rateLimitCountdown !== null
        ? rateLimitCountdown
        : state.retryAfter ?? null;

    const displayMessage =
      state.type === "rate_limited" && countdownValue !== null
        ? `Too many requests. Please wait ${countdownValue} seconds.`
        : state.message || "";

    const iconMap = {
      not_found: "🗂",
      unauthorized: "🔒",
      rate_limited: "⏳",
      server_error: "⚠️",
      network_error: "📡",
      job_failed: "🚫",
      timeout: "⌛",
      url_expired: "🔁",
      download_unavailable: "⬇️",
      forbidden: "🚫",
      default: "⚠️",
    };

    const actionButtons = (() => {
      switch (state.action) {
        case "return_to_library":
          return [
            <Link
              to="/library"
              className="btn-return-library"
              onClick={goToLibrary}
              key="return-library"
            >
              ← Return to Library
            </Link>,
          ];
        case "redirect_to_login":
          return [
            <p className="error-info" key="redirect">
              Redirecting to login...
            </p>,
          ];
        case "countdown":
          return [
            <div className="countdown-timer" key="countdown">
              Retrying in {countdownValue ?? "..."}s
            </div>,
            <button
              type="button"
              className="btn-retry"
              onClick={handleRetry}
              key="retry-now"
            >
              Retry Now
            </button>,
          ];
        case "retry":
          return [
            <button
              type="button"
              className="btn-retry"
              onClick={handleRetry}
              key="retry"
            >
              Retry
            </button>,
          ];
        case "auto_retry":
          return [
            <p className="error-info" key="auto-msg">
              Retrying automatically...
            </p>,
            <button
              type="button"
              className="btn-retry"
              onClick={handleRetry}
              key="retry-now"
            >
              Retry Now
            </button>,
          ];
        case "try_again":
          return [
            <button
              type="button"
              className="btn-retry"
              onClick={handleTryAgain}
              key="try-again"
            >
              Try Again
            </button>,
            <Link
              to="/library"
              className="btn-return-library"
              onClick={goToLibrary}
              key="return-library"
            >
              ← Return to Library
            </Link>,
          ];
        case "refreshing":
          return [
            <p className="error-info" key="refreshing">
              Refreshing video link...
            </p>,
          ];
        default:
          return [
            <Link
              to="/library"
              className="btn-return-library"
              onClick={goToLibrary}
              key="return-library"
            >
              ← Return to Library
            </Link>,
          ];
      }
    })();

    return (
      <div className="workspace-page">
        <div className="workspace-content loaded">
          <section
            className="workspace-error fade-in"
            role="alert"
            aria-live="assertive"
          >
            <div className="error-icon" aria-hidden="true">
              {iconMap[state.type] || iconMap.default}
            </div>
            <p className="error-title">{state.title}</p>
            <p className="error-message">{displayMessage}</p>
            {state.type === "rate_limited" && countdownValue !== null && (
              <p className="error-submessage">
                We’ll retry automatically, or refresh now.
              </p>
            )}
            <div className="error-actions">{actionButtons}</div>
          </section>
        </div>
      </div>
    );
  };

  /**
   * Renders the processing state while the video is still generating.
   *
   * @returns {JSX.Element} Processing UI
   */
  const renderProcessingView = () => {
    const rawProgress =
      jobData?.progress_percentage ??
      jobData?.progress_percent ??
      jobData?.progress ??
      null;
    const progressNumber =
      typeof rawProgress === "number"
        ? rawProgress
        : typeof rawProgress === "string"
        ? Number(rawProgress)
        : null;
    const hasProgress = Number.isFinite(progressNumber);
    const clampedProgress = hasProgress
      ? Math.min(Math.max(progressNumber, 0), 100)
      : null;

    return (
      <div className="workspace-page">
        <div className="workspace-content loaded">
          <section className="workspace-processing fade-in" aria-live="polite">
            <p className="processing-message pulse">
              {hasProgress
                ? `Processing... ${Math.round(clampedProgress)}%`
                : "Processing your video..."}
            </p>
            <div className="processing-progress" aria-hidden="true">
              <div
                className={`processing-progress-bar${
                  hasProgress ? "" : " processing-progress-indeterminate"
                }`}
                style={
                  hasProgress
                    ? { width: `${Math.round(clampedProgress)}%` }
                    : undefined
                }
              />
            </div>
            <p className="processing-status">
              This usually takes 2-5 minutes. Feel free to navigate away—we’ll
              keep checking for you.
            </p>
            <button type="button" className="btn-retry" onClick={handleRetry}>
              Refresh Status
            </button>
          </section>
        </div>
      </div>
    );
  };

  if (!videoId || (!SKIP_UUID_VALIDATION && !isValidJobID(videoId))) {
    return renderErrorView({
      type: "invalid_video_id",
      title: "Invalid Video ID",
      message: "The video ID is invalid or missing.",
      action: "return_to_library",
    });
  }

  if (loading) {
    return (
      <div className="workspace-page">
        <div className="workspace-content loading">
          <section className="workspace-loading fade-in" aria-live="polite">
            <div className="loading-spinner" />
            <p className="loading-message">Loading video workspace...</p>
            <p className="loading-submessage">
              Getting everything ready for you
            </p>
          </section>
        </div>
      </div>
    );
  }

  if (errorState) {
    return renderErrorView();
  }

  if (!jobData) {
    return renderErrorView({
      type: "no_data",
      title: "No Data Available",
      message: "We couldn't load this video. Please try again later.",
      action: "retry",
    });
  }

  if (jobData?.status === "processing" || jobData?.status === "pending") {
    return renderProcessingView();
  }

  const videoPlayerKey = jobData.video_url
    ? `${jobData.job_id}-${jobData.video_url}`
    : jobData.job_id;

  return (
    <div className="workspace-page">
      <div className="workspace-content loaded">
        <nav className="workspace-breadcrumbs" aria-label="Breadcrumb">
          <Link to="/library" className="breadcrumb-link">
            Library
          </Link>
          <span className="breadcrumb-separator"> / </span>
          <span className="breadcrumb-current">Video Editor</span>
        </nav>

        <div className="workspace-top-bar">
          <div>
            <h1 className="workspace-title">
              {jobData?.title || "Video Workspace"}
            </h1>
            <p className="workspace-video-id">
              Video ID: <code>{jobData.job_id}</code>
            </p>
          </div>
          <ActionsToolbar
            jobData={jobData}
            onDownload={handleDownload}
            onDelete={handleDelete}
          />
        </div>

        <main className="workspace-main">
          <VideoPlayer
            key={videoPlayerKey}
            videoUrl={jobData.video_url}
            status={jobData.status}
            aspectRatio={jobData.aspect_ratio || "16:9"}
            onRefresh={handleVideoRefresh}
          />

          <div className="workspace-grid">
            <VideoMetadata key={jobData.job_id} jobData={jobData} />
            <ChatInterface jobData={jobData} />
          </div>
        </main>
      </div>
    </div>
  );
}

export default Workspace;
