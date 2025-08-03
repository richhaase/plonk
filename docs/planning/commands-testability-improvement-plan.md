# Commands Package Testability Improvement Plan

**Created**: 2025-08-04
**Status**: PROPOSED
**Goal**: Improve commands package coverage from 14.6% to 40%+ using simple function extraction

## Overview

This document outlines a pragmatic approach to improving test coverage in the commands package by extracting pure functions without changing the architecture or CLI behavior.

## Guiding Principles

1. **No architectural changes** - Keep cobra commands as-is
2. **Extract only pure logic** - Functions with clear inputs/outputs
3. **Keep it simple** - No patterns, abstractions, or frameworks
4. **Maintain Go idioms** - Simple functions, not objects
5. **Zero CLI risk** - No changes to command flow or output

## Current State

- **Package**: `internal/commands`
- **Current Coverage**: 14.6%
- **Lines of Code**: ~5,200
- **Testable Functions**: Only helpers and output utilities
- **Untestable**: Command orchestration (`run*` functions)

## Extraction Strategy

### What to Extract

1. **Validation Logic**
   - Argument validation
   - Flag compatibility checks
   - Package spec validation

2. **Parsing Functions**
   - Package specifications (manager:package)
   - Flag normalization
   - Path resolution

3. **Business Logic**
   - Filtering logic (missing, unmanaged, etc.)
   - Sorting and grouping
   - State calculations

4. **Formatting Logic**
   - Error message formatting
   - Output preparation
   - Summary calculations

### What NOT to Extract

1. **Cobra integration** - Leave flag parsing in commands
2. **Orchestrator calls** - Keep dependencies as-is
3. **I/O operations** - Don't abstract filesystem
4. **Simple flag reads** - Not worth testing

## Implementation Plan

### Phase 1: Install/Uninstall Commands (Highest Impact)

**File**: `install.go` / `uninstall.go`

**Extract these functions:**

```go
// parsePackageSpecs parses and validates package specifications
func parsePackageSpecs(args []string, defaultManager string) ([]PackageSpec, error) {
    // Move logic from runInstall lines 67-101
}

// validatePackageSpec validates a single package specification
func validatePackageSpec(spec string) (manager, packageName string, err error) {
    // Extract validation from current ParsePackageSpec
    // Add empty checks and format validation
}

// groupPackagesByManager groups package specs by their manager
func groupPackagesByManager(specs []PackageSpec) map[string][]PackageSpec {
    // Useful for batch operations
}
```

**Estimated Coverage Gain**: +8-10%

### Phase 2: Status Command (Complex Logic)

**File**: `status.go`

**Extract these functions:**

```go
// validateStatusFlags checks for incompatible flag combinations
func validateStatusFlags(showUnmanaged, showMissing bool) error {
    if showUnmanaged && showMissing {
        return fmt.Errorf("--unmanaged and --missing are mutually exclusive")
    }
    return nil
}

// normalizeDisplayFlags sets defaults when no flags specified
func normalizeDisplayFlags(showPackages, showDotfiles bool) (packages, dotfiles bool) {
    if !showPackages && !showDotfiles {
        return true, true
    }
    return showPackages, showDotfiles
}

// filterItemsByState filters items based on state flags
func filterItemsByState(items []resources.Item, showMissing, showUnmanaged bool) []resources.Item {
    // Extract filtering logic from output formatting
}

// calculateStatusSummary builds summary statistics
func calculateStatusSummary(items []resources.Item) (total, missing, unmanaged int) {
    // Count logic currently embedded in output
}
```

**Estimated Coverage Gain**: +10-12%

### Phase 3: Apply Command (Strategy Logic)

**File**: `apply.go`

**Extract these functions:**

```go
// validateApplyFlags ensures flag combinations are valid
func validateApplyFlags(packagesOnly, dotfilesOnly bool) error {
    // Currently handled by cobra, but we can test the logic
}

// determineApplyScope calculates what to apply based on flags
func determineApplyScope(packagesOnly, dotfilesOnly bool) (applyPackages, applyDotfiles bool) {
    // Logic for what operations to perform
}

// convertApplyResultToOutput transforms orchestrator results for display
func convertApplyResultToOutput(result *orchestrator.Result, scope string) *ApplyOutput {
    // Currently inline in runApply
}
```

**Estimated Coverage Gain**: +5-7%

### Phase 4: Common Patterns (Cross-Command)

**File**: `helpers.go` (extend) or new `validation.go`

**Extract these functions:**

```go
// validateOutputFormat checks if output format is valid
func validateOutputFormat(format string) error {
    // Common across all commands
}

// resolvePaths expands and validates file paths
func resolvePaths(paths []string, homeDir string) ([]string, error) {
    // Used in multiple commands
}

// checkFileExists safely checks file existence
func checkFileExists(path string) (exists bool, readable bool) {
    // Abstracted file checks
}
```

**Estimated Coverage Gain**: +3-5%

## Testing Approach

### Standard Table-Driven Tests

```go
func TestValidatePackageSpec(t *testing.T) {
    tests := []struct {
        name    string
        spec    string
        wantMgr string
        wantPkg string
        wantErr bool
    }{
        {"valid brew spec", "brew:wget", "brew", "wget", false},
        {"valid npm spec", "npm:prettier", "npm", "prettier", false},
        {"no manager", "wget", "", "wget", false},
        {"empty package", "brew:", "", "", true},
        {"empty manager", ":wget", "", "", true},
        {"empty spec", "", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mgr, pkg, err := validatePackageSpec(tt.spec)
            if (err != nil) != tt.wantErr {
                t.Errorf("validatePackageSpec() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if mgr != tt.wantMgr {
                t.Errorf("manager = %v, want %v", mgr, tt.wantMgr)
            }
            if pkg != tt.wantPkg {
                t.Errorf("package = %v, want %v", pkg, tt.wantPkg)
            }
        })
    }
}
```

## Success Metrics

1. **Coverage**: Increase from 14.6% to 40%+
2. **No Breaking Changes**: All existing tests pass
3. **Maintainability**: Extracted functions are reusable
4. **Clarity**: Business logic is more visible and documented

## Risk Mitigation

1. **Extract one function at a time** - Test immediately
2. **Keep original code** - Extract, don't move initially
3. **Run full test suite** - Ensure no regressions
4. **Manual CLI testing** - Verify behavior unchanged

## Timeline

- **Week 1**: Phase 1 (Install/Uninstall) + Phase 2 (Status)
- **Week 2**: Phase 3 (Apply) + Phase 4 (Common)
- **Total Effort**: ~3-4 developer days

## Future Considerations

Once these extractions are complete and tested:

1. Consider subprocess testing for CLI behavior
2. Add integration tests with mock binaries
3. Extract more complex orchestration logic if needed

## Conclusion

This plan provides a low-risk path to significantly improve test coverage through simple function extraction. It maintains Go idioms, requires no architectural changes, and can be implemented incrementally with immediate benefits.
