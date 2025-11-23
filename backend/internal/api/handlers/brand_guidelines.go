package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// BrandGuidelinesHandler handles brand guideline requests
type BrandGuidelinesHandler struct {
	brandGuidelinesRepo repository.BrandGuidelinesRepository
	s3Service          *repository.S3AssetRepository
	assetsBucket       string
	logger             *zap.Logger
}

// NewBrandGuidelinesHandler creates a new brand guidelines handler
func NewBrandGuidelinesHandler(
	brandGuidelinesRepo repository.BrandGuidelinesRepository,
	s3Service *repository.S3AssetRepository,
	assetsBucket string,
	logger *zap.Logger,
) *BrandGuidelinesHandler {
	return &BrandGuidelinesHandler{
		brandGuidelinesRepo: brandGuidelinesRepo,
		s3Service:          s3Service,
		assetsBucket:       assetsBucket,
		logger:             logger,
	}
}

// CreateBrandGuidelines handles POST /api/v1/brand-guidelines
// @Summary Create brand guidelines
// @Description Create new brand guidelines for the authenticated user
// @Tags brand-guidelines
// @Accept json
// @Produce json
// @Param body body domain.BrandGuidelineRequest true "Brand guideline data"
// @Success 201 {object} domain.BrandGuidelines
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines [post]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) CreateBrandGuidelines(c *gin.Context) {
	userID := auth.MustGetUserID(c)

	var req domain.BrandGuidelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid brand guidelines request",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Generate unique ID
	guidelineID := uuid.New().String()
	now := time.Now().Unix()

	// Create brand guidelines object
	guidelines := &domain.BrandGuidelines{
		GuidelineID:    guidelineID,
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		IsActive:       false, // New guidelines start as inactive
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Set optional fields if provided
	if req.Colors != nil {
		guidelines.Colors = *req.Colors
	}
	if req.Typography != nil {
		guidelines.Typography = *req.Typography
	}
	if req.LogoAssets != nil {
		guidelines.LogoAssets = req.LogoAssets
	}
	if req.BrandVoice != nil {
		guidelines.BrandVoice = *req.BrandVoice
	}
	if req.ImageStyle != nil {
		guidelines.ImageStyle = *req.ImageStyle
	}
	if req.VideoStyle != nil {
		guidelines.VideoStyle = *req.VideoStyle
	}
	if req.SourceDocument != nil {
		guidelines.SourceDocument = req.SourceDocument
	}

	// Save to repository
	if err := h.brandGuidelinesRepo.CreateBrandGuidelines(c.Request.Context(), guidelines); err != nil {
		h.logger.Error("Failed to create brand guidelines",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	h.logger.Info("Brand guidelines created successfully",
		zap.String("user_id", userID),
		zap.String("guideline_id", guidelineID),
		zap.String("name", req.Name),
	)

	c.JSON(http.StatusCreated, guidelines)
}

// GetBrandGuidelines handles GET /api/v1/brand-guidelines/:id
// @Summary Get brand guidelines by ID
// @Description Get specific brand guidelines by ID
// @Tags brand-guidelines
// @Produce json
// @Param id path string true "Guideline ID"
// @Success 200 {object} domain.BrandGuidelines
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines/{id} [get]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) GetBrandGuidelines(c *gin.Context) {
	userID := auth.MustGetUserID(c)
	guidelineID := c.Param("id")

	if guidelineID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	guidelines, err := h.brandGuidelinesRepo.GetBrandGuidelines(c.Request.Context(), guidelineID)
	if err != nil {
		h.logger.Error("Failed to get brand guidelines",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	if guidelines == nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	// Verify user owns this guideline
	if guidelines.UserID != userID {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, guidelines)
}

// ListBrandGuidelines handles GET /api/v1/brand-guidelines
// @Summary List user's brand guidelines
// @Description Get all brand guidelines for the authenticated user
// @Tags brand-guidelines
// @Produce json
// @Success 200 {object} []domain.BrandGuidelineSummary
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines [get]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) ListBrandGuidelines(c *gin.Context) {
	userID := auth.MustGetUserID(c)

	guidelines, err := h.brandGuidelinesRepo.GetBrandGuidelinesByUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list brand guidelines",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Convert to summaries for list view
	summaries := make([]domain.BrandGuidelineSummary, len(guidelines))
	for i, guideline := range guidelines {
		summaries[i] = guideline.ToSummary()
	}

	c.JSON(http.StatusOK, summaries)
}

// UpdateBrandGuidelines handles PUT /api/v1/brand-guidelines/:id
// @Summary Update brand guidelines
// @Description Update existing brand guidelines
// @Tags brand-guidelines
// @Accept json
// @Produce json
// @Param id path string true "Guideline ID"
// @Param body body domain.BrandGuidelineRequest true "Updated brand guideline data"
// @Success 200 {object} domain.BrandGuidelines
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines/{id} [put]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) UpdateBrandGuidelines(c *gin.Context) {
	userID := auth.MustGetUserID(c)
	guidelineID := c.Param("id")

	if guidelineID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	var req domain.BrandGuidelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid brand guidelines update request",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Get existing guidelines
	guidelines, err := h.brandGuidelinesRepo.GetBrandGuidelines(c.Request.Context(), guidelineID)
	if err != nil {
		h.logger.Error("Failed to get brand guidelines for update",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	if guidelines == nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	// Verify user owns this guideline
	if guidelines.UserID != userID {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	// Update fields
	guidelines.Name = req.Name
	guidelines.Description = req.Description
	guidelines.UpdatedAt = time.Now().Unix()

	// Update optional fields if provided
	if req.Colors != nil {
		guidelines.Colors = *req.Colors
	}
	if req.Typography != nil {
		guidelines.Typography = *req.Typography
	}
	if req.LogoAssets != nil {
		guidelines.LogoAssets = req.LogoAssets
	}
	if req.BrandVoice != nil {
		guidelines.BrandVoice = *req.BrandVoice
	}
	if req.ImageStyle != nil {
		guidelines.ImageStyle = *req.ImageStyle
	}
	if req.VideoStyle != nil {
		guidelines.VideoStyle = *req.VideoStyle
	}
	if req.SourceDocument != nil {
		guidelines.SourceDocument = req.SourceDocument
	}

	// Save updates
	if err := h.brandGuidelinesRepo.UpdateBrandGuidelines(c.Request.Context(), guidelines); err != nil {
		h.logger.Error("Failed to update brand guidelines",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	h.logger.Info("Brand guidelines updated successfully",
		zap.String("user_id", userID),
		zap.String("guideline_id", guidelineID),
		zap.String("name", req.Name),
	)

	c.JSON(http.StatusOK, guidelines)
}

// SetActiveBrandGuidelines handles POST /api/v1/brand-guidelines/:id/activate
// @Summary Set active brand guidelines
// @Description Set specific brand guidelines as active for the user
// @Tags brand-guidelines
// @Produce json
// @Param id path string true "Guideline ID"
// @Success 200 {object} domain.BrandGuidelines
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines/{id}/activate [post]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) SetActiveBrandGuidelines(c *gin.Context) {
	userID := auth.MustGetUserID(c)
	guidelineID := c.Param("id")

	if guidelineID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Verify guideline exists and belongs to user
	guidelines, err := h.brandGuidelinesRepo.GetBrandGuidelines(c.Request.Context(), guidelineID)
	if err != nil {
		h.logger.Error("Failed to get brand guidelines for activation",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	if guidelines == nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	if guidelines.UserID != userID {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	// Set as active (this will deactivate other guidelines for the user)
	if err := h.brandGuidelinesRepo.SetActiveBrandGuidelines(c.Request.Context(), userID, guidelineID); err != nil {
		h.logger.Error("Failed to set active brand guidelines",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Return updated guidelines
	guidelines.IsActive = true
	guidelines.UpdatedAt = time.Now().Unix()

	h.logger.Info("Brand guidelines activated",
		zap.String("user_id", userID),
		zap.String("guideline_id", guidelineID),
	)

	c.JSON(http.StatusOK, guidelines)
}

// DeleteBrandGuidelines handles DELETE /api/v1/brand-guidelines/:id
// @Summary Delete brand guidelines
// @Description Delete specific brand guidelines
// @Tags brand-guidelines
// @Produce json
// @Param id path string true "Guideline ID"
// @Success 204
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines/{id} [delete]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) DeleteBrandGuidelines(c *gin.Context) {
	userID := auth.MustGetUserID(c)
	guidelineID := c.Param("id")

	if guidelineID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Verify guideline exists and belongs to user
	guidelines, err := h.brandGuidelinesRepo.GetBrandGuidelines(c.Request.Context(), guidelineID)
	if err != nil {
		h.logger.Error("Failed to get brand guidelines for deletion",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	if guidelines == nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	if guidelines.UserID != userID {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound,
		})
		return
	}

	// Delete from repository
	if err := h.brandGuidelinesRepo.DeleteBrandGuidelines(c.Request.Context(), guidelineID); err != nil {
		h.logger.Error("Failed to delete brand guidelines",
			zap.String("user_id", userID),
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	h.logger.Info("Brand guidelines deleted",
		zap.String("user_id", userID),
		zap.String("guideline_id", guidelineID),
	)

	c.Status(http.StatusNoContent)
}

// GetPresignedURLForBrandDocument handles POST /api/v1/brand-guidelines/upload/presigned-url
// @Summary Get presigned URL for brand document upload
// @Description Generate a presigned URL for uploading a brand guidelines document
// @Tags brand-guidelines
// @Accept json
// @Produce json
// @Param body body PresignedURLRequest true "Upload request"
// @Success 200 {object} PresignedURLResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/brand-guidelines/upload/presigned-url [post]
// @Security BearerAuth
func (h *BrandGuidelinesHandler) GetPresignedURLForBrandDocument(c *gin.Context) {
	userID := auth.MustGetUserID(c)

	var req PresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid presigned URL request for brand document",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Validate file size (max 25MB for documents)
	const maxFileSize = 25 * 1024 * 1024 // 25MB
	if req.FileSize > maxFileSize {
		h.logger.Warn("Brand document file size exceeds limit",
			zap.String("user_id", userID),
			zap.Int64("file_size", req.FileSize),
			zap.Int64("max_size", maxFileSize),
		)
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest,
		})
		return
	}

	// Generate S3 key for brand documents
	filename := sanitizeFilename(req.Filename)
	timestamp := time.Now().Unix()
	s3Key := fmt.Sprintf("users/%s/brand-guidelines/%d_%s", userID, timestamp, filename)

	// Generate presigned PUT URL (valid for 1 hour)
	uploadURL, err := h.s3Service.GetPresignedPutURL(
		c.Request.Context(),
		s3Key,
		req.ContentType,
		1*time.Hour,
	)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL for brand document",
			zap.String("user_id", userID),
			zap.String("s3_key", s3Key),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Generate asset URL for reference
	assetURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", h.assetsBucket, s3Key)

	h.logger.Info("Presigned URL generated for brand document",
		zap.String("user_id", userID),
		zap.String("s3_key", s3Key),
		zap.String("content_type", req.ContentType),
		zap.Int64("file_size", req.FileSize),
	)

	c.JSON(http.StatusOK, PresignedURLResponse{
		UploadURL: uploadURL,
		AssetURL:  assetURL,
	})
}