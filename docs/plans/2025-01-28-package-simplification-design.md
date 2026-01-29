# Package Management Simplification

## Overview

Simplify package management to mirror the dotfiles approach: track what you have, deploy to other machines. No upgrades, no uninstalls, no magic.

## Core Principles

1. **Track, don't manage** - Plonk records what packages you want, doesn't try to be a package manager
2. **Explicit always** - `brew:ripgrep` not `ripgrep` with default manager inference
3. **Track only installed** - Can't track arbitrary packages, must be installed first
4. **Apply installs missing** - Idempotent deployment to new machines

## Interface

```go
type PackageManager interface {
    IsInstalled(ctx context.Context, name string) (bool, error)
    Install(ctx context.Context, name string) error
}
```

Two methods. No mocks, no CommandExecutor - test with BATS.

## Supported Managers

| Manager | Ecosystem |
|---------|-----------|
| brew | System packages (macOS) |
| pnpm | JavaScript/Node |
| cargo | Rust |
| go | Go |
| uv | Python |

Removed: npm, pipx, gem, conda.

## Lock File Format (v3)

```yaml
version: 3
packages:
  brew:
    - fd
    - jq
    - ripgrep
  cargo:
    - bat
    - eza
  go:
    - golang.org/x/tools/gopls
```

- Grouped by manager
- Alphabetically sorted
- No metadata, no timestamps
- Auto-migrate from v2 on first load

## Commands

### `plonk track <manager:pkg>...`

```
$ plonk track brew:ripgrep cargo:bat
Tracking brew:ripgrep
Tracking cargo:bat
```

- Verifies package is installed before tracking
- Rejects if not installed
- Adds to lock file

### `plonk untrack <manager:pkg>...`

```
$ plonk untrack brew:ripgrep
Untracking brew:ripgrep
```

- Removes from lock file
- Does NOT uninstall

### `plonk apply`

```
$ plonk apply
Installing brew:ripgrep... done
Skipping cargo:bat (installed)
```

- Checks IsInstalled for each tracked package
- Installs only missing packages
- Skips managers entirely if nothing needed

## Commands Deleted

- `plonk install` - replaced by manual install + track
- `plonk rm` - replaced by untrack
- `plonk upgrade` - gone
- `plonk add` - gone (if it existed for packages)

## File Structure

```
internal/packages/
├── manager.go        # PackageManager interface
├── registry.go       # Maps names to managers
├── brew.go           # ~20 lines
├── pnpm.go           # ~20 lines
├── cargo.go          # ~20 lines
├── go.go             # ~20 lines
├── uv.go             # ~20 lines
├── apply.go          # Check + install missing
└── track.go          # Verify + update lock
```

Delete: base.go, interfaces.go (merge), upgrade.go, npm.go, pipx.go, gem.go, conda.go, all *_test.go files.

## Migration

- Lock v2 auto-migrates to v3 on first load
- Extract manager:name from ResourceEntry, rewrite grouped format
- Only migrate package resources, dotfiles handled separately

## Workflow

**On main machine:**
1. `brew install ripgrep`
2. `plonk track brew:ripgrep`
3. Commit plonk.lock

**On new machine:**
1. Clone dotfiles repo
2. `plonk apply`
3. ripgrep gets installed
