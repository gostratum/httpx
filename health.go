package httpx

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core"
	"github.com/spf13/viper"
)

// RegisterHealthRoutes registers the actuator-style health endpoints
// These are essential for Kubernetes liveness and readiness probes
func RegisterHealthRoutes(e *gin.Engine, reg core.Registry, v *viper.Viper, opts ...Option) {
	// Apply settings from options
	var s settings
	if basePath := v.GetString("http.base_path"); basePath != "" {
		s.basePath = basePath
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
	healthzPath := v.GetString("http.health.readiness_path")
	if healthzPath == "" {
		healthzPath = "/healthz"
	}

	livezPath := v.GetString("http.health.liveness_path")
	if livezPath == "" {
		livezPath = "/livez"
	}

	infoPath := v.GetString("http.health.info_path")
	if infoPath == "" {
		infoPath = "/actuator/info"
	}

	// Get configurable timeout
	healthTimeout := v.GetDuration("http.health.timeout")
	if healthTimeout == 0 {
		healthTimeout = 300 * time.Millisecond
	}

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
