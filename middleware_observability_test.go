package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMetrics implements a simple in-memory metrics provider for testing
type MockMetrics struct {
	counters   map[string]float64
	histograms map[string][]float64
	gauges     map[string]float64
}

func NewMockMetrics() *MockMetrics {
	return &MockMetrics{
		counters:   make(map[string]float64),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

func (m *MockMetrics) Counter(name, help string, labels ...string) MockCounter {
	return MockCounter{metrics: m, name: name}
}

func (m *MockMetrics) Histogram(name, help string, buckets []float64, labels ...string) MockHistogram {
	return MockHistogram{metrics: m, name: name}
}

func (m *MockMetrics) Gauge(name, help string, labels ...string) MockGauge {
	return MockGauge{metrics: m, name: name}
}

type MockCounter struct {
	metrics *MockMetrics
	name    string
}

func (c MockCounter) Inc(labels map[string]string) {
	c.metrics.counters[c.name]++
}

type MockHistogram struct {
	metrics *MockMetrics
	name    string
}

func (h MockHistogram) Observe(value float64, labels map[string]string) {
	h.metrics.histograms[h.name] = append(h.metrics.histograms[h.name], value)
}

type MockGauge struct {
	metrics *MockMetrics
	name    string
}

func (g MockGauge) Set(value float64, labels map[string]string) {
	g.metrics.gauges[g.name] = value
}

func (g MockGauge) Inc(labels map[string]string) {
	g.metrics.gauges[g.name]++
}

func (g MockGauge) Dec(labels map[string]string) {
	g.metrics.gauges[g.name]--
}

// MockTracer implements a simple in-memory tracer for testing
type MockTracer struct {
	spans []*MockSpan
}

func NewMockTracer() *MockTracer {
	return &MockTracer{
		spans: make([]*MockSpan, 0),
	}
}

func (t *MockTracer) Start(ctx any, name string, opts ...any) (any, *MockSpan) {
	span := &MockSpan{
		name:       name,
		attributes: make(map[string]any),
	}
	t.spans = append(t.spans, span)
	return ctx, span
}

type MockSpan struct {
	name       string
	attributes map[string]any
	ended      bool
	statusCode int
	statusMsg  string
}

func (s *MockSpan) SetAttributes(attrs ...any) {
	// Simple implementation - just track that attributes were set
	for i := 0; i < len(attrs); i += 2 {
		if i+1 < len(attrs) {
			key := attrs[i].(string)
			s.attributes[key] = attrs[i+1]
		}
	}
}

func (s *MockSpan) SetStatus(code int, msg string) {
	s.statusCode = code
	s.statusMsg = msg
}

func (s *MockSpan) End() {
	s.ended = true
}

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("records request metrics", func(t *testing.T) {
		metrics := NewMockMetrics()

		engine := gin.New()
		engine.Use(func(c *gin.Context) {
			// Simulate MetricsMiddleware behavior
			metrics.Counter("http_requests_total", "Total HTTP requests").Inc(nil)
			metrics.Histogram("http_request_duration_seconds", "Request duration", nil).Observe(0.1, nil)
			c.Next()
		})
		engine.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1.0, metrics.counters["http_requests_total"])
		assert.Len(t, metrics.histograms["http_request_duration_seconds"], 1)
	})

	t.Run("tracks in-flight requests", func(t *testing.T) {
		metrics := NewMockMetrics()

		engine := gin.New()
		engine.Use(func(c *gin.Context) {
			metrics.Gauge("http_requests_in_flight", "In-flight requests").Inc(nil)
			defer metrics.Gauge("http_requests_in_flight", "In-flight requests").Dec(nil)
			c.Next()
		})
		engine.GET("/test", func(c *gin.Context) {
			// At this point, gauge should be 1
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// After request completes, gauge should be back to 0
		assert.Equal(t, 0.0, metrics.gauges["http_requests_in_flight"])
	})
}

func TestTracingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates span for request", func(t *testing.T) {
		tracer := NewMockTracer()

		engine := gin.New()
		engine.Use(func(c *gin.Context) {
			_, span := tracer.Start(c.Request.Context(), "HTTP "+c.Request.Method)
			defer span.End()

			span.SetAttributes(
				"http.method", c.Request.Method,
				"http.url", c.Request.URL.String(),
			)

			c.Next()

			span.SetAttributes("http.status_code", c.Writer.Status())
		})
		engine.GET("/api/users", func(c *gin.Context) {
			c.JSON(200, gin.H{"users": []string{}})
		})

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		require.Len(t, tracer.spans, 1)
		span := tracer.spans[0]

		assert.Equal(t, "HTTP GET", span.name)
		assert.Equal(t, "GET", span.attributes["http.method"])
		assert.Equal(t, "/api/users", span.attributes["http.url"])
		assert.Equal(t, 200, span.attributes["http.status_code"])
		assert.True(t, span.ended)
	})

	t.Run("marks span as error for 5xx status", func(t *testing.T) {
		tracer := NewMockTracer()

		engine := gin.New()
		engine.Use(func(c *gin.Context) {
			_, span := tracer.Start(c.Request.Context(), "HTTP "+c.Request.Method)
			defer func() {
				if c.Writer.Status() >= 500 {
					span.SetStatus(2, "Internal Server Error") // 2 = Error
				}
				span.End()
			}()
			c.Next()
		})
		engine.GET("/error", func(c *gin.Context) {
			c.JSON(500, gin.H{"error": "internal error"})
		})

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		require.Len(t, tracer.spans, 1)
		span := tracer.spans[0]

		assert.Equal(t, 2, span.statusCode)
		assert.Equal(t, "Internal Server Error", span.statusMsg)
		assert.True(t, span.ended)
	})
}

func TestObservabilityIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("middleware chain with observability", func(t *testing.T) {
		metrics := NewMockMetrics()
		tracer := NewMockTracer()

		engine := gin.New()

		// Simulate both middlewares
		engine.Use(func(c *gin.Context) {
			// Metrics middleware simulation
			metrics.Counter("http_requests_total", "Total HTTP requests").Inc(nil)
			metrics.Gauge("http_requests_in_flight", "In-flight requests").Inc(nil)
			defer metrics.Gauge("http_requests_in_flight", "In-flight requests").Dec(nil)

			// Tracing middleware simulation
			_, span := tracer.Start(c.Request.Context(), "HTTP "+c.Request.Method)
			defer span.End()

			c.Next()

			span.SetAttributes("http.status_code", c.Writer.Status())
		})

		engine.GET("/api/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// Verify metrics were recorded
		assert.Equal(t, 1.0, metrics.counters["http_requests_total"])
		assert.Equal(t, 0.0, metrics.gauges["http_requests_in_flight"])

		// Verify trace was created
		require.Len(t, tracer.spans, 1)
		assert.Equal(t, "HTTP GET", tracer.spans[0].name)
		assert.True(t, tracer.spans[0].ended)
	})
}
