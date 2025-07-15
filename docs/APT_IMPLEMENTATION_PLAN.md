# APT Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding APT (Advanced Package Tool) support to plonk. APT is the primary package manager for Debian-based Linux distributions (Debian, Ubuntu, Linux Mint, etc.).

**Status**: Planning Phase

## Design Principles

1. **System-wide packages only** - Track packages installed system-wide (requires sudo)
2. **Privilege escalation** - Handle sudo requirements gracefully
3. **Distribution agnostic** - Work across different Debian-based distributions
4. **Package availability** - Check package availability before operations

## Key Challenges and Solutions

### 1. Privilege Requirements
**Challenge**: APT operations typically require sudo/root privileges

**Solution**:
- Detect if running with sufficient privileges
- For operations requiring sudo, provide clear error messages
- Consider read-only operations (list, search) that don't require sudo
- Document that plonk should be run with sudo for APT operations

### 2. Package Name Variations
**Challenge**: Package names can vary between distributions and versions

**Solution**:
- Use `apt-cache search` to verify package existence
- Handle virtual packages and meta-packages
- Store exact package names as returned by APT

### 3. System State Detection
**Challenge**: Distinguishing manually installed vs dependency packages

**Solution**:
- Use `apt-mark showmanual` to list manually installed packages
- Filter out packages installed as dependencies
- Track only explicitly installed packages

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/apt.go`
```go
type AptManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for apt binary and verify it's functional
- `ListInstalled()` - Use `apt-mark showmanual` to list user-installed packages
- `Install()` - Use `apt-get install -y` (will fail without sudo)
- `Uninstall()` - Use `apt-get remove -y` (will fail without sudo)
- `IsInstalled()` - Use `dpkg -l` to check installation status
- `Search()` - Use `apt-cache search` for package discovery
- `Info()` - Use `apt-cache show` for package details
- `GetInstalledVersion()` - Parse `dpkg -l` output for version

#### 1.2 Register in Manager Registry
- Add "apt" to `internal/managers/registry.go`
- Ensure Linux-only availability check

#### 1.3 Handle APT-specific edge cases
- Deal with held packages (`apt-mark showhold`)
- Handle packages in broken states
- Manage apt lock file conflicts
- Support both apt and apt-get commands

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/apt_test.go`)
- Mock command executor for all apt commands
- Test privilege detection
- Test package name parsing
- Test version extraction

#### 2.2 Integration Tests
- Test on Ubuntu/Debian containers if available
- Test with common packages (curl, git, vim)
- Test error conditions (no sudo, locked apt)

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add apt to PACKAGE_MANAGERS.md
- Update CLI.md with apt examples
- Document sudo requirements clearly

#### 3.2 Error Messages
- Add apt-specific error messages
- Provide helpful suggestions for common issues
- Include distribution-specific hints

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `apt --version` | Verify apt is functional |
| List installed | `apt-mark showmanual` | Lists manually installed packages |
| Install | `apt-get install -y <package>` | Requires sudo |
| Uninstall | `apt-get remove -y <package>` | Requires sudo |
| Check if installed | `dpkg -l <package>` | Check package state |
| Search | `apt-cache search <query>` | No sudo required |
| Get info | `apt-cache show <package>` | No sudo required |
| Get version | `dpkg -l <package>` | Parse version column |

### Package States

APT packages can have various states:
- `ii` - Installed
- `rc` - Removed but config files remain
- `un` - Not installed
- `hI` - Held installed

### Error Handling

Following plonk's error patterns:
```go
// No sudo access
return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "apt",
    "sudo access required for apt operations").
    WithSuggestionMessage("Run plonk with sudo or as root")

// Package not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "apt",
    fmt.Sprintf("package '%s' not found", name)).
    WithSuggestionMessage(fmt.Sprintf("Search available packages: apt-cache search %s", name))

// APT locked
return errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "apt",
    "apt database is locked").
    WithSuggestionMessage("Wait for other apt processes to complete or remove lock file")
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - apt found and functional
   - apt not found (non-Debian system)
   - apt found but not functional

2. **List Tests**
   - Parse apt-mark showmanual output
   - Handle empty package list
   - Filter out dependencies

3. **Install/Uninstall Tests**
   - Successful operations (mocked)
   - Permission denied errors
   - Package not found errors
   - APT lock conflicts

4. **Version Detection Tests**
   - Parse dpkg -l output format
   - Handle multi-arch packages
   - Handle epoch versions

### Mock Examples
```go
// Mock successful install
executor.EXPECT().CommandContext(ctx, "apt-get", "install", "-y", "curl").
    Return("", nil)

// Mock permission denied
executor.EXPECT().CommandContext(ctx, "apt-get", "install", "-y", "curl").
    Return("E: Could not open lock file", errors.New("exit status 100"))
```

## Platform Considerations

1. **Linux-only** - APT is only available on Debian-based Linux
2. **Distribution differences** - Ubuntu vs Debian package names
3. **Architecture support** - Handle i386, amd64, arm64
4. **Repository configuration** - Work with default repos only

## Security Considerations

1. **Privilege escalation** - Never automatically invoke sudo
2. **Package verification** - APT handles GPG verification
3. **Repository trust** - Use system-configured repositories only

## Future Enhancements

1. **PPA support** - Handle Ubuntu Personal Package Archives
2. **Snap integration** - Ubuntu's snap packages
3. **Automated updates** - Track security updates
4. **Dependency insights** - Show reverse dependencies

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Comprehensive test coverage (>80%)
3. ✅ Clear error messages for permission issues
4. ✅ Works on Ubuntu 20.04+ and Debian 10+
5. ✅ Documentation updated
6. ✅ Follows plonk's existing patterns

## Common APT Packages to Test With

- `curl` - HTTP client
- `git` - Version control
- `vim` - Text editor
- `build-essential` - Compiler tools
- `htop` - Process viewer

## Timeline Estimate

- Phase 1 (Core Implementation): 4-5 hours
- Phase 2 (Testing): 3-4 hours
- Phase 3 (Documentation): 1-2 hours

Total: ~8-11 hours of development time
