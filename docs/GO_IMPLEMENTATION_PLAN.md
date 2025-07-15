# Go Package Manager Implementation Plan

## Overview

This document outlines the implementation plan for adding Go package manager support to plonk. The implementation will follow plonk's existing patterns while handling Go-specific behaviors around module-based installations and GOPATH/GOBIN considerations.

**Status**: Planning Phase

## Design Principles

1. **Focus on Global Tools Only** - Track only binaries installed with `go install`
2. **Module-Aware** - Work with Go modules (post-1.11) as the standard
3. **Environment Agnostic** - Use whatever `go` binary is in PATH
4. **Binary-Only Tracking** - Only track executable tools, not libraries

## Key Challenges and Solutions

### 1. Installation Location
**Challenge**: Go installs binaries to `$GOBIN` or `$GOPATH/bin` or `$HOME/go/bin`

**Solution**:
- Use `go env GOBIN` to detect installation directory
- Fall back to `go env GOPATH` + `/bin` if GOBIN not set
- Ensure the bin directory is in user's PATH for execution

### 2. Package Naming
**Challenge**: Go uses full module paths (e.g., `github.com/user/tool@version`)

**Solution**:
- Accept both short names and full paths in commands
- Store full module path in lock file for accurate reinstallation
- Extract binary name from module path for display
- Handle version specifications (@latest, @v1.2.3)

### 3. Version Detection
**Challenge**: Go doesn't have a simple "list installed" command for binaries

**Solution**:
- Use `go version -m <binary>` to get module information
- Parse module path and version from binary
- Scan GOBIN directory for Go binaries
- Filter out non-Go executables

## Implementation Steps

### Phase 1: Core Implementation

#### 1.1 Create `internal/managers/goinstall.go`
```go
type GoInstallManager struct{}
```

Implement all PackageManager interface methods:
- `IsAvailable()` - Check for go binary and verify version >= 1.16
- `ListInstalled()` - Scan GOBIN for Go binaries, use `go version -m`
- `Install()` - Use `go install <module>@<version>`
- `Uninstall()` - Remove binary from GOBIN (no built-in uninstall)
- `IsInstalled()` - Check if binary exists in GOBIN
- `Search()` - Use pkg.go.dev API or return helpful message
- `Info()` - Parse `go version -m` output for details
- `GetInstalledVersion()` - Extract version from binary with `go version -m`

#### 1.2 Register in Manager Registry
- Add "go" or "goinstall" to `internal/managers/registry.go`
- Ensure proper initialization in factory method

#### 1.3 Handle Go-specific edge cases
- Deal with GOBIN vs GOPATH/bin detection
- Handle missing binaries (user manually deleted)
- Parse module paths correctly (github.com/user/repo/cmd/tool)
- Support version specifications (@latest, @v1.2.3, @master)

### Phase 2: Testing

#### 2.1 Unit Tests (`internal/managers/goinstall_test.go`)
- Mock command executor for all go commands
- Test all interface methods
- Test error conditions (go not found, module not found, etc.)
- Test module path parsing and version extraction

#### 2.2 Integration Tests
- Test with real go installation if available
- Test installation of common tools (golangci-lint, gopls)
- Test version specifications
- Test GOBIN detection

### Phase 3: Documentation and Polish

#### 3.1 Update Documentation
- Add go to PACKAGE_MANAGERS.md
- Update CLI.md with go examples
- Document GOBIN/GOPATH behavior

#### 3.2 Error Messages
- Add go-specific error messages and suggestions
- Handle common Go issues (GOBIN not in PATH, module not found)

## Technical Specifications

### Command Mappings

| Operation | Command | Notes |
|-----------|---------|-------|
| Check availability | `go version` | Verify go is functional and >= 1.16 |
| List installed | Custom implementation | Scan GOBIN + `go version -m` |
| Install | `go install <module>@<version>` | Always specify version |
| Uninstall | `rm <GOBIN>/<binary>` | No built-in uninstall |
| Check if installed | Check file exists | Look in GOBIN directory |
| Search | pkg.go.dev API | Or provide helpful URL |
| Get info | `go version -m <binary>` | Parse module info |
| Get version | `go version -m <binary>` | Extract version string |

### Data Structures

```go
// Store full module information
type GoModuleInfo struct {
    ModulePath string  // e.g., "github.com/golangci/golangci-lint/cmd/golangci-lint"
    BinaryName string  // e.g., "golangci-lint"
    Version    string  // e.g., "v1.55.2"
}
```

### Error Handling

Following plonk's error patterns:
```go
// go not found
return false, nil  // Not an error, just unavailable

// Module not found
return errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "go",
    fmt.Sprintf("module '%s' not found", name)).
    WithSuggestionMessage("Search on pkg.go.dev or check module path")

// GOBIN not in PATH
return errors.NewError(errors.ErrConfiguration, errors.DomainPackages, "go",
    "GOBIN directory not in PATH").
    WithSuggestionMessage(fmt.Sprintf("Add %s to your PATH", gobin))
```

## Testing Strategy

### Unit Test Scenarios
1. **Availability Tests**
   - go found and version >= 1.16
   - go not found
   - go version too old

2. **List Tests**
   - Multiple Go binaries installed
   - Empty GOBIN directory
   - Mixed Go and non-Go executables

3. **Install/Uninstall Tests**
   - Successful installation with version
   - Module not found
   - Network errors
   - Binary already exists

4. **Version Detection Tests**
   - Parse version from go version -m output
   - Handle different version formats
   - Handle development versions

### Mock Examples
```go
// Mock successful install
executor.EXPECT().CommandContext(ctx, "go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2").
    Return("", nil)

// Mock version detection
executor.EXPECT().CommandContext(ctx, "go", "version", "-m", "/home/user/go/bin/golangci-lint").
    Return("golangci-lint: go1.21.5\n\tpath\tgithub.com/golangci/golangci-lint/cmd/golangci-lint\n\tmod\tgithub.com/golangci/golangci-lint\tv1.55.2\t", nil)
```

## Future Considerations

1. **Module Caching** - Leverage Go's module cache for faster reinstalls
2. **Workspace Support** - Handle go.work files for multi-module projects
3. **Binary Renaming** - Support custom binary names during installation
4. **Proxy Configuration** - Respect GOPROXY settings
5. **Private Modules** - Handle authentication for private repositories

## Success Criteria

1. ✅ All PackageManager interface methods implemented
2. ✅ Comprehensive test coverage (>80%)
3. ✅ Handles GOBIN/GOPATH configurations correctly
4. ✅ Clear error messages with actionable suggestions
5. ✅ Documentation updated
6. ✅ Works with standard Go installation
7. ✅ Follows plonk's existing patterns and conventions

## Key Differences from Other Managers

1. **No Built-in Uninstall** - Must manually remove binaries
2. **Module Paths** - Uses full paths not simple package names
3. **Binary Detection** - Must scan directory and verify with go version -m
4. **Version in Install** - Best practice to always specify version

## Common Go Tools to Test With

- `golang.org/x/tools/gopls@latest` - Go language server
- `github.com/golangci/golangci-lint/cmd/golangci-lint@latest` - Linter
- `mvdan.cc/gofumpt@latest` - Stricter gofmt
- `github.com/air-verse/air@latest` - Live reload
- `github.com/go-delve/delve/cmd/dlv@latest` - Debugger

## Timeline Estimate

- Phase 1 (Core Implementation): 3-4 hours
- Phase 2 (Testing): 2-3 hours
- Phase 3 (Documentation): 1 hour

Total: ~6-8 hours of development time
