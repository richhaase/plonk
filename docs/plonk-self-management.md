# Plonk Self-Management

**Status**: 🚧 **IN DEVELOPMENT** - Implementation planned
**Priority**: High
**Target Release**: v1.2.0

## Overview

Plonk self-management provides a streamlined bootstrap and maintenance experience for seasoned users. Instead of the traditional "install package manager → install plonk → use plonk" flow, self-management enables a direct "install plonk → use plonk" experience.

This feature maintains plonk's core philosophy as a "package manager manager" - it's about bootstrapping, not replacing the underlying package managers.

## Motivation

For experienced plonk users setting up new development environments, the fastest possible path is:

```bash
# Traditional flow
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew install richhaase/tap/plonk
plonk clone username/dotfiles

# Self-managed flow
curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash
plonk clone username/dotfiles
```

While the difference is minor, eliminating the intermediate Homebrew step provides value for rapid environment setup, especially in automated scenarios, containers, or when setting up multiple machines.

## Core Principles

### 1. Bootstrap Only
Self-management is about getting plonk onto a system quickly. It does not:
- Install other package managers
- Replace plonk's delegation to existing package managers
- Change plonk's architectural role as a coordinator

### 2. Maintain Package Manager Integration
Plonk continues to delegate all package operations to appropriate package managers (brew, npm, cargo, etc.). Self-management only affects how plonk itself is installed and updated.

### 3. Coexist with Traditional Installation
Users can freely switch between installation methods:
- Install via script, update via Homebrew
- Install via Homebrew, update via self-update
- Mix and match as needed

## Features

### Installation Script

**Command**: `curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash`

**Functionality**:
- Detects platform (macOS/Linux, Intel/ARM)
- Downloads latest plonk binary from GitHub releases
- Installs to `/usr/local/bin/plonk`
- Verifies installation with version check
- **Scope**: Only installs plonk binary - no package managers, no automatic clone

**Platform Support**:
- macOS (Intel and Apple Silicon)
- Linux (x86_64 and ARM64)
- Windows support via WSL2

### Self-Update Command

**Command**: `plonk self-update`

**Functionality**:
- Queries GitHub releases API for latest version
- Compares against currently installed version
- Downloads and verifies new binary if update available
- Replaces current installation atomically
- Provides `--dry-run` option to preview updates

**Update Process**:
1. Check GitHub API for latest release
2. Compare versions (semantic version comparison)
3. Download binary and checksums.txt
4. Verify SHA256 checksum for integrity
5. Replace current binary
6. Verify new installation

**Safety Features**:
- SHA256 checksum verification against official checksums.txt
- Fail-fast on download errors
- Version verification after update
- Clear error messages for troubleshooting

### Self-Uninstall Command

**Command**: `plonk self-uninstall`

**Functionality**:
- Removes plonk binary from system
- Removes `~/.config/plonk/` directory completely
- Complete cleanup approach

**What Gets Removed**:
- Plonk executable binary
- Configuration directory (`~/.config/plonk/`)
- Lock file (`plonk.lock`)
- Stored dotfile copies

**What Gets Preserved**:
- **User's actual dotfiles** - These are copied to `$HOME`, not symlinked
- **Installed packages** - Package managers retain all installed software
- **Package managers themselves** - Homebrew, npm, etc. remain installed

**Data Recovery**:
- Dotfiles remain in their deployed locations (`~/.zshrc`, `~/.vimrc`, etc.)
- Packages remain installed and functional
- Lock file can be recreated with `plonk status` after reinstallation

## Security Considerations

### Checksum Verification
All downloads are verified using SHA256 checksums:
- Each GitHub release includes `checksums.txt` with SHA256 hashes
- Installation script and self-update verify binary integrity
- Protection against corrupted downloads and basic tampering

### Code Signing
macOS binaries are code-signed and notarized:
- Eliminates security warnings on macOS
- Provides cryptographic verification of binary authenticity
- Automated via GoReleaser configuration

### Future: GPG Signing
If user demand justifies the additional complexity:
- GPG signatures for releases
- Enhanced cryptographic verification
- Protection against sophisticated attacks

## Use Cases

### 1. Rapid Environment Setup
**Scenario**: Setting up a new development machine or VM
```bash
# Single-command bootstrap
curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash
plonk clone myusername/dotfiles
# Fully configured development environment
```

### 2. Container Environments
**Scenario**: Building development containers or CI environments
```dockerfile
# Dockerfile
RUN curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash
RUN plonk clone organization/standard-config
```

### 3. Automated Provisioning
**Scenario**: Infrastructure-as-code or automated machine setup
```bash
#!/bin/bash
# Bootstrap script for new team members
curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash
plonk clone company/developer-setup
```

### 4. Maintenance Operations
**Scenario**: Keeping plonk updated without package manager dependency
```bash
plonk self-update  # Update plonk itself
# Continue using plonk for package management
```

## Installation Method Comparison

| Method | Install Command | Update Command | Pros | Cons |
|--------|----------------|----------------|------|------|
| **Homebrew** | `brew install richhaase/tap/plonk` | `brew upgrade plonk` | Package manager integration, automatic dependency resolution | Requires Homebrew first |
| **Install Script** | `curl ... \| bash` | `plonk self-update` | Direct installation, no prerequisites | Manual checksum verification |
| **Go Install** | `go install github.com/...` | `go install github.com/...@latest` | Simple for Go developers | Requires Go toolchain |
| **Manual Binary** | Download from releases | `plonk self-update` | Full control, offline capable | Manual process |

## Technical Implementation

### Architecture Overview
```
plonk self-update/self-uninstall commands
├── internal/commands/self_update.go      # Command implementation
├── internal/commands/self_uninstall.go   # Command implementation
├── internal/self/                        # Shared utilities
│   ├── github.go                         # GitHub API client
│   ├── integrity.go                      # Checksum verification
│   ├── platform.go                       # Platform detection
│   └── binary.go                         # Binary management
└── install.sh                            # Installation script
```

### GitHub Integration
- Uses GitHub Releases API for version discovery
- Downloads from `https://github.com/richhaase/plonk/releases/latest/download/`
- Verifies checksums against `checksums.txt` from same release
- Follows GitHub API best practices (user-agent, error handling)

### Platform Detection
Supports the same platforms as current plonk releases:
- macOS: Darwin_amd64, Darwin_arm64
- Linux: Linux_amd64, Linux_arm64
- Binary naming: `plonk_${OS}_${ARCH}.tar.gz`

### Error Handling
- Network errors: Clear messages with retry suggestions
- Checksum failures: Security-focused error messages
- Permission errors: Guidance on installation locations
- Version parsing: Robust semantic version handling

## Migration Path

### From Other Installation Methods
Users can seamlessly switch to self-management:

**From Homebrew**:
```bash
# Current state: installed via brew
brew uninstall plonk

# Switch to self-management
curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash
plonk self-update  # Future updates
```

**From Go Install**:
```bash
# Remove go-installed version
rm $(go env GOPATH)/bin/plonk

# Switch to self-management
curl -fsSL https://raw.githubusercontent.com/richhaase/plonk/main/install.sh | bash
```

### To Other Installation Methods
Self-managed installations can switch back:
```bash
# Remove self-managed version
plonk self-uninstall

# Install via Homebrew
brew install richhaase/tap/plonk
```

## Documentation Updates Required

### Primary Documentation
- **README.md**: Add installation script option to installation methods
- **docs/installation.md**: New section for self-management installation
- **docs/CLI.md**: Document `self-update` and `self-uninstall` commands

### Command Documentation
- **docs/cmds/self-update.md**: Detailed self-update command reference
- **docs/cmds/self-uninstall.md**: Detailed self-uninstall command reference

## Future Enhancements

### Phase 2 Improvements
- **Rollback capability**: `plonk self-update --rollback` to previous version
- **Version pinning**: `plonk self-update --version v1.1.0` for specific versions
- **Update channels**: Beta/stable release channels

### Integration Enhancements
- **Shell completion**: Tab completion for self-management commands
- **Configuration options**: Update check intervals, preferred installation location
- **Telemetry integration**: Track self-update usage (with user consent)

### Advanced Security
- **GPG signature verification**: If user demand justifies complexity
- **Supply chain security**: SLSA attestations, reproducible builds
- **Vulnerability scanning**: Automated security scanning of releases

## Risks and Mitigations

### Security Risks
**Risk**: Malicious script execution via curl | bash
**Mitigation**:
- Encourage users to inspect scripts before execution
- Provide checksummed script downloads as alternative
- Clear documentation about security considerations

**Risk**: Man-in-the-middle attacks on downloads
**Mitigation**:
- HTTPS for all downloads
- SHA256 checksum verification
- Future GPG signing if needed

### Operational Risks
**Risk**: GitHub API rate limiting or outages
**Mitigation**:
- Graceful handling of API failures
- Fallback to manual installation instructions
- Cached version information where appropriate

**Risk**: Binary corruption or availability issues
**Mitigation**:
- Multiple download mirrors (future enhancement)
- Comprehensive checksum verification
- Clear error messages with manual recovery steps

## Success Criteria

### Functional Requirements
- ✅ Installation script successfully installs plonk on supported platforms
- ✅ `plonk self-update` correctly identifies and applies updates
- ✅ `plonk self-uninstall` completely removes plonk while preserving user data
- ✅ SHA256 checksum verification prevents corrupted installations
- ✅ Commands integrate seamlessly with existing plonk CLI patterns

### User Experience Requirements
- ✅ Bootstrap flow reduces setup time for experienced users
- ✅ Self-update provides clear feedback about available updates
- ✅ Error messages provide actionable guidance for common issues
- ✅ Documentation clearly explains when to use each installation method

### Technical Requirements
- ✅ Compatible with existing plonk architecture and patterns
- ✅ Maintains separation between plonk installation and package management
- ✅ Robust error handling for network, permission, and platform issues
- ✅ Follows security best practices for self-updating software

## Conclusion

Plonk self-management provides a valuable enhancement for experienced users while maintaining the tool's core philosophy and architecture. By enabling direct installation and self-maintenance, it reduces friction in the critical "new environment setup" workflow that plonk is designed to streamline.

The implementation maintains plonk's principles of simplicity, security, and focused functionality while adding meaningful convenience for power users and automated environments.
