# Command Pipeline Dismantling Plan

## 1. Overview & Goal (Revised)

**The Problem:** While the top-level command pipeline has been removed, individual commands (e.g., `runStatus`, `runPkgList`, `runDotList`, `add.go`, `install.go`) may still contain excessive layers of abstraction and indirection before reaching the core business logic in `internal/core` or `internal/services`. This contributes to cognitive load and makes the code harder to understand and maintain.

**The Goal:** To flatten the internal call chains within individual commands, ensuring they directly call the relevant business logic functions in `internal/core` or `internal/services`. This will reduce unnecessary indirection and improve code clarity.

**Guiding Principles:**
*   **Directness:** Favor direct function calls over complex abstractions where no clear value is added.
*   **Clarity:** Make the flow of execution for each command explicit and easy to follow.
*   **No UX Change:** The user experience of the CLI must remain identical.
*   **Test-Driven:** All changes must be verified by existing tests (`just test`, `just test-ux`).

## 2. Detailed Refactoring Phases

We will tackle this command by command, focusing on the `RunE` function of each Cobra command.

### Phase 0: Identify Commands and Current Call Chains

1.  **List all commands:** Identify all `.go` files in `internal/commands` that define Cobra commands.
2.  **Trace Call Chains:** For each command, trace its execution flow from its `RunE` function down to the core business logic. Document the current layers of indirection.

### Phase 1: Refactor Simple Commands

Start with commands that have relatively straightforward logic.

1.  **Target: `runStatus()` (from `root.go`)**
    *   **Action:** Analyze `runStatus()` to identify its dependencies and the core logic it performs. Refactor it to directly call functions in `internal/core` or `internal/services` as much as possible, removing any intermediate helper functions or unnecessary layers.
    *   **Verification:** Run `just test` and `just test-ux`. Manually run `plonk status`.
2.  **Target: `runPkgList()` (from `shared.go`)**
    *   **Action:** Analyze `runPkgList()` and refactor it to directly call functions in `internal/core` or `internal/services`, simplifying its internal logic.
    *   **Verification:** Run `just test` and `just test-ux`. Manually run `plonk ls --packages`.
3.  **Target: `runDotList()` (from `shared.go`)**
    *   **Action:** Analyze `runDotList()` and refactor it to directly call functions in `internal/core` or `internal/services`, simplifying its internal logic.
    *   **Verification:** Run `just test` and `just test-ux`. Manually run `plonk ls --dotfiles`.

### Phase 2: Refactor Complex Commands

Tackle commands with more involved logic, such as `add`, `install`, `sync`, `remove`.

1.  **Target: `add.go` (and related `addSingleDotfile`, `AddSingleFile`, `AddDirectoryFiles` in `core/dotfiles.go`)**
    *   **Action:** Analyze the `add` command's `RunE` and the functions it calls. Flatten the call chain by having `add.go` directly orchestrate calls to `internal/core/dotfiles.go` functions, removing any unnecessary intermediate functions or wrappers.
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk add`.
2.  **Target: `install.go`**
    *   **Action:** Analyze the `install` command's `RunE` and its call chain. Refactor to directly call `internal/core/packages.go` functions.
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk install`.
3.  **Target: `sync.go`**
    *   **Action:** Analyze the `sync` command's `RunE` and its call chain. Refactor to directly call `internal/core` functions for both packages and dotfiles.
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk sync`.
4.  **Target: `rm.go`**
    *   **Action:** Analyze the `rm` command's `RunE` and its call chain. Refactor to directly call `internal/core` functions.
    *   **Verification:** Run `just test` and `just test-ux`. Manually test `plonk rm`.

### Phase 3: Final Cleanup

1.  **Remove Obsolete Functions/Files:** Delete any helper functions or files that become obsolete after flattening the call chains.
2.  **Review Remaining `shared.go`:** Re-evaluate the remaining functions in `shared.go` (`applyPackages`, `applyDotfiles`, `convertToDotfileInfo`, UI type aliases). If they can be further simplified or moved to more specific packages (e.g., `internal/services` or `internal/ui`), do so. The ultimate goal is still to minimize or eliminate `shared.go`.

## 3. Risk Analysis & Mitigation

*   **Risk: Breaking Command Logic:**
    *   **Mitigation:** Small, incremental changes. Thorough testing (`just test`, `just test-ux`) after each refactored command. Manual testing of the specific command.
*   **Risk: Introducing Circular Dependencies:**
    *   **Mitigation:** Adhere strictly to package boundaries. `internal/commands` can import `internal/core` and `internal/services`, but `internal/core` and `internal/services` should not import `internal/commands`.
*   **Risk: Increased Complexity in `RunE`:**
    *   **Mitigation:** The goal is to flatten, not to create a monolithic `RunE`. If a `RunE` becomes too large, it indicates that the core logic it's calling might need further decomposition within `internal/core` or `internal/services`.
