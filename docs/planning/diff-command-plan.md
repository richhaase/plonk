# Plonk Diff Command Implementation Plan

**Status**: 📝 Planning (2025-07-30)

## Overview
Add `plonk diff` command to show differences between source and deployed dotfiles that have drifted.

## User Experience

### Command Syntax
```bash
plonk diff              # Show diff for all drifted dotfiles
plonk diff ~/.vimrc     # Show diff for specific file
plonk diff vimrc        # Also accepts config name
```

### Default Behavior
- Uses `git diff --no-index` by default (zero-config)
- Shows differences between source (in $PLONK_DIR) and deployed (in $HOME)
- Only shows output for files that have actually drifted
- GUI diff tools open in their own windows

### Configuration
```yaml
# plonk.yaml
diff_tool: "delta"  # Optional, defaults to "git diff --no-index"
```

Common configurations:
- `diff_tool: "delta"`
- `diff_tool: "vimdiff"`
- `diff_tool: "code --diff --wait"`
- `diff_tool: "meld"`

## Design Decisions

1. **Simple argument handling**: Accept file path or config name
2. **Zero-config default**: Use `git diff --no-index` (git is already required)
3. **Simple execution**: `{tool} {source} {dest}` - no complex templating
4. **Direct output**: Stream diff tool output directly to user
5. **GUI support**: Let GUI tools open naturally, don't capture output
6. **No flags initially**: Keep it simple, can add later if needed

## Implementation Plan

### 1. Command Structure
Create `internal/commands/diff.go`:
```go
var diffCmd = &cobra.Command{
    Use:   "diff [file]",
    Short: "Show differences for drifted dotfiles",
    Long:  `Show differences between source and deployed dotfiles...`,
    Args:  cobra.MaximumNArgs(1),
    RunE:  runDiff,
}
```

### 2. Implementation Flow

1. **Parse arguments**:
   - No args: Find all drifted files
   - With arg: Resolve to specific dotfile

2. **Reconcile to find drifted files**:
   - Use existing reconciliation system
   - Filter for StateDegraded items

3. **For each drifted file**:
   - Get source path from $PLONK_DIR
   - Get deployed path from $HOME
   - Execute diff tool

4. **Execute diff tool**:
   - Load config for custom diff_tool
   - Default to "git diff --no-index"
   - Run: `exec.Command(tool, sourcePath, deployedPath)`
   - Stream output directly to stdout/stderr

### 3. Error Handling

- **No drifted files**: Message "No drifted dotfiles found"
- **File not found**: "Dotfile not found: X"
- **File not drifted**: "Dotfile is not drifted: X"
- **Diff tool not found**: Fall back to default with warning
- **Diff tool error**: Show error but continue with other files

### 4. File Resolution

Support multiple input formats:
- `~/.vimrc` → Resolve to vimrc in config
- `/home/user/.vimrc` → Resolve to vimrc in config
- `vimrc` → Direct config name
- `.vimrc` → Strip dot and use as config name

### 5. Configuration Loading

Add to `internal/config/config.go`:
```go
type Config struct {
    // ... existing fields
    DiffTool string `yaml:"diff_tool,omitempty"`
}
```

Default value handled in code if empty.

## Testing Strategy

### Unit Tests
1. File argument resolution (path → config name)
2. Configuration loading with defaults
3. Diff command selection logic

### Integration Tests
1. No drifted files scenario
2. Single drifted file
3. Multiple drifted files
4. Specific file request
5. Custom diff tool configuration

### Manual Testing
1. Test with various diff tools:
   - git diff --no-index (default)
   - delta
   - vimdiff
   - VS Code
   - meld
2. Test with binary files
3. Test with missing/invalid files
4. Test with no arguments vs specific file

## Success Criteria

1. ✅ `plonk diff` shows all drifted files
2. ✅ `plonk diff ~/.vimrc` shows specific file
3. ✅ Default to git diff without configuration
4. ✅ Respects configured diff_tool
5. ✅ GUI tools work naturally
6. ✅ Clear error messages
7. ✅ No output when no drift

## Future Enhancements (Not for v1)

1. `--all` flag to include missing files
2. `--name-only` flag to just list drifted files
3. `--tool` flag to override configured tool
4. Side-by-side diff in terminal
5. Colorized output options

## Implementation Order

1. Create diff command structure
2. Add file resolution logic
3. Integrate with reconciliation
4. Add diff tool execution
5. Add configuration support
6. Write tests
7. Update documentation

## Estimated Effort

- Command structure: 1 hour
- File resolution: 1 hour
- Reconciliation integration: 30 minutes
- Diff execution: 1 hour
- Configuration: 30 minutes
- Tests: 2 hours
- Documentation: 30 minutes

**Total: ~6 hours**
