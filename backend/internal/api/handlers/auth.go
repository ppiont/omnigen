package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

type AuthHandler struct {
	jwtValidator *auth.JWTValidator
	cookieConfig auth.CookieConfig
	logger       *zap.Logger
}

func NewAuthHandler(jwtValidator *auth.JWTValidator, cookieConfig auth.CookieConfig, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		jwtValidator: jwtValidator,
		cookieConfig: cookieConfig,
		logger:       logger,
	}
}

// LoginRequest represents the request body for login endpoint
type LoginRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	IDToken      string `json:"id_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse represents the user data returned to frontend
type UserResponse struct {
	UserID           string `json:"user_id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	SubscriptionTier string `json:"subscription_tier"`
}

// @Summary Exchange Cognito tokens for httpOnly cookies
// @Description Frontend calls this after authenticating with Cognito SDK to exchange tokens for secure httpOnly cookies
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Cognito tokens from frontend"
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Invalid token"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewAPIError(
			errors.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			nil,
		))
		return
	}

	// Validate the ID token
	claims, err := h.jwtValidator.ValidateToken(req.IDToken)
	if err != nil {
		h.logger.Warn("Invalid ID token during login", zap.Error(err))
		c.JSON(http.StatusUnauthorized, errors.NewAPIError(
			errors.ErrUnauthorized,
			"Invalid or expired token",
			nil,
		))
		return
	}

	// Set httpOnly cookies
	auth.SetAuthCookies(c, req.AccessToken, req.IDToken, req.RefreshToken, h.cookieConfig)

	h.logger.Info("User logged in successfully",
		zap.String("user_id", claims.Sub),
		zap.String("email", claims.Email),
	)

	c.JSON(http.StatusOK, UserResponse{
		UserID:           claims.Sub,
		Email:            claims.Email,
		Name:             claims.Name,
		SubscriptionTier: claims.SubscriptionTier,
	})
}

// @Summary Refresh access token
// @Description Refresh the access token using the refresh token from httpOnly cookie
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "Returns new tokens"
// @Failure 401 {object} errors.ErrorResponse "Missing or invalid refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken := auth.GetRefreshTokenFromCookie(c)
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, errors.NewAPIError(
			errors.ErrUnauthorized,
			"No refresh token found in cookies",
			nil,
		))
		return
	}

	// Note: Frontend will handle the actual token refresh with Cognito SDK
	// This endpoint just validates the refresh token exists
	// The frontend should call Cognito's refresh API and then call /auth/login again

	c.JSON(http.StatusOK, gin.H{
		"message": "Refresh token found. Frontend should refresh with Cognito SDK and call /auth/login with new tokens.",
	})
}

// @Summary Get current user information
// @Description Returns the current authenticated user's information from JWT
// @Tags auth
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} errors.ErrorResponse "Not authenticated"
// @Router /api/v1/auth/me [get]
// @Security BearerAuth
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := auth.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewAPIError(
			errors.ErrUnauthorized,
			"User not authenticated",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		UserID:           claims.Sub,
		Email:            claims.Email,
		Name:             claims.Name,
		SubscriptionTier: claims.SubscriptionTier,
	})
}

// @Summary Logout
// @Description Clear httpOnly cookies and log out the user
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user info before clearing (for logging)
	claims, ok := auth.GetUserClaims(c)

	// Clear cookies
	auth.ClearAuthCookies(c, h.cookieConfig)

	if ok {
		h.logger.Info("User logged out",
			zap.String("user_id", claims.Sub),
			zap.String("email", claims.Email),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
