# TASK 008 WORKER INSTRUCTIONS - Transform Runtime to Orchestrator

## Overview
You are transforming the complex `runtime` package into a clean `orchestrator` package. This preserves essential coordination functionality while eliminating over-engineering.

## Phase 1: Create Orchestrator Package Structure

### Step 1.1: Create the new package directory
```bash
mkdir -p internal/orchestrator
```

### Step 1.2: Move logging.go unchanged
```bash
cp internal/runtime/logging.go internal/orchestrator/logging.go
# Update package declaration from "package runtime" to "package orchestrator"
```

### Step 1.3: Create orchestrator/paths.go
Create simple path utilities (no caching, no singleton):
```go
package orchestrator

import (
    "os"
    "github.com/richhaase/plonk/internal/config"
)

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
    homeDir, _ := os.UserHomeDir()
    return homeDir
}

// GetConfigDir returns the plonk configuration directory
func GetConfigDir() string {
    return config.GetDefaultConfigDirectory()
}
```

### Step 1.4: Create orchestrator/reconcile.go
Extract the reconciliation logic from runtime without the SharedContext complexity:

```go
package orchestrator

import (
    "context"
    "github.com/richhaase/plonk/internal/config"
    "github.com/richhaase/plonk/internal/lock"
    "github.com/richhaase/plonk/internal/managers"
    "github.com/richhaase/plonk/internal/state"
)

// ReconcileDotfiles performs dotfile reconciliation without SharedContext
func ReconcileDotfiles(ctx context.Context, homeDir, configDir string) (state.Result, error) {
    // Create fresh provider (no caching)
    cfg := config.LoadConfigWithDefaults(configDir)
    configAdapter := config.NewConfigAdapter(cfg)
    dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
    provider := state.NewDotfileProvider(homeDir, configDir, dotfileConfigAdapter)

    // Get items
    configured, err := provider.GetConfiguredItems()
    if err != nil {
        return state.Result{}, err
    }

    actual, err := provider.GetActualItems(ctx)
    if err != nil {
        return state.Result{}, err
    }

    // Reconcile (copy exact logic from runtime/context_simple.go)
    return reconcileDotfileItems(provider, configured, actual), nil
}

// ReconcilePackages performs package reconciliation without SharedContext
func ReconcilePackages(ctx context.Context, configDir string) (state.Result, error) {
    // Create fresh providers (no caching)
    lockService := lock.NewYAMLLockService(configDir)
    lockAdapter := lock.NewLockFileAdapter(lockService)
    registry := managers.NewManagerRegistry()
    provider, err := registry.CreateMultiProvider(ctx, lockAdapter)
    if err != nil {
        return state.Result{}, err
    }

    // Get items
    configured, err := provider.GetConfiguredItems()
    if err != nil {
        return state.Result{}, err
    }

    actual, err := provider.GetActualItems(ctx)
    if err != nil {
        return state.Result{}, err
    }

    // Reconcile (copy exact logic from runtime/context_simple.go)
    return reconcilePackageItems(provider, configured, actual), nil
}

// ReconcileAll reconciles all domains
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]state.Result, error) {
    results := make(map[string]state.Result)

    // Reconcile dotfiles
    dotfileResult, err := ReconcileDotfiles(ctx, homeDir, configDir)
    if err != nil {
        return nil, err
    }
    results["dotfile"] = dotfileResult

    // Reconcile packages
    packageResult, err := ReconcilePackages(ctx, configDir)
    if err != nil {
        return nil, err
    }
    results["package"] = packageResult

    return results, nil
}

// Copy the reconcile helper functions exactly from runtime/context_simple.go:
// - reconcileDotfileItems()
// - reconcilePackageItems()
```

## Phase 2: Update All Command Files

You need to update 13 command files. The pattern is always:

### Before (using runtime.SharedContext):
```go
import "github.com/richhaase/plonk/internal/runtime"

sharedCtx := runtime.GetSharedContext()
homeDir := sharedCtx.HomeDir()
configDir := sharedCtx.ConfigDir()
result, err := sharedCtx.SimplifiedReconcileDotfiles(ctx)
```

### After (using orchestrator functions):
```go
import "github.com/richhaase/plonk/internal/orchestrator"

homeDir := orchestrator.GetHomeDir()
configDir := orchestrator.GetConfigDir()
result, err := orchestrator.ReconcileDotfiles(ctx, homeDir, configDir)
```

### Files to Update:
1. **internal/commands/add.go**: Replace `sharedCtx.HomeDir()`, `sharedCtx.ConfigDir()`
2. **internal/commands/rm.go**: Replace `sharedCtx.HomeDir()`, `sharedCtx.ConfigDir()`
3. **internal/commands/sync.go**: Replace reconciliation calls
4. **internal/commands/install.go**: Replace `sharedCtx.ConfigWithDefaults()`, `sharedCtx.ManagerRegistry()`
5. **internal/commands/uninstall.go**: Replace `sharedCtx.ManagerRegistry()`
6. **internal/commands/doctor.go**: Replace reconciliation calls
7. **internal/commands/status.go**: Replace reconciliation calls
8. **internal/commands/ls.go**: Replace reconciliation calls
9. **internal/commands/env.go**: Replace directory access
10. **internal/commands/info.go**: Replace directory access
11. **internal/commands/search.go**: Replace registry access
12. **internal/commands/helpers.go**: Replace utility calls
13. **internal/commands/shared.go**: Replace any runtime usage

### Special Cases to Handle:

#### For ManagerRegistry() calls:
```go
// Before:
registry := sharedCtx.ManagerRegistry()

// After:
registry := managers.NewManagerRegistry()  // Create fresh instance
```

#### For Config calls:
```go
// Before:
cfg := sharedCtx.ConfigWithDefaults()

// After:
cfg := config.LoadConfigWithDefaults(orchestrator.GetConfigDir())
```

## Phase 3: Clean Up and Test

### Step 3.1: Delete the runtime package
```bash
rm -rf internal/runtime/
```

### Step 3.2: Verify no remaining imports
```bash
grep -r "internal/runtime" internal/
# Should return no results
```

### Step 3.3: Run tests
```bash
go test ./...
just test-ux
```

## Critical Guidelines

### DO:
- ✅ Copy reconciliation logic exactly (preserve Managed/Missing/Untracked semantics)
- ✅ Replace every SharedContext call with direct function calls
- ✅ Create fresh instances instead of using cached ones
- ✅ Keep logging functionality identical
- ✅ Test after each phase

### DON'T:
- ❌ Change any reconciliation logic (must work identically)
- ❌ Add any caching or singleton patterns
- ❌ Modify logging behavior
- ❌ Skip updating any command file

## Expected Results
- **Package count**: Remains 13 (transformation, not deletion)
- **Code reduction**: ~100 LOC eliminated from over-engineering
- **Performance**: Same or better (no caching overhead)
- **Functionality**: Identical behavior, cleaner code

## Testing Verification
After completion, verify these specific behaviors work:
1. `plonk status` shows same Managed/Missing/Untracked items
2. `plonk doctor` reconciliation works identically
3. `plonk sync` applies changes correctly
4. All package manager operations work
5. Logging/debug output unchanged

The goal is simpler code with identical functionality.
