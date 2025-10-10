package responsex

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	ctxStartTimeKey     = "responsex.start_time"
	ctxServerVersionKey = "responsex.server_version"
	ctxRateLimitKey     = "responsex.rate_limit"
)

// MetaMiddleware injects headers and metadata into each request.
// It records start_time in the gin.Context so duration_ms can be computed.
func MetaMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Date header
		c.Header("Date", time.Now().UTC().Format(time.RFC1123))

		// X-Request-Id: prefer incoming header, otherwise generate
		rid := c.Request.Header.Get("X-Request-Id")
		if strings.TrimSpace(rid) == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set("X-Request-Id", rid)

		// Optional traceparent: propagate if present
		if tp := c.Request.Header.Get("traceparent"); tp != "" {
			c.Writer.Header().Set("traceparent", tp)
		}

		// X-Server-Version
		if version != "" {
			c.Writer.Header().Set("X-Server-Version", version)
			c.Set(ctxServerVersionKey, version)
		}

		// Record start time
		c.Set(ctxStartTimeKey, time.Now())

		// proceed
		c.Next()

		// After request: ensure request id header available in context too
		c.Set("X-Request-Id", c.Writer.Header().Get("X-Request-Id"))
	}
}

// WithLogger is an optional helper to attach a zap logger into context for middleware
func WithLogger(c *gin.Context, logger *zap.Logger) {
	c.Set("responsex.logger", logger)
}
