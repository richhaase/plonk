# CLI Reference

Complete command-line interface reference for plonk. All commands support structured output formats for AI agents.

## Global Options

```bash
--output, -o    Output format: table|json|yaml (default: table)
--version, -v   Show version information
--help, -h      Show help for any command
```

## Commands Overview

| Command | Purpose | AI Usage |
|---------|---------|----------|
| `init` | Create config template | Initial setup |
| `status` | Show system state | State reconciliation analysis |
| `add` | Add packages/dotfiles | Intelligent addition |
| `rm` | Remove packages/dotfiles | Intelligent removal |
| `ls` | List managed items | Smart overview |
| `sync` | Apply configuration | Automated deployment |
| `install` | Install packages | One-command package installation |
| `link` | Link dotfiles explicitly | Dotfile deployment |
| `unlink` | Unlink dotfiles explicitly | Dotfile removal |
| `dotfiles` | List dotfiles specifically | Dotfile overview |
| `env` | Environment info | Debugging context |
| `doctor` | Health checks | System validation |
| `search` | Find packages | Package discovery |
| `info` | Package details | Package analysis |
| `config` | Configuration management | Config operations |

## Core Commands

### `plonk init`

Create a configuration file template with all available options and helpful comments.

**Usage:**
```bash
plonk init [--force]
```

**Options:**
- `--force` - Overwrite existing configuration file

**Behavior:**
- Creates `~/.config/plonk/plonk.yaml` with all default values
- Includes helpful comments explaining each option
- Shows current default values that plonk uses
- Configuration is optional - plonk works without any config file

**Generated file includes:**
- All timeout settings with explanations
- Default package manager selection
- Directory expansion settings for dotfile listing
- Ignore patterns for dotfile discovery
- Extensive comments explaining each option

**Example:**
```bash
plonk init           # Create config template
plonk config show    # View effective configuration
plonk config edit    # Edit configuration
```

### `plonk status`

Display overall system state across all domains (equivalent to `plonk` with no arguments).

**Usage:**
```bash
plonk status [--output format]
plonk              # Zero-argument status (like git)
```

**Configuration File Status:**
- `ℹ️ using defaults` - No configuration file (zero-config mode)
- `✅ valid` - Configuration file exists and is valid
- `❌ invalid` - Configuration file exists but has syntax/validation errors

**Lock File Status:**
- `ℹ️ using defaults` - No lock file (no packages managed yet)
- `✅ exists` - Lock file exists with managed packages

**Item States:**
- `Managed` - In configuration AND present on system
- `Missing` - In configuration BUT NOT present on system
- `Untracked` - Present on system BUT NOT in configuration

**JSON output structure:**
```json
{
  "config_path": "/Users/user/.config/plonk/plonk.yaml",
  "lock_path": "/Users/user/.config/plonk/plonk.lock",
  "config_exists": false,
  "config_valid": false,
  "lock_exists": false,
  "state_summary": {
    "total_managed": 5,
    "total_missing": 1,
    "total_untracked": 23,
    "results": [
      {
        "domain": "package",
        "manager": "homebrew",
        "managed": [{"name": "git", "state": "managed"}],
        "missing": [{"name": "curl", "state": "missing"}],
        "untracked": [{"name": "vim", "state": "untracked"}]
      },
      {
        "domain": "dotfile",
        "managed": [{"name": "zshrc", "state": "managed"}],
        "missing": [],
        "untracked": [{"name": "vimrc", "state": "untracked"}]
      }
    ]
  }
}
```

### `plonk ls [filter]`

Smart overview of managed items with filtering options.

**Usage:**
```bash
plonk ls [--packages] [--dotfiles] [--manager manager] [--verbose] [--output format]
```

**Filtering Options:**
- `--packages` - Show packages only
- `--dotfiles` - Show dotfiles only
- `--manager` - Filter by package manager (homebrew, npm, cargo)
- `--verbose` - Show all items including untracked

**Behavior:**
- By default shows managed + missing items with untracked count
- Use `--verbose` to see all items including full untracked list
- Sorts by state (managed, missing, untracked), then alphabetically

**Example output:**
```bash
# Default overview
$ plonk ls
Overview: 43 total | ✓ 25 managed | ⚠ 3 missing | ? 15 untracked

Packages (25):
  Status Package                        Manager    Version
  ------ ------------------------------ ---------- --------
  ✓      git                            homebrew   2.43.0
  ✓      neovim                         homebrew   0.9.5
  ⚠      htop                           homebrew   -
  ✓      typescript                     npm        5.3.3

Dotfiles (18):
  Status Target                         Source
  ------ ------------------------------ --------------
  ✓      ~/.zshrc                       zshrc
  ✓      ~/.config/nvim/init.lua        config/nvim/init.lua
  ⚠      ~/.vimrc                       vimrc

15 untracked items (use --verbose to show details)

# Packages only
$ plonk ls --packages
Package Summary: 30 total | ✓ 25 managed | ⚠ 3 missing | ? 2 untracked

# Specific manager
$ plonk ls --manager homebrew
Homebrew packages: 20 total | ✓ 18 managed | ⚠ 2 missing
```

## Workflow Commands

### `plonk sync`

Apply all pending changes from your plonk configuration to your system (replaces `apply`).

**Usage:**
```bash
plonk sync [--dry-run] [--backup] [--packages] [--dotfiles]
```

**Options:**
- `--dry-run, -n` - Show changes without applying
- `--backup` - Create backups before overwriting dotfiles
- `--packages` - Sync packages only (mutually exclusive with --dotfiles)
- `--dotfiles` - Sync dotfiles only (mutually exclusive with --packages)

**Behavior:**
- Installs packages marked as missing in the lock file
- Deploys dotfiles from the configuration directory to their target locations
- Processes both missing dotfiles (new deployments) and managed dotfiles (updates)

**Example output:**
```
Plonk Sync
==========

📦📄 Syncing packages and dotfiles

📦 Packages: 2 installed, 0 failed
📄 Dotfiles: 3 deployed, 1 skipped

Summary: All changes applied successfully
```

### `plonk install <packages...>`

Install packages on your system and add them to plonk management.

**Usage:**
```bash
plonk install <packages...> [--brew] [--npm] [--cargo] [--dry-run] [--force]
```

**Options:**
- `--brew` - Use Homebrew package manager
- `--npm` - Use NPM package manager
- `--cargo` - Use Cargo package manager
- `--dry-run, -n` - Show what would be installed without making changes
- `--force, -f` - Force installation even if already managed

**Behavior:**
- Installs packages using the specified package manager
- Adds packages to the lock file for management
- Continues processing all packages even if some fail
- Uses the default manager from config or Homebrew as fallback

**Examples:**
```bash
plonk install htop                      # Install htop using default manager
plonk install git neovim ripgrep        # Install multiple packages
plonk install git --brew                # Install git specifically with Homebrew
plonk install lodash --npm              # Install lodash with npm global packages
plonk install ripgrep --cargo           # Install ripgrep with cargo packages
plonk install --dry-run htop neovim     # Preview what would be installed
```

### `plonk uninstall <packages...>`

Uninstall packages from your system and remove them from plonk management.

**Usage:**
```bash
plonk uninstall <packages...> [--brew] [--npm] [--cargo] [--dry-run] [--force]
```

**Options:**
- `--brew` - Use Homebrew package manager
- `--npm` - Use NPM package manager
- `--cargo` - Use Cargo package manager
- `--dry-run, -n` - Show what would be removed without making changes
- `--force, -f` - Force removal even if not managed

**Behavior:**
- Uninstalls packages from the system using the appropriate package manager
- Removes packages from the lock file
- Continues processing all packages even if some fail
- Automatically detects which manager manages each package

**Examples:**
```bash
plonk uninstall htop                    # Uninstall htop and remove from lock file
plonk uninstall git neovim              # Uninstall multiple packages
plonk uninstall --dry-run htop          # Preview what would be uninstalled
```

### `plonk dotfiles`

List dotfiles specifically with enhanced detail.

**Usage:**
```bash
plonk dotfiles [--verbose] [--output format]
```

**Behavior:**
- Same as `plonk ls --dotfiles` but with dotfile-specific formatting
- Shows source → target mappings clearly
- Includes deployment status and last modified times

## Package Discovery

### `plonk search <package>`

Search for packages across available package managers.

**Usage:**
```bash
plonk search <package> [--output format]
```

**Search behavior:**
1. If installed: Show managing package manager
2. If not installed + in default manager: Show default manager results
3. If not installed + not in default: Show all managers with package
4. If no default manager: Show all managers with package

**JSON output:**
```json
{
  "query": "git",
  "installed": true,
  "manager": "homebrew",
  "results": [
    {
      "manager": "homebrew",
      "name": "git",
      "version": "2.42.0",
      "description": "Distributed revision control system"
    }
  ]
}
```

### `plonk info <package>`

Show detailed package information.

**Usage:**
```bash
plonk info <package> [--output format]
```

**Information includes:**
- Name and version
- Description and homepage
- Installation status
- Dependencies (when available)
- Size information

## System Commands

### `plonk env`

Show environment information for debugging.

**Usage:**
```bash
plonk env [--output format]
```

**Output includes:**
- System information (OS, architecture)
- Package manager availability
- Configuration paths
- Environment variables

### `plonk doctor`

Comprehensive health check with actionable diagnostics.

**Usage:**
```bash
plonk doctor [--output format]
```

**Check categories:**
- System requirements
- Environment variables
- File permissions
- Configuration validation
- Package manager functionality

**JSON output:**
```json
{
  "status": "healthy|warning|error",
  "checks": [
    {
      "category": "system",
      "name": "go_version",
      "status": "pass|fail|warn",
      "message": "Go 1.24.4 found",
      "suggestion": "Update to latest version"
    }
  ]
}
```

## Configuration Management

### `plonk config show`

Display effective configuration (defaults merged with user settings).

Shows the complete configuration that plonk is actually using, including all default values merged with any user-specified overrides from the config file. This provides a comprehensive view of the active configuration regardless of whether a config file exists or contains only partial settings.

**Usage:**
```bash
plonk config show [--output format]
```

**Examples:**
```bash
plonk config show                 # Show effective config as YAML
plonk config show --output json   # Show as JSON
plonk config show --output yaml   # Show as YAML (default)
```

### `plonk config validate`

Validate configuration syntax and structure.

**Usage:**
```bash
plonk config validate [--output format]
```

**Exit codes:**
- `0` - Valid configuration
- `1` - Syntax error
- `2` - Validation error

### `plonk config edit`

Edit configuration file using `$EDITOR`.

**Usage:**
```bash
plonk config edit
```

## CLI 2.0 Migration from Legacy Commands

The new CLI provides significant typing reduction and improved ergonomics:

| Legacy Command | New Command | Typing Reduction |
|----------------|-------------|------------------|
| `plonk pkg add htop` | `plonk add htop` | 33% fewer characters |
| `plonk dot add ~/.vimrc` | `plonk add ~/.vimrc` | 25% fewer characters |
| `plonk pkg list` | `plonk ls --packages` | Similar length but more flexible |
| `plonk dot list` | `plonk ls --dotfiles` | Similar length but more flexible |
| `plonk apply` | `plonk sync` | 17% fewer characters |
| `plonk pkg add htop && plonk apply` | `plonk install htop` | 60% fewer characters |

**Key improvements:**
- **Intelligent detection**: No need to specify pkg/dot - plonk detects automatically
- **Mixed operations**: Add packages and dotfiles in single command
- **Unix-style**: Familiar commands like `ls`, `rm`, `add`
- **Workflow shortcuts**: `install` installs packages and adds to management
- **Zero-argument status**: Just type `plonk` for status (like git)

## Error Handling

### Exit Codes
- `0` - Success
- `1` - User input or configuration error
- `2` - System error (permissions, package manager unavailable)

### Structured Error System

Plonk uses a structured error system that provides:
- **Consistent error codes** across all commands
- **Domain-based categorization** for better error handling
- **User-friendly messages** with actionable guidance
- **Technical details** available in debug mode

### Error Categories

| Code | Domain | Description |
|------|---------|-------------|
| `CONFIG_NOT_FOUND` | config | Configuration file missing or unreadable |
| `CONFIG_PARSE_FAILURE` | config | Configuration syntax errors |
| `CONFIG_VALIDATION` | config | Configuration validation failures |
| `INVALID_INPUT` | commands | Invalid user input or arguments |
| `PACKAGE_INSTALL` | packages | Package installation failures |
| `MANAGER_UNAVAILABLE` | packages | Package manager not available |
| `FILE_IO` | dotfiles | File operation errors |
| `FILE_PERMISSION` | dotfiles | Permission-related file errors |
| `RECONCILIATION` | state | State reconciliation failures |

### Error Output Format

**Table format (default):**
```
Error: Package 'nonexistent-package' not found in any package manager

Troubleshooting steps:
  1. Check if the package name is correct: plonk search nonexistent-package
  2. Try updating package manager: brew update (for Homebrew)
  3. Check network connectivity

If the problem persists, run: plonk doctor
```

**JSON format:**
```json
{
  "error": {
    "code": "PACKAGE_INSTALL",
    "domain": "packages",
    "operation": "install",
    "message": "Package 'nonexistent-package' not found in any package manager",
    "item": "nonexistent-package",
    "severity": "error",
    "user_message": "Package 'nonexistent-package' not found in any package manager\n\nTroubleshooting steps:\n  1. Check if the package name is correct: plonk search nonexistent-package\n  2. Try updating package manager: brew update (for Homebrew)\n  3. Check network connectivity\n\nIf the problem persists, run: plonk doctor"
  }
}
```

### Debug Mode

Enable detailed error information:
```bash
export PLONK_DEBUG=1
plonk install nonexistent-package  # Shows technical details for errors
```

### Common Error Solutions

**Configuration not found:**
```bash
# Create initial configuration
plonk config edit
```

**Package manager unavailable:**
```bash
# Check system health
plonk doctor

# Install Homebrew (macOS)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Node.js/NPM
brew install node
```

**Permission errors:**
```bash
# Check permissions
ls -la ~/.config/plonk/

# Fix permissions
chmod 750 ~/.config/plonk/
chmod 600 ~/.config/plonk/plonk.yaml
```

## AI Agent Integration

### Structured Output
All commands support `--output json` for machine parsing:

```bash
plonk status --output json | jq '.packages.homebrew[] | select(.state == "missing")'
```

### Batch Operations
Commands can be chained for automated workflows:

```bash
# Check system health, then apply if healthy
plonk doctor --output json && plonk sync --dry-run
```

### Configuration Validation
Always validate before applying:

```bash
plonk config validate && plonk sync
```

## Shell Completion

Plonk provides intelligent tab completion for enhanced productivity across all major shells.

### Installation

**Temporary (current session only):**
```bash
# Bash
source <(plonk completion bash)

# Zsh
source <(plonk completion zsh)

# Fish
plonk completion fish | source

# PowerShell
plonk completion powershell | Out-String | Invoke-Expression
```

**Permanent installation:**

#### Bash
```bash
# Linux
plonk completion bash | sudo tee /etc/bash_completion.d/plonk

# macOS with Homebrew
plonk completion bash > $(brew --prefix)/etc/bash_completion.d/plonk

# Manual (add to ~/.bashrc)
echo 'source <(plonk completion bash)' >> ~/.bashrc
```

#### Zsh
```bash
# Create completion directory if needed
mkdir -p ~/.local/share/zsh/site-functions

# Install completion
plonk completion zsh > ~/.local/share/zsh/site-functions/_plonk

# Add to ~/.zshrc if not present
echo 'fpath=(~/.local/share/zsh/site-functions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit && compinit' >> ~/.zshrc
```

#### Fish
```bash
# Fish auto-discovers completions
plonk completion fish > ~/.config/fish/completions/plonk.fish
```

#### PowerShell
```powershell
# Save to profile for persistence
plonk completion powershell >> $PROFILE
```

### Completion Features

**Command and subcommand completion:**
```bash
plonk <TAB>          # status, install, uninstall, add, rm, ls, sync, etc.
plonk install <TAB>  # Package name suggestions
plonk add <TAB>      # Dotfile path suggestions
plonk ls <TAB>       # --packages, --dotfiles, --manager, etc.
```

**Package name completion:**
```bash
plonk install <TAB>          # git, curl, htop, ripgrep, etc.
plonk install ri<TAB>        # ripgrep
plonk install --brew <TAB>   # homebrew packages
plonk install --npm <TAB>    # npm packages
```

**Dotfile path completion:**
```bash
plonk add <TAB>      # ~/.zshrc, ~/.vimrc, ~/.config/, etc.
plonk add ~/.<TAB>   # ~/.zshrc, ~/.bashrc, ~/.gitconfig, etc.
plonk add ~/.c<TAB>  # ~/.config/, falls back to system completion
```

**Flag and option completion:**
```bash
plonk ls --output <TAB>      # table, json, yaml
plonk install --<TAB>        # brew, npm, cargo, dry-run, force
plonk add --<TAB>            # dry-run, force
plonk sync --<TAB>           # dry-run, backup, packages, dotfiles
```

**Manager-aware suggestions:**
- **Homebrew**: Development tools, system utilities, CLI apps
- **NPM**: JavaScript packages, build tools, frameworks
- **Cargo**: Rust command-line tools and utilities

### Verification

Test that completion is working:
```bash
plonk install <TAB><TAB>     # Should show package suggestions
plonk add ~/.<TAB>           # Should show dotfile suggestions
plonk ls --output <TAB>      # Should show: table, json, yaml
```

### Debugging Completion

If completion isn't working, you can test it directly:
```bash
# Test completion manually
plonk __complete install ""
plonk __complete add "~/"
plonk __complete ls --output ""
```

## See Also

- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file format
- [ARCHITECTURE.md](ARCHITECTURE.md) - State reconciliation concepts
- [DEVELOPMENT.md](DEVELOPMENT.md) - Contributing and development
