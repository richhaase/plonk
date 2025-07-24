# Orchestrator Context - Plonk Refactoring Project

## Current State (2025-07-24)
- **Branch**: refactor/simplify
- **Package Count**: 22 → 9 (59% reduction achieved)
- **Goal**: Clean, domain-focused architecture with idiomatic Go patterns
- **Philosophy**: Not forcing arbitrary package count, but achieving maintainable domain separation

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
12. ✅ **Task 010**: Deleted `errors` package - Eliminated 766 LOC over-engineered error system
13. ✅ **Task 012**: Simplified `config` package - 68% reduction achieved (593 → 278 LOC)

## Current Task Queue
1. **Task 013** (Simplify State): READY - Eliminate provider pattern (60-70% reduction)

## Recent Achievements
- **Config simplification complete**: Dual system eliminated, getters removed, idiomatic patterns
- **Error handling simplified**: 766 LOC of complexity replaced with `fmt.Errorf()`
- **All tests passing**: Both unit and UX integration tests validate changes

## Package Architecture Vision
**Current Packages (9)**:
- `commands` - CLI handlers (needs business logic extraction)
- `config` - Configuration management (simplifying in Task 012)
- `dotfiles` - Dotfile operations (clear domain)
- `lock` - Lock file handling (clear domain)
- `managers` - Package managers (needs inheritance removal)
- `orchestrator` - Coordination logic (preserved for AI Lab)
- `paths` - Path resolution/validation (may merge with dotfiles)
- `state` - State types (simplifying in Task 013)
- `ui` - Output formatting (clear domain)

**Target Architecture (7-9 well-defined packages)**:
Not forcing an arbitrary count, but achieving clean domain separation with idiomatic Go patterns. Each package should have a clear purpose and minimal cross-dependencies.

## Architectural Philosophy
- **Domain Clarity**: Each package represents a clear business domain
- **Idiomatic Go**: Following community patterns, not enterprise/Java patterns
- **Developer/AI Friendly**: Easy to understand, navigate, and extend
- **Right-sized Packages**: Not too granular, not too monolithic

## Critical Preservation Points (AI Lab Requirements)
1. **Orchestrator Pattern** - Coordination layer for future features
2. **Reconciliation Logic** - Managed/Missing/Untracked pattern for extensibility
3. **Output Formats** - All 3 formats (table, json, yaml) for automation
4. **Clean Interfaces** - Well-defined package boundaries
5. **Lock File Design** - Extensible for future resource types

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
