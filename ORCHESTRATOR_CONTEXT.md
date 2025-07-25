# Orchestrator Context - Plonk Resource-Core Refactor

## Current State (2025-07-25)
- **Branch**: refactor/ai-lab-prep
- **Current Package Count**: 9 packages
- **Current LOC**: 13,536
- **Target**: 5 packages, ~8,000 LOC (41% reduction from current)
- **Goal**: Introduce Resource abstraction for AI Lab features while simplifying to idiomatic Go

## Key Documents
- **REFACTOR.md** - Master plan with 5-phase approach and progress tracking
- **docs/ARCHITECTURE.md** - DOES NOT EXIST YET - Will be created in Phase 5

## Previous Refactoring Summary
The codebase has already undergone significant refactoring:
- Reduced from 22 → 9 packages (59% reduction)
- Reduced from ~26,000 → 13,536 LOC (48% reduction)
- Eliminated anti-patterns: Java-style getters, complex errors, unnecessary abstractions
- Key deletions: cli, constants, executor, types, interfaces, services, operations, core, mocks, testing, errors packages

## Current Refactor Phase Plan

### Phase 1: Directory Structure (Day 1) - NOT STARTED
Create resource-focused architecture and migrate files

### Phase 2: Resource Abstraction (Day 2-3 + ½ day buffer) - NOT STARTED
Introduce Resource interface and adapt existing code

### Phase 3: Simplification & Edge-case Fixes (Day 4-5 + ½ day buffer) - NOT STARTED
Remove remaining abstractions, flatten implementations, and update code comments

### Phase 4: Lock v2 & Hooks (Day 6) - NOT STARTED
Implement extensible lock file and hook system

### Phase 5: Code Quality & Naming (Day 7) - NOT STARTED
Remove unused code and improve naming consistency

### Phase 6: Testing & Documentation (Day 8) - NOT STARTED
Update tests, docs, and ensure performance targets

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
