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
plonk status                          # Show managed items + remote sync status
plonk diff                            # Show modified dotfiles

# Git
plonk push                            # Push committed changes to remote
plonk pull                            # Pull remote changes
plonk pull --apply                    # Pull and apply changes

# Utilities
plonk doctor                          # Check system health
plonk config show                     # View settings
plonk clone user/dotfiles             # Clone repo and apply
```

## Migration Notes

- **v0.31**: Templates support macOS Keychain directives such as `{{keychain:plonk/openrouter}}`. Keychain-backed values stay out of shell environment variables and are masked by `plonk diff`.
- **v0.30**: `dotfiles.rules` supports an explicit deploy mode such as `"0600"` for a secret template's rendered target.
- **v0.27**: Mutating commands (`add`, `rm`, `track`, `untrack`, `config edit`) auto-commit to git by default. Disable with `git.auto_commit: false` in `plonk.yaml`.
- **v0.27**: `plonk push` and `plonk pull` synchronize a dotfiles repository.
- **v0.28**: `plonk status`, `plonk packages`, and `plonk dotfiles` show ahead/behind status when a remote is configured.
- `plonk install`/`uninstall`/`upgrade` were removed in v0.26. Install with your package manager, then use `plonk track` / `plonk untrack`.
- Supported managers: `brew`, `cargo`, `go`, `pnpm`, `uv`; the lock format is `version: 3` and automatically migrates from v2 on read.

## Supported Package Managers

| Manager | Prefix | Example |
|---------|--------|---------|
| Homebrew | `brew:` | `plonk track brew:ripgrep` |
| Cargo | `cargo:` | `plonk track cargo:bat` |
| Go | `go:` | `plonk track go:golang.org/x/tools/gopls` |
| PNPM | `pnpm:` | `plonk track pnpm:typescript` |
| UV | `uv:` | `plonk track uv:ruff` |

## Templates and Secrets

Files ending in `.tmpl` are rendered before deployment. Use legacy `{{VAR_NAME}}` or explicit `{{env:VAR_NAME}}` for ordinary machine-specific environment values:

```ini
# ~/.config/plonk/gitconfig.tmpl → ~/.gitconfig
[user]
    email = {{EMAIL}}
    name = {{env:GIT_USER_NAME}}
```

For a macOS secret, use Keychain instead of exporting a credential into your shell:

```json
// ~/.config/plonk/pi/agent/auth.json.tmpl → ~/.pi/agent/auth.json
{"key":"{{keychain:plonk/openrouter}}"}
```

The locator is `keychain:service/account`; omitting `/account` defaults to the current macOS user. Create a Keychain item interactively (do not put the value in a shell command or shell profile):

```bash
security add-generic-password -s plonk -a openrouter -w
```

For a rendered secret file, explicitly set a restrictive deployment mode:

```yaml
dotfiles:
  rules:
    - name: pi/agent/auth.json.tmpl
      mode: "0600"
```

**Rules:**
- Environment, Keychain, and legacy directives have no defaults or conditionals.
- Missing or inaccessible directives make `apply` fail before that file is written; `plonk doctor` identifies the locator and offers remediation without printing secret values.
- Keychain directives require macOS. On other platforms Plonk reports the provider as unavailable.
- A plain file and `.tmpl` file cannot target the same destination.
- `plonk status` compares rendered content in memory. `plonk diff` masks Keychain-derived values as `[REDACTED_SECRET]` before invoking an external diff tool.
- Never run `plonk add` on an existing secret-bearing file. Create a `.tmpl` with a Keychain directive instead.

## How It Works

```
~/.config/plonk/
├── plonk.lock          # Tracked packages (auto-managed)
├── plonk.yaml          # Settings (optional, usually not needed)
├── zshrc               # → ~/.zshrc
├── vimrc               # → ~/.vimrc
├── gitconfig.tmpl      # → ~/.gitconfig (rendered template)
└── config/
    └── nvim/
        └── init.lua    # → ~/.config/nvim/init.lua
```

- **Packages**: Listed in `plonk.lock`, installed on `apply` if missing
- **Dotfiles**: Files in this directory deploy to `$HOME` with a dot prefix
- **Templates**: `.tmpl` files resolve environment values and, on macOS, Keychain values before deployment

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
git:
  auto_commit: true                  # Auto-commit after mutations (default: true)
diff_tool: delta                     # Custom diff viewer
operation_timeout: 600               # Seconds (default: 300)
ignore_patterns:
  - "*.swp"
  - ".DS_Store"

# Optional restrictive permissions for an individual deployed file
dotfiles:
  rules:
    - name: pi/agent/auth.json.tmpl
      mode: "0600"
```

See [docs/reference.md](docs/reference.md) for all options.

## Documentation

- **[CLI & Config Reference](docs/reference.md)** - Complete command and configuration details
- **[Internals](docs/internals.md)** - Architecture for contributors

## Development

```bash
git clone https://github.com/richhaase/plonk
cd plonk
make dev-setup && go test ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT
