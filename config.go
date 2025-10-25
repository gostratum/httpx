package httpx

import (
	"time"

	"github.com/gostratum/core/configx"
)

// Config contains configuration for the HTTP server module
type Config struct {
	// Addr is the address to listen on (e.g., ":8080" or "localhost:8080")
	Addr string `mapstructure:"addr" default:":8080"`

	// BasePath is the base path for all routes (e.g., "/api/v1")
	BasePath string `mapstructure:"base_path"`

	// Health contains health check endpoint configuration
	Health HealthConfig `mapstructure:"health"`

	// Request contains request-specific configuration
	Request RequestConfig `mapstructure:"request"`
}

// Prefix enables configx.Bind
func (Config) Prefix() string { return "http" }

// HealthConfig contains health check endpoint configuration
type HealthConfig struct {
	// ReadinessPath is the path for readiness checks (default: /healthz)
	ReadinessPath string `mapstructure:"readiness_path" default:"/healthz"`

	// LivenessPath is the path for liveness checks (default: /livez)
	LivenessPath string `mapstructure:"liveness_path" default:"/livez"`

	// InfoPath is the path for info endpoint (default: /actuator/info)
	InfoPath string `mapstructure:"info_path" default:"/actuator/info"`

	// Timeout is the maximum duration for health checks
	Timeout time.Duration `mapstructure:"timeout" default:"300ms"`
}

// RequestConfig contains request-specific configuration
type RequestConfig struct {
	// Logging contains request logging configuration
	Logging LoggingConfig `mapstructure:"logging"`
}

// LoggingConfig contains request logging configuration
type LoggingConfig struct {
	// DisabledURLs are URL patterns to skip in request logging
	DisabledURLs []DisabledURL `mapstructure:"disabled_urls"`
}

// NewConfig creates a new Config from the configuration loader
func NewConfig(loader configx.Loader) (Config, error) {
	var cfg Config
	if err := loader.Bind(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ConfigSummary returns a compact diagnostic map for HTTP configuration
func (c Config) ConfigSummary() map[string]any {
	return map[string]any{
		"addr":           c.Addr,
		"base_path":      c.BasePath,
		"readiness_path": c.Health.ReadinessPath,
		"liveness_path":  c.Health.LivenessPath,
		"health_timeout": c.Health.Timeout,
	}
}
