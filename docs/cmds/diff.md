# Diff Command

The `plonk diff` command shows differences between source and deployed dotfiles that have drifted.

## Description

The diff command identifies dotfiles whose deployed versions differ from their source versions in `$PLONK_DIR` and displays the differences using a configurable diff tool. By default, it uses `git diff --no-index` for a familiar, colorized output. The command supports showing diffs for all drifted files or for a specific file.

## Behavior

### Core Operation

Diff performs the following steps:
1. **Reconciliation** - Identifies all drifted dotfiles by comparing SHA256 checksums
2. **Filtering** - Optionally filters to a specific file if argument provided
3. **Diff Display** - Executes the configured diff tool for each drifted file

### Command Syntax

```bash
plonk diff              # Show diff for all drifted dotfiles
plonk diff [file]       # Show diff for specific file only
```

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

### Configuration

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

### Output

The diff command streams the output of the diff tool directly to the terminal. The exact format depends on the configured tool, but typically shows:
- File paths being compared
- Line-by-line differences
- Context around changes

### Error Handling

- **No drifted files**: Displays "No drifted dotfiles found"
- **File not found**: Returns error if specified file doesn't exist or isn't drifted
- **Diff tool errors**: Reports error but continues with other files
- **Missing diff tool**: Falls back to default if configured tool not found

### Integration with Other Commands

- Use `plonk status` to see which files have drifted
- Use `plonk apply` to restore drifted files from source
- Use `plonk add` to update source with current deployed version

## Implementation Notes

The diff command leverages plonk's reconciliation system to identify drift:

**Command Structure:**
- Entry point: `internal/commands/diff.go`
- Uses existing reconciliation to find drifted files
- Path normalization handles ~, $HOME, and environment variables

**Key Implementation Details:**

1. **Drift Detection:**
   - Reuses reconciliation system from status command
   - Filters for items with `StateDegraded` (drifted state)
   - Relies on SHA256 checksum comparison

2. **Path Resolution:**
   - `normalizePath()` handles all path formats
   - Expands environment variables with `os.ExpandEnv()`
   - Resolves relative paths to absolute
   - Cleans redundant path elements

3. **File Mapping:**
   - Source files in `$PLONK_DIR` have no leading dot
   - Deployed files in `$HOME` have leading dot
   - Metadata preserves mapping between source and destination

4. **Diff Execution:**
   - Splits tool command to handle flags (e.g., "git diff --no-index")
   - Streams output directly to stdout/stderr
   - Non-zero exit codes from diff tools are expected (files differ)

**Configuration Loading:**
- Uses `LoadWithDefaults()` for zero-config behavior
- `DiffTool` field in Config struct
- Empty value defaults to "git diff --no-index"

## Improvements

- Add `--name-only` flag to list drifted files without showing diffs
- Add `--tool` flag to override configured diff tool
- Support diff options pass-through (e.g., `--word-diff`)
- Add support for three-way diff when drift history is available
