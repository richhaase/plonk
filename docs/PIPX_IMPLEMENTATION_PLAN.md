# pipx Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding pipx support to plonk. pipx is a tool for installing and running Python applications in isolated environments, making it ideal for Python CLI tools.

**Status**: Planning Phase

## Design Principles

1. **Isolation first** - Each tool in its own virtual environment
2. **CLI tools only** - Focus on applications with entry points
3. **Python version aware** - Handle multiple Python versions
4. **pipx managed only** - Don't mix with pip packages

## Key Challenges and Solutions

### 1. Virtual Environment Management
**Challenge**: pipx creates isolated venvs for each package

**Solution**:
- Let pipx handle all venv management
- Don't track venv locations
- Use pipx commands exclusively
- Trust pipx's environment handling

### 2. Python Version Selection
**Challenge**: pipx can use different Python versions per package

**Solution**:
- Default to pipx's default Python
- Support `--python` flag if specified
- Store Python version used (from pipx list)
- Note but don't enforce Python versions

### 3. Injected Packages
**Challenge**: pipx supports injecting packages into existing venvs

**Solution**:
- Initial implementation: ignore inject feature
- Track only main applications
- Future: consider inject support
- Document limitation clearly

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/pipx.go`
```go
type PipxManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for pipx command and verify it works
- `ListInstalled()` - Use `pipx list --json` for accurate parsing
- `Install()` - Use `pipx install`
- `Uninstall()` - Use `pipx uninstall`
- `IsInstalled()` - Parse pipx list output
- `Search()` - Delegate to pip search or return unsupported
- `Info()` - Parse `pipx list --json` for specific package
- `GetInstalledVersion()` - Extract from pipx list JSON

#### 1.2 Register in Manager Registry
- Add "pipx" to `internal/managers/registry.go`
- Cross-platform (wherever pipx runs)

#### 1.3 Handle pipx-specific features
- JSON output parsing for reliability
- Metadata about Python version
- venv location (informational only)
- Entry point information
- Upgrade vs reinstall semantics

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/pipx_test.go`)
- Mock pipx commands
- Test JSON parsing
- Test error conditions
- Test version extraction

#### 2.2 Integration Tests
- Test with real pipx if available
- Install common tools
- Test Python version handling
- Test alongside pip

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add pipx to PACKAGE_MANAGERS.md
- Explain isolation benefits
- Compare with pip approach

#### 3.2 Error Messages
- pipx not installed
- Package has no entry points
- Python version issues

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `pipx --version` | Verify pipx is functional |
| List installed | `pipx list --json` | JSON for reliable parsing |
| Install | `pipx install <package>` | Isolated install |
| Install with Python | `pipx install --python <python> <package>` | Specific Python |
| Uninstall | `pipx uninstall <package>` | Remove app and venv |
| Check installed | Parse `pipx list --json` | No direct command |
| Search | Not supported | pipx has no search |
| Get info | Parse `pipx list --json` | Extract package info |
| Get version | Parse `pipx list --json` | From package metadata |

### pipx JSON Output Format
```json
{
  "pipx_spec_version": "0.1",
  "venvs": {
    "black": {
      "metadata": {
        "injected_packages": {},
        "main_package": {
          "app_paths": [
            {
              "app": "black",
              "app_path": "/home/user/.local/bin/black"
            },
            {
              "app": "blackd",
              "app_path": "/home/user/.local/bin/blackd"
            }
          ],
          "package": "black",
          "package_version": "23.12.1",
          "python": "/usr/bin/python3.11"
        },
        "pipx_metadata_version": "0.3",
        "python_version": "3.11.7"
      }
    }
  }
}
```

### Error Handling

```go
// pipx not installed
return errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "pipx",
    "pipx is not installed").
    WithSuggestionMessage("Install pipx: python3 -m pip install --user pipx")

// No entry points
return errors.NewError(errors.ErrPackageInstall, errors.DomainPackages, "pipx",
    fmt.Sprintf("package '%s' has no console scripts", name)).
    WithSuggestionMessage("pipx can only install packages with CLI entry points")

// Already installed
return errors.NewError(errors.ErrPackageInstall, errors.DomainPackages, "pipx",
    fmt.Sprintf("'%s' already installed", name)).
    WithSuggestionMessage("Use 'pipx reinstall' or 'pipx upgrade' directly")
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - pipx installed and working
   - pipx not found
   - pipx broken installation

2. **JSON Parsing**
   - Single package
   - Multiple packages
   - Package with multiple entry points
   - Empty venv list

3. **Installation Tests**
   - Successful install
   - Package not found
   - No entry points
   - Already installed

4. **Python Version Handling**
   - Default Python
   - Specific Python version
   - Multiple Python versions

### Mock Examples
```go
// Mock pipx list JSON
executor.EXPECT().CommandContext(ctx, "pipx", "list", "--json").
    Return(`{"pipx_spec_version": "0.1", "venvs": {"black": {"metadata": {"main_package": {"package": "black", "package_version": "23.12.1"}}}}}`, nil)

// Mock successful install
executor.EXPECT().CommandContext(ctx, "pipx", "install", "black").
    Return("installed package black 23.12.1", nil)

// Mock package without entry points
executor.EXPECT().CommandContext(ctx, "pipx", "install", "requests").
    Return("", errors.New("No apps associated with package requests"))
```

## pipx-Specific Considerations

1. **Virtual Environment Isolation**
   - Each app in separate venv
   - No dependency conflicts
   - Clean uninstalls
   - Larger disk usage

2. **Entry Point Requirement**
   - Only installs packages with console scripts
   - Can't install libraries
   - Perfect for CLI tools
   - Clear error messages

3. **Python Version Flexibility**
   - Can use different Python per app
   - Survives Python upgrades better
   - More complex tracking
   - Version info in metadata

4. **Upgrade Semantics**
   - `pipx upgrade` vs `pipx reinstall`
   - Different from pip behavior
   - Preserve injected packages
   - Handle carefully

## Comparison with pip

| Feature | pip | pipx |
|---------|-----|------|
| Isolation | Shared environment | Separate venvs |
| Disk usage | Lower | Higher |
| Dependency conflicts | Possible | Impossible |
| Library support | Yes | No |
| CLI tools | Yes | Yes (only) |
| Python version | Single | Multiple |

## Security Considerations

1. **Isolation benefits** - No dependency conflicts
2. **Update safety** - Can update without breaking others
3. **Python trust** - Uses system/user Python
4. **Package verification** - Same as pip

## Future Enhancements

1. **Inject support** - Track injected packages
2. **Python version management** - Specify Python versions
3. **Upgrade handling** - Smart upgrade vs reinstall
4. **venv inspection** - Show venv details

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ JSON parsing working reliably
3. ✅ Clear entry point error messages
4. ✅ Proper isolation verification
5. ✅ Works alongside pip without conflicts
6. ✅ Comprehensive test coverage

## Common Packages to Test

- `black` - Code formatter
- `poetry` - Package manager
- `httpie` - HTTP client
- `flake8` - Linter
- `tox` - Test runner

## Timeline Estimate

- Phase 1 (Core Implementation): 3-4 hours
- Phase 2 (Testing): 2-3 hours
- Phase 3 (Documentation): 1 hour

Total: ~6-8 hours of development time
