# Package Management Commands

Commands for managing packages: `track` and `untrack`.

## Description

Package management in plonk follows a "tracking" model similar to dotfiles. Instead of installing and uninstalling packages, plonk tracks packages that are already installed on your system. This approach is simpler and more predictable:

- **track** - Record an installed package in the lock file
- **untrack** - Remove a package from tracking (does NOT uninstall)
- **apply** - Install any tracked packages that are missing

Package state is tracked in `plonk.lock`, which is updated atomically with each operation.

## Supported Package Managers

Plonk supports 5 package managers:

- `brew:` - Homebrew (macOS and Linux)
- `cargo:` - Cargo (Rust)
- `go:` - Go (Go binaries)
- `pnpm:` - PNPM (fast, disk-efficient Node.js packages)
- `uv:` - UV (Python tool manager)

The `manager:package` format is always required - there is no default manager for package operations.

---

## Track Command

Records an installed package in the lock file for management.

### Synopsis

```bash
plonk track <manager:package>...
```

### Behavior

- Verifies the package is actually installed before tracking
- Adds package to `plonk.lock` if installed
- Fails if package is not installed (install it first using the native package manager)
- Fails if manager is not supported

### Examples

```bash
# Track packages (must already be installed)
plonk track brew:ripgrep
plonk track cargo:bat go:gopls
plonk track pnpm:typescript uv:ruff

# Track multiple packages at once
plonk track brew:wget brew:jq cargo:fd
```

### Error Cases

```bash
# Package not installed - fails
plonk track brew:nonexistent-package
# Error: package 'nonexistent-package' is not installed

# Missing manager prefix - fails
plonk track ripgrep
# Error: invalid format 'ripgrep', expected manager:package

# Unsupported manager - fails
plonk track npm:typescript
# Error: unsupported manager 'npm'
```

---

## Untrack Command

Removes a package from tracking. Does NOT uninstall the package.

### Synopsis

```bash
plonk untrack <manager:package>...
```

### Behavior

- Removes package from `plonk.lock`
- Does NOT uninstall the package from the system
- Silently succeeds if package was not being tracked

### Examples

```bash
# Stop tracking packages (leaves them installed)
plonk untrack brew:ripgrep
plonk untrack cargo:bat go:gopls

# Untrack multiple packages
plonk untrack brew:wget brew:jq
```

---

## Workflow

The typical workflow for package management:

1. **Install packages using native tools:**
   ```bash
   brew install ripgrep fd bat
   cargo install tokei
   go install golang.org/x/tools/gopls@latest
   ```

2. **Track the installed packages:**
   ```bash
   plonk track brew:ripgrep brew:fd brew:bat
   plonk track cargo:tokei
   plonk track go:gopls
   ```

3. **On a new machine, apply to install missing packages:**
   ```bash
   plonk apply
   ```

## Integration

- Use `plonk status` to see tracked packages
- Use `plonk apply` to install missing tracked packages
- Lock file can be version controlled for syncing across machines
- See [Configuration Guide](../configuration.md) for timeout settings

## Notes

- The `manager:package` format is always required
- Track only works with packages that are already installed
- Untrack does not uninstall - use the native package manager to remove packages
- To upgrade packages, use the native package manager directly (e.g., `brew upgrade`)
