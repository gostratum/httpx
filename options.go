package httpx

import "github.com/gin-gonic/gin"

// moduleConfig holds programmatic configuration for the HTTP module
// Simple configuration values (strings, bools, numbers) should be in config YAML instead
type moduleConfig struct {
	extraMW []gin.HandlerFunc // Go functions - cannot be in YAML
	info    *BuildInfo        // Build metadata - could be programmatic or config
}

// Option configures the HTTP module
type Option func(*moduleConfig)

// WithMiddleware adds custom middleware to the Gin engine
// This is programmatic-only as it requires Go functions
func WithMiddleware(mw ...gin.HandlerFunc) Option {
	return func(s *moduleConfig) {
		s.extraMW = append(s.extraMW, mw...)
	}
}

// WithInfo enables the /actuator/info endpoint with build information
// This can be programmatic or from config, keeping as option for convenience
func WithInfo(b BuildInfo) Option {
	return func(s *moduleConfig) {
		s.info = &b
	}
}

// BuildInfo contains build metadata for the /actuator/info endpoint
type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	BuiltAt string `json:"builtAt"`
}
