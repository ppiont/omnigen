package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/pkg/errors"
)

// Auth creates an authentication middleware that validates API keys
func Auth(validKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")

		if apiKey == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				errors.ErrorResponse{Error: errors.ErrMissingAPIKey},
			)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		valid := false
		for _, key := range validKeys {
			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
				valid = true
				break
			}
		}

		if !valid {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				errors.ErrorResponse{Error: errors.ErrInvalidAPIKey},
			)
			return
		}

		c.Next()
	}
}
