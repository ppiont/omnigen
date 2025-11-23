import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { FileText, Info, Film } from "lucide-react";
import VideoPlayer from "../components/workspace/VideoPlayer";
import VideoMetadata from "../components/workspace/VideoMetadata";
import ScriptEditor from "../components/workspace/ScriptEditor";
import ScenePanel from "../components/workspace/ScenePanel";
import Timeline from "../components/workspace/Timeline";
import ActionsToolbar from "../components/workspace/ActionsToolbar";
import { jobs } from "../utils/api";
import { addRecentlyOpenedVideo } from "../utils/recentVideos";
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
  const [activeSidebarTab, setActiveSidebarTab] = useState("metadata");
  const [script, setScript] = useState("");
  const scriptJobIdRef = useRef(null);
  const pollingIntervalRef = useRef(null);
  const pollingTimeoutRef = useRef(null);
  const retryTimeoutRef = useRef(null);
  const linkRefreshAttemptsRef = useRef(0);

  /**
   * Flag used to disable UUID validation during early development. Set to
   * false when enforcing strict routing requirements.
   */
  const SKIP_UUID_VALIDATION = false;

  const normalizeStatus = (status) => (status || "").toLowerCase();
  const isProcessingStatus = (status) => {
    const normalized = normalizeStatus(status);
    return normalized === "processing" || normalized === "pending";
  };
  const isCompletedStatus = (status) => {
    const normalized = normalizeStatus(status);
    return normalized === "completed" || normalized === "complete";
  };
  const isTerminalStatus = (status) => {
    const normalized = normalizeStatus(status);
    return (
      normalized === "failed" ||
      normalized === "cancelled" ||
      normalized === "canceled" ||
      isCompletedStatus(normalized)
    );
  };

  /**
   * Validates whether the provided ID string is a UUID or job-{UUID} format.
   *
   * @param {string} id - Identifier to validate
   * @returns {boolean} True when the id matches UUID or job-{UUID} formatting
   */
  const isValidUUID = (id) => {
    if (!id) return false;

    // Check if it's a job-{uuid} format (backend format)
    if (id.startsWith("job-")) {
      const uuidPart = id.substring(4); // Remove 'job-' prefix
      const uuidRegex =
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
      return uuidRegex.test(uuidPart);
    }

    // Check if it's a plain UUID
    const uuidRegex =
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    return uuidRegex.test(id);
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

      console.log(
        `[WORKSPACE] 📥 Fetching job: ${id} (showLoader: ${showLoader})`
      );
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

        // Track this video as recently opened
        if (data.job_id) {
          addRecentlyOpenedVideo(data.job_id);
        }

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
        console.log(
          `[WORKSPACE] 🔄 Polling job status (poll #${pollCount})...`
        );
        const updatedJob = await fetchJob(videoId, { showLoader: false });
        if (isTerminalStatus(updatedJob?.status)) {
          console.log(
            `[WORKSPACE] ✅ Polling complete. Final status: ${updatedJob.status}`
          );
          clearPolling();
        }
      } catch (error) {
        console.warn(`[WORKSPACE] ⚠️ Poll error (poll #${pollCount}):`, error);
        // Swallow transient errors; fetchJob manages error UI state
      }
    };

    console.log("[WORKSPACE] ⏰ Setting up polling interval (every 2 seconds)");
    pollingIntervalRef.current = setInterval(pollOnce, 2000);

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
   * @param {string} [format='mp4'] - Download format ('mp4' or 'webm')
   */
  const handleDownload = (job = jobData, format = "mp4") => {
    console.log("[WORKSPACE] ⬇️ Download requested for job:", job?.job_id, "format:", format);

    const videoUrl = format === "webm" && job?.webm_video_url
      ? job.webm_video_url
      : job?.video_url;

    if (!videoUrl) {
      console.warn("[WORKSPACE] ⚠️ Cannot download: No video URL for format:", format);
      setErrorState({
        type: "download_unavailable",
        title: "Video Not Ready",
        message: format === "webm"
          ? "WebM format is not available for this video."
          : "Video is not ready for download yet.",
        action: "retry",
      });
      return;
    }

    console.log("[WORKSPACE] ⬇️ Starting download:", videoUrl);
    const link = document.createElement("a");
    link.href = videoUrl;
    link.download = `video-${job.job_id}.${format}`;
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

  /**
   * Handles script changes from the script editor.
   *
   * @param {string} newScript - Updated script text
   */
  const handleScriptChange = useCallback((newScript) => {
    setScript(newScript);
    // TODO: Save script to backend or local storage
  }, []);

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
    if (isProcessingStatus(jobData?.status)) {
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
    if (!jobData?.job_id) {
      return;
    }

    if (scriptJobIdRef.current !== jobData.job_id) {
      scriptJobIdRef.current = jobData.job_id;
      setScript(jobData.prompt || "");
      return;
    }

    if (!script && jobData.prompt) {
      setScript(jobData.prompt);
    }
  }, [jobData?.job_id, jobData?.prompt, script]);

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

  if (!videoId || (!SKIP_UUID_VALIDATION && !isValidUUID(videoId))) {
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

  if (isProcessingStatus(jobData?.status)) {
    return renderProcessingView();
  }

  const videoPlayerKey = jobData.video_url
    ? `${jobData.job_id}-${jobData.video_url}`
    : jobData.job_id;

  const sidebarTabs = [
    {
      id: "metadata",
      label: "Metadata",
      icon: <Info size={20} />,
    },
    {
      id: "scenes",
      label: "Scenes",
      icon: <Film size={20} />,
    },
    {
      id: "script",
      label: "Script",
      icon: <FileText size={20} />,
    },
  ];

  let scenes =
    jobData.scenes ||
    jobData.metadata?.scenes ||
    jobData.metadata?.script?.scenes ||
    [];

  if (
    scenes.length === 0 &&
    Array.isArray(jobData.scene_video_urls) &&
    jobData.scene_video_urls.length > 0
  ) {
    const numScenes = jobData.scene_video_urls.length;
    const totalDuration = jobData.duration || 30;
    const sceneDuration = totalDuration / numScenes;

    scenes = jobData.scene_video_urls.map((url, index) => ({
      scene_number: index + 1,
      start_time: index * sceneDuration,
      duration: sceneDuration,
      location: `Scene ${index + 1}`,
      action: `Scene ${index + 1} video clip`,
    }));

    console.log("[WORKSPACE] Created scenes from scene_video_urls:", scenes);
  }

  let audioSpec =
    jobData.audio_spec ||
    jobData.metadata?.audio_spec ||
    jobData.metadata?.script?.audio_spec ||
    null;

  if (!audioSpec && jobData.audio_url) {
    audioSpec = {
      enable_audio: true,
      music_mood: "background",
      music_style: "background",
    };
    console.log("[WORKSPACE] Created audio spec from audio_url");
  }

  const backgroundMusicUrl =
    jobData.audio_url || jobData.metadata?.audio_url || null;
  const narratorAudioUrl =
    jobData.narrator_audio_url ||
    jobData.metadata?.narrator_audio_url ||
    audioSpec?.narrator_audio_url ||
    null;
  const sideEffectsText =
    jobData.side_effects_text ||
    jobData.side_effects ||
    audioSpec?.side_effects_text ||
    jobData.metadata?.audio_spec?.side_effects_text ||
    null;
  const sideEffectsStartTime =
    jobData.side_effects_start_time ??
    audioSpec?.side_effects_start_time ??
    jobData.metadata?.audio_spec?.side_effects_start_time ??
    null;

  console.log("[WORKSPACE] Timeline data:", {
    videoDuration: jobData.duration || 30,
    scenesCount: scenes.length,
    scenes,
    audioSpec,
    backgroundMusicUrl,
    narratorAudioUrl,
    hasSideEffectsText: Boolean(sideEffectsText),
    sideEffectsStartTime,
    sceneVideoURLs: jobData.scene_video_urls,
  });

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

        <div className="workspace-body">
          <section className="workspace-main-column">
            <div className="workspace-player-card">
              <VideoPlayer
                key={videoPlayerKey}
                videoUrl={jobData.video_url}
                status={jobData.status}
                aspectRatio={jobData.aspect_ratio || "16:9"}
                onRefresh={handleVideoRefresh}
                backgroundMusicUrl={backgroundMusicUrl}
                narratorAudioUrl={narratorAudioUrl}
              />
            </div>
          </section>

          <aside className="workspace-inspector" aria-label="Video details">
            <div
              className="workspace-inspector-tabs"
              role="tablist"
              aria-orientation="vertical"
            >
              {sidebarTabs.map((tab) => {
                const isActive = activeSidebarTab === tab.id;
                return (
                  <button
                    key={tab.id}
                    type="button"
                    className={`workspace-inspector-tab ${
                      isActive ? "active" : ""
                    }`}
                    onClick={() => setActiveSidebarTab(tab.id)}
                    role="tab"
                    aria-selected={isActive}
                    aria-controls={`workspace-inspector-panel-${tab.id}`}
                    tabIndex={isActive ? 0 : -1}
                  >
                    {tab.icon}
                    <span>{tab.label}</span>
                  </button>
                );
              })}
            </div>

            <div
              className="workspace-inspector-panel"
              role="tabpanel"
              id={`workspace-inspector-panel-${activeSidebarTab}`}
            >
              {activeSidebarTab === "metadata" && (
                <VideoMetadata key={jobData.job_id} jobData={jobData} />
              )}
              {activeSidebarTab === "scenes" && (
                <ScenePanel
                  jobId={jobData.job_id}
                  scenes={scenes}
                  sceneVideoUrls={jobData.scene_video_urls}
                  sceneVersions={jobData.scene_versions}
                  onSceneRegenerated={() => fetchJob(jobData.job_id)}
                  disabled={!isCompletedStatus(jobData.status)}
                />
              )}
              {activeSidebarTab === "script" && (
                <ScriptEditor script={script} onChange={handleScriptChange} />
              )}
            </div>
          </aside>
        </div>

        {/* Timeline Editor - Full Width at Bottom */}
        <div className="workspace-timeline-section">
          <div className="workspace-timeline-card">
            <Timeline
              videoDuration={jobData.duration || 30}
              scenes={scenes}
              audioSpec={audioSpec}
              backgroundMusicUrl={backgroundMusicUrl}
              audioUrl={backgroundMusicUrl}
              narratorAudioUrl={narratorAudioUrl}
              sideEffectsText={sideEffectsText}
              sideEffectsStartTime={sideEffectsStartTime}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export default Workspace;
