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

# Generate mocks for testing
generate-mocks:
    @echo "Generating mocks..."
    @go run go.uber.org/mock/mockgen@latest -source=internal/managers/common.go -destination=internal/managers/mock_manager.go -package=managers
    @go run go.uber.org/mock/mockgen@latest -source=internal/state/reconciler.go -destination=internal/state/mock_provider.go -package=state
    @go run go.uber.org/mock/mockgen@latest -source=internal/state/package_provider.go -destination=internal/state/mock_package_interfaces.go -package=state
    @go run go.uber.org/mock/mockgen@latest -source=internal/config/interfaces.go -destination=internal/config/mock_config.go -package=config
    @echo "✅ Generated mocks successfully!"

# Build the plonk binary with version information
build:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building plonk with version information..."
    mkdir -p build
    
    # Get version info using helper
    eval $(just _get-version-info)
    
    if ! go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o build/plonk ./cmd/plonk; then
        echo "❌ Build failed"
        exit 1
    fi
    echo "✅ Built versioned plonk binary to build/ (version: $VERSION)"

# Run all unit tests  
test:
    @echo "Running unit tests..."
    go test ./...
    @echo "✅ Unit tests passed!"

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

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf build dist
    rm -f coverage.out coverage.html
    go clean
    @echo "✅ Build artifacts cleaned"

# Install plonk globally
install:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing plonk globally with version information..."
    
    # Get version info using helper
    eval $(just _get-version-info)
    
    if ! go install -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" ./cmd/plonk; then
        echo "❌ Installation failed"
        exit 1
    fi
    echo "✅ Plonk installed globally! (version: $VERSION)"
    echo "Run 'plonk --help' to get started"

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
