# Test Coverage Improvement Plan: Orchestrator and Config Packages

**Date**: 2025-08-03
**Author**: Analysis by Claude
**Current Status**: Planning phase completed, ready for implementation

## Executive Summary

This document provides a comprehensive plan for improving test coverage in the `orchestrator` and `config` packages. The orchestrator package currently shows 0% coverage and requires targeted improvements to reach 40-50%. The config package also shows 0% coverage but likely has good actual coverage that isn't being reported correctly.

## Package Analysis Summary

### Orchestrator Package
- **Current Coverage**: 0% (confirmed absent from coverage.out)
- **Target Coverage**: 40-50%
- **Lines of Code**: 490 LOC
- **Complexity**: 112
- **Main Challenge**: Direct exec calls and tight coupling to domain packages

### Config Package
- **Current Coverage**: 0% (likely a reporting issue)
- **Expected Actual Coverage**: 70-80%
- **Lines of Code**: 855 LOC
- **Complexity**: 174
- **Main Challenge**: Coverage not being captured, not testability

## Orchestrator Package Improvement Plan

### Current Architecture

The orchestrator package coordinates between:
```
Commands Layer (apply, status, diff)
    ↓
Orchestrator (coordination & hooks)
    ↓
Domain Layer (packages, dotfiles)
    ↓
Infrastructure (lock service, config)
```

### Key Components

1. **Orchestrator struct** - Main coordinator with options pattern
2. **HookRunner** - Executes shell commands with timeouts
3. **Apply functions** - Legacy compatibility for package/dotfile operations
4. **ReconcileAll** - Coordinates domain reconciliation

### Testing Challenges

1. **Direct exec.CommandContext calls** in HookRunner
2. **Tight coupling** to packages.Reconcile() and dotfiles.Reconcile()
3. **Complex Apply method** coordinating multiple subsystems
4. **Direct output package calls** for UI updates

### Recommended Implementation Approach

#### Phase 1: Hook Testing Infrastructure (Days 1-2)

**1. Create Command Runner Interface**

```go
// internal/orchestrator/interfaces.go
package orchestrator

import "context"

// CommandRunner abstracts command execution for testing
type CommandRunner interface {
    RunCommand(ctx context.Context, shell, command string) ([]byte, error)
}

// RealCommandRunner implements CommandRunner using exec
type RealCommandRunner struct{}

func (r *RealCommandRunner) RunCommand(ctx context.Context, shell, command string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, shell, "-c", command)
    return cmd.CombinedOutput()
}

// Package-level variable for test injection
var defaultCommandRunner CommandRunner = &RealCommandRunner{}

// SetCommandRunner allows tests to override the command runner
func SetCommandRunner(runner CommandRunner) {
    defaultCommandRunner = runner
}
```

**2. Refactor HookRunner**

```go
// Modify HookRunner.Run to use the interface
func (h *HookRunner) Run(ctx context.Context, hooks []config.Hook, phase string) error {
    for _, hook := range hooks {
        // ... timeout setup ...

        // Use interface instead of direct exec
        output, err := defaultCommandRunner.RunCommand(ctx, "sh", hook.Command)

        // ... error handling ...
    }
}
```

**3. Create Comprehensive Hook Tests**

```go
// internal/orchestrator/hooks_test.go
func TestHookRunner_Run(t *testing.T) {
    // Save and restore original runner
    originalRunner := defaultCommandRunner
    defer func() { defaultCommandRunner = originalRunner }()

    // Create mock runner
    mockRunner := &MockCommandRunner{
        outputs: map[string][]byte{
            "echo test": []byte("test"),
            "exit 1":    []byte("error"),
        },
        errors: map[string]error{
            "exit 1": fmt.Errorf("exit status 1"),
        },
    }
    SetCommandRunner(mockRunner)

    // Test successful hook
    // Test failing hook with continue_on_error
    // Test timeout handling
    // Test multiple hooks
}
```

**Expected Coverage Gain**: 15-20% from HookRunner and related functions

#### Phase 2: Resource Orchestration Testing (Days 3-4)

**1. Create Minimal Interfaces for Dependencies**

```go
// internal/orchestrator/dependencies.go
package orchestrator

// PackageOrchestrator abstracts package operations
type PackageOrchestrator interface {
    Reconcile(ctx context.Context) (resources.Result, error)
    ApplyChanges(ctx context.Context, items []resources.Item, dryRun bool) (PackageApplyResult, error)
}

// DotfileOrchestrator abstracts dotfile operations
type DotfileOrchestrator interface {
    Reconcile(ctx context.Context) (resources.Result, error)
    ApplyChanges(ctx context.Context, items []resources.Item, dryRun bool) (DotfileApplyResult, error)
}

// Default implementations that delegate to real packages
type RealPackageOrchestrator struct {
    configDir string
}

func (r *RealPackageOrchestrator) Reconcile(ctx context.Context) (resources.Result, error) {
    return packages.Reconcile(ctx, r.configDir)
}

// Similar for dotfiles...
```

**2. Add Test Hooks to Orchestrator**

```go
// Package-level variables for test injection
var (
    packageOrchestrator PackageOrchestrator
    dotfileOrchestrator DotfileOrchestrator
)

// Initialize with real implementations
func init() {
    // These would be set based on configDir when needed
}

// Test helper functions
func SetPackageOrchestrator(po PackageOrchestrator) {
    packageOrchestrator = po
}
```

**3. Test the Apply Method**

```go
func TestOrchestrator_Apply(t *testing.T) {
    tests := []struct {
        name           string
        options        []Option
        mockPackages   *MockPackageOrchestrator
        mockDotfiles   *MockDotfileOrchestrator
        expectedResult ApplyResult
        wantErr        bool
    }{
        {
            name: "successful apply all",
            options: []Option{
                WithDryRun(false),
            },
            mockPackages: &MockPackageOrchestrator{
                reconcileResult: resources.Result{
                    Missing: []resources.Item{{Name: "vim"}},
                },
                applyResult: PackageApplyResult{
                    TotalInstalled: 1,
                },
            },
            // ... test cases
        },
    }

    // Run tests with mocks
}
```

**Expected Coverage Gain**: 20-25% from Apply and related orchestration logic

#### Phase 3: Pure Function Testing (Day 5)

**1. Test All Option Functions**

```go
func TestOptions(t *testing.T) {
    tests := []struct {
        name   string
        option Option
        check  func(*Orchestrator) bool
    }{
        {
            name:   "WithDryRun",
            option: WithDryRun(true),
            check:  func(o *Orchestrator) bool { return o.dryRun },
        },
        // Test all options...
    }
}
```

**2. Test Result Conversion Functions**

```go
func TestConvertApplyResult(t *testing.T) {
    // Test conversion logic
}
```

**3. Test ReconcileAll with Mocks**

```go
func TestReconcileAll(t *testing.T) {
    // Mock both domain reconcile functions
    // Test successful reconciliation
    // Test error handling
}
```

**Expected Coverage Gain**: 5-10% from utility functions

### Implementation Guidelines

1. **Use existing patterns** - Follow the CommandExecutor pattern from packages
2. **Minimal changes** - Add interfaces only where necessary
3. **Preserve behavior** - All tests should verify existing behavior
4. **Incremental approach** - Implement one phase at a time
5. **Run tests frequently** - Ensure no regressions

### Test File Structure

```
internal/orchestrator/
├── orchestrator_test.go      # Main orchestrator tests
├── hooks_test.go            # Existing + expanded hook tests
├── apply_test.go            # New tests for apply functions
├── test_helpers.go          # Mock implementations
└── interfaces.go            # New file for interfaces
```

## Config Package Investigation Plan

### Current Situation

The config package shows 0% coverage but has 649 lines of test code across 3 test files. This suggests a coverage reporting issue, not a testability problem.

### Investigation Steps

**1. Verify Local Coverage**

```bash
# Run tests with coverage
go test ./internal/config -cover -v

# Generate detailed coverage
go test ./internal/config -coverprofile=config.coverage
go tool cover -html=config.coverage
```

**2. Check CI Configuration**

Look for issues in:
- Test command flags in CI
- Coverage merge process
- Test file patterns

**3. Common Issues to Check**

- Build tags excluding tests
- Test files not matching `*_test.go` pattern
- Package name mismatches
- Integration test tags

### If Coverage Is Actually Low

The config package is already well-designed for testing:

1. **Uses temp directories** effectively
2. **No need for mocking** - file I/O is tested with real files
3. **Pure functions** where possible
4. **Clear separation** of concerns

Minor improvements if needed:

```go
// Add these helpers to make tests even cleaner
func TestConfig_LoadScenarios(t *testing.T) {
    tests := []struct {
        name      string
        setupFunc func(dir string) error
        validate  func(cfg *Config) error
    }{
        // Comprehensive test scenarios
    }
}
```

## Success Metrics

### Orchestrator Package
- [ ] Achieve 40-50% test coverage
- [ ] All public functions have tests
- [ ] Mock infrastructure in place
- [ ] No changes to external API
- [ ] All existing functionality preserved

### Config Package
- [ ] Identify why coverage shows 0%
- [ ] Fix coverage reporting
- [ ] Verify actual coverage is 70%+
- [ ] Document any minor improvements

## Risk Mitigation

1. **API Compatibility** - Use interfaces internally only
2. **Test Isolation** - Use defer to restore package variables
3. **Incremental Changes** - Test after each phase
4. **Documentation** - Comment all test helpers

## Next Steps

1. **Orchestrator Package**:
   - Start with Phase 1 (Hook testing)
   - Create interfaces.go
   - Refactor HookRunner incrementally
   - Add comprehensive tests

2. **Config Package**:
   - Run local coverage analysis
   - Investigate CI configuration
   - Fix coverage reporting
   - Document findings

## Timeline Estimate

- **Orchestrator improvements**: 5 days
- **Config investigation**: 1 day
- **Total effort**: 6 days

This plan provides a clear path to improving test coverage while maintaining the stability and functionality of the plonk codebase.
