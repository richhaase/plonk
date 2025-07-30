# Dotfile Drift Detection - Implementation Tasks

## Phase 1: MVP Implementation (2-3 days)

### Day 1: Core Detection
- [ ] Add SHA256 checksum computation to `dotfiles/manager.go`
  - [ ] Implement `computeFileHash(path string) (string, error)`
  - [ ] Implement `CompareFiles(path1, path2 string) (bool, error)`
  - [ ] Add unit tests for hash computation
  - [ ] Test edge cases (empty files, missing files, permissions)

- [ ] Update reconciliation logic in `resources/reconcile.go`
  - [ ] Modify `ReconcileItems` to check for drift
  - [ ] Use `StateDegraded` as `StateDrifted`
  - [ ] Add "drift_status" metadata
  - [ ] Update `GroupItemsByState` to handle drifted state

- [ ] Update dotfile resource in `dotfiles/resource.go`
  - [ ] Add comparison function to item metadata
  - [ ] Ensure `GetConfiguredDotfiles` includes compare_fn
  - [ ] Update `Apply` to handle `StateDegraded`

### Day 2: Status Integration
- [ ] Update status display
  - [ ] Add `Drifted()` function to `output/status.go`
  - [ ] Update `commands/status.go` to show drifted files
  - [ ] Add drift count to summary
  - [ ] Update table formatting for drift state

- [ ] Update types and constants
  - [ ] Update `resources/types.go` comments for StateDegraded
  - [ ] Add `StateDriftedStr = "drifted"` constant
  - [ ] Update `String()` method to return "drifted"

- [ ] Add integration tests
  - [ ] Test drift detection flow
  - [ ] Test status output with drifted files
  - [ ] Test apply with drifted files

### Day 3: Polish and Testing
- [ ] Documentation updates
  - [ ] Update `docs/cmds/status.md` with drift information
  - [ ] Update `docs/cmds/apply.md` with drift behavior
  - [ ] Add drift detection to `README.md` features

- [ ] Comprehensive testing
  - [ ] Manual test on macOS
  - [ ] Manual test on Linux
  - [ ] Test with various file types
  - [ ] Performance testing with many files

- [ ] Edge cases
  - [ ] Symlinks
  - [ ] Directories
  - [ ] Permission-denied files
  - [ ] Binary files

## Phase 2: Enhanced Features (Future)

### Diff Tool Integration
- [ ] Add `diff_tool` to config schema
- [ ] Implement internal diff display
- [ ] Add `--preview` flag to apply
- [ ] Execute external diff tools

### Additional Commands
- [ ] Add `plonk diff [file]` command
- [ ] Add `--show-diff` flag to status
- [ ] Add selective apply functionality

## Testing Checklist

### Unit Tests Required
1. `manager_test.go`
   - `TestComputeFileHash`
   - `TestCompareFiles`
   - `TestCompareFilesNotExist`

2. `reconcile_test.go`
   - `TestReconcileWithDrift`
   - `TestGroupItemsWithDrift`

3. `resource_test.go`
   - `TestDotfileResourceWithDrift`
   - `TestApplyDriftedFile`

### Integration Tests Required
1. `drift_test.go` (new file)
   - Full flow: add, deploy, modify, detect, apply
   - Multiple files with mixed states
   - Directory handling

### Manual Test Scenarios
1. **Basic Flow**
   ```bash
   echo "original" > ~/.vimrc
   plonk add ~/.vimrc
   echo "modified" > ~/.vimrc
   plonk status  # Should show drifted
   plonk apply   # Should restore original
   ```

2. **Multiple Files**
   - Some drifted, some not
   - Some missing, some drifted

3. **Edge Cases**
   - Large files (>10MB)
   - Binary files
   - Symlinks
   - No read permission

## Definition of Done

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual testing complete on macOS and Linux
- [ ] Documentation updated
- [ ] No performance regression (status < 200ms)
- [ ] Code reviewed and cleaned up
- [ ] Feature works as specified in plan
