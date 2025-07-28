# Setup Command

The `plonk setup` command initializes plonk or clones an existing dotfiles repository.

## Description

The setup command provides two primary workflows: initializing a fresh plonk installation or cloning an existing dotfiles repository. It ensures the system has necessary package managers installed and creates required configuration files. When cloning a repository, setup automatically runs `plonk apply` to configure the system according to the cloned configuration.

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
```
1. Create $PLONK_DIR if not exists
2. Generate plonk.yaml with defaults
3. Run doctor --fix to check/install package managers
   - Prompt for each missing manager (unless --yes)
4. Complete setup (no apply needed)
```

**Clone Repository:**
```
1. Verify $PLONK_DIR doesn't exist or is empty
2. Clone repository to $PLONK_DIR
3. Run doctor --fix to check/install package managers
   - Prompt for each missing manager (unless --yes)
4. Run plonk apply to sync system state
```

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

## Improvements

- Add flags to skip specific package managers during setup (e.g., `--no-cargo`, `--no-npm`)
- Consider auto-detecting required package managers from cloned plonk.lock
- Make clone + apply operation more intelligent by only installing managers for tracked packages
- Consider separating setup into two distinct commands: `plonk init` for initial setup and `plonk setup` for clone + apply workflow
