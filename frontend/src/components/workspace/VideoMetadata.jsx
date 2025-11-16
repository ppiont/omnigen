import { useEffect, useId, useMemo, useState } from "react";
import PropTypes from "prop-types";
import { formatDistanceToNow } from "date-fns";

const COST_PER_SECOND = 0.07;
const DEFAULT_TITLE = "Untitled Video";
const MODEL_NAME = "Kling v2.5 Turbo Pro";
const DEFAULT_RESOLUTION = "1080p";
const TITLE_STORAGE_PREFIX = "video-title-";

const STATUS_VARIANTS = {
  completed: { label: "Completed", className: "status-completed" },
  processing: { label: "Processing", className: "status-processing" },
  pending: { label: "Pending", className: "status-pending" },
  failed: { label: "Failed", className: "status-failed" },
};

const getStorageKey = (jobId) => `${TITLE_STORAGE_PREFIX}${jobId}`;

const readLocalTitle = (jobId) => {
  if (typeof window === "undefined" || !jobId) return null;
  return window.localStorage.getItem(getStorageKey(jobId));
};

const persistLocalTitle = (jobId, value) => {
  if (typeof window === "undefined" || !jobId) return;
  window.localStorage.setItem(getStorageKey(jobId), value);
};

const resolveStoredTitle = (jobId) =>
  readLocalTitle(jobId)?.trim() || DEFAULT_TITLE;

const MOBILE_VIEWPORT_QUERY = "(max-width: 767px)";

const getIsMobileViewport = () =>
  typeof window !== "undefined" &&
  typeof window.matchMedia === "function" &&
  window.matchMedia(MOBILE_VIEWPORT_QUERY).matches;

/**
 * VideoMetadata displays backend-provided metadata plus derived details with an
 * editable title that persists to localStorage.
 *
 * @param {{jobData: Object}} props - Component props
 * @returns {JSX.Element} Metadata panel
 */
function VideoMetadata({ jobData }) {
  const metadataContentId = useId();
  const [isMobileViewport, setIsMobileViewport] = useState(() =>
    getIsMobileViewport()
  );
  const [isDetailsOpen, setIsDetailsOpen] = useState(
    () => !getIsMobileViewport()
  );
  const [title, setTitle] = useState(() => resolveStoredTitle(jobData?.job_id));

  useEffect(() => {
    if (
      typeof window === "undefined" ||
      typeof window.matchMedia !== "function"
    ) {
      return undefined;
    }

    const mediaQuery = window.matchMedia(MOBILE_VIEWPORT_QUERY);

    const handleViewportChange = (event) => {
      setIsMobileViewport(event.matches);
      setIsDetailsOpen(event.matches ? false : true);
    };

    mediaQuery.addEventListener("change", handleViewportChange);
    return () => mediaQuery.removeEventListener("change", handleViewportChange);
  }, []);

  const durationSeconds = useMemo(() => {
    if (jobData?.duration === undefined || jobData?.duration === null) {
      return null;
    }
    const numeric = Number(jobData.duration);
    return Number.isFinite(numeric) ? numeric : null;
  }, [jobData?.duration]);

  const formattedDuration = useMemo(() => {
    if (durationSeconds === null) {
      return "Not specified";
    }
    return `${durationSeconds} second${durationSeconds === 1 ? "" : "s"}`;
  }, [durationSeconds]);

  const formattedCost = useMemo(() => {
    if (durationSeconds === null) {
      return "Not specified";
    }
    const total = durationSeconds * COST_PER_SECOND;
    return `$${total.toFixed(2)}`;
  }, [durationSeconds]);

  const createdAtTimestamp = jobData?.created_at ?? null;

  const formattedCreatedTime = useMemo(() => {
    if (!createdAtTimestamp) {
      return "Unknown";
    }
    try {
      return formatDistanceToNow(createdAtTimestamp * 1000, {
        addSuffix: true,
      });
    } catch {
      return "Unknown";
    }
  }, [createdAtTimestamp]);

  const normalizedStatus = jobData?.status?.toLowerCase();
  const statusVariant = STATUS_VARIANTS[normalizedStatus] || {
    label: jobData?.status || "Unknown",
    className: "status-unknown",
  };

  const handleTitleChange = (event) => {
    setTitle(event.target.value);
  };

  const handleTitleBlur = () => {
    const nextTitle = title.trim() || DEFAULT_TITLE;
    setTitle(nextTitle);
    persistLocalTitle(jobData?.job_id, nextTitle);
  };

  const handleToggleDetails = () => {
    if (!isMobileViewport) {
      return;
    }
    setIsDetailsOpen((prev) => !prev);
  };

  if (!jobData) {
    return (
      <section className="video-metadata-panel">
        <p className="metadata-empty">No metadata available.</p>
      </section>
    );
  }

  const detailFields = [
    { label: "Duration", value: formattedDuration },
    {
      label: "Aspect Ratio",
      value: jobData.aspect_ratio || "Not specified",
    },
    { label: "Style", value: jobData.style || "Not specified" },
    { label: "Generated", value: formattedCreatedTime },
  ];

  const derivedFields = [
    { label: "Model", value: MODEL_NAME },
    { label: "Resolution", value: DEFAULT_RESOLUTION },
    { label: "Estimated Cost", value: formattedCost },
  ];

  const collapseContentClassName = [
    "metadata-collapse-content",
    isMobileViewport ? (isDetailsOpen ? "is-open" : "is-collapsed") : "is-open",
  ].join(" ");

  return (
    <section
      className="video-metadata-panel"
      aria-labelledby="video-title-input"
    >
      {isMobileViewport && (
        <button
          type="button"
          className="metadata-collapse-toggle"
          aria-expanded={isDetailsOpen}
          aria-controls={metadataContentId}
          onClick={handleToggleDetails}
        >
          <span>Video details</span>
          <svg
            className={`metadata-collapse-icon ${
              isDetailsOpen ? "is-open" : ""
            }`}
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>
      )}

      <div
        id={metadataContentId}
        className={collapseContentClassName}
        aria-hidden={isMobileViewport ? !isDetailsOpen : false}
      >
        <div className="metadata-title-row">
          <input
            id="video-title-input"
            type="text"
            className="editable-title"
            value={title}
            onChange={handleTitleChange}
            onBlur={handleTitleBlur}
            placeholder={DEFAULT_TITLE}
            aria-label="Video title"
          />
          <p className="metadata-helper">Title saves locally on this device</p>
        </div>

        <div className="metadata-field metadata-status">
          <span className="metadata-label">Status</span>
          <span className={`status-badge ${statusVariant.className}`}>
            {statusVariant.label}
          </span>
        </div>

        <div className="metadata-field metadata-job-id">
          <span className="metadata-label">Job ID</span>
          <code className="metadata-code">{jobData.job_id}</code>
        </div>

        <div className="metadata-grid">
          {detailFields.map((field) => (
            <div className="metadata-field" key={field.label}>
              <span className="metadata-label">{field.label}</span>
              <span className="metadata-value">{field.value}</span>
            </div>
          ))}
        </div>

        <div className="metadata-grid metadata-grid-secondary">
          {derivedFields.map((field) => (
            <div className="metadata-field" key={field.label}>
              <span className="metadata-label">{field.label}</span>
              <span className="metadata-value">{field.value}</span>
            </div>
          ))}
        </div>

        <div className="metadata-field metadata-prompt">
          <span className="metadata-label">Original Prompt</span>
          <p className="metadata-prompt-text">
            {jobData.prompt?.trim() || "No prompt provided."}
          </p>
        </div>
      </div>
    </section>
  );
}

VideoMetadata.propTypes = {
  jobData: PropTypes.shape({
    job_id: PropTypes.string.isRequired,
    status: PropTypes.string.isRequired,
    prompt: PropTypes.string,
    duration: PropTypes.number,
    style: PropTypes.string,
    aspect_ratio: PropTypes.string,
    created_at: PropTypes.number,
    completed_at: PropTypes.number,
  }).isRequired,
};

export default VideoMetadata;
