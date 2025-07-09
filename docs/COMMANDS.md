# Plonk CLI Commands

Plonk manages packages and dotfiles consistently across multiple machines using Homebrew and NPM package managers.

## Global Options

- `--output, -o`: Output format (table|json|yaml) - default: table
- `--version, -v`: Show version information
- `--help, -h`: Show help for any command

## Commands

### `plonk status`
Display overall plonk status including configuration, packages, and dotfiles.

### Package Management (`plonk pkg`)

- `plonk pkg list [filter]` - List packages across all managers
- `plonk pkg add <package>` - Add a package to plonk configuration and install it
- `plonk pkg remove <package>` - Remove a package from plonk configuration
- `plonk pkg apply` - Install all packages defined in configuration

**List filters:** `managed`, `untracked`, `missing`

### Dotfile Management (`plonk dot`)

- `plonk dot list [filter]` - List dotfiles across all locations
- `plonk dot add <dotfile>` - Add a dotfile to plonk configuration
- `plonk dot re-add <dotfile>` - Re-add a dotfile to plonk configuration
- `plonk dot apply` - Apply dotfile configuration (link/copy dotfiles to home directory)

**List filters:** `managed`, `untracked`, `missing`

### Configuration (`plonk config`)

- `plonk config show` - Display current configuration content

## Configuration File

Location: `~/.config/plonk/plonk.yaml`

Contains package definitions for Homebrew and NPM, dotfile definitions with source and destination paths, and manager-specific settings.