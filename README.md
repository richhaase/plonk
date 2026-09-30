# Plonk

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)

**One command to set up your development environment.**

```bash
brew install --cask richhaase/tap/plonk
plonk clone user/dotfiles
# Install any required package managers first.
```

## What It Does

Plonk manages packages and dotfiles together. Add the files and tools you want, then replicate them across machines.

**Key ideas:**
- **Add files and packages** - Track installed tools or install missing ones
- **Filesystem as state** - Your `~/.config/plonk/` directory IS your dotfiles
- **Copy, don't symlink** - Simpler and more compatible

## Quick Start

```bash
# Track your dotfiles
plonk add ~/.zshrc ~/.vimrc ~/.config/nvim/

# Install missing packages and track them
plonk add brew:ripgrep brew:fd brew:bat

# See everything plonk manages
plonk status --all

# On a new machine: clone and apply
plonk clone your-github/dotfiles
```

## Commands

```bash
# Packages
plonk add brew:ripgrep cargo:bat  # Install if missing and track
plonk rm brew:ripgrep             # Untrack (keeps installed)
plonk rm -f brew:ripgrep          # Uninstall and untrack

# Dotfiles
plonk add ~/.vimrc ~/.zshrc           # Start tracking
plonk rm ~/.vimrc                     # Stop managing (keeps deployed file)
plonk rm -f ~/.vimrc                  # Delete source and deployed file

# Sync
plonk apply                           # Install missing packages, deploy dotfiles
plonk apply --dry-run                 # Preview changes
plonk status                          # Show remote status and actionable items
plonk status --all                    # Include healthy managed items
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

## Supported Package Managers

| Manager | Prefix | Example |
|---------|--------|---------|
| Homebrew | `brew:` | `plonk add brew:ripgrep` |
| Cargo | `cargo:` | `plonk add cargo:bat` |
| Go | `go:` | `plonk add go:golang.org/x/tools/gopls` |
| PNPM | `pnpm:` | `plonk add pnpm:typescript` |
| UV | `uv:` | `plonk add uv:ruff` |

## Templates and Secrets

Files ending in `.tmpl` are rendered before deployment. Use legacy `{{VAR_NAME}}` or explicit `{{env:VAR_NAME}}` for ordinary machine-specific environment values:

```ini
# ~/.config/plonk/gitconfig.tmpl → ~/.gitconfig
[user]
    email = {{EMAIL}}
    name = {{env:GIT_USER_NAME}}
```

For a macOS secret, use Keychain instead of exporting a credential into your shell:

`pi/agent/auth.json.tmpl` deploys to `~/.pi/agent/auth.json`:

```json
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

`plonk apply` also corrects explicitly configured permissions when file contents are unchanged; `--dry-run` previews the update.

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

- **Packages**: Listed in `plonk.lock`, installed on `apply` if missing. Each manager finishes before the next is checked, so tracked tools such as `brew:pnpm` can supply a later manager on the current `PATH`.
- **Dotfiles**: Files in this directory deploy to `$HOME` with a dot prefix
- **Templates**: `.tmpl` files resolve environment values and, on macOS, Keychain values before deployment

## Installation

```bash
# Homebrew (recommended)
brew install --cask richhaase/tap/plonk

# Or via Go
go install github.com/richhaase/plonk/cmd/plonk@latest
```

Install the package managers you use and ensure their executables are on `PATH`.
Git is needed for clone, repository synchronization, auto-commit, and the default
diff tool. Keychain templates require macOS.

## Configuration

Plonk works without configuration. If needed, create `~/.config/plonk/plonk.yaml`:

```yaml
git:
  auto_commit: false       # Default: true
operation_timeout: 600    # Default: 300 seconds
```

Use `plonk config show` to inspect effective settings and `plonk config edit` to
edit them. Color is automatic on terminals; set `NO_COLOR=1` to disable it.

See [docs/reference.md](docs/reference.md) for all options.

## Documentation

- **[CLI & Config Reference](docs/reference.md)** - Complete command and configuration details
- **[Internals](docs/internals.md)** - Architecture for contributors

## Development

```bash
git clone https://github.com/richhaase/plonk
cd plonk
make build
go test ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT
