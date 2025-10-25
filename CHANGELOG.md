# Changelog

## [Unreleased]

## [0.1.5] - 2025-10-25

### Added
- Release version 0.1.5

### Changed
- Updated gostratum dependencies to latest versions


# Changelog — httpx

All notable changes to the `httpx` module are documented in this file.

## Unreleased

### Added
- (8fbe840) Add `Makefile` and release tooling
  - Add top-level `Makefile` with targets for tests, lint, version bumping, release creation and local dev helpers.
  - Add `scripts/` helpers: dependency updater, changelog helper, release automation and version bump scripts.
  - Improves developer ergonomics for releasing and dependency management.

### Changed
- (0bf2ddb) Refactor to typed configuration and update middleware/health routes
  - Replace untyped `viper` usage with typed `httpx.Config` and `configx.Loader` for safer config handling.
  - Update health endpoints and middleware wiring to use the new typed config and clearer lifecycle hooks.

### Dependencies
- (0636001) Bump gostratum dependencies (metricsx, tracingx)
  - Update `go.mod`/`go.sum` to use newer `gostratum/metricsx` and `gostratum/tracingx` versions (v0.1.3 and friends).

---

### Notes
- Each entry includes the short commit hash in parentheses for traceability.
- This changelog was generated from recent commits in the `httpx` module (commits at indices 0, 1, and 4 from `git log`).

If you'd like a different changelog layout (Keep a CHANGELOG per semver tag, or use "Unreleased" then add releases), I can reformat and include the full commit messages and dates.