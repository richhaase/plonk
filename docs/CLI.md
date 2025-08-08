# Plonk CLI Reference

## Global Options

All commands support these options:

- `--output, -o` - Output format: `table` (default), `json`, `yaml`
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

Clone a dotfiles repository and intelligently set up plonk.

```bash
plonk clone user/dotfiles             # Clone GitHub repository (shorthand)
plonk clone https://github.com/user/repo.git  # Clone with full URL
plonk clone user/repo --no-apply      # Clone without running apply
```

Options:
- `--no-apply` - Skip running apply after clone


### plonk install

Install packages and add them to management.

```bash
plonk install ripgrep                 # Default manager
plonk install brew:wget npm:prettier uv:ruff pixi:tree composer:phpunit/phpunit dotnet:dotnetsay  # Specific managers
plonk install --dry-run ripgrep       # Preview changes
```

### plonk uninstall

Uninstall packages and remove from management.

```bash
plonk uninstall ripgrep
plonk uninstall brew:wget npm:prettier uv:ruff pixi:tree composer:phpunit/phpunit dotnet:dotnetsay
plonk uninstall --dry-run ripgrep
```

### plonk add

Add dotfiles to management.

```bash
plonk add ~/.vimrc ~/.zshrc           # Add files
plonk add ~/.config/nvim/             # Add directory
plonk add --dry-run ~/.vimrc          # Preview
```

### plonk rm

Remove dotfiles from management (files are not deleted).

```bash
plonk rm ~/.vimrc
plonk rm --dry-run ~/.vimrc
```

### plonk status

Show managed packages and dotfiles.

```bash
plonk status                          # Show all
plonk status --packages               # Only packages
plonk status --dotfiles               # Only dotfiles
plonk status --unmanaged              # Show only unmanaged items
plonk status --missing                # Show only missing resources
plonk status --missing --packages     # Show only missing packages
```

Options:
- `--packages` - Show only package status
- `--dotfiles` - Show only dotfile status
- `--unmanaged` - Show only unmanaged items
- `--missing` - Show only missing resources (mutually exclusive with --unmanaged)

### plonk diff

Show differences for drifted dotfiles.

```bash
plonk diff                            # Show all drifted files
plonk diff ~/.zshrc                   # Show diff for specific file
plonk diff $HOME/.bashrc              # Supports environment variables
```

### plonk apply

Install missing packages and deploy missing dotfiles.

```bash
plonk apply                           # Apply all changes
plonk apply --dry-run                 # Preview changes
plonk apply --packages               # Apply packages only
plonk apply --dotfiles               # Apply dotfiles only
```

### plonk search

Search for packages across all managers.

```bash
plonk search ripgrep                  # Search all managers
plonk search brew:ripgrep             # Search specific manager
```

### plonk info

Show detailed package information.

```bash
plonk info ripgrep
plonk info brew:ripgrep
```

### plonk doctor

Check system health and configuration.

```bash
plonk doctor                          # Check system health
plonk doctor -o json                  # Output as JSON
plonk doctor -o yaml                  # Output as YAML
```

Note: To automatically install missing package managers, use `plonk clone`.

### plonk config

Manage plonk configuration.

```bash
plonk config show                     # Show current config with user-defined values highlighted
plonk config edit                     # Edit config in visudo-style (only saves non-defaults)
```

### plonk dotfiles

List dotfiles specifically.

```bash
plonk dotfiles                        # List all managed dotfiles
plonk dotfiles -o json                # Output as JSON
```

### plonk completion

Generate shell completion scripts.

```bash
plonk completion bash                 # Bash completion
plonk completion zsh                  # Zsh completion
plonk completion fish                 # Fish completion
plonk completion powershell           # PowerShell completion
```

## Package Manager Prefixes

Use prefixes to specify package managers:

- `brew:` - Homebrew (macOS and Linux)
- `npm:` - NPM (global packages)
- `cargo:` - Cargo (Rust)
- `gem:` - RubyGems
- `go:` - Go modules
- `uv:` - UV (Python tool manager)
- `pixi:` - Pixi (Conda-forge packages)
- `composer:` - Composer (PHP global packages)
- `dotnet:` - .NET Global Tools

Examples:
```bash
plonk install brew:wget
plonk install npm:typescript
plonk install cargo:ripgrep
plonk install gem:bundler
plonk install go:golang.org/x/tools/cmd/goimports
plonk install uv:ruff
plonk install pixi:tree
plonk install composer:phpunit/phpunit
plonk install dotnet:dotnetsay
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
