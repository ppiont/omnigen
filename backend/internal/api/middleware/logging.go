package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger creates a logging middleware
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for noisy endpoints (health checks, job polling)
		path := c.Request.URL.Path
		method := c.Request.Method

		// Skip health checks and GET job status polling
		if path == "/health" || (method == "GET" && len(path) > 14 && path[:14] == "/api/v1/jobs/") {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Log request start
		logger.Info("Request started",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request completion
		logger.Info("Request completed",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Int("response_size", c.Writer.Size()),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request error",
					zap.String("method", method),
					zap.String("path", path),
					zap.Error(e.Err),
				)
			}
		}
	}
}
