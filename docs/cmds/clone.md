# Clone Command

Clones a dotfiles repository and intelligently sets up plonk.

## Synopsis

```bash
plonk clone [options] <git-repo>
```

## Description

The clone command provides an intelligent setup experience by cloning an existing dotfiles repository and detecting which package managers are needed based on the repository's lock file. It highlights any missing managers so you can install them manually using the install hints surfaced by `plonk doctor`.

The command clones the repository directly into `$PLONK_DIR`, detects required package managers from `plonk.lock`, and runs `plonk apply` to complete the setup.

## Options

- `--dry-run, -n` - Show what would be cloned without making changes

## Behavior

### Core Operation

`plonk clone` performs the following intelligent steps:
1. Validates and parses the git repository URL
2. Clones the repository directly into `$PLONK_DIR`
3. Reads `plonk.lock` to detect required package managers
4. Reports which managers are missing and provides installation guidance
5. Runs `plonk apply` to configure your system

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

The clone command:
- Detects required package managers from the lock file
- Skips automatic installation for security reasons, but clearly lists the missing managers so you can install them yourself
- Automatically runs apply to complete setup
- Requires no additional flags to perform detection and reporting

### Error Handling

- **Directory exists**: Fails if `$PLONK_DIR` already exists
- **Clone failures**: Cleans up and exits (bad URL, network, auth)
- **Lock file issues**: Warns but continues (no managers installed)
- **Package manager failures**: Reports but continues with others
- **Apply failures**: Reports but doesn't fail entire operation

## Examples

```bash
# Clone and detect required managers
plonk clone user/dotfiles

# Clone specific user's dotfiles
plonk clone richhaase/dotfiles

# Preview what would happen
plonk clone --dry-run user/dotfiles

# Full URL examples
plonk clone https://github.com/user/repo.git
plonk clone git@github.com:user/repo.git
```

## Integration

- Automatically runs `plonk apply` after setup
- Creates default `plonk.yaml` if missing
- Works with v2 lock file format for accurate manager detection
- Use `plonk doctor` after clone to install any missing package managers using the provided hints

## Notes

- Homebrew must be installed before using plonk (it's a prerequisite)
- Detects which package managers are needed and lists any that are missing (it does not install them automatically)
- The repository is cloned directly into `$PLONK_DIR` (no subdirectory)
