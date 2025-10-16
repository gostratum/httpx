package httpx

import (
	"github.com/gin-gonic/gin"
	"github.com/gostratum/core"
	"github.com/gostratum/metricsx"
	"github.com/gostratum/tracingx"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ObservabilityParams contains optional observability dependencies
type ObservabilityParams struct {
	fx.In

	Metrics metricsx.Metrics `optional:"true"`
	Tracer  tracingx.Tracer  `optional:"true"`
}

// Module returns an fx.Option that wires the HTTP server with default configuration.
// Callers may pass Option overrides to customize behavior.
//
// Observability is automatically enabled if metricsx and/or tracingx modules are present.
// The middleware will be no-op if the modules are not available.
func Module(opts ...Option) fx.Option {
	return fx.Options(
		// Provide the log skipper function
		fx.Provide(NewSkipper),

		// Provide the Gin engine with all dependencies and options
		fx.Provide(func(log *zap.Logger, v *viper.Viper, skip func(string, string) bool, obs ObservabilityParams) *gin.Engine {
			return NewEngineWithObservability(log, v, skip, obs, opts...)
		}),

		// Start the HTTP server as part of the application lifecycle
		fx.Invoke(func(lc fx.Lifecycle, v *viper.Viper, log *zap.Logger, reg core.Registry, e *gin.Engine) {
			StartServer(lc, v, log, reg, e, opts...)
		}),
	)
}
