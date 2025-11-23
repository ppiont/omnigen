import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import PropTypes from "prop-types";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Volume2,
  Maximize,
  Minimize,
} from "lucide-react";

const ASPECT_RATIO_CLASSES = {
  "16:9": "video-aspect-16-9",
  "9:16": "video-aspect-9-16",
  "1:1": "video-aspect-1-1",
};

const STATUS_VARIANTS = [
  "pending",
  "processing",
  "completed",
  "complete",
  "failed",
];

const normalizeStatus = (status) => (status || "").toLowerCase();
const isProcessingStatus = (status) => {
  const normalized = normalizeStatus(status);
  return normalized === "processing" || normalized === "pending";
};
const isCompletedStatus = (status) => {
  const normalized = normalizeStatus(status);
  return normalized === "completed" || normalized === "complete";
};
const isFailedStatus = (status) => normalizeStatus(status) === "failed";

/**
 * VideoPlayer renders the primary workspace video experience, offering custom
 * controls, keyboard navigation, and robust loading/error states.
 */
function VideoPlayer({
  videoUrl,
  status,
  aspectRatio,
  onError,
  onRefresh,
  // DEPRECATED: Audio is now embedded in video. These props are kept for backwards compatibility.
  // eslint-disable-next-line no-unused-vars
  backgroundMusicUrl,
  // eslint-disable-next-line no-unused-vars
  narratorAudioUrl,
}) {
  const videoRef = useRef(null);
  const containerRef = useRef(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [videoError, setVideoError] = useState(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [progressIsActive, setProgressIsActive] = useState(false);
  const [reloadNonce, setReloadNonce] = useState(0);

  const aspectClass = useMemo(
    () => ASPECT_RATIO_CLASSES[aspectRatio] || ASPECT_RATIO_CLASSES["16:9"],
    [aspectRatio]
  );

  const formatTime = useCallback((seconds) => {
    if (!Number.isFinite(seconds)) return "0:00";
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }, []);

  const resetPlaybackState = useCallback(() => {
    setIsPlaying(false);
    setCurrentTime(0);
    setDuration(0);
  }, []);

  const applyVolumeMix = useCallback(() => {
    if (videoRef.current) {
      videoRef.current.volume = volume;
      videoRef.current.muted = false;
    }
  }, [volume]);

  const handleVideoElementPlay = useCallback(() => {
    setIsPlaying(true);
  }, []);

  const handleVideoElementPause = useCallback(() => {
    setIsPlaying(false);
  }, []);

  const handleVideoEnded = useCallback(() => {
    setIsPlaying(false);
  }, []);

  useEffect(() => {
    applyVolumeMix();
  }, [applyVolumeMix]);

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(Boolean(document.fullscreenElement));
    };

    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
    };
  }, []);

  useEffect(() => {
    if (!isCompletedStatus(status) && videoRef.current) {
      videoRef.current.pause();
    }
  }, [status]);

  const handleLoadedMetadata = () => {
    if (!videoRef.current) return;
    setDuration(videoRef.current.duration || 0);
  };

  const handleTimeUpdate = () => {
    if (!videoRef.current) return;
    setCurrentTime(videoRef.current.currentTime);
  };

  const handleVideoError = (event) => {
    const errorCode = event?.target?.error?.code;
    const isExpired = errorCode === 4; // MEDIA_ERR_SRC_NOT_SUPPORTED
    const nextError = {
      type: isExpired ? "url_expired" : "load_failed",
      message: isExpired
        ? "Video link expired. Refreshing..."
        : "Unable to load video. Please try again.",
    };

    setVideoError(nextError);
    setIsPlaying(false);

    if (isExpired && onRefresh) {
      onRefresh("url_expired");
    }

    if (onError) {
      onError(nextError.type);
    }
  };

  const togglePlay = async () => {
    if (!videoRef.current) {
      return;
    }

    if (isPlaying) {
      videoRef.current.pause();
      setIsPlaying(false);
      return;
    }

    try {
      await videoRef.current.play();
      setIsPlaying(true);
    } catch {
      const playbackError = {
        type: "playback_error",
        message: "Unable to start playback. Please try again.",
      };
      setVideoError(playbackError);
      if (onError) onError(playbackError.type);
    }
  };

  const seekTo = (timeInSeconds) => {
    if (!videoRef.current) return;

    const clipped = Math.min(
      Math.max(timeInSeconds, 0),
      videoRef.current.duration || duration || 0
    );

    videoRef.current.currentTime = clipped;
    setCurrentTime(clipped);
  };

  const handleProgressChange = (event) => {
    seekTo(Number(event.target.value));
  };

  const handleProgressInteraction = (nextState) => () => {
    setProgressIsActive(nextState);
  };

  const handleVolumeChange = (event) => {
    const nextVolume = Number(event.target.value);
    setVolume(Number.isFinite(nextVolume) ? nextVolume : 1);
  };

  const seekByOffset = (offset) => {
    if (!videoRef.current) return;
    seekTo(videoRef.current.currentTime + offset);
  };

  const seekToBeginning = () => seekTo(0);
  const seekToEnd = () => seekTo(duration || videoRef.current?.duration || 0);

  const toggleFullscreen = () => {
    if (!containerRef.current) return;
    if (!document.fullscreenElement) {
      containerRef.current.requestFullscreen?.();
    } else {
      document.exitFullscreen?.();
    }
  };

  const handleShortcut = (event) => {
    const { key } = event;

    switch (key) {
      case " ":
      case "Spacebar":
        event.preventDefault();
        togglePlay();
        break;
      case "ArrowLeft":
        event.preventDefault();
        seekByOffset(-5);
        break;
      case "ArrowRight":
        event.preventDefault();
        seekByOffset(5);
        break;
      case "ArrowUp":
        event.preventDefault();
        setVolume((prev) => Math.min(1, Number((prev + 0.1).toFixed(2))));
        break;
      case "ArrowDown":
        event.preventDefault();
        setVolume((prev) => Math.max(0, Number((prev - 0.1).toFixed(2))));
        break;
      case "f":
      case "F":
        event.preventDefault();
        toggleFullscreen();
        break;
      default:
        break;
    }
  };

  const handleVideoErrorRefresh = () => {
    const currentErrorType = videoError?.type;
    resetPlaybackState();
    setVideoError(null);
    setReloadNonce((prev) => prev + 1);

    if (onRefresh) {
      onRefresh(currentErrorType || "video_refresh");
    }
  };

  const progressPercent = duration ? (currentTime / duration) * 100 : 0;

  const renderStateBlock = (variant, message, extraAction = null) => (
    <div className={`video-state-block ${variant} ${aspectClass}`}>
      <div className="video-state-content">
        {variant === "loading" && <div className="loading-spinner" />}
        {variant === "error" && (
          <span className="state-icon" role="img" aria-label="Error">
            ⚠️
          </span>
        )}
        <p>{message}</p>
        {extraAction}
      </div>
    </div>
  );

  if (isProcessingStatus(status)) {
    const normalizedStatus = normalizeStatus(status);
    const loadingMessage =
      normalizedStatus === "processing"
        ? "Processing your video..."
        : "Preparing your video...";
    return (
      <div className="video-player-container">
        {renderStateBlock("loading", loadingMessage)}
      </div>
    );
  }

  if (isFailedStatus(status)) {
    return (
      <div className="video-player-container">
        {renderStateBlock(
          "error",
          "Video generation failed. Please try again from the Create page."
        )}
      </div>
    );
  }

  if (videoError) {
    const actionLabel =
      videoError.type === "url_expired" ? "Refresh Link" : "Refresh";
    return (
      <div className="video-player-container">
        {renderStateBlock(
          "error",
          videoError.message,
          <button
            type="button"
            className="retry-btn"
            onClick={handleVideoErrorRefresh}
          >
            {actionLabel}
          </button>
        )}
      </div>
    );
  }

  if (!videoUrl) {
    return (
      <div className="video-player-container">
        {renderStateBlock("empty", "Video is not available yet.")}
      </div>
    );
  }

  return (
    <section
      className={`video-player-container ${
        isFullscreen ? "is-fullscreen" : ""
      }`}
      ref={containerRef}
      tabIndex={0}
      onKeyDown={handleShortcut}
      aria-label="Video player"
    >
      <div className={`video-player-shell ${aspectClass}`}>
        <video
          key={reloadNonce}
          ref={videoRef}
          className="video-element"
          src={videoUrl}
          playsInline
          preload="auto"
          onLoadedMetadata={handleLoadedMetadata}
          onTimeUpdate={!progressIsActive ? handleTimeUpdate : undefined}
          onPlay={handleVideoElementPlay}
          onPause={handleVideoElementPause}
          onEnded={handleVideoEnded}
          onError={handleVideoError}
        />
      </div>

      <div className="video-controls" aria-label="Video controls">
        <div className="controls-left">
          <button
            type="button"
            className="control-btn play-pause-btn"
            onClick={togglePlay}
            aria-label={isPlaying ? "Pause video" : "Play video"}
          >
            {isPlaying ? <Pause size={20} /> : <Play size={20} />}
          </button>

          <button
            type="button"
            className="control-btn seek-btn"
            onClick={seekToBeginning}
            aria-label="Seek to beginning"
            disabled={!duration}
          >
            <SkipBack size={18} />
          </button>

          <button
            type="button"
            className="control-btn seek-btn"
            onClick={seekToEnd}
            aria-label="Seek to end"
            disabled={!duration}
          >
            <SkipForward size={18} />
          </button>

          <span className="time-display">
            {formatTime(currentTime)} / {formatTime(duration)}
          </span>
        </div>

        <div className="video-progress-wrapper">
          <input
            type="range"
            min="0"
            max={duration || 0}
            step="0.1"
            value={duration ? currentTime : 0}
            className="video-progress-slider"
            onChange={handleProgressChange}
            onMouseDown={handleProgressInteraction(true)}
            onMouseUp={handleProgressInteraction(false)}
            onTouchStart={handleProgressInteraction(true)}
            onTouchEnd={handleProgressInteraction(false)}
            aria-label="Video progress"
            disabled={!duration}
            style={{ "--progress": `${progressPercent}%` }}
          />
        </div>

        <div className="controls-right">
          <div className="volume-control">
            <span className="volume-icon" aria-hidden="true">
              <Volume2 size={18} />
            </span>
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              value={volume}
              onChange={handleVolumeChange}
              className="volume-slider"
              aria-label="Volume"
            />
          </div>

          <button
            type="button"
            className="control-btn fullscreen-btn"
            onClick={toggleFullscreen}
            aria-label={isFullscreen ? "Exit fullscreen" : "Enter fullscreen"}
          >
            {isFullscreen ? <Minimize size={18} /> : <Maximize size={18} />}
          </button>
        </div>
      </div>
    </section>
  );
}

VideoPlayer.propTypes = {
  videoUrl: PropTypes.string,
  status: PropTypes.oneOf(STATUS_VARIANTS).isRequired,
  aspectRatio: PropTypes.oneOf(Object.keys(ASPECT_RATIO_CLASSES)),
  onError: PropTypes.func,
  onRefresh: PropTypes.func,
  // DEPRECATED: Audio is now embedded in video. These props are kept for backwards compatibility.
  backgroundMusicUrl: PropTypes.string,
  narratorAudioUrl: PropTypes.string,
};

VideoPlayer.defaultProps = {
  videoUrl: null,
  aspectRatio: "16:9",
  onError: undefined,
  onRefresh: undefined,
  backgroundMusicUrl: null,
  narratorAudioUrl: null,
};

export default VideoPlayer;
