# Phase 0: Command Call Chain Analysis

## Overview

This document traces the execution flow from each command's `RunE` function to the core business logic, identifying layers of indirection that can be removed.

## Command Analysis

### 1. runStatus() (from root.go)

**Current Call Chain:**
```
runStatus()
  → Creates StatusOutput struct
  → Populates with package/dotfile counts
  → RenderOutput()
```

**Observations:**
- Already fairly direct
- Gets data from lock service and config adapter
- Main logic is UI formatting

**Refactoring Potential:** Low - already straightforward

### 2. runPkgList() (from shared.go)

**Current Call Chain:**
```
runPkgList()
  → core.CreatePackageProvider()
  → provider.GetCurrentState()
  → convertToPackageInfo()
  → RenderOutput()
```

**Observations:**
- Uses provider pattern from core
- Has conversion layer (convertToPackageInfo)
- Reasonably direct

**Refactoring Potential:** Medium - could remove conversion layer

### 3. runDotList() (from shared.go)

**Current Call Chain:**
```
runDotList()
  → core.CreateDotfileProvider()
  → provider.GetCurrentState()
  → convertToDotfileInfo()
  → RenderOutput()
```

**Observations:**
- Similar pattern to runPkgList
- Has conversion layer (convertToDotfileInfo)
- Uses provider from core

**Refactoring Potential:** Medium - could remove conversion layer

### 4. add.go

**Current Call Chain:**
```
runAdd()
  → ParseSimpleFlags()
  → addSingleDotfiles()
    → core.AddSingleDotfile() (for each file)
  → OR operations.BatchProcess()
    → core.AddDirectoryFiles()
  → RenderOutput()
```

**Observations:**
- Has wrapper function (addSingleDotfiles)
- Uses batch processing for directories
- Multiple paths through the code

**Refactoring Potential:** High - can simplify flow and remove wrappers

### 5. install.go

**Current Call Chain:**
```
runInstall()
  → ParseSimpleFlags()
  → operations.PackageProcessor()
    → installSinglePackage()
      → lockService operations
      → packageManager.Install()
  → operations.BatchProcess()
  → RenderOutput()
```

**Observations:**
- Uses operations.PackageProcessor abstraction
- Has local wrapper (installSinglePackage)
- Batch processing layer

**Refactoring Potential:** High - can flatten the processor pattern

### 6. sync.go

**Current Call Chain:**
```
runSync()
  → syncPackages()
    → applyPackages()
      → lockService.GetPackages()
      → operations.BatchProcess()
  → syncDotfiles()
    → applyDotfiles()
      → core.CreateDotfileProvider()
      → operations.BatchProcess()
  → RenderOutput()
```

**Observations:**
- Multiple layers of wrappers
- Reuses apply logic through indirection
- Complex flow through shared functions

**Refactoring Potential:** Very High - many layers to flatten

### 7. rm.go

**Current Call Chain:**
```
runRemove()
  → processor function (closure)
    → operations.SimpleProcessor()
      → core.RemoveSingleDotfile()
  → operations.BatchProcess()
  → RenderOutput()
```

**Observations:**
- Uses closure pattern for processor
- Goes through operations layer
- Has SimpleProcessor abstraction

**Refactoring Potential:** High - can simplify processor pattern

### 8. ls.go

**Current Call Chain:**
```
runList()
  → reconciler.New()
  → core.CreatePackageProvider()
  → core.CreateDotfileProvider()
  → reconciler.ReconcileAll()
  → formatters for output
  → RenderOutput()
```

**Observations:**
- Uses reconciler pattern from state package
- Already calls core providers directly
- Complex due to reconciliation logic

**Refactoring Potential:** Low - reconciler provides value

### 9. uninstall.go

**Current Call Chain:**
```
runUninstall()
  → operations.PackageProcessor()
    → uninstallSinglePackage()
      → lockService operations
      → packageManager.Uninstall()
  → operations.BatchProcess()
  → RenderOutput()
```

**Observations:**
- Similar pattern to install.go
- Uses operations abstractions
- Has local wrapper function

**Refactoring Potential:** High - same as install.go

## Summary of Findings

### High Priority Refactoring Targets:
1. **sync.go** - Most layers of indirection
2. **install.go/uninstall.go** - Similar patterns, can be simplified together
3. **add.go** - Complex flow with multiple paths
4. **rm.go** - Unnecessary processor abstractions

### Medium Priority:
1. **runPkgList()/runDotList()** - Conversion layers could be removed

### Low Priority:
1. **runStatus()** - Already direct
2. **ls.go** - Reconciler pattern provides value

## Common Patterns to Address:

1. **Operations Layer**: The `operations.BatchProcess` and processor patterns add indirection
2. **Wrapper Functions**: Many commands have thin wrappers that just forward calls
3. **Conversion Functions**: Converting between internal types and output types
4. **Shared Functions**: The `applyPackages`/`applyDotfiles` pattern in shared.go

## Recommended Approach:

1. Start with simple commands (Phase 1) to establish patterns
2. Move to complex commands (Phase 2) using lessons learned
3. Remove obsolete abstractions in Phase 3
