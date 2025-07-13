# Plonk Architecture Refactoring Implementation Plan

**Created**: 2025-07-13
**Objective**: Address architectural issues identified in CLAUDE_CODE_REVIEW.md through systematic refactoring

## ⚠️ Critical Constraint
**NON-NEGOTIABLE**: The UI/UX must remain completely unchanged. All CLI commands, outputs, behaviors, and user interactions must work exactly as they do today. This is the primary constraint for all refactoring work.

## Progress Tracking Grid

### Phase Overview
| Phase | Name | Priority | Status | Progress |
|-------|------|----------|--------|----------|
| 1 | Interface Consolidation | High | 🟡 In Progress | 67% |
| 2 | Abstraction Cleanup | High | 🔴 Not Started | 0% |
| 3 | Service Layer Extraction | Medium | 🔴 Not Started | 0% |
| 4 | Error Standardization | Medium | 🔴 Not Started | 0% |
| 5 | Legacy Cleanup | Low | 🔴 Not Started | 0% |

### Detailed Task Tracking

#### Phase 1: Interface Consolidation (High Priority)
| Task | Description | Status | Notes |
|------|-------------|--------|-------|
| P1.1 | Audit all interface duplications | ✅ DONE | See P1.1_INTERFACE_AUDIT.md |
| P1.2 | Create migration strategy | ✅ DONE | See P1.2_MIGRATION_STRATEGY.md |
| P1.3 | Consolidate config interfaces | ✅ DONE | See P1.3_CONSOLIDATION_PROGRESS.md |
| P1.4 | Standardize adapter layers | ✅ DONE | See P1.4_ADAPTER_STANDARDIZATION.md |
| P1.5 | Update all implementations | ⬜ TODO | |
| P1.6 | Test and validate | ⬜ TODO | |

#### Phase 2: Abstraction Cleanup (High Priority)
| Task | Description | Status | Notes |
|------|-------------|--------|-------|
| P2.1 | Remove RuntimeState | ⬜ TODO | |
| P2.2 | Enhance SharedContext | ⬜ TODO | |
| P2.3 | Update all commands | ⬜ TODO | |
| P2.4 | Remove TODOs | ⬜ TODO | |
| P2.5 | Documentation update | ⬜ TODO | |

#### Phase 3: Service Layer Extraction (Medium Priority)
| Task | Description | Status | Notes |
|------|-------------|--------|-------|
| P3.1 | Extract item detection logic | ⬜ TODO | |
| P3.2 | Extract apply operations | ⬜ TODO | |
| P3.3 | Extract sync operations | ⬜ TODO | |
| P3.4 | Break up shared.go | ⬜ TODO | |
| P3.5 | Create service interfaces | ⬜ TODO | |
| P3.6 | Update commands to use services | ⬜ TODO | |

#### Phase 4: Error Standardization (Medium Priority)
| Task | Description | Status | Notes |
|------|-------------|--------|-------|
| P4.1 | Audit error usage | ⬜ TODO | |
| P4.2 | Replace fmt.Errorf calls | ⬜ TODO | |
| P4.3 | Standardize error domains | ⬜ TODO | |
| P4.4 | Add error suggestions | ⬜ TODO | |
| P4.5 | Test error handling | ⬜ TODO | |

#### Phase 5: Legacy Cleanup (Low Priority)
| Task | Description | Status | Notes |
|------|-------------|--------|-------|
| P5.1 | Remove legacy output types | ⬜ TODO | |
| P5.2 | Standardize naming (Provider/Manager) | ⬜ TODO | |
| P5.3 | Clean up TODOs | ⬜ TODO | |
| P5.4 | Update documentation | ⬜ TODO | |

---

## Phase 1: Interface Consolidation (Week 1)

### Objective
Eliminate duplicate interface definitions and establish a single source of truth in `/internal/interfaces/`.

### Current State
- Duplicate interfaces in `/internal/config/interfaces.go` and `/internal/interfaces/config.go`
- Different method signatures for same interfaces
- Multiple adapter layers bridging duplicates
- `PackageConfigItem` struct duplicated

### Tasks

#### P1.1: Audit All Interface Duplications (Day 1)
**Goal**: Create comprehensive list of all duplicate interfaces and their usage

**Actions**:
1. Document all interfaces in both locations with their signatures
2. Identify all files importing from each location
3. Map adapter usage and dependencies
4. Create migration impact analysis

**Deliverables**:
- Duplication audit spreadsheet
- Dependency graph
- Risk assessment

#### P1.2: Create Migration Strategy (Day 1)
**Goal**: Define approach for consolidation without breaking existing code

**Strategy Options**:
1. **Big Bang**: Change all at once (higher risk, faster)
2. **Gradual Migration**: Use type aliases temporarily (safer, slower)
3. **Adapter Pattern**: Create temporary adapters (most complex)

**Recommended**: Gradual Migration with type aliases

**Actions**:
1. Define target interface signatures in `/internal/interfaces/`
2. Create migration order based on dependencies
3. Plan type alias strategy
4. Define testing approach

#### P1.3: Consolidate Config Interfaces (Days 2-3)
**Goal**: Merge config-related interfaces into unified definitions

**Actions**:
1. Update `/internal/interfaces/config.go` with final signatures
2. Add type aliases in `/internal/config/interfaces.go`:
   ```go
   // Temporary aliases for backward compatibility
   type ConfigReader = interfaces.ConfigReader
   type ConfigWriter = interfaces.ConfigWriter
   ```
3. Update implementations to use new interfaces
4. Remove method signature differences

**Key Decisions**:
- Use concrete types (`*Config`) vs generic (`interface{}`)
- Unified error handling approach
- Consistent naming conventions

#### P1.4: Standardize Adapter Layers (Day 4)
**Goal**: Establish consistent adapter patterns across the codebase

**Actions**:
1. Document adapter architecture:
   - Purpose: Prevent circular dependencies
   - Pattern: Bridge between package boundaries
   - Naming: *Adapter suffix convention
2. Create adapter guidelines:
   - When to use adapters (cross-package boundaries)
   - When to use type aliases (identical interfaces)
   - Standard implementation patterns
3. Standardize existing adapters:
   - Ensure consistent naming
   - Add interface compliance checks
   - Document adapter purposes
4. Performance optimization:
   - Measure adapter overhead
   - Consider caching if needed

#### P1.5: Update All Implementations (Day 5)
**Goal**: Ensure all code uses consolidated interfaces where possible

**Actions**:
1. Update imports to use consolidated interfaces from `/internal/interfaces/`
2. Replace direct interface usage where type aliases exist
3. Keep adapters for cross-package boundaries
4. Update mock generation if needed
5. Verify all interface satisfaction
6. Run full test suite to ensure no regressions

#### P1.6: Test and Validate (Day 6)
**Goal**: Comprehensive validation of interface consolidation

**Actions**:
1. Run all unit tests with coverage report
2. Run integration tests
3. Manual testing of key workflows:
   - Package management (add, remove, list)
   - Dotfile management (add, remove, sync)
   - Configuration operations
4. UI/UX snapshot comparison
5. Performance validation comparing before/after
6. Document any issues or regressions

---

## Phase 2: Abstraction Cleanup (Week 2)

### Objective
Remove RuntimeState pattern and standardize on SharedContext for resource management.

### Current State
- RuntimeState barely used with TODOs for future migration
- SharedContext widely adopted (10+ commands)
- Confusion about which pattern to use

### Tasks

#### P2.1: Remove RuntimeState (Days 1-2)
**Goal**: Eliminate RuntimeState pattern completely

**Actions**:
1. Remove `/internal/runtime/runtime_state.go`
2. Remove RuntimeState interface and implementation
3. Update any references (mainly TODOs)
4. Clean up related types

#### P2.2: Enhance SharedContext (Day 2)
**Goal**: Add any missing functionality from RuntimeState to SharedContext

**Actions**:
1. Review RuntimeState features
2. Identify any unique capabilities
3. Add useful features to SharedContext
4. Improve SharedContext documentation

**Enhancements**:
- Add state reconciliation helpers
- Add provider access methods
- Improve caching strategies

#### P2.3: Update All Commands (Days 3-4)
**Goal**: Ensure consistent SharedContext usage

**Commands to Update**:
- Commands not using SharedContext
- Commands with TODO comments about RuntimeState
- Ensure consistent initialization pattern

**Pattern**:
```go
sharedCtx := runtime.GetSharedContext()
cfg := sharedCtx.ConfigWithDefaults()
registry := sharedCtx.ManagerRegistry()
```

#### P2.4: Remove TODOs (Day 4)
**Goal**: Clean up migration TODOs

**Actions**:
1. Remove "TODO: Replace with RuntimeState" comments
2. Update any related documentation
3. Clean up test code

#### P2.5: Documentation Update (Day 5)
**Goal**: Update architecture docs

**Actions**:
1. Update ARCHITECTURE.md
2. Update CLAUDE.md guidelines
3. Create SharedContext usage guide

---

## Phase 3: Service Layer Extraction (Week 3-4)

### Objective
Extract business logic from commands layer into properly organized services.

### Current State
- 959-line shared.go with mixed responsibilities
- Services layer exists but only has 2 files
- Business logic embedded in presentation layer

### Tasks

#### P3.1: Extract Item Detection Logic (Days 1-2)
**Goal**: Create dedicated detection package

**New Package**: `/internal/detection/`

**Contents**:
```go
// detector.go
type ItemDetector interface {
    DetectItemType(item string) (ItemType, float64) // returns type and confidence
    IsPackage(item string) bool
    IsDotfile(item string) bool
}

// patterns.go
var packagePatterns = []Pattern{...}
var dotfilePatterns = []Pattern{...}
```

**Actions**:
1. Extract detection logic from shared.go
2. Create comprehensive pattern matching
3. Add confidence scoring
4. Create thorough tests

#### P3.2: Extract Apply Operations (Days 3-4)
**Goal**: Move apply logic to services

**New Service**: `/internal/services/apply_service.go`

**Interface**:
```go
type ApplyService interface {
    ApplyPackages(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
    ApplyDotfiles(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
    ApplyAll(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
}
```

**Actions**:
1. Extract from shared.go
2. Create clean service interface
3. Inject dependencies properly
4. Update command to use service

#### P3.3: Extract Sync Operations (Days 5-6)
**Goal**: Move sync logic to services

**Similar to Apply**:
- Create SyncService
- Extract sync-specific logic
- Clean interface design

#### P3.4: Break Up shared.go (Days 7-8)
**Goal**: Decompose shared.go into focused modules

**Target Structure**:
- `shared_types.go` - Shared type definitions
- `shared_utils.go` - Utility functions
- Individual files for remaining logic

**Actions**:
1. Categorize current contents
2. Create new focused files
3. Move code systematically
4. Update imports

#### P3.5: Create Service Interfaces (Days 9-10)
**Goal**: Define clean service layer APIs

**Services to Create**:
```go
// Package operations
type PackageService interface {
    Add(ctx context.Context, packages []string, options AddOptions) ([]Result, error)
    Remove(ctx context.Context, packages []string, options RemoveOptions) ([]Result, error)
    List(ctx context.Context, options ListOptions) ([]PackageInfo, error)
    Search(ctx context.Context, query string, options SearchOptions) ([]SearchResult, error)
}

// Dotfile operations
type DotfileService interface {
    Add(ctx context.Context, paths []string, options AddOptions) ([]Result, error)
    Remove(ctx context.Context, paths []string, options RemoveOptions) ([]Result, error)
    Link(ctx context.Context, options LinkOptions) ([]Result, error)
    Unlink(ctx context.Context, options UnlinkOptions) ([]Result, error)
}
```

#### P3.6: Update Commands to Use Services (Days 11-12)
**Goal**: Commands become thin presentation layer

**Pattern**:
```go
func runAddCommand(cmd *cobra.Command, args []string) error {
    // 1. Parse flags
    // 2. Get services from context
    // 3. Call service method
    // 4. Handle output formatting
    // 5. Return appropriate error
}
```

---

## Phase 4: Error Standardization (Week 5)

### Objective
Standardize error handling across the codebase using structured errors.

### Current State
- Mix of structured errors and fmt.Errorf
- 67 error calls in managers alone
- Inconsistent error domains and codes

### Tasks

#### P4.1: Audit Error Usage (Day 1)
**Goal**: Comprehensive error usage analysis

**Actions**:
1. Search for all fmt.Errorf usage
2. Categorize by package and type
3. Identify error patterns
4. Create conversion strategy

#### P4.2: Replace fmt.Errorf Calls (Days 2-3)
**Goal**: Convert to structured errors

**Priority Packages**:
1. `/internal/managers/` (highest usage)
2. `/internal/state/`
3. `/internal/commands/`

**Pattern**:
```go
// Before
return fmt.Errorf("failed to install package %s: %w", name, err)

// After
return errors.Wrap(err, errors.ErrPackageInstall, errors.DomainPackages, "install",
    "failed to install package").WithItem(name)
```

#### P4.3: Standardize Error Domains (Day 4)
**Goal**: Consistent error categorization

**Actions**:
1. Review current domains
2. Define complete domain list
3. Document domain usage
4. Update all error calls

#### P4.4: Add Error Suggestions (Day 4)
**Goal**: Helpful error messages

**Actions**:
1. Add suggestions to common errors
2. Include recovery actions
3. Add diagnostic information

#### P4.5: Test Error Handling (Day 5)
**Goal**: Validate error behavior

**Actions**:
1. Update tests for new errors
2. Test error messages
3. Verify error codes
4. Check user experience

---

## Phase 5: Legacy Cleanup (Week 6)

### Objective
Remove legacy code and establish consistent naming.

### Current State
- Legacy output types marked for compatibility
- Mix of Provider/Manager terminology
- Various TODOs and outdated comments

### Tasks

#### P5.1: Remove Legacy Output Types (Days 1-2)
**Goal**: Clean up output.go

**Actions**:
1. Identify legacy types
2. Find current usage
3. Migrate to new types
4. Remove legacy code

#### P5.2: Standardize Naming (Days 3-4)
**Goal**: Consistent terminology

**Decisions**:
- Provider vs Manager (recommend: Provider for state, Manager for packages)
- Result vs Output types
- Operation vs Action terminology

#### P5.3: Clean Up TODOs (Day 5)
**Goal**: Address or remove all TODOs

**Actions**:
1. Audit all TODO comments
2. Implement quick fixes
3. Create issues for larger items
4. Remove outdated TODOs

#### P5.4: Update Documentation (Day 5)
**Goal**: Reflect all changes

**Actions**:
1. Update ARCHITECTURE.md
2. Update CODEMAP.md
3. Update CLAUDE.md
4. Update inline documentation

---

## Success Metrics

### Code Quality
- [ ] Zero duplicate interfaces
- [ ] Single abstraction pattern (SharedContext)
- [ ] No fmt.Errorf usage
- [ ] shared.go under 200 lines
- [ ] All TODOs addressed

### Architecture
- [ ] Clear layer boundaries
- [ ] Thin command layer
- [ ] Rich service layer
- [ ] Consistent patterns

### Testing
- [ ] All tests pass
- [ ] No regression in functionality
- [ ] Improved test coverage
- [ ] Better error testing

### Documentation
- [ ] Updated architecture docs
- [ ] Clear development guidelines
- [ ] Accurate code maps
- [ ] Helpful error messages

---

## Risk Mitigation

### High Risk Areas
1. **Interface consolidation** - May break existing code
   - Mitigation: Use type aliases for gradual migration

2. **SharedContext changes** - Wide impact
   - Mitigation: Enhance without breaking existing API

3. **Service extraction** - Large refactoring
   - Mitigation: Incremental extraction with tests

### Testing Strategy
1. Maintain 100% test pass rate throughout
2. Add tests before refactoring
3. Manual testing of critical paths
4. Performance benchmarking

### Rollback Plan
1. Git branches for each phase
2. Feature flags for major changes
3. Ability to revert individual commits
4. Parallel implementation where possible

---

## Timeline Summary

**Total Duration**: 6 weeks

- **Week 1**: Interface Consolidation (High Priority)
- **Week 2**: Abstraction Cleanup (High Priority)
- **Week 3-4**: Service Layer Extraction (Medium Priority)
- **Week 5**: Error Standardization (Medium Priority)
- **Week 6**: Legacy Cleanup (Low Priority)

**Critical Path**: Phases 1 and 2 must complete before major service extraction

**Flexibility**: Phases 4 and 5 can be done in parallel or reordered based on needs
