# Plonk development tasks

# Show available recipes
default:
    @just --list

# Generate mocks for testing
generate-mocks:
    @echo "Generating mocks..."
    @mockgen -source=internal/managers/common.go -destination=internal/managers/mock_manager.go -package=managers
    @mockgen -source=internal/state/reconciler.go -destination=internal/state/mock_provider.go -package=state
    @mockgen -source=internal/state/package_provider.go -destination=internal/state/mock_package_interfaces.go -package=state
    @mockgen -source=internal/config/interfaces.go -destination=internal/config/mock_config.go -package=config
    @echo "✅ Generated mocks successfully!"

# Build the plonk binary
build:
    @echo "Building plonk..."
    @mkdir -p build
    go build -o build/plonk ./cmd/plonk
    @echo "Built plonk binary to build/"

# Build the plonk binary with version information
build-versioned:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building plonk with version information..."
    mkdir -p build
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o build/plonk ./cmd/plonk
    echo "Built versioned plonk binary to build/"

# Run all unit tests  
test:
    @echo "Running unit tests..."
    go test ./...
    @echo "✅ Unit tests passed!"

# Run integration tests (requires Docker)
test-integration:
    @echo "Running integration tests..."
    @echo "🐳 Building Docker image..."
    @docker build -t plonk-test -f test/integration/docker/Dockerfile .
    @echo "🧪 Running integration tests..."
    go test -tags=integration -v ./test/integration/... -timeout=10m
    @echo "✅ Integration tests passed!"

# Run integration tests with faster timeout for development
test-integration-fast:
    @echo "Running fast integration tests..."
    @echo "🐳 Building Docker image..."
    @docker build -t plonk-test -f test/integration/docker/Dockerfile .
    @echo "🧪 Running fast integration tests..."
    go test -tags=integration -v ./test/integration/... -timeout=5m -short
    @echo "✅ Fast integration tests passed!"

# Build Docker image for integration tests
test-integration-setup:
    @echo "Building Docker image for integration tests..."
    docker build -t plonk-test -f test/integration/docker/Dockerfile .
    @echo "✅ Docker image built successfully!"

# Run all tests (unit + integration)
test-all:
    @echo "Running all tests..."
    @just test
    @just test-integration
    @echo "✅ All tests passed!"

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf build
    go clean
    @echo "✅ Build artifacts cleaned"

# Clean Docker images and artifacts
clean-docker:
    @echo "Cleaning Docker images and artifacts..."
    -docker rmi plonk-test
    -docker system prune -f
    @echo "✅ Docker artifacts cleaned"

# Install plonk globally
install:
    @echo "Installing plonk globally..."
    go install ./cmd/plonk
    @echo "✅ Plonk installed globally!"
    @echo "Run 'plonk --help' to get started"

# Run pre-commit checks (format, lint, test, security)
precommit:
    @echo "Running pre-commit checks..."
    @just format
    @just lint
    @just test
    @just security
    @echo "✅ Pre-commit checks passed!"

# Run pre-commit checks with integration tests
precommit-full:
    @echo "Running full pre-commit checks..."
    @just format
    @just lint
    @just test-all
    @just security
    @echo "✅ Full pre-commit checks passed!"

# Format Go code and organize imports
format:
    @echo "🔧 Formatting Go code and organizing imports..."
    go run golang.org/x/tools/cmd/goimports -w .

# Run linter
lint:
    @echo "🔍 Running linter..."
    go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

# Run security checks
security:
    @echo "Running govulncheck..."
    go run golang.org/x/vuln/cmd/govulncheck ./...
    @echo "Running gosec..."
    go run github.com/securego/gosec/v2/cmd/gosec ./...

# Interactive release command (simplified for testing)
release:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Plonk Release Manager"
    echo "======================="
    CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "Current version: $CURRENT_VERSION"
    echo "Use this to create a manual tag:"
    echo "git tag -a v1.0.0 -m 'Release v1.0.0'"
