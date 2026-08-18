# Plonk Internals

Architecture and code organization for contributors.

## Overview

Plonk is a state reconciliation engine. It compares desired state (lock file, config directory) with actual state (installed packages, deployed files) and applies changes to reconcile them.

## Directory Structure

```
plonk/
├── cmd/plonk/main.go           # Entry point
├── internal/
│   ├── commands/               # CLI commands
│   │   ├── root.go             # Root command, global flags
│   │   ├── track.go            # Package tracking
│   │   ├── untrack.go          # Package untracking
│   │   ├── add.go              # Dotfile addition
│   │   ├── rm.go               # Dotfile removal
│   │   ├── apply.go            # State application
│   │   ├── status.go           # Status display
│   │   ├── diff.go             # Drift display
│   │   ├── clone.go            # Repository cloning
│   │   ├── push.go             # Git push
│   │   ├── pull.go             # Git pull (with optional apply)
│   │   ├── doctor.go           # Health checks
│   │   └── config*.go          # Configuration commands
│   ├── packages/               # Package management
│   │   ├── manager.go          # Manager interface
│   │   ├── registry.go         # Manager lookup
│   │   ├── apply.go            # Package application
│   │   ├── brew.go             # Homebrew
│   │   ├── cargo.go            # Cargo
│   │   ├── go.go               # Go
│   │   ├── pnpm.go             # PNPM
│   │   └── uv.go               # UV
│   ├── dotfiles/               # Dotfile management
│   │   ├── dotfiles.go         # Manager + operations
│   │   ├── reconcile.go        # State reconciliation
│   │   ├── apply.go            # Selective apply
│   │   ├── types.go            # Dotfile/Status types
│   │   └── fs.go               # FileSystem abstraction
│   ├── orchestrator/           # Coordination
│   │   ├── coordinator.go      # Apply coordination
│   │   └── reconcile.go        # Cross-domain reconciliation
│   ├── config/                 # Configuration
│   │   └── config.go           # Config loading/defaults
│   ├── lock/                   # Lock file
│   │   ├── v3.go               # V3 format + migration
│   │   └── types.go            # Lock types
│   ├── gitops/                 # Git automation
│   │   ├── gitops.go           # Git client (commit, push, pull)
│   │   └── autocommit.go       # Post-mutation auto-commit hook
│   ├── clone/                  # Clone operations
│   │   ├── setup.go            # Clone + apply
│   │   └── git.go              # Git operations
│   ├── diagnostics/            # Health checks
│   │   └── health.go           # System checks
│   ├── template/               # Template parser and value resolvers
│   │   ├── template.go         # Directive grammar and parsing
│   │   ├── render.go           # Rendering and secret masking
│   │   ├── env.go              # Environment resolver
│   │   └── keychain.go         # macOS Keychain resolver
│   └── output/                 # Output formatting
│       ├── formatters.go       # Table/JSON/YAML
│       └── colors.go           # Terminal colors
└── tests/bats/                 # Integration tests
```

## Key Interfaces

### Manager (packages)

```go
type Manager interface {
    IsInstalled(ctx context.Context, name string) (bool, error)
    Install(ctx context.Context, name string) error
}
```

That's it. Two methods per package manager.

### Lock Service

```go
type LockV3Service interface {
    Read() (*LockV3, error)
    Write(lock *LockV3) error
}
```

## State Model

### Package State

Stored in `plonk.lock`:
```yaml
version: 3
packages:
  brew: [fd, ripgrep]
  cargo: [bat]
```

### Dotfile State

The filesystem IS the state. Files in `$PLONK_DIR` (excluding `plonk.yaml`, `plonk.lock`) are managed dotfiles. Files with the `.tmpl` extension are rendered before deployment using environment variables and, on macOS, Keychain generic-password items.

### Resource States

- **managed** - Tracked and exists
- **missing** - Tracked but doesn't exist
- **drifted** - Exists but modified (dotfiles only)
- **unmanaged** - Exists but not tracked

## Data Flow

### Track Flow
```
User → track command → Verify installed → Update lock file
```

### Apply Flow
```
Lock file → List tracked → Check installed → Install missing
Config dir → List files → Resolve and render .tmpl → Check deployed → Deploy missing/drifted
```

### Auto-Commit Flow
```
Mutation command succeeds → gitops.AutoCommit → Check config → Check git repo → git add -A → git commit
```

Auto-commit is best-effort: failures are warnings, not errors. The mutation itself already succeeded. Controlled by `git.auto_commit` in `plonk.yaml` (default: `true`). If `$PLONK_DIR` is not a git repo, warns and skips.

### Status Flow
```
Lock file + system state → Reconcile → Display differences
```

## Adding a Package Manager

1. Create `internal/packages/newmanager.go`:

```go
type NewManagerSimple struct{}

func NewNewManagerSimple() *NewManagerSimple {
    return &NewManagerSimple{}
}

func (m *NewManagerSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
    // Check if package is installed
}

func (m *NewManagerSimple) Install(ctx context.Context, name string) error {
    // Install package
}
```

2. Register in `internal/packages/registry.go`:

```go
case "newmanager":
    return NewNewManagerSimple(), nil
```

3. Add to `SupportedManagers` in `internal/packages/manager.go`

4. Add tests in `tests/bats/behavioral/03-package-managers.bats`

## Testing

### Unit Tests

```bash
go test ./...
go test -v ./internal/packages/...
```

### BATS Integration Tests

```bash
bats tests/bats/behavioral/
```

BATS tests call the real CLI and real package managers. Use the safe package list in `tests/bats/config/safe-packages.list`.

## Template Rendering and Secret Resolution

Files ending in `.tmpl` are rendered before deployment or comparison. The shared `internal/template` package owns directive parsing, resolution, redaction, and error classification; template grammar is not duplicated across commands and diagnostics.

### Directive forms

- `{{VAR_NAME}}` — legacy environment variable.
- `{{env:VAR_NAME}}` — explicit environment variable.
- `{{keychain:service/account}}` — macOS Keychain generic-password item. Omitting `/account` defaults to the current macOS username.

### Implementation

- **`internal/template`**: Parses directives in two phases, dispatches to registered resolvers, and exposes typed errors for missing secrets, unavailable providers, locked Keychain, access denial, and malformed directives.
- **Resolvers**: `EnvResolver` preserves legacy behavior. `MacOSKeychainResolver` invokes literal `/usr/bin/security` with separated arguments, a restricted environment, and a timeout. `MockSecretResolver` supports deterministic tests without a host Keychain.
- **`DotfileManager`**: Uses the shared renderer for deployment, drift detection, and source rendering. Resolved values remain in memory until the normal atomic deployment path writes the authorized target under `$HOME`.
- **`toTarget()`**: Strips `.tmpl` before adding the dot prefix, so `gitconfig.tmpl` targets `~/.gitconfig`.
- **Conflict detection**: `List()` builds a target-path map and errors if two sources (e.g., `gitconfig` and `gitconfig.tmpl`) resolve to the same target.
- **`diagnostics/health.go`**: Parses templates through the shared package and reports unresolved locators with provider-owned remediation hints, never resolved values.
- **`commands/diff.go`**: Uses a redacted render for templates containing Keychain directives. `[REDACTED_SECRET]`, not a plaintext secret, is written to temporary diff files or passed to an external diff tool.

### Secret boundary

Plaintext secrets are permitted only while rendering in process memory and in the explicitly authorized deployed target file. They must not be emitted in command output, errors, logs, environment variables, command arguments, or temporary diff files.

## Error Handling

- Commands return structured results with per-item status
- Partial failures supported (some succeed, some fail)
- Non-zero exit codes for scripting

## Output Formatting

All commands support `-o table|json|yaml`. Table is default for humans, JSON/YAML for scripting.
