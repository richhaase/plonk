# Package Management Commands

Commands for managing packages: `install`, `uninstall`, `search`, and `info`.

## Description

The package management commands handle system package operations across multiple package managers. All commands support package manager prefixes (e.g., `brew:htop`) to target specific managers, defaulting to the configured `default_manager` when no prefix is specified.

Package state is tracked in `plonk.lock`, which is updated atomically with each operation. The v2 lock format stores both binary names and source paths for accurate reinstallation.

## Package Manager Prefixes

- `brew:` - Homebrew (macOS and Linux)
- `npm:` - NPM (global packages)
- `pnpm:` - PNPM (fast, disk-efficient Node.js packages)
- `cargo:` - Cargo (Rust)
- `pipx:` - Pipx (Python applications in isolated environments)
- `conda:` - Conda (scientific computing and data science packages)
- `gem:` - RubyGems
- `go:` - Go modules
- `uv:` - UV (Python tool manager)
- `pixi:` - Pixi (Conda-forge packages)


Without prefix, uses `default_manager` from configuration (default: brew).

---

## Install Command

Installs packages and adds them to plonk management. Also supports automatic package manager bootstrapping.

### Synopsis

```bash
plonk install [options] <package>...
```

### Options

- `--dry-run, -n` - Preview changes without installing

### Behavior

**Package Manager Self-Installation:**
- **Bare manager names** (e.g., `pnpm`, `cargo`) → triggers manager self-installation
- **Prefixed names** (e.g., `brew:npm`) → installs package normally via specified manager

**Regular Package Installation:**
- **Not installed** → installs package, adds to plonk.lock
- **Already installed** → adds to plonk.lock (success)
- **Already managed** → skips (no reinstall)
- Updates plonk.lock atomically with each success
- Processes multiple packages independently
- Failures don't block other installations

### Examples

```bash
# Bootstrap package managers (automatic detection)
plonk install pnpm cargo uv pipx

# Install packages with default manager
plonk install ripgrep fd bat

# Install packages with specific managers
plonk install brew:wget npm:prettier pnpm:typescript cargo:exa pipx:black conda:numpy uv:ruff pixi:tree

# Mixed operations (manager bootstrap + package install)
plonk install pnpm ripgrep npm:prettier

# Preview installation
plonk install --dry-run pnpm ripgrep
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
plonk uninstall brew:wget npm:prettier pnpm:typescript pipx:black conda:numpy uv:ruff pixi:tree

# Preview removal
plonk uninstall --dry-run ripgrep
```

---

## Search Command

Searches for packages across package managers.

### Synopsis

```bash
plonk search [options] <query>
```

### Options

- `-o, --output` - Output format (table/json/yaml)

### Behavior

- Without prefix: searches all managers in parallel
- With prefix: searches only specified manager
- Shows package names only (no descriptions by default)
- Uses configurable timeout (default: 5 minutes)
- Slow managers may not return results due to timeout

**Search Support by Manager:**
- **Supported**: brew, npm, cargo, conda, gem, pixi
- **Not Supported**: pnpm, pipx, uv, go (return empty results)

### Examples

```bash
# Search all managers
plonk search ripgrep

# Search specific manager (only supported managers)
plonk search brew:ripgrep conda:numpy pixi:tree

# Managers without search support return empty results
plonk search uv:ruff          # No results (UV doesn't support search)
plonk search pnpm:typescript  # No results (PNPM doesn't support search)
plonk search go:test          # No results (Go doesn't support search)

# Output as JSON
plonk search -o json ripgrep
```

---

## Info Command

Shows detailed package information and installation status.

### Synopsis

```bash
plonk info [options] <package>
```

### Options

- `-o, --output` - Output format (table/json/yaml)

### Behavior

Priority order for information:
1. Managed by plonk (shows from lock file)
2. Installed but not managed
3. Available but not installed

Displays:
- Package name and status
- Manager and version
- Description and homepage
- Installation command

### Examples

```bash
# Get package info
plonk info ripgrep

# Info for specific manager
plonk info brew:ripgrep pipx:black conda:numpy uv:ruff pixi:tree

# Output as JSON
plonk info -o json ripgrep
```

---

## Common Behaviors

### State Management

**Install/Uninstall:**
- Modifies `plonk.lock` atomically
- Updates system packages via manager
- Lock file preserves full module paths for Go packages

**Search/Info:**
- Read-only operations
- No state modifications
- Query package managers directly

### Error Handling

- Individual package failures don't stop batch operations
- Summary shows succeeded/skipped/failed counts
- Package conflicts during install are considered successful
- Manager unavailability results in operation failure

### Timeout Configuration

Operations have configurable timeouts via `plonk.yaml`:
- `package_timeout` - Install/uninstall operations (default: 180s)
- `operation_timeout` - Search operations (default: 300s / 5 minutes)

## Integration

- Use `plonk status` to see managed packages
- Use `plonk apply` to install all packages from lock file
- Lock file can be version controlled for team sharing
- See [Configuration Guide](../configuration.md) for timeout settings

## Notes

- Empty package names are rejected with validation errors
- Invalid managers show helpful error messages
- Go packages store both binary name and full source path in v2 lock format
- Network timeouts are handled gracefully in search operations
