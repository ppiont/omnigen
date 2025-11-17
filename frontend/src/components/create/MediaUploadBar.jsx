import { useState, useRef } from "react";
import "../../styles/create.css";

const ACCEPTED_TYPES = ["image/jpeg", "image/jpg", "image/png", "image/webp"];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

/**
 * MediaUploadBar - Compact icon-based media upload buttons
 * Displays below the prompt textarea as icon buttons
 */
function MediaUploadBar({
  referenceImage,
  onReferenceImageSelect,
  templateImage,
  onTemplateImageSelect,
  durations,
  selectedDuration,
  onDurationChange
}) {
  const [referenceError, setReferenceError] = useState("");
  const [templateError, setTemplateError] = useState("");
  const referenceInputRef = useRef(null);
  const templateInputRef = useRef(null);

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const validateFile = (file) => {
    if (!ACCEPTED_TYPES.includes(file.type)) {
      return "Invalid file type. Please upload JPG, PNG, or WebP images.";
    }
    if (file.size > MAX_FILE_SIZE) {
      return `File too large. Maximum size is ${formatFileSize(MAX_FILE_SIZE)}.`;
    }
    return null;
  };

  const handleReferenceSelect = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setReferenceError("");
    const validationError = validateFile(file);
    if (validationError) {
      setReferenceError(validationError);
      onReferenceImageSelect(null);
      return;
    }

    const reader = new FileReader();
    reader.onloadend = () => {
      onReferenceImageSelect({
        file,
        preview: reader.result,
        name: file.name,
        size: file.size,
      });
    };
    reader.readAsDataURL(file);
  };

  const handleTemplateSelect = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setTemplateError("");
    const validationError = validateFile(file);
    if (validationError) {
      setTemplateError(validationError);
      onTemplateImageSelect(null);
      return;
    }

    const reader = new FileReader();
    reader.onloadend = () => {
      onTemplateImageSelect({
        file,
        preview: reader.result,
        name: file.name,
        size: file.size,
      });
    };
    reader.readAsDataURL(file);
  };

  const handleReferenceRemove = () => {
    setReferenceError("");
    onReferenceImageSelect(null);
    if (referenceInputRef.current) {
      referenceInputRef.current.value = "";
    }
  };

  const handleTemplateRemove = () => {
    setTemplateError("");
    onTemplateImageSelect(null);
    if (templateInputRef.current) {
      templateInputRef.current.value = "";
    }
  };

  return (
    <div className="media-upload-bar">
      {/* Reference Image Button */}
      <div className="media-upload-item">
        <button
          type="button"
          className={`media-upload-btn ${referenceImage ? "has-media" : ""}`}
          onClick={() => !referenceImage && referenceInputRef.current?.click()}
          title="Optional: Add image to guide style"
        >
          <span className="media-icon">+</span>
          <span className="media-label">
            {referenceImage ? "Reference Style Img" : "Add Reference Style Img"}
          </span>
          {referenceImage && (
            <button
              className="media-clear-btn"
              onClick={(e) => {
                e.stopPropagation();
                handleReferenceRemove();
              }}
              aria-label="Remove reference image"
              title="Remove image"
            >
              ✕
            </button>
          )}
        </button>
        <input
          ref={referenceInputRef}
          type="file"
          accept=".jpg,.jpeg,.png,.webp"
          onChange={handleReferenceSelect}
          className="media-upload-input"
          aria-label="Upload reference image"
        />

        {referenceImage && (
          <div className="media-preview-tooltip">
            <div className="media-preview-header">
              <span className="media-preview-title">Reference Image</span>
              <button
                className="media-remove-btn"
                onClick={handleReferenceRemove}
                aria-label="Remove reference image"
              >
                ✕
              </button>
            </div>
            <div className="media-preview-image">
              <img src={referenceImage.preview} alt="Reference" />
            </div>
            <div className="media-preview-info">
              <span className="media-filename">{referenceImage.name}</span>
              <span className="media-filesize">
                {formatFileSize(referenceImage.size)}
              </span>
            </div>
            <p className="media-preview-description">
              Guides the visual style and aesthetic
            </p>
          </div>
        )}

        {referenceError && (
          <div className="media-upload-error">{referenceError}</div>
        )}
      </div>

      {/* Template Image Button */}
      <div className="media-upload-item">
        <button
          type="button"
          className={`media-upload-btn ${templateImage ? "has-media" : ""}`}
          onClick={() => !templateImage && templateInputRef.current?.click()}
          title="Optional: Add image to start your video"
        >
          <span className="media-icon">+</span>
          <span className="media-label">
            {templateImage ? "Starting Img" : "Add Starting Img"}
          </span>
          {templateImage && (
            <button
              className="media-clear-btn"
              onClick={(e) => {
                e.stopPropagation();
                handleTemplateRemove();
              }}
              aria-label="Remove template image"
              title="Remove image"
            >
              ✕
            </button>
          )}
        </button>
        <input
          ref={templateInputRef}
          type="file"
          accept=".jpg,.jpeg,.png,.webp"
          onChange={handleTemplateSelect}
          className="media-upload-input"
          aria-label="Upload template image"
        />

        {templateImage && (
          <div className="media-preview-tooltip">
            <div className="media-preview-header">
              <span className="media-preview-title">Starting Clip</span>
              <button
                className="media-remove-btn"
                onClick={handleTemplateRemove}
                aria-label="Remove template image"
              >
                ✕
              </button>
            </div>
            <div className="media-preview-image">
              <img src={templateImage.preview} alt="Template" />
            </div>
            <div className="media-preview-info">
              <span className="media-filename">{templateImage.name}</span>
              <span className="media-filesize">
                {formatFileSize(templateImage.size)}
              </span>
            </div>
            <p className="media-preview-description">
              Used as the first frame/scene of your video
            </p>
          </div>
        )}

        {templateError && (
          <div className="media-upload-error">{templateError}</div>
        )}
      </div>

      {/* Duration Buttons */}
      {durations && (
        <div className="media-duration-group">
          <label className="duration-label">Duration</label>
          <div className="duration-button-group">
            {durations.map((dur) => (
              <button
                key={dur}
                type="button"
                className={`duration-btn ${
                  selectedDuration === dur ? "is-active" : ""
                }`}
                onClick={() => onDurationChange(dur)}
              >
                {dur}
              </button>
            ))}
          </div>
          <p className="duration-helper">
            Longer videos cost more but allow complex narratives
          </p>
        </div>
      )}
    </div>
  );
}

export default MediaUploadBar;
