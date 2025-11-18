import { useState, useRef } from "react";
import { Upload, Image as ImageIcon, X } from "lucide-react";
import "../../styles/create.css";

const ACCEPTED_TYPES = ["image/jpeg", "image/jpg", "image/png", "image/webp"];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

/**
 * MediaUploadBar - Professional product image upload and duration selector
 * Displays below the prompt textarea
 */
function MediaUploadBar({
  productImage,
  onProductImageSelect,
  durations,
  selectedDuration,
  onDurationChange
}) {
  const [productError, setProductError] = useState("");
  const productInputRef = useRef(null);

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

  const handleProductSelect = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setProductError("");
    const validationError = validateFile(file);
    if (validationError) {
      setProductError(validationError);
      onProductImageSelect(null);
      return;
    }

    const reader = new FileReader();
    reader.onloadend = () => {
      onProductImageSelect({
        file,
        preview: reader.result,
        name: file.name,
        size: file.size,
      });
    };
    reader.readAsDataURL(file);
  };

  const handleProductRemove = () => {
    setProductError("");
    onProductImageSelect(null);
    if (productInputRef.current) {
      productInputRef.current.value = "";
    }
  };

  return (
    <div className="media-upload-bar">
      {/* Product Image Upload Button */}
      <div className="media-upload-item">
        <button
          type="button"
          className={`media-upload-btn ${productImage ? "has-media" : ""}`}
          onClick={() => !productImage && productInputRef.current?.click()}
          title="Upload product image (optional)"
        >
          <ImageIcon size={18} className="media-icon" />
          <span className="media-label">
            {productImage ? "Product Image" : "Upload Product Image"}
          </span>
          {productImage && (
            <button
              className="media-clear-btn"
              onClick={(e) => {
                e.stopPropagation();
                handleProductRemove();
              }}
              aria-label="Remove product image"
              title="Remove image"
            >
              <X size={14} />
            </button>
          )}
        </button>
        <input
          ref={productInputRef}
          type="file"
          accept=".jpg,.jpeg,.png,.webp"
          onChange={handleProductSelect}
          className="media-upload-input"
          aria-label="Upload product image"
        />

        {productImage && (
          <div className="media-preview-tooltip">
            <div className="media-preview-header">
              <span className="media-preview-title">Product Image</span>
              <button
                className="media-remove-btn"
                onClick={handleProductRemove}
                aria-label="Remove product image"
              >
                <X size={16} />
              </button>
            </div>
            <div className="media-preview-image">
              <img src={productImage.preview} alt="Product" />
            </div>
            <div className="media-preview-info">
              <span className="media-filename">{productImage.name}</span>
              <span className="media-filesize">
                {formatFileSize(productImage.size)}
              </span>
            </div>
            <p className="media-preview-description">
              Used as the starting frame for your video
            </p>
          </div>
        )}

        {productError && (
          <div className="media-upload-error">{productError}</div>
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
