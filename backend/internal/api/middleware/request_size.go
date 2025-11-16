package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/pkg/errors"
)

// MaxRequestBodySize returns a middleware that limits the request body size
func MaxRequestBodySize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		// Try to parse the body
		if err := c.Request.ParseForm(); err != nil {
			if err.Error() == "http: request body too large" {
				c.JSON(http.StatusRequestEntityTooLarge, errors.ErrorResponse{
					Error: &errors.APIError{
						Code:    "REQUEST_TOO_LARGE",
						Message: "Request body size exceeds maximum allowed limit",
						Status:  http.StatusRequestEntityTooLarge,
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
