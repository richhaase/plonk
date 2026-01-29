# Plonk CLI Reference

## Global Options

All commands support these options:

- `--help, -h` - Show help for any command
- `--version, -v` - Show plonk version

## Commands

Note: For detailed command behaviors and examples, see the individual command documentation in docs/cmds/

### plonk

Show help information for plonk commands.

```bash
plonk                                     # Show help and available commands
```

### plonk clone

Clone a dotfiles repository and intelligently set up plonk. Automatically detects required package managers from the lock file, reports which ones are missing, and guides you to install them manually.

```bash
plonk clone user/dotfiles             # Clone GitHub repository (shorthand)
plonk clone https://github.com/user/repo.git  # Clone with full URL
plonk clone --dry-run user/dotfiles   # Preview what would happen
```

Options:
- `--dry-run, -n` - Show what would be cloned without making changes

### plonk track

Track packages that are already installed on your system. Adds them to the lock file for management.

```bash
plonk track brew:ripgrep              # Track a single package
plonk track brew:fd cargo:bat         # Track multiple packages
plonk track go:gopls pnpm:typescript  # Track packages from different managers
```

**Note**: The `manager:package` format is always required. Packages must already be installed.

**📖 [Complete Package Management Documentation →](cmds/package-management.md)**

### plonk untrack

Stop tracking packages. Does NOT uninstall the package from your system.

```bash
plonk untrack brew:ripgrep            # Stop tracking a package
plonk untrack cargo:bat go:gopls      # Stop tracking multiple packages
```

**📖 [Complete Package Management Documentation →](cmds/package-management.md)**

### plonk add

Add dotfiles to management or sync drifted files back.

```bash
plonk add ~/.vimrc ~/.zshrc           # Add files
plonk add ~/.config/nvim/             # Add directory
plonk add -y                          # Sync all drifted files from $HOME to $PLONK_DIR
plonk add -y --dry-run                # Preview drift sync
plonk add --dry-run ~/.vimrc          # Preview add
```

### plonk rm

Remove dotfiles from management (files are not deleted).

```bash
plonk rm ~/.vimrc
plonk rm --dry-run ~/.vimrc
```

### plonk status

Show managed packages and dotfiles (combined view).

```bash
plonk status                          # Show all
plonk st                              # Short alias
```

**📖 [Complete Status Documentation →](cmds/status.md)**

Alias: `st`

**Note**: For a focused view of dotfiles only, use:
- `plonk dotfiles` (or `plonk d`) - Show only dotfile status

### plonk diff

Show differences for drifted dotfiles.

```bash
plonk diff                            # Show all drifted files
plonk diff ~/.zshrc                   # Show diff for specific file
plonk diff $HOME/.bashrc              # Supports environment variables
```

### plonk apply

Install missing tracked packages and deploy missing dotfiles.

```bash
plonk apply                           # Apply all changes
plonk apply ~/.vimrc ~/.zshrc         # Apply only specific dotfiles
plonk apply --dry-run                 # Preview changes
plonk apply --packages                # Apply packages only
plonk apply --dotfiles                # Apply dotfiles only
```

### plonk doctor

Check system health and configuration.

```bash
plonk doctor                          # Check system health
```

Note: Use `plonk clone` to detect which package managers your repo requires; install any missing managers using the hints from `plonk doctor`.

### plonk config

Manage plonk configuration.

```bash
plonk config show                     # Show current config with user-defined values highlighted
plonk config edit                     # Edit config in visudo-style (only saves non-defaults)
```

### plonk dotfiles

Show dotfile status (focused view).

```bash
plonk dotfiles                        # Show all managed dotfiles
plonk d                               # Short alias
```

**📖 [Complete Dotfiles Documentation →](cmds/dotfiles.md)**

Alias: `d`

### plonk completion

Generate shell completion scripts.

```bash
plonk completion bash                 # Bash completion
plonk completion zsh                  # Zsh completion
plonk completion fish                 # Fish completion
plonk completion powershell           # PowerShell completion
```

## Package Manager Prefixes

Use prefixes to specify package managers when tracking:

- `brew:` - Homebrew (macOS and Linux)
- `cargo:` - Cargo (Rust)
- `go:` - Go (Go binaries)
- `pnpm:` - PNPM (fast, disk efficient Node.js packages)
- `uv:` - UV (Python tool manager)

Examples:
```bash
plonk track brew:wget
plonk track cargo:ripgrep
plonk track go:gopls
plonk track pnpm:typescript
plonk track uv:ruff
```

## Output and Colors

Plonk uses minimal colorization for status indicators:
- Green: Success, managed, available states
- Red: Error, missing, failed states
- Yellow: Warning, unmanaged states
- Blue: Informational annotations

To disable colors, set the `NO_COLOR` environment variable:
```bash
NO_COLOR=1 plonk status
```

## Exit Codes

- `0` - Success
- `1` - General error or command failure
