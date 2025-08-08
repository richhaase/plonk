# SelfInstall Implementation - Execution Guide

## Overview

This document provides step-by-step execution instructions for implementing the SelfInstall() method across all package managers. This work should be executed by a development agent following the detailed implementation plan in `selfinstall-implementation-plan.md`.

## Pre-Implementation Requirements

### Required Reading
1. `docs/selfinstall-implementation-plan.md` - Complete technical specification
2. `internal/resources/packages/interfaces.go` - Current PackageManager interface
3. `internal/clone/setup.go` - Existing installation logic to refactor
4. `CLAUDE.md` - Development rules (especially scope control and testing safety)

### Key Development Rules to Follow
- **NEVER** add unrequested features beyond the SelfInstall interface method
- **ALWAYS** use existing patterns from other package manager implementations
- **FORBIDDEN** to create new files unless absolutely necessary
- **REQUIRED** to maintain professional output without emojis
- **CRITICAL** unit tests must never modify the host system

## Implementation Tasks

### Phase 1: Interface Foundation ⏳
**Estimated Time**: 4-6 hours

#### Task 1.1: Extend PackageManager Interface
**File**: `internal/resources/packages/interfaces.go`

```go
// Add to existing PackageManager interface
type PackageManager interface {
    // ... existing methods
    SelfInstall(ctx context.Context) error
}

// Add new structs for installation metadata
type InstallationInfo struct {
    Method         InstallMethod
    IsIndependent  bool
    RequiresManager string
    SecurityLevel   SecurityLevel
    PlatformNotes   map[string]string
}

type InstallMethod int
const (
    ShellScript InstallMethod = iota
    BinaryDownload
    PackageManager
    RuntimeBundled
    NotSupported
)

type SecurityLevel int
const (
    HighSecurity SecurityLevel = iota
    MediumSecurity
    StandardSecurity
    ManualOnly
)

// Optional interface for installation metadata
type SelfInstaller interface {
    GetInstallationInfo() InstallationInfo
}
```

#### Task 1.2: Create Installation Helpers
**File**: `internal/resources/packages/install_helpers.go` (NEW FILE - justified for shared utilities)

```go
// Helper functions for SelfInstall implementations
func executeInstallCommand(ctx context.Context, cmd *exec.Cmd, managerName string) error
func checkPackageManagerAvailable(ctx context.Context, managerName string) (bool, error)
func requiresUserConfirmation(installInfo InstallationInfo) bool
func describeInstallMethod(method InstallMethod) string
func describeSecurityLevel(level SecurityLevel) string
```

**Implementation Priority**: Complete Task 1.1 and 1.2 together as they form the foundation.

### Phase 2: Tier 1 Independent Installers ⏳
**Estimated Time**: 8-10 hours

#### Task 2.1: Homebrew SelfInstall
**File**: `internal/resources/packages/homebrew.go`

```go
func (h *HomebrewManager) SelfInstall(ctx context.Context) error {
    // Check if already available (idempotent)
    if available, _ := h.IsAvailable(ctx); available {
        return nil
    }

    // Execute official installer script
    script := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
    cmd := exec.CommandContext(ctx, "bash", "-c", script)

    return executeInstallCommand(ctx, cmd, "Homebrew")
}

func (h *HomebrewManager) GetInstallationInfo() InstallationInfo {
    return InstallationInfo{
        Method:        ShellScript,
        IsIndependent: true,
        SecurityLevel: HighSecurity,
        PlatformNotes: map[string]string{
            "darwin": "Installs to /opt/homebrew (Apple Silicon) or /usr/local (Intel)",
            "linux":  "Installs to /home/linuxbrew/.linuxbrew",
        },
    }
}
```

#### Task 2.2: Cargo SelfInstall
**File**: `internal/resources/packages/cargo.go`

```go
func (c *CargoManager) SelfInstall(ctx context.Context) error {
    if available, _ := c.IsAvailable(ctx); available {
        return nil
    }

    script := `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y`
    cmd := exec.CommandContext(ctx, "bash", "-c", script)

    return executeInstallCommand(ctx, cmd, "Rust/Cargo")
}
```

#### Task 2.3: Go SelfInstall
**File**: `internal/resources/packages/goinstall.go`

```go
func (g *GoManager) SelfInstall(ctx context.Context) error {
    if available, _ := g.IsAvailable(ctx); available {
        return nil
    }

    switch runtime.GOOS {
    case "darwin":
        return g.installMacOS(ctx)
    case "linux":
        return g.installLinux(ctx)
    default:
        return fmt.Errorf("unsupported platform for Go auto-installation: %s", runtime.GOOS)
    }
}

// Platform-specific installation methods
func (g *GoManager) installMacOS(ctx context.Context) error
func (g *GoManager) installLinux(ctx context.Context) error
```

#### Task 2.4: UV SelfInstall
**File**: `internal/resources/packages/uv.go`

```go
func (u *UVManager) SelfInstall(ctx context.Context) error {
    if available, _ := u.IsAvailable(ctx); available {
        return nil
    }

    script := `curl -LsSf https://astral.sh/uv/install.sh | sh`
    cmd := exec.CommandContext(ctx, "bash", "-c", script)

    return executeInstallCommand(ctx, cmd, "UV")
}
```

#### Task 2.5: Pixi SelfInstall
**File**: `internal/resources/packages/pixi.go`

```go
func (p *PixiManager) SelfInstall(ctx context.Context) error {
    if available, _ := p.IsAvailable(ctx); available {
        return nil
    }

    script := `curl -fsSL https://pixi.sh/install.sh | sh`
    cmd := exec.CommandContext(ctx, "bash", "-c", script)

    return executeInstallCommand(ctx, cmd, "Pixi")
}
```

**Implementation Priority**: Complete all Tier 1 managers (Tasks 2.1-2.5) as a group since they follow the same pattern.

### Phase 3: Tier 2 Runtime-Dependent Installers ⏳
**Estimated Time**: 6-8 hours

#### Task 3.1: Composer SelfInstall
**File**: `internal/resources/packages/composer.go`

```go
func (c *ComposerManager) SelfInstall(ctx context.Context) error {
    if available, _ := c.IsAvailable(ctx); available {
        return nil
    }

    // Check PHP prerequisite
    if !CheckCommandAvailable("php") {
        return fmt.Errorf("composer requires PHP to be installed first - install PHP and retry")
    }

    return c.installWithHashVerification(ctx)
}

func (c *ComposerManager) installWithHashVerification(ctx context.Context) error {
    // Implement 4-step secure installation process from getcomposer.org
}
```

#### Task 3.2: Pip SelfInstall
**File**: `internal/resources/packages/pip.go`

```go
func (p *PipManager) SelfInstall(ctx context.Context) error {
    if available, _ := p.IsAvailable(ctx); available {
        return nil
    }

    // Try ensurepip first (Python 3.4+)
    if err := p.tryEnsurePip(ctx); err == nil {
        return nil
    }

    // Fallback to get-pip.py
    return p.installWithGetPip(ctx)
}

func (p *PipManager) tryEnsurePip(ctx context.Context) error
func (p *PipManager) installWithGetPip(ctx context.Context) error
```

### Phase 4: Tier 3 Package Manager Dependencies ⏳
**Estimated Time**: 6-8 hours

#### Task 4.1: NPM SelfInstall
**File**: `internal/resources/packages/npm.go`

```go
func (n *NPMManager) SelfInstall(ctx context.Context) error {
    if available, _ := n.IsAvailable(ctx); available {
        return nil
    }

    // Try to install via Homebrew if available
    if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
        return n.installViaHomebrew(ctx)
    }

    return fmt.Errorf("npm requires Node.js installation - install Node.js manually from https://nodejs.org or ensure Homebrew is available")
}

func (n *NPMManager) installViaHomebrew(ctx context.Context) error
```

#### Task 4.2: Gem SelfInstall
**File**: `internal/resources/packages/gem.go`

```go
func (g *GemManager) SelfInstall(ctx context.Context) error {
    if available, _ := g.IsAvailable(ctx); available {
        return nil
    }

    if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
        return g.installViaHomebrew(ctx)
    }

    return fmt.Errorf("gem requires Ruby installation - install Ruby manually or ensure Homebrew is available")
}
```

#### Task 4.3: .NET SelfInstall
**File**: `internal/resources/packages/dotnet.go`

```go
func (d *DotnetManager) SelfInstall(ctx context.Context) error {
    if available, _ := d.IsAvailable(ctx); available {
        return nil
    }

    if homebrewAvailable, _ := checkPackageManagerAvailable(ctx, "brew"); homebrewAvailable {
        return d.installViaHomebrew(ctx)
    }

    return fmt.Errorf(".NET tools require .NET SDK installation - install .NET SDK manually from https://dotnet.microsoft.com/download or ensure Homebrew is available")
}
```

### Phase 5: Integration with Clone Command ⏳
**Estimated Time**: 4-6 hours

#### Task 5.1: Update Clone Integration
**File**: `internal/clone/setup.go`

Replace existing `installDetectedManagers` function to use new SelfInstall interface:

```go
func installDetectedManagers(ctx context.Context, managers []string, cfg Config) error {
    registry := packages.NewManagerRegistry()

    for _, managerName := range managers {
        mgr, err := registry.GetManager(managerName)
        if err != nil {
            output.Printf("Warning: Unknown package manager '%s', skipping\n", managerName)
            continue
        }

        // Check if already available
        if available, _ := mgr.IsAvailable(ctx); available {
            output.Printf("%s is already installed\n", getManagerDescription(managerName))
            continue
        }

        // Get installation metadata if available
        var installInfo packages.InstallationInfo
        if installer, ok := mgr.(packages.SelfInstaller); ok {
            installInfo = installer.GetInstallationInfo()
        }

        // Interactive confirmation for security-sensitive operations
        if cfg.Interactive && packages.RequiresUserConfirmation(installInfo) {
            if !promptForInstallation(managerName, installInfo) {
                output.Printf("Skipping installation of %s\n", managerName)
                continue
            }
        }

        // Attempt auto-installation
        output.Printf("Installing %s...\n", getManagerDescription(managerName))
        if err := mgr.SelfInstall(ctx); err != nil {
            output.Printf("Failed to auto-install %s: %v\n", managerName, err)
            output.Printf("Manual installation: %s\n", getManualInstallInstructions(managerName))
            continue
        }

        // Verify installation
        if available, _ := mgr.IsAvailable(ctx); !available {
            output.Printf("Warning: %s installation completed but not detected\n", managerName)
        } else {
            output.Printf("Successfully installed %s\n", getManagerDescription(managerName))
        }
    }

    return nil
}

func promptForInstallation(managerName string, info packages.InstallationInfo) bool {
    // Implementation for user consent flow
}
```

#### Task 5.2: Remove Obsolete Installation Functions
**File**: `internal/clone/setup.go`

Remove/refactor these functions since they're replaced by SelfInstall interface:
- `installSingleManager`
- `installLanguagePackage`
- `installCargo`
- Any other hardcoded installation logic

**CRITICAL**: Ensure no functionality is lost - migrate any useful logic into the SelfInstall implementations.

## Testing Requirements

### Unit Tests - MUST BE SAFE
**Critical Rule**: Unit tests must NEVER modify the host system or call real installation commands.

#### Test Files to Create/Update:
- `internal/resources/packages/install_helpers_test.go` (NEW)
- Update existing `*_test.go` files for each package manager

#### Required Test Cases:
```go
// Test SelfInstall is idempotent
func TestSelfInstall_AlreadyAvailable(t *testing.T)

// Test error handling when prerequisites missing
func TestSelfInstall_MissingPrerequisites(t *testing.T)

// Test InstallationInfo metadata accuracy
func TestGetInstallationInfo_Metadata(t *testing.T)

// Test helper functions with mocked commands
func TestExecuteInstallCommand_MockedSuccess(t *testing.T)
func TestExecuteInstallCommand_MockedFailure(t *testing.T)
```

#### Testing Strategy:
- Use `MockCommandExecutor` interface for all external commands
- Test logic paths without executing real installation scripts
- Verify error messages and return values
- Test context cancellation handling

### Integration Testing
- Test via existing BATS framework in `tests/integration/`
- Focus on clone command workflow with SelfInstall
- Use safe, reversible operations only

## Validation Steps

### After Each Phase:
1. **Compile Check**: Ensure all code compiles without errors
2. **Unit Tests**: All tests pass with no system modifications
3. **Interface Compliance**: All managers implement required methods
4. **Error Handling**: Graceful failures with helpful messages

### Before Final Submission:
1. **Full Test Suite**: Run complete test suite - `go test ./...`
2. **Integration Tests**: Run relevant BATS tests
3. **Clone Workflow**: Test `plonk clone` with real repository
4. **Cross-Platform**: Verify behavior on both macOS and Linux if possible

## Common Pitfalls to Avoid

### Development Rule Violations:
- ❌ **Adding unrequested features** like progress bars or retry logic
- ❌ **Creating unnecessary files** when existing ones can be modified
- ❌ **Adding emojis** to output messages
- ❌ **Making tests that modify the system**

### Technical Issues:
- ❌ **Blocking operations** without context cancellation support
- ❌ **Hardcoded paths** that don't work cross-platform
- ❌ **Missing error handling** for network failures
- ❌ **Non-idempotent operations** that fail when called multiple times

### Security Concerns:
- ❌ **Executing unverified scripts** without user confirmation
- ❌ **Missing HTTPS verification** for downloads
- ❌ **Ignoring platform differences** in installation methods

## Success Criteria

The implementation is complete when:

1. ✅ **All 10 package managers** implement `SelfInstall()` method
2. ✅ **Tier classification respected**: Independent > Runtime-dependent > Package manager dependent
3. ✅ **Security maintained**: User confirmation for shell scripts, HTTPS sources
4. ✅ **Error handling robust**: Clear messages when installation fails
5. ✅ **Clone integration**: `plonk clone` uses SelfInstall instead of hardcoded logic
6. ✅ **Testing comprehensive**: Unit tests cover all paths safely
7. ✅ **Cross-platform support**: Works on macOS and Linux
8. ✅ **Idempotent behavior**: Safe to call multiple times

## Handoff Requirements

When implementation is complete, provide:

1. **Summary report** of what was implemented
2. **Test results** showing all tests pass
3. **Demo of clone workflow** with auto-installation
4. **Documentation updates** if interface changed significantly
5. **Known limitations** or platform-specific issues

This execution guide provides the implementing agent with everything needed to successfully add SelfInstall functionality while maintaining plonk's quality and security standards.
