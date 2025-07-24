# Orchestrator Context - Plonk Refactoring Project

## Current State (2025-07-24)
- **Branch**: refactor/simplify
- **Package Count**: 22 → 10 (54% reduction achieved)
- **Goal**: Reduce to 5-6 packages while preserving extensibility for AI Lab features
- **Progress**: Excellent - only 4-5 more package changes needed to reach target

## Key Documents
- **CLAUDE_CODE_REVIEW.md** - Master refactoring plan and progress tracker
- **docs/ARCHITECTURE.md** - Target architecture

## Completed Tasks
1. ✅ Deleted `cli` package (merged to commands)
2. ✅ Deleted `constants` package (inlined)
3. ✅ Deleted `executor` package (use exec.Command directly)
4. ✅ Deleted `types` package (moved to state)
5. ✅ Fixed manager tests (converted to unit tests)
6. ✅ Deleted `interfaces` package (moved to consumers)
7. ✅ Deleted `services` package (moved to sync command)
8. ✅ Deleted `operations` package (moved to state/ui packages)
9. ✅ Deleted `core` package (moved to commands/managers packages)
10. ✅ Transformed `runtime` → `orchestrator` - Eliminated singleton, preserved coordination logic
11. ✅ Deleted `mocks` and `testing` packages - Eliminated 514 LOC of unused generated code

## Current Task Queue
1. **Task 010** (Delete Errors): IN PROGRESS by worker - 766 LOC elimination
2. **Task 012** (Simplify Config): READY - 65-70% reduction (593 → 150-200 LOC)

## Package Analysis Complete
- **errors**: 766 LOC over-engineered system - delete completely (IN PROGRESS)
- **config**: 593 LOC with migration debt - 65-70% reduction possible (PLANNED)
- **paths**: Keep - contains domain logic and security validation (DECISION FINAL)

## Critical Preservation Points (AI Lab Requirements)
1. **Orchestrator Pattern** - Transform runtime, don't delete
2. **Reconciliation Logic** - Keep Managed/Missing/Untracked pattern
3. **Output Formats** - Keep all 3 formats (table, json, yaml) for broad compatibility
4. **Clean Interfaces** - Maintain package boundaries for extensibility
5. **Lock File** - Design for future resource types (not just packages)

## Task Creation Pattern
```markdown
# Task XXX: [Action]

## Objective
[One sentence goal]

## Quick Context
[2-3 bullet points]

## Work Required
[Specific steps]

## Completion Report
Create `TASK_XXX_COMPLETION_REPORT.md` with:
- Summary of changes
- Files modified/deleted
- Test results
- Success criteria confirmation
```

## Testing
- Unit tests: `go test ./...`
- Integration tests: `just test-ux`
- Both must pass before marking task complete

## Worker Guidance
- Keep tasks focused and specific
- Provide just enough context
- Review completion reports before marking done
- Clean up task files after completion
