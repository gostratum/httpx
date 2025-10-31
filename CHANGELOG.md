# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Removed
- **BREAKING:** Removed `SanitizeViper()` function - was added in v0.1.4 but never used externally
  - Modern replacement: Use `Config.Sanitize()` method instead
  - Eliminates last viper dependency from httpx module
  - Part of framework-wide viper elimination effort

## [0.2.0] - 2025-10-29

### Breaking Changes

- Removed `WithBasePath(path string)` option. Base path is now a configuration value (YAML / env), not a programmatic option.

### Migration Guide

- Principle: "Configuration for values, Options for code" — move simple scalar values to YAML/configx and keep programmatic options for Go functions/objects.
- Example: move `WithBasePath("/api/v1")` to YAML as `http.base_path` and keep `WithMiddleware(...)` as an option.

### Added

- Observability support and lifecycle management for the HTTP server (start/stop hooks, metrics/tracing integration). (commit: 5c74853)

### Changed

- Update configuration management to use YAML and `core/configx` (remove programmatic base path option). (commit: 5a11a47)
- Update gostratum dependencies to v0.2.0 for core, metricsx, and tracingx. (commit: 64563d5)

### Chore

- Release bump to v0.2.0. (commit: 4f20e4a)


## [0.1.6] - 2025-10-26

### Added

- Small maintenance and release tooling improvements (Makefile and scripts for version management, dependency updates, and changelog maintenance). (commit: 8fbe840)

### Changed

- Refactor: HTTP module now uses typed configuration with `configx`; updated health routes and middleware accordingly. (commit: 0bf2ddb)
- Minor dependency bumps and test additions prior to release. (commits: 22cf7f1, others)


## [0.1.5] - 2025-10-25

### Added

- Add Makefile and scripts for version management, dependency updates, and changelog maintenance. (commit: 8fbe840)

### Changed

- Refactor HTTP module to use typed configuration with `configx`; update health routes and middleware accordingly. (commit: 0bf2ddb)

## [0.1.4] - 2025-10-21

### Changed

- Update gostratum/core, metricsx, and tracingx to v0.1.8, v0.1.4. (commit: 7dacb60)

### Added

- Implement `SanitizeViper` function to redact sensitive information from Viper configuration. (commit: 7aa493b)

## [0.1.3] - 2025-10-20

### Changed

- Update gostratum/metricsx and gostratum/tracingx to v0.1.3 in go.mod and go.sum. (commit: 0636001)
- Refactor middleware to use `any` types for improved type safety and update go.mod/go.sum accordingly. (commit: c72e920)

## [0.1.2] - 2025-10-17

### Changed

- Rename `RegisterHealthRoutes` to `registerHealthRoutes` for internal consistency. (commit: a0b5ea0)
- Update dependencies for gostratum/core, metricsx, and tracingx to latest versions. (commit: 62bf593)

## [0.1.1] - 2025-10-16

### Added

- Add `.gitignore` and update `go.mod`/`go.sum` with new dependencies; implement test for `MetaMiddleware` to ensure request ID is set correctly. (commit: 646895e)

### Changed

- Refactor logging to use `logx.Logger` across the application and update dependencies in `go.mod`/`go.sum`. (commit: 413187c)

### Added

- Add observability middleware for metrics and tracing. (commit: 6a0b2c4)

## [0.1.0] - 2025-10-09

### Added

- Implement `responsex` package with `MetaMiddleware`, pagination, and envelope structure for API responses. (commit: 01834db)
- Enhance `RecoveryMiddleware` to prioritize request ID from context. (commit: 044a58a)
- Add `httpx` module with health endpoints, middleware, and configuration options. (commit: 64715c2)