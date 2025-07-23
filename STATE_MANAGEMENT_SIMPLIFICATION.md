# State Management Simplification Plan

## 1. Overview & Goal

**The Problem:** The `CODE_REVIEW.md` identified the state reconciliation system as over-engineered, citing a "Generic `Provider` interface," "Complex diffing logic," and "Multiple abstraction layers." This leads to a system where "simple set operations (configured - actual = to_install) spans multiple files and interfaces," increasing cognitive load and making it difficult to understand and modify the core state management.

**The Goal:** To simplify the state management and reconciliation process by:
*   Reducing unnecessary interfaces and abstraction layers.
*   Making the reconciliation logic more direct and explicit.
*   Consolidating related state management concerns into fewer, more focused components.
*   Improving the clarity and maintainability of how `plonk` determines the difference between configured and actual system states.

**Guiding Principles:**
*   **Directness:** Favor direct function calls and concrete types over interfaces and abstractions where no clear value is added.
*   **Clarity:** Make the state reconciliation logic easy to follow and understand.
*   **No UX Change:** The user experience of the CLI (e.g., `plonk status`, `plonk sync`) must remain identical.
*   **Test-Driven:** All changes must be verified by existing tests (`just test`, `just test-ux`).

## 2. Current State Analysis (Phase 0) ✅ COMPLETE

*(See `STATE_ANALYSIS.md` for detailed findings)*

## 3. Refactoring Strategy: "Simplify Providers, Then Reconciler"

We will proceed iteratively, focusing on simplifying the data sources (providers) before tackling the core reconciliation logic.

### Phase 1: Provider Simplification ✅ COMPLETE

**Decision:** Retain `ConfigItem`, `ActualItem`, and `Item` as distinct types due to their distinct semantic purposes.

1.  **Target: `Provider` Interface (`internal/interfaces/core.go`)**
    *   **Action:** Analyze the `Provider` interface. Determine if it truly requires polymorphism, or can its functionality be directly integrated into the `Reconciler` or a simpler helper?
    *   **Action:** If the `Provider` interface can be removed, refactor `internal/state/reconciler.go` to directly interact with `dotfile_provider.go` and `package_provider.go` (or their simplified equivalents).
    *   **Action:** If the `Provider` interface is retained, simplify its methods and ensure they are minimal and focused.
2.  **Target: `DotfileProvider` (`internal/state/dotfile_provider.go`)**
    *   **Action:** Simplify the internal logic of `DotfileProvider`. Ensure it directly performs its function of providing dotfile state without unnecessary layers.
    *   **Action:** Remove any redundant interfaces or adapters within `DotfileProvider`.
3.  **Target: `PackageProvider` (`internal/state/package_provider.go`)**
    *   **Action:** Simplify the internal logic of `PackageProvider`. Ensure it directly performs its function of providing package state without unnecessary layers.
    *   **Action:** Remove any redundant interfaces or adapters within `PackageProvider`.
4.  **Verification:** Run `just test` and `just test-ux`. Manually test `plonk status` and `plonk ls`.
5.  **Commit:** Commit changes for each provider or interface.

*(See `PROVIDER_SIMPLIFICATION_SUMMARY.md` for detailed completion report of Phase 1)*

### Phase 2: Simplify Reconciler Logic

This phase targets `internal/state/reconciler.go` and `internal/state/types.go`.

1.  **Target: `Reconciler` (`internal/state/reconciler.go`)**
    *   **Action:** Refactor the `Reconciler` to directly perform the diffing logic (configured vs. actual). Eliminate any complex, generic "diffing" abstractions that obscure the simple set operations.
    *   **Action:** Ensure the `Reconciler` directly uses the simplified providers (from Phase 1) without additional indirection.
2.  **Target: State Types (`internal/state/types.go`)**
    *   **Action:** Review `Result` and `Item` types. Are they overly generic? Can they be simplified or made more specific to dotfiles/packages if the `Provider` interface is removed?
    *   **Action:** Ensure the types are clear and directly represent the state information needed by the `Reconciler` and for output.
3.  **Verification:** Run `just test` and `just test-ux`. Manually test `plonk status`, `plonk sync`.
4.  **Commit:** Commit changes for each logical simplification.

### Phase 3: Final Cleanup

1.  **Remove Obsolete Files/Interfaces:** Delete any interfaces, types, or files that become obsolete after the simplification.
2.  **Review Callers:** Ensure all parts of the codebase that interact with the state management system are updated to use the new, simplified API.
3.  **Quantitative Assessment:** Run `scc` to measure the reduction in lines of code and complexity within the `internal/state` package.

## 4. Risk Analysis & Mitigation

*   **Risk: Breaking State Consistency:**
    *   **Mitigation:** Small, incremental changes. Thorough testing (`just test`, `just test-ux`) after each modification. Manual testing of `plonk status`, `plonk sync`, `plonk add`, `plonk rm`.
*   **Risk: Reintroducing Complexity:**
    *   **Mitigation:** Strict adherence to the "Directness" and "Clarity" principles. Avoid adding new layers of abstraction unless absolutely necessary and clearly justified.
*   **Risk: Circular Dependencies:**
    *   **Mitigation:** Maintain strict package boundaries. `internal/state` should not import `internal/commands` or `internal/cli`.

## 5. Success Criteria

*   Significant reduction in the number of interfaces and abstraction layers within `internal/state`.
*   The state reconciliation logic is direct, explicit, and easy to understand.
*   The `Reconciler` directly interacts with simplified providers.
*   All unit and integration tests pass without modification.
*   The user experience of the CLI remains identical.
*   Quantifiable reduction in lines of code and complexity within the `internal/state` package.
