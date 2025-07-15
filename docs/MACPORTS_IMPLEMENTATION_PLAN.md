# MacPorts Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding MacPorts support to plonk. MacPorts is an alternative package manager for macOS that provides a large collection of open-source software compiled specifically for Mac.

**Status**: Planning Phase

## Design Principles

1. **macOS only** - Exclusively for macOS systems
2. **Port variants aware** - Handle build variants and options
3. **Prefix isolation** - Respect MacPorts' /opt/local prefix
4. **Build from source** - May compile packages from source

## Key Challenges and Solutions

### 1. Installation Prefix
**Challenge**: MacPorts installs to /opt/local, separate from system

**Solution**:
- Respect MacPorts prefix (/opt/local)
- Don't conflict with Homebrew (/usr/local or /opt/homebrew)
- Ensure PATH includes /opt/local/bin
- Check for MacPorts installation properly

### 2. Port Variants
**Challenge**: MacPorts supports variants (+universal, +quartz, etc.)

**Solution**:
- Parse variants from port names (e.g., `vim +python310`)
- Store variants with package name
- Use exact variant specification for reinstalls
- List installed variants with `port installed`

### 3. Build Dependencies
**Challenge**: MacPorts often builds from source, requiring build deps

**Solution**:
- Let MacPorts handle build dependencies automatically
- Don't track build-only dependencies
- Focus on requested ports only
- Note that operations may take longer due to compilation

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/macports.go`
```go
type MacPortsManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for port command in /opt/local/bin
- `ListInstalled()` - Use `port -q installed` for active ports
- `Install()` - Use `port install`
- `Uninstall()` - Use `port uninstall`
- `IsInstalled()` - Use `port -q installed <port>`
- `Search()` - Use `port search`
- `Info()` - Use `port info`
- `GetInstalledVersion()` - Parse `port installed` output

#### 1.2 Register in Manager Registry
- Add "macports" to `internal/managers/registry.go`
- macOS only (Darwin platform check)

#### 1.3 Handle MacPorts-specific features
- Variant parsing and storage
- Active vs inactive ports
- Port dependencies
- Portfile locations
- Binary archives vs source builds

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/macports_test.go`)
- Mock port commands
- Test variant parsing
- Test version extraction
- Test active/inactive detection

#### 2.2 Integration Tests
- Test on macOS with MacPorts
- Test ports with variants
- Test source builds
- Test alongside Homebrew

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add macports to PACKAGE_MANAGERS.md
- Document variant syntax
- Note compilation times

#### 3.2 Error Messages
- MacPorts not installed
- Port build failures
- Privilege requirements

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `/opt/local/bin/port version` | Check specific path |
| List installed | `port -q installed` | Active ports only |
| Install | `sudo port install <port>` | May compile from source |
| Uninstall | `sudo port uninstall <port>` | Removes port |
| Check installed | `port -q installed <port>` | Check specific port |
| Search | `port search <query>` | Search available ports |
| Get info | `port info <port>` | Port details |
| Get version | `port -q installed <port>` | Parse version |

### Port Output Formats

Port list output:
```
The following ports are currently installed:
  curl @8.5.0_0+ssl (active)
  git @2.43.0_0+credential_osxkeychain+diff_highlight+doc+pcre2+perl5_34 (active)
  python310 @3.10.13_0+lto+optimizations (active)
  vim @9.0.2153_0+huge (active)
```

Port info output:
```
vim @9.0.2153_0 (editors)
Variants:             athena, big, cscope, gtk2, gtk3, huge, lua, motif, perl5_32,
                     perl5_34, python310, python311, python312, ruby30, ruby31,
                     ruby32, ruby33, small, tiny, universal, x11, xim

Description:          Vim is a greatly improved version of the vi editor.
Homepage:             https://www.vim.org/

Platforms:            darwin
License:              Vim
Maintainers:          Email: raimue@macports.org, GitHub: raimue
```

### Variant Handling

Parse and preserve variants:
```go
// Input: "vim +python310 +huge"
// Stored as: name="vim", variants="+python310 +huge"

// Installation command:
port install vim +python310 +huge
```

### Error Handling

```go
// MacPorts not installed
return errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "macports",
    "MacPorts is not installed").
    WithSuggestionMessage("Install MacPorts from https://www.macports.org/")

// Sudo required
return errors.NewError(errors.ErrFilePermission, errors.DomainPackages, "macports",
    "sudo access required for port operations").
    WithSuggestionMessage("MacPorts requires sudo for installation")

// Port not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "macports",
    fmt.Sprintf("port '%s' not found", name)).
    WithSuggestionMessage("Search available ports: port search " + name)

// Build failure
return errors.NewError(errors.ErrPackageInstall, errors.DomainPackages, "macports",
    fmt.Sprintf("failed to build port '%s'", name)).
    WithSuggestionMessage("Check build logs: port log " + name)
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - MacPorts installed at /opt/local
   - MacPorts not installed
   - Wrong platform (Linux)

2. **Variant Operations**
   - Parse variants from port names
   - Install with variants
   - List shows variants
   - Preserve variants on reinstall

3. **Port States**
   - Active ports
   - Inactive ports
   - Multiple versions

4. **Error Handling**
   - Build failures
   - Permission denied
   - Network issues (fetching)

### Mock Examples
```go
// Mock port installed
executor.EXPECT().CommandContext(ctx, "/opt/local/bin/port", "-q", "installed").
    Return("curl @8.5.0_0+ssl (active)\nvim @9.0.2153_0+huge+python310 (active)\n", nil)

// Mock install with variants
executor.EXPECT().CommandContext(ctx, "sudo", "/opt/local/bin/port", "install", "vim", "+huge", "+python310").
    Return("", nil)

// Mock port info
executor.EXPECT().CommandContext(ctx, "/opt/local/bin/port", "info", "vim").
    Return("vim @9.0.2153_0 (editors)\nVariants: athena, big, cscope...\n", nil)
```

## MacPorts-Specific Considerations

1. **Installation location**
   - Always in /opt/local
   - Separate from system
   - PATH configuration needed

2. **Source builds**
   - Often compiles from source
   - Can take significant time
   - Requires Xcode command line tools

3. **Variant system**
   - Build-time options
   - Can't change without rebuild
   - Important for functionality

4. **Coexistence**
   - Can run alongside Homebrew
   - Different installation prefixes
   - Potential PATH conflicts

## Security Considerations

1. **Sudo requirement** - Most operations need sudo
2. **Source verification** - MacPorts verifies checksums
3. **Compiler trust** - Requires Xcode/CLT
4. **Port maintainers** - Community maintained

## Platform Requirements

1. **macOS version** - Supports older macOS versions
2. **Xcode/CLT** - Required for building
3. **Disk space** - Builds can use significant space
4. **Architecture** - Universal binary support

## Future Enhancements

1. **Variant discovery** - List available variants
2. **Inactive port handling** - Manage inactive versions
3. **Port selection** - Handle `port select`
4. **Local portfiles** - Support custom ports

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Variant support working correctly
3. ✅ Proper /opt/local prefix handling
4. ✅ Clear sudo requirement messages
5. ✅ Works alongside Homebrew
6. ✅ Handles long build times gracefully

## Common Ports to Test

- `curl +ssl` - With SSL variant
- `git +credential_osxkeychain` - With macOS integration
- `python311` - Language interpreter
- `vim +huge` - Editor with features
- `ImageMagick +x11` - Complex dependencies

## Timeline Estimate

- Phase 1 (Core Implementation): 4-5 hours
- Phase 2 (Testing): 3-4 hours
- Phase 3 (Documentation): 1-2 hours

Total: ~8-11 hours of development time
