# Flatpak Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding Flatpak support to plonk. Flatpak is a universal package management system for Linux that provides sandboxed desktop applications.

**Status**: Planning Phase

## Design Principles

1. **Desktop apps focus** - Primarily for GUI applications
2. **Sandboxed by default** - Respect Flatpak's security model
3. **Remote aware** - Handle multiple Flatpak remotes (Flathub, etc.)
4. **User vs System** - Support both installation scopes

## Key Challenges and Solutions

### 1. Application IDs
**Challenge**: Flatpak uses reverse DNS IDs (org.mozilla.firefox)

**Solution**:
- Accept both short names and full IDs
- Use `flatpak search` to resolve short names
- Store full application IDs
- Display friendly names where possible

### 2. Remotes and Repositories
**Challenge**: Apps come from different remotes (Flathub, GNOME, etc.)

**Solution**:
- Default to Flathub remote
- List remote with each app
- Don't manage remotes (system admin task)
- Work with configured remotes only

### 3. User vs System Installation
**Challenge**: Flatpak supports --user and --system scopes

**Solution**:
- Default to --user for better permissions
- Allow --system with sudo
- Track installation scope
- Query both scopes when listing

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/flatpak.go`
```go
type FlatpakManager struct{
    scope string // "user" or "system"
}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for flatpak command and runtime
- `ListInstalled()` - Use `flatpak list --app` for applications
- `Install()` - Use `flatpak install`
- `Uninstall()` - Use `flatpak uninstall`
- `IsInstalled()` - Check with `flatpak info`
- `Search()` - Use `flatpak search`
- `Info()` - Use `flatpak info` for details
- `GetInstalledVersion()` - Parse flatpak list output

#### 1.2 Register in Manager Registry
- Add "flatpak" to `internal/managers/registry.go`
- Linux-only (check for flatpak availability)

#### 1.3 Handle Flatpak-specific features
- Application ID resolution
- Remote detection
- Scope management (user/system)
- Runtime dependencies (note only)
- Permission overview

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/flatpak_test.go`)
- Mock flatpak commands
- Test ID resolution
- Test scope handling
- Test version parsing

#### 2.2 Integration Tests
- Test on Linux with Flatpak
- Test Flathub applications
- Test user and system scopes
- Test search functionality

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add flatpak to PACKAGE_MANAGERS.md
- Explain sandboxing benefits
- Note GUI app focus

#### 3.2 Error Messages
- No remotes configured
- Application ID not found
- Portal/runtime issues

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `flatpak --version` | Also check remotes |
| List installed | `flatpak list --app --columns=application,version,branch` | Apps only |
| Install (user) | `flatpak install --user -y <remote> <app>` | Non-interactive |
| Install (system) | `flatpak install --system -y <remote> <app>` | Needs sudo |
| Uninstall | `flatpak uninstall --user -y <app>` | Or --system |
| Check installed | `flatpak info <app>` | Exit code check |
| Search | `flatpak search <query>` | Search all remotes |
| Get info | `flatpak info --show-commit <app>` | Detailed info |
| Get version | Parse from list output | Branch/commit info |

### Flatpak Output Formats

List output:
```
Name                          Application ID                    Version   Branch
Firefox                       org.mozilla.firefox              120.0.1   stable
Visual Studio Code            com.visualstudio.code            1.85.1    stable
GIMP                         org.gimp.GIMP                    2.10.36   stable
```

Info output:
```
Name: Firefox
ID: org.mozilla.firefox
Ref: app/org.mozilla.firefox/x86_64/stable
Arch: x86_64
Branch: stable
Version: 120.0.1
License: MPL-2.0
Origin: flathub
Collection:
Installation: user
Installed: 742.4 MB
Runtime: org.freedesktop.Platform/x86_64/23.08
```

Search output:
```
Name          Description                               Application ID             Version   Branch   Remotes
Firefox       Fast, private web browser                org.mozilla.firefox        120.0.1   stable   flathub
```

### Error Handling

```go
// No remotes configured
return errors.NewError(errors.ErrConfiguration, errors.DomainPackages, "flatpak",
    "no flatpak remotes configured").
    WithSuggestionMessage("Add Flathub: flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo")

// Application not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "flatpak",
    fmt.Sprintf("application '%s' not found in remotes", name)).
    WithSuggestionMessage("Search applications: flatpak search " + name)

// Runtime missing
return errors.NewError(errors.ErrPackageInstall, errors.DomainPackages, "flatpak",
    "required runtime not available").
    WithSuggestionMessage("Flatpak will install required runtimes automatically")
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - Flatpak installed
   - Remotes configured
   - No Flatpak (non-Linux)

2. **ID Resolution**
   - Short name to full ID
   - Already full ID
   - Ambiguous names

3. **Scope Operations**
   - User installation
   - System installation
   - Query both scopes

4. **Search and Discovery**
   - Single result
   - Multiple results
   - No results

### Mock Examples
```go
// Mock flatpak list
executor.EXPECT().CommandContext(ctx, "flatpak", "list", "--app", "--columns=application,version,branch").
    Return("org.mozilla.firefox\t120.0.1\tstable\n", nil)

// Mock user install
executor.EXPECT().CommandContext(ctx, "flatpak", "install", "--user", "-y", "flathub", "org.mozilla.firefox").
    Return("", nil)

// Mock search
executor.EXPECT().CommandContext(ctx, "flatpak", "search", "firefox").
    Return("Firefox\tFast, private web browser\torg.mozilla.firefox\t120.0.1\tstable\tflathub\n", nil)
```

## Flatpak-Specific Considerations

1. **Sandboxing Model**
   - Apps run in isolation
   - Limited filesystem access
   - Portal system for permissions
   - Different from traditional packages

2. **Application Focus**
   - Primarily GUI applications
   - Not for CLI tools
   - Different use case from pip/npm
   - Complements system packages

3. **Remote System**
   - Multiple sources possible
   - Flathub is primary
   - Remote management out of scope
   - Trust system configuration

4. **Runtimes and SDKs**
   - Apps depend on runtimes
   - Shared between apps
   - Automatic installation
   - Don't track explicitly

## Security Considerations

1. **Sandboxing** - Apps are isolated by default
2. **Permissions** - Flatpak manages app permissions
3. **Signatures** - Remotes sign packages
4. **User installation** - No system-wide changes

## Platform Support

1. **Linux only** - Flatpak is Linux-specific
2. **Distribution agnostic** - Works on most distros
3. **Desktop environments** - Better integration with some DEs
4. **Architecture** - Multi-arch support

## Future Enhancements

1. **Permission management** - Show/modify permissions
2. **Remote management** - Add/remove remotes
3. **Override support** - Environment overrides
4. **Update policies** - Control app updates

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Application ID resolution working
3. ✅ User/system scope support
4. ✅ Search functionality implemented
5. ✅ Clear remote configuration messages
6. ✅ Works on major Linux distributions

## Common Applications to Test

- `org.mozilla.firefox` - Web browser
- `com.visualstudio.code` - Code editor
- `org.gimp.GIMP` - Image editor
- `com.spotify.Client` - Music streaming
- `org.libreoffice.LibreOffice` - Office suite

## Timeline Estimate

- Phase 1 (Core Implementation): 4-5 hours
- Phase 2 (Testing): 3-4 hours
- Phase 3 (Documentation): 1-2 hours

Total: ~8-11 hours of development time
