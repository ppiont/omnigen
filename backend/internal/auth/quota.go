package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/repository"
	pkgerrors "github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// QuotaEnforcementMiddleware creates a middleware that checks and decrements usage quotas
// This should be applied to endpoints that consume quota (e.g., video generation)
func QuotaEnforcementMiddleware(usageRepo *repository.DynamoDBUsageRepository, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context (set by JWT middleware)
		claims, ok := GetUserClaims(c)
		if !ok {
			logger.Error("User claims not found in context")
			c.JSON(http.StatusUnauthorized, pkgerrors.NewAPIError(
				pkgerrors.ErrUnauthorized,
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		userID := claims.Sub
		tier := claims.SubscriptionTier
		if tier == "" {
			tier = "free"
		}

		// Check and decrement quota
		err := usageRepo.CheckAndDecrementQuota(c.Request.Context(), userID, tier)
		if err != nil {
			if errors.Is(err, repository.ErrQuotaExceeded) {
				logger.Warn("User quota exceeded",
					zap.String("user_id", userID),
					zap.String("tier", tier),
				)

				c.JSON(http.StatusPaymentRequired, pkgerrors.NewAPIError(
					&pkgerrors.APIError{
						Code:    "QUOTA_EXCEEDED",
						Message: "Monthly video generation quota exceeded",
						Status:  http.StatusPaymentRequired,
					},
					"You have reached your monthly video generation limit",
					map[string]interface{}{
						"tier":    tier,
						"upgrade": "Upgrade your subscription for higher quotas",
					},
				))
				c.Abort()
				return
			}

			logger.Error("Failed to check quota",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, pkgerrors.NewAPIError(
				pkgerrors.ErrInternalServer,
				"Failed to check usage quota",
				nil,
			))
			c.Abort()
			return
		}

		// Store the fact that quota was decremented so we can rollback if needed
		c.Set("quota_decremented", true)

		c.Next()
	}
}
