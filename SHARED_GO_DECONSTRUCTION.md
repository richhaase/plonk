# Deconstruction Plan: `internal/commands/shared.go`

## 1. Overview & Goal

**The Problem:** The file `internal/commands/shared.go` is a 1000+ line "junk drawer" that violates the Single Responsibility Principle. It contains a mix of core business logic, UI rendering, state management, and error handling. This is the primary source of the "Business Logic Scatter" identified in the `CODE_REVIEW.md`, making the codebase impossible to reason about.

**The Goal:** The ultimate goal is the **complete deletion of the `internal/commands/shared.go` file**. All logic within it will be moved to a more appropriate, well-defined package.

**Guiding Principles:**
*   The user experience (UX) of the CLI must not change. All commands must function identically from a user's perspective.
*   The internal implementation is entirely flexible. No backward compatibility is required for the internal Go APIs we are refactoring.
*   Every step must leave the codebase in a working, testable state.

## 2. The Strategy: "Move, Verify, Repeat"

We will approach this refactor with a simple, methodical, and low-risk process for each function or closely-related group of functions:

1.  **Identify & Move:** Select a function and move it to its new, logical home.
2.  **Update Callers:** Update all places in the codebase that called the function to use its new location.
3.  **Verify with Tests:** Run the entire test suite (`just test` and `just test-ux`) to ensure no regressions were introduced.
4.  **Commit:** Commit the small, successful change.
5.  **Repeat:** Select the next function and repeat the process.

This iterative approach ensures that if any step introduces an issue, the cause is immediately obvious and easy to revert.

## 3. Function Analysis & New Homes

The logic within `shared.go` can be categorized into distinct domains. Each will be moved to a new or existing package that aligns with its purpose.

| Category              | Example Functions (from `shared.go`)                               | New Home                               | Justification                                                                                             |
| --------------------- | ------------------------------------------------------------------ | -------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| **Core Dotfile Logic**  | `addSingleDotfile`, `removeSingleDotfile`, `linkDotfile`             | `internal/core/dotfiles.go`            | Centralizes all business logic for managing dotfiles into a single, cohesive package.                     |
| **Core Package Logic**  | `installPackages`, `uninstallPackages`, `getManagerForPackage`     | `internal/core/packages.go`            | Centralizes all business logic for managing packages.                                                     |
| **State & Config**      | `loadState`, `getResolvedConfig`, `saveLockFile`                     | `internal/core/state.go`               | Consolidates state and configuration loading, which is core logic, not command logic.                     |
| **UI & Output**         | `printOperationReport`, `renderTable`, `formatCompletion`          | `internal/ui/`                         | Separates presentation logic from business logic, making both easier to test and modify independently.    |
| **Error Handling**      | `handlePipelineError`, `logError`                                  | `internal/errors/`                     | Consolidates error handling and reporting mechanisms.                                                     |
| **Command Helpers**     | `getManagerFromArgs`, `confirmAction`                              | `internal/cli/helpers.go`              | For logic tightly coupled to the CLI interaction itself, but not core business logic.                     |

## 4. Detailed Refactoring Phases

This is the step-by-step execution plan.

### Phase 0: Preparation (No-Risk Setup) ✅ COMPLETE

1.  **Create Destination Packages:** ✅
    *   Create directory `internal/core` with `dotfiles.go`, `packages.go`, and `state.go`. ✅
    *   Create directory `internal/ui` with `tables.go` and `formatters.go`. ✅
    *   Create directory `internal/cli` with `helpers.go`. ✅
2.  **Verify:** Run `just test`. No code has changed, so all tests should pass. This confirms the new structure is sound. ✅

**Status:** All destination packages created successfully. Tests pass. Ready to proceed with Phase 1.

### Phase 1: Migrate Low-Dependency Functions (Quick Wins)

We will start with functions that have few dependencies on other logic within `shared.go`.

1.  **Target: UI & Output Functions.**
    *   **Action:** Move table rendering and output formatting functions to `internal/ui/`.
    *   **Verification:** Run tests. Manually run a command that produces table output (e.g., `plonk ls`) to confirm the UX is identical.
2.  **Target: Error Handling Functions.**
    *   **Action:** Move the error handling and logging helpers to `internal/errors/`.
    *   **Verification:** Run tests.

### Phase 2: Migrate Core Business Logic (The Main Event)

This is the most critical phase. We will untangle the core application logic.

1.  **Target: Dotfile Logic.**
    *   **Action:** Move all functions related to adding, removing, and linking dotfiles from `shared.go` to `internal/core/dotfiles.go`. This will likely be a group of functions moved together.
    *   **Verification:** Run `just test`. Pay special attention to the integration tests by running `just test-ux`. Manually validate `plonk add`, `plonk rm`, and `plonk sync` for dotfiles.
2.  **Target: Package Logic.**
    *   **Action:** Move all functions related to installing and uninstalling packages to `internal/core/packages.go`.
    *   **Verification:** Run tests and `test-ux`. Manually validate `plonk install`, `plonk uninstall`, and `plonk sync` for packages.
3.  **Target: State and Config Logic.**
    *   **Action:** Move state and config loading functions to `internal/core/state.go`.
    *   **Verification:** Run tests. Manually validate commands that rely on the configuration file (`plonk.yaml`) and lock file.

### Phase 3: Final Cleanup

1.  **Target: Remaining Helpers.**
    *   **Action:** Move any remaining CLI-specific helpers to `internal/cli/helpers.go`.
    *   **Verification:** Run tests.
2.  **Target: Delete `shared.go`**
    *   **Action:** At this point, `internal/commands/shared.go` should be empty. Delete the file.
    *   **Verification:** Run `just test` one last time to ensure the project still builds and all tests pass. This is the final success metric.

## 5. Risk Analysis & Mitigation

The codebase's complexity presents several risks. Here is how we will mitigate them.

*   **Risk: Hidden Dependencies within `shared.go`.**
    *   **Description:** A function we move may depend on another function still in `shared.go`, making the move complex.
    *   **Mitigation:** Our function categorization in step 3 will help identify these coupled groups. We will move tightly coupled functions as a single unit. The small, incremental nature of the plan means we will discover these dependencies quickly and can adjust the plan for that unit.

*   **Risk: Introducing a Circular Dependency.**
    *   **Description:** Moving a function to `internal/core` could create a situation where `core` needs to import a higher-level package like `internal/commands`, which is an architectural anti-pattern.
    *   **Mitigation:** This is a hard rule: **`internal/core` must never import `internal/commands` or `internal/cli`**. If moving a function would cause this, it's a signal that the function is not pure business logic. We will stop, analyze, and break the function down further, leaving the part that depends on the CLI in the `cli` package and moving only the pure logic to `core`.

*   **Risk: UX Regression.**
    *   **Description:** A change could subtly alter the output format, confirmation prompts, or error messages.
    *   **Mitigation:** The `test-ux` integration test is our primary automated defense. Furthermore, the plan explicitly includes manual validation steps for key commands after each major logic migration. We will visually inspect the CLI output to ensure it remains identical.

*   **Risk: A Step Fails.**
    *   **Description:** A function is moved, and a test unexpectedly fails.
    *   **Mitigation:** Because each step is atomic and committed upon success, we can immediately revert the single change (`git reset --hard HEAD`). The cause of the failure is isolated to the last function moved. We will then analyze the failing test, adjust the plan for that specific function, and try again.

## 6. Progress Tracking

### Phase Status
- ✅ Phase 0: Preparation - COMPLETE
- ⏳ Phase 1: Migrate Low-Dependency Functions - NOT STARTED
- ⏳ Phase 2: Migrate Core Business Logic - NOT STARTED
- ⏳ Phase 3: Final Cleanup - NOT STARTED

### Functions Moved
(This section will be updated as functions are moved from shared.go)

| Function | Original Location | New Location | Status |
| -------- | ---------------- | ------------ | ------ |
| (none yet) | | | |
