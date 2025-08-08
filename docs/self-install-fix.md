# Self-Installation Independence Audit and Fix Plan

**Status**: Planning
**Priority**: Medium
**Estimated Effort**: 4-6 days
**Target**: Maximize package manager installation independence

## Executive Summary

This plan addresses unnecessary dependencies in package manager self-installation methods. Currently, 5 out of 11 package managers depend on other package managers (primarily Homebrew) when more independent installation methods are available. The goal is to make each manager use the most independent installation method available while maintaining reliability.

## Background Research

### Current State Analysis

A comprehensive audit of all 11 package manager `SelfInstall()` methods revealed the following dependency patterns:

#### ✅ Truly Independent Managers (5/11)
1. **Homebrew** - Uses standalone installer script
2. **Cargo** - Uses rustup installer script
3. **UV** - Uses standalone installer script
4. **Pixi** - Uses standalone installer script
5. **Composer** - Uses hash-verified installer (requires PHP runtime only)

#### ⚠️ Dependent Managers (5/11)
6. **NPM** - Depends on Homebrew (`brew install node`)
7. **Gem** - Depends on Homebrew (`brew install ruby`)
8. **.NET Tools** - Depends on Homebrew (`brew install dotnet`)
9. **Pipx** - Depends on pip3 OR Homebrew
10. **Go Install** - Platform-dependent (Homebrew on macOS, system pkg mgrs on Linux)

#### 🔍 Current Dependency Hierarchy
```
Homebrew (independent)
├── NPM (Node.js via Homebrew)
├── Gem (Ruby via Homebrew)
├── .NET Tools (SDK via Homebrew)
└── Pipx (fallback to Homebrew)

Go Install (mixed - system package managers)
```

### Problem Statement

**Issue**: Several package managers unnecessarily depend on Homebrew when independent installation methods exist.

**Impact**:
- Creates artificial dependency hierarchy
- Reduces installation success rate on systems without Homebrew
- Complicates `plonk clone` automation
- Violates principle of maximum independence

## Objective

**Transform package managers to use a SINGLE, most independent installation method available. Each package manager should have exactly one predictable installation method, treating each manager as a first-class citizen with no fallback dependencies on other package managers.**

## Detailed Manager Analysis

### 1. NPM (Node.js) - HIGH PRIORITY

**Current State**: Depends on Homebrew
```go
// Current implementation
if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
    return n.installViaHomebrew(ctx)
}
return fmt.Errorf("npm requires Node.js installation - install Node.js manually...")
```

**Independent Alternatives Available**:
- **Official Installer**: https://nodejs.org/en/download/package-manager/
- **Node Version Manager (nvm)**: `curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash`
- **Binary Downloads**: Direct from nodejs.org with platform detection
- **Snap**: `snap install node --classic` (Linux)

**Recommendation**: Use ONLY nvm installer script - no fallbacks

### 2. Gem (Ruby) - MEDIUM PRIORITY

**Current State**: Depends on Homebrew
```go
if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
    return g.installViaHomebrew(ctx)
}
```

**Independent Alternatives Available**:
- **rbenv**: `curl -fsSL https://github.com/rbenv/rbenv-installer/raw/HEAD/bin/rbenv-installer | bash`
- **RVM**: `curl -sSL https://get.rvm.io | bash`
- **ruby-build**: Standalone Ruby compilation
- **System packages**: Already used on Linux via system package managers

**Recommendation**: Use ONLY rbenv installer script - no fallbacks

### 3. .NET Tools - HIGH PRIORITY

**Current State**: Depends on Homebrew
```go
if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
    return d.installViaHomebrew(ctx)
}
```

**Independent Alternatives Available**:
- **Official Install Script**: `curl -sSL https://dot.net/v1/dotnet-install.sh | bash`
- **Windows PowerShell**: `winget install Microsoft.DotNet.SDK.8`
- **Snap**: `snap install dotnet-sdk --classic` (Linux)
- **Direct Download**: From Microsoft with platform detection

**Recommendation**: Use ONLY official dotnet-install.sh script - no fallbacks

### 4. Pipx - LOW PRIORITY

**Current State**: Already has good independence
```go
// Try pip3 first (good)
if pipAvailable, _ := checkPackageManagerAvailable(ctx, "pip3"); pipAvailable {
    return p.installViaPip(ctx)
}
// Falls back to Homebrew
if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
    return p.installViaHomebrew(ctx)
}
```

**Assessment**: Current implementation should be simplified to SINGLE method.

**Recommendation**: Use ONLY pip3 installation - no Homebrew fallback. pip3 is the logical and correct dependency for a Python package manager.

### 5. Go Install - MEDIUM PRIORITY

**Current State**: Platform-specific with reasonable approach
```go
switch runtime.GOOS {
case "darwin":
    return g.installMacOS(ctx) // Uses Homebrew
case "linux":
    return g.installLinux(ctx) // Uses system package managers
default:
    return fmt.Errorf("unsupported platform...")
}
```

**Independent Alternatives Available**:
- **Official Installer**: `curl -L https://go.dev/dl/go1.21.5.linux-amd64.tar.gz`
- **g (Go Version Manager)**: `curl -sSL https://git.io/g-install | sh`
- **Binary Installation**: Direct download and PATH setup

**Recommendation**: Use ONLY official Go installer script - no package manager dependencies

## Implementation Plan

### Phase 1: Research and Validation (1 day)

#### Task 1.1: Verify Independent Installation Methods
For each target manager, verify that independent installation methods:
- ✅ Work across target platforms (macOS, Linux)
- ✅ Are maintained and reliable
- ✅ Don't require root/admin privileges where possible
- ✅ Handle PATH configuration correctly

#### Task 1.2: Review Existing Patterns
Study successful independent installers (Homebrew, Cargo, UV, Pixi) to identify common patterns:
- Script download and execution methods
- Error handling approaches
- PATH management
- Verification steps

### Phase 2: Design New SelfInstall Methods (1 day)

#### Task 2.1: Define Single Installation Method
For each manager, define the ONE installation method:

**NPM Example**:
- **Single Method**: nvm installer script only
- **No Fallbacks**: If nvm installation fails, return clear error with manual instructions
- **Predictable**: Users always know exactly how NPM will be installed

#### Task 2.2: Design Helper Functions
Create reusable helper functions:
```go
// Download and execute installation script with verification
func executeInstallScript(ctx context.Context, scriptURL, name string) error

// Download and install binary with platform detection
func installBinaryWithPlatformDetection(ctx context.Context, baseURL, name string) error

// Verify installation success and PATH configuration
func verifyInstallationSuccess(ctx context.Context, binary, name string) error
```

#### Task 2.3: Update Health Check Suggestions
Modify `CheckHealth()` methods to suggest new installation methods in help text.

### Phase 3: Implementation (2-3 days)

#### Task 3.1: NPM Manager (Priority 1)
- Implement nvm installer as primary method
- Add fallback to official Node.js installer
- Keep Homebrew as final fallback
- Update tests and documentation

#### Task 3.2: .NET Tools Manager (Priority 1)
- Implement official dotnet-install.sh script
- Add Windows PowerShell method
- Keep Homebrew as fallback
- Update tests and documentation

#### Task 3.3: Gem Manager (Priority 2)
- Implement rbenv installer as primary method
- Add RVM as secondary option
- Keep Homebrew as fallback
- Update tests and documentation

#### Task 3.4: Go Install Manager (Priority 2)
- Add official Go installer method
- Keep platform-specific package managers as fallbacks
- Improve Windows support
- Update tests and documentation

### Phase 4: Testing and Validation (1 day)

#### Task 4.1: Cross-Platform Testing
Test new installation methods on:
- ✅ macOS (with and without Homebrew)
- ✅ Linux (Ubuntu, CentOS, Alpine)
- ✅ Systems without existing package managers

#### Task 4.2: BATS Integration Testing
- Add BATS tests for `plonk clone` scenarios with new installation methods
- Verify health checks provide accurate guidance via BATS behavioral tests
- Test single-method installation behavior in real CLI scenarios

#### Task 4.3: Error Handling Validation
- Ensure graceful degradation when scripts fail
- Verify error messages are helpful and actionable
- Test timeout and cancellation behavior

## Expected Outcomes

### Before/After Comparison

| Manager | Current Method | New Primary Method | Independence Level |
|---------|---------------|-------------------|-------------------|
| NPM | Homebrew | nvm installer | Independent ✅ |
| .NET | Homebrew | Official installer | Independent ✅ |
| Gem | Homebrew | rbenv installer | Independent ✅ |
| Go | Platform-mixed | Official installer | Independent ✅ |
| Pipx | pip3 → Homebrew | pip3 → pipx installer | Improved ✅ |

### Success Metrics
- **Reduced Dependencies**: 4+ managers become truly independent
- **Improved Success Rate**: Higher installation success on diverse systems
- **Better User Experience**: More consistent installation behavior
- **Enhanced Reliability**: Less dependency on external package managers

## Implementation Considerations

### Security Considerations
- ✅ Verify script signatures where available
- ✅ Use HTTPS for all downloads
- ✅ Validate checksums when provided by official sources
- ✅ Follow principle of least privilege

### Backwards Compatibility
- ✅ Keep existing fallback methods for compatibility
- ✅ Maintain current error message format
- ✅ Ensure no breaking changes to public interfaces

### Error Handling
- ✅ Clear error messages explaining what failed and why
- ✅ Suggest manual installation steps when auto-install fails
- ✅ Proper context cancellation and timeout handling

## Risk Assessment

### Technical Risks: LOW-MEDIUM
- **Mitigation**: Independent installers are maintained by official projects
- **Mitigation**: Keep existing Homebrew fallbacks for compatibility
- **Mitigation**: Comprehensive testing on multiple platforms

### Compatibility Risks: LOW
- **Mitigation**: Phased rollout with existing methods as fallbacks
- **Mitigation**: No changes to public interfaces
- **Mitigation**: Backwards compatibility preserved

### Security Risks: MEDIUM
- **Mitigation**: Use official installation scripts only
- **Mitigation**: HTTPS verification and script validation
- **Mitigation**: Follow security best practices from existing implementations

## Files to Modify

### Primary Implementation Files
```
internal/resources/packages/npm.go          # NPM self-installation
internal/resources/packages/dotnet.go       # .NET self-installation
internal/resources/packages/gem.go          # Gem self-installation
internal/resources/packages/goinstall.go    # Go self-installation
internal/resources/packages/pipx.go         # Pipx improvements
```

### Helper Files
```
internal/resources/packages/install_helpers.go  # Shared installation utilities
```

### Test Files
```
internal/resources/packages/npm_test.go     # Updated tests
internal/resources/packages/dotnet_test.go  # Updated tests
internal/resources/packages/gem_test.go     # Updated tests
internal/resources/packages/goinstall_test.go # Updated tests
```

### Documentation Files
```
docs/package-managers.md    # Updated installation documentation
```

## Success Criteria

### Functional Requirements
- ✅ Each manager uses most independent installation method available
- ✅ Fallback mechanisms preserved for compatibility
- ✅ All managers can self-install on clean systems
- ✅ Cross-platform functionality maintained or improved
- ✅ Error handling provides clear guidance

### Non-Functional Requirements
- **Reliability**: Installation success rate > 95% on supported platforms
- **Performance**: Installation completes within reasonable timeouts
- **Security**: All downloads use HTTPS and official sources
- **Maintainability**: Code follows existing patterns and conventions

## Future Considerations

### Post-Implementation Enhancements
1. **Usage Analytics**: Track which installation methods are most successful
2. **Platform Expansion**: Add support for additional platforms (Windows, BSD)
3. **Version Management**: Consider integration with version managers (nvm, rbenv, etc.)
4. **Caching**: Cache downloaded installers for faster subsequent installations

### Long-term Goals
1. **Complete Independence**: Achieve 100% independent installation capability
2. **Smart Fallbacks**: Machine learning-based selection of optimal installation method
3. **Containerized Installation**: Support for container-based isolation
4. **Air-gapped Installation**: Support for offline installation scenarios

## Conclusion

This audit and fix plan addresses a significant architectural improvement opportunity. By maximizing installation independence, we enhance plonk's reliability, reduce complexity, and improve the user experience across diverse system configurations.

The phased approach ensures minimal risk while delivering meaningful improvements. The focus on independent installation methods aligns with plonk's goals of unified, efficient package management without unnecessary dependencies.

---

**Ready for Implementation**: This plan provides sufficient context for an agent to complete the audit, suggest specific improvements, and prepare implementation upon user consent.
