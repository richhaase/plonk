# Setup Command (Deprecated)

**⚠️ DEPRECATED**: The `plonk setup` command is deprecated. Use `plonk init` or `plonk clone` instead.

The `plonk setup` command initializes plonk or clones an existing dotfiles repository.

## Description

**This command is deprecated and will be removed in a future version.**
- For initializing fresh plonk configuration, use `plonk init`
- For cloning dotfiles repositories, use `plonk clone`

The setup command provides two primary workflows: initializing a fresh plonk installation or cloning an existing dotfiles repository. It ensures the system has necessary package managers installed and creates required configuration files. When cloning a repository, setup automatically runs `plonk apply` to configure the system according to the cloned configuration.

## Prerequisites

- **Git**: Required for cloning repositories (clone mode only)
- **Package Manager**: At least one supported package manager should be available
- **Empty Directory**: For clone mode, `$PLONK_DIR` must not exist or be empty
- **Network Access**: Required for cloning repositories and installing package managers

## Behavior

### Two Primary Modes

1. **Fresh Setup** (`plonk setup`)
   - Creates `$PLONK_DIR` directory structure
   - Generates default `plonk.yaml` with all configuration values
   - Runs health checks and offers to install missing package managers
   - Does NOT create `plonk.lock` (only created by install/uninstall commands)

2. **Clone Repository** (`plonk setup [git-repo]`)
   - Clones repository directly into `$PLONK_DIR`
   - Preserves existing configuration files from the repository
   - Runs health checks and offers to install missing package managers
   - Executes `plonk apply` to sync system with cloned configuration

### Repository URL Support

Accepts multiple Git repository formats:
- **GitHub shorthand**: `user/repo` (defaults to HTTPS)
- **HTTPS URL**: `https://github.com/user/repo.git`
- **SSH URL**: `git@github.com:user/repo.git`
- **Git protocol**: `git://github.com/user/repo.git`

### Package Manager Installation

Setup uses custom tool installation logic that analyzes system health and installs missing package managers:

**Bootstrap Managers** (installed via official installers):
- Homebrew (required on macOS/Linux) - uses official installer script
- Cargo/Rust (via rustup) - uses rustup installer script

**Language Package Managers** (installed via default_manager, see [Configuration Guide](../configuration.md#package-manager-settings)):
- npm (Node.js) - installs `node` package via default package manager
- pip (Python) - installs `python` package via default package manager
- gem (Ruby) - installs `ruby` package via default package manager
- go - installs `go` package via default package manager

Note: Both `plonk setup` and `plonk doctor --fix` use the same underlying tool installation logic.

### Command Options

- `--yes` - Non-interactive mode, automatically accepts all prompts
- `-o, --output` - Output format (inherited global flag, limited utility for setup)

### Execution Flow

**Fresh Setup:**
1. Verify $PLONK_DIR doesn't exist
2. Create $PLONK_DIR directory
3. Generate plonk.yaml with defaults
4. Run health checks and install missing package managers
5. Complete setup (no apply needed)

**Clone Repository:**
1. Verify $PLONK_DIR doesn't exist
2. Clone repository to $PLONK_DIR
3. Run health checks and install missing package managers
4. Automatically run plonk apply to sync system state

### Interactive Behavior

In interactive mode (default):
- Prompts before installing each missing package manager
- Shows what will be installed and asks for confirmation
- Allows users to skip optional package managers

With `--yes` flag:
- Automatically installs all missing package managers
- No user prompts or confirmations

### Error Handling

- **Directory exists**: Fails if `$PLONK_DIR` already exists
- **Clone failures**: Exits with error (bad URL, network issues, auth problems)
- **Package manager failures**:
  - Homebrew (required): Exits with error
  - Others (optional): Reports failure and continues
- **Atomic operation**: On failure, attempts to leave system in clean state

### PATH Configuration Guidance

When package managers are installed but not immediately available in PATH, setup provides detailed configuration instructions:

**Homebrew on Apple Silicon (ARM64 Macs)**:
- Installed to `/opt/homebrew/bin/brew`
- Setup shows commands to add `/opt/homebrew/bin` to PATH
- Provides shell profile instructions for bash and zsh

**Cargo/Rust Installation**:
- Installed to `~/.cargo/bin/cargo`
- Setup shows commands to add `~/.cargo/bin` to PATH
- Provides shell profile instructions for bash and zsh
- Notes that restarting shell will automatically source `~/.cargo/env`

**Shell Profile Examples**:
```bash
# For bash users
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.bashrc
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc

# For zsh users
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.zshrc
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.zshrc
```

### Special Behaviors

- Only creates configuration files during fresh setup (not when cloning)
- `plonk.lock` is never created by setup - only by install/uninstall commands
- Cloned repositories preserve their existing configuration files
- Apply is only run after successful clone, not for fresh setup

### Cross-References

- Uses same tool installation logic as `plonk doctor --fix` for health checks and package manager installation
- Automatically runs `plonk apply` after successful repository clone

## Implementation Notes

The setup command provides initialization and repository cloning through a modular architecture:

**Command Structure:**
- Entry point: `internal/commands/setup.go`
- Core logic: `internal/setup/setup.go`
- Git operations: `internal/setup/git.go`
- Tool installation: `internal/setup/tools.go`
- User prompts: `internal/setup/prompts.go`

**Key Implementation Flow:**

1. **Command Processing:**
   - Accepts optional git repository argument
   - `--yes` flag enables non-interactive mode
   - Routes to `SetupWithoutRepo()` or `SetupWithRepo()` based on arguments

2. **Repository URL Parsing:**
   - Handles: GitHub shorthand, HTTPS URLs, SSH URLs, git:// protocol
   - Auto-appends `.git` to HTTPS URLs if missing
   - Uses regex validation for GitHub shorthand format
   - GitHub shorthand defaults to HTTPS URLs

3. **Directory Management:**
   - Checks for existing plonk directory
   - Fails if `$PLONK_DIR` already exists (ensures clean clone environment)
   - Uses atomic cleanup on clone failure
   - Creates directory structure with 0750 permissions

4. **Fresh Setup Flow:**
   - Creates `$PLONK_DIR` directory
   - **DISCREPANCY**: Creates actual config file with all defaults (not just "default plonk.yaml")
   - Generates complete plonk.yaml with expanded defaults
   - **DISCREPANCY**: Does NOT create plonk.lock during setup (docs are correct)
   - Runs health checks and tool installation

5. **Clone Repository Flow:**
   - Validates and parses git URL
   - Clones directly into `$PLONK_DIR` using git command
   - Preserves existing plonk.yaml if present
   - Only creates default config if plonk.yaml missing
   - Automatically runs orchestrator.Apply() after successful clone
   - Runs orchestrator.Apply() directly (not shell command)

6. **Tool Installation:**
   - Uses `diagnostics.RunHealthChecks()` for system analysis
   - Uses same logic as `plonk doctor --fix` but called directly
   - Implements custom installation logic for each manager:
     - Homebrew: Uses official installer script
     - Cargo: Uses rustup installer script
   - Provides PATH configuration guidance for Apple Silicon and standard installs
   - Handles network connectivity checks

7. **Interactive Behavior:**
   - Uses `promptYesNo()` for user confirmations
   - `--yes` flag bypasses all prompts including apply confirmation
   - Shows detailed PATH setup instructions when tools install outside PATH

**Configuration File Creation:**
- **DISCREPANCY**: Creates complete plonk.yaml with all default values expanded
- Includes detailed comments and structure
- Uses actual default values from config.GetDefaults()
- Does NOT create plonk.lock file during setup

**Error Handling:**
- Git clone failures trigger directory cleanup
- Network connectivity validation before downloads
- Detailed error messages with command output
- Graceful handling of existing installations

**Bugs Identified:**
None - all discrepancies have been resolved.

## Improvements

- Add flags to skip specific package managers during setup (e.g., `--no-cargo`, `--no-npm`)
- Consider auto-detecting required package managers from cloned plonk.lock
- Make clone + apply operation more intelligent by only installing managers for tracked packages
- Consider separating setup into two distinct commands: `plonk init` for initial setup and `plonk setup` for clone + apply workflow
