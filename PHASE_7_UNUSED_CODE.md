# Phase 7: Unused Code Analysis

**Analysis Date**: July 25, 2025
**Tool Used**: `deadcode ./...`

## Summary

The deadcode analysis identified 55+ functions/methods that are genuinely unreachable. These fall into several categories:

## Safe to Remove

### Legacy/Unused Output Functions
- `PackageListOutput.TableOutput` and `StructuredData` methods - old output format
- `PackageStatusOutput.TableOutput` and `StructuredData` methods - old output format
- `DotfileListOutput.TableOutput` and `StructuredData` methods - old output format

### Legacy Configuration Functions (compat.go)
- `TargetToSource` - legacy function
- `LoadConfig` - wrapper function (we eliminated this in Phase 6 but missed cleanup)
- `LoadConfigWithDefaults` - wrapper function
- `ValidationResult.IsValid` - unused validation method
- `SimpleValidator.ValidateConfig` - unused validation method

### Legacy Orchestrator Methods
- `NewOrchestrator` - old constructor (we have New() now)
- `Orchestrator.GetResources` - unused method
- `Orchestrator.SyncLegacy` - legacy sync method
- `Orchestrator.writeLock` - internal method no longer used
- `SyncWithLockUpdate` - standalone function replaced by orchestrator methods

### Unused Resource Functions
- `ReconcileResources` - generic function not used
- `CountByStatus` - utility function not used
- `HasFailures` - utility function not used

### Unused Dotfile Utilities
- `Expander.ExpandDirectory` and related methods - alternative implementation not used
- `FileOperations.CopyDirectory`, `RemoveFile`, etc. - alternative file ops not used
- `Filter.ShouldSkipFilesOnly` - unused filter method
- `Scanner.ScanDirectory` and `walkDirectory` - alternative scanning not used

### Unused Package Manager Functions
- `ManagerRegistry.GetAvailableManagers` - registry method not used
- `ManagerRegistry.GetManagerInfo` - registry method not used
- `PackageResource.ID` and `MultiPackageResource.ID` - resource methods not used
- `GetActualPackages` and `GetActualPackagesForManager` - state functions not used

### Test Utilities Not Used
- Various helper functions in `test_helpers.go` and `testing/test_utils.go`
- Parser test utilities that are no longer referenced

### Unused Parser Functions
- Many parser functions in `parsers/parsers.go` that appear to be unused utilities
- `ParseJSON`, `ParseLines`, `SimpleLineParser` methods, etc.

## Keep - May Be Used

### Struct Fields with JSON/YAML Tags
All struct fields appear to be used for serialization, so keeping them.

### Exported Types and Interfaces
Interface methods are kept even if only one implementation exists.

## Analysis Confidence

**High Confidence** (safe to remove):
- Functions clearly marked as legacy or replaced
- Internal utility functions with no references
- Old output formatting methods

**Medium Confidence** (should investigate):
- Some parser functions might be used indirectly
- Some test utilities might be used by specific tests

## Recommendation

Remove the "Safe to Remove" items in phases:
1. Legacy functions (compat.go, old orchestrator methods)
2. Unused output formatters
3. Unused utilities (after verifying no indirect usage)

**Estimated LOC Reduction**: 200-400 lines
