# httpx

A thin, reusable HTTP layer for Go applications built on top of `github.com/gostratum/core`. This module provides an fx-first composition approach for HTTP servers with built-in health endpoints, request logging, and middleware support.

## Features

- **Fx-first composition**: No globals, clean dependency injection with `go.uber.org/fx`
- **Typed configuration**: Uses `core/configx` with structured config following framework conventions
- **Universal health endpoints**: Built-in `/healthz` (readiness) and `/livez` (liveness) for all app types
- **Optional info endpoint**: `/actuator/info` with build metadata when enabled
- **Config-driven defaults**: HTTP configuration with sensible defaults
- **Request logging controls**: Pattern-based log skipping with regex support
- **Lifecycle management**: Graceful startup and shutdown via fx.Lifecycle
- **Built-in middleware**: Request ID, Zap logging, and recovery middleware
- **Kubernetes ready**: Perfect for web apps, workers, and migration jobs
- **Composable**: Easy integration with existing fx applications

## Installation

```bash
go get github.com/gostratum/httpx
```

## Quick Start

```go
package main

import (
    "github.com/gostratum/core"
    "github.com/gostratum/httpx"
    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
)

func main() {
    app := core.New(
        httpx.Module(
            httpx.WithBasePath("/api"),
            httpx.WithInfo(httpx.BuildInfo{
                Version: "v1.0.0",
                Commit:  "abcd123",
                BuiltAt: "2025-10-07T10:00:00Z",
            }),
        ),
        fx.Invoke(func(e *gin.Engine) {
            e.GET("/api/hello", func(c *gin.Context) {
                c.JSON(200, gin.H{"message": "Hello, World!"})
            })
        }),
    )
    
    app.Run()
}
```

## Configuration

The module uses typed configuration following the GoStratum framework pattern. Configuration is loaded via `core/configx` under the `http` prefix.

**Configuration Precedence (highest to lowest):**
1. Environment variables: `STRATUM_HTTP_*`
2. Environment-specific config file: `{APP_ENV}.yaml`
3. Base config file: `base.yaml`
4. Struct tag defaults

### YAML Configuration

```yaml
http:
  addr: ":8080"           # Server listen address
  base_path: "/"          # Base path for all routes
  
  health:                 # Health endpoint configuration
    readiness_path: "/healthz"      # Readiness endpoint path
    liveness_path: "/livez"         # Liveness endpoint path  
    info_path: "/actuator/info"     # Info endpoint path
    timeout: "300ms"                # Health check timeout
    
  request:
    logging:
      disabled_urls:      # URLs to skip in request logging
        - method: "GET"
          urlPattern: "^/metrics$"
        - method: "POST"
          urlPattern: "^/webhook/.*"
```

### Environment Variables

All configuration can be overridden with environment variables using the `STRATUM_HTTP_` prefix:

```bash
export STRATUM_HTTP_ADDR=":9090"
export STRATUM_HTTP_BASE_PATH="/api"
export STRATUM_HTTP_HEALTH_READINESS_PATH="/ready"
export STRATUM_HTTP_HEALTH_LIVENESS_PATH="/alive"
export STRATUM_HTTP_HEALTH_TIMEOUT="500ms"
```

### Default Configuration

- **Address**: `:8080`
- **Base Path**: `/`
- **Health Endpoints**:
  - Readiness: `/healthz` (configurable via `http.health.readiness_path`)
  - Liveness: `/livez` (configurable via `http.health.liveness_path`)
  - Info: `/actuator/info` (configurable via `http.health.info_path`)
  - Timeout: `300ms` (configurable via `http.health.timeout`)
- **Disabled Logging**: Health and actuator endpoints are excluded from request logs by default

## API Reference

### Module Function

```go
func Module(opts ...Option) fx.Option
```

Returns an fx.Option that can be included in your fx application. The module provides:
- Configured Gin engine with middleware
- HTTP server with lifecycle management
- Built-in health endpoints (`/healthz`, `/livez`) for Kubernetes probes
- Optional info endpoint (`/actuator/info`) when build metadata is provided

### Options

#### WithBasePath

```go
func WithBasePath(path string) Option
```

Overrides the base path for all routes. This takes precedence over the `http.base_path` configuration.

```go
httpx.Module(httpx.WithBasePath("/api/v1"))
```

#### WithMiddleware

```go
func WithMiddleware(mw ...gin.HandlerFunc) Option
```

Adds custom middleware to the Gin engine. Middleware is added after the built-in middleware (recovery, request ID, logging).

```go
httpx.Module(httpx.WithMiddleware(
    middleware.CORS(),
    middleware.RateLimit(),
))
```

#### WithInfo

```go
func WithInfo(info BuildInfo) Option
```

Enables the `/actuator/info` endpoint with build information.

```go
httpx.Module(httpx.WithInfo(httpx.BuildInfo{
    Version: "v1.0.0",
    Commit:  "abc123",
    BuiltAt: "2025-10-07T10:00:00Z",
}))
```

### BuildInfo

```go
type BuildInfo struct {
    Version string `json:"version"`
    Commit  string `json:"commit"`
    BuiltAt string `json:"builtAt"`
}
```

Build information structure for the info endpoint. Typically populated via ldflags during build:

```bash
go build -ldflags="-X main.version=v1.0.0 -X main.commit=abc123 -X main.builtAt=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Health Endpoints

### /healthz (Readiness)

Readiness check endpoint that aggregates all readiness checks from the core registry. Used by Kubernetes readiness probes.

- **Returns 200**: All readiness checks pass (ready to serve traffic)
- **Returns 503**: One or more readiness checks fail (not ready)
- **Timeout**: 300ms

### /livez (Liveness)

Liveness check endpoint that aggregates all liveness checks from the core registry. Used by Kubernetes liveness probes.

- **Returns 200**: All liveness checks pass (application is alive)
- **Returns 503**: One or more liveness checks fail (should restart)
- **Timeout**: 300ms

### /actuator/info (Optional)

Returns build information when enabled with `WithInfo()`.

```json
{
  "version": "v1.0.0",
  "commit": "abc123",
  "builtAt": "2025-10-07T10:00:00Z"
}
```

## Middleware

The module includes several built-in middleware components:

### Request ID Middleware

- Extracts or generates X-Request-ID header
- Adds request ID to response headers
- Makes request ID available in request context

### Logging Middleware

- Logs HTTP requests using Zap
- Includes request ID, method, path, status code, and duration
- Respects log skipping configuration
- Can be disabled for specific URL patterns

### Recovery Middleware

- Recovers from panics in handlers
- Logs panic details with request context
- Returns 500 status code for panics

## Request Log Skipping

You can configure URL patterns to skip request logging:

```yaml
http:
  request:
    logging:
      disabled_urls:
        - method: "GET"
          urlPattern: "^/metrics$"
        - method: ""              # Empty method matches all methods
          urlPattern: "^/static/.*"
```

**Default Skipped URLs:**
- `GET /healthz`
- `GET /livez`  
- `GET /actuator/*`

## Integration with Core

The httpx module is designed to work seamlessly with `github.com/gostratum/core`:

- Uses core's `configx.Loader` for typed configuration (following framework pattern)
- Exposes core's Registry health checks via HTTP endpoints
- Sets liveness status when HTTP server starts: `reg.Set(core.Liveness, "http.server", nil)`
- Follows core's fx-first architecture patterns
- Respects health endpoint log skipping by default
- **No direct viper access**: Uses typed Config struct with validation

## Examples

### Web Application

```go
// Full HTTP server with business endpoints + health endpoints
app := core.New(
    httpx.Module(
        httpx.WithBasePath("/api"),
        httpx.WithInfo(httpx.BuildInfo{
            Version: "v1.0.0",
            Commit:  "abc123", 
            BuiltAt: "2025-10-07T10:00:00Z",
        }),
    ),
    fx.Invoke(func(e *gin.Engine) {
        api := e.Group("/api/v1")
        api.GET("/users", getUsersHandler)
        api.POST("/users", createUserHandler)
    }),
)
```

### Worker Application

```go
// Health endpoints only (no business endpoints needed)
app := core.New(
    httpx.Module(), // Provides /healthz, /livez for Kubernetes
    workerModule(), // Your worker logic
    fx.Invoke(func(reg core.Registry) {
        // Worker can update health status
        reg.Set(core.Readiness, "worker.ready", nil)
    }),
)
```

### Migration Job

```go
// Health endpoints + migration lifecycle signaling
app := core.New(
    httpx.Module(),
    migrationModule(),
    fx.Invoke(func(reg core.Registry) {
        // Signal migration progress via health endpoints
        
        // Before migration starts
        reg.Set(core.Readiness, "migration.ready", nil)
        
        // During migration (not ready for traffic)
        reg.Set(core.Readiness, "migration.running", errors.New("migration in progress"))
        
        // After successful migration
        reg.Set(core.Readiness, "migration.complete", nil)
        
        // Application can exit gracefully, K8s will see healthy status
    }),
)
```

### Custom Health Endpoints

```go
// Configure custom health endpoint paths via environment or config files
// Environment variables:
// export STRATUM_HTTP_HEALTH_READINESS_PATH="/health/ready"
// export STRATUM_HTTP_HEALTH_LIVENESS_PATH="/health/alive"
// export STRATUM_HTTP_HEALTH_INFO_PATH="/info"
// export STRATUM_HTTP_HEALTH_TIMEOUT="500ms"

// Or via YAML config file (base.yaml or dev.yaml/prod.yaml)
// http:
//   health:
//     readiness_path: "/health/ready"
//     liveness_path: "/health/alive"
//     info_path: "/info"
//     timeout: "500ms"

app := core.New(
    httpx.Module(),
    // Configuration is automatically loaded from configx
)
```

### Custom Configuration

```go
app := core.New(
    httpx.Module(
        httpx.WithBasePath("/api"),
        httpx.WithMiddleware(customAuthMiddleware),
        httpx.WithInfo(buildInfo),
    ),
    // ... rest of your modules
)
```

## Kubernetes Integration

httpx provides health endpoints that work seamlessly with Kubernetes probes:

### Deployment Configuration (Default Paths)

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /livez
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Deployment Configuration (Custom Paths)

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        ports:
        - containerPort: 8080
        env:
        - name: HTTP_HEALTH_READINESS_PATH
          value: "/health/ready"
        - name: HTTP_HEALTH_LIVENESS_PATH
          value: "/health/alive"
        livenessProbe:
          httpGet:
            path: /health/alive  # Custom liveness path
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready  # Custom readiness path
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Migration Jobs

```yaml
apiVersion: batch/v1
kind: Job
spec:
  template:
    spec:
      containers:
      - name: migration
        image: migration:latest
        ports:
        - containerPort: 8080
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        # Job completes when migration finishes and health check passes
      restartPolicy: OnFailure
```

## Dependencies

- `github.com/gostratum/core` - Core framework with fx and health utilities
- `go.uber.org/fx` - Dependency injection framework
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/google/uuid` - UUID generation for request IDs

## License

This project follows the same license as the gostratum organization.