import { useState, useRef } from "react";
import "../../styles/create.css";

const ACCEPTED_TYPES = ["application/pdf", "application/json", "text/plain", "image/jpeg", "image/jpg", "image/png"];
const MAX_FILE_SIZE = 25 * 1024 * 1024; // 25MB for documents

function BrandGuidelineUpload({ onGuidelineSelect, currentGuideline, disabled }) {
  const [error, setError] = useState("");
  const [preview, setPreview] = useState(null);
  const fileInputRef = useRef(null);

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const getFileTypeDisplay = (type) => {
    switch (type) {
      case "application/pdf":
        return "PDF Document";
      case "application/json":
        return "JSON Document";
      case "text/plain":
        return "Text Document";
      case "image/jpeg":
      case "image/jpg":
      case "image/png":
        return "Image Document";
      default:
        return "Document";
    }
  };

  const validateFile = (file) => {
    // Check file type
    if (!ACCEPTED_TYPES.includes(file.type)) {
      return "Invalid file type. Please upload PDF, JSON, text files, or images (JPG, PNG).";
    }

    // Check file size
    if (file.size > MAX_FILE_SIZE) {
      return `File too large. Maximum size is ${formatFileSize(MAX_FILE_SIZE)}.`;
    }

    return null;
  };

  const handleFileSelect = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setError("");

    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      setPreview(null);
      onGuidelineSelect(null);
      return;
    }

    // Create preview for documents
    const preview = {
      name: file.name,
      size: file.size,
      type: file.type,
      displayType: getFileTypeDisplay(file.type),
    };

    // For images, also create a visual preview
    if (file.type.startsWith('image/')) {
      const reader = new FileReader();
      reader.onloadend = () => {
        preview.imageData = reader.result;
        setPreview(preview);
        onGuidelineSelect({
          file,
          preview,
          name: file.name,
          size: file.size,
        });
      };
      reader.readAsDataURL(file);
    } else {
      setPreview(preview);
      onGuidelineSelect({
        file,
        preview,
        name: file.name,
        size: file.size,
      });
    }
  };

  const handleRemove = () => {
    setPreview(null);
    setError("");
    onGuidelineSelect(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const handleClick = () => {
    if (!disabled) {
      fileInputRef.current?.click();
    }
  };

  const currentPreview = preview || (currentGuideline ? {
    name: currentGuideline.name,
    size: currentGuideline.size,
    type: currentGuideline.type,
    displayType: getFileTypeDisplay(currentGuideline.type),
    imageData: currentGuideline.preview?.imageData,
  } : null);

  return (
    <div className="image-upload-container">
      <label className="option-label">Brand Guidelines Document</label>

      <div className={`image-upload-area ${disabled ? 'upload-disabled' : ''}`}>
        {currentPreview ? (
          <div className="image-preview-container">
            <div className="brand-guideline-preview">
              {currentPreview.imageData ? (
                // Show image preview for image files
                <img
                  src={currentPreview.imageData}
                  alt="Brand guideline preview"
                  className="guideline-image-preview"
                />
              ) : (
                // Show document icon for other files
                <div className="document-preview">
                  <div className="document-icon">
                    {currentPreview.type === 'application/pdf' && '📄'}
                    {currentPreview.type === 'application/json' && '📋'}
                    {currentPreview.type === 'text/plain' && '📝'}
                    {!['application/pdf', 'application/json', 'text/plain'].includes(currentPreview.type) && '📄'}
                  </div>
                  <div className="document-type">{currentPreview.displayType}</div>
                </div>
              )}
            </div>
            <div className="image-info">
              <div className="image-details">
                <span className="image-name">
                  {currentPreview.name}
                </span>
                {currentPreview.size && (
                  <span className="image-size">
                    {formatFileSize(currentPreview.size)}
                  </span>
                )}
              </div>
              {!disabled && (
                <button
                  type="button"
                  className="btn-remove-image"
                  onClick={handleRemove}
                  aria-label="Remove brand guideline document"
                >
                  ✕ Remove
                </button>
              )}
            </div>
          </div>
        ) : (
          <button
            type="button"
            className="image-upload-button"
            onClick={handleClick}
            disabled={disabled}
          >
            <div className="upload-icon">📋</div>
            <div className="upload-text">
              <span className="upload-primary">
                {disabled ? "Brand guidelines upload disabled" : "Choose brand guidelines"}
              </span>
              <span className="upload-secondary">
                PDF, JSON, text files, or images • Max 25MB
              </span>
            </div>
          </button>
        )}

        <input
          ref={fileInputRef}
          type="file"
          accept=".pdf,.json,.txt,.jpg,.jpeg,.png"
          onChange={handleFileSelect}
          className="image-upload-input"
          disabled={disabled}
          aria-label="Upload brand guideline document"
        />
      </div>

      {error && <div className="image-upload-error">{error}</div>}

      <p className="option-helper">
        Upload your brand guidelines document to ensure consistent branding in generated videos (optional)
      </p>
    </div>
  );
}

export default BrandGuidelineUpload;