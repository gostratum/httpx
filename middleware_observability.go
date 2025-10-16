package httpx

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/metricsx"
	"github.com/gostratum/tracingx"
)

// MetricsMiddleware instruments HTTP requests with metrics if metricsx is available
func MetricsMiddleware(metrics metricsx.Metrics) gin.HandlerFunc {
	// Create metric collectors
	requestCounter := metrics.Counter(
		"http_requests_total",
		metricsx.WithHelp("Total number of HTTP requests"),
		metricsx.WithLabels("method", "path", "status"),
	)

	requestDuration := metrics.Histogram(
		"http_request_duration_seconds",
		metricsx.WithHelp("HTTP request duration in seconds"),
		metricsx.WithLabels("method", "path", "status"),
		metricsx.WithBuckets(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
	)

	requestSize := metrics.Histogram(
		"http_request_size_bytes",
		metricsx.WithHelp("HTTP request size in bytes"),
		metricsx.WithLabels("method", "path"),
		metricsx.WithBuckets(100, 1000, 10000, 100000, 1000000, 10000000),
	)

	responseSize := metrics.Histogram(
		"http_response_size_bytes",
		metricsx.WithHelp("HTTP response size in bytes"),
		metricsx.WithLabels("method", "path", "status"),
		metricsx.WithBuckets(100, 1000, 10000, 100000, 1000000, 10000000),
	)

	activeRequests := metrics.Gauge(
		"http_requests_in_flight",
		metricsx.WithHelp("Current number of HTTP requests being processed"),
	)

	return func(c *gin.Context) {
		start := time.Now()

		// Increment active requests
		activeRequests.Inc()
		defer activeRequests.Dec()

		// Record request size
		if c.Request.ContentLength > 0 {
			requestSize.Observe(float64(c.Request.ContentLength), c.Request.Method, c.FullPath())
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		requestCounter.Inc(c.Request.Method, c.FullPath(), status)
		requestDuration.Observe(duration, c.Request.Method, c.FullPath(), status)
		responseSize.Observe(float64(c.Writer.Size()), c.Request.Method, c.FullPath(), status)
	}
}

// TracingMiddleware instruments HTTP requests with distributed tracing if tracingx is available
func TracingMiddleware(tracer tracingx.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from incoming request headers
		ctx, err := tracer.Extract(c.Request.Context(), c.Request.Header)
		if err != nil {
			// If extraction fails, use original context
			ctx = c.Request.Context()
		}

		// Start a new span for this request
		ctx, span := tracer.Start(ctx, c.FullPath(),
			tracingx.WithSpanKind(tracingx.SpanKindServer),
			tracingx.WithAttributes(map[string]interface{}{
				"http.method":      c.Request.Method,
				"http.url":         c.Request.URL.String(),
				"http.host":        c.Request.Host,
				"http.scheme":      c.Request.URL.Scheme,
				"http.user_agent":  c.Request.UserAgent(),
				"http.request_id":  c.Writer.Header().Get("X-Request-ID"),
				"http.remote_addr": c.ClientIP(),
			}),
		)
		defer span.End()

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Inject trace context into response headers for downstream services
		_ = tracer.Inject(ctx, c.Writer.Header())

		// Process request
		c.Next()

		// Set response attributes
		status := c.Writer.Status()
		span.SetTag("http.status_code", status)
		span.SetTag("http.response_size", c.Writer.Size())

		// Mark span as error if status >= 500
		if status >= 500 {
			span.SetTag("error", true)
			if len(c.Errors) > 0 {
				span.SetError(c.Errors.Last().Err)
				for _, err := range c.Errors {
					span.LogFields(tracingx.Field{
						Key:   "error.message",
						Value: err.Error(),
					})
				}
			}
		}

		// Add trace IDs to response headers for debugging
		c.Writer.Header().Set("X-Trace-ID", span.TraceID())
		c.Writer.Header().Set("X-Span-ID", span.SpanID())
	}
}
