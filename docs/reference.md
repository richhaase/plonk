# Plonk Reference

Complete CLI and configuration reference.

## Migration Notes

- **v0.31**: Templates support macOS Keychain directives (`{{keychain:service/account}}`) and mask Keychain-derived values in `plonk diff`.
- **v0.30**: `dotfiles.rules` can set an explicit deploy mode, such as `"0600"`, for an individual dotfile.
- **v0.27**: Mutating commands (`add`, `rm`, `track`, `untrack`, `config edit`) auto-commit by default. Disable with `git.auto_commit: false` in `plonk.yaml`.
- **v0.27**: `plonk push` and `plonk pull` synchronize your dotfiles repository.
- `install`, `uninstall`, and `upgrade` were removed in v0.26; package operations use `track`, `untrack`, and `apply`.
- Supported package managers: `brew`, `cargo`, `go`, `pnpm`, `uv`.
- Lock files use `version: 3`; older v2 files migrate automatically.

## Commands

### plonk track

Track packages that are already installed.

```bash
plonk track <manager:package>...
```

- Verifies packages are installed before tracking
- Adds to `plonk.lock`
- Format `manager:package` is required (no default manager)

```bash
plonk track brew:ripgrep cargo:bat go:golang.org/x/tools/gopls
```

### plonk untrack

Stop tracking packages (does not uninstall).

```bash
plonk untrack <manager:package>...
```

```bash
plonk untrack brew:ripgrep
```

### plonk add

Add dotfiles to management.

```bash
plonk add <file>...
plonk add -y              # Sync all drifted files back to $PLONK_DIR
plonk add --dry-run       # Preview
```

Copies files from `$HOME` to `$PLONK_DIR`, stripping the dot prefix.

### plonk rm

Remove dotfiles from management (does not delete deployed files).

```bash
plonk rm <file>...
plonk rm --dry-run
```

### plonk apply

Install missing packages and deploy missing/drifted dotfiles.

```bash
plonk apply [options] [files...]
```

**Options:**
- `--dry-run, -n` - Preview changes
- `--packages` - Packages only
- `--dotfiles` - Dotfiles only

```bash
plonk apply                    # Everything
plonk apply --packages         # Packages only
plonk apply ~/.vimrc           # Specific dotfile
```

### plonk status

Show managed packages and dotfiles.

```bash
plonk status
plonk st                       # Alias
```

**States:**
- `managed` - Tracked and present
- `missing` - Tracked but not present
- `drifted` - Dotfile modified since deployment

### plonk dotfiles

Show dotfile status only.

```bash
plonk dotfiles
plonk d                        # Alias
```

### plonk diff

Show differences for drifted dotfiles.

```bash
plonk diff                     # All drifted
plonk diff ~/.zshrc            # Specific file
```

Uses `git diff` by default, or `diff_tool` from config.

### plonk clone

Clone a dotfiles repository and apply.

```bash
plonk clone <repo>
plonk clone user/dotfiles              # GitHub shorthand
plonk clone https://github.com/u/r.git # Full URL
plonk clone --dry-run user/dotfiles    # Preview
```

### plonk push

Push committed changes to the remote.

```bash
plonk push
```

- Warns if there are uncommitted changes in the working tree
- Requires a configured remote

### plonk pull

Pull remote changes into your plonk directory.

```bash
plonk pull [options]
```

**Options:**
- `--apply, -a` - Run `plonk apply` after pulling

If there are uncommitted local changes and `auto_commit` is enabled, they are committed first. If `auto_commit` is disabled and there are uncommitted changes, the pull is refused.

```bash
plonk pull                    # Pull remote changes
plonk pull --apply            # Pull and apply
```

### plonk doctor

Check system health.

```bash
plonk doctor
```

Reports: config directory, permissions, package manager availability, template variable readiness.

### plonk config

View and edit configuration.

```bash
plonk config show              # View current config
plonk config show -o json      # JSON output
plonk config edit              # Edit in $EDITOR
```

### plonk completion

Generate shell completions.

```bash
plonk completion bash
plonk completion zsh
plonk completion fish
```

## Package Managers

| Manager | Prefix | Install Command |
|---------|--------|-----------------|
| Homebrew | `brew:` | `brew install <pkg>` |
| Cargo | `cargo:` | `cargo install <pkg>` |
| Go | `go:` | `go install <pkg>@latest` |
| PNPM | `pnpm:` | `pnpm add -g <pkg>` |
| UV | `uv:` | `uv tool install <pkg>` |

## Configuration

Configuration file: `~/.config/plonk/plonk.yaml`

All settings are optional. Plonk uses sensible defaults.

### Settings

```yaml
# Git integration
git:
  auto_commit: true        # Auto-commit after mutations (default: true)

# Package manager default (for discovery, not tracking)
default_manager: brew

# Timeouts (seconds)
operation_timeout: 300     # General operations
dotfile_timeout: 60        # File operations
# Note: package installs use a fixed 10-minute per-package timeout.

# Diff tool for viewing drifted files
diff_tool: delta           # Default: git diff --no-index

# Directories to scan for dotfiles
expand_directories:
  - .config                # Default

# Files to ignore
ignore_patterns:
  - "*.swp"
  - "*.tmp"
  - ".DS_Store"
  - ".git/*"

# Per-dotfile deploy permissions (optional)
dotfiles:
  rules:
    - name: "pi/agent/auth.json.tmpl"   # Source path relative to $PLONK_DIR
      mode: "0600"                       # Octal permissions for the deployed file
```

#### Per-dotfile deploy mode (`dotfiles.rules`)

Optionally assign explicit permissions to specific dotfiles on deploy, overriding
the source/git file permissions. This is useful for secret templates (e.g. auth
files) that should be deployed with restrictive permissions such as `0600` even
when committed with standard `0644` permissions.

- `name` - Dotfile source path relative to `$PLONK_DIR` (e.g. `pi/agent/auth.json.tmpl`). Required.
- `mode` - Octal file permissions applied to the deployed target after write and rename. Must be in the range `0000`-`0777` (digits `0`-`7` only). Invalid values produce a configuration validation error.

When no rule matches, or a rule has no `mode`, behavior is unchanged: the deployed
file keeps the source file permissions.

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `PLONK_DIR` | Config directory (default: `~/.config/plonk`) |
| `VISUAL` | Editor for `config edit` |
| `EDITOR` | Fallback editor |
| `NO_COLOR` | Disable colored output |

### Precedence

1. Command-line flags
2. Environment variables
3. `plonk.yaml`
4. Built-in defaults

## Templates

Dotfiles with a `.tmpl` extension are rendered before deployment. A template may use an environment variable or, on macOS, a generic-password item from Keychain.

### Syntax

#### Environment variables

Legacy `{{VAR_NAME}}` and explicit `{{env:VAR_NAME}}` both resolve environment variables:

```ini
# gitconfig.tmpl
[user]
    email = {{EMAIL}}
    name = {{env:GIT_USER_NAME}}
```

#### macOS Keychain secrets

Use `{{keychain:service/account}}` for a generic-password item. If `/account` is omitted, Plonk uses the current macOS username as the account.

```json
// pi/agent/auth.json.tmpl
{
  "openrouter": {
    "type": "api_key",
    "key": "{{keychain:plonk/openrouter}}"
  }
}
```

Create the item interactively so its value is not exposed through shell history or a command argument:

```bash
security add-generic-password -s plonk -a openrouter -w
```

Plonk reads Keychain using the macOS `security` tool with a fixed executable path, a restricted process environment, and a timeout. It is a Keychain consumer only; it never stores or changes Keychain values.

### How It Works

1. Place a `.tmpl` file in `$PLONK_DIR` (e.g., `gitconfig.tmpl`)
2. On `plonk apply`, Plonk resolves each directive in memory
3. The rendered output is deployed to `$HOME` with `.tmpl` stripped (e.g., `~/.gitconfig`)
4. Configure `dotfiles.rules` mode `"0600"` for a template whose rendered output contains a secret

### Behavior and Security

- Missing, locked, inaccessible, malformed, or unavailable providers fail the affected apply without printing a secret value.
- `plonk doctor` checks directives and reports their locator plus a provider-specific remediation hint; it never reports a resolved value.
- A plain file and a `.tmpl` file must not target the same destination (e.g., `gitconfig` and `gitconfig.tmpl` cannot coexist).
- `plonk status` compares rendered content in memory.
- For templates containing Keychain directives, `plonk diff` masks resolved values as `[REDACTED_SECRET]` before writing temp files or invoking the configured external diff tool.
- `plonk rm gitconfig` recognizes and removes the `gitconfig.tmpl` source file.
- Keychain directives are supported only on macOS. On other operating systems Plonk reports the provider as unavailable.

### Limitations (By Design)

- No conditionals, loops, or template functions
- No default/fallback values
- No Keychain writes from Plonk
- No Linux Secret Service or Windows Credential Manager provider yet

## Lock File

`plonk.lock` is auto-managed. Format:

```yaml
version: 3
packages:
  brew:
    - fd
    - ripgrep
  cargo:
    - bat
  go:
    - golang.org/x/tools/gopls
```

## Exit Codes

- `0` - Success
- `1` - Error

Template-provider failures are reported with a specific cause in the error message: secret not found, provider unavailable, Keychain locked, access denied, or invalid directive syntax.

## Output Formats

Commands support `--output` / `-o`:
- `table` (default)
- `json`
- `yaml`
