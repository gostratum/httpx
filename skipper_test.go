package httpx

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkipper(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		method       string
		path         string
		expectedSkip bool
		expectError  bool
	}{
		{
			name: "skip default health endpoints",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
			},
			method:       "GET",
			path:         "/healthz",
			expectedSkip: true,
		},
		{
			name: "skip default liveness endpoints",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
			},
			method:       "GET",
			path:         "/livez",
			expectedSkip: true,
		},
		{
			name: "skip default actuator endpoints",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
			},
			method:       "GET",
			path:         "/actuator/info",
			expectedSkip: true,
		},
		{
			name: "don't skip regular endpoints",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
			},
			method:       "GET",
			path:         "/api/users",
			expectedSkip: false,
		},
		{
			name: "skip custom configured endpoints",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
				Request: RequestConfig{
					Logging: LoggingConfig{
						DisabledURLs: []DisabledURL{
							{
								Method:     "GET",
								URLPattern: "^/metrics$",
							},
						},
					},
				},
			},
			method:       "GET",
			path:         "/metrics",
			expectedSkip: true,
		},
		{
			name: "invalid regex should return error",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
				Request: RequestConfig{
					Logging: LoggingConfig{
						DisabledURLs: []DisabledURL{
							{
								Method:     "GET",
								URLPattern: "[invalid regex",
							},
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "case insensitive method matching",
			config: Config{
				Health: HealthConfig{
					ReadinessPath: "/healthz",
					LivenessPath:  "/livez",
					InfoPath:      "/actuator/info",
					Timeout:       300 * time.Millisecond,
				},
			},
			method:       "get",
			path:         "/healthz",
			expectedSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipper, err := NewSkipper(tt.config)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, skipper)

			result := skipper(tt.method, tt.path)
			assert.Equal(t, tt.expectedSkip, result)
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithBasePath", func(t *testing.T) {
		var s settings
		opt := WithBasePath("/api/v1")
		opt(&s)
		assert.Equal(t, "/api/v1", s.basePath)
	})

	t.Run("WithInfo", func(t *testing.T) {
		var s settings
		info := BuildInfo{
			Version: "v1.0.0",
			Commit:  "abc123",
			BuiltAt: "2025-10-07",
		}
		opt := WithInfo(info)
		opt(&s)
		require.NotNil(t, s.info)
		assert.Equal(t, "v1.0.0", s.info.Version)
		assert.Equal(t, "abc123", s.info.Commit)
		assert.Equal(t, "2025-10-07", s.info.BuiltAt)
	})

	t.Run("WithMiddleware", func(t *testing.T) {
		var s settings
		middleware1 := func(c *gin.Context) {}
		middleware2 := func(c *gin.Context) {}

		opt := WithMiddleware(middleware1, middleware2)
		opt(&s)

		assert.Len(t, s.extraMW, 2)
	})
}
