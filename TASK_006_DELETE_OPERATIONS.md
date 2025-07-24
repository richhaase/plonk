# Task 006: Delete Operations Package

## Objective
Remove the `operations` package by distributing its functionality to more appropriate domain packages.

## Quick Context
- `operations` contains shared types for batch operations (install, add, remove)
- Mostly simple data structures with minimal logic
- Part of simplification effort (reducing from 17 to 16 packages)

## Work Required

### 1. Move Core Types to `state` Package
Move from `operations/types.go` to `state/operations.go`:
- `OperationResult` struct
- `ResultSummary` struct
- `CalculateSummary()` function
- `CountByStatus()` function

Why: These types represent operation state and fit naturally with other state types.

### 2. Move Progress Reporting to `ui` Package
Move from `operations/progress.go` to `ui/progress.go`:
- `ProgressReporter` interface
- `DefaultProgressReporter` struct and all its methods
- Progress formatting functions

Why: This is UI logic that already imports from the ui package.

### 3. Distribute Context Utilities
Move from `operations/context.go`:
- `CreateOperationContext()` → Inline in commands that use it (simple one-liner)
- `CheckCancellation()` → Inline or move to individual commands
- `DetermineExitCode()` → Move to `commands/helpers.go`

Why: These are simple utilities that don't need their own package.

### 4. Update Imports
Update all files that import from operations:
- `internal/commands/install.go`
- `internal/commands/uninstall.go`
- `internal/commands/add.go`
- `internal/commands/rm.go`
- `internal/commands/helpers.go`
- `internal/core/dotfiles.go`
- `internal/ui/formatters.go`

### 5. Delete Package
- Remove `internal/operations/` directory completely

## Implementation Notes
- When moving types, check if you can simplify them
- The progress reporter might benefit from some cleanup when moved to ui
- Consider if any functions can be inlined rather than moved

## Completion Report

Create `TASK_006_COMPLETION_REPORT.md` with:
- List of what was moved where
- Any simplifications made
- Import updates count
- Test results
- Package count confirmation: 17 → 16

## Success Criteria
- No remaining operations package
- All functionality preserved in new locations
- All tests pass
- Cleaner package organization
