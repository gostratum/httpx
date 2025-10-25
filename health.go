package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core"
)

// registerHealthRoutes registers the actuator-style health endpoints internally
// These are essential for Kubernetes liveness and readiness probes
func registerHealthRoutes(e *gin.Engine, reg core.Registry, cfg Config, opts ...Option) {
	// Apply settings from options
	var s settings
	if cfg.BasePath != "" {
		s.basePath = cfg.BasePath
	}
	for _, o := range opts {
		o(&s)
	}

	// Determine base path
	base := s.basePath
	if base == "" {
		base = "/"
	}

	// Get configurable endpoint paths
	healthzPath := cfg.Health.ReadinessPath
	livezPath := cfg.Health.LivenessPath
	infoPath := cfg.Health.InfoPath

	// Get configurable timeout
	healthTimeout := cfg.Health.Timeout

	// Create route group at base path
	g := e.Group(strings.TrimRight(base, "/"))

	// Readiness check - aggregates all readiness checks from registry
	// Used by Kubernetes readiness probes
	g.GET(healthzPath, func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthTimeout)
		defer cancel()

		result := reg.Aggregate(ctx, core.Readiness)

		statusCode := http.StatusOK
		if !result.OK {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, result)
	})

	// Liveness check - aggregates all liveness checks from registry
	// Used by Kubernetes liveness probes
	g.GET(livezPath, func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthTimeout)
		defer cancel()

		result := reg.Aggregate(ctx, core.Liveness)

		statusCode := http.StatusOK
		if !result.OK {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, result)
	})

	// Optional info endpoint if WithInfo was provided
	if s.info != nil {
		g.GET(infoPath, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": s.info.Version,
				"commit":  s.info.Commit,
				"builtAt": s.info.BuiltAt,
			})
		})
	}
}
