# Setup Command

The `plonk setup` command initializes plonk or clones an existing dotfiles repository.

## Description

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
- **GitHub shorthand**: `user/repo`
- **HTTPS URL**: `https://github.com/user/repo.git`
- **SSH URL**: `git@github.com:user/repo.git`

### Package Manager Installation

Setup delegates health checking and package manager installation to `plonk doctor --fix`:

**Bootstrap Managers** (installed via official installers):
- Homebrew (required on macOS/Linux)
- Cargo/Rust (via rustup)

**Language Package Managers** (installed via default_manager):
- npm (Node.js)
- pip (Python)
- gem (Ruby)
- go

### Command Options

- `--yes` - Non-interactive mode, automatically accepts all prompts
- `-o, --output` - Output format (inherited global flag, limited utility for setup)

### Execution Flow

**Fresh Setup:**
1. Create $PLONK_DIR if not exists
2. Generate plonk.yaml with defaults
3. Run doctor --fix to check/install package managers
4. Complete setup (no apply needed)

**Clone Repository:**
1. Verify $PLONK_DIR doesn't exist or is empty
2. Clone repository to $PLONK_DIR
3. Run doctor --fix to check/install package managers
4. Run plonk apply to sync system state

### Interactive Behavior

In interactive mode (default):
- Prompts before installing each missing package manager
- Shows what will be installed and asks for confirmation
- Allows users to skip optional package managers

With `--yes` flag:
- Automatically installs all missing package managers
- No user prompts or confirmations

### Error Handling

- **Directory exists**: Fails if `$PLONK_DIR` is not empty when cloning
- **Clone failures**: Exits with error (bad URL, network issues, auth problems)
- **Package manager failures**:
  - Homebrew (required): Exits with error
  - Others (optional): Reports failure and continues
- **Atomic operation**: On failure, attempts to leave system in clean state

### Special Behaviors

- Only creates configuration files during fresh setup (not when cloning)
- `plonk.lock` is never created by setup - only by install/uninstall commands
- Cloned repositories preserve their existing configuration files
- Apply is only run after successful clone, not for fresh setup

### Cross-References

- Uses `plonk doctor --fix` internally for health checks and package manager installation
- Runs `plonk apply` after successful repository clone

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
   - **DISCREPANCY**: Supports additional formats beyond documented ones
   - Handles: GitHub shorthand, HTTPS URLs, SSH URLs, git:// protocol
   - Auto-appends `.git` to HTTPS URLs if missing
   - Uses regex validation for GitHub shorthand format

3. **Directory Management:**
   - Checks for existing plonk files (plonk.yaml OR plonk.lock)
   - **DISCREPANCY**: Fails if ANY plonk file exists, not just "not empty"
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
   - **DISCREPANCY**: Prompts user before running apply (not automatic)
   - Runs orchestrator.Apply() directly (not shell command)

6. **Tool Installation:**
   - Uses `diagnostics.RunHealthChecks()` for system analysis
   - **DISCREPANCY**: Does NOT delegate to `plonk doctor --fix` as documented
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
1. Documentation says setup delegates to `plonk doctor --fix` but actually uses custom implementation
2. Documentation says apply runs automatically after clone, but actually prompts user first
3. Documentation says "empty directory required" but code checks for specific plonk files
4. Missing documentation of git:// protocol support
5. Documentation doesn't mention PATH configuration guidance provided

## Improvements

- Add flags to skip specific package managers during setup (e.g., `--no-cargo`, `--no-npm`)
- Consider auto-detecting required package managers from cloned plonk.lock
- Make clone + apply operation more intelligent by only installing managers for tracked packages
- Consider separating setup into two distinct commands: `plonk init` for initial setup and `plonk setup` for clone + apply workflow
