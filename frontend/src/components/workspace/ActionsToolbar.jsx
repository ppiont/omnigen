import { useCallback, useEffect, useRef, useState } from "react";
import PropTypes from "prop-types";
import { ChevronDown } from "lucide-react";

/**
 * Renders Download and Delete controls for a workspace video. Heavy lifting
 * (API work) is delegated to parent callbacks.
 *
 * @param {{jobData: Object, onDownload?: Function, onDelete?: Function}} props
 * @returns {JSX.Element} Toolbar controls
 */
function ActionsToolbar({ jobData, onDownload, onDelete }) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showFormatMenu, setShowFormatMenu] = useState(false);
  const formatMenuRef = useRef(null);

  const hasWebM = Boolean(jobData?.webm_video_url);
  const isDownloadDisabled =
    !jobData?.video_url ||
    jobData?.status !== "completed" ||
    typeof onDownload !== "function";
  const isDeleteDisabled =
    !jobData ||
    jobData.status === "processing" ||
    typeof onDelete !== "function";

  const closeModal = useCallback(() => {
    if (isDeleting) {
      return;
    }
    setIsModalOpen(false);
  }, [isDeleting]);

  useEffect(() => {
    if (!isModalOpen) {
      return undefined;
    }

    const handleKeyDown = (event) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeModal();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [closeModal, isModalOpen]);

  // Close format menu when clicking outside
  useEffect(() => {
    if (!showFormatMenu) return undefined;

    const handleClickOutside = (event) => {
      if (formatMenuRef.current && !formatMenuRef.current.contains(event.target)) {
        setShowFormatMenu(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showFormatMenu]);

  const handleDownloadClick = () => {
    if (isDownloadDisabled || !jobData || typeof onDownload !== "function") {
      return;
    }

    // If WebM is available, show format menu; otherwise download MP4 directly
    if (hasWebM) {
      setShowFormatMenu(!showFormatMenu);
    } else {
      try {
        onDownload(jobData, "mp4");
      } catch {
        window.alert("Unable to download the video. Please try again.");
      }
    }
  };

  const handleFormatSelect = (format) => {
    setShowFormatMenu(false);
    if (!jobData || typeof onDownload !== "function") return;

    try {
      onDownload(jobData, format);
    } catch {
      window.alert("Unable to download the video. Please try again.");
    }
  };

  const handleDeleteClick = () => {
    if (isDeleteDisabled) {
      return;
    }
    setIsModalOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!jobData || isDeleteDisabled || typeof onDelete !== "function") {
      return;
    }

    setIsDeleting(true);
    try {
      await onDelete(jobData);
      setIsModalOpen(false);
    } catch {
      window.alert("Unable to delete the video. Please try again.");
    } finally {
      setIsDeleting(false);
    }
  };

  const handleModalBackdropClick = (event) => {
    if (event.target === event.currentTarget) {
      closeModal();
    }
  };

  return (
    <>
      <div
        className="actions-toolbar"
        role="toolbar"
        aria-label="Video actions"
      >
        <div className="download-button-wrapper" ref={formatMenuRef}>
          <button
            type="button"
            className="action-btn action-btn-download"
            onClick={handleDownloadClick}
            disabled={isDownloadDisabled}
            title={
              isDownloadDisabled
                ? "Video is not ready for download yet"
                : hasWebM
                ? "Choose download format"
                : "Download video (MP4)"
            }
            aria-disabled={isDownloadDisabled}
            aria-haspopup={hasWebM ? "true" : undefined}
            aria-expanded={hasWebM ? showFormatMenu : undefined}
          >
            <svg
              className="action-icon"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            <span className="action-btn-text">Download</span>
            {hasWebM && <ChevronDown size={16} className="dropdown-icon" />}
          </button>

          {showFormatMenu && hasWebM && (
            <div className="format-dropdown" role="menu">
              <button
                type="button"
                className="format-option"
                onClick={() => handleFormatSelect("mp4")}
                role="menuitem"
              >
                <span className="format-name">MP4</span>
                <span className="format-desc">Best compatibility</span>
              </button>
              <button
                type="button"
                className="format-option"
                onClick={() => handleFormatSelect("webm")}
                role="menuitem"
              >
                <span className="format-name">WebM</span>
                <span className="format-desc">Smaller file size</span>
              </button>
            </div>
          )}
        </div>

        <button
          type="button"
          className="action-btn action-btn-delete"
          onClick={handleDeleteClick}
          disabled={isDeleteDisabled}
          title={
            isDeleteDisabled
              ? "Video cannot be deleted while processing"
              : "Delete video"
          }
          aria-disabled={isDeleteDisabled}
        >
          <svg
            className="action-icon"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="3 6 5 6 21 6" />
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            <line x1="10" y1="11" x2="10" y2="17" />
            <line x1="14" y1="11" x2="14" y2="17" />
          </svg>
          <span className="action-btn-text">Delete</span>
        </button>
      </div>

      {isModalOpen && (
        <div
          className="modal-overlay"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-modal-title"
          onClick={handleModalBackdropClick}
        >
          <div className="modal-content">
            <div className="modal-header">
              <h2 id="delete-modal-title" className="modal-title">
                Delete Video?
              </h2>
              <button
                type="button"
                className="modal-close"
                aria-label="Close modal"
                onClick={closeModal}
              >
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            <div className="modal-body">
              <p className="modal-message">
                Are you sure you want to delete this video? This cannot be
                undone.
              </p>
            </div>

            <div className="modal-footer">
              <button
                type="button"
                className="modal-btn modal-btn-cancel"
                onClick={closeModal}
                disabled={isDeleting}
              >
                Cancel
              </button>
              <button
                type="button"
                className="modal-btn modal-btn-delete"
                onClick={handleDeleteConfirm}
                disabled={isDeleting}
              >
                {isDeleting ? "Deleting..." : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

ActionsToolbar.propTypes = {
  jobData: PropTypes.shape({
    job_id: PropTypes.string.isRequired,
    status: PropTypes.string.isRequired,
    video_url: PropTypes.string,
    webm_video_url: PropTypes.string,
  }).isRequired,
  onDownload: PropTypes.func, // (jobData, format: 'mp4' | 'webm') => void
  onDelete: PropTypes.func,
};

ActionsToolbar.defaultProps = {
  onDownload: undefined,
  onDelete: undefined,
};

export default ActionsToolbar;
