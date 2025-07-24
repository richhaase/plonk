# State Type Alias Cleanup Plan

## 1. Overview & Goal

**The Problem:** During the "State Management Simplification" refactor, type aliases in `internal/state/types.go` were retained for API stability, but were noted as being marked for future deprecation. These aliases contribute to "Type Confusion" and "Abstraction Maze" by obscuring the true underlying types and adding an unnecessary layer of indirection.

**The Goal:** To eliminate unnecessary type aliases within `internal/state/types.go` and update all their callers to use the direct, underlying types. This will:
*   Improve code clarity and readability.
*   Reduce cognitive load for developers.
*   Make the codebase more direct and easier to navigate.
*   Further align the codebase with idiomatic Go practices.

**Guiding Principles:**
*   **Directness:** Use the underlying concrete types directly.
*   **Clarity:** Ensure the code is explicit about the types being used.
*   **No UX Change:** The user experience of the CLI must remain identical.
*   **Test-Driven:** All changes must be verified by existing tests (`just test`, `just test-ux`).

## 2. Current State Analysis (Phase 0)

Before making any changes, we need to identify the specific type aliases and their usage.

### Phase 0.1: Identify Type Aliases

1.  **Action:** List all type aliases defined in `internal/state/types.go`.
    *   *(Self-correction: I will read `internal/state/types.go` now to get this list.)*

### Phase 0.2: Identify Callers

1.  **Action:** For each alias, identify all files that import `internal/state` and then use the alias (e.g., `state.Item`).
2.  **Action:** Document the number of occurrences for each alias.

### Phase 0.3: Document Findings

1.  **Action:** Create a new document (e.g., `STATE_TYPE_ALIAS_ANALYSIS.md`) to list all findings from Phase 0.1 and 0.2. This will serve as our working document for this refactor.

## 3. Refactoring Strategy: "Replace, Remove"

We will proceed iteratively, replacing alias usage with direct type usage.

### Phase 1: Replace `ItemState` Alias

1.  **Target:** All files using `state.ItemState` or `state.StateManaged`, `state.StateMissing`, `state.StateUntracked`.
2.  **Action:** Change `state.ItemState` to `interfaces.ItemState`.
3.  **Action:** Change `state.StateManaged` to `interfaces.StateManaged`, etc.
4.  **Verification:** Run `just test` and `just test-ux`.
5.  **Commit:** Commit the changes.

### Phase 2: Replace `Item` Alias

1.  **Target:** All files using `state.Item`.
2.  **Action:** Change `state.Item` to `interfaces.Item`.
3.  **Verification:** Run `just test` and `just test-ux`.
4.  **Commit:** Commit the changes.

### Phase 3: Replace `ConfigItem` and `ActualItem` Aliases

1.  **Target:** All files using `state.ConfigItem` and `state.ActualItem`.
2.  **Action:** Change `state.ConfigItem` to `interfaces.ConfigItem`.
3.  **Action:** Change `state.ActualItem` to `interfaces.ActualItem`.
4.  **Verification:** Run `just test` and `just test-ux`.
5.  **Commit:** Commit the changes.

### Phase 4: Replace `Result` and `Summary` Aliases

1.  **Target:** All files using `state.Result` and `state.Summary`.
2.  **Action:** Change `state.Result` to `types.Result`.
3.  **Action:** Change `state.Summary` to `types.Summary`.
4.  **Verification:** Run `just test` and `just test-ux`.
5.  **Commit:** Commit the changes.

### Phase 5: Final Cleanup

1.  **Remove Obsolete Aliases:**
    *   **Action:** Once all callers of an alias have been updated, delete the alias definition from `internal/state/types.go`.
    *   **Action:** Remove the corresponding `import "github.com/richhaase/plonk/internal/interfaces"` or `import "github.com/richhaase/plonk/internal/types"` from `internal/state/types.go` if no other types from those packages are used.
2.  **Quantitative Assessment:** Run `scc` to measure the reduction in lines of code and complexity within the `internal/state/types.go` file and overall.

## 4. Risk Analysis & Mitigation

*   **Risk: Breaking Functionality:**
    *   **Mitigation:** Small, incremental changes. Go's strong type system will catch many errors at compile time. Thorough testing (`just test`, `just test-ux`) after each alias replacement.
*   **Risk: Missing Callers:**
    *   **Mitigation:** Use IDE's "find all references" or `grep` to ensure all usages are updated. Compilation errors will also highlight missed callers.
*   **Risk: Circular Dependencies:**
    *   **Mitigation:** This refactor should *reduce* dependencies on `internal/state/types.go`, not create new circular ones. If a circular dependency arises, it indicates a deeper issue that needs to be addressed.

## 5. Success Criteria

*   All type aliases are removed from `internal/state/types.go`.
*   All callers use the direct, underlying types.
*   The `internal/state/types.go` file is significantly smaller and simpler.
*   All unit and integration tests pass without modification.
*   The user experience of the CLI remains identical.
*   Quantifiable reduction in lines of code and complexity within the `internal/state` package.
