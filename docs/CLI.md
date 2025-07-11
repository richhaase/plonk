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
| `apply` | Apply configuration | Automated deployment |
| `env` | Environment info | Debugging context |
| `doctor` | Health checks | System validation |
| `search` | Find packages | Package discovery |
| `info` | Package details | Package analysis |
| `pkg` | Package management | Package operations |
| `dot` | Dotfile management | Dotfile operations |
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

Display overall system state across all domains.

**Usage:**
```bash
plonk status [--output format]
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

### `plonk apply`

Apply managed state to system - installs missing packages and deploys dotfiles.

**Usage:**
```bash
plonk apply [--dry-run] [--backup]
```

**Options:**
- `--dry-run` - Show changes without applying
- `--backup` - Create backups before overwriting dotfiles

**Behavior:**
- Processes both missing dotfiles (new deployments) and managed dotfiles (updates)
- Installs packages marked as missing in the lock file
- Deploys dotfiles from the configuration directory to their target locations

**Backup functionality:**
- When `--backup` is used, creates timestamped backups before overwriting existing files
- Backup format: `{filename}.backup.{YYYYMMDD-HHMMSS}`
- Backups are created in the same directory as the original file
- Multiple backups can exist with unique timestamps
- To restore: manually copy the desired backup file over the current file

**Example output:**
```
Apply Configuration
===================

Packages:
✓ Installing htop via homebrew
✓ Installing ripgrep via homebrew
✓ All 2 missing packages installed

Dotfiles:
✓ Deployed ~/.zshrc from zshrc
✓ Deployed ~/.vimrc from vimrc (backup created)
✓ Updated ~/.config/nvim/init.lua from config/nvim/init.lua
✓ All 3 dotfiles processed

Summary: 2 packages installed, 3 dotfiles deployed
```

**Exit codes:**
- `0` - Success
- `1` - Configuration error
- `2` - Package manager error
- `3` - File operation error

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

## Package Management

### `plonk pkg list`

List packages across all managers with smart defaults.

**Usage:**
```bash
plonk pkg list [--verbose] [--manager manager] [--output format]
```

**Behavior:**
- By default shows managed + missing packages with untracked count
- Use `--verbose` to see all packages including full untracked list
- Use `--manager` to filter by package manager (homebrew, npm)
- Sort by state (managed, missing, untracked), then alphabetically

**Options:**
- `--verbose` - Show all packages including untracked
- `--manager` - Filter by package manager (homebrew, npm)

**Example output:**
```
Package Summary
===============
Total: 25 packages | ✓ Managed: 18 | ⚠ Missing: 2 | ? Untracked: 5

  Status Package                        Manager
  ------ ------------------------------ ----------
  ✓      git                            homebrew
  ✓      curl                           homebrew
  ⚠      htop                           homebrew
  ✓      typescript                     npm
  ✓      ripgrep                        cargo

5 untracked packages (use --verbose to show details)
```

### `plonk pkg add [package]`

Add packages to lock file and install them.

**Usage:**
```bash
plonk pkg add [package] [--manager manager] [--dry-run] [--output format]
```

**Behaviors:**
- `plonk pkg add` - Add all untracked packages
- `plonk pkg add htop` - Add specific package to lock file and install
- `plonk pkg add htop --manager homebrew` - Force specific manager
- `plonk pkg add ripgrep --manager cargo` - Force specific manager
- `plonk pkg add htop --dry-run` - Preview what would be added/installed

**Options:**
- `--manager` - Force specific package manager (homebrew, npm)
- `--dry-run` - Show what would be added without making changes

**Example output:**
```
Package Add
===========
✓ Added htop to homebrew configuration
✓ Successfully installed htop

Summary: Added to configuration and installed
```

### `plonk pkg remove <package>`

Remove package from lock file with enhanced output.

**Usage:**
```bash
plonk pkg remove <package> [--uninstall] [--dry-run] [--output format]
```

**Options:**
- `--uninstall` - Also uninstall the package from the system
- `--dry-run` - Show what would be removed without making changes

**Behaviors:**
- `plonk pkg remove htop` - Remove from lock file only
- `plonk pkg remove htop --uninstall` - Remove from lock file and uninstall
- `plonk pkg remove htop --dry-run` - Preview what would be removed

**Example output:**
```
Package Remove
==============
✓ Removed htop from homebrew configuration
✓ Successfully uninstalled htop from system

Summary: removed from configuration and uninstalled from system
```

**Note:** By default, only removes from config. Use `--uninstall` to also remove from system.

## Dotfile Management

### `plonk dot list`

List dotfiles with their states and smart defaults.

**Usage:**
```bash
plonk dot list [--verbose] [--output format]
```

**Behavior:**
- By default shows missing + managed files with untracked count
- Use `--verbose` to see all files including full untracked list
- Configured directories are expanded to show individual files
- Applies ignore patterns to filter out noise files

**Example output:**
```
Dotfiles Summary
================
Total: 59 files | ✓ Managed: 12 | ⚠ Missing: 0 | ? Untracked: 47

  Status Target                                    Source
  ------ ----------------------------------------- --------------------------------------
  ✓      ~/.config/nvim/init.lua                   config/nvim/init.lua
  ✓      ~/.zshrc                                  zshrc
  ?      ~/.aws/cli                                -
  ?      ~/.aws/config                             -

47 untracked files (use --verbose to show details)
```

**JSON output:**
```json
{
  "summary": {
    "total": 59,
    "managed": 12,
    "missing": 0,
    "untracked": 47,
    "verbose": false
  },
  "dotfiles": [
    {
      "name": ".zshrc",
      "state": "managed",
      "target": "~/.zshrc",
      "source": "zshrc"
    }
  ]
}
```

### `plonk dot add <dotfile>`

Add or update dotfile in plonk management.

**Usage:**
```bash
plonk dot add <dotfile>
```

**Behavior:**
- **New files**: Copies file to plonk config and marks as managed
- **Existing files**: Updates the managed copy with current system version
- **Directories**: Recursively processes all files, respecting ignore patterns

**Path Resolution:**
- **Absolute paths**: `plonk dot add /home/user/.vimrc`
- **Tilde paths**: `plonk dot add ~/.vimrc`
- **Relative paths**: First tries current directory, then home directory
  - `plonk dot add .vimrc` → looks for `./vimrc` then `~/.vimrc`
  - `plonk dot add init.lua` → looks for `./init.lua` then `~/init.lua`

**Examples:**
```bash
plonk dot add ~/.vimrc          # Explicit home directory path
plonk dot add .vimrc            # Finds ~/.vimrc (if not in current dir)
plonk dot add ~/.config/nvim/   # Add entire directory
cd ~/.config/nvim && plonk dot add init.lua  # Finds ./init.lua
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
plonk apply  # Shows technical details for errors
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
plonk doctor --output json && plonk apply --dry-run
```

### Configuration Validation
Always validate before applying:

```bash
plonk config validate && plonk apply
```

## See Also

- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file format
- [ARCHITECTURE.md](ARCHITECTURE.md) - State reconciliation concepts
- [DEVELOPMENT.md](DEVELOPMENT.md) - Contributing and development
