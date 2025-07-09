# Testing Guide

This guide covers the testing infrastructure and practices for the Plonk project.

## Overview

Plonk uses unit tests for comprehensive testing:
- **Unit tests**: Fast, isolated tests for individual components

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

## Running Tests

### Quick Commands (via justfile)

```bash
# Run unit tests
just test
```

### Direct Go Commands

```bash
# Unit tests
go test ./...
```

## Test Patterns and Conventions

### Unit Test Patterns
- **Mocking**: Mock providers implement interfaces for isolated testing
- **Context handling**: Tests verify context cancellation and timeouts
- **Error scenarios**: Comprehensive error condition testing
- **State management**: Mock providers for testing reconciliation logic

## Test Categories

### Unit Test Categories
- **Configuration parsing** (`internal/config/*_test.go`) - YAML parsing, validation, ignore patterns
- **State reconciliation** (`internal/state/*_test.go`) - State management and provider interfaces
- **Package manager operations** (`internal/managers/*_test.go`) - Homebrew and NPM implementations
- **Dotfile operations** (`internal/dotfiles/*_test.go`) - Auto-discovery, file operations
- **Apply command** (`internal/commands/apply_test.go`) - Unified apply functionality
- **Error handling** (`internal/errors/*_test.go`) - Structured error types

## Development Workflow

### Before Committing
```bash
# Run pre-commit checks (includes unit tests)
just precommit
```

### Adding Tests
1. **Unit tests**: Add `*_test.go` files alongside source code
2. **Follow existing patterns**: Use established helper functions and mocking

### Test Performance
- **Unit tests**: ~1-2 seconds (fast feedback)

## Troubleshooting

### Common Issues
- **Timeout errors**: Use `-timeout` flag if tests are timing out
- **Permission errors**: Ensure proper file permissions
- **Environment variables**: Tests may use `PLONK_DIR` for config directory override

### Debug Commands
```bash
# Run tests with verbose output
go test -v ./...

# Debug specific test
go test -v -run TestSpecificFunction ./path/to/test/

# Test with custom config directory
PLONK_DIR=/tmp/test-config go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Profile tests
go test -cpuprofile=cpu.prof ./...
go test -memprofile=mem.prof ./...
```

## CI/CD Integration

The test suite is designed to work in CI environments:
- Unit tests run on every commit
- Pre-commit hooks ensure code quality
- Security scanning with `gosec` and `govulncheck`