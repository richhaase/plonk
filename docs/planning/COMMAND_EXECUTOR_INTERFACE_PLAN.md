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

func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
    m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args, Context: ctx})

    // Find matching response
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
   - Replace direct exec calls with executor calls
   - Update function signatures to accept executor
   - Maintain backward compatibility with wrapper functions

3. **Update package manager structs**
   - Add `executor CommandExecutor` field to each manager
   - Update constructors to accept optional executor
   - Default to RealCommandExecutor if not provided

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
type HomebrewManager struct {
    binary   string
    executor CommandExecutor
}

func NewHomebrewManager(executor CommandExecutor) *HomebrewManager {
    if executor == nil {
        executor = &RealCommandExecutor{}
    }
    return &HomebrewManager{
        binary:   "brew",
        executor: executor,
    }
}

func (h *HomebrewManager) Install(ctx context.Context, name string) error {
    output, err := h.executor.CombinedOutput(ctx, h.binary, "install", name)
    if err != nil {
        return h.handleInstallError(err, output, name)
    }
    return nil
}
```

## Example: Testing with Mock

```go
func TestHomebrewManager_Install(t *testing.T) {
    mock := &MockCommandExecutor{
        Responses: map[string]CommandResponse{
            "brew install vim": {
                Output: []byte("Installing vim..."),
                Error:  nil,
            },
            "brew install nonexistent": {
                Output: []byte("Error: No available formula"),
                Error:  &exec.ExitError{ProcessState: &os.ProcessState{}},
            },
        },
    }

    manager := NewHomebrewManager(mock)

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

1. Keep existing public APIs unchanged
2. Add executor parameter as optional (nil = use real executor)
3. Existing code continues to work without modification
4. Only test code uses mock executors

## Benefits

1. **Immediate**: Unit test package managers without system modification
2. **Coverage**: Achieve 60%+ test coverage for package operations
3. **Safety**: Developers can run all tests locally
4. **Speed**: Unit tests run in milliseconds, not minutes
5. **Reliability**: No flaky tests due to network/package manager issues

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
