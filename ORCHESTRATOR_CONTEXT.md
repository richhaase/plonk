# Orchestrator Context - Plonk Refactoring Project

## Current State (2025-07-24)
- **Branch**: refactor/simplify
- **Package Count**: 22 → 13 (9 eliminated)
- **Goal**: Reduce to 5-6 packages while preserving extensibility for AI Lab features

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

## Next Priority Tasks
1. **Transform `runtime` → `orchestrator`** - Eliminate singleton, keep ~200-300 LOC coordination (IN PROGRESS - Task 008)
2. Delete `mocks` package - Replace with simple test doubles
3. Keep `paths` package - Contains important domain-specific logic and security validation

## Package Analysis Notes
- **paths**: More complex than expected, provides security validation and Plonk-specific logic
- **runtime**: Complex singleton pattern - transform to simple orchestrator functions
- **mocks**: Generated complexity - replace with simple test doubles

## Critical Preservation Points (AI Lab Requirements)
1. **Orchestrator Pattern** - Transform runtime, don't delete
2. **Reconciliation Logic** - Keep Managed/Missing/Untracked pattern
3. **YAML Output** - Keep both JSON and YAML
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
