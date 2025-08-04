# Plonk development tasks

# Show available recipes
default:
    @just --list

# =============================================================================
# INTERNAL HELPER FUNCTIONS (prefixed with _)
# =============================================================================

# Get version information for builds
_get-version-info:
    #!/usr/bin/env bash
    set -euo pipefail
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "export VERSION='$VERSION'"
    echo "export COMMIT='$COMMIT'"
    echo "export DATE='$DATE'"

# =============================================================================
# MAIN RECIPES
# =============================================================================

# Build the plonk binary with version information
build:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building plonk with version information..."
    mkdir -p bin

    # Get version info using helper
    eval $(just _get-version-info)

    if ! go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o bin/plonk ./cmd/plonk; then
        echo "❌ Build failed"
        exit 1
    fi
    echo "✅ Built versioned plonk binary to bin/ (version: $VERSION)"

# Run all unit tests
test:
    @echo "Running unit tests..."
    go test ./...
    @echo "✅ Unit tests passed!"

test-clear-cache:
    @echo "Clearing test cache..."
    @go clean -testcache
    @echo "✅ Cache cleared!"

# Run tests with coverage
test-coverage:
    @echo "Running unit tests with coverage..."
    @go test -coverprofile=coverage.out ./...
    @go tool cover -html=coverage.out -o coverage.html
    @echo "✅ Unit tests passed! Coverage report: coverage.html"

# Run tests with coverage for CI
test-coverage-ci:
    @echo "Running unit tests with coverage for CI..."
    @go test -race -coverprofile=coverage.out ./...
    @echo "✅ Unit tests passed with coverage!"


# Run integration tests (CI only - requires real package managers)
test-integration:
    @echo "Running integration tests..."
    @if [ -z "$CI" ]; then \
        echo "❌ Integration tests should only run in CI to protect your system"; \
        echo "   Set CI=true to override (at your own risk)"; \
        exit 1; \
    fi
    go test -v -tags=integration ./tests/integration/...
    @echo "✅ Integration tests completed!"

# Build Linux binary for Docker container (auto-detects architecture)
build-linux:
    @echo "🔨 Building Linux binary for Docker..."
    @if [ "$(uname -m)" = "arm64" ] || [ "$(uname -m)" = "aarch64" ]; then \
        echo "   Detected ARM64 architecture"; \
        CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o plonk-linux cmd/plonk/main.go; \
    else \
        echo "   Detected AMD64 architecture"; \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plonk-linux cmd/plonk/main.go; \
    fi

# Build test container
build-test-image:
    @echo "🐳 Building test container..."
    docker build -t plonk-test:poc -f Dockerfile.integration .

# Run containerized integration tests (safe on dev machines)
test-integration-container: build-linux build-test-image
    @echo "🧪 Running integration tests in Docker..."
    @echo "   Using Docker for safety (required on dev machines)"
    PLONK_INTEGRATION=1 go test -v -tags=integration ./tests/integration/container_test.go

# Quick verification of Linux binary
verify-linux-binary: build-linux
    @echo "✓ Testing Linux binary in Docker..."
    @docker run --rm \
        -v $$PWD/plonk-linux:/plonk \
        ubuntu:22.04 \
        /plonk --version || \
        (echo "❌ Linux binary failed" && exit 1)

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin dist
    rm -f coverage.out coverage.html
    go clean
    @echo "✅ Build artifacts cleaned"

# Complete development environment cleanup
clean-all: clean
    @echo "🧹 Performing complete cleanup..."
    @echo "  • Clearing Go module cache..."
    go clean -modcache
    @echo "  • Clearing pre-commit cache..."
    pre-commit clean || true
    rm -rf ~/.cache/pre-commit
    @echo "  • Clearing test cache..."
    go clean -testcache
    @echo "✅ Complete cleanup done!"

# Setup development environment for new contributors
dev-setup:
    @echo "🚀 Setting up development environment..."
    @echo "  • Downloading Go dependencies..."
    go mod download
    @echo "  • Installing pre-commit hooks..."
    @if command -v pre-commit &> /dev/null; then \
        pre-commit install; \
    else \
        echo "⚠️  pre-commit not found. Install with: brew install pre-commit"; \
        exit 1; \
    fi
    @echo "  • Running tests to verify setup..."
    just test
    @echo "✅ Development environment ready!"
    @echo ""
    @echo "Next steps:"
    @echo "  • Run 'just' to see available commands"
    @echo "  • Run 'just build' to build the binary"
    @echo "  • Run 'just precommit' before committing changes"

# Update all dependencies with safety checks
deps-update:
    @echo "🔄 Updating project dependencies..."
    @echo "  • Updating Go dependencies..."
    go get -u ./...
    go mod tidy
    @echo "  • Updating pre-commit hooks..."
    @if command -v pre-commit &> /dev/null; then \
        pre-commit autoupdate; \
    else \
        echo "⚠️  pre-commit not found, skipping hook updates"; \
    fi
    @echo "  • Running validation..."
    @echo "    - Testing..."
    @if ! just test; then \
        echo "❌ Tests failed after update. Review changes carefully."; \
        exit 1; \
    fi
    @echo "    - Linting..."
    @if ! just lint; then \
        echo "❌ Linting failed after update. Review changes carefully."; \
        exit 1; \
    fi
    @echo "✅ Dependencies updated successfully!"
    @echo ""
    @echo "📊 Review changes with:"
    @echo "  git diff go.mod go.sum .pre-commit-config.yaml"

# Run pre-commit checks (format, lint, test, security)
precommit:
    @echo "Running pre-commit checks..."
    @just format
    @just lint
    @just test
    @just security
    @echo "✅ Pre-commit checks passed!"


# Format Go code and organize imports
format:
    @echo "🔧 Formatting Go code and organizing imports..."
    go run golang.org/x/tools/cmd/goimports -w .

# Run linter
lint:
    @echo "🔍 Running linter..."
    go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

# Run security checks (non-blocking)
security:
    @echo "🔐 Running security checks..."
    @go run golang.org/x/vuln/cmd/govulncheck ./...
    @if go run github.com/securego/gosec/v2/cmd/gosec ./...; then \
        echo "✅ No security issues found!"; \
    else \
        echo "⚠️  Security warnings found (non-blocking)"; \
    fi
    @echo "✅ Security checks completed!"
