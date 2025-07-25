# Plonk Resource-Core Refactor

## Overview
Transform plonk from 9 packages to 5, introducing a Resource abstraction while aggressively simplifying within each package. This creates a lean core ready for AI Lab features (Docker Compose stacks, services like vLLM/Weaviate/Guardrails).

**Goals:**
- Reduce from ~13,500 LOC to ~8,000 LOC (41% reduction)
- Reduce from 9 packages to 5 packages
- Introduce minimal Resource interface for extensibility
- Preserve reconciliation semantics (Managed/Missing/Untracked)
- Maintain all current functionality with idiomatic Go

## Target Architecture

```
internal/
├── commands/      (~1,500 LOC) - Thin CLI handlers only
├── config/        (~300 LOC)   - Config loading and types
├── orchestrator/  (~300 LOC)   - Pure coordination
├── lock/          (~400 LOC)   - Lock v2 with resources section
├── output/        (~500 LOC)   - Table/JSON/YAML formatting
└── resources/     (~5,000 LOC)
    ├── resource.go           - Resource interface definition
    ├── reconcile.go          - Shared reconciliation logic
    ├── packages/             - All package managers
    │   ├── manager.go        - PackageManager interface
    │   ├── homebrew.go       - Direct implementation
    │   ├── npm.go           - Direct implementation
    │   └── ...              - Other managers
    └── dotfiles/            - Dotfile operations
        └── manager.go       - Direct implementation
```

## Phase Plan

### Phase 1: Directory Structure (Day 1)
- [ ] Create new directory structure under `internal/`
- [ ] Move files to new locations with git mv
- [ ] Update all imports
- [ ] Ensure tests pass with new structure

### Phase 2: Resource Abstraction (Day 2-3)
- [ ] Define minimal Resource interface
- [ ] Create shared reconciliation helper
- [ ] Adapt package managers to implement Resource
- [ ] Adapt dotfiles to implement Resource
- [ ] Update orchestrator to use Resource interface

**Checkpoint: Merge to main after Phase 2**

### Phase 3: Simplification (Day 4-5)
- [ ] Remove StandardManager abstraction
- [ ] Flatten all manager implementations
- [ ] Simplify state types to single Item struct
- [ ] Remove error matcher patterns
- [ ] Complete table output with tabwriter

### Phase 4: Lock v2 & Hooks (Day 6)
- [ ] Implement lock file v2 schema with resources section
- [ ] Add migration logic (v1 → v2)
- [ ] Implement hook execution in orchestrator
- [ ] Update plonk.yaml schema for hooks

### Phase 5: Testing & Documentation (Day 7)
- [ ] Update all tests for new structure
- [ ] Ensure <5s test execution
- [ ] Update ARCHITECTURE.md
- [ ] Add "future resource checklist"
- [ ] Final cleanup and optimization

## Key Design Decisions

### Resource Interface
```go
type Resource interface {
    ID() string
    Desired() []Item          // Set by orchestrator from config
    Actual(ctx) []Item
    Apply(ctx, Item) error
}
```

### Simplified State Type
```go
type Item struct {
    Name   string
    Type   string              // "package", "dotfile", "service"
    State  string              // "managed", "missing", "untracked", "degraded"
    Error  error
    Meta   map[string]string   // For future service health info
}
```

### Lock File v2
```yaml
version: 2
packages:           # Unchanged for compatibility
  homebrew:
    - name: jq
      installed_at: ...
resources:          # New generic section
  - type: docker-compose
    id: ai-lab-stack
    state: ...
```

### Hook Configuration
```yaml
# In plonk.yaml
hooks:
  pre_sync:
    - command: "echo Starting sync..."
  post_sync:
    - command: "./scripts/notify.sh"
      continue_on_error: true
```

## Progress Tracking

### Metrics
- **Starting LOC**: 13,536
- **Current LOC**: 13,536
- **Target LOC**: ~8,000 (±10%)
- **Starting Packages**: 9
- **Current Packages**: 9
- **Target Packages**: 5

### Package Status
- [ ] `commands` - Needs business logic extraction
- [ ] `config` - Already simplified, needs minor cleanup
- [ ] `dotfiles` - Move to resources/dotfiles
- [ ] `lock` - Needs v2 schema implementation
- [ ] `managers` - Move to resources/packages, needs flattening
- [ ] `orchestrator` - Needs reduction to ~300 LOC
- [ ] `output` - Create from current ui package
- [ ] `state` - Delete, move types to resources
- [ ] `ui` - Merge into output

### Deletions Planned
- `state/` package (move types to resources)
- `ui/` package (merge into output)
- StandardManager abstraction
- ErrorMatcher patterns
- Complex state types
- Unnecessary interfaces

## Success Criteria
- [ ] All tests passing
- [ ] <5s test execution time
- [ ] No package under 300 LOC (except config)
- [ ] No interfaces with single implementation
- [ ] All methods under 50 lines
- [ ] Direct error handling (no translation layers)
- [ ] Clean Resource abstraction for future extensions

## Branch Strategy
- Main branch: `main`
- Feature branch: `refactor/ai-lab-prep`
- Checkpoint merge after Phase 2
- Final merge after Phase 5
- Tag: `v0.8.0-core`

## Notes
- Preserve reconciliation semantics for AI Lab
- Keep orchestrator thin but essential
- Maintain backward compatibility for config files
- Focus on idiomatic Go patterns
- Document decisions in ARCHITECTURE.md
