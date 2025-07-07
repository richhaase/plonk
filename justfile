# Development tasks for Plonk
# Run `just` to see available commands

# Default recipe lists all available commands
default:
    @just --list

# Install all development dependencies
setup:
    @echo "Installing development tools..."
    asdf install
    @echo "✅ Development tools installed"

# Build the plonk binary
build:
    @echo "Building plonk..."
    @mkdir -p bin
    go build -o bin/plonk ./cmd/plonk
    @echo "✅ Built plonk binary to bin/"

# Run all tests
test:
    @echo "Running tests..."
    go test ./...

# Run tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    go test -cover ./...

# Run linter
lint:
    @echo "Running linter..."
    golangci-lint run --timeout=10m

# Fix linter issues automatically
lint-fix:
    @echo "Running linter with automatic fixes..."
    golangci-lint run --fix --timeout=10m

# Format code (gofmt)
format:
    @echo "Formatting code..."
    go fmt ./...

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin
    go clean

# Run full CI pipeline (format, lint, test, build)
ci: format lint test build
    @echo "✅ Full CI pipeline completed successfully"

# Install the binary to GOBIN (or GOPATH/bin if GOBIN unset)
install: build
    @echo "Installing plonk to $(go env GOBIN)..."
    go install ./cmd/plonk
    @echo "✅ Plonk installed"

# Show project status
status:
    @echo "=== Plonk Development Status ==="
    @echo "Go version: $(go version)"
    @echo "Linter version: $(golangci-lint version --format short)"
    @echo "Just version: $(just --version)"
    @echo ""
    @echo "Tools status:"
    @asdf current
    @echo ""
    @echo "Git status:"
    @git status --porcelain

# Development workflow: format, lint, test
dev: format lint test
    @echo "✅ Development checks passed"