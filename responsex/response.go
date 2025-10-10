package responsex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard API response wrapper.
type Envelope[T any] struct {
	Ok         bool        `json:"ok"`
	Data       T           `json:"data,omitempty"`
	Error      *APIError   `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
}

// Error represents an error in the unified response.
// APIError represents an error in the unified response.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details []ErrDetail `json:"details,omitempty"`
}

// ErrDetail gives additional context about an error.
type ErrDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

// Pagination is defined in pagination.go

// Meta holds request/response metadata.
type Meta struct {
	RequestID  string     `json:"request_id,omitempty"`
	Timestamp  time.Time  `json:"timestamp,omitempty"`
	DurationMS int64      `json:"duration_ms,omitempty"`
	Server     string     `json:"server,omitempty"`
	RateLimit  *RateLimit `json:"rate_limit,omitempty"`
}

// RateLimit mirrors common rate-limit metadata.
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

// WithETag sets ETag header on the gin context.
func WithETag(c *gin.Context, etag string) {
	if etag == "" {
		return
	}
	c.Header("ETag", etag)
}

// WithRateLimit sets rate-limit headers and returns a RateLimit meta.
func WithRateLimit(c *gin.Context, limit, remaining int, reset time.Time) *RateLimit {
	c.Header("X-RateLimit-Limit", intToStr(limit))
	c.Header("X-RateLimit-Remaining", intToStr(remaining))
	c.Header("X-RateLimit-Reset", int64ToStr(reset.Unix()))
	rl := &RateLimit{Limit: limit, Remaining: remaining, Reset: reset}
	c.Set(ctxRateLimitKey, rl)
	return rl
}

// OK sends a 200 response with the envelope.
func OK(c *gin.Context, data any, pg *Pagination) {
	sendEnvelope(c, http.StatusOK, true, data, nil, pg)
}

// Created sends a 201 response with Location header and envelope.
func Created(c *gin.Context, location string, data any) {
	if location != "" {
		c.Header("Location", location)
	}
	sendEnvelope(c, http.StatusCreated, true, data, nil, nil)
}

// Error sends an error envelope with provided HTTP status.
// Error sends an error envelope with provided HTTP status.
func Error(c *gin.Context, status int, code, message string, details []ErrDetail) {
	e := &APIError{Code: code, Message: message, Details: details}
	sendEnvelope(c, status, false, nil, e, nil)
}

// sendEnvelope builds and writes the JSON response envelope.
func sendEnvelope(c *gin.Context, status int, ok bool, data any, err *APIError, pg *Pagination) {
	env := Envelope[any]{
		Ok:         ok,
		Data:       data,
		Error:      err,
		Pagination: pg,
	}

	// Attach meta if available in context
	if v, exists := c.Get("responsex.start_time"); exists {
		if start, ok := v.(time.Time); ok {
			meta := &Meta{Timestamp: time.Now()}
			if rid, ok := c.Get("X-Request-Id"); ok {
				if rs, ok := rid.(string); ok {
					meta.RequestID = rs
				}
			}
			meta.DurationMS = time.Since(start).Milliseconds()
			if sv, ok := c.Get("responsex.server_version"); ok {
				if ss, ok := sv.(string); ok {
					meta.Server = ss
				}
			}
			if rl, ok := c.Get("responsex.rate_limit"); ok {
				if rlm, ok := rl.(*RateLimit); ok {
					meta.RateLimit = rlm
				}
			}
			env.Meta = meta
		}
	}

	c.Status(status)
	c.Header("Content-Type", "application/json; charset=utf-8")

	// Use json.Encoder to avoid html escaping
	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(env)
}

func intToStr(i int) string {
	return fmt.Sprintf("%d", i)
}

func int64ToStr(i int64) string {
	return fmt.Sprintf("%d", i)
}
