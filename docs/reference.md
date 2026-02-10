# Plonk Reference

Complete CLI and configuration reference.

## Migration Notes (v0.25+)

- `install`, `uninstall`, and `upgrade` commands were removed.
- Package operations are now centered on `track`, `untrack`, and `apply`.
- Supported package managers: `brew`, `cargo`, `go`, `pnpm`, `uv`.
- Lock files are `version: 3` and older v2 lock files are auto-migrated.

## Commands

### plonk track

Track packages that are already installed.

```bash
plonk track <manager:package>...
```

- Verifies packages are installed before tracking
- Adds to `plonk.lock`
- Format `manager:package` is required (no default manager)

```bash
plonk track brew:ripgrep cargo:bat go:gopls
```

### plonk untrack

Stop tracking packages (does not uninstall).

```bash
plonk untrack <manager:package>...
```

```bash
plonk untrack brew:ripgrep
```

### plonk add

Add dotfiles to management.

```bash
plonk add <file>...
plonk add -y              # Sync all drifted files back to $PLONK_DIR
plonk add --dry-run       # Preview
```

Copies files from `$HOME` to `$PLONK_DIR`, stripping the dot prefix.

### plonk rm

Remove dotfiles from management (does not delete deployed files).

```bash
plonk rm <file>...
plonk rm --dry-run
```

### plonk apply

Install missing packages and deploy missing/drifted dotfiles.

```bash
plonk apply [options] [files...]
```

**Options:**
- `--dry-run, -n` - Preview changes
- `--packages` - Packages only
- `--dotfiles` - Dotfiles only

```bash
plonk apply                    # Everything
plonk apply --packages         # Packages only
plonk apply ~/.vimrc           # Specific dotfile
```

### plonk status

Show managed packages and dotfiles.

```bash
plonk status
plonk st                       # Alias
```

**States:**
- `managed` - Tracked and present
- `missing` - Tracked but not present
- `drifted` - Dotfile modified since deployment

### plonk dotfiles

Show dotfile status only.

```bash
plonk dotfiles
plonk d                        # Alias
```

### plonk diff

Show differences for drifted dotfiles.

```bash
plonk diff                     # All drifted
plonk diff ~/.zshrc            # Specific file
```

Uses `git diff` by default, or `diff_tool` from config.

### plonk clone

Clone a dotfiles repository and apply.

```bash
plonk clone <repo>
plonk clone user/dotfiles              # GitHub shorthand
plonk clone https://github.com/u/r.git # Full URL
plonk clone --dry-run user/dotfiles    # Preview
```

### plonk doctor

Check system health.

```bash
plonk doctor
```

Reports: config directory, permissions, package manager availability.

### plonk config

View and edit configuration.

```bash
plonk config show              # View current config
plonk config show -o json      # JSON output
plonk config edit              # Edit in $EDITOR
```

### plonk completion

Generate shell completions.

```bash
plonk completion bash
plonk completion zsh
plonk completion fish
```

## Package Managers

| Manager | Prefix | Install Command |
|---------|--------|-----------------|
| Homebrew | `brew:` | `brew install <pkg>` |
| Cargo | `cargo:` | `cargo install <pkg>` |
| Go | `go:` | `go install <pkg>@latest` |
| PNPM | `pnpm:` | `pnpm add -g <pkg>` |
| UV | `uv:` | `uv tool install <pkg>` |

## Configuration

Configuration file: `~/.config/plonk/plonk.yaml`

All settings are optional. Plonk uses sensible defaults.

### Settings

```yaml
# Package manager default (for discovery, not tracking)
default_manager: brew

# Timeouts (seconds)
package_timeout: 180       # Package installation
operation_timeout: 300     # General operations
dotfile_timeout: 60        # File operations

# Diff tool for viewing drifted files
diff_tool: delta           # Default: git diff --no-index

# Directories to scan for dotfiles
expand_directories:
  - .config                # Default

# Files to ignore
ignore_patterns:
  - "*.swp"
  - "*.tmp"
  - ".DS_Store"
  - ".git/*"
```

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `PLONK_DIR` | Config directory (default: `~/.config/plonk`) |
| `VISUAL` | Editor for `config edit` |
| `EDITOR` | Fallback editor |
| `NO_COLOR` | Disable colored output |

### Precedence

1. Command-line flags
2. Environment variables
3. `plonk.yaml`
4. Built-in defaults

## Lock File

`plonk.lock` is auto-managed. Format:

```yaml
version: 3
packages:
  brew:
    - fd
    - ripgrep
  cargo:
    - bat
  go:
    - golang.org/x/tools/gopls
```

## Exit Codes

- `0` - Success
- `1` - Error

## Output Formats

Commands support `--output` / `-o`:
- `table` (default)
- `json`
- `yaml`
