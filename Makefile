.PHONY: help test test-coverage lint clean install-tools
.PHONY: version validate-version update-deps bump-patch bump-minor bump-major
.PHONY: release release-dry-run release-patch release-minor release-major

# Get current version
VERSION := $(shell cat .version 2>/dev/null || echo "0.0.0")

# Default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Testing & Quality (httpx module):"
	@echo "  test              - Run httpx unit tests"
	@echo "  test-coverage     - Run httpx tests with coverage"
	@echo "  lint              - Run linters for httpx module"
	@echo "  clean             - Clean httpx build artifacts"
	@echo ""
	@echo "Version Management (httpx module):"
	@echo "  version           - Show current httpx module version"
	@echo "  validate-version  - Validate .version file for httpx"
	@echo "  update-deps       - Update gostratum dependencies used by httpx"
	@echo "  bump-patch        - Bump httpx patch version (0.0.X)"
	@echo "  bump-minor        - Bump httpx minor version (0.X.0)"
	@echo "  bump-major        - Bump httpx major version (X.0.0)"
	@echo ""
	@echo "Release Management (httpx module):"
	@echo "  release           - Create new httpx release (default: patch)"
	@echo "  release-patch     - Create httpx patch release"
	@echo "  release-minor     - Create httpx minor release"
	@echo "  release-major     - Create httpx major release"
	@echo "  release-dry-run   - Test httpx release without committing"
	@echo ""
	@echo "Current version: v$(VERSION)"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@GOLANGCI_BIN=$(go env GOPATH)/bin/golangci-lint; \
	if [ -x "$GOLANGCI_BIN" ]; then \
		"$GOLANGCI_BIN" run ./...; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: make install-tools"; exit 1; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html
	go clean -cache

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "Installing golangci-lint..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Tools installed successfully"


# Version management
version:
	@echo "Current version: v$(VERSION)"

validate-version:
	@./scripts/validate-version.sh

update-deps:
	@./scripts/update-deps.sh

bump-patch:
	@./scripts/bump-version.sh patch

bump-minor:
	@./scripts/bump-version.sh minor

bump-major:
	@./scripts/bump-version.sh major

# Release management
release:
	@./scripts/release.sh $(or $(TYPE),patch)

release-dry-run:
	@DRY_RUN=true ./scripts/release.sh $(or $(TYPE),patch)

release-patch:
	@./scripts/release.sh patch

release-minor:
	@./scripts/release.sh minor

release-major:
	@./scripts/release.sh major


