package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// MockRegistry implements core.Registry for testing
type MockRegistry struct {
	liveValues map[string]error
}

func (m *MockRegistry) Set(checkType core.Kind, key string, value error) {
	if m.liveValues == nil {
		m.liveValues = make(map[string]error)
	}
	m.liveValues[key] = value
}

func (m *MockRegistry) Aggregate(ctx context.Context, checkType core.Kind) core.Result {
	return core.Result{OK: true}
}

func (m *MockRegistry) Register(check core.Check) {
	// No-op for testing
}

// MockResult implements core.Result for testing
type MockResult struct {
	OK bool `json:"ok"`
}

func TestModule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("module provides expected dependencies", func(t *testing.T) {
		logger := zap.NewNop()
		v := viper.New()
		v.Set("http.addr", ":0") // Use random port for testing
		reg := &MockRegistry{}

		app := fx.New(
			fx.Provide(func() *zap.Logger { return logger }),
			fx.Provide(func() *viper.Viper { return v }),
			fx.Provide(func() core.Registry { return reg }), // Mock core.Registry
			Module(),
			fx.Invoke(func(e *gin.Engine) {
				// Verify we can get the engine
				assert.NotNil(t, e)
			}),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := app.Start(ctx)
		require.NoError(t, err)

		err = app.Stop(ctx)
		require.NoError(t, err)
	})
}

func TestRegisterHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("registers health endpoints", func(t *testing.T) {
		engine := gin.New()
		v := viper.New()
		reg := &MockRegistry{}

		RegisterHealthRoutes(engine, reg, v)

		// Test /healthz
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)
		engine.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Test /livez
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/livez", nil)
		engine.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("registers info endpoint with build info", func(t *testing.T) {
		engine := gin.New()
		v := viper.New()
		reg := &MockRegistry{}

		info := BuildInfo{
			Version: "v1.0.0",
			Commit:  "abc123",
			BuiltAt: "2025-10-07T10:00:00Z",
		}

		RegisterHealthRoutes(engine, reg, v, WithInfo(info))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/actuator/info", nil)

		engine.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "v1.0.0")
		assert.Contains(t, w.Body.String(), "abc123")
		assert.Contains(t, w.Body.String(), "2025-10-07T10:00:00Z")
	})

	t.Run("registers endpoints with custom base path", func(t *testing.T) {
		engine := gin.New()
		v := viper.New()
		v.Set("http.base_path", "/api")
		reg := &MockRegistry{}

		RegisterHealthRoutes(engine, reg, v)

		// Test health endpoints with base path
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/healthz", nil)
		engine.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/livez", nil)
		engine.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
}
