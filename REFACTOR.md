# Plonk Resource-Core Refactor

## Overview
Transform plonk from 9 packages to 5, introducing a Resource abstraction while aggressively simplifying within each package. This creates a lean core ready for AI Lab features (Docker Compose stacks, services like vLLM/Weaviate/Guardrails).

**Goals:**
- Reduce from ~14,300 LOC to ~11,000-12,000 LOC (15-20% reduction)
- Reduce from 8 packages to 5 packages
- Preserve Resource interface for AI Lab extensibility
- Preserve reconciliation semantics (Managed/Missing/Untracked)
- Maintain ALL current functionality (all 6 package managers)
- Focus on idiomatic Go simplification, not aggressive abstraction
- Keep orchestrator and all extensibility points for AI Lab

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

### Phase 3: Simplification & Edge-case Fixes (Day 4-5 + ½ day buffer) ✅ COMPLETE
- [x] Remove StandardManager abstraction
- [x] Create `resources/packages/helpers.go` for 3-4 common helpers
- [x] Flatten all manager implementations
- [x] Simplify state types to single Item struct
- [x] Remove error matcher patterns (verify with grep before deletion)
- [x] Complete table output with tabwriter
- [x] Review and update all code comments for accuracy
  - [x] Remove outdated comments referencing deleted packages/patterns
  - [x] Update comments to reflect new architecture
  - [x] Ensure comments describe "why" not "what"
  - [x] Remove TODO comments that are no longer relevant

**Result**: Only ~500 LOC reduction achieved (not the expected 6,000)

### Phase 3.5: Comprehensive Code Analysis ✅ COMPLETE
- [x] Analyzed commands package for duplication
- [x] Compared package manager implementations
- [x] Investigated dotfiles over-engineering
- [x] Identified non-essential features
- [x] Found cross-package duplication
- [x] Proposed architecture alternatives

**Result**: Identified 6,500-8,000 LOC potential reduction, but many conflict with AI Lab goals

### Phase 4: Idiomatic Go Simplification (Day 6-7) ✅ COMPLETE
- [x] Simplify error handling without losing context
- [x] Consolidate package manager tests
- [x] Simplify dotfiles package, trust stdlib
- [x] Merge doctor into status command (177 lines)
- [x] Remove genuinely unused code (1,165 lines)
- [x] Inline trivial single-use helpers
- [x] Final cleanup and formatting

**Result**: 474 lines reduction per scc (13,826 LOC), focused on genuine simplification

### Phase 5: Lock v2 & Hooks (Day 8) ✅ COMPLETE*
- [x] Implement lock file v2 schema with resources section
- [x] Add migration logic (v1 → v2, auto-upgrade on write)
- [x] Add single lock version constant to prevent drift
- [x] Implement hook execution in orchestrator (10min default timeout)
- [x] Update plonk.yaml schema for hooks
- [x] Log version migration during apply operations

**Note**: Infrastructure complete but not integrated into sync command (deferred to Phase 6)

### Phase 6: Final Structural Cleanup (Day 9) ✅ COMPLETE
- [x] Integrate new Orchestrator into sync command
  - [x] Update sync.go to use new Orchestrator.Sync() method
  - [x] Enable hook execution (pre/post sync)
  - [x] Enable v2 lock file generation
  - [x] Remove legacy sync functions
- [x] Extract business logic from commands package
- [x] Simplify orchestrator to pure coordination
- [x] Remove unnecessary abstractions (eliminated compat layer)
- [x] Consolidate related code within packages
- [x] Full integration testing

**Result**: Clean architecture with proper separation of concerns, import cycles resolved

### Phase 7: Code Quality & Naming (Day 10) ✅ COMPLETE
- [x] Find and remove unused code
  - [x] Run `staticcheck -unused ./...` to find unused functions/types
  - [x] Use `go mod why` to check for unnecessary dependencies
  - [x] Remove dead code paths and unreachable functions
- [x] Improve naming consistency
  - [x] Rename variables/functions that don't follow Go conventions
  - [x] Fix inconsistent naming patterns (e.g., GetX vs X)
  - [x] Ensure package names match their purpose
- [x] Identify and refactor confusing names
  - [x] Replace generic names (e.g., "data", "info", "item") with specific ones
  - [x] Clarify ambiguous function names
  - [x] Standardize terminology across packages
- [x] Run `golint` and `go vet` for additional issues

**Result**: 1,550+ lines of dead code removed through two passes. Initial cleanup reduced from 55+ to 16 dead code items (70% improvement), followed by additional cleanup reducing to just 2 test helpers (87.5% total improvement). Major parser cleanup (parsers.go reduced 86%). Standardized function naming throughout codebase. All tests passing with clean linter results.

### Phase 8: Comprehensive UX Review (Day 11)
- [ ] Review all CLI commands and patterns with stakeholder
- [ ] Identify opportunities to simplify without sacrificing functionality
- [ ] Document UX improvement recommendations
- [ ] Prioritize changes based on user impact and implementation effort
- [ ] Create detailed plan for UX improvements

### Phase 9+: UX Implementation (Days 12-13)
- [ ] Implement approved UX improvements
- [ ] Update command structure as needed
- [ ] Ensure backward compatibility where appropriate
- [ ] Update integration tests for new UX
- [ ] Update documentation for new patterns

### Final Phase: Testing & Documentation (Day 14)
- [ ] Update all tests for new structure and UX
- [ ] Ensure reasonable test execution time
- [ ] Update ARCHITECTURE.md with "How to add a new Resource" section
- [ ] Update README with new command patterns
- [ ] Add comprehensive examples and use cases
- [ ] Final verification and optimization

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
- **After Phase 1**: 13,978 (restructuring)
- **After Phase 2**: ~14,800 (added Resource abstraction)
- **After Phase 3**: ~14,300 (only 500 LOC reduction)
- **After Phase 4**: 13,826 (per scc - idiomatic simplification)
- **After Phase 5**: 13,826 (infrastructure added, no reduction)
- **After Phase 6**: ~13,800 (structural cleanup, consolidation)
- **After Phase 7**: ~12,250 (1,550+ LOC reduction from dead code removal)
- **Target LOC**: ~11,000-12,000 (revised for idiomatic approach)
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

## Future Considerations (Post-Refactor)

### Data-Driven Package Managers
After completing the refactor, consider investigating a data-driven approach where package managers are configured via YAML rather than implemented in code. This could:
- Reduce 2,340 LOC to ~400 LOC
- Make adding new managers trivial
- Align with AI Lab's declarative philosophy
- BUT: May lose flexibility for manager-specific quirks

Evaluate once the codebase is simplified and we understand the true commonalities across managers.

### Note on Code Reduction
The revised Phase 4 targets a more realistic 2,000-3,000 LOC reduction through idiomatic Go simplification rather than aggressive abstraction. This approach:
- Preserves code clarity and maintainability
- Avoids introducing complexity for the sake of line count
- Keeps Go idioms and explicit behavior
- Results in a final target of ~11,000-12,000 LOC instead of 10,000-11,000
