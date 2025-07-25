# Plonk Resource-Core Refactor

## Overview
Transform plonk from 9 packages to 5, introducing a Resource abstraction while aggressively simplifying within each package. This creates a lean core ready for AI Lab features (Docker Compose stacks, services like vLLM/Weaviate/Guardrails).

**Goals:**
- Reduce from ~13,500 LOC to ~8,000 LOC (41% reduction)
- Reduce from 9 packages to 5 packages
- Introduce minimal Resource interface for extensibility
- Preserve reconciliation semantics (Managed/Missing/Untracked)
- Maintain all current functionality with idiomatic Go
- Remove unused code and improve naming consistency

## Target Architecture

```
internal/
├── commands/      (~1,500 LOC) - Thin CLI handlers only
├── config/        (~300 LOC)   - Config loading and types
├── orchestrator/  (~300 LOC)   - Pure coordination
├── lock/          (~400 LOC)   - Lock v2 with resources section
├── output/        (≤300 LOC)   - Table/JSON/YAML formatting (stretch: 250)
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

### Phase 1: Directory Structure (Day 1) ✅ COMPLETE
- [x] Create new directory structure under `internal/`
- [x] Move files to new locations with git mv
- [x] Update all imports
- [x] Ensure tests pass with new structure
- [x] Verify no circular dependencies with `go list -f '{{ join .Imports "\n" }}' ./...`

### Phase 2: Resource Abstraction (Day 2-3 + ½ day buffer) ✅ COMPLETE
- [x] Define minimal Resource interface
- [x] Create shared reconciliation helper
- [x] Adapt package managers to implement Resource
- [x] Adapt dotfiles to implement Resource
- [x] Update orchestrator to use Resource interface
- [x] Add integration test: orchestrator Sync with 1 package + 1 dotfile → verify lock v2

**Checkpoint: Test runtime ~8.9s (exceeds 5s target) - proceeding to Phase 3**

### Phase 3: Simplification & Edge-case Fixes (Day 4-5 + ½ day buffer)
- [ ] Remove StandardManager abstraction
- [ ] Create `resources/packages/helpers.go` for 3-4 common helpers
- [ ] Flatten all manager implementations
- [ ] Simplify state types to single Item struct
- [ ] Remove error matcher patterns (verify with grep before deletion)
- [ ] Complete table output with tabwriter
- [ ] Review and update all code comments for accuracy
  - [ ] Remove outdated comments referencing deleted packages/patterns
  - [ ] Update comments to reflect new architecture
  - [ ] Ensure comments describe "why" not "what"
  - [ ] Remove TODO comments that are no longer relevant

### Phase 4: Lock v2 & Hooks (Day 6)
- [ ] Implement lock file v2 schema with resources section
- [ ] Add migration logic (v1 → v2, auto-upgrade on write)
- [ ] Add single lock version constant to prevent drift
- [ ] Implement hook execution in orchestrator (10min default timeout)
- [ ] Update plonk.yaml schema for hooks
- [ ] Log version migration during apply operations

### Phase 5: Code Quality & Naming (Day 7)
- [ ] Find and remove unused code
  - [ ] Run `staticcheck -unused ./...` to find unused functions/types
  - [ ] Use `go mod why` to check for unnecessary dependencies
  - [ ] Remove dead code paths and unreachable functions
- [ ] Improve naming consistency
  - [ ] Rename variables/functions that don't follow Go conventions
  - [ ] Fix inconsistent naming patterns (e.g., GetX vs X)
  - [ ] Ensure package names match their purpose
- [ ] Identify and refactor confusing names
  - [ ] Replace generic names (e.g., "data", "info", "item") with specific ones
  - [ ] Clarify ambiguous function names
  - [ ] Standardize terminology across packages
- [ ] Run `golint` and `go vet` for additional issues

### Phase 6: Testing & Documentation (Day 8)
- [ ] Update all tests for new structure
- [ ] Ensure <5s test execution (hard CI gate on unit + fast integration)
- [ ] Update ARCHITECTURE.md with "How to add a new Resource" section
- [ ] Update README quick-start paths for new structure
- [ ] Add "future resource checklist"
- [ ] Final cleanup and optimization

## Key Design Decisions

### Resource Interface
```go
type Resource interface {
    ID() string
    Desired() []Item          // Set by orchestrator from config (ordering handled by orchestrator)
    Actual(ctx) []Item
    Apply(ctx, Item) error
}
```

**Note**: Orchestrator handles ordering when needed (e.g., for future Docker services with dependencies)

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

**Note**: "degraded" state reserved for future use; orchestrator ignores it until health checks exist

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

**Migration**: Reader accepts v1 & v2; writer always upgrades to newest schema

### Hook Configuration
```yaml
# In plonk.yaml
hooks:
  pre_sync:
    - command: "echo Starting sync..."
      timeout: 30s  # Optional, defaults to 10m
  post_sync:
    - command: "./scripts/notify.sh"
      continue_on_error: true  # Optional, defaults to false (fail-fast)
```

**Notes**:
- Default timeout: 10 minutes (configurable with Go durations: 30s, 5m, 1h)
- Default behavior: fail-fast unless `continue_on_error: true`
- No rollback mechanism in this refactor

## Progress Tracking

### Metrics
- **Starting LOC**: 13,536
- **Current LOC**: ~14,800 (after Phase 2, includes new abstractions)
- **Target LOC**: ~8,000 (±10%)
- **Starting Packages**: 9
- **Current Packages**: 8 (state package removed)
- **Target Packages**: 5

### Package Status
- [ ] `commands` - Needs business logic extraction
- [ ] `config` - Already simplified, needs minor cleanup
- [x] `dotfiles` - ✅ Moved to resources/dotfiles
- [ ] `lock` - Needs v2 schema implementation
- [x] `managers` - ✅ Moved to resources/packages, still needs flattening
- [ ] `orchestrator` - Needs reduction to ~300 LOC
- [x] `output` - ✅ Renamed from ui package
- [x] `state` - ✅ Deleted, types moved to resources
- [x] `ui` - ✅ Renamed to output

### Deletions Completed
- [x] `state/` package - ✅ Types moved to resources
- [x] `ui/` package - ✅ Renamed to output

### Deletions Remaining
- [ ] StandardManager abstraction
- [ ] ErrorMatcher patterns
- [ ] Complex state types
- [ ] Unnecessary interfaces

## Success Criteria
- [ ] All tests passing
- [ ] <5s test execution time (hard CI gate on unit + fast integration tests)
- [ ] No package under 400 LOC unless trivial by nature
- [ ] No interfaces with single implementation
- [ ] All methods under 50 lines
- [ ] Direct error handling (no translation layers)
- [ ] Clean Resource abstraction for future extensions
- [ ] Orchestrator stays ≤300 LOC including hook runner

## Branch Strategy
- Main branch: `main`
- Feature branch: `refactor/ai-lab-prep`
- Checkpoint merge after Phase 2
- Final merge after Phase 5
- Tag: `v0.8.0-core`

## Risk Register

| Risk | Mitigation |
|------|------------|
| Circular dependencies creep back during moves | Run `go list -f '{{ join .Imports "\n" }}' ./...` after each phase |
| Lock migration breaks existing users silently | Log detected→target version during apply operations |
| Human output columns misalign on narrow terminals | Acceptable; JSON/YAML is the fallback |
| ErrorMatcher removal breaks tests | Grep for ErrorMatcher usage and exact error strings before deletion |

## Testing Strategy

### Integration Test
- Single test exercising full orchestrator flow:
  - Load config with 1 brew package + 1 dotfile
  - Run Sync on temp directory
  - Verify lock v2 produced with correct content

### Table Output Test
- Capture stdout and assert non-zero length (avoid brittle column checks)

### Performance Testing
- Separate slow e2e tests from unit/fast integration tests
- CI gate blocks merge if unit+fast tests >5s

## Notes
- Preserve reconciliation semantics for AI Lab
- Keep orchestrator thin but essential
- Maintain backward compatibility for config files
- Focus on idiomatic Go patterns
- Document decisions in ARCHITECTURE.md
- Create `resources/packages/helpers.go` for truly common functions (3-4 max)
