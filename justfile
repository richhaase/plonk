# Plonk development tasks

# Show available recipes
default:
    @just --list

# Build the plonk binary with version information
build:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building plonk with version information..."
    mkdir -p bin

    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

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
        echo "BATS not found. Install with: brew install bats-core"; \
        exit 1; \
    fi
    @cd tests/bats && PLONK_TEST_CLEANUP_PACKAGES=1 PLONK_TEST_CLEANUP_DOTFILES=1 bats behavioral/
    @echo "BATS tests completed!"

# Run tests with coverage
test-coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running unit tests with normalized coverage..."

    # Normalize the denominator by excluding low-signal packages from coverage totals.
    # Override with COVER_EXCLUDE_REGEX to customize (must match go list package import paths).
    # Default excludes:
    #   - internal/testutil: helper/test-only utilities
    #   - tools: tool stubs not part of runtime
    #   - cmd/*: CLI entry binaries (thin wrappers over libraries)
    COVER_EXCLUDE_REGEX=${COVER_EXCLUDE_REGEX:-'/(internal/testutil|tools|cmd/.*)$'}

    # Resolve the list of packages to test and instrument
    mapfile -t PKGS < <(go list ./... | grep -Ev "${COVER_EXCLUDE_REGEX}")
    if [ ${#PKGS[@]} -eq 0 ]; then
        echo "No packages selected for coverage after filtering. Check COVER_EXCLUDE_REGEX." >&2
        exit 2
    fi

    # Build comma-separated coverpkg list from the filtered packages
    COVERPKG=$(printf '%s,' "${PKGS[@]}" | sed 's/,$//')

    # Run tests with cross-package instrumentation over the normalized package set
    go test -coverpkg="${COVERPKG}" -covermode=atomic -coverprofile=coverage.out "${PKGS[@]}"

    # Generate textual summary and total coverage
    go tool cover -func=coverage.out > coverage.txt
    awk 'END{printf "Total coverage (normalized): %s\n", $3}' coverage.txt

    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    echo "Unit tests passed! Coverage report: coverage.html (see also coverage.txt)"

# Clean build artifacts and test cache
clean:
    @echo "Cleaning build artifacts and caches..."
    rm -rf bin dist
    rm -f coverage.out coverage.html coverage.txt
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
    @echo "Development environment ready!"
    @echo ""
    @echo "Next steps:"
    @echo "  • Run 'just' to see available commands"
    @echo "  • Run 'just build' to build the binary"
    @echo "  • Run 'just test' to run tests"

# Find dead code (unreachable functions)
find-dead-code:
    @echo "Finding dead code..."
    @if ! command -v deadcode &> /dev/null; then \
        echo "Installing deadcode tool..."; \
        go install golang.org/x/tools/cmd/deadcode@latest; \
    fi
    deadcode -test ./...
