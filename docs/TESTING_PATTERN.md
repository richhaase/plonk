# Package Manager Testing Pattern

This document describes the new testing pattern for package managers in plonk, which enables comprehensive unit testing without requiring actual package managers to be installed.

## Overview

The new pattern introduces three key components:
1. **CommandExecutor Interface** - Abstracts command execution for testing
2. **ErrorMatcher** - Standardizes error detection across package managers
3. **Mock-based Unit Tests** - Test all scenarios without real commands

## Components

### 1. CommandExecutor Interface

Located in `internal/interfaces/executor.go`:

```go
type CommandExecutor interface {
    Execute(ctx context.Context, name string, args ...string) ([]byte, error)
    ExecuteCombined(ctx context.Context, name string, args ...string) ([]byte, error)
    LookPath(name string) (string, error)
}
```

**Implementations:**
- `RealCommandExecutor` - Uses `os/exec` for production
- `MockCommandExecutor` - Generated mock for testing

### 2. ErrorMatcher

Located in `internal/managers/error_matcher.go`:

```go
type ErrorMatcher struct {
    patterns []ErrorPattern
}

func (m *ErrorMatcher) MatchError(output string) ErrorType
```

**Error Types:**
- `ErrorTypeNotFound` - Package not found
- `ErrorTypePermission` - Permission denied
- `ErrorTypeAlreadyInstalled` - Already installed (success for install)
- `ErrorTypeNotInstalled` - Not installed (success for uninstall)
- `ErrorTypeLocked` - Resource locked (e.g., apt lock)

### 3. Package Manager Implementation

Example structure for a testable package manager:

```go
type PipManagerV2 struct {
    executor     CommandExecutor
    errorMatcher *ErrorMatcher
    // Manager-specific fields
}

// Production constructor
func NewPipManagerV2() *PipManagerV2 {
    return &PipManagerV2{
        executor:     &RealCommandExecutor{},
        errorMatcher: NewCommonErrorMatcher(),
    }
}

// Test constructor
func NewPipManagerV2WithExecutor(exec CommandExecutor) *PipManagerV2 {
    return &PipManagerV2{
        executor:     exec,
        errorMatcher: NewCommonErrorMatcher(),
    }
}
```

## Writing Unit Tests

### Test Structure

```go
func TestPipManagerV2_Install(t *testing.T) {
    tests := []struct {
        name        string
        packageName string
        mockSetup   func(m *mocks.MockCommandExecutor)
        wantErr     bool
        wantErrCode errors.ErrorCode
    }{
        {
            name:        "successful install",
            packageName: "requests",
            mockSetup: func(m *mocks.MockCommandExecutor) {
                m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
                m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "--user", "requests").
                    Return([]byte("Successfully installed requests-2.28.0"), nil)
            },
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockExecutor := mocks.NewMockCommandExecutor(ctrl)
            tt.mockSetup(mockExecutor)

            manager := NewPipManagerV2WithExecutor(mockExecutor)
            err := manager.Install(context.Background(), tt.packageName)

            // Assertions...
        })
    }
}
```

### Common Test Scenarios

1. **Success Cases**
   - Successful install/uninstall
   - Already installed (for install)
   - Not installed (for uninstall)

2. **Error Cases**
   - Package not found
   - Permission denied
   - Context cancellation
   - Command execution failure

3. **Edge Cases**
   - Fallback commands (pip → pip3)
   - Output parsing failures
   - Special error conditions

## Migration Guide

To migrate an existing package manager:

1. **Add executor field:**
   ```go
   type MyManager struct {
       executor CommandExecutor
       errorMatcher *ErrorMatcher
   }
   ```

2. **Update constructor:**
   ```go
   func NewMyManager() *MyManager {
       return &MyManager{
           executor: &RealCommandExecutor{},
           errorMatcher: NewCommonErrorMatcher(),
       }
   }
   ```

3. **Replace exec.Command calls:**
   ```go
   // Before:
   cmd := exec.CommandContext(ctx, "apt", "install", pkg)
   output, err := cmd.CombinedOutput()

   // After:
   output, err := m.executor.ExecuteCombined(ctx, "apt", "install", pkg)
   ```

4. **Use ErrorMatcher for error detection:**
   ```go
   errorType := m.errorMatcher.MatchError(string(output))
   switch errorType {
   case ErrorTypeNotFound:
       return errors.NewError(errors.ErrPackageNotFound, ...)
   case ErrorTypePermission:
       return errors.NewError(errors.ErrFilePermission, ...)
   // ...
   }
   ```

5. **Write comprehensive unit tests** using mocks

## Benefits

1. **Fast Tests** - No need for actual package managers
2. **Comprehensive Coverage** - Test all error scenarios
3. **CI/CD Friendly** - Tests run anywhere
4. **Consistent Error Handling** - Shared error patterns
5. **Easier Maintenance** - Less code duplication

## Best Practices

1. **Mock at the command level**, not the manager level
2. **Use `.AnyTimes()` for frequently called methods** like `LookPath`
3. **Test both stdout and stderr** scenarios
4. **Include context cancellation tests**
5. **Test manager-specific edge cases** (e.g., pip/pip3 fallback)

## Example: Complete Test File

See `internal/managers/pip_refactored_test.go` for a complete example of:
- Availability testing with fallbacks
- List/Install/Uninstall operations
- Error scenario testing
- Context cancellation handling
- Version detection

## Future Improvements

1. **Base Manager Class** - Extract common patterns
2. **Structured Output Parsing** - Prefer JSON where available
3. **Integration Test Suite** - Separate tests for real commands
4. **Performance Benchmarks** - Measure mock vs real performance
