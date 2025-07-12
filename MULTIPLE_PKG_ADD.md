# Multiple Package Add Implementation Plan

## Overview

This document provides a detailed implementation plan for adding multiple package support to the `plonk pkg add` command. The enhancement allows users to add multiple packages in a single command while maintaining full backward compatibility.

## Goals

- Enable `plonk pkg add git neovim ripgrep htop`
- Support manager-specific flags: `plonk pkg add --manager npm typescript prettier eslint`
- Maintain backward compatibility with single package usage
- Provide clear progress indication and error handling
- Support dry-run mode for multiple packages

## Current Implementation Analysis

### Current Command Structure
```go
// internal/commands/pkg_add.go (current)
func runPkgAdd(cmd *cobra.Command, args []string) error {
    // Currently expects exactly 1 argument or 0 (for add all untracked)
    if len(args) == 0 {
        return addAllUntracked(cmd)
    }

    if len(args) != 1 {
        return errors.NewError(...)
    }

    packageName := args[0]
    // Process single package...
}
```

### Current Add Logic Flow
1. Validate single package argument
2. Load configuration
3. Get package manager
4. Check if package already managed
5. Install package
6. Update lock file
7. Show result

## Implementation Plan

**⚠️ IMPORTANT: See CONTEXT.md for pre-work requirements and shared utilities that must be implemented first.**

### Phase 0: Pre-work (✅ COMPLETED)

The following pre-work has been completed:

1. **✅ Add `GetInstalledVersion()` to PackageManager interface**
   - ✅ Updated `internal/managers/common.go` interface
   - ✅ Implemented in all package managers (homebrew, npm, cargo)
   - ✅ Updated mocks with `just generate-mocks`
   - ✅ All tests passing

2. **✅ Create shared utilities in `internal/operations/`**
   - ✅ Common result types and progress reporting interfaces (`types.go`)
   - ✅ Error suggestion formatting utilities (`progress.go`)
   - ✅ Summary display logic and context management (`context.go`)
   - ✅ Comprehensive test coverage (`types_test.go`)

3. **✅ Extend error system**
   - ✅ Added suggestion support to PlonkError type
   - ✅ Created helper methods: `WithSuggestion`, `WithSuggestionCommand`, `WithSuggestionMessage`
   - ✅ Updated `UserMessage()` to include suggestions

### Phase 1: Core Multiple Package Support

#### 1.1 Update Command Arguments Handling

**File:** `internal/commands/pkg_add.go`

```go
func runPkgAdd(cmd *cobra.Command, args []string) error {
    switch len(args) {
    case 0:
        // Add all untracked (existing behavior)
        return addAllUntracked(cmd)
    case 1:
        // Single package (existing behavior, but call new function)
        return addPackages(cmd, args)
    default:
        // Multiple packages (new behavior)
        return addPackages(cmd, args)
    }
}

func addPackages(cmd *cobra.Command, packageNames []string) error {
    // New function to handle both single and multiple packages
}
```

#### 1.2 Implement Multiple Package Processing

**New function structure:**
```go
// NOTE: This will be replaced by shared OperationResult type from internal/operations/
type PackageAddResult struct {
    Name           string
    Manager        string
    Version        string // Package version after successful installation
    Status         string // "added", "skipped", "failed"
    Error          error
    AlreadyManaged bool
}

func addPackages(cmd *cobra.Command, packageNames []string) error {
    results := make([]PackageAddResult, 0, len(packageNames))

    // Load configuration once
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    // Process each package
    for _, packageName := range packageNames {
        result := addSinglePackage(cfg, packageName, manager)
        results = append(results, result)

        // Show progress for each package
        showPackageProgress(result)
    }

    // Show summary
    showAddSummary(results)

    // Determine exit code
    return determineExitCode(results)
}
```

#### 1.3 Refactor Single Package Logic

**Extract existing logic into reusable function:**
```go
func addSinglePackage(cfg *config.Config, packageName string, managerName string) PackageAddResult {
    result := PackageAddResult{Name: packageName}

    // Determine manager
    manager, err := getPackageManager(cfg, managerName)
    if err != nil {
        result.Status = "failed"
        result.Error = err
        return result
    }
    result.Manager = manager.Name()

    // Check if already managed
    if isAlreadyManaged(cfg, packageName, manager) {
        result.Status = "skipped"
        result.AlreadyManaged = true
        return result
    }

    // Dry run check
    if dryRun {
        result.Status = "would-add"
        return result
    }

    // Install package
    err = manager.Install(ctx, packageName)
    if err != nil {
        result.Status = "failed"
        result.Error = err
        return result
    }

    // Get installed version for reporting (requires pre-work: GetInstalledVersion method)
    version, err := manager.GetInstalledVersion(ctx, packageName)
    if err != nil {
        // Version lookup failed, but installation succeeded
        result.Version = "unknown"
    } else {
        result.Version = version
    }

    // Update lock file immediately for clean state on cancellation
    err = updateLockFile(cfg, packageName, manager, result.Version)
    if err != nil {
        result.Status = "failed"
        result.Error = err
        return result
    }

    result.Status = "added"
    return result
}
```

### Phase 2: Output and User Experience

#### 2.1 Progress Indication

```go
func showPackageProgress(result PackageAddResult) {
    switch result.Status {
    case "added":
        fmt.Printf("✓ %s@%s (%s)\n", result.Name, result.Version, result.Manager)
    case "skipped":
        fmt.Printf("✗ %s (%s) - already managed\n", result.Name, result.Manager)
    case "failed":
        fmt.Printf("✗ %s (%s) - %s\n", result.Name, result.Manager, formatErrorWithSuggestion(result.Error))
    case "would-add":
        fmt.Printf("+ %s (%s) - would add\n", result.Name, result.Manager)
    }
}

// NOTE: This will be replaced by shared utility from internal/operations/
func formatErrorWithSuggestion(err error, packageName string) string {
    // This function will be replaced by operations.FormatErrorWithSuggestion
    // See CONTEXT.md for the shared implementation
    msg := err.Error()

    // Add suggestions based on error type
    if strings.Contains(msg, "not found") {
        return fmt.Sprintf("%s\n     Try: plonk search %s", msg, packageName)
    }
    if strings.Contains(msg, "manager unavailable") {
        return fmt.Sprintf("%s\n     Try: plonk doctor", msg)
    }
    if strings.Contains(msg, "network") || strings.Contains(msg, "timeout") {
        return fmt.Sprintf("%s\n     Check network connectivity", msg)
    }

    return msg
}
```

#### 2.2 Summary Output

```go
func showAddSummary(results []PackageAddResult) {
    added := countByStatus(results, "added")
    skipped := countByStatus(results, "skipped")
    failed := countByStatus(results, "failed")

    fmt.Printf("\nSummary: %d added, %d skipped, %d failed\n", added, skipped, failed)

    // Show failed packages with suggestions
    if failed > 0 {
        fmt.Println("\nFailed packages:")
        for _, result := range results {
            if result.Status == "failed" {
                fmt.Printf("  %s: %v\n", result.Name, result.Error)
            }
        }
        fmt.Println("\nTry running 'plonk doctor' to check system health")
    }
}
```

#### 2.3 Structured Output Support

```go
// Support for --output json/yaml
type MultipleAddOutput struct {
    Summary struct {
        Total   int `json:"total"`
        Added   int `json:"added"`
        Skipped int `json:"skipped"`
        Failed  int `json:"failed"`
    } `json:"summary"`
    Results []PackageAddResult `json:"results"`
}
```

### Phase 3: Error Handling and Edge Cases

#### 3.1 Error Handling Strategy

**Continue on Error Approach (Sequential Processing):**
- Process packages one at a time in order specified
- Report success or failure immediately after each package
- Continue processing remaining packages even if some fail
- Update lock file after each successful installation for clean state on cancellation
- Exit code 0 if any packages succeeded
- Exit code 1 only if all packages failed

**Error Scenarios:**
- **Cancellation (Ctrl-C)**: Clean termination with up-to-date lock file
- **Package manager unavailable**: Clear error with suggestion to run `plonk doctor`
- **Network failures**: Informative error with connectivity suggestions
- **Permission errors**: Clear error about file/directory permissions
- **Package not found**: Error with suggestion to try `plonk search <package>`
- **Lock file write failures**: Error about disk space or permissions

```go
func determineExitCode(results []PackageAddResult) error {
    failed := 0
    succeeded := 0

    for _, result := range results {
        if result.Status == "failed" {
            failed++
        } else if result.Status == "added" {
            succeeded++
        }
    }

    // Success if any packages were added
    if succeeded > 0 {
        return nil
    }

    // Failure only if all packages failed
    if failed > 0 {
        return errors.NewError(
            errors.ErrPackageInstall,
            errors.DomainPackages,
            "add-multiple",
            fmt.Sprintf("failed to add %d package(s)", failed),
        )
    }

    return nil
}
```

#### 3.2 Context and Timeout Handling

```go
func addPackages(cmd *cobra.Command, packageNames []string) error {
    // Create context with timeout for entire operation
    ctx, cancel := context.WithTimeout(context.Background(),
        time.Duration(cfg.OperationTimeout)*time.Second)
    defer cancel()

    for _, packageName := range packageNames {
        // Check if context cancelled
        if ctx.Err() != nil {
            return errors.Wrap(ctx.Err(), errors.ErrInternal,
                errors.DomainPackages, "add-multiple",
                "operation cancelled or timed out")
        }

        // Pass context to individual package operations
        result := addSinglePackageWithContext(ctx, cfg, packageName, manager)
        // ...
    }
}
```

#### 3.3 Lock File Management

**Strategy: Update after each successful package for clean state on cancellation**
```go
func addSinglePackage(cfg *config.Config, packageName string, managerName string) PackageAddResult {
    // ... install package ...

    // Update lock file immediately after successful install
    // This ensures plonk.lock is always in a clean, up-to-date state
    // even if user cancels with Ctrl-C during multi-package operation
    err = updateLockFile(cfg, packageName, manager, version)
    if err != nil {
        // Lock file update failure is a hard error
        result.Status = "failed"
        result.Error = errors.Wrap(err, errors.ErrFileIO,
            errors.DomainDotfiles, "update-lock",
            "failed to update lock file")
        return result
    }

    return result
}
```

### Phase 4: Testing Strategy

#### 4.1 Unit Tests

**File:** `internal/commands/pkg_add_test.go`

```go
func TestMultiplePackageAdd(t *testing.T) {
    tests := []struct {
        name          string
        packages      []string
        mockSetup     func(*mocks.MockPackageManager)
        expectedAdded int
        expectedFailed int
        expectError   bool
    }{
        {
            name:     "add multiple packages successfully",
            packages: []string{"git", "neovim", "ripgrep"},
            mockSetup: func(m *mocks.MockPackageManager) {
                m.EXPECT().Install(gomock.Any(), "git").Return(nil)
                m.EXPECT().Install(gomock.Any(), "neovim").Return(nil)
                m.EXPECT().Install(gomock.Any(), "ripgrep").Return(nil)
            },
            expectedAdded: 3,
            expectedFailed: 0,
            expectError: false,
        },
        {
            name:     "continue on partial failure",
            packages: []string{"git", "nonexistent", "ripgrep"},
            mockSetup: func(m *mocks.MockPackageManager) {
                m.EXPECT().Install(gomock.Any(), "git").Return(nil)
                m.EXPECT().Install(gomock.Any(), "nonexistent").Return(fmt.Errorf("package not found"))
                m.EXPECT().Install(gomock.Any(), "ripgrep").Return(nil)
            },
            expectedAdded: 2,
            expectedFailed: 1,
            expectError: false, // Success because some packages added
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 4.2 Integration Tests

```go
func TestBackwardCompatibility(t *testing.T) {
    // Test that single package add still works
    // Test that zero arguments (add all untracked) still works
    // Test that existing flags still work
}

func TestDryRunWithMultiplePackages(t *testing.T) {
    // Test dry run shows what would be added
    // Test no actual changes are made
}
```

#### 4.3 Error Scenarios Testing

```go
func TestErrorHandling(t *testing.T) {
    // Test context cancellation
    // Test package manager unavailable
    // Test network failures
    // Test lock file write failures
}
```

### Phase 5: Documentation Updates

#### 5.1 CLI Help Text

```go
var pkgAddCmd = &cobra.Command{
    Use:   "add [package1] [package2] ...",
    Short: "Add packages to plonk management",
    Long: `Add one or more packages to plonk management.

Examples:
  plonk pkg add git                    # Add single package
  plonk pkg add git neovim ripgrep     # Add multiple packages
  plonk pkg add --manager npm typescript prettier  # Specify manager
  plonk pkg add --dry-run git neovim   # Preview changes`,
    Args: cobra.MinimumNArgs(0),
}
```

#### 5.2 Update CLI.md Documentation

Add examples and update command reference to show multiple package support.

### Phase 6: Implementation Checklist

#### Code Changes
- [ ] Update `pkg_add.go` argument handling
- [ ] Implement `addPackages()` function
- [ ] Refactor single package logic into `addSinglePackage()`
- [ ] Add progress indication
- [ ] Add summary output
- [ ] Add structured output support
- [ ] Update error handling
- [ ] Add context/timeout support

#### Testing
- [ ] Add unit tests for multiple package scenarios
- [ ] Add backward compatibility tests
- [ ] Add error handling tests
- [ ] Add dry-run tests
- [ ] Test with different output formats
- [ ] Test context cancellation

#### Documentation
- [ ] Update command help text
- [ ] Update CLI.md with examples
- [ ] Add usage examples to README.md

#### Quality Assurance
- [ ] Run existing test suite (ensure no regressions)
- [ ] Run linter and security checks
- [ ] Test with real package managers
- [ ] Verify structured error handling
- [ ] Test edge cases (empty args, invalid packages, etc.)

## Migration and Compatibility

### Backward Compatibility Guarantees

1. **Existing single package usage unchanged**
   - `plonk pkg add git` works exactly as before
   - All existing flags continue to work
   - Output format for single packages unchanged

2. **Zero arguments behavior unchanged**
   - `plonk pkg add` still adds all untracked packages
   - No changes to this workflow

3. **Exit codes preserved**
   - Same exit codes for single package scenarios
   - Multiple package exit codes follow same patterns

### Performance Considerations

1. **Sequential processing** - Avoid package manager conflicts
2. **Early context checking** - Cancel gracefully on timeout
3. **Incremental lock file updates** - Don't lose progress on failure
4. **Memory efficient** - Process packages one at a time, don't load all into memory

## Future Enhancements

Once basic multiple package support is implemented, these enhancements could be considered:

1. **Multi-manager fallback** - If package not found in default manager, automatically try other managers or prompt user for choice. Example: `ripgrep` exists in both homebrew and cargo - which should be preferred?

2. **Manager auto-detection** - Automatically detect best manager for each package based on availability and user preferences

3. **Parallel processing** - For package managers that support it safely

4. **Dependency resolution** - Smart ordering of package installation

5. **Rollback support** - Undo partial installations on failure

6. **Progress bars** - For long-running installations

7. **Manager optimization** - Group packages by manager for efficiency

## Implementation Requirements Summary

Based on design review, the implementation should:

1. **Sequential processing** - Process packages one at a time to avoid conflicts
2. **Immediate lock file updates** - Update after each successful package for clean state on cancellation
3. **Immediate progress reporting** - Report success/failure after each package installation
4. **Informative error messages** - Include helpful suggestions for common error scenarios
5. **Version information** - Show package version in success messages (`✓ git@2.43.0 (homebrew)`)
6. **Continue on failure** - Process all packages even if some fail, show summary at end

## Additional Questions for Implementation

1. Should we add a `GetInstalledVersion()` method to the PackageManager interface, or use existing methods?
2. How should we handle cases where version lookup fails but installation succeeded?
3. Should the progress output be configurable (e.g., quiet mode for scripting)?
4. Any specific timeout values for the overall multi-package operation vs individual packages?
