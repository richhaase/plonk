# Clone Command

Clones a dotfiles repository and intelligently sets up plonk.

## Synopsis

```bash
plonk clone [options] <git-repo>
```

## Description

The clone command provides an intelligent setup experience by cloning an existing dotfiles repository and automatically detecting which package managers are needed based on the repository's lock file. This eliminates the need to manually specify which tools to install - plonk figures it out for you.

The command clones the repository directly into `$PLONK_DIR`, detects required package managers from `plonk.lock`, installs only what's needed, and optionally runs `plonk apply` to complete the setup.

## Options

- `--no-apply` - Skip running `plonk apply` after setup

## Behavior

### Core Operation

`plonk clone` performs the following intelligent steps:
1. Validates and parses the git repository URL
2. Clones the repository directly into `$PLONK_DIR`
3. Reads `plonk.lock` to detect required package managers
4. Installs ONLY the package managers needed by your managed packages
5. Runs `plonk apply` to configure your system (unless --no-apply)

### Repository URL Support

Accepts multiple Git repository formats:
- **GitHub shorthand**: `user/repo` (defaults to HTTPS)
- **HTTPS URL**: `https://github.com/user/repo.git`
- **SSH URL**: `git@github.com:user/repo.git`
- **Git protocol**: `git://github.com/user/repo.git`

### Package Manager Detection

The key feature of clone is automatic detection:
- Reads the cloned `plonk.lock` file
- Extracts package managers from resource metadata
- Only installs managers that are actually needed
- No unnecessary tool installation

If no lock file exists or it cannot be read:
- No package managers are installed
- User can manually install package managers if needed

### Automated Behavior

The clone command operates fully automatically:
- Detects required package managers from the lock file
- Automatically installs detected managers using official installation methods
- Automatically runs apply to complete setup (unless --no-apply)
- No user interaction required - completely hands-off operation

### Error Handling

- **Directory exists**: Fails if `$PLONK_DIR` already exists
- **Clone failures**: Cleans up and exits (bad URL, network, auth)
- **Lock file issues**: Warns but continues (no managers installed)
- **Package manager failures**: Reports but continues with others
- **Apply failures**: Reports but doesn't fail entire operation

## Examples

```bash
# Clone and auto-detect managers
plonk clone user/dotfiles

# Clone specific user's dotfiles
plonk clone richhaase/dotfiles

# Clone without running apply
plonk clone user/repo --no-apply

# Clone with manual apply step
plonk clone user/dotfiles --no-apply

# Full URL examples
plonk clone https://github.com/user/repo.git
plonk clone git@github.com:user/repo.git
```

## Integration

- Automatically runs `plonk apply` after setup (unless --no-apply)
- Uses `plonk doctor` health checks to verify package managers
- Creates default `plonk.yaml` if missing
- Works with v2 lock file format for accurate manager detection

## Notes

- Homebrew must be installed before using plonk (it's a prerequisite)
- Only installs package managers that are actually needed by your managed packages
- The repository is cloned directly into `$PLONK_DIR` (no subdirectory)
