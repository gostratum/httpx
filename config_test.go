package httpx

import (
	"testing"
	"time"

	"github.com/gostratum/core/configx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Run("loads config with defaults", func(t *testing.T) {
		loader := configx.New()
		cfg, err := NewConfig(loader)
		require.NoError(t, err)

		// Check defaults
		assert.Equal(t, ":8080", cfg.Addr)
		assert.Equal(t, "/healthz", cfg.Health.ReadinessPath)
		assert.Equal(t, "/livez", cfg.Health.LivenessPath)
		assert.Equal(t, "/actuator/info", cfg.Health.InfoPath)
		assert.Equal(t, 300*time.Millisecond, cfg.Health.Timeout)
	})

	t.Run("Prefix returns 'http'", func(t *testing.T) {
		cfg := Config{}
		assert.Equal(t, "http", cfg.Prefix())
	})
}

func TestConfigSummary(t *testing.T) {
	cfg := Config{
		Addr:     ":8080",
		BasePath: "/api",
		Health: HealthConfig{
			ReadinessPath: "/healthz",
			LivenessPath:  "/livez",
			Timeout:       300 * time.Millisecond,
		},
	}

	summary := cfg.ConfigSummary()
	assert.Equal(t, ":8080", summary["addr"])
	assert.Equal(t, "/api", summary["base_path"])
	assert.Equal(t, "/healthz", summary["readiness_path"])
	assert.Equal(t, "/livez", summary["liveness_path"])
	assert.Equal(t, 300*time.Millisecond, summary["health_timeout"])
}
