# Orchestrator Context - Plonk Refactoring Project

## Current State (2025-07-24)
- **Branch**: refactor/simplify
- **Package Count**: 22 → 9 (59% reduction achieved)
- **LOC Count**: ~26,000 → 14,573 (44% overall reduction)
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
14. ✅ **Task 013**: Simplified `state` package - 87% reduction achieved (1,011 → 131 LOC)
15. ✅ **Task 014**: Removed BaseManager inheritance - Eliminated Java-style patterns (qualitative win)
16. ✅ **Task 015**: Improved integration test implementation - Enhanced error handling and Go idioms
17. ✅ **Task 016**: Analyzed managers package - Identified 1,180-1,450 LOC reduction potential (21-26%)

## Current Task Queue
1. **Task 017** (Commands Refactoring): READY - Extract business logic from commands (20-30% reduction target)

## Next Priorities Based on Metrics
1. **Commands refactoring**: Extract business logic from commands package (20-30% reduction, 5,076 LOC base)
2. **Managers optimization**: Implement identified duplication fixes (21-26% reduction, detailed roadmap available)
3. **Path/Dotfiles merge**: Consider consolidating related functionality (package count reduction)
4. **UI enhancements**: Improve output consistency across commands

## Package Architecture Vision

### Current Metrics (9 packages, 14,573 LOC total)
| Package | LOC | % of Total | Assessment |
|---------|-----|------------|------------|
| commands | 5,076 | 34.8% | Too large - business logic extraction needed |
| managers | 4,402 | 30.2% | BaseManager inheritance needs removal |
| dotfiles | 2,245 | 15.4% | Core domain, appropriate size |
| paths | 1,067 | 7.3% | Consider merging with dotfiles |
| config | 579 | 4.0% | Recently simplified, good size |
| ui | 464 | 3.2% | Well-focused |
| orchestrator | 334 | 2.3% | Minimal coordination layer |
| lock | 304 | 2.1% | Focused domain |
| state | 102 | 0.7% | Successfully simplified (87% reduction) |

### Key Achievements
- **44% overall LOC reduction** from ~26,000 to 14,573
- **59% package reduction** from 22 to 9 packages
- **Eliminated anti-patterns**: Java-style getters, complex errors, unnecessary abstractions
- **Preserved extensibility**: AI Lab coordination patterns intact
- **Major simplifications**: state (87% reduction), config (68% reduction), errors (100% deletion)

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
