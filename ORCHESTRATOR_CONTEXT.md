# Orchestrator Context - Plonk Resource-Core Refactor

## Current State (2025-07-26)
- **Branch**: refactor/ai-lab-prep
- **Current Package Count**: 8 packages (was 9, state package removed)
- **Current LOC**: ~12,250 (after Phases 1-7, 1,550+ LOC reduction in Phase 7)
- **Revised Target**: 5 packages, ~11,000-12,000 LOC (15-20% reduction)
- **Goal**: Simplify while preserving ALL functionality for AI Lab features

## Key Documents
- **REFACTOR.md** - Master plan with 5-phase approach and progress tracking
- **docs/ARCHITECTURE.md** - DOES NOT EXIST YET - Will be created in Phase 5

## Previous Refactoring Summary
The codebase has already undergone significant refactoring:
- Reduced from 22 → 9 packages (59% reduction)
- Reduced from ~26,000 → 13,536 LOC (48% reduction)
- Eliminated anti-patterns: Java-style getters, complex errors, unnecessary abstractions
- Key deletions: cli, constants, executor, types, interfaces, services, operations, core, mocks, testing, errors packages

## Key Decisions from Phase 3.5 Analysis
- **Keep all 6 package managers** - All are critical to plonk's value
- **Preserve orchestrator** - Essential for AI Lab Docker Compose coordination
- **Maintain Resource abstraction** - Needed for future resource types
- **Keep JSON/YAML output** - Required for automation
- **Focus on internal simplification** - Not feature removal

## Key Decisions from Phase 6 Implementation
- **Preserved reconciliation abstraction** - Generic pattern in resources package
- **Domain-specific reconcile in domain packages** - Avoids import cycles
- **ReconcileAll in orchestrator** - Coordination logic belongs there
- **Eliminated compatibility layer** - Direct function calls are clearer
- **Consolidated small files** - Reduced fragmentation, improved cohesion

## Current Refactor Phase Plan

### Phase 1: Directory Structure (Day 1) - ✅ COMPLETE
Successfully migrated to resource-focused architecture (6 commits, 0.474s test time)

### Phase 2: Resource Abstraction (Day 2-3 + ½ day buffer) - ✅ COMPLETE
Successfully introduced Resource interface, adapted packages/dotfiles, simplified orchestrator (-79 lines)

### Phase 3: Simplification & Edge-case Fixes (Day 4-5 + ½ day buffer) - ✅ COMPLETE
Removed abstractions but only achieved ~500 LOC reduction

### Phase 3.5: Code Analysis - ✅ COMPLETE
Identified 6,500-8,000 LOC reduction potential, but many conflict with AI Lab requirements

### Phase 4: Idiomatic Go Simplification (Day 6-7) - ✅ COMPLETE
Achieved 474 LOC reduction through genuine simplification (removed 1,165 lines of dead code)

### Phase 5: Lock v2 & Hooks (Day 8) - ✅ COMPLETE*
Infrastructure implemented but sync command integration deferred to Phase 6

### Phase 6: Final Structural Cleanup (Day 9) - ✅ COMPLETE
Orchestrator integrated, business logic extracted, abstractions removed, import cycles resolved

### Phase 7: Code Quality & Naming (Day 10) - ✅ COMPLETE
Removed 1,550+ lines of dead code (87.5% reduction to just 2 test helpers), standardized naming conventions, major parser cleanup

### Phase 8: Comprehensive UX Review (Day 11) - ✅ COMPLETE
Reviewed all CLI commands and created detailed UX improvement plan

### Phase 9-14: UX Implementation (Days 12-14) - NOT STARTED
Implement command consolidation, prefix syntax, config changes, search/info improvements, and sync→apply rename

### Phase 15: Output Standardization (Day 14) - NOT STARTED
Standardize output formatting across all commands

### Final Phase: Testing & Documentation (Day 15) - NOT STARTED
Update tests, docs, and final verification

## Target Architecture (5 packages)

```
internal/
├── commands/      (~1,500 LOC) - Thin CLI handlers only
├── config/        (~300 LOC)   - Config loading and types
├── orchestrator/  (~300 LOC)   - Pure coordination
├── lock/          (~400 LOC)   - Lock v2 with resources section
├── output/        (≤300 LOC)   - Table/JSON/YAML formatting
└── resources/     (~5,000 LOC)
    ├── resource.go           - Resource interface definition
    ├── reconcile.go          - Shared reconciliation logic
    ├── packages/             - All package managers
    └── dotfiles/            - Dotfile operations
```

### Package Migration Mapping
| Current Package | Target Location | Action |
|----------------|-----------------|---------|
| commands | commands | Keep, extract remaining logic |
| config | config | Keep, minor cleanup |
| dotfiles | resources/dotfiles | Move and adapt to Resource |
| lock | lock | Keep, add v2 schema |
| managers | resources/packages | Move and flatten |
| orchestrator | orchestrator | Keep, reduce to ~300 LOC |
| state | DELETE | Move types to resources |
| ui | output | Rename and simplify |

## Key Design Decisions

### Resource Interface
- Minimal 4-method interface (ID, Desired, Actual, Apply)
- Orchestrator handles ordering for future dependencies
- Resources don't parse config - orchestrator sets Desired()

### State Management
- Single Item type for all resources (packages, dotfiles, future services)
- Three core states: managed, missing, untracked (+ degraded reserved)
- Meta map for extensibility without schema changes

### Hook System
- Default 10min timeout, configurable per hook
- Fail-fast by default, continue_on_error optional
- Shell commands only, no rollback in this refactor

### Lock File v2
- Reader accepts v1 & v2, writer always upgrades
- Single version constant to prevent drift
- Resources section for future Docker/service types

## Critical Preservation Points (AI Lab Requirements)
1. **Orchestrator Pattern** - Thin coordination layer (~300 LOC)
2. **Reconciliation Logic** - Managed/Missing/Untracked semantics
3. **Resource Abstraction** - Clean interface for Docker Compose integration
4. **Output Formats** - Table/JSON/YAML for automation
5. **Lock File v2** - Extensible for any resource type

## Task Creation Guidelines

### For Phase Tasks
Create focused task files for agents with:
1. **Clear objective** - One sentence describing the outcome
2. **Current state** - What exists now (file locations, package structure)
3. **Target state** - What should exist after (new structure, moved files)
4. **Specific steps** - Numbered list of actions to take
5. **Validation** - How to verify success (tests pass, no circular deps)

### Testing Requirements
- Unit tests: `go test ./...` must pass
- Integration tests: `just test-ux` must pass
- Performance: Unit + fast integration tests must run in <5s
- Circular deps: Check with `go list -f '{{ join .Imports "\n" }}' ./...`

### Success Metrics
- All tests green
- No new circular dependencies
- Target LOC achieved (±10%)
- Clean git history (atomic commits)

## Phase 1 Readiness Checklist
- [x] REFACTOR.md defines complete plan
- [x] ORCHESTRATOR_CONTEXT.md provides architecture guidance
- [ ] docs/ARCHITECTURE.md exists (or will be created in Phase 5)
- [x] Target package structure defined
- [x] Migration mapping clear
- [x] Risk mitigations identified
