import { useCallback, useEffect, useState } from "react";
import PropTypes from "prop-types";

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

  const handleDownloadClick = () => {
    if (isDownloadDisabled || !jobData || typeof onDownload !== "function") {
      return;
    }

    try {
      onDownload(jobData);
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
        <button
          type="button"
          className="action-btn action-btn-download"
          onClick={handleDownloadClick}
          disabled={isDownloadDisabled}
          title={
            isDownloadDisabled
              ? "Video is not ready for download yet"
              : "Download video"
          }
          aria-disabled={isDownloadDisabled}
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
        </button>

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
  }).isRequired,
  onDownload: PropTypes.func,
  onDelete: PropTypes.func,
};

ActionsToolbar.defaultProps = {
  onDownload: undefined,
  onDelete: undefined,
};

export default ActionsToolbar;
