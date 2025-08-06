# Diff Command

Shows differences between source and deployed dotfiles that have drifted.

## Synopsis

```bash
plonk diff [file]
```

## Description

The diff command identifies dotfiles whose deployed versions differ from their source versions in `$PLONK_DIR` and displays the differences using a configurable diff tool. By default, it uses `git diff --no-index` for a familiar, colorized output.

The command supports showing diffs for all drifted files or for a specific file. It's useful for understanding what changes have been made to deployed dotfiles before deciding whether to restore them with `plonk apply` or update the source with `plonk add`.

## Options

This command has no options.

## Behavior

### Core Operation

Diff performs the following steps:
1. **Reconciliation** - Identifies all drifted dotfiles by comparing SHA256 checksums
2. **Filtering** - Optionally filters to a specific file if argument provided
3. **Diff Display** - Executes the configured diff tool for each drifted file

### File Path Resolution

The diff command accepts various path formats for the file argument:
- `~/.zshrc` - Tilde expansion
- `$HOME/.zshrc` - Environment variable expansion
- `/Users/rdh/.zshrc` - Absolute paths
- `$CUSTOM_VAR` - Any environment variable containing a path
- `.zshrc` - Relative paths (resolved to current directory)

### Default Behavior

- Uses `git diff --no-index` by default (zero-config)
- Shows differences with source file first, deployed file second
- Processes all drifted files when no argument provided
- Continues with remaining files if individual diff fails
- GUI diff tools open in their own windows

### Diff Tool Configuration

Configure a custom diff tool in `plonk.yaml`:

```yaml
diff_tool: "delta"  # or any diff command
```

Common configurations:
- `diff_tool: "delta"` - Modern diff with syntax highlighting
- `diff_tool: "vimdiff"` - Vim's diff mode
- `diff_tool: "code --diff --wait"` - VS Code diff
- `diff_tool: "meld"` - GUI diff tool

The diff tool is executed as: `{tool} {source_path} {deployed_path}`

### Error Handling

- **No drifted files**: Displays "No drifted dotfiles found"
- **File not found**: Returns error if specified file doesn't exist or isn't drifted
- **Diff tool errors**: Reports error but continues with other files
- **Missing diff tool**: Falls back to default if configured tool not found

## Examples

```bash
# Show diff for all drifted dotfiles
plonk diff

# Show diff for specific file
plonk diff ~/.zshrc

# Show diff using environment variable
plonk diff $HOME/.gitconfig

# Show diff for relative path
plonk diff .vimrc
```

## Integration

- Use `plonk status` to see which files have drifted
- Use `plonk apply` to restore drifted files from source
- Use `plonk add` to update source with current deployed version

## Notes

- The diff tool output is streamed directly to the terminal
- Non-zero exit codes from diff tools are expected (files differ)
- File paths are normalized and expanded before comparison
