# Task 007: Delete Core Package

## Objective
Remove the `core` package by merging its business logic directly into domain packages.

## Quick Context
- `core` is a thin abstraction layer over operations
- Contains only 3 files with business logic that belongs in domain packages
- Part of simplification effort (reducing from 15 to 14 packages)

## Work Required

### 1. Analyze Core Package Contents
Examine all files in `internal/core/` to understand:
- What business logic exists
- Which domain packages should receive each piece
- Current usage patterns across the codebase

### 2. Merge Business Logic into Domain Packages
Move from `core/` to appropriate domains:
- Dotfile operations → `dotfiles/` package
- Package operations → `managers/` package
- Config operations → `config/` package
- Any shared utilities → `commands/helpers.go`

Why: Business logic should live in its domain package, not in a generic "core" layer.

### 3. Update Imports
Update all files that import from core:
- `internal/commands/` files
- Any other consumers found during analysis

### 4. Delete Package
- Remove `internal/core/` directory completely

## Implementation Notes
- When moving functions, preserve all existing functionality
- Look for opportunities to inline simple wrapper functions
- Consider if any functions can be simplified when moved to their natural domain

## Completion Report

Create `TASK_007_COMPLETION_REPORT.md` with:
- List of what was moved where
- Any simplifications made
- Import updates count
- Test results
- Package count confirmation: 15 → 14

## Success Criteria
- No remaining core package
- All functionality preserved in domain packages
- All tests pass
- Cleaner domain organization
