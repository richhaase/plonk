# Testing Guide

This guide covers the testing infrastructure and practices for the Plonk project.

## Overview

Plonk uses a comprehensive testing strategy with two main layers:
- **Unit tests**: Fast, isolated tests for individual components
- **Integration tests**: End-to-end tests running in Docker containers

## Test Organization

### Unit Tests
- Location: Alongside source code files as `*_test.go`
- Framework: Go's built-in `testing` package
- Coverage: Core business logic, state management, configuration parsing, package managers

Key unit test areas:
- `internal/config/` - Configuration parsing and validation
- `internal/state/` - State reconciliation and provider interfaces
- `internal/managers/` - Package manager implementations
- `internal/dotfiles/` - Dotfile operations and atomic updates
- `internal/errors/` - Error handling and structured errors

### Integration Tests
- Location: `test/integration/`
- Framework: Go testing with Docker containers
- Environment: Ubuntu 22.04 with Homebrew, NPM, and system packages
- Build tag: `integration` (requires `-tags=integration`)

## Running Tests

### Quick Commands (via justfile)

```bash
# Run unit tests only
just test

# Run integration tests (requires Docker)
just test-integration

# Run all tests
just test-all

# Fast integration tests (shorter timeout)
just test-integration-fast
```

### Direct Go Commands

```bash
# Unit tests
go test ./...

# Integration tests (with Docker setup)
go test -tags=integration -v ./test/integration/... -timeout=10m

# Integration tests (short mode)
go test -tags=integration -v ./test/integration/... -timeout=5m -short
```

## Integration Test Setup

### Prerequisites
- Docker installed and running
- Sufficient disk space for Ubuntu container (~500MB)

### Docker Environment
Integration tests run in a controlled Ubuntu 22.04 environment with:
- Go 1.21.6
- Homebrew (Linux)
- Node.js and NPM
- System package managers (apt)
- Non-root test user with sudo access

### Setup Process
1. **Automatic setup**: Integration tests automatically build the Docker image
2. **Manual setup**: `just test-integration-setup` to build image only
3. **Cleanup**: `just clean-docker` to remove test images

## Test Patterns and Conventions

### Unit Test Patterns
- **Mocking**: Mock providers implement interfaces for isolated testing
- **Context handling**: Tests verify context cancellation and timeouts
- **Error scenarios**: Comprehensive error condition testing
- **State management**: Mock providers for testing reconciliation logic

### Integration Test Patterns
- **Docker isolation**: Each test runs in a fresh container
- **Helper functions**: Shared utilities in `test/integration/helpers.go`
- **Fixture files**: Test configurations in `test/integration/fixtures/`
- **Real package managers**: Tests interact with actual Homebrew, NPM, etc.

### Test Helpers
Key helper functions available:
- `DockerRunner`: Execute commands in Docker containers
- `RequireDockerImage()`: Skip tests if Docker unavailable
- `CreateTempConfigFile()`: Generate test configuration files
- `CleanupBuildArtifacts()`: Clean up after tests

## Test Categories

### Unit Test Categories
- **Configuration parsing** (`internal/config/*_test.go`)
- **State reconciliation** (`internal/state/*_test.go`)
- **Package manager operations** (`internal/managers/*_test.go`)
- **Dotfile operations** (`internal/dotfiles/*_test.go`)
- **Apply command** (`internal/commands/apply_test.go`) - Unified apply functionality
- **Error handling** (`internal/errors/*_test.go`)

### Integration Test Categories
- **Binary functionality** (`binary_test.go`)
- **Apply command** (`apply_test.go`) - Unified package and dotfile application
- **Import command** (`import_test.go`, `import_modes_test.go`, `import_large_test.go`) - Package and dotfile discovery and import
- **Config commands** (`config_commands_test.go`) - Configuration validation, editing, and environment information
- **Package management** (`packages_test.go`)
- **Configuration validation** (`config_test.go`)
- **Dotfile operations** (`dotfiles_test.go`)
- **State management** (`state_test.go`)
- **Cross-manager workflows** (`cross_manager_test.go`)
- **Error recovery** (`error_recovery_test.go`)
- **Security validation** (`security_test.go`)
- **Performance benchmarks** (`performance_test.go`)

## Development Workflow

### Before Committing
```bash
# Run pre-commit checks (includes unit tests)
just precommit

# Run full checks (includes integration tests)
just precommit-full
```

### Adding Tests
1. **Unit tests**: Add `*_test.go` files alongside source code
2. **Integration tests**: Add to `test/integration/` with `//go:build integration`
3. **Test fixtures**: Place in `test/integration/fixtures/`
4. **Follow existing patterns**: Use established helper functions and mocking

### Test Performance
- **Unit tests**: ~1-2 seconds (fast feedback)
- **Integration tests**: ~2-5 minutes (Docker overhead)
- **Full test suite**: ~3-7 minutes

## Troubleshooting

### Common Issues
- **Docker not available**: Integration tests will be skipped
- **Docker image build fails**: Check Dockerfile and network connectivity
- **Timeout errors**: Use `-timeout` flag or `test-integration-fast`
- **Permission errors**: Ensure Docker daemon is running

### Debug Commands
```bash
# Check Docker setup
docker image inspect plonk-test

# Run tests with verbose output
go test -v -tags=integration ./test/integration/...

# Debug specific test
go test -v -run TestSpecificFunction ./path/to/test/
```

## CI/CD Integration

The test suite is designed to work in CI environments:
- Unit tests run on every commit
- Integration tests require Docker support
- Pre-commit hooks ensure code quality
- Security scanning with `gosec` and `govulncheck`