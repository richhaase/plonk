# Plonk development tasks

# Show available recipes
default:
    @just --list
    @echo
    @echo "Quick release workflow:"
    @echo "  just release-version-suggest  # Get version suggestions"
    @echo "  just release-check           # Validate GoReleaser config"
    @echo "  just release-snapshot        # Test release (no publishing)"
    @echo "  just release v1.2.3          # Create release with GoReleaser"

# =============================================================================
# INTERNAL HELPER FUNCTIONS (prefixed with _)
# =============================================================================

# Check if GoReleaser is installed
_check-goreleaser:
    #!/usr/bin/env bash
    if ! command -v goreleaser &> /dev/null; then
        echo "❌ GoReleaser not found. Install with:"
        echo "   brew install goreleaser"
        echo "   or visit: https://goreleaser.com/install/"
        exit 1
    fi

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

# Validate git working directory is clean
_require-clean-git:
    #!/usr/bin/env bash
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo "❌ Working directory not clean. Please commit or stash changes."
        exit 1
    fi

# Check if we're on main/master branch with optional override
_check-main-branch:
    #!/usr/bin/env bash
    CURRENT_BRANCH=$(git branch --show-current)
    if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
        echo "⚠️  Warning: Not on main/master branch (currently on: $CURRENT_BRANCH)"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Operation cancelled"
            exit 1
        fi
    fi

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
    @go test -race -coverprofile=coverage.out -covermode=atomic ./...
    @echo "✅ Unit tests passed with coverage!"

# Complete UX validation - ensures all commands work as expected
test-ux: test-clear-cache
    @echo "Running complete UX integration tests..."
    @go test -tags=integration ./tests/integration -run TestCompleteUserExperience -v -timeout 10m
    @echo "✅ UX integration tests passed!"

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
    @echo "  • Generating test mocks..."
    just generate-mocks
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

# Release using GoReleaser (requires GoReleaser to be installed)
release VERSION:
    #!/usr/bin/env bash
    set -euo pipefail

    VERSION="{{VERSION}}"

    # Validate version format
    if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
        echo "❌ Invalid version format. Use vX.Y.Z or vX.Y.Z-suffix (e.g., v1.2.3, v1.2.3-rc1)"
        exit 1
    fi

    # Check if tag already exists
    if git tag -l | grep -q "^$VERSION$"; then
        echo "❌ Tag $VERSION already exists!"
        exit 1
    fi

    # Use helper functions for validation
    just _check-goreleaser
    just _check-main-branch
    just _require-clean-git

    echo "🚀 Starting GoReleaser release process for $VERSION"
    echo "================================================="

    # Create and push tag (GoReleaser will build from this tag)
    echo "🏷️  Creating release tag..."
    git tag -a "$VERSION" -m "Release $VERSION"

    echo "📤 Pushing tag to remote..."
    git push origin "$VERSION"

    # Run GoReleaser
    echo "🚀 Running GoReleaser..."
    goreleaser release --clean

    echo
    echo "✅ Release $VERSION completed successfully!"
    echo "🎉 Check GitHub releases: https://github.com/richhaase/plonk/releases"

# Test release process without publishing (dry run)
release-snapshot:
    #!/usr/bin/env bash
    set -euo pipefail

    # Use helper function for validation
    just _check-goreleaser

    echo "🔍 Running GoReleaser in snapshot mode (no publishing)..."
    if ! goreleaser release --snapshot --clean; then
        echo "❌ Snapshot build failed"
        exit 1
    fi

    echo
    echo "✅ Snapshot build completed!"
    echo "📦 Check dist/ directory for generated binaries"

# Validate GoReleaser configuration
release-check:
    #!/usr/bin/env bash
    set -euo pipefail

    # Use helper function for validation
    just _check-goreleaser

    echo "🔍 Validating GoReleaser configuration..."
    if ! goreleaser check; then
        echo "❌ GoReleaser configuration is invalid"
        exit 1
    fi

    echo "✅ GoReleaser configuration is valid!"


# Show suggested next version based on current tags
release-version-suggest:
    #!/usr/bin/env bash
    set -euo pipefail

    # Get current version
    CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "Current version: $CURRENT_VERSION"
    echo

    # Show recent commits since last tag
    echo "Recent commits since $CURRENT_VERSION:"
    git log --oneline --no-merges $CURRENT_VERSION..HEAD 2>/dev/null || echo "  (no commits since last tag)"
    echo

    # Parse current version for increment suggestions
    if [[ $CURRENT_VERSION =~ ^v?([0-9]+)\.([0-9]+)\.([0-9]+)(-rc([0-9]+))?$ ]]; then
        MAJOR=${BASH_REMATCH[1]}
        MINOR=${BASH_REMATCH[2]}
        PATCH=${BASH_REMATCH[3]}
        RC=${BASH_REMATCH[5]:-""}
    else
        MAJOR=0
        MINOR=0
        PATCH=0
        RC=""
    fi

    # Calculate version options
    PATCH_VERSION="v$MAJOR.$MINOR.$((PATCH + 1))"
    MINOR_VERSION="v$MAJOR.$((MINOR + 1)).0"
    MAJOR_VERSION="v$((MAJOR + 1)).0.0"
    if [[ -n $RC ]]; then
        RC_VERSION="v$MAJOR.$MINOR.$PATCH-rc$((RC + 1))"
    else
        RC_VERSION="v$MAJOR.$((MINOR + 1)).0-rc1"
    fi

    echo "Suggested versions:"
    echo "  Patch: $PATCH_VERSION (bug fixes)"
    echo "  Minor: $MINOR_VERSION (new features)"
    echo "  Major: $MAJOR_VERSION (breaking changes)"
    echo "  RC: $RC_VERSION (release candidate)"
    echo
    echo "Usage: just release <version>"
    echo "Example: just release $PATCH_VERSION"
