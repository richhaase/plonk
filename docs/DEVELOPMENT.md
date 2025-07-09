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

# Install development dependencies
go mod download

# Generate mocks for testing
just generate-mocks

# Run tests
just test

# Build binary
just build

# Install locally
just install
```

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

# Code quality
just format             # Format code and organize imports
just lint               # Run linter
just security           # Run security checks
just precommit          # Run all pre-commit checks

# Mock generation
just generate-mocks     # Generate test mocks

# Release (interactive)
just release            # Create release tag
just goreleaser-release # Build release binaries
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

### Key Interfaces
For complete interface specifications, see:
- **Configuration interfaces:** `docs/api/config.md`
- **Package manager interface:** `docs/api/managers.md`
- **State provider interface:** `docs/api/state.md`
- **Error handling types:** `docs/api/errors.md`

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

### Pre-commit Checks

```bash
# Run all pre-commit checks
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

Use structured errors for user-facing messages:
- See `docs/api/errors.md` for complete error type specifications
- Use `errors.NewPlonkError()` for creating structured errors
- Include error codes, domains, and user-friendly messages
- Follow existing error patterns in the codebase

## Contributing

### Development Workflow

1. **Fork and clone** the repository
2. **Create feature branch** from main
3. **Make changes** following code style
4. **Add tests** for new functionality
5. **Run pre-commit checks** (`just precommit`)
6. **Commit changes** with descriptive messages
7. **Push branch** and create pull request

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
   - See `docs/api/managers.md` for complete PackageManager interface
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

### Version Management

Plonk uses semantic versioning with Git tags:

```bash
# Create release (interactive)
just release

# Manual tagging
git tag -a v1.2.3 -m "Release v1.2.3 - Description"
git push origin v1.2.3
```

### Release Build

```bash
# Build release binaries
just goreleaser-release
```

### Release Checklist

1. ✅ All tests pass (`just precommit`)
2. ✅ Documentation updated
3. ✅ Version bumped appropriately
4. ✅ Release notes prepared
5. ✅ Tag created and pushed
6. ✅ Release binaries built
7. ✅ GitHub release created

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
- [GODOC.md](GODOC.md) - API documentation