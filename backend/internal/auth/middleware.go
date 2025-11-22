package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// DevAuthMiddleware creates a middleware that bypasses authentication for local development
// It sets a mock user context with a test user ID
func DevAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("Dev auth: bypassing JWT validation")

		// Set mock user claims for development
		mockClaims := &domain.UserClaims{
			Sub:              "dev-user-123",
			Email:            "dev@localhost",
			CognitoUsername:  "devuser",
			SubscriptionTier: "pro",
		}
		SetUserClaims(c, mockClaims)
		c.Next()
	}
}

// JWTAuthMiddleware creates a middleware that validates JWT tokens
// If SKIP_AUTH=true environment variable is set, uses DevAuthMiddleware instead
func JWTAuthMiddleware(validator *JWTValidator, logger *zap.Logger) gin.HandlerFunc {
	// Check for SKIP_AUTH environment variable for local development
	if os.Getenv("SKIP_AUTH") == "true" {
		logger.Info("SKIP_AUTH=true: Using development auth middleware (NO AUTHENTICATION)")
		return DevAuthMiddleware(logger)
	}
	return func(c *gin.Context) {
		// First, try to get token from cookie
		tokenString := GetTokenFromCookie(c)

		// If not in cookie, try Authorization header (for backwards compatibility)
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				// Check if it's a Bearer token
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}
		}

		// If still no token, return unauthorized
		if tokenString == "" {
			logger.Warn("No authentication token found in cookies or Authorization header")
			c.JSON(http.StatusUnauthorized, errors.NewAPIError(
				errors.ErrUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		// Validate token
		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Token validation failed",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, errors.NewAPIError(
				errors.ErrUnauthorized,
				"Invalid or expired token",
				map[string]interface{}{
					"error": err.Error(),
				},
			))
			c.Abort()
			return
		}

		// Store user information in context
		SetUserClaims(c, claims)

		// Continue to next handler
		c.Next()
	}
}

// OptionalJWTAuthMiddleware creates a middleware that optionally validates JWT tokens
// If a token is present, it validates and stores user context
// If no token is present, it continues without authentication
func OptionalJWTAuthMiddleware(validator *JWTValidator, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// Invalid format, continue without authentication
			logger.Debug("Invalid Authorization header format in optional auth",
				zap.String("header", authHeader))
			c.Next()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			logger.Debug("Token validation failed in optional auth",
				zap.Error(err))
			c.Next()
			return
		}

		// Store user information in context
		SetUserClaims(c, claims)

		logger.Debug("User authenticated (optional)",
			zap.String("user_id", claims.Sub),
			zap.String("email", claims.Email))

		// Continue to next handler
		c.Next()
	}
}

// RequireSubscriptionTier creates a middleware that requires a specific subscription tier
func RequireSubscriptionTier(minTier string, logger *zap.Logger) gin.HandlerFunc {
	// Define tier hierarchy
	tierLevel := map[string]int{
		"free":       1,
		"pro":        2,
		"enterprise": 3,
	}

	return func(c *gin.Context) {
		// Get user claims from context
		claims, ok := GetUserClaims(c)
		if !ok {
			logger.Warn("Subscription tier check failed: user not authenticated")
			c.JSON(http.StatusUnauthorized, errors.NewAPIError(
				errors.ErrUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		// Get user's subscription tier
		userTier := claims.SubscriptionTier
		if userTier == "" {
			userTier = "free" // Default to free tier
		}

		// Check if user's tier meets the minimum requirement
		userLevel := tierLevel[strings.ToLower(userTier)]
		minLevel := tierLevel[strings.ToLower(minTier)]

		if userLevel < minLevel {
			logger.Warn("Insufficient subscription tier",
				zap.String("user_id", claims.Sub),
				zap.String("user_tier", userTier),
				zap.String("required_tier", minTier))
			c.JSON(http.StatusForbidden, errors.NewAPIError(
				errors.ErrForbidden,
				"This feature requires a higher subscription tier",
				map[string]interface{}{
					"current_tier":  userTier,
					"required_tier": minTier,
				},
			))
			c.Abort()
			return
		}

		// User has sufficient tier, continue
		c.Next()
	}
}
