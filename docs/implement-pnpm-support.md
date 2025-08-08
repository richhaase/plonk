# Implementation Plan: pnpm Package Manager Support

**Status**: Planning
**Priority**: High
**Estimated Effort**: 3-4 days
**Target Release**: Next minor version

## Executive Summary

This document outlines the implementation plan for adding pnpm package manager support to plonk. pnpm is a fast, disk space efficient package manager that uses content-addressable storage and hard-linking. Adding pnpm support aligns perfectly with plonk's paradigm while providing significant performance benefits to users.

## Background & Justification

### Market Trends
- **Performance Focus**: pnpm is ~70% faster than npm for most operations
- **Disk Efficiency**: Uses ~50% less disk space via content-addressable storage
- **Growing Adoption**: Increasingly recommended for new JavaScript projects in 2024-2025
- **Developer Experience**: Better dependency isolation prevents phantom dependencies

### Strategic Alignment
- **Perfect Paradigm Fit**: pnpm maintains global package state compatible with plonk.lock
- **Unified Interface**: Commands map cleanly to plonk's PackageManager interface
- **Performance Benefits**: Enhances plonk's efficiency goals
- **Competitive Advantage**: Positions plonk ahead of alternatives

## Technical Analysis

### Command Mapping
```bash
# Core Operations
Install:    pnpm add -g <package>       → PackageManager.Install()
Uninstall:  pnpm remove -g <package>    → PackageManager.Uninstall()
List:       pnpm list -g --json         → PackageManager.ListInstalled()
Version:    pnpm list -g <pkg> --json   → PackageManager.InstalledVersion()
Info:       pnpm view <package> --json  → PackageManager.Info()
Check:      pnpm list -g <package>      → PackageManager.IsInstalled()
Upgrade:    pnpm update -g [packages]   → PackageManager.Upgrade()
Health:     pnpm --version              → PackageManager.CheckHealth()
Available:  which pnpm                  → PackageManager.IsAvailable()
```

### Interface Compatibility
```go
type PackageManager interface {
    // Core operations - all supported by pnpm
    IsAvailable(ctx context.Context) (bool, error)        ✅
    ListInstalled(ctx context.Context) ([]string, error)  ✅
    Install(ctx context.Context, name string) error       ✅
    Uninstall(ctx context.Context, name string) error     ✅
    IsInstalled(ctx context.Context, name string) (bool, error) ✅
    InstalledVersion(ctx context.Context, name string) (string, error) ✅
    Info(ctx context.Context, name string) (*PackageInfo, error) ✅

    // Required operations
    CheckHealth(ctx context.Context) (*HealthCheck, error)      ✅
    SelfInstall(ctx context.Context) error                      ✅
    Upgrade(ctx context.Context, packages []string) error       ✅

    // Optional operations
    Search(ctx context.Context, query string) ([]string, error) ❌ Not supported
}
```

### Single Limitation: Search Support
**Issue**: pnpm lacks a native `search` command
**Solution**: Simply return empty results from `Search()` method
**Impact**: Not a blocker - search returns empty results gracefully like other managers

### Full Interface Compatibility
pnpm provides **complete implementation** for all required methods:
- ✅ **All core operations** - standard package management
- ✅ **CheckHealth()** - via `pnpm --version` and `pnpm root -g`
- ✅ **SelfInstall()** - perfect single-method installation
- ✅ **Upgrade()** - via `pnpm update -g [packages]`
- ✅ **Search()** - returns empty results (graceful degradation)

### Perfect Self-Installation Support
pnpm has **perfect self-installation capabilities** using a single, fully independent method:

#### Single Method Approach
**Standalone script** - `curl -fsSL https://get.pnpm.io/install.sh | sh -` (FULLY INDEPENDENT)

#### Why Single Method Is Superior
- **Predictable Behavior**: Users always know exactly how pnpm will be installed
- **Maximum Independence**: Requires no other package managers or dependencies
- **First-Class Treatment**: pnpm is treated as an independent package manager, not subordinate to others
- **Platform Agnostic**: Works consistently across macOS, Linux, and Windows
- **No Unexpected Fallbacks**: Eliminates potential surprises or configuration drift

This makes pnpm an **exemplar** of proper self-installation design - simple, predictable, and independent.

#### Comparison with Existing Managers
Based on the self-installation audit (see `docs/self-install-fix.md`), pnpm represents best practices:

| Manager | Independence Level | Single Method | Fallbacks |
|---------|-------------------|---------------|-----------|
| **pnpm (proposed)** | ✅ **Perfect** | Standalone script | None |
| Homebrew | ✅ Perfect | Standalone script | None |
| Cargo | ✅ Perfect | rustup script | None |
| UV | ✅ Perfect | Standalone script | None |
| NPM | ❌ Dependent | brew install node | None (fails if no Homebrew) |
| Gem | ❌ Dependent | brew install ruby | None (fails if no Homebrew) |
| .NET | ❌ Dependent | brew install dotnet | None (fails if no Homebrew) |

**pnpm joins the elite group of truly independent package managers**, setting a positive example for the codebase.

## Implementation Plan

### Phase 1: Core Implementation (2 days)

#### File Creation
- `internal/resources/packages/pnpm.go` - Main implementation
- `internal/resources/packages/pnpm_test.go` - Unit tests

#### Core Structure
```go
// PnpmManager manages pnpm packages
type PnpmManager struct {
    binary string
}

// NewPnpmManager creates a new pnpm manager
func NewPnpmManager() *PnpmManager {
    return &PnpmManager{
        binary: "pnpm",
    }
}
```

#### Method Implementation Priority
1. **IsAvailable()** - Check if pnpm binary exists and is functional
2. **ListInstalled()** - Parse `pnpm list -g --json` output
3. **Install()/Uninstall()** - Core package management operations
4. **IsInstalled()** - Check specific package installation
5. **InstalledVersion()** - Extract version from list output
6. **Info()** - Use `pnpm view <package> --json`
7. **CheckHealth()** - Comprehensive health checking (REQUIRED)
8. **SelfInstall()** - Install pnpm via available package managers (REQUIRED)
9. **Upgrade()** - Use `pnpm update -g [packages]` (REQUIRED)
10. **Search()** - Return `ErrOperationNotSupported` (only optional method)

#### JSON Output Parsing
```go
// pnpm list -g --json structure (similar to npm)
type PnpmListOutput struct {
    Dependencies map[string]struct {
        Version string `json:"version"`
        Path    string `json:"path,omitempty"`
    } `json:"dependencies"`
}

// pnpm view output structure
type PnpmViewOutput struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Homepage    string            `json:"homepage"`
    Dependencies map[string]string `json:"dependencies"`
}
```

#### Error Handling Patterns
```go
// Handle pnpm-specific error patterns
func (p *PnpmManager) handleInstallError(err error, output []byte, packageName string) error {
    outputStr := string(output)

    if exitCode, ok := ExtractExitCode(err); ok {
        if strings.Contains(outputStr, "ERR_PNPM_PEER_DEP_ISSUES") {
            return fmt.Errorf("peer dependency issues installing %s", packageName)
        }
        if strings.Contains(outputStr, "ERR_PNPM_PACKAGE_NOT_FOUND") {
            return fmt.Errorf("package '%s' not found", packageName)
        }
        // Continue with standard error handling...
    }
    return err
}
```

#### Registration
```go
func init() {
    RegisterManager("pnpm", func() PackageManager {
        return NewPnpmManager()
    })
}
```

### Phase 2: Required Operations (1 day)

#### Health Checking
```go
func (p *PnpmManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
    check := &HealthCheck{
        Name:     "PNPM Manager",
        Category: "package-managers",
        Status:   "pass",
        Message:  "PNPM is available and properly configured",
    }

    // Check availability
    available, err := p.IsAvailable(ctx)
    if !available {
        check.Status = "warn"
        check.Message = "PNPM is not available"
        check.Suggestions = []string{
            "Install pnpm via standalone script: curl -fsSL https://get.pnpm.io/install.sh | sh -",
            "See https://pnpm.io/installation for manual installation options",
        }
    }

    // Discover global directory
    globalDir, err := p.getGlobalDirectory(ctx)
    if err == nil {
        check.Details = append(check.Details, fmt.Sprintf("PNPM global directory: %s", globalDir))
    }

    return check, nil
}
```

#### Self-Installation
```go
func (p *PnpmManager) SelfInstall(ctx context.Context) error {
    if available, _ := p.IsAvailable(ctx); available {
        return nil // Already available
    }

    // Use ONLY the most independent installation method - standalone script
    // This ensures predictable behavior and treats pnpm as a first-class citizen
    return p.installViaStandaloneScript(ctx)
}

// installViaStandaloneScript uses pnpm's official installation script
func (p *PnpmManager) installViaStandaloneScript(ctx context.Context) error {
    script := `curl -fsSL https://get.pnpm.io/install.sh | sh`
    return executeInstallScript(ctx, script, "pnpm")
}
```

#### Global Directory Discovery
```go
func (p *PnpmManager) getGlobalDirectory(ctx context.Context) (string, error) {
    // Get pnpm global directory
    output, err := ExecuteCommand(ctx, p.binary, "root", "-g")
    if err != nil {
        return "", fmt.Errorf("failed to get pnpm global directory: %w", err)
    }
    return strings.TrimSpace(string(output)), nil
}
```

### Phase 3: Testing & Integration (1 day)

#### Unit Test Coverage
```go
func TestPnpmManager_ListInstalled(t *testing.T) {
    tests := []struct {
        name           string
        mockOutput     []byte
        mockError      error
        expectedPkgs   []string
        expectedError  bool
    }{
        {
            name: "successful list with packages",
            mockOutput: []byte(`{
                "dependencies": {
                    "typescript": {"version": "5.3.3"},
                    "prettier": {"version": "3.1.0"}
                }
            }`),
            expectedPkgs: []string{"prettier", "typescript"},
        },
        {
            name: "empty list",
            mockOutput: []byte(`{"dependencies": {}}`),
            expectedPkgs: []string{},
        },
        // Additional test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := NewPnpmManager()
            // Use mock executor
            // Test implementation...
        })
    }
}
```

#### BATS Integration Tests
- Add BATS tests for pnpm package installation/uninstall scenarios
- Test pnpm integration with `plonk clone` workflows
- Verify error handling in real CLI usage scenarios
- Cross-platform compatibility testing via BATS

#### Test Helper Functions
```go
// Test helpers for pnpm operations
func setupMockPnpmManager(t *testing.T) *PnpmManager {
    return &PnpmManager{binary: "mock-pnpm"}
}

func createMockPnpmListOutput(packages map[string]string) []byte {
    deps := make(map[string]map[string]string)
    for name, version := range packages {
        deps[name] = map[string]string{"version": version}
    }

    output := map[string]interface{}{
        "dependencies": deps,
    }

    data, _ := json.Marshal(output)
    return data
}
```

### Phase 4: Documentation & Polish (0.5 days)

#### Documentation Updates

**Update `docs/CLI.md`**:
```markdown
- `pnpm:` - PNPM (fast, disk efficient JavaScript packages)

Examples:
```bash
plonk install pnpm:typescript
plonk install pnpm:prettier
```

**Update `docs/package-managers.md`**:
```markdown
### PNPM (pnpm) - JavaScript Package Manager
- **Status**: Fast and disk efficient package manager
- **Performance**: Up to 70% faster than npm
- **Commands**: `pnpm add -g`, `pnpm remove -g`, `pnpm list -g`
- **Storage**: Content-addressable store with hard-linking
- **Examples**: `typescript`, `prettier`, `eslint`
```

**Update `README.md`**:
```markdown
### Core Package Managers
- **Homebrew** (brew) - macOS/Linux packages and system tools
- **NPM** (npm) - Node.js packages (global)
- **PNPM** (pnpm) - Fast, efficient Node.js packages (global)
- **Cargo** (cargo) - Rust packages
```

#### CLI Examples
```bash
# Core managers
plonk install brew:wget npm:prettier pnpm:typescript cargo:ripgrep
```

## Implementation Details

### File Structure
```
internal/resources/packages/
├── pnpm.go           # Main implementation (~600 lines)
├── pnpm_test.go      # Comprehensive unit tests (~800 lines)
├── npm.go            # Reference implementation
└── ...
```

### Code Patterns to Follow
1. **Error Handling**: Follow npm.go patterns for robust error handling
2. **JSON Parsing**: Reuse parsing utilities where possible
3. **Command Execution**: Use existing ExecuteCommand/ExecuteCommandCombined
4. **Context Support**: Ensure all methods properly handle context cancellation
5. **Testing**: Comprehensive test coverage following existing patterns

### Performance Considerations
- Parse JSON output efficiently
- Handle large package lists gracefully
- Implement proper timeout handling
- Cache global directory discovery

## Risk Analysis

### Technical Risks: **LOW**
- **Mitigation**: Well-documented pnpm command interface
- **Mitigation**: Stable JSON output format similar to npm
- **Mitigation**: Can reuse 90% of npm.go implementation patterns

### Implementation Risks: **LOW**
- **Mitigation**: Comprehensive testing strategy
- **Mitigation**: Incremental implementation approach
- **Mitigation**: Search limitation is manageable (precedent exists)

### Compatibility Risks: **LOW**
- **Mitigation**: pnpm is compatible with npm registry
- **Mitigation**: Global packages work identically to npm
- **Mitigation**: No breaking changes to existing functionality

## Success Criteria

### Functional Requirements
- ✅ Install/uninstall global packages via pnpm
- ✅ List installed global packages with versions
- ✅ Check package installation status
- ✅ Retrieve package information from registry
- ✅ Upgrade packages to latest versions
- ✅ Health checking and diagnostics
- ✅ Self-installation when possible
- ❌ Package search (limitation documented)

### Non-Functional Requirements
- **Performance**: Commands execute within timeout limits
- **Reliability**: Proper error handling for all failure scenarios
- **Usability**: Clear error messages and status reporting
- **Maintainability**: Code follows existing patterns and conventions

### Integration Requirements
- ✅ Works with plonk's lock file system
- ✅ Integrates with plonk's orchestration layer
- ✅ Compatible with existing CLI commands
- ✅ Supports all output formats (table, JSON, YAML)

## Timeline

### Week 1
- **Days 1-2**: Core implementation (Phase 1)
- **Day 3**: Required operations - CheckHealth, SelfInstall, Upgrade (Phase 2)
- **Day 4**: Testing and integration (Phase 3)

### Week 2
- **Day 1**: Documentation and polish (Phase 4)
- **Day 2**: Code review and refinement
- **Day 3**: Final testing and release preparation

## Future Considerations

### Post-Implementation Enhancements
1. **Search Fallback**: Consider using npm search as fallback when available
2. **Performance Optimization**: Leverage pnpm's speed advantages
3. **Workspace Support**: Future consideration for pnpm workspace features
4. **Store Integration**: Potential integration with pnpm's content-addressable store

### Monitoring & Maintenance
1. **Usage Analytics**: Track pnpm adoption among plonk users
2. **Performance Metrics**: Compare install/upgrade times vs npm
3. **Error Monitoring**: Track common failure patterns
4. **Version Compatibility**: Monitor pnpm version compatibility

## Conclusion

Adding pnpm support to plonk represents a high-value, low-risk enhancement that aligns perfectly with plonk's goals of unified, efficient package management. The implementation follows established patterns and provides significant performance benefits to users while maintaining full compatibility with plonk's architecture and paradigm.

### Model Implementation
**pnpm serves as an exemplar of proper self-installation design** within the plonk ecosystem:
- ✅ **Single independent installation method** - no fallbacks or dependencies
- ✅ **Predictable behavior** - users always know exactly how pnpm will be installed
- ✅ **First-class treatment** - treated as independent package manager
- ✅ **No artificial dependencies** on other package managers

This implementation can serve as a **reference pattern** for improving other package managers' self-installation methods (see `docs/self-install-fix.md` for opportunities).

The straightforward implementation plan, comprehensive testing strategy, and clear success criteria make this an ideal candidate for the next minor release.
