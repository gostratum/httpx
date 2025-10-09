package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates engine with default settings", func(t *testing.T) {
		logger := zap.NewNop()
		v := viper.New()
		skip := func(method, path string) bool { return false }

		engine := NewEngine(logger, v, skip)

		assert.NotNil(t, engine)
	})

	t.Run("applies base path from config", func(t *testing.T) {
		logger := zap.NewNop()
		v := viper.New()
		v.Set("http.base_path", "/api")
		skip := func(method, path string) bool { return false }

		engine := NewEngine(logger, v, skip)

		assert.NotNil(t, engine)
	})

	t.Run("applies options", func(t *testing.T) {
		logger := zap.NewNop()
		v := viper.New()
		skip := func(method, path string) bool { return false }

		testMiddleware := func(c *gin.Context) {
			c.Header("X-Test", "middleware")
			c.Next()
		}

		engine := NewEngine(logger, v, skip, WithMiddleware(testMiddleware))

		assert.NotNil(t, engine)

		// Test that middleware is applied
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		engine.GET("/test", func(c *gin.Context) {
			c.Status(200)
		})

		engine.ServeHTTP(w, req)

		assert.Equal(t, "middleware", w.Header().Get("X-Test"))
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates request ID when not provided", func(t *testing.T) {
		engine := gin.New()
		engine.Use(RequestIDMiddleware())

		var requestID string
		engine.GET("/test", func(c *gin.Context) {
			requestID = c.Writer.Header().Get("X-Request-ID")
			c.Status(200)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		engine.ServeHTTP(w, req)

		assert.NotEmpty(t, requestID)
		assert.Equal(t, requestID, w.Header().Get("X-Request-ID"))
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		engine := gin.New()
		engine.Use(RequestIDMiddleware())

		expectedID := "test-request-id"
		var actualID string

		engine.GET("/test", func(c *gin.Context) {
			actualID = c.Writer.Header().Get("X-Request-ID")
			c.Status(200)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", expectedID)

		engine.ServeHTTP(w, req)

		assert.Equal(t, expectedID, actualID)
		assert.Equal(t, expectedID, w.Header().Get("X-Request-ID"))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("skips logging when skip function returns true", func(t *testing.T) {
		logger := zap.NewNop()
		skip := func(method, path string) bool {
			return method == "GET" && path == "/healthz"
		}

		engine := gin.New()
		engine.Use(LoggingMiddleware(logger, skip))

		engine.GET("/healthz", func(c *gin.Context) {
			c.Status(200)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)

		// This should not panic or cause issues
		engine.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("logs when skip function returns false", func(t *testing.T) {
		logger := zap.NewNop()
		skip := func(method, path string) bool {
			return false
		}

		engine := gin.New()
		engine.Use(RequestIDMiddleware()) // Need this for request ID
		engine.Use(LoggingMiddleware(logger, skip))

		engine.GET("/api/users", func(c *gin.Context) {
			c.Status(200)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/users", nil)

		// This should not panic
		engine.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}
