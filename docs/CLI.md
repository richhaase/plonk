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

### `plonk status`

Display overall system state across all domains.

**Usage:**
```bash
plonk status [--output format]
```

**Output states:**
- `Managed` - In configuration AND present on system
- `Missing` - In configuration BUT NOT present on system
- `Untracked` - Present on system BUT NOT in configuration

**JSON output structure:**
```json
{
  "packages": {
    "homebrew": [
      {"name": "git", "state": "managed"},
      {"name": "curl", "state": "missing"}
    ],
    "npm": [
      {"name": "typescript", "state": "managed"}
    ]
  },
  "dotfiles": [
    {"name": "zshrc", "state": "managed", "source": "~/.config/plonk/zshrc", "target": "~/.zshrc"},
    {"name": "vimrc", "state": "missing", "source": "~/.config/plonk/vimrc", "target": "~/.vimrc"}
  ]
}
```

### `plonk apply`

Apply configuration to system - installs missing packages and deploys dotfiles.

**Usage:**
```bash
plonk apply [--dry-run] [--backup]
```

**Options:**
- `--dry-run` - Show changes without applying
- `--backup` - Create backups before overwriting dotfiles

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

### `plonk pkg list [filter]`

List packages across all managers.

**Usage:**
```bash
plonk pkg list [managed|missing|untracked] [--output format]
```

**Filters:**
- `managed` - In config AND installed
- `missing` - In config BUT NOT installed
- `untracked` - Installed BUT NOT in config
- No filter - All packages

### `plonk pkg add [package]`

Add packages to configuration.

**Usage:**
```bash
plonk pkg add [package] [--manager manager] [--dry-run]
```

**Behaviors:**
- `plonk pkg add` - Add all untracked packages
- `plonk pkg add htop` - Add specific package
- `plonk pkg add htop --manager homebrew` - Force specific manager

### `plonk pkg remove <package>`

Remove package from configuration.

**Usage:**
```bash
plonk pkg remove <package> [--dry-run]
```

**Note:** Only removes from config, does not uninstall from system.

## Dotfile Management

### `plonk dot list [filter]`

List dotfiles with their states.

**Usage:**
```bash
plonk dot list [managed|missing|untracked] [--output format]
```

**JSON output:**
```json
{
  "dotfiles": [
    {
      "name": "zshrc",
      "state": "managed",
      "source": "~/.config/plonk/zshrc",
      "target": "~/.zshrc",
      "exists": true
    }
  ]
}
```

### `plonk dot add <dotfile>`

Add dotfile to plonk management.

**Usage:**
```bash
plonk dot add <dotfile>
```

**Example:**
```bash
plonk dot add ~/.vimrc  # Copies to ~/.config/plonk/vimrc
```

### `plonk dot re-add <dotfile>`

Re-add existing dotfile (overwrites managed copy).

**Usage:**
```bash
plonk dot re-add <dotfile>
```

## Configuration Management

### `plonk config show`

Display current configuration.

**Usage:**
```bash
plonk config show [--output format]
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
- `1` - Configuration error
- `2` - Package manager error
- `3` - File operation error
- `4` - Validation error

### Error Output Format
```json
{
  "error": {
    "code": "CONFIG_NOT_FOUND",
    "domain": "config",
    "operation": "load",
    "message": "Configuration file not found",
    "suggestion": "Run 'plonk config edit' to create configuration"
  }
}
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