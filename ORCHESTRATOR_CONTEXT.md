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

## Next Priorities Based on Metrics
1. **Task 013**: State package simplification (689 → ~200-300 LOC)
2. **Managers refactoring**: Remove BaseManager inheritance pattern
3. **Commands extraction**: Move business logic to domain packages
4. **Path/Dotfiles merge**: Consider consolidating related functionality

## Package Architecture Vision

### Current Metrics (9 packages, 15,166 LOC total)
| Package | LOC | % of Total | Assessment |
|---------|-----|------------|------------|
| commands | 5,087 | 33.5% | Too large - business logic extraction needed |
| managers | 4,513 | 29.8% | BaseManager inheritance needs removal |
| dotfiles | 2,142 | 14.1% | Core domain, appropriate size |
| paths | 1,067 | 7.0% | Consider merging with dotfiles |
| state | 689 | 4.5% | Ready for provider pattern removal |
| config | 579 | 3.8% | Recently simplified, good size |
| ui | 464 | 3.1% | Well-focused |
| lock | 328 | 2.2% | Focused domain |
| orchestrator | 297 | 2.0% | Minimal coordination layer |

### Key Achievements
- **42% overall LOC reduction** from ~26,000 to 15,166
- **59% package reduction** from 22 to 9 packages
- **Eliminated anti-patterns**: Java-style getters, complex errors, unnecessary abstractions
- **Preserved extensibility**: AI Lab coordination patterns intact

**Target Architecture (7-9 well-defined packages)**:
Focus on domain clarity and idiomatic Go patterns rather than arbitrary package count.

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
