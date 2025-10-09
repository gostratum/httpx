package httpx

import "github.com/gin-gonic/gin"

// settings holds the configuration for the HTTP module
type settings struct {
	basePath string
	extraMW  []gin.HandlerFunc
	info     *BuildInfo
}

// Option configures the HTTP module
type Option func(*settings)

// WithMiddleware adds custom middleware to the Gin engine
func WithMiddleware(mw ...gin.HandlerFunc) Option {
	return func(s *settings) {
		s.extraMW = append(s.extraMW, mw...)
	}
}

// WithBasePath overrides the base path for all routes
func WithBasePath(path string) Option {
	return func(s *settings) {
		s.basePath = path
	}
}

// WithInfo enables the /actuator/info endpoint with build information
func WithInfo(b BuildInfo) Option {
	return func(s *settings) {
		s.info = &b
	}
}

// BuildInfo contains build metadata for the /actuator/info endpoint
type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	BuiltAt string `json:"builtAt"`
}
