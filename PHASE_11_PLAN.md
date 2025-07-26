# Phase 11: Command Consolidation

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase consolidates commands by removing redundant ones and improving the no-arguments behavior. We're removing the `ls` command (redundant with status), changing no-args behavior to show help, and adding a short alias for status.

## Objectives

1. Remove `plonk ls` command entirely
2. Change `plonk` (no args) to show help instead of status
3. Add `plonk st` as alias for `plonk status` command
4. Update all help text and documentation
5. Maintain separation of add/rm (dotfiles) and install/uninstall (packages)

## Current State

- `plonk ls` exists as a separate command showing packages/dotfiles
- `plonk` with no arguments shows status
- `plonk status` has no short alias
- Both add/rm and install/uninstall commands exist and should remain

## Implementation Details

### 1. Remove ls Command

**Files to delete:**
- `internal/commands/ls.go`

**Files to modify:**
- `internal/commands/root.go` - Remove ls command registration

### 2. Update Root Command Behavior

**In `internal/commands/root.go`:**
- Remove the `Run` function that currently shows status
- Let cobra's default behavior show help when no command is provided
- This happens automatically when a command has no `Run` function

Example:
```go
var rootCmd = &cobra.Command{
    Use:   "plonk",
    Short: "A developer environment manager",
    Long: `Plonk manages your development environment by installing packages
and managing dotfiles across multiple package managers.`,
    // Remove the Run: function entirely
}
```

### 3. Add Status Alias

**In `internal/commands/status.go`:**
```go
var statusCmd = &cobra.Command{
    Use:     "status",
    Aliases: []string{"st"},
    Short:   "Show the current state of packages and dotfiles",
    // ... rest of command definition
}
```

### 4. Update Help Text

Ensure all command help text reflects:
- No mention of `ls` command
- Clear indication that `status` can be invoked as `st`
- Root help text shows available commands

### 5. Verify Command Separation

Confirm these commands remain unchanged:
- `plonk add` / `plonk rm` - for dotfiles only
- `plonk install` / `plonk uninstall` - for packages only

## Testing Requirements

### Unit Tests
- Remove any tests for the ls command
- Update root command tests to expect help output
- Add test for `st` alias functionality

### Integration Tests
1. `plonk` (no args) shows help text, not status
2. `plonk ls` returns "unknown command" error
3. `plonk st` works identically to `plonk status`
4. `plonk status` continues to work as before
5. Help text no longer mentions ls command

### Manual Testing
- Run `plonk` and verify it shows help
- Run `plonk help` and verify ls is not listed
- Run `plonk st` and `plonk status` and verify identical output
- Verify add/rm and install/uninstall remain separate

## Expected Changes

1. **Deleted files:**
   - `internal/commands/ls.go`
   - Any test files specifically for ls command

2. **Modified files:**
   - `internal/commands/root.go` - Remove Run function and ls registration
   - `internal/commands/status.go` - Add alias
   - Integration test files that test ls or no-args behavior

3. **Behavior changes:**
   - `plonk` shows help instead of status
   - `plonk ls` no longer exists
   - `plonk st` works as alias for status

## Validation Checklist

Before marking complete:
- [ ] `plonk ls` command is completely removed
- [ ] `plonk` with no args shows help text
- [ ] `plonk st` works as alias for `plonk status`
- [ ] No references to ls command in help text
- [ ] All unit tests pass
- [ ] All integration tests updated and passing
- [ ] Add/rm remain dotfile-only commands
- [ ] Install/uninstall remain package-only commands

## Notes

- This is a breaking change - users of `plonk ls` will need to use `plonk status`
- The status command already shows both packages and dotfiles, making ls redundant
- Showing help by default (no args) is more standard CLI behavior
- The `st` alias follows common conventions (like `git st` for `git status`)

Remember to create `PHASE_11_COMPLETION.md` when finished!
