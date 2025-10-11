# Package Management Commands

Commands for managing packages: `install` and `uninstall`.

## Description

The package management commands handle system package operations across multiple package managers. All commands support package manager prefixes (e.g., `brew:htop`) to target specific managers, defaulting to the configured `default_manager` when no prefix is specified.

Package state is tracked in `plonk.lock`, which is updated atomically with each operation.

## Package Manager Prefixes

- `brew:` - Homebrew (macOS and Linux)
- `npm:` - NPM (global packages)
- `pnpm:` - PNPM (fast, disk-efficient Node.js packages)
- `cargo:` - Cargo (Rust)
- `pipx:` - Pipx (Python applications in isolated environments)
- `conda:` - Conda (scientific computing and data science packages)
- `gem:` - RubyGems
- `uv:` - UV (Python tool manager)


Without prefix, uses `default_manager` from configuration (default: brew).

---

## Install Command

Installs packages and adds them to plonk management.

### Synopsis

```bash
plonk install [options] <package>...
```

### Options

- `--dry-run, -n` - Preview changes without installing

### Behavior

**Package Installation:**
- Package managers must be available before installing packages
- Use `plonk doctor` to check for missing package managers and view installation instructions
- **Not installed** → installs package, adds to plonk.lock
- **Already installed** → adds to plonk.lock (success)
- **Already managed** → skips (no reinstall)
- Updates plonk.lock atomically with each success
- Processes multiple packages independently
- Failures don't block other installations

### Examples

```bash
# Install packages with default manager
plonk install ripgrep fd bat

# Install packages with specific managers
plonk install brew:wget npm:prettier pnpm:typescript cargo:exa pipx:black conda:numpy uv:ruff

# Preview installation
plonk install --dry-run ripgrep
```

---

## Uninstall Command

Removes packages from system and plonk management.

### Synopsis

```bash
plonk uninstall [options] <package>...
```

### Options

- `--dry-run, -n` - Preview changes without uninstalling

### Behavior

- Removes package from system and plonk.lock entry
- Only removes packages currently managed by plonk
- Dependency handling delegated to package manager
- Processes multiple packages independently
- Lock file updated atomically per operation

### Examples

```bash
# Uninstall packages
plonk uninstall ripgrep fd

# Uninstall with specific manager
plonk uninstall brew:wget npm:prettier pnpm:typescript pipx:black conda:numpy uv:ruff

# Preview removal
plonk uninstall --dry-run ripgrep
```

---


## Common Behaviors

### State Management

**Install/Uninstall:**
- Modifies `plonk.lock` atomically
- Updates system packages via manager

### Error Handling

- Individual package failures don't stop batch operations
- Summary shows succeeded/skipped/failed counts
- Package conflicts during install are considered successful
- Manager unavailability results in operation failure

### Timeout Configuration

Operations have configurable timeouts via `plonk.yaml`:
- `package_timeout` - Install/uninstall operations (default: 180s)

## Integration

- Use `plonk status` to see managed packages
- Use `plonk apply` to install all packages from lock file
- Lock file can be version controlled for team sharing
- See [Configuration Guide](../configuration.md) for timeout settings

## Notes

- Empty package names are rejected with validation errors
- Invalid managers show helpful error messages
