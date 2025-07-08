# Plonk development tasks

# Show available recipes
default:
    @just --list

# Build the plonk binary
build:
    @echo "Building plonk..."
    @mkdir -p build
    go build -o build/plonk ./cmd/plonk
    @echo "✅ Built plonk binary to build/"

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
    @echo "🔍 Running govulncheck..."
    go run golang.org/x/vuln/cmd/govulncheck ./...
    @echo "🔐 Running gosec..."
    go run github.com/securego/gosec/v2/cmd/gosec ./...