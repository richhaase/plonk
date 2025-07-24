# TASK 007 - WORKER INSTRUCTIONS: Delete Core Package

## Orchestrator Notes
This task moves core business logic to appropriate domain packages. The core package is a thin abstraction that should be eliminated.

## Current State Analysis
**Core package contains 3 files:**
- `dotfiles.go` (341 lines) - Dotfile business logic → Move to `dotfiles/` package
- `packages.go` (33 lines) - Go package utilities → Move to `managers/` package
- `state.go` (37 lines) - Factory functions → Inline or move appropriately

**Consumers:** 5 command files import from core

## Detailed Worker Instructions

### STEP 1: Move Dotfile Operations (HIGH PRIORITY)
**Target:** Move all dotfile functions from `core/dotfiles.go` to `dotfiles/operations.go`

**Functions to move:**
- `AddSingleDotfile()` - Main entry point for adding dotfiles
- `AddSingleFile()` - Single file processing logic
- `AddDirectoryFiles()` - Directory expansion and processing
- `RemoveSingleDotfile()` - Remove dotfile operations
- `ProcessDotfileForApply()` - Apply operations for sync command

**Types to move:**
- `ProcessDotfileForApplyOptions` struct
- `ProcessDotfileForApplyResult` struct

**Action:** Append these functions to `/internal/dotfiles/operations.go` (don't overwrite existing content)

### STEP 2: Move Package Utilities (MEDIUM PRIORITY)
**Target:** Move `ExtractBinaryNameFromPath()` from `core/packages.go` to `managers/` package

**Action:** Add this function to `/internal/managers/go.go` (where it's used) or create `managers/utilities.go`

### STEP 3: Handle Factory Functions (MEDIUM PRIORITY)
**Target:** Eliminate or inline factory functions from `core/state.go`

**Functions to handle:**
- `LoadOrCreateConfig()` - **INLINE** directly in commands (it's just a 2-line wrapper)
- `CreatePackageProvider()` - **MOVE** to commands/helpers.go or inline
- `CreateDotfileProvider()` - **MOVE** to commands/helpers.go or inline

### STEP 4: Update All Imports (HIGH PRIORITY)
**Files to update:** (Found via grep)
- `/internal/commands/uninstall.go`
- `/internal/commands/rm.go`
- `/internal/commands/install.go`
- `/internal/commands/add.go`
- `/internal/commands/sync.go`

**Import changes:**
- Remove: `"github.com/richhaase/plonk/internal/core"`
- Add where needed:
  - `"github.com/richhaase/plonk/internal/dotfiles"` (for dotfile operations)
  - `"github.com/richhaase/plonk/internal/managers"` (for package utilities)

**Function call updates:**
- `core.AddSingleDotfile()` → `dotfiles.AddSingleDotfile()`
- `core.RemoveSingleDotfile()` → `dotfiles.RemoveSingleDotfile()`
- `core.ProcessDotfileForApply()` → `dotfiles.ProcessDotfileForApply()`
- `core.ExtractBinaryNameFromPath()` → `managers.ExtractBinaryNameFromPath()`
- `core.LoadOrCreateConfig()` → Inline the 2-line logic
- Factory functions → Use direct constructors or move to helpers

### STEP 5: Delete Core Package
**Action:** `rm -rf /Users/rdh/src/plonk/internal/core`

### STEP 6: Verify and Test
**Actions:**
1. Run `go test ./...` - All tests must pass
2. Run `just test` - Integration tests must pass
3. Verify no remaining imports to core package: `grep -r "internal/core" internal/`

## Success Criteria Checklist
- [ ] All dotfile functions moved to dotfiles package
- [ ] Package utilities moved to managers package
- [ ] Factory functions eliminated or properly relocated
- [ ] All 5 command files updated with correct imports
- [ ] Core package directory deleted
- [ ] All tests passing
- [ ] No remaining references to core package
- [ ] Package count reduced from 15 → 14

## Notes for Worker
- **Preserve all functionality** - Don't change logic, just move it
- **Test frequently** - Run tests after each major move
- **Handle imports carefully** - Make sure all dependencies follow to new locations
- **The core package contains legitimate business logic** - Unlike previous deletions, these functions do important work and must be preserved

## Completion Report Required
Create `TASK_007_COMPLETION_REPORT.md` documenting:
- What was moved where
- Any simplifications made during the move
- Import update count
- Test results
- Confirmation of package count reduction (15 → 14)

## Expected Outcome
The core package will be eliminated, with all business logic properly relocated to domain packages where it belongs. This removes an unnecessary abstraction layer while preserving all functionality.
