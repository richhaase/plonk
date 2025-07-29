# Init Command

The `plonk init` command initializes a fresh plonk configuration.

For CLI syntax and flags, see [CLI Reference](../cli.md#plonk-init).

## Description

The init command sets up plonk for first-time use by creating the configuration directory, generating default configuration files, and installing required package managers. This command gives you full control over which package managers to install through skip flags.

## Behavior

### Primary Function

`plonk init` performs the following steps:
1. Creates `$PLONK_DIR` directory structure if it doesn't exist
2. Generates default `plonk.yaml` with all configuration values
3. Creates an empty `plonk.lock` file (v2 format)
4. Runs health checks to detect missing package managers
5. Installs missing package managers (respecting skip flags)

### Package Manager Control

By default, init attempts to install all missing package managers. You can control this behavior with skip flags:

- `--no-homebrew` - Skip Homebrew installation
- `--no-cargo` - Skip Cargo/Rust installation
- `--no-npm` - Skip npm/Node.js installation
- `--no-pip` - Skip pip/Python installation
- `--no-gem` - Skip gem/Ruby installation
- `--no-go` - Skip Go installation
- `--all` - Explicitly install all managers (default behavior)

### Package Manager Categories

**Bootstrap Managers** (installed via official installers):
- Homebrew (macOS/Linux) - uses official installer script
- Cargo/Rust - uses rustup installer script

**Language Package Managers** (installed via default_manager):
- npm (Node.js) - installs `node` package
- pip (Python) - installs `python` package
- gem (Ruby) - installs `ruby` package
- go - installs `go` package

### Command Options

- `--yes` - Non-interactive mode, automatically accepts all prompts
- `--no-<manager>` - Skip installation of specific package manager
- `--all` - Install all package managers (explicit default)

### Execution Flow

1. Check if $PLONK_DIR already exists (fail if it does)
2. Create $PLONK_DIR directory with proper permissions
3. Generate plonk.yaml with expanded default values
4. Create empty plonk.lock file (v2 format)
5. Run health checks to identify missing package managers
6. Filter out managers marked for skipping
7. Prompt to install remaining missing managers (or auto-install with --yes)
8. Install each approved package manager
9. Provide PATH configuration guidance if needed

### Interactive Behavior

In interactive mode (default):
- Shows list of missing package managers
- Prompts for confirmation before installation
- Allows user to decline installation

With `--yes` flag:
- Automatically installs all non-skipped managers
- No user prompts or confirmations

### Error Handling

- **Directory exists**: Exits with message if `$PLONK_DIR` already exists
- **Package manager failures**:
  - Reports individual failures but continues with others
  - Shows manual installation instructions for failed managers
  - Partial success is not treated as failure

### PATH Configuration

When package managers install outside PATH, init provides configuration instructions:

```bash
# Homebrew on Apple Silicon
export PATH="/opt/homebrew/bin:$PATH"

# Cargo/Rust
export PATH="$HOME/.cargo/bin:$PATH"
```

### Examples

```bash
# Initialize with all package managers
plonk init

# Skip specific managers
plonk init --no-cargo --no-gem

# Non-interactive mode
plonk init --yes

# Skip all optional managers except npm
plonk init --no-cargo --no-pip --no-gem --no-go
```

### Cross-References

- Related to `plonk clone` for repository-based setup
- Uses same health check system as `plonk doctor`
- Package installations tracked in `plonk.lock`

## Implementation Notes

The init command provides controlled initialization through:

**Command Structure:**
- Entry point: `internal/commands/init.go`
- Core logic: `internal/setup/setup.go` (InitializeNew function)
- Reuses setup package infrastructure

**Key Features:**
- Skip flags implemented as SkipManagers struct
- Filters missing managers before installation
- Preserves existing setup error handling and PATH guidance
- Creates v2 format lock file by default
