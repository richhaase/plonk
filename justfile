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
        echo "Build failed"
        exit 1
    fi
    echo "Built versioned plonk binary to bin/ (version: $VERSION)"

# Run all unit tests
test:
    @echo "Running unit tests..."
    go test ./...
    @echo "Unit tests passed!"


# Run tests with coverage
test-coverage:
    @echo "Running unit tests with coverage..."
    @go test -coverprofile=coverage.out ./...
    @go tool cover -html=coverage.out -o coverage.html
    @echo "Unit tests passed! Coverage report: coverage.html"



# Run all tests (unit + integration)
test-all: _build-linux _build-test-image
    @echo "Running all tests (unit + integration)..."
    go test ./...
    go test -v -tags=integration ./tests/integration/...
    @echo "All tests passed!"

# Run all tests with coverage
test-all-coverage: _build-linux _build-test-image
    @echo "Running all tests with coverage..."
    @go test -coverprofile=unit.coverage ./...
    @go test -tags=integration -coverprofile=integration.coverage ./tests/integration/...
    @echo "Merging coverage reports..."
    @go run github.com/wadey/gocovmerge unit.coverage integration.coverage > coverage.out
    @go tool cover -html=coverage.out -o coverage.html
    @rm unit.coverage integration.coverage
    @echo "All tests passed! Coverage report: coverage.html"

# Build Linux binary for Docker container (auto-detects architecture)
_build-linux:
    @echo "Building Linux binary for Docker..."
    @if [ "$(uname -m)" = "arm64" ] || [ "$(uname -m)" = "aarch64" ]; then \
        echo "   Detected ARM64 architecture"; \
        CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o plonk-linux cmd/plonk/main.go; \
    else \
        echo "   Detected AMD64 architecture"; \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plonk-linux cmd/plonk/main.go; \
    fi

# Build test container
_build-test-image:
    @echo "Building test container..."
    docker build -t plonk-test:poc -f Dockerfile.integration .



# Clean build artifacts and test cache
clean:
    @echo "Cleaning build artifacts and caches..."
    rm -rf bin dist
    rm -f coverage.out coverage.html
    go clean
    go clean -testcache
    @echo "Build artifacts and test cache cleaned"


# Setup development environment for new contributors
dev-setup:
    @echo "Setting up development environment..."
    @echo "  • Downloading Go dependencies..."
    go mod download
    @echo "  • Installing pre-commit hooks..."
    @if command -v pre-commit &> /dev/null; then \
        pre-commit install; \
    else \
        echo "pre-commit not found. Install with: brew install pre-commit"; \
        exit 1; \
    fi
    @echo "  • Running tests to verify setup..."
    just test
    @echo "Development environment ready!"
    @echo ""
    @echo "Next steps:"
    @echo "  • Run 'just' to see available commands"
    @echo "  • Run 'just build' to build the binary"




# Format Go code and organize imports
format:
    @echo "Formatting Go code and organizing imports..."
    go run golang.org/x/tools/cmd/goimports -w .

# Run linter
lint:
    @echo "Running linter..."
    go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m
