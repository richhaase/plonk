# Command Executor Interface Implementation Plan

## Overview

This plan describes how to introduce a Command Executor Interface to enable unit testing of package managers without executing real system commands. This is a minimal refactoring approach that maintains backward compatibility while enabling comprehensive testing.

## Problem Statement

Current package managers directly call `exec.CommandContext`, making them impossible to unit test without:
- Installing real packages
- Modifying the developer's system
- Running in containers/VMs

This tight coupling prevents us from achieving adequate test coverage for v1.0.

## Solution: Command Executor Interface

### 1. Interface Definition

```go
// CommandExecutor abstracts command execution for testability
type CommandExecutor interface {
    // Execute runs a command and returns stdout
    Execute(ctx context.Context, name string, args ...string) ([]byte, error)

    // CombinedOutput runs a command and returns combined stdout/stderr
    CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)

    // LookPath searches for an executable in PATH
    LookPath(name string) (string, error)
}
```

### 2. Real Implementation

```go
// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, name, args...)
    return cmd.Output()
}

func (r *RealCommandExecutor) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, name, args...)
    return cmd.CombinedOutput()
}

func (r *RealCommandExecutor) LookPath(name string) (string, error) {
    return exec.LookPath(name)
}

// Package-level default executor
var defaultExecutor CommandExecutor = &RealCommandExecutor{}

// SetDefaultExecutor allows tests to override the executor
func SetDefaultExecutor(executor CommandExecutor) {
    defaultExecutor = executor
}

// Updated helper functions use defaultExecutor
func ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
    return defaultExecutor.Execute(ctx, name, args...)
}

func ExecuteCommandCombined(ctx context.Context, name string, args ...string) ([]byte, error) {
    return defaultExecutor.CombinedOutput(ctx, name, args...)
}

func CheckCommandAvailable(name string) bool {
    _, err := defaultExecutor.LookPath(name)
    return err == nil
}
```

### 3. Mock Implementation

```go
// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
    // Commands records all executed commands
    Commands []ExecutedCommand

    // Responses maps command patterns to responses
    Responses map[string]CommandResponse

    // DefaultResponse is used when no pattern matches
    DefaultResponse CommandResponse
}

type ExecutedCommand struct {
    Name string
    Args []string
    Context context.Context
}

type CommandResponse struct {
    Output []byte
    Error  error
}

// MockExitError implements the minimal interface that package managers check for
type MockExitError struct {
    Code int
}

func (e *MockExitError) Error() string {
    return fmt.Sprintf("exit status %d", e.Code)
}

func (e *MockExitError) ExitCode() int {
    return e.Code
}

func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
    m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args, Context: ctx})

    // Find matching response using simple string matching
    // This is intentionally simple - exact matches only for v1.0
    key := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
    if resp, ok := m.Responses[key]; ok {
        return resp.Output, resp.Error
    }

    return m.DefaultResponse.Output, m.DefaultResponse.Error
}
```

## Implementation Steps

### Phase 1: Core Infrastructure (2-3 hours)

1. **Create executor.go in packages/**
   - Define CommandExecutor interface
   - Implement RealCommandExecutor
   - Implement MockCommandExecutor
   - Add factory function: `NewDefaultExecutor() CommandExecutor`

2. **Update helpers.go**
   - Add package-level default executor variable
   - Add SetDefaultExecutor function for test override
   - Replace direct exec calls with defaultExecutor calls
   - No changes to function signatures needed

3. **Update package managers**
   - No struct changes needed
   - Managers continue to use helper functions
   - Helper functions now use defaultExecutor internally

### Phase 2: Package Manager Updates (3-4 hours)

Update each package manager to use the executor:

1. **HomebrewManager**
   - Replace exec.CommandContext calls
   - Update tests to use MockCommandExecutor
   - Test all methods without real brew commands

2. **NPMManager**
   - Similar updates
   - Mock npm responses

3. **Other Managers** (pip, cargo, gem, go)
   - Apply same pattern
   - Create appropriate mock responses

### Phase 3: Testing (4-6 hours)

1. **Create test fixtures**
   - Common command outputs for each package manager
   - Error responses
   - Edge cases

2. **Write comprehensive unit tests**
   - Test all package manager methods
   - Test error handling
   - Test parsing of various output formats

3. **Verify coverage improvement**
   - Target 60%+ coverage for package managers
   - Document remaining gaps

## Example: Updated HomebrewManager

```go
// HomebrewManager struct remains unchanged
type HomebrewManager struct {
    binary string
}

// Constructor remains unchanged
func NewHomebrewManager() *HomebrewManager {
    return &HomebrewManager{
        binary: "brew",
    }
}

// Methods remain unchanged - they use helpers which now use defaultExecutor
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
    output, err := ExecuteCommandCombined(ctx, h.binary, "install", name)
    if err != nil {
        return h.handleInstallError(err, output, name)
    }
    return nil
}
```

## Example: Testing with Mock

```go
func TestHomebrewManager_Install(t *testing.T) {
    // Save original executor and restore after test
    originalExecutor := defaultExecutor
    defer func() { defaultExecutor = originalExecutor }()

    // Set up mock executor
    mock := &MockCommandExecutor{
        Responses: map[string]CommandResponse{
            "brew install vim": {
                Output: []byte("Installing vim..."),
                Error:  nil,
            },
            "brew install nonexistent": {
                Output: []byte("Error: No available formula"),
                Error:  &MockExitError{Code: 1},
            },
        },
    }

    // Override the default executor
    SetDefaultExecutor(mock)

    // Create manager - it will use the mock executor via helpers
    manager := NewHomebrewManager()

    // Test successful install
    err := manager.Install(context.Background(), "vim")
    assert.NoError(t, err)
    assert.Len(t, mock.Commands, 1)
    assert.Equal(t, "brew", mock.Commands[0].Name)
    assert.Equal(t, []string{"install", "vim"}, mock.Commands[0].Args)

    // Test failed install
    err = manager.Install(context.Background(), "nonexistent")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

## Backward Compatibility

To maintain backward compatibility:

1. No changes to package manager APIs or constructors
2. No changes to helper function signatures
3. Existing code continues to work without modification
4. Only test code calls SetDefaultExecutor to override behavior
5. Tests must restore original executor to avoid affecting other tests

## Benefits

1. **Immediate**: Unit test package managers without system modification
2. **Coverage**: Achieve 60%+ test coverage for package operations
3. **Safety**: Developers can run all tests locally
4. **Speed**: Unit tests run in milliseconds, not minutes
5. **Reliability**: No flaky tests due to network/package manager issues

## Known Limitations

### Go Install Manager
The Go install manager has methods that depend on `os.Stat` to check if binary files exist:
- `InstalledVersion` - checks if binary exists before getting version
- `Info` - checks if binary is installed before retrieving info
- `IsInstalled` - directly checks file existence

These methods cannot be fully tested with the Command Executor pattern since we only mock command execution, not file system operations. Integration tests should cover these methods.

## Risks and Mitigations

### Risk: Mock responses become outdated
**Mitigation**: Regular integration tests in CI verify mock accuracy

### Risk: Added complexity
**Mitigation**: Minimal interface, clear separation of concerns

### Risk: Missing edge cases
**Mitigation**: Gradually add test cases as bugs are found

## Success Criteria

1. All package managers use CommandExecutor
2. Unit test coverage for packages/ exceeds 60%
3. All tests can run safely on developer machines
4. No breaking changes to existing code
5. Clear documentation for adding new package managers

## Timeline

- Day 1: Implement core infrastructure and update 2-3 package managers
- Day 2: Complete remaining package managers and write tests
- Day 3: Fix edge cases, improve coverage, update documentation

Total: 2-3 days of focused work

## Future Enhancements

After v1.0, consider:
1. Async command execution support
2. Command timeout configuration
3. Retry logic abstraction
4. Command output streaming
