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
- **Configuration parsing** (`internal/config/*_test.go`)
- **State reconciliation** (`internal/state/*_test.go`)
- **Package manager operations** (`internal/managers/*_test.go`)
- **Dotfile operations** (`internal/dotfiles/*_test.go`)
- **Apply command** (`internal/commands/apply_test.go`) - Unified apply functionality
- **Error handling** (`internal/errors/*_test.go`)

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

### Debug Commands
```bash
# Run tests with verbose output
go test -v ./...

# Debug specific test
go test -v -run TestSpecificFunction ./path/to/test/
```

## CI/CD Integration

The test suite is designed to work in CI environments:
- Unit tests run on every commit
- Pre-commit hooks ensure code quality
- Security scanning with `gosec` and `govulncheck`