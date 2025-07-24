# STUCK - Import Cycle in Core Package Deletion - RESOLVED

## Task Status
Working on TASK 007 - Delete Core Package

## Current Situation
Encountered an import cycle while moving functions from `core` package to appropriate domain packages.

## What Was Completed
1. ✅ Moved all dotfile operations from `core/dotfiles.go` to `dotfiles/operations.go`
2. ✅ Moved `ExtractBinaryNameFromPath` from `core/packages.go` to `managers/goinstall.go`
3. ✅ Inlined `LoadOrCreateConfig` in `commands/add.go` (was just a 2-line wrapper)
4. ✅ Updated imports in all 5 command files to use new locations
5. ✅ Deleted the `core` package directory

## The Problem
After moving the functions and updating imports, running `go test ./...` reveals an import cycle:

```
package github.com/richhaase/plonk/cmd/plonk
	imports github.com/richhaase/plonk/internal/commands
	imports github.com/richhaase/plonk/internal/config
	imports github.com/richhaase/plonk/internal/managers
	imports github.com/richhaase/plonk/internal/state
	imports github.com/richhaase/plonk/internal/dotfiles
	imports github.com/richhaase/plonk/internal/state: import cycle not allowed
```

## Root Cause Analysis
The moved functions in `dotfiles/operations.go` use `state.OperationResult` type, creating a dependency:
- `dotfiles` → `state` (for OperationResult type)
- `state` → `dotfiles` (state/dotfile_provider.go imports dotfiles)

This creates a circular dependency that wasn't present when these functions were in the `core` package.

## ORCHESTRATOR RESOLUTION

### Analysis
The issue is architectural - the core package was actually **preventing** this import cycle by being a neutral location. When we moved the functions, we exposed a pre-existing tight coupling between state and dotfiles packages.

### **SOLUTION: Move Functions to Commands Package Instead**

The functions should NOT go to the `dotfiles` package because they:
1. Use `state.OperationResult` types (creating the import dependency)
2. Are primarily called by commands
3. Represent business logic orchestration, not core dotfile domain logic

### **Worker Instructions for Resolution:**

1. **MOVE functions from `dotfiles/operations.go` to `commands/dotfile_operations.go`:**
   - `AddSingleDotfile()`
   - `AddSingleFile()`
   - `AddDirectoryFiles()`
   - `RemoveSingleDotfile()`
   - `ProcessDotfileForApply()`
   - Both types: `ProcessDotfileForApplyOptions`, `ProcessDotfileForApplyResult`

2. **UPDATE imports in command files** (add.go, rm.go, sync.go):
   - Change `dotfiles.AddSingleDotfile()` → `AddSingleDotfile()` (same package)
   - Change `dotfiles.RemoveSingleDotfile()` → `RemoveSingleDotfile()` (same package)
   - Change `dotfiles.ProcessDotfileForApply()` → `ProcessDotfileForApply()` (same package)

3. **REMOVE the state import from dotfiles/operations.go**:
   - This will break the import cycle completely

### Why This Is The Correct Solution
- **Commands package is the natural orchestrator** - it coordinates between domains
- **No import cycles** - commands already imports both state and dotfiles
- **Follows Go best practices** - business logic orchestration belongs in the orchestrating package
- **Functions stay close to their usage** - all these functions are called primarily from commands

### Architecture Insight
The core package was actually serving as a **business logic orchestration layer**. When we deleted it, that orchestration responsibility should have gone to the `commands` package, not deeper into domain packages.

## Worker Action Plan
1. Create `internal/commands/dotfile_operations.go`
2. Move the 5 functions and 2 types from `dotfiles/operations.go`
3. Update 3 command files to use local functions instead of dotfiles package functions
4. Remove state import from dotfiles package
5. Test that import cycle is resolved
6. Complete Task 007

This maintains all functionality while resolving the architectural issue properly.
