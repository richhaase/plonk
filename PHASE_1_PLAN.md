# Phase 1: Directory Structure Migration

## Objective
Transform the current 9-package structure into the resource-focused 5-package architecture by creating new directories and migrating files.

## Timeline
Day 1 (8 hours)

## Current State
```
internal/
├── commands/      (3,990 LOC)
├── config/        (640 LOC)
├── dotfiles/      (2,550 LOC)
├── lock/          (304 LOC)
├── managers/      (3,600 LOC)
├── orchestrator/  (1,054 LOC)
├── state/         (102 LOC)
└── ui/            (464 LOC)
```

## Target State
```
internal/
├── commands/      (unchanged for Phase 1)
├── config/        (unchanged for Phase 1)
├── lock/          (unchanged for Phase 1)
├── orchestrator/  (unchanged for Phase 1)
├── output/        (renamed from ui)
└── resources/
    ├── packages/  (moved from managers)
    └── dotfiles/  (moved from internal/dotfiles)
```

## Task Breakdown

### Task 1.1: Create Resources Structure (30 min)
**Agent Instructions:**
1. Create directory `internal/resources`
2. Create subdirectories `internal/resources/packages` and `internal/resources/dotfiles`
3. Verify structure with `tree internal/`
4. Commit with message: "refactor: create resources directory structure"

**Validation:**
- Directories exist at correct paths
- Git shows new directories staged

### Task 1.2: Rename UI to Output (45 min)
**Agent Instructions:**
1. Use `git mv internal/ui internal/output` to rename directory
2. Update all imports from `"github.com/rdhelms/plonk/internal/ui"` to `"github.com/rdhelms/plonk/internal/output"`
3. Run `go build ./...` to verify no import errors
4. Run tests: `go test ./internal/output/...`
5. Commit with message: "refactor: rename ui package to output"

**Files to update imports in:**
- Search with: `grep -r "internal/ui" --include="*.go" .`
- Update all occurrences

**Validation:**
- No import errors
- All tests pass
- No references to "internal/ui" remain

### Task 1.3: Move Managers to Resources/Packages (2 hours)
**Agent Instructions:**
1. Use `git mv internal/managers/* internal/resources/packages/`
2. Update all imports from `"github.com/rdhelms/plonk/internal/managers"` to `"github.com/rdhelms/plonk/internal/resources/packages"`
3. Update package declarations in moved files from `package managers` to `package packages`
4. Fix any internal imports within the packages directory
5. Run `go build ./...` to verify no errors
6. Run tests: `go test ./internal/resources/packages/...`
7. Commit with message: "refactor: move managers to resources/packages"

**Special attention:**
- The `parsers` subdirectory should move to `internal/resources/packages/parsers`
- The `testing` subdirectory should move to `internal/resources/packages/testing`
- Update imports in all command files that use managers

**Validation:**
- No build errors
- All package tests pass
- No circular dependencies

### Task 1.4: Move Dotfiles to Resources/Dotfiles (1.5 hours)
**Agent Instructions:**
1. Use `git mv internal/dotfiles/* internal/resources/dotfiles/`
2. Update all imports from `"github.com/rdhelms/plonk/internal/dotfiles"` to `"github.com/rdhelms/plonk/internal/resources/dotfiles"`
3. Package declaration remains `package dotfiles`
4. Fix any internal imports
5. Run `go build ./...` to verify no errors
6. Run tests: `go test ./internal/resources/dotfiles/...`
7. Commit with message: "refactor: move dotfiles to resources/dotfiles"

**Validation:**
- No build errors
- All dotfile tests pass
- Commands that use dotfiles still work

### Task 1.5: Move State Types (1 hour)
**Agent Instructions:**
1. Copy types from `internal/state/types.go` to a new file `internal/resources/types.go`
2. Update the package declaration to `package resources`
3. Find all imports of `"github.com/rdhelms/plonk/internal/state"` and update to `"github.com/rdhelms/plonk/internal/resources"`
4. Delete the `internal/state` directory
5. Run `go build ./...` to verify no errors
6. Run all tests: `go test ./...`
7. Commit with message: "refactor: move state types to resources package"

**Note:** This is a copy-then-delete operation, not a move, to handle the package change

**Validation:**
- No build errors
- All tests pass
- state package no longer exists

### Task 1.6: Fix Import Cycles (1 hour)
**Agent Instructions:**
1. Run `go list -f '{{ join .Imports "\n" }}' ./... | sort | uniq -c | sort -rn` to check for issues
2. Look for any circular dependency errors from `go build ./...`
3. Common fixes:
   - If config imports resources, and resources imports config, move shared types to a common location
   - If orchestrator imports resources, ensure resources doesn't import orchestrator
4. Document any import cycle fixes
5. Commit fixes with message: "refactor: resolve import cycles from restructure"

**Validation:**
- `go build ./...` succeeds
- No circular dependency errors

### Task 1.7: Update Go Module and Verify (1 hour)
**Agent Instructions:**
1. Run `go mod tidy` to clean up module dependencies
2. Run `go build ./...` to ensure everything builds
3. Run all unit tests: `go test ./...`
4. Run integration tests: `just test-ux`
5. Check test execution time: `time go test ./...`
6. Commit with message: "refactor: update go.mod after restructure"

**Validation:**
- All tests pass
- Test execution time <5s (note the actual time)
- No missing dependencies

### Task 1.8: Final Verification (30 min)
**Agent Instructions:**
1. Run circular dependency check:
   ```bash
   go list -f '{{ join .Imports "\n" }}' ./... 2>&1 | grep -E "import cycle|circular"
   ```
2. Verify directory structure matches target:
   ```bash
   tree internal/ -d -L 2
   ```
3. Count LOC to ensure no code was lost:
   ```bash
   find internal/ -name "*.go" -not -path "*/test/*" | xargs wc -l
   ```
4. Create a summary report as comment in the PR

**Validation:**
- No circular dependencies detected
- Directory structure matches target
- Total LOC is approximately the same (±100 lines)

## Risk Mitigations

1. **Import Cycles**: Most likely between config ↔ resources
   - Solution: Move shared types to resources or create minimal interface

2. **Test Failures**: Paths in tests may need updating
   - Solution: Update test fixtures and paths

3. **Missing Imports**: Some files may be missed
   - Solution: Use comprehensive grep before moving

## Success Criteria
- [ ] All directories moved to new structure
- [ ] Zero import errors
- [ ] All tests passing
- [ ] Test execution <5s
- [ ] No circular dependencies
- [ ] Clean git history (8 atomic commits)

## Notes for Agents
- Use `git mv` for all moves to preserve history
- Commit after each successful task
- If you encounter errors, document them clearly
- Run tests after each major change
- Keep commits atomic and well-described
