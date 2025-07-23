# Path Resolution Refactor: Completion Plan

## Current Status (Updated: Phase 2.1 Refactoring)

### Completed Actions:
1. ✅ **internal/dotfiles/operations.go** - GetDestinationPath now delegates to PathResolver
2. ✅ **internal/services/dotfile_operations.go** - AddSingleFile now properly handles path resolution errors without fallback
3. ✅ **internal/commands/shared.go** - Path resolution functions removed (moved to core/dotfiles.go in Phase 2.1)
4. ✅ **internal/config/yaml_config.go**:
   - GetDefaultConfigDirectory now delegates to paths.GetDefaultConfigDirectory()
   - Removed shouldSkipDotfile, sourceToTarget functions
   - TargetToSource now delegates to paths.TargetToSource()
   - GetDotfileTargets refactored to use PathResolver.ExpandConfigDirectory()
5. ✅ **internal/paths/resolver.go** - Enhanced with:
   - Exported GetDefaultConfigDirectory() function
   - Added ConfigDir() and HomeDir() getter methods
   - Added SourceToTarget() and TargetToSource() conversion functions
   - Added ExpandConfigDirectory() method for config directory traversal

### Questions/Decisions Needed:

1.  **CopyFileWithAttributes Duplication**
    -   Found two different implementations:
        -   `internal/services/dotfile_operations.go`: Creates destination directories
        -   `internal/core/dotfiles.go`: Simpler version without directory creation
    -   Both are actively used by their respective packages
    -   **Question**: Should we consolidate these into internal/dotfiles/fileops.go (which already exists with more sophisticated copy operations)?

2.  **PathResolver Design**
    -   Currently PathResolver mixes concerns: path resolution, validation, AND business logic (ExpandConfigDirectory)
    -   **Question**: Should ExpandConfigDirectory remain in PathResolver or move to a higher-level service?

3.  **Error Handling Strategy**
    -   Some refactored functions now return empty results on error (e.g., GetDotfileTargets)
    -   **Question**: Should we propagate errors up the call chain or maintain the current silent failure approach?

## My Responses to Questions/Decisions Needed

1.  **CopyFileWithAttributes Duplication**
    -   **Decision**: Yes, consolidate. Move both implementations into `internal/dotfiles/fileops.go`. This centralizes file operations and removes duplication. The `SHARED_GO_DECONSTRUCTION.md` plan already suggested this. Please ensure all callers are updated to use the single, consolidated function.

2.  **PathResolver Design (`ExpandConfigDirectory`)**
    -   **Decision**: For now, it's acceptable to keep `ExpandConfigDirectory` in `internal/paths`. The `paths` package is intended to be the central utility for *all* path-related concerns, including discovery. It uses the `PathResolver`'s core functions, rather than duplicating them. If it starts to accumulate too much business logic (e.g., directly interacting with `config.Config` objects for filtering), we can reconsider in a future refactor.

3.  **Error Handling Strategy**
    -   **Decision**: **Propagate errors up the call chain.** Silent failures are a major source of bugs and make debugging extremely difficult. This is standard Go idiom and crucial for improving reliability and maintainability. This will likely require updating many function signatures and error handling throughout the codebase, but it's a critical step for code quality.

## Remaining Work (Action Items for Bob)

To fully complete the path resolution refactor, please address the following:

1.  **Address Error Handling in `internal/dotfiles/operations.go`**: In `GetDestinationPath`, ensure that errors from `m.pathResolver.GetDestinationPath` are propagated up the call chain instead of being silently ignored (i.e., remove the `return destination` fallback on error).

2.  **Refactor `internal/core/dotfiles.go`**:
    *   **Remove Redundant Wrappers**:
        *   Delete the `ResolveDotfilePath` function. Callers should directly use `paths.NewPathResolver(...).ResolveDotfilePath(...)`.
        *   Delete the `GeneratePaths` function. Callers should directly use `paths.NewPathResolver(...).GeneratePaths(...)`. **Crucially, ensure the problematic fallback logic within this function (the `if err != nil { ... }` block that manually generates paths) is completely removed and that `paths.PathResolver.GeneratePaths` is robust enough to handle all cases without error.**
    *   **Move `CopyFileWithAttributes`**: Move this function to `internal/dotfiles/fileops.go` (or `internal/util/fileops.go` if it's truly generic and used outside dotfiles). Update all callers.

3.  **Consolidate `CopyFileWithAttributes`**: Ensure that the `CopyFileWithAttributes` in `internal/services/dotfile_operations.go` is also removed and replaced with a call to the consolidated function in `internal/dotfiles/fileops.go` (or `internal/util/fileops.go`).

## 3. Verification Steps

Upon completion of the above actions, the following verification steps must be performed:

1.  **Run Unit Tests:** Execute `just test`. All unit tests must pass.
2.  **Run UX Integration Tests:** Execute `just test-ux`. All integration tests must pass, confirming no user-facing behavior changes.
3.  **Code Inspection:** Manually inspect the modified files to ensure:
    *   No manual path manipulation (tilde expansion, `filepath.Join`, `os.Stat` for path existence/type) remains outside `internal/paths`.
    *   All path-related concerns are delegated to `internal/paths`.
    *   The `internal/paths` package is now the sole authority for these operations.
4.  **Confirm Deletion:** Ensure `shouldSkipDotfile`, `sourceToTarget`, and `TargetToSource` are removed from `yaml_config.go`.

## 4. Success Criteria

This phase of the refactor will be considered complete when:
*   All action items listed in section 2 are implemented.
*   All tests (`just test` and `just test-ux`) pass.
*   A manual code inspection confirms the complete delegation of path logic to `internal/paths`.

## 5. Next Steps

Once this plan is successfully executed and verified, we can confidently proceed with other refactoring efforts, knowing that our foundational path logic is sound and consistent.

## 6. Implementation Notes and Decisions Log

### Implementation Approach Taken:
1. **Incremental Refactoring**: Rather than a big-bang rewrite, we're updating each component to delegate to paths package
2. **Backward Compatibility**: Maintaining existing function signatures where possible to minimize disruption
3. **Centralization Over Distribution**: Moving all path-related logic to paths package, even if it means a larger package

### Key Design Decisions:
1. **PathResolver Enhancements**:
   - Added getter methods (ConfigDir(), HomeDir()) to expose internal state
   - Exported GetDefaultConfigDirectory() for use by config package
   - Added business-specific method ExpandConfigDirectory() - this may need reconsideration

2. **Source/Target Conversion**:
   - Moved SourceToTarget and TargetToSource to paths package as pure functions
   - These are core to plonk's dotfile management convention and belong with path logic

3. **Error Handling**:
   - Currently maintaining existing behavior (silent failures in some cases)
   - This preserves backward compatibility but may hide issues

### Testing Status:
- ✅ Unit tests: All tests pass after refactoring
- ⏳ Integration tests (test-ux): Not yet run after refactoring
- ⏳ Manual testing of key commands: Not yet performed

### Completion Status:
1. ✅ CopyFileWithAttributes consolidated into internal/dotfiles/fileops.go
2. ✅ All error handling updated to propagate errors (no silent failures)
3. ✅ Redundant path resolution wrappers removed from core/dotfiles.go
4. ✅ All path resolution now delegated to internal/paths package

### Summary of Changes Made:
1. **GetDestinationPath** in operations.go now returns error and all callers updated
2. **CopyFileWithAttributes** consolidated into single implementation in fileops.go
3. **ResolveDotfilePath** and **GeneratePaths** wrappers removed from core/dotfiles.go
4. **shouldSkipDotfile**, **sourceToTarget**, **TargetToSource** moved to paths package
5. **GetDefaultConfigDirectory** now delegates to paths.GetDefaultConfigDirectory()
6. **GetDotfileTargets** refactored to use PathResolver.ExpandConfigDirectory()

### Final Verification Needed:
- Run integration tests (test-ux) to ensure no user-facing changes
- Manual testing of key dotfile commands (add, rm, sync)
