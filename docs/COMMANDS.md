# Plonk CLI Commands

Plonk manages packages and dotfiles consistently across multiple machines using Homebrew and NPM package managers.

## Global Options

- `--output, -o`: Output format (table|json|yaml) - default: table
- `--version, -v`: Show version information
- `--help, -h`: Show help for any command

## Commands

### `plonk status`
Display overall plonk status including configuration, packages, and dotfiles.

### `plonk apply`
Apply entire plonk configuration (packages and dotfiles) to your system.

**Options:**
- `--dry-run` - Show what would be applied without making changes
- `--backup` - Create backups before overwriting existing dotfiles

Applies all configured packages and dotfiles in a single operation.

### `plonk env`
Show environment information for plonk including system details, package manager availability, configuration status, and path information. Useful for debugging and troubleshooting.

### `plonk doctor`
Perform comprehensive health checks on your plonk installation. Diagnoses common issues with system requirements, environment variables, permissions, configuration, and package manager functionality.

**Health check categories:**
- System requirements (OS, shell, basic tools)
- Environment variables and PATH configuration
- File permissions and directory access
- Configuration file validation
- Package manager availability and functionality
- Executable installation and PATH setup

Each check provides actionable suggestions for fixing identified issues.

### `plonk search <package>`
Search for packages across available package managers. Provides intelligent search behavior based on installation status and configuration.

**Search behavior:**
- If package is installed: Shows installation status and managing package manager
- If not installed and available in default manager: Shows results from default manager
- If not installed and not in default manager: Shows all managers that have the package
- If no default manager configured: Shows all managers that have the package

**Examples:**
```bash
plonk search git              # Search for git package
plonk search typescript       # Search for typescript package  
plonk search --output json go # Search with JSON output
```

### `plonk info <package>`
Show detailed information about a specific package including version, description, homepage, dependencies, and installation status.

**Information displayed:**
- Package name and version
- Description and homepage URL
- Installation status and managing package manager
- Dependencies (when available)
- Size information (when available)

**Examples:**
```bash
plonk info git                    # Show git package information
plonk info typescript             # Show typescript package information
plonk info --output json webpack  # Show information in JSON format
```

### Package Management (`plonk pkg`)

- `plonk pkg list [filter]` - List packages across all managers
- `plonk pkg add <package>` - Add a package to plonk configuration and install it
- `plonk pkg remove <package>` - Remove a package from plonk configuration

**List filters:** `managed`, `untracked`, `missing`

### Dotfile Management (`plonk dot`)

- `plonk dot list [filter]` - List dotfiles across all locations
- `plonk dot add <dotfile>` - Add a dotfile to plonk configuration
- `plonk dot re-add <dotfile>` - Re-add a dotfile to plonk configuration

**List filters:** `managed`, `untracked`, `missing`

### Configuration (`plonk config`)

- `plonk config show` - Display current configuration content
- `plonk config validate` - Validate configuration file for syntax and structural errors
- `plonk config edit` - Edit configuration file using your preferred editor

## Configuration File

Location: `~/.config/plonk/plonk.yaml`

Contains package definitions for Homebrew and NPM, dotfile definitions with source and destination paths, and manager-specific settings.