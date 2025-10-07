# Dotfile Management Commands

Commands for managing dotfiles: `add` and `rm`.

## Description

The dotfile management commands handle the addition and removal of configuration files in plonk. The `add` command copies files from their current locations (typically in `$HOME`) to the plonk configuration directory (`$PLONK_DIR`), while `rm` removes files from plonk management.

The dotfiles themselves are the source of truth - any file in `$PLONK_DIR` (excluding plonk.yaml and plonk.lock) is considered a managed dotfile. No separate tracking file is needed.

## File Mapping

Plonk uses automatic dot-prefix handling:
- `$HOME/.zshrc` ↔ `$PLONK_DIR/zshrc`
- `$HOME/.config/nvim/init.lua` ↔ `$PLONK_DIR/config/nvim/init.lua`

The leading dot is removed when storing in `$PLONK_DIR` and re-added when deploying to `$HOME`.

---

## Add Command

Copies dotfiles from their current location to `$PLONK_DIR` for management.

### Synopsis

```bash
plonk add [options] [files...]
```

### Options

- `--dry-run, -n` - Preview changes without adding files
- `--sync-drifted, -y` - Sync all drifted files from $HOME back to $PLONKDIR

### Behavior

- Accepts single files, multiple files, or entire directories
- Always copies files (never moves originals)
- Overwrites existing files in `$PLONK_DIR` if present
- Maintains directory structure
- Respects `ignore_patterns` from configuration

**Directory Handling:**
- Recursively walks directory tree
- Adds each file individually (leaf nodes)
- Skips directories themselves

**Path Resolution:**
- Accepts absolute paths: `/home/user/.vimrc`
- Accepts tilde paths: `~/.vimrc`
- Relative paths resolved from current directory

### Examples

```bash
# Add single file
plonk add ~/.vimrc

# Add multiple files
plonk add ~/.zshrc ~/.gitconfig ~/.tmux.conf

# Add entire directory
plonk add ~/.config/nvim/

# Sync all drifted files back to $PLONKDIR
plonk add -y

# Preview drift sync without making changes
plonk add -y --dry-run

# Preview changes
plonk add --dry-run ~/.vimrc
```

---

## Remove Command

Removes dotfiles from plonk management by deleting them from `$PLONK_DIR`.

### Synopsis

```bash
plonk rm [options] <file>...
```

### Options

- `--dry-run, -n` - Preview what would be removed

### Behavior

- Deletes files from `$PLONK_DIR` only
- Original files in `$HOME` are never touched
- Files are no longer managed after removal
- No backups created (use git in `$PLONK_DIR` for recovery)
- Accepts same path formats as `add` command

### Examples

```bash
# Remove single file
plonk rm ~/.vimrc

# Remove multiple files
plonk rm ~/.zshrc ~/.gitconfig

# Preview removal
plonk rm --dry-run ~/.vimrc
```

---

## Common Behaviors

### State Management

**Add Command:**
- Modifies: `$PLONK_DIR` filesystem (copies file)
- System changes: None (original file unchanged)
- Immediate effect: File available for deployment via `apply`

**Remove Command:**
- Modifies: `$PLONK_DIR` filesystem (deletes file)
- System changes: None (deployed file in `$HOME` unchanged)
- Immediate effect: File no longer managed or deployable

### Error Handling

**Add errors:**
- File not found: Reports error and continues with other files
- Permission denied: Reports error for that file
- Ignored file: Silently skips (not considered an error)

**Remove errors:**
- File not in `$PLONK_DIR`: Reports as already not managed
- Permission denied: Reports error for that file

### File System Structure

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

## Integration

- After `add`, run `plonk apply` to deploy dotfiles to `$HOME`
- Use `plonk status` to see managed dotfiles
- Use `plonk diff` to see changes in deployed dotfiles
- Files added are immediately reflected in status output
- Recommended to use git in `$PLONK_DIR` for version control

## Notes

- Add always overwrites without warning (assumes user intent)
- No deployment happens during add - use `plonk apply` to deploy
- The `.plonk/` directory is reserved for future plonk metadata
- Dotfiles within `$PLONK_DIR` (e.g., `.git`, `.gitignore`) are ignored
- Ignore patterns prevent files from being added but not removed
