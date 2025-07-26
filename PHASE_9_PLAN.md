# Phase 9: Config Command Updates

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This is a foundational phase that adds auto-validation to all config loading operations. This will benefit all subsequent phases by ensuring configs are always valid when loaded.

## Objectives

1. Add automatic validation whenever config is loaded
2. Remove the explicit `plonk config validate` command
3. Update `plonk config show` to display the file path
4. Ensure clear, actionable error messages on validation failures

## Current State

- Config validation exists as a separate command: `plonk config validate`
- Config loading doesn't automatically validate
- Validation errors may not be user-friendly
- `plonk config show` doesn't display the config file path

## Implementation Details

### 1. Auto-Validation Implementation

**Files to modify:**
- `internal/config/config.go` - Add validation to Load functions
- `internal/config/compat.go` - Ensure LoadWithDefaults validates

**Key changes:**
```go
// In LoadConfig or similar functions, after loading:
if err := ValidateConfig(cfg); err != nil {
    return nil, fmt.Errorf("invalid configuration: %w", err)
}
```

### 2. Remove Validate Command

**Files to delete/modify:**
- `internal/commands/config_validate.go` - Delete this file
- `internal/commands/config.go` - Remove validate subcommand registration

### 3. Update Config Show

**Files to modify:**
- `internal/commands/config_show.go`

**Changes:**
- Add config file path to the output (before the config content)
- For table format: Add as a header line
- For JSON/YAML: Consider adding as a comment or metadata field

Example output:
```
Config file: /home/user/.config/plonk/plonk.yaml

default_manager: homebrew
packages:
  homebrew:
    - ripgrep
    - jq
```

### 4. Improve Error Messages

When validation fails, provide clear guidance:

**Bad:** `validation failed: invalid config`

**Good:** `invalid configuration: unknown package manager "hombrew" at line 3. Valid managers are: homebrew, npm, pip, cargo, go, gem`

### 5. Update Error Handling

Ensure all commands that load config handle validation errors gracefully:
- Show the validation error
- Suggest running `plonk config edit` to fix
- Exit with appropriate error code

## Testing Requirements

### Unit Tests
- Test validation happens on config load
- Test various validation error scenarios
- Test error message clarity

### Integration Tests
1. Test that invalid configs prevent commands from running
2. Test `plonk config show` displays path
3. Test removal of `plonk config validate` command

**Note:** When creating test configs with intentional errors, use minimal examples that clearly show the issue.

## Expected Changes

1. **Deleted files:**
   - `internal/commands/config_validate.go`

2. **Modified files:**
   - `internal/config/config.go`
   - `internal/config/compat.go` (if it exists)
   - `internal/commands/config.go`
   - `internal/commands/config_show.go`
   - Various test files

3. **Behavior changes:**
   - Any command that loads config will now fail fast on invalid config
   - `plonk config validate` command no longer exists
   - `plonk config show` displays file path

## Validation Checklist

Before marking complete:
- [ ] Auto-validation works for all config loading paths
- [ ] `plonk config validate` command is removed
- [ ] `plonk config show` displays file path
- [ ] Error messages are clear and actionable
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] No references to `config validate` remain in help text or docs

## Notes

- This is foundational work - other phases depend on this
- Focus on clear error messages that help users fix problems
- Don't add new validation rules, just integrate existing validation
- Keep changes focused - don't refactor unrelated code

Remember to create `PHASE_9_COMPLETION.md` when finished!
