# Dotfiles Simplification Implementation Plan

## Phase 1: Create New Structure (Green Path)

Build the new simplified module alongside the old one, then swap.

### Task 1.1: Create types.go
- [ ] Define `Dotfile` struct
- [ ] Define `State` enum (Managed, Missing, Drifted, Unmanaged)
- [ ] Define `DotfileStatus` struct
- [ ] Define `ApplyResult` struct

### Task 1.2: Create fs.go
- [ ] Define `FileSystem` interface (7 methods)
- [ ] Implement `OSFileSystem` struct
- [ ] Implement `MemoryFS` struct for testing

### Task 1.3: Create dotfiles.go
- [ ] Define `Manager` struct (configDir, homeDir, fs, ignore)
- [ ] Implement `New()` and `NewWithFS()` constructors
- [ ] Implement `toTarget()` helper (source → target path)
- [ ] Implement `toSource()` helper (target → source path)
- [ ] Implement `shouldIgnore()` helper
- [ ] Implement `List()` - enumerate dotfiles in $PLONK_DIR
- [ ] Implement `Add()` - copy from $HOME to $PLONK_DIR
- [ ] Implement `Remove()` - delete from $PLONK_DIR
- [ ] Implement `Deploy()` - copy from $PLONK_DIR to $HOME (atomic write)
- [ ] Implement `IsDrifted()` - compare source and target content
- [ ] Implement `Diff()` - return content difference

### Task 1.4: Create reconcile.go
- [ ] Implement `Reconcile()` - return status for all dotfiles
- [ ] Implement `Apply()` - deploy missing/drifted files

### Task 1.5: Create dotfiles_test.go
- [ ] Test `List()` with MemoryFS
- [ ] Test `toTarget()` / `toSource()` path conversion
- [ ] Test `shouldIgnore()` pattern matching
- [ ] Test `IsDrifted()` comparison
- [ ] Test `Reconcile()` state detection
- [ ] Test `Add()` / `Remove()` / `Deploy()` operations

## Phase 2: Update Command Layer

### Task 2.1: Update commands to use new Manager
- [ ] Update `add.go` to use new dotfiles.Manager
- [ ] Update `rm.go` to use new dotfiles.Manager
- [ ] Update `apply.go` to use new dotfiles.Manager
- [ ] Update `status.go` to use new dotfiles.Manager
- [ ] Update `diff.go` to use new dotfiles.Manager
- [ ] Update `dotfiles.go` command to use new Manager

### Task 2.2: Update orchestrator
- [ ] Update `coordinator.go` to use new Apply signature

## Phase 3: Delete Old Code

### Task 3.1: Delete template code
- [ ] Delete template.go, template_test.go
- [ ] Delete template_fileops.go, template_fileops_test.go
- [ ] Delete template_comparator.go, template_comparator_test.go
- [ ] Delete expander.go

### Task 3.2: Delete old abstractions
- [ ] Delete atomic.go, atomic_test.go
- [ ] Delete config_handler.go, config_handler_more_test.go
- [ ] Delete directory_scanner.go
- [ ] Delete file_comparator.go, file_comparator_more_test.go
- [ ] Delete fileops.go, fileops_test.go, fileops_backup_test.go
- [ ] Delete filter.go, filter_test.go
- [ ] Delete path_resolver.go
- [ ] Delete path_validator.go, path_validator_test.go
- [ ] Delete scanner.go, scanner_test.go
- [ ] Delete old manager.go and manager_*_test.go files
- [ ] Delete old reconcile.go, reconcile_items.go
- [ ] Delete old types.go
- [ ] Delete remaining test files

## Phase 4: Verify

### Task 4.1: Run tests
- [ ] `go test ./...` passes
- [ ] `go build ./...` passes
- [ ] `golangci-lint run` passes

### Task 4.2: Run BATS integration tests
- [ ] `bats tests/bats/behavioral/` passes

### Task 4.3: Manual verification
- [ ] `plonk status` shows dotfiles correctly
- [ ] `plonk add ~/.testfile` works
- [ ] `plonk rm testfile` works
- [ ] `plonk apply` deploys dotfiles
- [ ] `plonk diff` shows drifted content

## Execution Order

1. Phase 1 (create new) - Can be done without breaking existing code
2. Phase 2 (update commands) - Swap to new implementation
3. Phase 4.1 (verify tests) - Ensure nothing broke
4. Phase 3 (delete old) - Remove dead code
5. Phase 4.2-4.3 (final verify) - Full validation
