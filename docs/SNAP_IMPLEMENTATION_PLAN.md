# Snap Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding Snap support to plonk. Snap is a universal package management system developed by Canonical, primarily used on Ubuntu but available across many Linux distributions.

**Status**: Planning Phase

## Design Principles

1. **Cross-distribution** - Work on any Linux distribution with snapd
2. **Channel aware** - Handle stable, candidate, beta, edge channels
3. **Classic vs Strict** - Distinguish confinement modes
4. **Service aware** - Some snaps run as services

## Key Challenges and Solutions

### 1. Snap Confinement
**Challenge**: Snaps can be strictly confined or classic (unconfined)

**Solution**:
- Detect confinement mode using `snap info`
- Track classic snaps that may need `--classic` flag
- Warn when installing classic snaps
- Store confinement mode in metadata

### 2. Channels and Tracks
**Challenge**: Snaps support multiple release channels and tracks

**Solution**:
- Default to stable channel
- Parse channel from install command (e.g., `--channel=edge`)
- Store channel information with package
- Use stored channel for updates

### 3. Service Management
**Challenge**: Some snaps provide services that need management

**Solution**:
- Detect service snaps using snap info
- Don't attempt to manage snap services
- Note service status in package info
- Let systemd/snapd handle services

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/snap.go`
```go
type SnapManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for snap command and snapd service
- `ListInstalled()` - Use `snap list` to get installed snaps
- `Install()` - Use `snap install` with appropriate flags
- `Uninstall()` - Use `snap remove`
- `IsInstalled()` - Parse `snap list` or check exit code
- `Search()` - Use `snap find` for discovery
- `Info()` - Use `snap info` for detailed information
- `GetInstalledVersion()` - Parse `snap list` output

#### 1.2 Register in Manager Registry
- Add "snap" to `internal/managers/registry.go`
- Linux-only, requires snapd

#### 1.3 Handle Snap-specific features
- Classic confinement detection and handling
- Channel/track management
- Revision tracking
- Automatic refresh awareness
- Interface connections (note but don't manage)

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/snap_test.go`)
- Mock snap commands
- Test channel parsing
- Test confinement detection
- Test version/revision parsing

#### 2.2 Integration Tests
- Test on Ubuntu with snapd
- Test classic and strict snaps
- Test different channels
- Test service snaps

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add snap to PACKAGE_MANAGERS.md
- Document channel behavior
- Explain classic confinement

#### 3.2 Error Messages
- Snapd not running errors
- Classic confinement warnings
- Connection errors to snap store

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `snap version` | Also check snapd service |
| List installed | `snap list` | Parse table output |
| Install | `snap install <snap>` | May need --classic |
| Install w/channel | `snap install <snap> --channel=<ch>` | Channel selection |
| Uninstall | `snap remove <snap>` | Removes snap |
| Check installed | `snap list <snap>` | Exit code check |
| Search | `snap find <query>` | Search store |
| Get info | `snap info <snap>` | Detailed info |
| Get version | `snap list <snap>` | Parse version/rev |

### Snap List Output Format
```
Name      Version    Rev    Tracking       Publisher   Notes
code      1.85.1     147    latest/stable  vscode✓     classic
firefox   120.0      3358   latest/stable  mozilla✓    -
```

### Snap Info Format
```
name:      code
summary:   Code editing. Redefined.
publisher: Microsoft Visual Studio Code (vscode✓)
store-url: https://snapcraft.io/code
contact:   https://github.com/Microsoft/vscode
license:   Proprietary
description: |
  Visual Studio Code is a lightweight but powerful source code editor...
channels:
  latest/stable:    1.85.1 2023-12-14 (147) 394MB classic
  latest/candidate: ↑
  latest/beta:      ↑
  latest/edge:      ↑
```

### Error Handling

```go
// Snapd not running
return errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "snap",
    "snapd service is not running").
    WithSuggestionMessage("Start snapd with: sudo systemctl start snapd")

// Classic confinement required
return errors.NewError(errors.ErrPackageInstall, errors.DomainPackages, "snap",
    fmt.Sprintf("snap '%s' requires classic confinement", name)).
    WithSuggestionMessage("Install with: snap install " + name + " --classic")

// Snap not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "snap",
    fmt.Sprintf("snap '%s' not found in store", name)).
    WithSuggestionMessage("Search snaps with: snap find " + name)
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - snapd running
   - snapd not installed
   - snapd installed but not running

2. **Snap Types**
   - Classic snaps
   - Strictly confined snaps
   - Service snaps
   - Desktop application snaps

3. **Channel Operations**
   - Default (stable) channel
   - Edge channel install
   - Channel tracking

4. **Error Conditions**
   - Network failures
   - Store authentication
   - Insufficient space

### Mock Examples
```go
// Mock snap list
executor.EXPECT().CommandContext(ctx, "snap", "list").
    Return("Name      Version    Rev    Tracking       Publisher   Notes\n" +
           "code      1.85.1     147    latest/stable  vscode✓     classic\n", nil)

// Mock classic snap install
executor.EXPECT().CommandContext(ctx, "snap", "install", "code", "--classic").
    Return("", nil)

// Mock snap info for confinement check
executor.EXPECT().CommandContext(ctx, "snap", "info", "code").
    Return("name:      code\n...\nchannels:\n  latest/stable: 1.85.1 2023-12-14 (147) 394MB classic\n", nil)
```

## Snap-Specific Considerations

1. **Automatic Updates**
   - Snaps auto-update by default
   - Can't pin versions easily
   - Refresh schedule managed by snapd

2. **Confinement Modes**
   - Strict: Default, sandboxed
   - Classic: Full system access
   - Devmode: Development mode

3. **Interfaces**
   - Snaps connect to interfaces
   - Some require manual connection
   - Don't manage connections

4. **Parallel Installation**
   - Snaps support parallel installs
   - Different feature set
   - Not covered in initial implementation

## Security Considerations

1. **Store authentication** - Handled by snapd
2. **Signature verification** - Automatic by snapd
3. **Confinement** - Respect security model
4. **Auto-connections** - Let snapd handle

## Platform Support

1. **Distribution support**
   - Ubuntu (primary)
   - Fedora, Debian, Arch (with snapd)
   - Not on some distributions (Mint)

2. **Architecture support**
   - amd64, arm64, armhf
   - Package availability varies

## Future Enhancements

1. **Refresh control** - Hold/unhold snaps
2. **Channel switching** - Change tracking channel
3. **Interface management** - Connect/disconnect
4. **Parallel installs** - Support instance names

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Handles classic and strict snaps correctly
3. ✅ Channel support implemented
4. ✅ Clear snapd availability checking
5. ✅ Works on Ubuntu and other snapd-enabled distros
6. ✅ Comprehensive test coverage

## Common Snaps to Test

- `hello` - Simple test snap
- `code` - VS Code (classic)
- `firefox` - Browser (strict)
- `lxd` - Container system (service)
- `spotify` - Desktop application

## Timeline Estimate

- Phase 1 (Core Implementation): 4-5 hours
- Phase 2 (Testing): 3-4 hours
- Phase 3 (Documentation): 1-2 hours

Total: ~8-11 hours of development time
