# Diagnostics Package Testing Plan

**Date**: 2025-08-02
**Author**: Analysis by Claude
**Current Coverage**: 13.7% (109 of 798 statements)
**Target Coverage**: 50-60% (adding ~300-400 statements)

## Executive Summary

The `internal/diagnostics` package performs system health checks for plonk, directly interacting with the operating system, filesystem, and external commands. Currently at 13.7% coverage, this document provides a testing strategy that balances pragmatism with thorough testing, using idiomatic Go patterns to achieve 50-60% coverage.

## Current State Analysis

### Package Structure

```
internal/diagnostics/
├── health.go (730 lines)
│   ├── 13 check functions (0% coverage each)
│   ├── 5 utility functions (50-100% coverage)
│   └── Direct system calls throughout
└── health_test.go (269 lines)
    ├── Tests for utilities only
    └── No mocking infrastructure
```

### Coverage Breakdown

| Function Type | Coverage | Count | Challenge |
|--------------|----------|-------|-----------|
| System checks | 0% | 13 functions | Direct OS/exec calls |
| Pure utilities | 76-100% | 5 functions | Already well tested |
| External command helpers | 0% | 3 functions | Need exec mocking |

### Key Functions by Category

#### 1. Environment & Path Checks (0% coverage)
- `checkEnvironmentVariables()` - Reads environment vars
- `checkPathConfiguration()` - Complex PATH analysis
- `checkExecutablePath()` - Uses exec.LookPath

#### 2. File System Checks (0% coverage)
- `checkPermissions()` - Tests write access
- `checkConfigurationFile()` - Reads config files
- `checkLockFile()` - Validates lock files

#### 3. Package Manager Checks (0% coverage)
- `checkPackageManagerAvailability()` - Tests command existence
- `checkSystemRequirements()` - Verifies Homebrew/Git

#### 4. External Command Utilities (0% coverage)
- `getPythonUserBinDir()` - Runs python3 command
- `getGoBinDir()` - Runs go env commands

#### 5. Pure Logic Functions (76-100% coverage)
- `detectShell()` - Shell detection from string
- `generatePathExport()` - String formatting
- `generateShellCommands()` - Command generation
- `calculateOverallHealth()` - Status aggregation

## Testing Strategy: Pragmatic Go Approach

### Core Principle: Abstract Only What's Necessary

Instead of heavy mocking frameworks, use minimal interfaces for the three main external dependencies:

1. **Command execution**
2. **File system operations**
3. **Environment variables**

### Proposed Solution: Lightweight Abstraction Layer

#### Step 1: Create Minimal Interfaces

```go
// internal/diagnostics/interfaces.go (new file)
package diagnostics

// SystemInterface provides mockable system operations
type SystemInterface interface {
    // Command execution
    LookPath(file string) (string, error)
    CommandOutput(name string, args ...string) ([]byte, error)

    // File operations
    FileExists(path string) (bool, error)
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error

    // Environment
    Getenv(key string) string
}

// RealSystem implements SystemInterface with actual OS calls
type RealSystem struct{}

func (RealSystem) LookPath(file string) (string, error) {
    return exec.LookPath(file)
}

func (RealSystem) CommandOutput(name string, args ...string) ([]byte, error) {
    return exec.Command(name, args...).Output()
}

// ... implement other methods

// Global variable for easy testing
var system SystemInterface = RealSystem{}
```

#### Step 2: Refactor Functions to Use Interface

Example refactoring of `checkExecutablePath`:

```go
// Before (untestable)
func checkExecutablePath() HealthCheck {
    plonkPath, err := exec.LookPath("plonk")
    // ...
}

// After (testable)
func checkExecutablePath() HealthCheck {
    plonkPath, err := system.LookPath("plonk")
    // ...
}
```

#### Step 3: Simple Test Pattern

```go
func TestCheckExecutablePath(t *testing.T) {
    tests := []struct {
        name     string
        system   *MockSystem
        want     HealthCheck
    }{
        {
            name: "plonk in PATH",
            system: &MockSystem{
                lookPathReturns: map[string]string{
                    "plonk": "/usr/local/bin/plonk",
                },
            },
            want: HealthCheck{
                Status: "pass",
                Message: "plonk is in PATH",
            },
        },
        {
            name: "plonk not in PATH",
            system: &MockSystem{
                lookPathErrors: map[string]error{
                    "plonk": exec.ErrNotFound,
                },
            },
            want: HealthCheck{
                Status: "warn",
                Message: "plonk is not in PATH",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            oldSystem := system
            system = tt.system
            defer func() { system = oldSystem }()

            got := checkExecutablePath()
            assert.Equal(t, tt.want.Status, got.Status)
            assert.Equal(t, tt.want.Message, got.Message)
        })
    }
}
```

## Implementation Plan

### Phase 1: Infrastructure Setup (Day 1)

1. **Create interfaces.go** with SystemInterface
2. **Implement RealSystem** with actual OS calls
3. **Create MockSystem** for testing
4. **Add helper functions** for common test scenarios

### Phase 2: Refactor & Test Simple Checks (Days 2-3)

Start with functions that have simple external dependencies:

1. **checkExecutablePath()** - Only uses exec.LookPath
   - Mock: Return path or error
   - Test cases: Found, not found, permission error
   - Expected coverage gain: +2-3%

2. **checkEnvironmentVariables()** - Only reads env vars
   - Mock: Return predefined environment
   - Test cases: All vars set, some missing, empty PATH
   - Expected coverage gain: +3-4%

3. **getPythonUserBinDir()** & **getGoBinDir()** - Run simple commands
   - Mock: Return command output
   - Test cases: Success, command not found, empty output
   - Expected coverage gain: +4-5%

### Phase 3: Test File System Checks (Days 4-5)

These require more complex mocking:

4. **checkConfigurationFile()** - File existence and readability
   - Mock: FileExists and ReadFile results
   - Test cases: File exists, missing, unreadable
   - Expected coverage gain: +3-4%

5. **checkLockFile()** - Similar to config file
   - Mock: FileExists and ReadFile results
   - Test cases: Valid lock, invalid lock, missing
   - Expected coverage gain: +3-4%

6. **checkPermissions()** - Most complex, creates temp files
   - Mock: MkdirAll, WriteFile, Remove results
   - Test cases: Writable, read-only, missing dir
   - Expected coverage gain: +4-5%

### Phase 4: Test Complex Logic (Week 2)

7. **checkPathConfiguration()** - Complex PATH analysis
   - Mock: Multiple env vars and file checks
   - Test cases: Various PATH scenarios
   - Expected coverage gain: +8-10%

8. **checkPackageManagerAvailability()** - Checks multiple commands
   - Mock: Command existence for each manager
   - Test cases: All available, some missing, none available
   - Expected coverage gain: +5-7%

9. **RunHealthChecks()** - Orchestrates all checks
   - Use existing mocked functions
   - Test cases: All pass, mixed results, all fail
   - Expected coverage gain: +3-5%

## Mock Implementation Reference

### Simple MockSystem

```go
type MockSystem struct {
    // Command mocking
    lookPathReturns map[string]string
    lookPathErrors  map[string]error
    commandOutputs  map[string][]byte
    commandErrors   map[string]error

    // File mocking
    files          map[string][]byte
    fileErrors     map[string]error
    writeFileError error
    mkdirAllError  error

    // Environment mocking
    env map[string]string

    // Tracking
    calls []string
}

func (m *MockSystem) LookPath(file string) (string, error) {
    m.calls = append(m.calls, fmt.Sprintf("LookPath:%s", file))
    if err, ok := m.lookPathErrors[file]; ok {
        return "", err
    }
    if path, ok := m.lookPathReturns[file]; ok {
        return path, nil
    }
    return "", exec.ErrNotFound
}

func (m *MockSystem) Getenv(key string) string {
    m.calls = append(m.calls, fmt.Sprintf("Getenv:%s", key))
    return m.env[key]
}

// ... other methods
```

## Test Data Helpers

Create realistic test scenarios:

```go
// Common test environments
func NewHealthySystemMock() *MockSystem {
    return &MockSystem{
        lookPathReturns: map[string]string{
            "brew":   "/opt/homebrew/bin/brew",
            "git":    "/usr/bin/git",
            "plonk":  "/usr/local/bin/plonk",
            "python3": "/usr/bin/python3",
        },
        commandOutputs: map[string][]byte{
            "python3 -m site --user-base": []byte("/Users/test/.local"),
            "go env GOBIN":                []byte("/Users/test/go/bin"),
        },
        env: map[string]string{
            "HOME":   "/Users/test",
            "PATH":   "/usr/local/bin:/usr/bin:/bin",
            "SHELL":  "/bin/zsh",
        },
        files: map[string][]byte{
            "/Users/test/.config/plonk/plonk.yaml": []byte("version: 2"),
            "/Users/test/.config/plonk/plonk.lock": []byte("version: 2"),
        },
    }
}

func NewBrokenSystemMock() *MockSystem {
    return &MockSystem{
        lookPathErrors: map[string]error{
            "brew": exec.ErrNotFound,
            "git":  exec.ErrNotFound,
        },
        env: map[string]string{
            "PATH": "",
        },
        fileErrors: map[string]error{
            "/Users/test/.config/plonk/plonk.yaml": os.ErrNotExist,
        },
    }
}
```

## Expected Coverage Improvements

| Phase | Functions Tested | Current → Target | Gain |
|-------|-----------------|------------------|------|
| Current | Utilities only | 13.7% | - |
| Phase 1 | Infrastructure | 13.7% | 0% |
| Phase 2 | Simple checks (4) | 13.7% → 25% | +11.3% |
| Phase 3 | File checks (3) | 25% → 35% | +10% |
| Phase 4 | Complex checks (3) | 35% → 50-55% | +15-20% |

Total improvement: **36-42% coverage gain**

## Alternative Approach: Integration Tests

For functions that are inherently system-dependent, consider integration tests:

```go
// internal/diagnostics/health_integration_test.go
// +build integration

func TestHealthChecksIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Run actual health checks
    report := RunHealthChecks()

    // Verify structure, not specific results
    assert.NotEmpty(t, report.Checks)
    assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, report.Overall.Status)

    // Verify all expected checks ran
    checkNames := make(map[string]bool)
    for _, check := range report.Checks {
        checkNames[check.Name] = true
    }

    assert.True(t, checkNames["System Requirements"])
    assert.True(t, checkNames["Environment Variables"])
    // ... etc
}
```

## Best Practices

1. **Keep mocks simple** - Don't over-engineer the mock system
2. **Test behavior, not implementation** - Focus on the health check results
3. **Use table-driven tests** - Group related scenarios
4. **Separate unit and integration tests** - Use build tags
5. **Don't mock what you don't need** - Some functions might be better tested via integration

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Interface changes break compatibility | Keep interfaces minimal and stable |
| Mocks become too complex | Start simple, add complexity only as needed |
| Tests become brittle | Test outcomes, not exact messages |
| Refactoring takes too long | Do incrementally, function by function |

## Success Criteria

- [ ] Diagnostics package reaches 50%+ coverage
- [ ] All critical health checks have unit tests
- [ ] Mock system is simple and maintainable
- [ ] Tests run quickly (< 1 second)
- [ ] No changes to external API

## Next Steps

1. Create `internal/diagnostics/interfaces.go`
2. Implement `RealSystem` and `MockSystem`
3. Refactor `checkExecutablePath()` as proof of concept
4. Continue with priority order list

This approach provides a pragmatic balance between testability and simplicity, avoiding heavy mocking frameworks while still achieving good test coverage for critical health check logic.
