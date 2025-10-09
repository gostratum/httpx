package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewEngine creates and configures a new Gin engine with middleware and routes
func NewEngine(log *zap.Logger, v *viper.Viper, skip func(string, string) bool, opts ...Option) *gin.Engine {
	// Apply configuration from options
	var s settings
	if basePath := v.GetString("http.base_path"); basePath != "" {
		s.basePath = basePath
	}
	for _, o := range opts {
		o(&s)
	}

	// Create new Gin engine
	e := gin.New()

	// Add core middleware in order
	e.Use(RecoveryMiddleware(log))
	e.Use(RequestIDMiddleware())
	e.Use(LoggingMiddleware(log, skip))

	// Add any extra middleware provided via options
	for _, mw := range s.extraMW {
		e.Use(mw)
	}

	return e
}

// StartServer starts the HTTP server with lifecycle management and graceful shutdown
func StartServer(lc fx.Lifecycle, v *viper.Viper, log *zap.Logger, reg core.Registry, e *gin.Engine, opts ...Option) {
	// Get server address from config
	addr := v.GetString("http.addr")
	if addr == "" {
		addr = ":8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	// Register health routes for Kubernetes probes
	RegisterHealthRoutes(e, reg, v, opts...)

	// Add lifecycle hooks for graceful startup and shutdown
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Set liveness status for HTTP server
			reg.Set(core.Liveness, "http.server", nil)

			// Start server in a goroutine
			go func() {
				log.Info("http: starting server", zap.String("addr", addr))
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("http: server error", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("http: shutting down server")

			// Create shutdown context with timeout
			shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			// Gracefully shutdown the server
			return srv.Shutdown(shutdownCtx)
		},
	})
}
