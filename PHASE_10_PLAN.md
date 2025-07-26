# Phase 10: Sync to Apply Command Rename

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase renames the core `sync` command to `apply` throughout the codebase. This is a pervasive change that affects commands, functions, hooks, and documentation. Getting this done early prevents having to update new code later.

## Objectives

1. Rename `plonk sync` command to `plonk apply`
2. Update all internal function names from Sync* to Apply*
3. Update hook names from pre_sync/post_sync to pre_apply/post_apply
4. Remove all references to "sync" in user-facing text
5. Ensure `--dry-run` flag continues to work

## Current State

- Main reconciliation command: `plonk sync`
- Internal functions: SyncPackages, SyncDotfiles, orchestrator.Sync
- Config hooks: pre_sync, post_sync
- Many references to "sync" in help text and error messages

## Implementation Details

### 1. Command File Rename

**Files to rename:**
- `internal/commands/sync.go` → `internal/commands/apply.go`

### 2. Command Registration Update

**In `internal/commands/apply.go` (renamed from sync.go):**
```go
var applyCmd = &cobra.Command{
    Use:   "apply",
    Short: "Apply configuration to reconcile system state",
    Long:  `Apply reads your plonk configuration and reconciles the system state
to match, installing missing packages and managing dotfiles.`,
    // ... rest of command definition
}
```

### 3. Internal Function Renames

**Key renames needed:**
- `orchestrator.Sync()` → `orchestrator.Apply()`
- `orchestrator.SyncResult` → `orchestrator.ApplyResult`
- `SyncPackages()` → `ApplyPackages()`
- `SyncDotfiles()` → `ApplyDotfiles()`
- Any other Sync* functions → Apply*

**Note:** Use IDE refactoring tools if available to ensure all references are updated.

### 4. Hook Configuration Updates

**In config structures and YAML:**
- `pre_sync` → `pre_apply`
- `post_sync` → `post_apply`

**Backward compatibility:** For now, we're NOT maintaining backward compatibility. Users must update their configs.

Example config change:
```yaml
# Old
hooks:
  pre_sync:
    - command: "echo Starting sync"
  post_sync:
    - command: "echo Sync complete"

# New
hooks:
  pre_apply:
    - command: "echo Starting apply"
  post_apply:
    - command: "echo Apply complete"
```

### 5. Update Hook Runner

**In `internal/orchestrator/hooks.go` (or similar):**
- Update method names: `RunPreSync` → `RunPreApply`, etc.
- Update any hook-related constants or types

### 6. Update Error Messages and Help Text

Search for all occurrences of "sync" in:
- Error messages
- Help text
- Command descriptions
- Comments (where user-facing)

Common replacements:
- "sync" → "apply"
- "synchronize" → "apply"
- "synchronization" → "application"
- "syncing" → "applying"

### 7. Test Updates

**Integration tests will need updates:**
- Test function names
- Expected command outputs
- Config fixtures with hooks

**Unit tests:**
- Function names
- Test descriptions

## Testing Requirements

### Unit Tests
- Verify all renamed functions work correctly
- Ensure hook runner uses new names

### Integration Tests
1. `plonk apply` works with all flags
2. `plonk sync` returns "unknown command" error
3. Hooks with new names (pre_apply/post_apply) execute correctly
4. `--dry-run` flag still works
5. Help text shows "apply" not "sync"

### Manual Testing
- Run `plonk apply` with a real config
- Verify hooks execute
- Check all help text: `plonk help`, `plonk apply --help`
- Verify error messages use "apply" terminology

## Expected Changes

1. **Renamed files:**
   - `internal/commands/sync.go` → `internal/commands/apply.go`

2. **Modified files (extensive):**
   - `internal/orchestrator/orchestrator.go`
   - `internal/orchestrator/hooks.go`
   - `internal/commands/apply.go` (formerly sync.go)
   - Multiple test files
   - Any file referencing sync operations

3. **Config changes:**
   - Hook names in YAML structure
   - Example configs
   - Test fixtures

## Search and Replace Strategy

1. **First pass - exact matches:**
   - `"sync"` → `"apply"`
   - `'sync'` → `'apply'`
   - ` sync ` → ` apply `

2. **Second pass - function names:**
   - `Sync(` → `Apply(`
   - `sync(` → `apply(`
   - Variable names like `syncResult` → `applyResult`

3. **Third pass - hooks:**
   - `pre_sync` → `pre_apply`
   - `post_sync` → `post_apply`
   - `PreSync` → `PreApply`
   - `PostSync` → `PostApply`

4. **Manual review required:**
   - Comments mentioning sync
   - Documentation strings
   - Test descriptions

## Validation Checklist

Before marking complete:
- [ ] `plonk apply` command works
- [ ] `plonk sync` returns "unknown command"
- [ ] All internal Sync functions renamed to Apply
- [ ] Hooks use pre_apply/post_apply names
- [ ] No "sync" references in help text
- [ ] All tests updated and passing
- [ ] `--dry-run` flag works with apply command
- [ ] Error messages use "apply" terminology

## Notes

- This is a breaking change - no backward compatibility
- Be thorough - use grep to find all occurrences
- Don't forget to update test configs and fixtures
- Some third-party library calls might use "sync" - leave those alone
- Focus only on plonk's use of the term

Remember to create `PHASE_10_COMPLETION.md` when finished!
