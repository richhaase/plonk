# Plonk

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)

**One command to set up your development environment.**

```bash
brew install --cask richhaase/tap/plonk
plonk clone user/dotfiles
# Done.
```

## What It Does

Plonk manages packages and dotfiles together. Install tools with your package manager, tell plonk to remember them, replicate everywhere.

**Key ideas:**
- **Track, don't install** - Record what's already installed
- **Filesystem as state** - Your `~/.config/plonk/` directory IS your dotfiles
- **Copy, don't symlink** - Simpler and more compatible

## Quick Start

```bash
# Track your dotfiles
plonk add ~/.zshrc ~/.vimrc ~/.config/nvim/

# Install packages normally, then track them
brew install ripgrep fd bat
plonk track brew:ripgrep brew:fd brew:bat

# See what plonk manages
plonk status

# On a new machine: clone and apply
plonk clone your-github/dotfiles
```

## Commands

```bash
# Packages (must be installed first, then tracked)
plonk track brew:ripgrep cargo:bat    # Remember installed packages
plonk untrack brew:ripgrep            # Forget (doesn't uninstall)

# Dotfiles
plonk add ~/.vimrc ~/.zshrc           # Start tracking
plonk rm ~/.vimrc                     # Stop tracking (doesn't delete)

# Sync
plonk apply                           # Install missing packages, deploy dotfiles
plonk apply --dry-run                 # Preview changes
plonk status                          # Show managed items
plonk diff                            # Show modified dotfiles

# Utilities
plonk doctor                          # Check system health
plonk config show                     # View settings
plonk clone user/dotfiles             # Clone repo and apply
```

## Migration Notes (v0.25+)

If you're upgrading from older releases:

- `plonk install`/`uninstall`/`upgrade` were removed.
  - Use your package manager directly, then `plonk track` / `plonk untrack`.
- Supported managers are now: `brew`, `cargo`, `go`, `pnpm`, `uv`.
  - Legacy lock entries for unsupported managers are reported clearly in status/apply output.
- Lock file format is now `version: 3` and migrates automatically from v2 on read.

## Supported Package Managers

| Manager | Prefix | Example |
|---------|--------|---------|
| Homebrew | `brew:` | `plonk track brew:ripgrep` |
| Cargo | `cargo:` | `plonk track cargo:bat` |
| Go | `go:` | `plonk track go:golang.org/x/tools/gopls` |
| PNPM | `pnpm:` | `plonk track pnpm:typescript` |
| UV | `uv:` | `plonk track uv:ruff` |

## Templates

Dotfiles can use environment variable substitution via `.tmpl` files. This lets you keep machine-specific values (email, paths, hostnames) out of your dotfiles repo.

**Create a template** in `$PLONK_DIR` with the `.tmpl` extension:

```ini
# ~/.config/plonk/gitconfig.tmpl → deploys to ~/.gitconfig
[user]
    email = {{EMAIL}}
    name = {{GIT_USER_NAME}}
```

**Set the variables** in your shell, then apply:

```bash
export EMAIL="me@example.com"
export GIT_USER_NAME="My Name"
plonk apply
```

**Rules:**
- Syntax: `{{VAR_NAME}}` (environment variables only, no defaults or conditionals)
- All referenced variables must be set or `apply` fails with a clear error
- A plain file and `.tmpl` file cannot target the same destination
- `plonk doctor` warns about missing template variables
- `plonk diff` and `plonk status` compare rendered output, not raw templates

## How It Works

```
~/.config/plonk/
├── plonk.lock          # Tracked packages (auto-managed)
├── plonk.yaml          # Settings (optional, usually not needed)
├── zshrc               # → ~/.zshrc
├── vimrc               # → ~/.vimrc
├── gitconfig.tmpl      # → ~/.gitconfig (rendered with env vars)
└── config/
    └── nvim/
        └── init.lua    # → ~/.config/nvim/init.lua
```

- **Packages**: Listed in `plonk.lock`, installed on `apply` if missing
- **Dotfiles**: Files in this directory deploy to `$HOME` with a dot prefix
- **Templates**: `.tmpl` files are rendered (env var substitution) before deployment

## Installation

```bash
# Homebrew (recommended)
brew install --cask richhaase/tap/plonk

# Or via Go
go install github.com/richhaase/plonk/cmd/plonk@latest
```

**Requirements:** Homebrew, Git, macOS/Linux/WSL

## Configuration

Plonk works without configuration. If needed, create `~/.config/plonk/plonk.yaml`:

```yaml
# All settings are optional
diff_tool: delta                    # Custom diff viewer
package_timeout: 300                # Seconds (default: 180)
ignore_patterns:
  - "*.swp"
  - ".DS_Store"
```

See [docs/reference.md](docs/reference.md) for all options.

## Documentation

- **[CLI & Config Reference](docs/reference.md)** - Complete command and configuration details
- **[Internals](docs/internals.md)** - Architecture for contributors

## Development

```bash
git clone https://github.com/richhaase/plonk
cd plonk
just dev-setup && go test ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT
