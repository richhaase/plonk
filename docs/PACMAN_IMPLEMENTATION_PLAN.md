# Pacman Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding Pacman support to plonk. Pacman is the package manager for Arch Linux and Arch-based distributions (Manjaro, EndeavourOS, Artix).

**Status**: Planning Phase

## Design Principles

1. **Rolling release aware** - Handle frequent updates gracefully
2. **AUR conscious** - Distinguish official packages from AUR
3. **Explicit only** - Track explicitly installed packages, not dependencies
4. **Minimal approach** - Don't interfere with Arch philosophy

## Key Challenges and Solutions

### 1. Explicit vs Dependency Packages
**Challenge**: Pacman installs many dependencies automatically

**Solution**:
- Use `pacman -Qe` to list explicitly installed packages
- Ignore packages installed as dependencies (`pacman -Qd`)
- Track installation reason in package database

### 2. AUR Packages
**Challenge**: AUR packages are built differently, not in official repos

**Solution**:
- Detect AUR packages using `pacman -Qm` (foreign packages)
- Don't attempt to install AUR packages directly
- Provide clear messages about AUR helper requirements
- Track AUR packages separately if installed

### 3. Package Groups
**Challenge**: Pacman supports package groups (base-devel, gnome)

**Solution**:
- Detect group syntax in install commands
- Expand groups to individual packages for tracking
- Use `pacman -Sg` to list group contents
- Store individual packages, not group names

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/pacman.go`
```go
type PacmanManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for pacman binary and /etc/arch-release
- `ListInstalled()` - Use `pacman -Qe` for explicitly installed
- `Install()` - Use `pacman -S --noconfirm`
- `Uninstall()` - Use `pacman -R --noconfirm`
- `IsInstalled()` - Use `pacman -Q` for quick check
- `Search()` - Use `pacman -Ss` for repository search
- `Info()` - Use `pacman -Si` for repo info, `-Qi` for installed
- `GetInstalledVersion()` - Parse `pacman -Q` output

#### 1.2 Register in Manager Registry
- Add "pacman" to `internal/managers/registry.go`
- Arch Linux only (check /etc/arch-release)

#### 1.3 Handle Pacman-specific features
- Package groups expansion
- Foreign (AUR) package detection
- Optional dependency handling
- Provider/conflict resolution
- Partial upgrades warning

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/pacman_test.go`)
- Mock pacman commands
- Test group expansion
- Test version parsing
- Test AUR detection

#### 2.2 Integration Tests
- Test on Arch Linux container
- Test with core packages
- Test group installations
- Test foreign package detection

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add pacman to PACKAGE_MANAGERS.md
- Include Arch-specific notes
- Document AUR limitations

#### 3.2 Error Messages
- AUR package warnings
- Partial upgrade warnings
- Keyring issues

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `pacman --version` | Also check /etc/arch-release |
| List installed | `pacman -Qe` | Explicitly installed only |
| List foreign | `pacman -Qm` | AUR/manual packages |
| Install | `pacman -S --noconfirm <pkg>` | Requires sudo |
| Uninstall | `pacman -R --noconfirm <pkg>` | Basic remove |
| Check installed | `pacman -Q <pkg>` | Exit code check |
| Search | `pacman -Ss <query>` | Search repos |
| Get info (repo) | `pacman -Si <pkg>` | Repository info |
| Get info (local) | `pacman -Qi <pkg>` | Installed info |
| Get version | `pacman -Q <pkg>` | Parse output |
| List group | `pacman -Sg <group>` | Expand groups |

### Package Output Formats

Pacman query output examples:
```
# pacman -Q git
git 2.43.0-1

# pacman -Qe (explicitly installed)
base 3-2
git 2.43.0-1
linux 6.6.8.arch1-1

# pacman -Si git (repository info)
Repository      : extra
Name            : git
Version         : 2.43.0-1
Description     : Fast distributed version control system
```

### Error Handling

```go
// No sudo access
return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "pacman",
    "sudo access required for pacman operations").
    WithSuggestionMessage("Run plonk with sudo or as root")

// AUR package
return errors.NewError(errors.ErrUnsupported, errors.DomainPackages, "pacman",
    fmt.Sprintf("'%s' is an AUR package", name)).
    WithSuggestionMessage("Use an AUR helper like yay or paru to install AUR packages")

// Package not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "pacman",
    fmt.Sprintf("package '%s' not found in repositories", name)).
    WithSuggestionMessage("Update database with: sudo pacman -Sy")

// Database lock
return errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "pacman",
    "pacman database is locked").
    WithSuggestionMessage("Remove /var/lib/pacman/db.lck if no pacman process is running")
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - Arch Linux detection
   - Non-Arch system handling
   - Pacman binary check

2. **List Operations**
   - Explicitly installed filtering
   - Foreign package detection
   - Group member listing

3. **Install/Uninstall Tests**
   - Regular packages
   - Package groups
   - AUR package rejection
   - Already installed handling

4. **Search and Info**
   - Repository search
   - Local vs repo info
   - Version extraction

### Mock Examples
```go
// Mock successful install
executor.EXPECT().CommandContext(ctx, "pacman", "-S", "--noconfirm", "git").
    Return("", nil)

// Mock AUR package detection
executor.EXPECT().CommandContext(ctx, "pacman", "-Si", "yay").
    Return("", errors.New("error: package 'yay' was not found"))

// Mock group expansion
executor.EXPECT().CommandContext(ctx, "pacman", "-Sg", "base-devel").
    Return("base-devel autoconf\nbase-devel automake\nbase-devel binutils\n", nil)
```

## Arch-Specific Considerations

1. **Rolling release model**
   - Frequent updates
   - No version pinning
   - Partial upgrades discouraged

2. **Package signing**
   - GPG key management
   - Keyring updates
   - Trust database

3. **AUR ecosystem**
   - Not official packages
   - Require building
   - Different update mechanism

4. **Minimalism philosophy**
   - No automatic dependency installation
   - User responsibility
   - Explicit configuration

## Security Considerations

1. **Package signing** - Respect pacman's GPG verification
2. **AUR safety** - Don't enable AUR without user consent
3. **Partial upgrades** - Warn about system breakage risks
4. **Mirror trust** - Use configured mirrors only

## Future Enhancements

1. **AUR helper integration** - Optional yay/paru support
2. **Hook awareness** - Respect pacman hooks
3. **Package file tracking** - Use `pacman -Ql`
4. **Orphan detection** - Find unneeded packages

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Correctly identifies explicitly installed packages
3. ✅ Handles package groups appropriately
4. ✅ Clear messaging about AUR limitations
5. ✅ Works on Arch Linux and derivatives
6. ✅ Respects Arch philosophy

## Common Packages to Test

- `git` - Version control
- `base-devel` - Development group
- `firefox` - Web browser
- `neovim` - Text editor
- `htop` - System monitor

## Timeline Estimate

- Phase 1 (Core Implementation): 4-5 hours
- Phase 2 (Testing): 3-4 hours
- Phase 3 (Documentation): 1-2 hours

Total: ~8-11 hours of development time
