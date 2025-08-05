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

# Run BATS behavioral tests
test-bats:
    @echo "Running BATS behavioral tests..."
    @if ! command -v bats &> /dev/null; then \
        echo "❌ BATS not found. Install with: brew install bats-core"; \
        exit 1; \
    fi
    @cd tests/bats && PLONK_TEST_CLEANUP_PACKAGES=1 PLONK_TEST_CLEANUP_DOTFILES=1 bats behavioral/
    @echo "✅ BATS tests completed!"

# Run tests with coverage
test-coverage:
    @echo "Running unit tests with coverage..."
    @go test -coverprofile=coverage.out ./...
    @go tool cover -html=coverage.out -o coverage.html
    @echo "Unit tests passed! Coverage report: coverage.html"

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
