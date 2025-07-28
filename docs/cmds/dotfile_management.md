# Dotfile Management Commands

Commands for managing dotfiles: `add` and `rm`.

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

### Add Command

Copies dotfiles from their current location to `$PLONK_DIR` for management. Accepts single files, multiple files, or entire directories.

**Command Options**:
- `--dry-run, -n` - Preview what would be added without making changes
- `--force, -f` - (Currently non-functional, should be removed)

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
- `--force, -f` - (Currently non-functional, should be removed)

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

## Implementation Notes

## Improvements

- Remove non-functional --force flag from both commands
- Update add command help to remove "preserve original files" language
- Improve path resolution documentation in help text
- Consider adding --backup flag for rm command
- Add verbose output option to show ignore pattern matches
- Consider warning when re-adding files that differ from current version
