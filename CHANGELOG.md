# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


## [0.2.0] - 2025-10-28

### Breaking Changes

**BREAKING:** The following module option has been removed in favor of YAML configuration:
1. ❌ `WithBasePath(path string)` - Use `http.base_path` in YAML config

### Migration Guide

**Principle:** "Configuration for values, Options for code"

#### Before (Phase 1)
```go
httpx.Module(
    httpx.WithBasePath("/api/v1"),
    httpx.WithMiddleware(authMW),
)
```

#### After (Phase 2)
```go
httpx.Module(
    httpx.WithMiddleware(authMW),  // Only programmatic options remain
)
```

**YAML config (base.yaml or dev.yaml):**
```yaml
http:
  addr: ":8080"
  base_path: "/api/v1"  # Moved to YAML
  
  health:
    readiness_path: "/healthz"
    liveness_path: "/livez"
```

#### Options Kept (Still Needed)

These options remain because they require Go functions/objects:
- ✅ `WithMiddleware(mw ...gin.HandlerFunc)` - Go functions cannot be in YAML
- ✅ `WithInfo(BuildInfo)` - Build metadata for programmatic injection

### Changed

- Refactored module to follow "Configuration for values, Options for code" principle
- Base path now managed via YAML with `core/configx`
- Updated examples and documentation with new pattern

### Benefits

1. **Simpler API:** 3 module options → 2 options (33% reduction)
2. **Configuration Consistency:** BasePath value in YAML, middleware in options
3. **Environment Flexibility:** BasePath can be overridden via `STRATUM_HTTP_BASE_PATH` env var
4. **Clearer Intent:** Only programmatic options remain (Go functions/objects)
5. **Better Discoverability:** Base path documented in YAML schema
6. **Alignment with Framework:** Consistent with DBX and other modules

### Added
- Release version 0.1.6

### Changed
- Updated gostratum dependencies to latest versions


## [0.1.5] - 2025-10-26

### Added

- Add Makefile and scripts for version management, dependency updates, and changelog maintenance

### Changed

- Refactor HTTP module to use typed configuration with configx; update health routes and middleware accordingly

## [0.1.4] - 2025-10-26

### Changed

- Update gostratum/core, metricsx, and tracingx to v0.1.8, v0.1.4

### Added

- Implement SanitizeViper function to redact sensitive information from Viper configuration

## [0.1.3] - 2025-10-26

### Changed

- Update gostratum/metricsx and gostratum/tracingx to v0.1.3 in go.mod and go.sum
- Update go.mod and go.sum to use latest versions of gostratum/core, metricsx, and tracingx; refactor middleware to use 'any' type for improved type safety

## [0.1.2] - 2025-10-26

### Changed

- Rename RegisterHealthRoutes to registerHealthRoutes for consistency

### Changed

- Update dependencies for gostratum/core, metricsx, and tracingx to latest versions

## [0.1.1] - 2025-10-26

### Added

- Add .gitignore and update go.mod/go.sum with new dependencies; implement test for MetaMiddleware to ensure request ID is set correctly

### Changed

- Refactor logging to use logx.Logger across the application and update dependencies in go.mod/go.sum

### Added

- Add observability middleware for metrics and tracing

## [0.1.0] - 2025-10-26

### Added

- Implement responsex package with MetaMiddleware, pagination, and envelope structure for API responses
- Enhance RecoveryMiddleware to prioritize request ID from context
- Add httpx module with health endpoints, middleware, and configuration options