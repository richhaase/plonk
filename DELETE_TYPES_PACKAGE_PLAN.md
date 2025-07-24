# Plan: Delete Types Package

## Objective
Delete the `internal/types` package and move its functionality to more appropriate locations, reducing unnecessary abstraction.

## Current State
- `internal/types/common.go` contains:
  - `Result` struct - for state reconciliation results
  - `Summary` struct - for aggregate counts
  - Methods on Result: `Count()`, `IsEmpty()`, `AddToSummary()`
  - Type aliases for `Item` and `ItemState` from interfaces package
  - Constant re-exports for state constants

- Used by 5 files:
  - commands/ls.go
  - commands/shared.go
  - commands/status.go
  - runtime/context.go
  - runtime/context_simple.go

## Analysis
The types package is mostly re-exporting types from interfaces and adding two structs (Result and Summary) that are used for state reconciliation. This adds an unnecessary layer of indirection.

## Plan

### Step 1: Decide where to move Result and Summary
- These types are used for state reconciliation results
- Best location: `internal/state` package (where state reconciliation happens)

### Step 2: Update all imports
Replace imports in 5 files:
- Change `"github.com/richhaase/plonk/internal/types"` to:
  - `"github.com/richhaase/plonk/internal/state"` for Result/Summary
  - `"github.com/richhaase/plonk/internal/interfaces"` for Item/ItemState (if needed)

### Step 3: Move the code
1. Copy Result and Summary structs + methods to `internal/state/types.go` (new file)
2. Update package declaration from `package types` to `package state`
3. Remove the type aliases and constant re-exports (use interfaces directly)

### Step 4: Update references
- Change `types.Result` → `state.Result`
- Change `types.Summary` → `state.Summary`
- Change `types.Item` → `interfaces.Item`
- Change `types.ItemState` → `interfaces.ItemState`
- Change state constants to use interfaces directly

### Step 5: Clean up
1. Delete `internal/types/` directory
2. Ensure all references are updated

## Expected Changes
- 5 files will have import statements changed
- Type references will be more direct (no unnecessary aliases)
- Package count reduced by 1

## Risks
- Medium risk - need to ensure all type references are updated correctly
- The state package is a logical home for Result/Summary since they're used in reconciliation

## Validation
1. Run `go build ./...` to ensure compilation
2. Run unit tests: `just test`
3. Run UX/integration tests: `just test-ux`
