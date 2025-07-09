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

# Build the plonk binary with version information
build:
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

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf build
    go clean
    @echo "✅ Build artifacts cleaned"

# Install plonk globally
install:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing plonk globally with version information..."
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    go install -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" ./cmd/plonk
    echo "✅ Plonk installed globally!"
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

# Run security checks
security:
    @echo "Running govulncheck..."
    go run golang.org/x/vuln/cmd/govulncheck ./...
    @echo "Running gosec..."
    go run github.com/securego/gosec/v2/cmd/gosec ./...

# Interactive release command
release:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Plonk Release Manager"
    echo "======================="
    
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
    
    read -p "Enter version (e.g., v1.2.3): " NEW_VERSION
    if [[ ! $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?$ ]]; then
        echo "Invalid version format. Use vX.Y.Z or vX.Y.Z-rcN"
        exit 1
    fi
    
    echo
    echo "Creating release: $NEW_VERSION"
    echo
    
    # Create annotated tag
    read -p "Enter release notes: " RELEASE_NOTES
    git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION - $RELEASE_NOTES"
    
    echo "Created tag: $NEW_VERSION"
    echo
    echo "Next steps:"
    echo "  1. Push the tag: git push origin $NEW_VERSION"
    echo "  2. Run release build: just goreleaser-release"
    echo "  3. Check GitHub releases"

# Run goreleaser for actual release (requires tag)
goreleaser-release:
    @echo "Running goreleaser for release..."
    goreleaser release
