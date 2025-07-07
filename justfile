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

# Run all tests  
test:
    @echo "Running tests..."
    go test ./...
    @echo "✅ Tests passed!"

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf build
    go clean
    @echo "✅ Build artifacts cleaned"

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