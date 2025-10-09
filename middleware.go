package httpx

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ctxKey is used as a key for context values
type ctxKey string

const requestIDKey ctxKey = "rid"

// RequestIDMiddleware adds X-Request-ID header and context value to requests
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate a new one
		id := c.Request.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}

		// Add request ID to context and response header
		ctx := context.WithValue(c.Request.Context(), requestIDKey, id)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Request-ID", id)

		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests using Zap logger with configurable skipping
func LoggingMiddleware(log *zap.Logger, skip func(method, path string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging if configured
		if skip != nil && skip(c.Request.Method, c.FullPath()) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		// Extract request ID from response header (set by RequestIDMiddleware)
		requestID := c.Writer.Header().Get("X-Request-ID")

		// Log the request
		log.Info("http",
			zap.String("rid", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("dur", time.Since(start)),
		)
	}
}

// RecoveryMiddleware handles panics and converts them to 500 errors
func RecoveryMiddleware(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		requestID := c.Writer.Header().Get("X-Request-ID")
		log.Error("http panic recovered",
			zap.String("rid", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Any("error", err),
		)
		c.AbortWithStatus(500)
	})
}
