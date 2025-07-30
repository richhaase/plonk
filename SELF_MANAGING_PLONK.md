# Self-Managing Plonk: Bootstrap and Update Solutions

## Problem Statement

Plonk is a "package manager manager" that helps users manage multiple package managers. However, it faces a bootstrapping paradox: how do you install a package manager manager without using a package manager? This document explores solutions for making plonk self-managing, including installation and updates.

## Current State Analysis

### Installation Methods
1. **Go Install** (current primary method)
   ```bash
   go install github.com/richhaase/plonk/cmd/plonk@latest
   ```
   - Requires Go 1.23+
   - Not ideal for non-Go developers
   - Depends on Go module proxy availability

2. **Clone and Build** (development method)
   ```bash
   git clone https://github.com/richhaase/plonk.git
   cd plonk
   go build -o plonk cmd/plonk/main.go
   ```
   - Requires Go and Git
   - More steps for end users

3. **Pre-built Releases** (configured but not active)
   - GoReleaser configuration exists
   - No GitHub Actions workflow yet
   - No releases published

### Version Management
- Version info injected via ldflags during build
- Fallback to Go build info for `go install` builds
- `plonk --version` shows version information
- No update checking or self-update capability

## Proposed Solutions

### Solution 1: Bootstrap Shell Script

Create an installer script that downloads pre-built binaries without dependencies.

**Features:**
- Single command installation: `curl -sSL https://get.plonk.dev | bash`
- Platform detection (OS and architecture)
- Checksum verification
- Progress indication
- Error handling and rollback
- Installation location options (`/usr/local/bin`, `~/.local/bin`, custom)

**Script Structure:**
```bash
#!/usr/bin/env bash
# install.sh - Bootstrap installer for plonk

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map to GoReleaser naming
case "$ARCH" in
  x86_64) ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Download latest release
LATEST_URL="https://api.github.com/repos/richhaase/plonk/releases/latest"
# ... implementation details ...
```

### Solution 2: Self-Update Command

Add `plonk update` command to the CLI.

**Command Structure:**
```
plonk update              # Update to latest stable version
plonk update --check      # Check for updates without installing
plonk update --force      # Force update even if current
plonk update v1.2.3       # Update to specific version
```

**Implementation Considerations:**
- Use GitHub Releases API to check versions
- Download to temporary location first
- Verify checksum before replacing
- Handle permission issues gracefully
- Show changelog between versions
- Backup current binary before update

### Solution 3: GitHub Releases Automation

Set up automated release pipeline using GitHub Actions.

**Workflow Components:**
1. **Release Trigger:**
   - On tag push (v*.*.*)
   - Manual workflow dispatch

2. **Build Matrix:**
   - OS: linux, darwin, windows
   - Arch: amd64, arm64
   - Exclude unsupported combinations

3. **Release Assets:**
   - Binary archives (tar.gz, zip)
   - checksums.txt
   - Installation script
   - Release notes

4. **Version Embedding:**
   - Use ldflags to embed version info
   - Include commit hash and build date

### Solution 4: Update Awareness

Integrate update checking into existing commands.

**Doctor Command Enhancement:**
```
$ plonk doctor
...
Plonk Version:
  Current: v1.0.0
  Latest: v1.1.0 (update available)
  Run 'plonk update' to upgrade
```

**Periodic Check:**
- Check once per day maximum
- Store last check timestamp in config
- Respect offline/airgapped environments
- Optional opt-out via config

### Solution 5: Alternative Distribution Methods

**Future Considerations:**

1. **Container Image:**
   ```bash
   docker run -v ~/.config/plonk:/config richhaase/plonk
   ```

2. **Static Binary Hosting:**
   - Host on plonk.dev domain
   - CDN distribution
   - Regional mirrors

3. **Package Manager Integration:**
   - Homebrew formula (after stable releases)
   - Snap package
   - Flatpak

## Implementation Plan

### Phase 1: Enable Releases (Week 1)
- [ ] Create GitHub Actions workflow for releases
- [ ] Test GoReleaser configuration
- [ ] Create first test release (v0.1.0-alpha)
- [ ] Verify binary artifacts work correctly

### Phase 2: Bootstrap Script (Week 2)
- [ ] Write install.sh script
- [ ] Add platform detection logic
- [ ] Implement checksum verification
- [ ] Test on multiple platforms
- [ ] Add uninstall functionality

### Phase 3: Self-Update Command (Week 3-4)
- [ ] Design update command structure
- [ ] Implement version comparison logic
- [ ] Add GitHub API integration
- [ ] Handle binary replacement safely
- [ ] Add rollback capability

### Phase 4: Integration (Week 5)
- [ ] Add update checks to doctor command
- [ ] Update documentation
- [ ] Create installation landing page
- [ ] Add telemetry for update adoption (opt-in)

## Security Considerations

1. **Checksum Verification:**
   - All downloads must verify SHA256 checksums
   - Checksums file must be signed (future)

2. **HTTPS Only:**
   - All downloads over HTTPS
   - Certificate pinning for critical endpoints

3. **Binary Signing:**
   - Consider code signing for macOS/Windows
   - GPG signatures for Linux

4. **Update Safety:**
   - Never update during active operations
   - Verify binary works before replacing
   - Keep backup of previous version

## Success Metrics

1. **Installation Success Rate:**
   - Track script completion rate
   - Monitor error reports

2. **Update Adoption:**
   - Percentage using self-update
   - Time to adopt new versions

3. **Platform Coverage:**
   - Supported OS/arch combinations
   - Installation method diversity

## Open Questions

1. Should plonk manage itself in plonk.lock?
2. How to handle update conflicts with OS package managers?
3. Should we support downgrades?
4. What about airgapped/offline environments?
5. How to handle custom builds vs official releases?

## Alternatives Considered

1. **Single Binary Go Tool:**
   - Use tools like `eget` or `bin`
   - Requires another tool to bootstrap

2. **Compile on Demand:**
   - Download source and compile
   - Requires Go toolchain

3. **WebAssembly Distribution:**
   - Platform independent
   - Performance concerns

## Conclusion

Making plonk self-managing requires a multi-faceted approach. The combination of shell script installer, self-update command, and automated releases provides a complete solution for users to install and maintain plonk without external package managers.

The phased implementation allows for iterative improvements while maintaining stability for existing users who install via `go install`.
