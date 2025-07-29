# Clone Command

The `plonk clone` command clones a dotfiles repository and intelligently sets up plonk.

For CLI syntax and flags, see [CLI Reference](../cli.md#plonk-clone).

## Description

The clone command provides an intelligent setup experience by cloning an existing dotfiles repository and automatically detecting which package managers are needed based on the repository's lock file. This eliminates the need to manually specify which tools to install - plonk figures it out for you.

## Behavior

### Primary Function

`plonk clone <git-repo>` performs the following intelligent steps:
1. Validates and parses the git repository URL
2. Clones the repository directly into `$PLONK_DIR`
3. Reads `plonk.lock` to detect required package managers
4. Installs ONLY the package managers needed by your dotfiles
5. Runs `plonk apply` to configure your system (unless --no-apply)

### Intelligent Package Manager Detection

The key feature of clone is automatic detection:
- Reads the cloned `plonk.lock` file
- Extracts package managers from resource metadata
- Only installs managers that are actually needed
- No unnecessary tool installation

If no lock file exists or it cannot be read:
- No package managers are installed
- User can manually install package managers if needed

### Repository URL Support

Accepts multiple Git repository formats:
- **GitHub shorthand**: `user/repo` (defaults to HTTPS)
- **HTTPS URL**: `https://github.com/user/repo.git`
- **SSH URL**: `git@github.com:user/repo.git`
- **Git protocol**: `git://github.com/user/repo.git`

### Command Options

- `--yes` - Non-interactive mode, automatically accepts all prompts
- `--no-apply` - Skip running `plonk apply` after setup

### Execution Flow

1. Parse and validate git repository URL
2. Verify $PLONK_DIR doesn't exist
3. Clone repository to $PLONK_DIR
4. Check for existing plonk.yaml (create default if missing)
5. Read plonk.lock and detect required package managers
6. Run health checks to find missing managers
7. Install only the required missing managers
8. Run `plonk apply` (unless --no-apply)

### Package Manager Detection

Detection process:
1. Read lock file resources
2. Filter to package type resources
3. Extract manager from metadata (v2 format)
4. Fall back to ID prefix parsing if needed
5. Build unique list of required managers

### Interactive Behavior

In interactive mode (default):
- Shows detected package managers
- Prompts before installation
- Asks for apply confirmation

With `--yes` flag:
- Automatically installs detected managers
- Automatically runs apply
- No user prompts

### Error Handling

- **Directory exists**: Fails if `$PLONK_DIR` already exists
- **Clone failures**: Cleans up and exits (bad URL, network, auth)
- **Lock file issues**: Warns but continues (no managers installed)
- **Package manager failures**: Reports but continues with others
- **Apply failures**: Reports but doesn't fail entire operation

### Examples

```bash
# Clone and auto-detect managers
plonk clone user/dotfiles

# Clone specific user's dotfiles
plonk clone richhaase/dotfiles

# Clone without running apply
plonk clone user/repo --no-apply

# Non-interactive mode
plonk clone --yes user/dotfiles

# Full URL examples
plonk clone https://github.com/user/repo.git
plonk clone git@github.com:user/repo.git
```

### Intelligence Features

**What Makes Clone Smart:**
1. **No Manual Configuration**: Unlike init, no need for skip flags
2. **Minimal Installation**: Only installs what's actually needed
3. **Lock File Aware**: Uses v2 metadata for accurate detection
4. **Graceful Degradation**: Works even without lock file

**Detection Example:**

Given a lock file with:
```yaml
resources:
  - type: package
    id: brew:git
    metadata:
      manager: brew
  - type: package
    id: npm:prettier
    metadata:
      manager: npm
```

Clone will:
- Detect that brew and npm are needed
- Check if they're already installed
- Only install the missing ones

### Cross-References

- Automatically runs `plonk apply` after setup
- Uses lock file format from `plonk install`
- Works with `plonk doctor` for package manager health checks

## Implementation Notes

The clone command provides intelligent setup through:

**Command Structure:**
- Entry point: `internal/commands/clone.go`
- Core logic: `internal/setup/setup.go` (CloneAndSetup function)
- Detection: `internal/setup/setup.go` (DetectRequiredManagers function)

**Key Features:**
- No skip flags - intelligence replaces manual control
- DetectRequiredManagers reads v2 lock format
- installDetectedManagers only installs what's needed
- Preserves repository's existing configuration files
