# Architectural Improvements Plan

This document tracks the architectural improvements identified in the code review.

## Critical Priority (Must Fix)

### 1. Fix Dry-Run Handling in Dotfiles ⚠️
**File**: `internal/resources/dotfiles/resource.go`
**Issue**: Line 118-121 hard-codes `DryRun: false`, ignoring top-level dry-run flag
**Impact**: Violates user expectations, could cause unintended file mutations
**Fix**: Thread dry-run flag through DotfileResource and respect it in all apply operations

### 2. Fix Orchestrator Success Semantics 🐛
**File**: `internal/orchestrator/coordinator.go`
**Issue**: Lines 78-93 treat "no changes needed" as failure
**Impact**: Idempotent operations (a core design goal) appear to fail
**Fix**: Set `Success = !result.HasErrors()`, add separate `Changed` field

### 3. Split PackageManager Interface 🏗️
**File**: `internal/resources/packages/interfaces.go`
**Issue**: Monolithic interface with 10 mandatory methods, violates ISP
**Impact**: Forces stub implementations, unclear feature support
**Fix**: Split into capability interfaces:
- `PackageManagerCore` (IsAvailable, Install, Uninstall, ListInstalled)
- `PackageSearcher` (Search)
- `PackageInfoProvider` (Info, InstalledVersion)
- `PackageUpgrader` (Upgrade)
- `PackageHealthChecker` (CheckHealth)
- `PackageSelfInstaller` (SelfInstall)

### 4. Create CommandExecutor Interface 🧪
**File**: `internal/resources/packages/executor.go`
**Issue**: Package-level functions can't be mocked
**Impact**: Testing requires complex setup, tight coupling
**Fix**:
- Define `CommandExecutor` interface
- Inject into all manager constructors
- Provide default and mock implementations

### 5. Refactor Orchestrator to Use Resource Abstraction 🔄
**File**: `internal/orchestrator/coordinator.go`
**Issue**: Lines 57-76 hardcode domain calls instead of using Resource interface
**Impact**: Adding new resource types requires orchestrator changes
**Fix**:
- Orchestrator should iterate over `[]Resource` from registry
- Remove `packagesOnly`/`dotfilesOnly` branching
- Enable/disable resources via registry configuration

## High Priority

### 6. Replace Function-in-Metadata with Typed Interface 🔐
**File**: `internal/resources/reconcile.go`
**Issue**: Lines 35-45, 125-135 store function values in metadata map
**Impact**: Not type-safe, runtime panics possible
**Fix**: Define `DriftComparator` interface, store typed comparator

### 7. Remove Unused Code 🧹
**Files**: Multiple
- `homebrew.go:493-501` - unused `contains()` function
- `coordinator.go:18` - unused `ctx` field
- `coordinator.go:36-38` - lock service construction not used
**Fix**: Remove or properly integrate

### 8. Consolidate Reconciliation Functions 📦
**File**: `internal/resources/reconcile.go`
**Issue**: `ReconcileItems` and `ReconcileItemsWithKey` are ~90% identical
**Fix**: Single generic function with key selector parameter

### 9. Add Deep Copy for Metadata Maps 🐛
**File**: `internal/resources/reconcile.go`
**Issue**: Shallow map assignment causes aliasing
**Impact**: Mutations can affect unexpected items
**Fix**: Implement deep copy utility for metadata merging

### 10. Fix Security Documentation Mismatch 📄
**Files**: `architecture.md`, `homebrew.go:353-355`
**Issue**: Doc says "no automatic script execution" but SelfInstall uses `curl | bash`
**Fix**: Either update docs to clarify opt-in nature or change implementation

## Implementation Order

1. **Phase 1: Correctness Bugs** (Items 1, 2, 9)
   - Fix dry-run handling
   - Fix success semantics
   - Fix metadata aliasing

2. **Phase 2: Testability** (Items 4, 7)
   - Create CommandExecutor interface
   - Remove unused code

3. **Phase 3: Architecture** (Items 3, 5, 6, 8)
   - Split interfaces
   - Refactor orchestrator
   - Fix metadata type safety
   - Consolidate reconciliation

4. **Phase 4: Documentation** (Item 10)
   - Update security docs

## Testing Requirements

Each fix must include:
- Unit tests demonstrating the bug/issue
- Unit tests verifying the fix
- Integration tests where applicable
- No reduction in existing test coverage

## Success Criteria

- [ ] All tests pass (`go test ./...`)
- [ ] Test coverage remains ≥60% overall
- [ ] Linter passes (`just lint`)
- [ ] Build succeeds (`just build`)
- [ ] Manual testing of core workflows (install, apply, status, clone)
- [ ] Documentation updated to reflect changes
