# Commands Package Testing Plan

**Date**: 2025-08-02
**Author**: Analysis by Claude
**Current Coverage**: 9.2% (421 of 4,577 statements)
**Target Coverage**: 45-50% (adding ~2,000 statements)

## Executive Summary

The `internal/commands` package is the largest untested package in plonk, representing 4,577 lines of code with only 9.2% coverage. This document provides a comprehensive plan to improve test coverage using idiomatic Go patterns that leverage existing interfaces and require minimal refactoring.

## Current State Analysis

### Package Structure

```
internal/commands/
├── Core Commands (22 files, ~4,577 LOC)
│   ├── status.go (524 lines) - 0% coverage
│   ├── env.go (358 lines) - 0% coverage
│   ├── info.go (345 lines) - 0% coverage
│   ├── config_edit.go (341 lines) - 0% coverage
│   └── ... (18 more command files)
├── Test Files (8 files)
│   ├── helpers_test.go - 100% coverage
│   ├── output_*_test.go - partial coverage
│   └── diff_test.go - 81.8% coverage
```

### Key Architectural Findings

1. **Clean Separation**: Commands are thin CLI adapters that delegate to orchestrator/resources
2. **Existing Test Infrastructure**:
   - `CommandExecutor` interface in `internal/resources/packages/executor.go`
   - `LockService` interface pattern exists
   - Functional options pattern for dependency injection
3. **No Business Logic**: Most commands just:
   - Parse CLI flags
   - Create orchestrator with options
   - Call orchestrator methods
   - Format output

### Coverage Analysis by Function Type

| Function Type | Current Coverage | Files |
|--------------|------------------|-------|
| `init()` functions | 100% | All command files |
| Helper functions | ~90% | helpers.go, output_utils.go |
| `run*()` functions | 0% | All command files |
| Output formatting | 0% | TableOutput(), StructuredData() |
| Error handling | 0% | Most error paths |

## Testing Strategy: Idiomatic Go Approach

### Core Principle: Test Through Existing Interfaces

The codebase already has the right testing seams. We don't need factories or heavy abstractions - just expose and use what's already there.

### Key Interfaces Already Available

```go
// internal/resources/packages/executor.go
type CommandExecutor interface {
    Run(command string, args ...string) (string, error)
    RunWithInput(command string, input string, args ...string) (string, error)
    RunInDir(dir string, command string, args ...string) (string, error)
    CommandExists(command string) bool
}

// This can be mocked for testing!
```

## Implementation Plan

### Phase 1: Enable Testing Infrastructure (Day 1)

#### 1.1 Expose Test Hooks in Package Layer

```go
// internal/resources/packages/testing.go (new file)
package packages

import "sync"

var (
    testExecutor CommandExecutor
    testMutex    sync.RWMutex
)

// SetTestExecutor sets a custom executor for testing
func SetTestExecutor(executor CommandExecutor) {
    testMutex.Lock()
    defer testMutex.Unlock()
    testExecutor = executor
}

// ResetTestExecutor restores the default executor
func ResetTestExecutor() {
    testMutex.Lock()
    defer testMutex.Unlock()
    testExecutor = nil
}

// getExecutor returns test executor if set, otherwise default
func getExecutor() CommandExecutor {
    testMutex.RLock()
    defer testMutex.RUnlock()
    if testExecutor != nil {
        return testExecutor
    }
    return &RealCommandExecutor{}
}
```

#### 1.2 Create Test Helper Package

```go
// internal/commands/testutil/helpers.go (new file)
package testutil

import (
    "bytes"
    "testing"
    "github.com/spf13/cobra"
)

// ExecuteCommand is a helper to test cobra commands
func ExecuteCommand(root *cobra.Command, args ...string) (output string, err error) {
    buf := new(bytes.Buffer)
    root.SetOut(buf)
    root.SetErr(buf)
    root.SetArgs(args)

    err = root.Execute()
    return buf.String(), err
}

// MockExecutor provides a simple mock for CommandExecutor
type MockExecutor struct {
    responses map[string]string
    errors    map[string]error
    calls     []string
}

func (m *MockExecutor) Run(command string, args ...string) (string, error) {
    call := command + " " + strings.Join(args, " ")
    m.calls = append(m.calls, call)

    if err, ok := m.errors[call]; ok {
        return "", err
    }
    if resp, ok := m.responses[call]; ok {
        return resp, nil
    }
    return "", nil
}
```

### Phase 2: Test High-Value Commands (Week 1)

Focus on commands that provide maximum coverage increase:

#### 2.1 Status Command Testing Pattern

```go
// internal/commands/status_test.go
package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/richhaase/plonk/internal/commands/testutil"
    "github.com/richhaase/plonk/internal/resources/packages"
)

func TestStatusCommand(t *testing.T) {
    tests := []struct {
        name      string
        args      []string
        setup     func(*testutil.MockExecutor)
        wantErr   bool
        contains  []string
        excludes  []string
    }{
        {
            name: "empty status",
            args: []string{"status"},
            setup: func(m *testutil.MockExecutor) {
                m.responses["brew list"] = ""
                m.responses["npm list -g --depth=0"] = ""
            },
            wantErr: false,
            contains: []string{"No packages"},
        },
        {
            name: "with packages",
            args: []string{"status"},
            setup: func(m *testutil.MockExecutor) {
                m.responses["brew list"] = "ripgrep\nfd\nbat"
                m.responses["brew info ripgrep"] = "ripgrep: stable 13.0.0"
                // ... more responses
            },
            wantErr: false,
            contains: []string{"ripgrep", "13.0.0", "3 packages"},
        },
        {
            name: "json output",
            args: []string{"status", "--output", "json"},
            setup: func(m *testutil.MockExecutor) {
                m.responses["brew list"] = "ripgrep"
            },
            wantErr: false,
            contains: []string{`"name":"ripgrep"`, `"manager":"brew"`},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock
            mock := &testutil.MockExecutor{
                responses: make(map[string]string),
                errors:    make(map[string]error),
            }
            if tt.setup != nil {
                tt.setup(mock)
            }

            // Set test executor
            packages.SetTestExecutor(mock)
            defer packages.ResetTestExecutor()

            // Execute command
            cmd := NewRootCmd()
            output, err := testutil.ExecuteCommand(cmd, tt.args...)

            // Verify
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }

            for _, want := range tt.contains {
                assert.Contains(t, output, want)
            }
            for _, exclude := range tt.excludes {
                assert.NotContains(t, output, exclude)
            }
        })
    }
}
```

#### 2.2 Command Priority Order

Test these commands in order for maximum impact:

1. **status** (524 lines) - Shows current state
2. **apply** (254 lines) - Core functionality
3. **install** (simple but critical)
4. **info** (345 lines) - Package information
5. **search** (328 lines) - Complex parallel logic
6. **env** (358 lines) - Environment info
7. **doctor** - System diagnostics

### Phase 3: Test Output Formatting (Week 2)

Each command has `TableOutput()` and `StructuredData()` methods that need testing:

```go
func TestStatusTableOutput(t *testing.T) {
    tests := []struct {
        name     string
        result   *StatusResult
        expected []string
    }{
        {
            name: "formats package table",
            result: &StatusResult{
                Packages: []PackageInfo{{
                    Name:    "ripgrep",
                    Manager: "brew",
                    Version: "13.0.0",
                }},
            },
            expected: []string{
                "Package Summary",
                "ripgrep",
                "brew",
                "13.0.0",
            },
        },
    }
    // ... test implementation
}
```

### Phase 4: Edge Cases and Error Paths (Week 3)

Focus on error conditions and edge cases:

```go
func TestCommandErrors(t *testing.T) {
    tests := []struct {
        name    string
        command string
        setup   func(*testutil.MockExecutor)
        wantErr string
    }{
        {
            name:    "brew not found",
            command: "status",
            setup: func(m *testutil.MockExecutor) {
                m.errors["brew list"] = errors.New("brew: command not found")
            },
            wantErr: "brew is not installed",
        },
        // ... more error cases
    }
}
```

## File-by-File Testing Guide

### High Priority Files (test first)

1. **status.go** (524 lines)
   - Test: Empty status, with packages, missing packages, multiple managers
   - Mock: CommandExecutor responses for brew/npm/cargo list commands
   - Coverage target: 70%

2. **apply.go** (254 lines)
   - Test: Dry run, normal run, partial failures
   - Mock: Orchestrator results
   - Coverage target: 80%

3. **install.go** (relatively simple)
   - Test: Single package, multiple packages, with manager, invalid package
   - Mock: CommandExecutor for install commands
   - Coverage target: 90%

### Medium Priority Files

4. **info.go** (345 lines)
   - Test: Package found/not found, multiple managers
   - Mock: Package manager info commands

5. **search.go** (328 lines)
   - Test: Parallel search, single manager search
   - Mock: Search command responses

6. **env.go** (358 lines)
   - Test: Environment detection
   - Mock: File system checks, command existence

### Lower Priority Files

- config_*.go - Complex file I/O, consider integration tests
- doctor.go - Mostly delegates to diagnostics package
- clone.go - Better tested via integration tests

## Mock Patterns Reference

### Simple Response Mock
```go
mock.responses["brew list"] = "package1\npackage2"
```

### Error Mock
```go
mock.errors["npm install foo"] = errors.New("E404 Not Found")
```

### Conditional Mock
```go
func (m *MockExecutor) Run(cmd string, args ...string) (string, error) {
    if cmd == "brew" && len(args) > 0 && args[0] == "info" {
        pkg := args[1]
        return fmt.Sprintf("%s: stable 1.0.0", pkg), nil
    }
    return "", nil
}
```

## Testing Best Practices

1. **Use Table-Driven Tests**: Group related test cases
2. **Test One Thing**: Each test case should verify one behavior
3. **Mock at the Right Level**: Mock CommandExecutor, not internal methods
4. **Test Output Formats**: Both human-readable and JSON/YAML
5. **Cleanup**: Always defer cleanup of test executors
6. **Parallel Tests**: Use `t.Parallel()` where possible

## Expected Coverage Improvements

| Phase | Commands Tested | Expected Coverage | Time |
|-------|----------------|-------------------|------|
| Current | helpers only | 9.2% | - |
| Phase 1 | Infrastructure | 9.2% | 1 day |
| Phase 2 | status, apply, install | ~25% | 1 week |
| Phase 3 | Output formatting | ~35% | 3 days |
| Phase 4 | Error paths | ~45% | 3 days |

## Success Metrics

- [ ] Commands package reaches 45%+ coverage
- [ ] All critical commands (status, apply, install) have tests
- [ ] No production code changes required
- [ ] Tests run in under 5 seconds
- [ ] Mock patterns established for future tests

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking production code | Tests only, no production changes |
| Complex mocking | Use existing CommandExecutor interface |
| Time investment | Focus on high-value commands first |
| Maintenance burden | Establish clear patterns early |

## Next Steps

1. Create `internal/resources/packages/testing.go` with test hooks
2. Create `internal/commands/testutil/` package with helpers
3. Write first test for `status` command as pattern reference
4. Continue with priority order list

This plan provides ~35% improvement in commands package coverage, which translates to ~15% improvement in overall project coverage (from 30% to 45%), meeting the v1.0 target.
