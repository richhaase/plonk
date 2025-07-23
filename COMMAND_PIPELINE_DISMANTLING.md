# Command Pipeline Dismantling Plan

## Status Update

**Last Updated:** Phase 1 Complete (All simple commands refactored)

### Progress Summary:
- ✅ Phase 0: Call chain analysis complete
- ✅ Phase 1.1: runStatus() reviewed (already direct)
- ✅ Phase 1.2: runPkgList() refactored to eliminate conversion layer
- ✅ Phase 1.3: runDotList() refactored to eliminate conversion layer
- ⏳ Phase 2: Complex commands pending (ready to start)
- ⏳ Phase 3: Final cleanup pending

### Key Achievements:
- Successfully eliminated conversion layers in both runPkgList() and runDotList()
- Established consistent pattern for wrapping raw domain objects for OutputData interface
- Maintained 100% test compatibility while simplifying code
- Reduced shared.go from ~600 to 534 lines

## 1. Overview & Goal (Revised)

**The Problem:** While the top-level command pipeline has been removed, individual commands (e.g., `runStatus`, `runPkgList`, `runDotList`, `add.go`, `install.go`) may still contain excessive layers of abstraction and indirection before reaching the core business logic in `internal/core` or `internal/services`. This contributes to cognitive load and makes the code harder to understand and maintain.

**The Goal:** To flatten the internal call chains within individual commands, ensuring they directly call the relevant business logic functions in `internal/core` or `internal/services`. This will reduce unnecessary indirection and improve code clarity.

## Current State Analysis (Phase 0 Complete)

The detailed call chain analysis has been completed and documented in `PHASE_0_CALL_CHAIN_ANALYSIS.md`. Key findings:

### Patterns to Eliminate:
1. **Operations Layer**: `operations.BatchProcess` and processor patterns add unnecessary indirection
2. **Wrapper Functions**: Thin wrappers like `syncPackages()` that just forward calls
3. **Conversion Functions**: Type converters like `convertToPackageInfo()` between internal types and output
4. **Shared Functions**: The `applyPackages`/`applyDotfiles` pattern in shared.go

### Command Prioritization:
- **High Priority**: sync.go (most layers), install/uninstall.go, add.go, rm.go
- **Medium Priority**: runPkgList/runDotList (conversion layers)
- **Low Priority**: runStatus (already direct), ls.go (reconciler provides value)

**Guiding Principles:**
*   **Directness:** Favor direct function calls over complex abstractions where no clear value is added.
*   **Clarity:** Make the flow of execution for each command explicit and easy to follow.
*   **No UX Change:** The user experience of the CLI must remain identical.
*   **Test-Driven:** All changes must be verified by existing tests (`just test`, `just test-ux`).

## 2. Detailed Refactoring Phases

We will tackle this command by command, focusing on the `RunE` function of each Cobra command.

### Phase 0: Identify Commands and Current Call Chains ✅ COMPLETE

See `PHASE_0_CALL_CHAIN_ANALYSIS.md` for detailed analysis. All commands have been analyzed and prioritized based on their refactoring potential.

### Phase 1: Refactor Simple Commands

Start with commands that have relatively straightforward logic.

1.  **Target: `runStatus()` (from `status.go`)** ✅ COMPLETE
    *   **Current State:** Already fairly direct - minimal changes needed
    *   **Action:** Reviewed and confirmed it already follows the direct call pattern
    *   **Result:** No changes needed - already a good example of direct calls
    *   **Verification:** All tests pass (`just test`, `just test-ux`)

2.  **Target: `runPkgList()` (from `shared.go`)** ✅ COMPLETE
    *   **Current State:** Had conversion layer with `EnhancedPackageOutput` and `PackageListOutput`
    *   **Actions Taken:**
        - Removed old `runPkgListOld()` function with conversion layers
        - Created `packageListResultWrapper` to wrap `state.Result` and implement `OutputData`
        - Eliminated intermediate data structures (`EnhancedPackageOutput`, `PackageListOutput`)
        - Function now passes raw `state.Result` directly to `RenderOutput()`
    *   **Result:** Direct data flow from reconciler to output formatter
    *   **Verification:** All tests pass, manually tested `plonk ls --packages`

3.  **Target: `runDotList()` (from `shared.go`)** ✅ COMPLETE
    *   **Current State:** Had `convertToDotfileInfo()` conversion layer
    *   **Actions Taken:**
        - Removed `convertToDotfileInfo()` function
        - Created `dotfileListResultWrapper` to wrap `state.Result` and implement `OutputData`
        - Eliminated intermediate `DotfileListOutput` structure
        - Function now passes raw `state.Result` directly to `RenderOutput()`
        - Maintained backward compatibility for structured output (JSON/YAML)
    *   **Result:** Direct data flow from reconciler to output formatter
    *   **Verification:** All tests pass, manually tested `plonk ls --dotfiles`

### Lessons Learned from Phase 1

The successful completion of Phase 1 has validated our approach and established patterns for Phase 2:

1. **Pattern Success**: The wrapper pattern (`state.Result` → thin wrapper → `OutputData`) works well
2. **No Data Loss**: Raw domain objects contain all necessary information for display
3. **Backward Compatibility**: Structured output (JSON/YAML) can be maintained while simplifying code
4. **Test Stability**: All tests continue to pass without modification
5. **Code Clarity**: Direct data flow is much easier to understand than conversion layers

### Phase 2: Refactor Complex Commands

Tackle commands with more involved logic, such as `add`, `install`, `sync`, `remove`.

**Key Difference from Phase 1**: These commands use `operations.BatchProcess` and processor patterns that need to be eliminated, requiring more significant restructuring.

1.  **Target: `add.go`**
    *   **Current State:** Uses `addSingleDotfiles()` wrapper and `operations.BatchProcess()`
    *   **Action:**
        - Remove `addSingleDotfiles()` wrapper function
        - Eliminate `operations.BatchProcess()` usage
        - Have `RunE` directly iterate and call `core.AddSingleDotfile()` or `core.AddDirectoryFiles()`
        - Handle error collection and result aggregation directly in `RunE`
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk add`.

2.  **Target: `install.go`**
    *   **Current State:** Uses `operations.PackageProcessor()` and `installSinglePackage()` wrapper
    *   **Action:**
        - Remove `operations.PackageProcessor()` and `operations.BatchProcess()`
        - Remove `installSinglePackage()` wrapper
        - Have `RunE` directly iterate over packages
        - Call `lockService` and `packageManager.Install()` directly
        - Handle results and error collection in `RunE`
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk install`.

3.  **Target: `sync.go`**
    *   **Current State:** Most complex with multiple wrapper layers
    *   **Action:**
        - Remove `syncPackages()`, `syncDotfiles()`, `applyPackages()`, and `applyDotfiles()` functions
        - Have `RunE` directly call `services.ApplyPackages()` and equivalent dotfile services
        - Handle the orchestration and result combination directly in `RunE`
        - Eliminate intermediate result transformations
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk sync`.

4.  **Target: `rm.go`**
    *   **Current State:** Uses `operations.SimpleProcessor()` and closure pattern
    *   **Action:**
        - Remove `operations.SimpleProcessor()` and `operations.BatchProcess()`
        - Eliminate the processor closure pattern
        - Have `RunE` directly iterate over dotfiles and call `core.RemoveSingleDotfile()`
        - Handle results directly in `RunE`
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk rm`.

5.  **Target: `uninstall.go`**
    *   **Current State:** Similar pattern to install.go
    *   **Action:**
        - Remove `operations.PackageProcessor()` and `operations.BatchProcess()`
        - Remove `uninstallSinglePackage()` wrapper
        - Have `RunE` directly orchestrate the uninstallation
        - Call lock service and package manager directly
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk uninstall`.

### Phase 3: Final Cleanup

1.  **Remove Obsolete Functions/Files:**
    - Delete helper functions that become unused after refactoring
    - Remove `operations.BatchProcess` and related processor patterns if no longer used
    - Clean up any conversion functions that are no longer needed

2.  **Review Remaining `shared.go`:**
    - After Phase 2, most functions in `shared.go` should be obsolete
    - Move any remaining UI type aliases to `internal/ui` or relevant output files
    - Goal: Minimize or completely eliminate `shared.go`

3.  **Operations Package Review:**
    - Assess if the `operations` package is still needed after removing batch processors
    - Consider moving any remaining valuable abstractions to more appropriate locations

## 3. Risk Analysis & Mitigation

*   **Risk: Breaking Command Logic:**
    *   **Mitigation:** Small, incremental changes. Thorough testing (`just test`, `just test-ux`) after each refactored command. Manual testing of the specific command.
*   **Risk: Introducing Circular Dependencies:**
    *   **Mitigation:** Adhere strictly to package boundaries. `internal/commands` can import `internal/core` and `internal/services`, but `internal/core` and `internal/services` should not import `internal/commands`.
*   **Risk: Increased Complexity in `RunE`:**
    *   **Mitigation:** The goal is to flatten, not to create a monolithic `RunE`. If a `RunE` becomes too large, it indicates that the core logic it's calling might need further decomposition within `internal/core` or `internal/services`.

## 4. Expected Outcomes

After completing all phases:

1. **Direct Command Execution**: Each command's `RunE` function will directly orchestrate business logic calls
2. **Eliminated Abstractions**: No more `operations.BatchProcess`, processor patterns, or unnecessary wrappers
3. **Clearer Data Flow**: Data flows directly from business logic to UI rendering without conversions
4. **Reduced File Count**: `shared.go` minimized or eliminated, operations package simplified
5. **Improved Maintainability**: Developers can trace command execution without jumping through layers

## 5. Next Steps

With Phase 0 complete and Ed's approval, we're ready to begin Phase 1 with `runStatus()` (minimal work needed) followed by `runPkgList()` and `runDotList()` to establish patterns for the more complex refactoring in Phase 2.
