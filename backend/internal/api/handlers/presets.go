package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/service"
	"go.uber.org/zap"
)

// PresetsHandler handles brand preset requests
type PresetsHandler struct {
	mockService *service.MockService
	logger      *zap.Logger
}

// NewPresetsHandler creates a new presets handler
func NewPresetsHandler(
	mockService *service.MockService,
	logger *zap.Logger,
) *PresetsHandler {
	return &PresetsHandler{
		mockService: mockService,
		logger:      logger,
	}
}

// PresetsResponse represents the list of brand presets
type PresetsResponse struct {
	Presets []service.BrandPreset `json:"presets"`
}

// GetPresets handles GET /api/v1/presets
// @Summary Get brand style presets
// @Description Get a list of predefined brand style presets for quick video generation
// @Tags presets
// @Produce json
// @Success 200 {object} PresetsResponse
// @Router /api/v1/presets [get]
func (h *PresetsHandler) GetPresets(c *gin.Context) {
	presets := h.mockService.GetMockPresets()

	h.logger.Info("Presets retrieved",
		zap.Int("count", len(presets)),
	)

	c.JSON(http.StatusOK, PresetsResponse{
		Presets: presets,
	})
}
