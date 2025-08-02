# Plonk Development Context

## Current Phase: Ready for v1.0 Release (2025-08-02)

### Status
All critical features, bug fixes, quality assurance tasks, and distribution improvements are complete. The codebase is ready for v1.0 release.

### v1.0 Release Checklist
- [x] Remove stability warning from README.md
- [x] Update version from "dev" to "1.0.0" (via build injection)
- [ ] Create and push v1.0.0 tag (pending final manual checks)
- [x] Verify GitHub Actions release workflow (v0.9.5 tested with signing/notarization)
- [x] Update installation docs with release download links (Homebrew support added)

### Completed Pre-v1.0 Work

#### Quality Assurance Phase ✅
1. **Test Coverage Improvement** - Achieved 61.7% coverage for packages/ (exceeded 60% target)
2. **Code Complexity Review** - Analyzed codebase, determined acceptable for v1.0
3. **Critical Documentation Review** - Fixed all outdated command references
4. **Build System Review** - Cleaned up Justfile and GitHub Actions
5. **Security Updates** - Updated Go to 1.24.5 to fix vulnerabilities

#### Feature Implementation Phase ✅
1. **Dotfile drift detection** - SHA256-based comparison with `plonk diff`
2. **Linux support** - Full parity via Homebrew on Linux
3. **Progress indicators** - Spinner-based feedback for long operations
4. **.plonk/ directory exclusion** - Reserved for future metadata
5. **All critical bug fixes** - Linux testing revealed and fixed all issues

#### Distribution Improvements ✅
1. **Homebrew Support** - Added homebrew-tap repository with Cask support
2. **Apple Code Signing** - Implemented Developer ID signing for macOS binaries
3. **Notarization** - Automated notarization via App Store Connect API
4. **Windows Builds Removed** - Focused on supported platforms (macOS/Linux)

### Post-v1.0 Roadmap

#### High Priority
- Test consolidation (hybrid BATS + Go integration approach)
- Package update command (`plonk update`)
- Verbose/debug modes

#### Medium Priority
- Code complexity reduction (extract common patterns)
- Coverage enforcement (ratcheting or minimum threshold)
- Enhanced error messages with remediation

#### Future Considerations
- Hook system using .plonk/hooks/
- Native Windows support (beyond WSL)
- Performance optimizations

## Critical Implementation Guidelines

### STRICT RULE: No Unauthorized Features
**NEVER independently add features or enhancements that were not explicitly requested.**
- You MAY propose improvements, but that is all
- Do NOT implement anything beyond the exact scope requested
- Do NOT add "helpful" extras without explicit approval
- When in doubt, implement ONLY what was explicitly requested

### UI/UX Guidelines
- **NEVER use emojis in plonk output** - Use colored text status indicators instead
- Status indicators should be colored minimally (only the status word, not full lines)
- Professional, clean output similar to tools like git, docker, kubectl

## Key Learnings

### Implementation Patterns
- **Lock File v2**: Breaking changes can lead to cleaner implementations
- **Metadata Design**: Using `map[string]interface{}` provides flexibility
- **Zero-Config Philosophy**: Sometimes removal is the best solution (init command)
- **State Model**: Successfully repurposed StateDegraded for drift detection
- **Path Resolution**: Clear separation between source (no dot) and deployed (with dot)

### Testing Philosophy
- Unit tests for business logic only, no mocks for CLIs
- Integration tests in CI only to protect developer systems
- BATS tests for behavioral validation

### UI/UX Philosophy
- **No emojis ever**: Professional colored text only
- **Minimal colorization**: Only status words colored
- **Professional appearance**: Similar to git, docker, kubectl

### Bug Fix Learnings
- **Test directories carefully**: Special handling can cause unexpected bugs
- **Error propagation matters**: Always show actual errors to users
- **Flag cleanup**: Remove non-functional flags to avoid confusion
- **Filter scope**: UnmanagedFilters should only apply to unmanaged file discovery

## Technical Details

### System Requirements
- **Go**: 1.23+ (required by tool dependencies)
- **Platforms**: macOS, Linux (including WSL)
- **Prerequisites**: Homebrew, Git

### Test Coverage Status
- **packages/**: 61.7% (up from 21.7%)
- **Key achievement**: Unit testing without system modification via Command Executor Interface
- **Future goal**: Consider coverage enforcement post-v1.0

### Build System
- **Justfile**: Development workflow automation
- **GitHub Actions**: CI/CD with matrix testing (Go 1.23, 1.24 on Ubuntu, macOS)
- **Release Process**: Automated via GoReleaser on tag push
  - Automatic macOS binary signing with Developer ID certificate
  - Apple notarization for Gatekeeper approval
  - Homebrew Cask generation and publishing to richhaase/homebrew-tap
  - Multi-platform archives (tar.gz) with checksums

### Known Limitations
- No native Windows support (use WSL)
- No package update command yet (use uninstall/install)
- Basic error messages (enhancement planned)
