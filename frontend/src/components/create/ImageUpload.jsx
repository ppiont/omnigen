import { useState, useRef } from "react";
import "../../styles/create.css";

const ACCEPTED_TYPES = ["image/jpeg", "image/jpg", "image/png", "image/webp"];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

function ImageUpload({ onImageSelect, currentImage }) {
  const [error, setError] = useState("");
  const [preview, setPreview] = useState(null);
  const fileInputRef = useRef(null);

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const validateFile = (file) => {
    // Check file type
    if (!ACCEPTED_TYPES.includes(file.type)) {
      return "Invalid file type. Please upload JPG, PNG, or WebP images.";
    }

    // Check file size
    if (file.size > MAX_FILE_SIZE) {
      return `File too large. Maximum size is ${formatFileSize(
        MAX_FILE_SIZE
      )}.`;
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
      onImageSelect(null);
      return;
    }

    // Create preview
    const reader = new FileReader();
    reader.onloadend = () => {
      setPreview(reader.result);
      onImageSelect({
        file,
        preview: reader.result,
        name: file.name,
        size: file.size,
      });
    };
    reader.readAsDataURL(file);
  };

  const handleRemove = () => {
    setPreview(null);
    setError("");
    onImageSelect(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="image-upload-container">
      <label className="option-label">Reference Image</label>

      <div className="image-upload-area">
        {preview || currentImage ? (
          <div className="image-preview-container">
            <div className="image-preview">
              <img
                src={preview || currentImage.preview}
                alt="Reference preview"
              />
            </div>
            <div className="image-info">
              <div className="image-details">
                <span className="image-name">
                  {currentImage?.name || "Image"}
                </span>
                {currentImage?.size && (
                  <span className="image-size">
                    {formatFileSize(currentImage.size)}
                  </span>
                )}
              </div>
              <button
                type="button"
                className="btn-remove-image"
                onClick={handleRemove}
                aria-label="Remove image"
              >
                ✕ Remove
              </button>
            </div>
          </div>
        ) : (
          <button
            type="button"
            className="image-upload-button"
            onClick={handleClick}
          >
            <div className="upload-icon">📁</div>
            <div className="upload-text">
              <span className="upload-primary">Choose reference image</span>
              <span className="upload-secondary">
                JPG, PNG, or WebP • Max 10MB
              </span>
            </div>
          </button>
        )}

        <input
          ref={fileInputRef}
          type="file"
          accept=".jpg,.jpeg,.png,.webp"
          onChange={handleFileSelect}
          className="image-upload-input"
          aria-label="Upload reference image"
        />
      </div>

      {error && <div className="image-upload-error">{error}</div>}

      <p className="option-helper">
        Upload a reference image to guide the visual style (optional)
      </p>
    </div>
  );
}

export default ImageUpload;
