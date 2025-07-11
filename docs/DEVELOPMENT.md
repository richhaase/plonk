# Development Guide

Development setup, build processes, and contributing guidelines for plonk.

## Development Setup

### Requirements

- **Go 1.24.4+** (see `.tool-versions`)
- **Just** (command runner)
- **Git** (version control)
- **Development tools** (automatically managed via `go.mod`)

### Quick Start

```bash
# Clone repository
git clone https://github.com/your-username/plonk
cd plonk

# One-command development setup (recommended)
just dev-setup

# Or manual setup:
go mod download
pre-commit install
just generate-mocks
just test

# Build and install
just build
just install
```

The `dev-setup` command handles all development environment setup automatically, including dependencies, git hooks, mocks, and verification tests.

## Build System

### Just Commands

All development tasks use the `just` command runner:

```bash
# Show all available commands
just

# Common development tasks
just build              # Build binary with version info
just test               # Run unit tests
just install            # Install binary globally
just clean              # Clean build artifacts

# Developer automation
just dev-setup          # Complete development environment setup
just deps-update        # Update dependencies with safety checks
just clean-all          # Deep clean including caches

# Code quality
just format             # Format code and organize imports
just lint               # Run linter
just security           # Run security checks
just precommit          # Run all pre-commit checks

# Mock generation
just generate-mocks     # Generate test mocks

# Release (automated)
just release-auto v1.2.3 # Complete automated release
```

### Build Process

The build process injects version information:

```bash
# Build with version info
just build

# Manual build
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o build/plonk ./cmd/plonk
```

## Code Organization

### Project Structure

```
cmd/plonk/              # CLI entry point
internal/
├── commands/           # CLI command implementations
├── config/             # Configuration management
├── dotfiles/           # Dotfile operations
├── errors/             # Structured error handling
├── managers/           # Package manager implementations
└── state/              # State reconciliation engine
docs/                   # Documentation
```

## Git Hooks & Pre-commit

### Pre-commit Framework

Plonk uses the industry-standard pre-commit framework for better developer experience:

```bash
# Automatic setup (included in dev-setup)
just dev-setup

# Or manual installation
brew install pre-commit  # macOS
pip install pre-commit   # Python
pre-commit install

# Common pre-commit commands
pre-commit run --all-files     # Run all hooks manually
pre-commit autoupdate          # Update hook versions
SKIP=go-test git commit        # Skip specific hooks
```

**Benefits:**
- ⚡ **94% faster** on non-Go file changes
- 🎯 **File-specific filtering** (Go hooks only on .go files)
- 🔄 **Automatic updates** via `just deps-update`
- 🛠 **Rich ecosystem** of community hooks
- 📊 **Better error reporting** with file context

### Manual Checks

You can still run checks manually using just:

```bash
# Run all pre-commit checks manually
just precommit
```

### Available Checks

The pre-commit framework includes:
- **Format**: `goimports` code formatting
- **Lint**: `golangci-lint` static analysis
- **Test**: Go unit tests
- **Security**: `govulncheck` vulnerability scanning

Plus additional checks in pre-commit framework:
- YAML/TOML syntax validation
- Trailing whitespace removal
- End-of-file fixing
- Large file detection
- Merge conflict detection

## Testing

### Test Structure

```bash
# Unit tests alongside source code
internal/config/yaml_config_test.go
internal/managers/homebrew_test.go
internal/state/reconciler_test.go
```

### Running Tests

```bash
# Run all tests
just test

# Run specific package tests
go test ./internal/config/

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestSpecificFunction ./internal/config/
```

### Test Patterns

#### Table-Driven Tests
```go
func TestPackageManager(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid package", "git", "git", false},
        {"invalid package", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

#### Mock Usage
```go
func TestWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockManager := managers.NewMockPackageManager(ctrl)
    mockManager.EXPECT().
        IsAvailable(gomock.Any()).
        Return(true, nil)

    // test with mock
}
```

#### Context Testing
```go
func TestContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    _, err := manager.ListInstalled(ctx)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context canceled")
}
```

### Mock Generation

```bash
# Generate all mocks
just generate-mocks

# Manual mock generation
mockgen -source=internal/managers/common.go -destination=internal/managers/mock_manager.go -package=managers
```

## Code Quality

See [Pre-commit Migration Guide](../.github/PRE_COMMIT_MIGRATION.md) for complete setup information.

### Pre-commit Checks

```bash
# Run all pre-commit checks manually
just precommit

# Individual checks
just format     # goimports
just lint       # golangci-lint
just security   # gosec + govulncheck
just test       # unit tests
```

### Code Style

- **goimports** for formatting and import organization
- **golangci-lint** for comprehensive linting
- **gosec** for security vulnerability scanning
- **govulncheck** for dependency vulnerability scanning

### Error Handling

Plonk uses a structured error system for consistent error handling across all commands.

#### Error Creation Patterns

**Use structured errors for all user-facing operations:**

```go
// Import the errors package
import "plonk/internal/errors"

// Create a new error
return errors.NewError(
    errors.ErrPackageInstall,    // Error code
    errors.DomainPackages,       // Domain
    "install",                   // Operation
    "failed to install package" // Message
)

// Wrap an existing error
return errors.Wrap(
    err,                         // Original error
    errors.ErrConfigNotFound,    // Error code
    errors.DomainConfig,         // Domain
    "load",                      // Operation
    "failed to load configuration" // Message
)

// Wrap with item context
return errors.WrapWithItem(
    err,                         // Original error
    errors.ErrPackageInstall,    // Error code
    errors.DomainPackages,       // Domain
    "install",                   // Operation
    packageName,                 // Item (package name)
    "failed to install package" // Message
)
```

#### Error Codes and Domains

**Always use appropriate error codes:**

```go
// Configuration errors
errors.ErrConfigNotFound      // Config file missing
errors.ErrConfigParseFailure  // Config syntax error
errors.ErrConfigValidation    // Config validation failed

// Package management errors
errors.ErrPackageInstall      // Package installation failed
errors.ErrManagerUnavailable  // Package manager not available

// File operation errors
errors.ErrFileIO             // General file I/O error
errors.ErrFilePermission     // Permission denied
errors.ErrFileNotFound       // File not found

// User input errors
errors.ErrInvalidInput       // Invalid command arguments

// System errors
errors.ErrInternal           // Internal system error
errors.ErrReconciliation     // State reconciliation failed
```

**Always use appropriate domains:**

```go
errors.DomainConfig          // Configuration-related operations
errors.DomainPackages        // Package management operations
errors.DomainDotfiles        // Dotfile operations
errors.DomainCommands        // Command-level operations
errors.DomainState           // State reconciliation
```

#### Command Error Handling

**Standard pattern for command error handling:**

```go
func runCommand(cmd *cobra.Command, args []string) error {
    // Parse output format
    format, err := ParseOutputFormat(outputFormat)
    if err != nil {
        return errors.WrapWithItem(err, errors.ErrInvalidInput,
            errors.DomainCommands, "command-name", "output-format",
            "invalid output format")
    }

    // Load configuration
    cfg, err := config.LoadConfig(configDir)
    if err != nil {
        return errors.Wrap(err, errors.ErrConfigNotFound,
            errors.DomainConfig, "load", "failed to load configuration")
    }

    // Perform operation
    if err := performOperation(cfg); err != nil {
        return errors.WrapWithItem(err, errors.ErrPackageInstall,
            errors.DomainPackages, "install", packageName,
            "failed to install package")
    }

    return nil
}
```

#### Error Testing

**Test error conditions explicitly:**

```go
func TestCommandErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        setupError  error
        expectedErr string
        expectedCode errors.ErrorCode
    }{
        {
            name:        "config not found",
            setupError:  os.ErrNotExist,
            expectedErr: "failed to load configuration",
            expectedCode: errors.ErrConfigNotFound,
        },
        {
            name:        "invalid package manager",
            setupError:  fmt.Errorf("invalid manager"),
            expectedErr: "manager not available",
            expectedCode: errors.ErrManagerUnavailable,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup error condition
            err := runCommand(cmd, args)

            // Verify structured error
            var plonkErr *errors.PlonkError
            assert.True(t, errors.As(err, &plonkErr))
            assert.Equal(t, tt.expectedCode, plonkErr.Code)
            assert.Contains(t, plonkErr.UserMessage(), tt.expectedErr)
        })
    }
}
```

#### Error Guidelines

1. **Never use `fmt.Errorf` for user-facing errors** - always use structured errors
2. **Include operation context** - specify what operation was being performed
3. **Use appropriate error codes** - select the most specific error code available
4. **Add item context when relevant** - package names, file paths, etc.
5. **Provide user-friendly messages** - explain what went wrong and how to fix it
6. **Test error conditions** - ensure error handling works correctly
7. **Use debug mode** - support `PLONK_DEBUG=1` for detailed error information

#### Exit Code Mapping

The error handling system automatically maps error codes to exit codes:

```go
// In HandleError function
switch plonkErr.Code {
case errors.ErrConfigNotFound, errors.ErrConfigParseFailure,
     errors.ErrConfigValidation, errors.ErrInvalidInput:
    return 1  // User error
case errors.ErrFilePermission, errors.ErrManagerUnavailable,
     errors.ErrInternal:
    return 2  // System error
default:
    return 1  // Default to user error
}
```

## Contributing

### Development Workflow

1. **Fork and clone** the repository
2. **Setup environment** (`just dev-setup`)
3. **Create feature branch** from main
4. **Make changes** following code style
5. **Add tests** for new functionality
6. **Run pre-commit checks** (automatic on commit)
7. **Commit changes** with descriptive messages
8. **Push branch** and create pull request

### Commit Messages

Use descriptive commit messages:

```bash
feat: add NPM package manager support
fix: resolve dotfile path resolution issue
docs: update CLI command documentation
test: add context cancellation tests
refactor: simplify configuration loading
```

### Pull Request Process

1. **Ensure tests pass** (`just precommit`)
2. **Update documentation** if needed
3. **Add examples** for new features
4. **Request review** from maintainers
5. **Address feedback** and update PR

### Adding Package Managers

1. **Implement interface** in `internal/managers/`
   - Follow existing patterns from `HomebrewManager` and `NpmManager`
2. **Add tests** with mocks and context support
3. **Register manager** in command layer
4. **Update documentation**
5. **Add configuration examples**

**Implementation requirements:**
- All methods must accept context for cancellation/timeout
- Handle expected conditions vs real errors appropriately
- Return structured errors using `internal/errors` package

### Adding Commands

1. **Create command file** in `internal/commands/`
2. **Register command** in `root.go`
3. **Add tests** for command logic
4. **Update CLI documentation**
5. **Add usage examples**

## Release Process

### Automated Release (Recommended)

Single command for complete release automation:

```bash
# Get version suggestions
just release-version-suggest

# Create automated release
just release-auto v1.2.3
```

**What it does:**
1. ✅ Validates version format and checks for duplicates
2. ✅ Ensures clean working directory
3. ✅ Runs full pre-release validation:
   - Run tests
   - Run linter
   - Run security checks
   - Test build
4. ✅ Creates and pushes git tag
5. ✅ Builds cross-platform release binaries
6. ✅ Provides clear success/failure feedback

### Manual Release (If Needed)

For emergency or special cases, you can manually execute the release steps:

```bash
# Manual steps (emergency use only)
git tag -a v1.2.3 -m "Release v1.2.3 - Description"
git push origin v1.2.3
# Build binaries manually for each platform if needed
```

### Release Checklist

The automated process (`just release-auto`) handles everything:
- ✅ Pre-release validation (tests, lint, security, build)
- ✅ Documentation generation
- ✅ Version validation
- ✅ Tag creation and push
- ✅ Cross-platform binary building

### Version Guidelines

- **Patch** (v1.2.3 → v1.2.4): Bug fixes, small improvements
- **Minor** (v1.2.3 → v1.3.0): New features, non-breaking changes
- **Major** (v1.2.3 → v2.0.0): Breaking changes, major rewrites
- **RC** (v1.2.3 → v1.3.0-rc1): Release candidates for testing

## Troubleshooting

### Common Issues

#### Build Failures
```bash
# Clean and rebuild
just clean
go mod tidy
just build
```

#### Test Failures
```bash
# Run specific failing test
go test -v -run TestFailingTest ./internal/config/

# Check test with race detection
go test -race ./...
```

#### Mock Issues
```bash
# Regenerate mocks
just generate-mocks

# Check mock interfaces match
go mod tidy
```

#### Dependency Issues
```bash
# Update all dependencies safely
just deps-update

# Or complete cleanup and reinstall
just clean-all
just dev-setup
```

### Debug Commands

```bash
# Verbose test output
go test -v ./...

# Build with debug info
go build -gcflags="-N -l" ./cmd/plonk

# Run with debug environment
PLONK_DEBUG=1 plonk status
```

## Performance Considerations

### Optimization Guidelines

- **Use context** for cancellation in long-running operations
- **Lazy initialization** of package managers
- **Minimal command execution** for system calls
- **Efficient file operations** with proper buffering

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## See Also

- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture
- [TESTING.md](TESTING.md) - Testing infrastructure
- [CODEMAP.md](CODEMAP.md) - Code navigation
