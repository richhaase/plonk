# Capability Interface Refactoring Status

## Overview
This document tracks the progress of refactoring the monolithic PackageManager interface into smaller, focused capability interfaces (Issue #3 from ARCHITECTURAL_FIXES.md).

## Completed ✅

### Interface Design
- [x] Split PackageManager into core interface + 5 capability interfaces:
  - `PackageManager` (core: IsAvailable, Install, Uninstall, ListInstalled, IsInstalled, InstalledVersion, Dependencies)
  - `PackageSearcher` (optional: Search)
  - `PackageInfoProvider` (optional: Info)
  - `PackageUpgrader` (optional: Upgrade)
  - `PackageHealthChecker` (optional: CheckHealth)
  - `PackageSelfInstaller` (optional: SelfInstall)

### Helper Functions
- [x] Added capability detection functions:
  - `SupportsSearch(pm PackageManager) bool`
  - `SupportsInfo(pm PackageManager) bool`
  - `SupportsUpgrade(pm PackageManager) bool`
  - `SupportsHealthCheck(pm PackageManager) bool`
  - `SupportsSelfInstall(pm PackageManager) bool`

### Tests
- [x] Created `capabilities_test.go` with comprehensive tests:
  - `TestCapabilityDetection` - Verifies all 12 managers
  - `TestAllManagersSupportCore` - Ensures core interface compliance
  - `TestCapabilityFunctionsReturnBool` - Tests helper functions

### Production Code Updates
- [x] `internal/clone/setup.go` - SelfInstall capability check (line 351)
- [x] `internal/diagnostics/health.go` - CheckHealth capability check (line 380)
- [x] `internal/commands/search.go` - Search capability checks (lines 128, 202)

## In Progress 🚧

### Production Code (Needs Completion)
- [ ] `internal/commands/info.go` - 5 Info() calls need capability checks
  - Line 127, 175, 238, 259, and others
- [ ] `internal/commands/upgrade.go` - 1 Upgrade() call needs capability check
  - Line 417
- [ ] `internal/commands/install.go` - May have SelfInstall calls to check

### Test Files (Needs Update)
- [ ] `compliance_test.go` - Partially updated (line 82-105), needs full completion
  - 10+ test functions need capability checks
- [ ] Individual manager test files (20+ files):
  - `homebrew_test.go`, `npm_test.go`, `pnpm_test.go`, etc.
  - Each has Info(), Search(), Upgrade() calls that may need updates

## Not Started ❌

### Documentation
- [ ] Update architecture.md to reflect new capability model
- [ ] Update CONTRIBUTING.md with capability implementation guidance
- [ ] Add examples showing how to add new capabilities

### Optional Enhancements
- [ ] Add capability metadata to registry
- [ ] Create CLI command to list manager capabilities
- [ ] Add capability-based filtering in status/doctor commands

## Benefits of This Refactoring

1. **Clearer Contracts** - Managers explicitly declare what they support
2. **Easier Testing** - Can test capabilities independently
3. **Better Error Messages** - Can tell users when a feature isn't supported
4. **Simpler Implementation** - New managers don't need stub methods
5. **Future-Proof** - Easy to add new optional capabilities

## Breaking Changes

- Code calling Info(), Search(), Upgrade(), CheckHealth(), or SelfInstall() must now use type assertions
- Tests expecting all methods to exist need capability checks

## Migration Guide

### Before:
```go
result, err := manager.Search(ctx, "package")
```

### After:
```go
if searcher, ok := manager.(packages.PackageSearcher); ok {
    result, err := searcher.Search(ctx, "package")
} else {
    // Handle unsupported case
}
```

## Estimated Remaining Work

- **Info calls**: ~30 minutes (5 locations + tests)
- **Upgrade calls**: ~20 minutes (1 location + tests)
- **Compliance tests**: ~1 hour (comprehensive update)
- **Individual tests**: ~2 hours (20+ files, mostly mechanical)
- **Documentation**: ~30 minutes

**Total**: ~4 hours of focused work

## Next Steps

1. Complete remaining production code updates (info.go, upgrade.go)
2. Finish compliance_test.go updates
3. Run full test suite to identify remaining failures
4. Update failing tests systematically
5. Update documentation
6. Submit PR with all changes

## Notes

- All 12 package managers still implement all capabilities (they haven't changed)
- The refactoring is purely about making the interface more flexible
- No functional changes to manager behavior
- Backward compatible at runtime (managers still work the same)
