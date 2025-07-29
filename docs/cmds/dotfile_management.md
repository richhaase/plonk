# Dotfile Management Commands

Commands for managing dotfiles: `add` and `rm`.

For CLI syntax and flags, see [CLI Reference](../cli.md#plonk-add).

## Description

The dotfile management commands handle the addition and removal of configuration files in plonk. The `add` command copies files from their current locations (typically in `$HOME`) to the plonk configuration directory (`$PLONK_DIR`), while `rm` removes files from plonk management by deleting them from `$PLONK_DIR`. The dotfiles themselves are the source of truth - any file in `$PLONK_DIR` (excluding plonk.yaml and plonk.lock) is considered a managed dotfile.

## Behavior

### Core Concepts

**File Mapping**:
- `$HOME/.zshrc` ↔ `$PLONK_DIR/zshrc`
- `$HOME/.config/nvim/init.lua` ↔ `$PLONK_DIR/config/nvim/init.lua`

The leading dot is removed when storing in `$PLONK_DIR` and re-added when deploying to `$HOME`.

**State Management**:
- Dotfile state is determined by the contents of `$PLONK_DIR`
- No separate tracking file - the filesystem IS the state
- Files in `$PLONK_DIR` (except plonk.yaml/plonk.lock) are managed dotfiles
- Dotfiles within `$PLONK_DIR` (e.g., `.git`, `.gitignore`) are ignored to prevent incorrect deployment

**File System Example**:
```
$PLONK_DIR/                          $HOME/
├── zshrc                     →      ├── .zshrc
├── gitconfig                 →      ├── .gitconfig
├── config/                   →      └── .config/
│   └── nvim/                 →          └── nvim/
│       └── init.lua          →              └── init.lua
├── plonk.yaml                       (not deployed)
└── plonk.lock                       (not deployed)
```

### Add Command

Copies dotfiles from their current location to `$PLONK_DIR` for management. Accepts single files, multiple files, or entire directories.


**Directory Handling**:
- Recursively walks directory tree
- Adds each file individually (leaf nodes)
- Maintains directory structure in `$PLONK_DIR`
- Respects `ignore_patterns` from configuration

**Path Resolution**:
- Accepts absolute paths: `/home/user/.vimrc`
- Accepts tilde paths: `~/.vimrc`
- Relative paths are resolved from current directory

**Add Behavior**:
- Always copies file to `$PLONK_DIR` (never moves)
- Overwrites existing file in `$PLONK_DIR` if present (re-add)
- Original file in `$HOME` remains unchanged
- No symlinks are created or used

### Remove Command

Removes dotfiles from plonk management by deleting them from `$PLONK_DIR`. Accepts single or multiple file paths.

**Command Options**:
- `--dry-run, -n` - Preview what would be removed without making changes

**Remove Behavior**:
- Deletes file from `$PLONK_DIR` only
- Original file in `$HOME` is never touched
- File is no longer managed by plonk after removal
- No backups are created (relies on user's git repo)

### Special Behaviors

- Add always overwrites without warning (assumes user knows what they're doing)
- Recommended to use git in `$PLONK_DIR` for version control and recovery
- No deployment happens during add - use `plonk apply` to deploy
- Ignore patterns prevent files from being added but not removed
- Both commands work with the actual filesystem as state

### Error Handling

**Add errors**:
- File not found: Reports error and continues with other files
- Permission denied: Reports error for that file
- Ignored file: Silently skips (not considered an error)

**Remove errors**:
- File not in `$PLONK_DIR`: Reports as already not managed
- Permission denied: Reports error for that file

### Integration with Other Commands

- After `add`, run `plonk apply` to deploy dotfiles to `$HOME`
- Use `plonk status` to see managed dotfiles
- Files added are immediately reflected in status output
- No lock file updates needed (filesystem is the state)

### State Impact

**Add Command**:
- Modifies: `$PLONK_DIR` filesystem (copies file)
- System changes: None (original file unchanged)
- Immediate effect: File available for deployment via `apply`

**Remove Command**:
- Modifies: `$PLONK_DIR` filesystem (deletes file)
- System changes: None (deployed file in `$HOME` unchanged)
- Immediate effect: File no longer managed or deployable

Both commands directly modify the filesystem which IS the state tracking mechanism for dotfiles.

## Implementation Notes

The dotfile management commands provide filesystem-based state management through a comprehensive file operations system:

**Command Structure:**
- Entry points: `internal/commands/add.go`, `internal/commands/rm.go`
- Core logic: `internal/resources/dotfiles/manager.go`
- File operations: `internal/resources/dotfiles/fileops.go`
- Path utilities: `internal/resources/dotfiles/scanner.go`, `internal/resources/dotfiles/filter.go`

**Key Implementation Flow:**

1. **Add Command Processing:**
   - Entry point parses flags and validates arguments
   - Uses `dotfiles.Manager.AddFiles()` for batch processing
   - Each path resolved through `ResolveDotfilePath()` with fallback logic
   - Supports both single files and recursive directory processing

2. **Path Resolution Logic:**
   - Tilde paths (`~/file`) → `$HOME/file`
   - Absolute paths used as-is
   - Relative paths: tries current directory first, then `$HOME`
   - All paths validated to be within `$HOME` boundary
   - All paths validated to be within `$HOME` boundary

3. **File Mapping System:**
   - Uses `TargetToSource()` and `SourceToTarget()` functions
   - Target `~/.zshrc` → Source `zshrc` (removes `~/.` prefix)
   - Source `config/nvim/init.lua` → Target `~/.config/nvim/init.lua` (adds `~/.` prefix)
   - Maintains directory structure in both directions

4. **Add Operation Flow:**
   - Single files: Direct processing via `AddSingleFile()`
   - Directories: Recursive walk via `AddDirectoryFiles()` with ignore pattern filtering
   - Always copies to `$PLONK_DIR` (never moves original)
   - Creates parent directories as needed with 0750 permissions
   - Uses `CopyFileWithAttributes()` for attribute preservation

5. **Remove Command Processing:**
   - Entry point uses `ParseSimpleFlags()` for flag parsing
   - Uses `dotfiles.Manager.RemoveFiles()` for batch processing
   - Only removes files from `$PLONK_DIR`, never touches `$HOME`

6. **State Management:**
   - No lock file involvement - filesystem IS the state
   - `GetConfiguredDotfiles()` scans `$PLONK_DIR` to determine managed files
   - Excludes `plonk.yaml` and `plonk.lock` from being treated as dotfiles
   - Uses `ExpandConfigDirectory()` to walk and categorize files

7. **Error Handling:**
   - Individual file failures don't stop batch operations
   - Returns `resources.OperationResult` arrays for detailed per-file status
   - `ValidateOperationResults()` determines overall command exit status
   - File not found, permission errors, and ignore pattern skips are differentiated

8. **Directory Processing:**
   - Add: Uses `ExpandDirectoryPaths()` to find all files recursively
   - Respects `ignore_patterns` from configuration during directory walks
   - Remove: Not supported for directories (only individual file paths)
   - Skips directories themselves, only processes leaf files

**Architecture Patterns:**
- Manager pattern centralizes path resolution and validation
- Atomic file operations for reliable copying
- Comprehensive path validation with security checks
- Graceful error handling with detailed result tracking

**Error Conditions:**
- Path resolution failures result in operation failure
- Permission denied errors are captured per-file
- Files outside `$HOME` boundary are rejected
- Ignore pattern matches are silently skipped (not errors)

**Integration Points:**
- Both commands use `config.LoadWithDefaults()` for consistent zero-config behavior
- Results compatible with generic `resources.OperationResult` system
- File operations support backup creation for apply command

**Bugs Identified:**
None - all discrepancies have been resolved.

## Improvements

- Improve path resolution documentation in help text
- Consider adding --backup flag for rm command
- Add verbose output option to show ignore pattern matches
- Consider warning when re-adding files that differ from current version
