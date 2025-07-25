# Phase 6 Progress and Import Cycle Issue

## What Has Been Accomplished

### Task 6.1: Integrate New Orchestrator ✅ COMPLETED
- Added new orchestrator constructor with options pattern (WithConfig, WithConfigDir, WithHomeDir, WithDryRun)
- Updated sync command to use new orchestrator for full sync operations
- Maintained backward compatibility with legacy sync functions for scoped operations
- Added SyncResult type for structured output
- Preserved existing functionality while integrating hooks and v2 lock support

### Task 6.2: Extract Business Logic from Commands ✅ COMPLETED
- Moved status summary functions (ConvertResultsToSummary, CreateDomainSummary, ExtractManagedItems) from commands to resources package
- Added operation validation utilities (ValidateOperationResults, HasFailures) to resources package
- Updated status, add, install, and rm commands to use extracted business logic
- Simplified error checking in commands using generic ValidateOperationResults function
- Removed duplicate code and improved code organization

### Task 6.3: Simplify Orchestrator Package 🔄 IN PROGRESS

#### Completed Steps:
1. **Moved directory utility functions**:
   - Moved `GetHomeDir()` and `GetConfigDir()` from `orchestrator/paths.go` to `config` package
   - Updated all imports and function calls throughout the codebase
   - Removed `orchestrator/paths.go` as it no longer contained orchestrator-specific functionality

2. **Moved health check functionality**:
   - Created new `internal/diagnostics` package
   - Moved `orchestrator/health.go` to `diagnostics/health.go`
   - Updated package declaration and imports
   - Updated status command to use `diagnostics.RunHealthChecks()`

3. **Started moving reconciliation logic**:
   - Added generic resource reconciliation functions (`ReconcileResource`, `ReconcileResources`) to `resources/reconcile.go`
   - Attempted to move domain-specific reconciliation functions

## Current Issue: Import Cycle

### The Problem
When attempting to move domain-specific reconciliation functions (`ReconcileDotfiles`, `ReconcilePackages`, `ReconcileAll`) from orchestrator to the resources package, I encountered an import cycle:

```
package github.com/richhaase/plonk/cmd/plonk
    imports github.com/richhaase/plonk/internal/commands from main.go
    imports github.com/richhaase/plonk/internal/diagnostics from status.go
    imports github.com/richhaase/plonk/internal/lock from health.go
    imports github.com/richhaase/plonk/internal/resources/dotfiles from yaml_lock.go
    imports github.com/richhaase/plonk/internal/resources from manager.go
    imports github.com/richhaase/plonk/internal/resources/packages from domain_reconcile.go
    imports github.com/richhaase/plonk/internal/resources from operations.go: import cycle not allowed
```

### Root Cause
The domain reconciliation functions need to import both:
- `github.com/richhaase/plonk/internal/resources/dotfiles`
- `github.com/richhaase/plonk/internal/resources/packages`

But these packages already import the base `resources` package, creating a circular dependency when the domain reconciliation functions are placed in the `resources` package.

### Current Workaround
I moved the domain reconciliation functions to a new `internal/utilities` package to break the cycle:
- Created `internal/utilities/reconcile.go`
- Updated all imports to use `utilities.ReconcileAll`, `utilities.ReconcilePackages`, etc.
- This resolves the import cycle but may not be the ideal architectural solution

## Remaining Tasks

### Task 6.3 Remaining Work:
- Remove any remaining non-coordination code from orchestrator
- Ensure orchestrator only handles: configuration loading, resource initialization, operation coordination, hook running, and lock file updates

### Task 6.4: Remove Unnecessary Abstractions
- Find interfaces with single implementations
- Remove overly generic code
- Simplify to direct implementations where appropriate

### Task 6.5: Consolidate Related Code
- Look for split logic that belongs together
- Move misplaced code to appropriate packages
- Consolidate within packages

### Task 6.6: Final Integration Testing
- Full workflow testing
- Run all tests
- Check for issues
- Create summary

## Architectural Considerations

The import cycle issue highlights a potential architectural problem. The domain reconciliation functions are utility functions that:
1. Don't belong in the base `resources` package (due to circular imports)
2. Don't belong in orchestrator (they're not coordination logic)
3. Could potentially be moved to domain-specific packages or a dedicated utilities package

The current solution (utilities package) works but might need refinement for better organization.

## Files Modified
- `internal/config/config.go` - Added directory utilities
- `internal/diagnostics/health.go` - Moved from orchestrator
- `internal/resources/reconcile.go` - Added generic reconciliation functions
- `internal/resources/types.go` - Added business logic extraction utilities
- `internal/utilities/reconcile.go` - Domain reconciliation functions
- Multiple command files - Updated imports and function calls
- Removed: `internal/orchestrator/paths.go`, `internal/orchestrator/health.go`, `internal/orchestrator/reconcile.go`

## Status
Currently able to build successfully with the utilities package workaround. Need to decide whether to:
1. Keep utilities package as-is
2. Move domain reconciliation functions to their respective domain packages
3. Restructure the package hierarchy to avoid circular dependencies
