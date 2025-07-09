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
- **Structured error testing**: Verify error codes, domains, and user messages

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

## Error Handling Testing

### Testing Structured Errors

Plonk uses a structured error system that requires specific testing patterns:

#### Basic Error Testing
```go
func TestCommandError(t *testing.T) {
    // Test error creation
    err := runCommand(cmd, []string{"invalid"})
    
    // Verify it's a structured error
    var plonkErr *errors.PlonkError
    assert.True(t, errors.As(err, &plonkErr))
    
    // Check error properties
    assert.Equal(t, errors.ErrInvalidInput, plonkErr.Code)
    assert.Equal(t, errors.DomainCommands, plonkErr.Domain)
    assert.Equal(t, "validate", plonkErr.Operation)
    assert.Contains(t, plonkErr.UserMessage(), "invalid input")
}
```

#### Table-Driven Error Testing
```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name         string
        setupError   error
        expectedCode errors.ErrorCode
        expectedDomain errors.ErrorDomain
        expectedMessage string
    }{
        {
            name:         "config not found",
            setupError:   os.ErrNotExist,
            expectedCode: errors.ErrConfigNotFound,
            expectedDomain: errors.DomainConfig,
            expectedMessage: "configuration file not found",
        },
        {
            name:         "package manager unavailable",
            setupError:   fmt.Errorf("command not found"),
            expectedCode: errors.ErrManagerUnavailable,
            expectedDomain: errors.DomainPackages,
            expectedMessage: "package manager not available",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup error condition
            err := functionUnderTest(tt.setupError)
            
            // Verify structured error
            var plonkErr *errors.PlonkError
            require.True(t, errors.As(err, &plonkErr))
            assert.Equal(t, tt.expectedCode, plonkErr.Code)
            assert.Equal(t, tt.expectedDomain, plonkErr.Domain)
            assert.Contains(t, plonkErr.UserMessage(), tt.expectedMessage)
        })
    }
}
```

#### Error Context Testing
```go
func TestErrorContext(t *testing.T) {
    packageName := "test-package"
    
    // Create error with item context
    err := errors.WrapWithItem(
        fmt.Errorf("install failed"),
        errors.ErrPackageInstall,
        errors.DomainPackages,
        "install",
        packageName,
        "failed to install package"
    )
    
    var plonkErr *errors.PlonkError
    require.True(t, errors.As(err, &plonkErr))
    
    // Verify context is preserved
    assert.Equal(t, packageName, plonkErr.Item)
    assert.Contains(t, plonkErr.UserMessage(), packageName)
}
```

#### Error Wrapping Testing
```go
func TestErrorWrapping(t *testing.T) {
    originalErr := fmt.Errorf("original error")
    
    // Wrap error with context
    wrappedErr := errors.Wrap(
        originalErr,
        errors.ErrFileIO,
        errors.DomainDotfiles,
        "copy",
        "failed to copy file"
    )
    
    // Verify wrapping preserves original error
    assert.ErrorIs(t, wrappedErr, originalErr)
    
    // Verify structured error properties
    var plonkErr *errors.PlonkError
    require.True(t, errors.As(wrappedErr, &plonkErr))
    assert.Equal(t, errors.ErrFileIO, plonkErr.Code)
    assert.Equal(t, errors.DomainDotfiles, plonkErr.Domain)
}
```

### Testing Exit Codes

```go
func TestExitCodes(t *testing.T) {
    tests := []struct {
        name         string
        error        error
        expectedCode int
    }{
        {
            name: "success",
            error: nil,
            expectedCode: 0,
        },
        {
            name: "user error",
            error: errors.NewError(errors.ErrInvalidInput, errors.DomainCommands, "test", "invalid input"),
            expectedCode: 1,
        },
        {
            name: "system error",
            error: errors.NewError(errors.ErrFilePermission, errors.DomainDotfiles, "test", "permission denied"),
            expectedCode: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            exitCode := commands.HandleError(tt.error)
            assert.Equal(t, tt.expectedCode, exitCode)
        })
    }
}
```

### Testing Error Messages

```go
func TestErrorMessages(t *testing.T) {
    err := errors.NewError(
        errors.ErrManagerUnavailable,
        errors.DomainPackages,
        "check",
        "homebrew is not available"
    ).WithItem("homebrew")
    
    userMessage := err.UserMessage()
    
    // Verify user-friendly message
    assert.Contains(t, userMessage, "homebrew")
    assert.Contains(t, userMessage, "not available")
    
    // Verify it doesn't contain technical details
    assert.NotContains(t, userMessage, "stack trace")
    assert.NotContains(t, userMessage, "internal error")
}
```

### Testing Debug Mode

```go
func TestDebugMode(t *testing.T) {
    // Set debug mode
    originalDebug := os.Getenv("PLONK_DEBUG")
    defer os.Setenv("PLONK_DEBUG", originalDebug)
    
    os.Setenv("PLONK_DEBUG", "1")
    
    err := errors.NewError(
        errors.ErrInternal,
        errors.DomainCommands,
        "test",
        "internal error"
    )
    
    // In debug mode, HandleError should show technical details
    // This would typically be tested by capturing stderr output
    exitCode := commands.HandleError(err)
    assert.Equal(t, 2, exitCode)
}
```

### Error Testing Guidelines

1. **Always test error conditions** - Every error path should have a test
2. **Use structured error assertions** - Verify error codes, domains, and messages
3. **Test error context** - Ensure item and operation context is preserved
4. **Test error wrapping** - Verify original errors are preserved
5. **Test user messages** - Ensure messages are user-friendly and actionable
6. **Test exit codes** - Verify proper exit code mapping
7. **Test debug mode** - Ensure debug information is available when needed

### Error Test Examples

See these test files for comprehensive error testing examples:
- `internal/commands/apply_test.go` - Command error handling
- `internal/errors/types_test.go` - Error type behavior
- `internal/managers/homebrew_test.go` - Package manager error handling
- `internal/config/yaml_config_test.go` - Configuration error handling

## CI/CD Integration

The test suite is designed to work in CI environments:
- Unit tests run on every commit
- Pre-commit hooks ensure code quality
- Security scanning with `gosec` and `govulncheck`
- Error handling validation in all test suites