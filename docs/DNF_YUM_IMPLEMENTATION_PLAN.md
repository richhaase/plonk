# DNF/YUM Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding DNF/YUM support to plonk. DNF (Dandified YUM) is the next-generation package manager for RPM-based distributions (Fedora, RHEL, CentOS, Rocky Linux, AlmaLinux).

**Status**: Planning Phase

## Design Principles

1. **Unified interface** - Support both dnf and yum commands transparently
2. **System packages only** - Track system-wide installations (requires sudo)
3. **Enterprise ready** - Handle RHEL/CentOS specific behaviors
4. **Backwards compatible** - Work with older yum-based systems

## Key Challenges and Solutions

### 1. DNF vs YUM Detection
**Challenge**: Some systems have dnf, others have yum, some have both

**Solution**:
- Check for dnf first (preferred on modern systems)
- Fall back to yum if dnf not available
- Use consistent interface regardless of backend
- Store which tool was detected for consistent usage

### 2. Package Groups
**Challenge**: DNF/YUM support package groups (@development-tools)

**Solution**:
- Detect group syntax (@groupname)
- Use appropriate group commands (groupinstall/groupremove)
- List group members when querying
- Track groups separately in lock file

### 3. Repository Management
**Challenge**: Packages availability depends on enabled repositories

**Solution**:
- Work with currently enabled repositories only
- Don't modify repository configuration
- Report which repository provides a package
- Handle EPEL and other common third-party repos

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/dnf.go`
```go
type DnfManager struct{
    backend string // "dnf" or "yum"
}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for dnf, fall back to yum
- `ListInstalled()` - Use `dnf/yum list installed` with user-installed filter
- `Install()` - Use `dnf/yum install -y`
- `Uninstall()` - Use `dnf/yum remove -y`
- `IsInstalled()` - Use `rpm -q` for fast checking
- `Search()` - Use `dnf/yum search`
- `Info()` - Use `dnf/yum info`
- `GetInstalledVersion()` - Parse rpm query output

#### 1.2 Register in Manager Registry
- Add "dnf" to `internal/managers/registry.go`
- Linux-only, RPM-based distributions

#### 1.3 Handle DNF/YUM-specific features
- Package groups (@group syntax)
- Module streams (dnf module)
- Transaction history
- Weak dependencies
- Repository metadata caching

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/dnf_test.go`)
- Mock both dnf and yum commands
- Test fallback behavior
- Test group handling
- Test version parsing

#### 2.2 Integration Tests
- Test on Fedora/RHEL containers
- Test dnf-to-yum fallback
- Test with common packages
- Test with package groups

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add dnf to PACKAGE_MANAGERS.md
- Document dnf/yum compatibility
- Add enterprise Linux examples

#### 3.2 Error Messages
- Repository-specific errors
- Group installation errors
- Module stream conflicts

## Technical Specifications

### Command Mappings

| Operation | DNF Command | YUM Command | Notes |
|-----------|-------------|-------------|-------|
| Check availability | `dnf --version` | `yum --version` | Try dnf first |
| List installed | `dnf list --installed` | `yum list installed` | Filter user packages |
| Install | `dnf install -y <pkg>` | `yum install -y <pkg>` | Requires sudo |
| Install group | `dnf groupinstall -y <grp>` | `yum groupinstall -y <grp>` | For @groups |
| Uninstall | `dnf remove -y <pkg>` | `yum remove -y <pkg>` | Requires sudo |
| Check installed | `rpm -q <pkg>` | `rpm -q <pkg>` | Fast check |
| Search | `dnf search <query>` | `yum search <query>` | Search all fields |
| Get info | `dnf info <pkg>` | `yum info <pkg>` | Package details |
| Get version | `rpm -q --qf "%{VERSION}-%{RELEASE}" <pkg>` | Same | RPM query |

### Package Name Formats

DNF/YUM supports various package specifications:
- Simple name: `git`
- With version: `git-2.31.1`
- With arch: `git.x86_64`
- Groups: `@development-tools`
- Modules: `nodejs:14`

### Repository Information

Parse repository info from package listings:
```
git.x86_64    2.31.1-2.fc34    @updates
```

### Error Handling

```go
// No sudo access
return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "dnf",
    "sudo access required for dnf/yum operations").
    WithSuggestionMessage("Run plonk with sudo or as root")

// Package not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "dnf",
    fmt.Sprintf("package '%s' not found in enabled repositories", name)).
    WithSuggestionMessage("Check available packages with: dnf search " + name)

// Repository error
return errors.NewError(errors.ErrCommandExecution, errors.DomainPackages, "dnf",
    "failed to download repository metadata").
    WithSuggestionMessage("Check network connection and repository configuration")
```

## Testing Strategy

### Unit Test Scenarios
1. **Backend Detection**
   - System with only dnf
   - System with only yum
   - System with both (prefer dnf)
   - Neither available

2. **Package Operations**
   - Regular package install/remove
   - Group install/remove
   - Module operations
   - Version-specific installs

3. **Error Handling**
   - Permission denied
   - Package not found
   - Repository errors
   - Network failures

### Mock Examples
```go
// Mock DNF availability check
executor.EXPECT().LookPath("dnf").Return("/usr/bin/dnf", nil)
executor.EXPECT().CommandContext(ctx, "dnf", "--version").Return("4.10.0", nil)

// Mock package group install
executor.EXPECT().CommandContext(ctx, "dnf", "groupinstall", "-y", "@development-tools").
    Return("", nil)
```

## Platform Considerations

1. **Distribution differences**
   - Fedora (latest DNF features)
   - RHEL 8+ (dnf default)
   - RHEL 7/CentOS 7 (yum only)
   - Rocky/Alma Linux compatibility

2. **Architecture support**
   - Handle multiarch packages
   - x86_64, aarch64, ppc64le

3. **Repository ecosystems**
   - EPEL (Extra Packages)
   - RPM Fusion
   - Corporate repositories

## Security Considerations

1. **GPG verification** - Respect system GPG check settings
2. **Repository trust** - Only use enabled repositories
3. **Privilege escalation** - Never auto-invoke sudo
4. **Package signatures** - Let DNF/YUM handle verification

## Feature Comparison

| Feature | DNF | YUM |
|---------|-----|-----|
| Module streams | ✓ | ✗ |
| Parallel downloads | ✓ | ✗ |
| Weak dependencies | ✓ | Limited |
| Python 3 | ✓ | ✗ |
| Transaction rollback | ✓ | Limited |

## Future Enhancements

1. **Module stream management** - DNF module commands
2. **History integration** - Use transaction history
3. **Copr support** - Fedora personal repos
4. **Weak dependency handling** - Recommends/Suggests

## Success Criteria

1. ✅ Works with both DNF and YUM transparently
2. ✅ All PackageManager interface methods implemented
3. ✅ Handles package groups correctly
4. ✅ Clear sudo/permission error messages
5. ✅ Tested on Fedora, RHEL 8, and CentOS 7
6. ✅ Documentation covers both tools

## Common Packages to Test

- `git` - Version control
- `vim-enhanced` - Full vim editor
- `@development-tools` - Group install
- `httpd` - Apache web server
- `postgresql` - Database server

## Timeline Estimate

- Phase 1 (Core Implementation): 5-6 hours
- Phase 2 (Testing): 3-4 hours
- Phase 3 (Documentation): 1-2 hours

Total: ~9-12 hours of development time
