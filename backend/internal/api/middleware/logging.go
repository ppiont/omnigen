package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Logger creates a logging middleware
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate trace ID
		traceID := uuid.New().String()
		c.Set("trace_id", traceID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Log request start
		logger.Info("Request started",
			zap.String("trace_id", traceID),
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
			zap.String("trace_id", traceID),
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
					zap.String("trace_id", traceID),
					zap.Error(e.Err),
				)
			}
		}
	}
}
