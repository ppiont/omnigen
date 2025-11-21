package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	s3Service    *repository.S3AssetRepository
	assetsBucket string
	logger       *zap.Logger
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(
	s3Service *repository.S3AssetRepository,
	assetsBucket string,
	logger *zap.Logger,
) *UploadHandler {
	return &UploadHandler{
		s3Service:    s3Service,
		assetsBucket: assetsBucket,
		logger:       logger,
	}
}

// PresignedURLRequest represents the request body for presigned URL generation
type PresignedURLRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,min=1"`
}

// PresignedURLResponse represents the response for presigned URL generation
type PresignedURLResponse struct {
	UploadURL string `json:"upload_url"`
	AssetURL  string `json:"asset_url"`
}

// GetPresignedURL handles POST /api/v1/upload/presigned-url
// @Summary Get presigned URL for file upload
// @Description Generate a presigned URL for uploading a file directly to S3
// @Tags upload
// @Accept json
// @Produce json
// @Param type query string true "Asset type (e.g., 'product_image')"
// @Param body body PresignedURLRequest true "Upload request"
// @Success 200 {object} PresignedURLResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/upload/presigned-url [post]
// @Security BearerAuth
func (h *UploadHandler) GetPresignedURL(c *gin.Context) {
	userID := auth.MustGetUserID(c)
	assetType := c.Query("type")

	if assetType == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	var req PresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid presigned URL request",
			zap.String("user_id", userID),
			zap.String("asset_type", assetType),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if req.FileSize > maxFileSize {
		h.logger.Warn("File size exceeds limit",
			zap.String("user_id", userID),
			zap.String("asset_type", assetType),
			zap.Int64("file_size", req.FileSize),
			zap.Int64("max_size", maxFileSize),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Generate S3 key based on asset type
	var s3Key string
	switch assetType {
	case "product_image":
		// Sanitize filename
		filename := sanitizeFilename(req.Filename)
		// Use timestamp to ensure uniqueness
		timestamp := time.Now().Unix()
		s3Key = fmt.Sprintf("users/%s/uploads/product_images/%d_%s", userID, timestamp, filename)
	default:
		h.logger.Warn("Unsupported asset type",
			zap.String("user_id", userID),
			zap.String("asset_type", assetType),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Generate presigned PUT URL (valid for 1 hour)
	uploadURL, err := h.s3Service.GetPresignedPutURL(
		c.Request.Context(),
		s3Key,
		req.ContentType,
		1*time.Hour,
	)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL",
			zap.String("user_id", userID),
			zap.String("asset_type", assetType),
			zap.String("s3_key", s3Key),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Generate asset URL (for reference after upload)
	assetURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", h.assetsBucket, s3Key)

	h.logger.Info("Presigned URL generated",
		zap.String("user_id", userID),
		zap.String("asset_type", assetType),
		zap.String("s3_key", s3Key),
		zap.String("content_type", req.ContentType),
		zap.Int64("file_size", req.FileSize),
	)

	c.JSON(http.StatusOK, PresignedURLResponse{
		UploadURL: uploadURL,
		AssetURL:  assetURL,
	})
}

// sanitizeFilename removes dangerous characters from filename
func sanitizeFilename(filename string) string {
	// Get base filename (remove path)
	filename = filepath.Base(filename)
	// Remove any remaining path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	// Remove other dangerous characters
	filename = strings.ReplaceAll(filename, "..", "_")
	// Limit length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}
	return filename
}

