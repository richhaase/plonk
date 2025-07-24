# Orchestrator Context - Plonk Refactoring Project

## Current State (2025-07-24)
- **Branch**: refactor/simplify
- **Package Count**: 22 → 16 (6 eliminated)
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

## Next Priority Tasks
1. Delete `operations` package - Move types to state, UI to ui package (READY - Task 006)
2. Transform `runtime` → `orchestrator` - Keep ~200-300 LOC coordination
3. Delete `core` package - Merge into domain packages
4. Keep `paths` package - Contains important domain-specific logic and security validation

## Package Analysis Notes
- **paths**: More complex than expected, provides security validation and Plonk-specific logic
- **operations**: Good candidate - simple types that fit naturally in state/ui packages
- **services**: Easiest target - thin wrapper used only by sync command

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
