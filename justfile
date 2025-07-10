# Plonk development tasks

# Show available recipes
default:
    @just --list
    @echo
    @echo "Quick release workflow:"
    @echo "  just release-version-suggest  # Get version suggestions"
    @echo "  just release-auto v1.2.3      # Automated release"

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

# Automated single-command release process
release-auto VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    
    VERSION="{{VERSION}}"
    
    # Validate version format
    if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?$ ]]; then
        echo "❌ Invalid version format. Use vX.Y.Z or vX.Y.Z-rcN (e.g., v1.2.3)"
        exit 1
    fi
    
    # Check if tag already exists
    if git tag -l | grep -q "^$VERSION$"; then
        echo "❌ Tag $VERSION already exists!"
        exit 1
    fi
    
    # Check if we're on main branch (optional safety check)
    CURRENT_BRANCH=$(git branch --show-current)
    if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
        echo "⚠️  Warning: Not on main/master branch (currently on: $CURRENT_BRANCH)"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Release cancelled"
            exit 1
        fi
    fi
    
    # Check working directory is clean
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo "❌ Working directory not clean. Please commit or stash changes."
        exit 1
    fi
    
    echo "🚀 Starting automated release process for $VERSION"
    echo "=================================================="
    
    # Pre-release validation
    echo "📋 Running pre-release validation..."
    
    # Run tests
    echo "  🧪 Running tests..."
    just test
    
    # Run linter
    echo "  🔍 Running linter..."
    just lint
    
    # Run security checks
    echo "  🔐 Running security checks..."
    just security
    
    # Test build
    echo "  🔨 Testing build..."
    just build
    
    echo "✅ Pre-release validation passed!"
    echo
    
    # Get release notes
    CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "📝 Preparing release notes..."
    echo "Recent commits since $CURRENT_VERSION:"
    git log --oneline --no-merges $CURRENT_VERSION..HEAD 2>/dev/null || echo "  (no commits since last tag)"
    echo
    
    read -p "Enter release notes: " RELEASE_NOTES
    if [[ -z "$RELEASE_NOTES" ]]; then
        RELEASE_NOTES="Release $VERSION"
    fi
    
    # Create and push tag
    echo "🏷️  Creating release tag..."
    git tag -a "$VERSION" -m "Release $VERSION - $RELEASE_NOTES"
    
    echo "📤 Pushing tag to remote..."
    git push origin "$VERSION"
    
    # Build release binaries
    echo "🚀 Building release binaries..."
    mkdir -p dist
    
    # Build for common platforms
    echo "  Building for linux/amd64..."
    GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/plonk-linux-amd64 ./cmd/plonk
    
    echo "  Building for darwin/amd64..."
    GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/plonk-darwin-amd64 ./cmd/plonk
    
    echo "  Building for darwin/arm64..."
    GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/plonk-darwin-arm64 ./cmd/plonk
    
    echo "  Building for windows/amd64..."
    GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/plonk-windows-amd64.exe ./cmd/plonk
    
    echo
    echo "✅ Release $VERSION completed successfully!"
    echo "📦 Release binaries built in dist/ directory"
    echo "🏷️  Git tag $VERSION created and pushed"


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
    echo "Usage: just release-auto <version>"
    echo "Example: just release-auto $PATCH_VERSION"
