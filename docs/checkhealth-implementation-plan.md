# CheckHealth() Implementation Plan

This document provides a comprehensive plan for implementing and integrating the `CheckHealth()` method into plonk doctor, following the development rules in `~/.claude/CLAUDE.md`.

## Executive Summary

The `CheckHealth()` implementation will replace hardcoded PATH checking logic in `diagnostics/health.go` with a decentralized approach where each package manager is responsible for its own health validation. This follows the principle of separation of concerns and eliminates the need for hardcoded assumptions about package manager installation paths.

## Current State Analysis

### Existing Problems to Address

1. **Hardcoded PATH assumptions** in `health.go:470-478`:
   ```go
   importantPaths := map[string]string{
       "System":     "/usr/local/bin",
       "Homebrew":   getHomebrewPath(), // Still uses runtime detection
       "Cargo":      filepath.Join(homeDir, ".cargo/bin"),
       "Go":         goBinDir, // Good - uses dynamic detection
       "Python/pip": pythonUserBin, // Good - uses dynamic detection
       "Gem":        filepath.Join(homeDir, ".gem/ruby/bin"),
       "NPM":        filepath.Join(homeDir, ".npm-global/bin"),
   }
   ```

2. **Missing package managers**: Composer, .NET, UV, Pixi are not checked

3. **Inconsistent discovery methods**: Some use dynamic detection (Go, Python), others use hardcoded paths (Cargo, Gem, NPM)

4. **Duplicate logic**: PATH checking logic is separate from package manager availability logic

### Code to be Removed/Replaced

**Files requiring changes:**
- `/Users/rdh/src/plonk/internal/diagnostics/health.go`
  - `checkPackageManagerAvailability()` - Replace with individual health checks
  - `checkPathConfiguration()` - Replace with aggregated health check results
  - `getHomebrewPath()` - Move to homebrew package manager
  - `getPythonUserBinDir()` - Move to pip package manager
  - `getGoBinDir()` - Move to go package manager
  - Hardcoded path map (lines 470-478)

**Functions to preserve:**
- `detectShell()`, `generatePathExport()`, `generateShellCommands()` - Shell configuration helpers still needed

## Implementation Plan

### Phase 1: Interface Extension

#### 1.1 Add CheckHealth to PackageManager Interface

**File**: `/Users/rdh/src/plonk/internal/resources/packages/interfaces.go`

```go
// Add to PackageManager interface
type PackageManager interface {
    PackageManagerCapabilities

    // Existing methods...
    IsAvailable(ctx context.Context) (bool, error)
    ListInstalled(ctx context.Context) ([]string, error)
    // ... other existing methods

    // New method
    CheckHealth(ctx context.Context) (*HealthCheck, error)
}

// Add HealthCheck struct to packages package
type HealthCheck struct {
    Name        string   `json:"name" yaml:"name"`
    Category    string   `json:"category" yaml:"category"`
    Status      string   `json:"status" yaml:"status"`
    Message     string   `json:"message" yaml:"message"`
    Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
    Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
    Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}
```

#### 1.2 Add Default Implementation Helper

**File**: `/Users/rdh/src/plonk/internal/resources/packages/health_helpers.go` (NEW FILE)

```go
// DefaultCheckHealth provides basic health check implementation
func DefaultCheckHealth(ctx context.Context, manager PackageManager, managerName string) (*HealthCheck, error) {
    check := &HealthCheck{
        Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
        Category: "package-managers",
        Status:   "pass",
        Message:  fmt.Sprintf("%s is available and functional", managerName),
    }

    // Check basic availability
    available, err := manager.IsAvailable(ctx)
    if err != nil {
        if IsContextError(err) {
            return nil, err
        }
        check.Status = "fail"
        check.Message = fmt.Sprintf("%s availability check failed", managerName)
        check.Issues = []string{fmt.Sprintf("Error checking %s: %v", managerName, err)}
        return check, nil
    }

    if !available {
        check.Status = "warn"
        check.Message = fmt.Sprintf("%s is not available", managerName)
        check.Issues = []string{fmt.Sprintf("%s command not found or not functional", managerName)}
        check.Suggestions = []string{fmt.Sprintf("Install %s package manager", managerName)}
        return check, nil
    }

    check.Details = []string{fmt.Sprintf("%s binary found and functional", managerName)}
    return check, nil
}
```

### Phase 2: Package Manager Implementations

Each package manager will implement `CheckHealth()` with proper dynamic path discovery. The implementation pattern:

1. **Check availability** using existing `IsAvailable()` logic
2. **Discover bin directory** using package manager-specific commands
3. **Check PATH configuration** for discovered directories
4. **Provide actionable suggestions** for fixing issues

#### 2.1 Homebrew Implementation

**File**: `/Users/rdh/src/plonk/internal/resources/packages/homebrew.go`

```go
func (h *HomebrewManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
    check := &HealthCheck{
        Name:     "Homebrew Manager",
        Category: "package-managers",
        Status:   "pass",
        Message:  "Homebrew is available and properly configured",
    }

    // Check basic availability first
    available, err := h.IsAvailable(ctx)
    if err != nil {
        if IsContextError(err) {
            return nil, err
        }
        check.Status = "fail"
        check.Message = "Homebrew availability check failed"
        check.Issues = []string{fmt.Sprintf("Error checking homebrew: %v", err)}
        return check, nil
    }

    if !available {
        check.Status = "fail"
        check.Message = "Homebrew is required but not available"
        check.Issues = []string{"Homebrew is required for plonk to function properly"}
        check.Suggestions = []string{
            `Install Homebrew: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
            "After installation, ensure brew is in your PATH",
        }
        return check, nil
    }

    // Discover homebrew bin directory dynamically
    binDir, err := h.getBinDirectory(ctx)
    if err != nil {
        check.Status = "warn"
        check.Message = "Could not determine Homebrew bin directory"
        check.Issues = []string{fmt.Sprintf("Error discovering bin directory: %v", err)}
        return check, nil
    }

    // Check if bin directory is in PATH
    pathCheck := checkDirectoryInPath(binDir)
    check.Details = append(check.Details, fmt.Sprintf("Homebrew bin directory: %s", binDir))

    if !pathCheck.inPath && pathCheck.exists {
        check.Status = "warn"
        check.Message = "Homebrew bin directory exists but not in PATH"
        check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
        check.Suggestions = pathCheck.suggestions
    } else if !pathCheck.exists {
        check.Status = "warn"
        check.Message = "Homebrew bin directory does not exist"
        check.Issues = []string{fmt.Sprintf("Directory %s does not exist", binDir)}
    } else {
        check.Details = append(check.Details, "Homebrew bin directory is in PATH")
    }

    return check, nil
}

// getBinDirectory discovers the actual homebrew bin directory
func (h *HomebrewManager) getBinDirectory(ctx context.Context) (string, error) {
    output, err := ExecuteCommand(ctx, h.binary, "--prefix")
    if err != nil {
        return "", fmt.Errorf("failed to get homebrew prefix: %w", err)
    }

    prefix := strings.TrimSpace(string(output))
    return filepath.Join(prefix, "bin"), nil
}
```

#### 2.2 NPM Implementation

**File**: `/Users/rdh/src/plonk/internal/resources/packages/npm.go`

```go
func (n *NpmManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
    check := &HealthCheck{
        Name:     "NPM Manager",
        Category: "package-managers",
        Status:   "pass",
        Message:  "NPM is available and properly configured",
    }

    // Check basic availability
    available, err := n.IsAvailable(ctx)
    if err != nil {
        if IsContextError(err) {
            return nil, err
        }
        check.Status = "warn"
        check.Message = "NPM availability check failed"
        check.Issues = []string{fmt.Sprintf("Error checking npm: %v", err)}
        return check, nil
    }

    if !available {
        check.Status = "warn"
        check.Message = "NPM is not available"
        check.Issues = []string{"NPM command not found or not functional"}
        check.Suggestions = []string{
            "Install Node.js (includes NPM): brew install node",
            "Or install NPM separately if Node.js is already installed",
        }
        return check, nil
    }

    // Discover NPM global bin directory
    binDir, err := n.getGlobalBinDirectory(ctx)
    if err != nil {
        check.Status = "warn"
        check.Message = "Could not determine NPM global bin directory"
        check.Issues = []string{fmt.Sprintf("Error discovering bin directory: %v", err)}
        check.Details = []string{"Consider configuring NPM prefix: npm config set prefix ~/.npm-global"}
        return check, nil
    }

    check.Details = append(check.Details, fmt.Sprintf("NPM global bin directory: %s", binDir))

    // Check PATH configuration
    pathCheck := checkDirectoryInPath(binDir)
    if !pathCheck.inPath && pathCheck.exists {
        check.Status = "warn"
        check.Message = "NPM global bin directory exists but not in PATH"
        check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
        check.Suggestions = pathCheck.suggestions
    } else if !pathCheck.exists {
        check.Details = append(check.Details, "NPM global bin directory does not exist (no global packages installed)")
    } else {
        check.Details = append(check.Details, "NPM global bin directory is in PATH")
    }

    return check, nil
}

func (n *NpmManager) getGlobalBinDirectory(ctx context.Context) (string, error) {
    // Get npm global prefix
    output, err := ExecuteCommand(ctx, n.binary, "config", "get", "prefix")
    if err != nil {
        return "", fmt.Errorf("failed to get npm prefix: %w", err)
    }

    prefix := strings.TrimSpace(string(output))
    if prefix == "" {
        return "", fmt.Errorf("npm prefix is empty")
    }

    return filepath.Join(prefix, "bin"), nil
}
```

#### 2.3 Pip Implementation

**File**: `/Users/rdh/src/plonk/internal/resources/packages/pip.go`

```go
func (p *PipManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
    check := &HealthCheck{
        Name:     "Pip Manager",
        Category: "package-managers",
        Status:   "pass",
        Message:  "Pip is available and properly configured",
    }

    // Check availability
    available, err := p.IsAvailable(ctx)
    if err != nil {
        if IsContextError(err) {
            return nil, err
        }
        check.Status = "warn"
        check.Message = "Pip availability check failed"
        check.Issues = []string{fmt.Sprintf("Error checking pip: %v", err)}
        return check, nil
    }

    if !available {
        check.Status = "warn"
        check.Message = "Pip is not available"
        check.Issues = []string{"pip/pip3 command not found"}
        check.Suggestions = []string{
            "Install Python 3: brew install python3",
            "Ensure pip is installed: python3 -m ensurepip",
        }
        return check, nil
    }

    // Discover user bin directory
    binDir, err := p.getUserBinDirectory(ctx)
    if err != nil {
        check.Status = "warn"
        check.Message = "Could not determine Python user bin directory"
        check.Issues = []string{fmt.Sprintf("Error discovering bin directory: %v", err)}
        return check, nil
    }

    check.Details = append(check.Details, fmt.Sprintf("Python user bin directory: %s", binDir))

    // Check PATH
    pathCheck := checkDirectoryInPath(binDir)
    if !pathCheck.inPath && pathCheck.exists {
        check.Status = "warn"
        check.Message = "Python user bin directory exists but not in PATH"
        check.Issues = []string{fmt.Sprintf("Directory %s exists but not in PATH", binDir)}
        check.Suggestions = pathCheck.suggestions
    } else if !pathCheck.exists {
        check.Details = append(check.Details, "Python user bin directory does not exist (no user packages installed)")
    } else {
        check.Details = append(check.Details, "Python user bin directory is in PATH")
    }

    return check, nil
}

func (p *PipManager) getUserBinDirectory(ctx context.Context) (string, error) {
    // Use python3 to get user base directory
    output, err := ExecuteCommand(ctx, "python3", "-m", "site", "--user-base")
    if err != nil {
        return "", fmt.Errorf("failed to get python user base: %w", err)
    }

    userBase := strings.TrimSpace(string(output))
    if userBase == "" {
        return "", fmt.Errorf("python user base is empty")
    }

    return filepath.Join(userBase, "bin"), nil
}
```

#### 2.4 Additional Package Manager Implementations

Following the same pattern for:
- **Cargo**: Use `$CARGO_HOME/bin` or default `~/.cargo/bin`
- **Go**: Use `go env GOBIN` or `go env GOPATH`+"/bin" or `~/go/bin`
- **Gem**: Parse `gem environment` for executable directory
- **UV**: Use `uv tool dir` to discover tool directory
- **Pixi**: Use `pixi global bin` to discover bin directory
- **Composer**: Use `composer global config bin-dir --absolute`
- **.NET**: Use predictable `~/.dotnet/tools` path

### Phase 3: PATH Checking Helper

**File**: `/Users/rdh/src/plonk/internal/resources/packages/path_helpers.go` (NEW FILE)

```go
type PathCheckResult struct {
    inPath      bool
    exists      bool
    suggestions []string
}

// checkDirectoryInPath checks if a directory exists and is in PATH
func checkDirectoryInPath(directory string) PathCheckResult {
    result := PathCheckResult{}

    // Check if directory exists
    if _, err := os.Stat(directory); err == nil {
        result.exists = true
    }

    // Check if in PATH
    path := os.Getenv("PATH")
    pathDirs := strings.Split(path, string(os.PathListSeparator))
    for _, pathDir := range pathDirs {
        if pathDir == directory {
            result.inPath = true
            break
        }
    }

    // Generate suggestions if not in PATH but exists
    if result.exists && !result.inPath {
        shellPath := os.Getenv("SHELL")
        shell := detectShellFromDiagnostics(shellPath) // Move from diagnostics
        pathExport := generatePathExportFromDiagnostics([]string{directory}) // Move from diagnostics
        commands := generateShellCommandsFromDiagnostics(shell, pathExport) // Move from diagnostics

        result.suggestions = append(result.suggestions, fmt.Sprintf("Detected shell: %s", shell.name))
        result.suggestions = append(result.suggestions, "Add to PATH with:")
        for _, cmd := range commands {
            result.suggestions = append(result.suggestions, fmt.Sprintf("  %s", cmd))
        }
    }

    return result
}
```

### Phase 4: Integration with Diagnostics

#### 4.1 Update health.go

**File**: `/Users/rdh/src/plonk/internal/diagnostics/health.go`

Replace `checkPackageManagerAvailability()` and `checkPathConfiguration()`:

```go
// checkPackageManagerHealth runs health checks for all package managers
func checkPackageManagerHealth(ctx context.Context) []HealthCheck {
    var checks []HealthCheck

    registry := packages.NewManagerRegistry()
    homebrewAvailable := false

    for _, managerName := range registry.GetAllManagerNames() {
        mgr, err := registry.GetManager(managerName)
        if err != nil {
            // Create a basic failure check for registry errors
            check := HealthCheck{
                Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
                Category: "package-managers",
                Status:   "fail",
                Message:  "Package manager registration failed",
                Issues:   []string{fmt.Sprintf("Error getting %s manager: %v", managerName, err)},
            }
            checks = append(checks, check)
            continue
        }

        // Call the manager's CheckHealth method
        healthCheck, err := mgr.CheckHealth(ctx)
        if err != nil {
            if IsContextError(err) {
                // Context errors should bubble up
                return checks // Return what we have so far
            }
            // Convert to basic health check
            check := HealthCheck{
                Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
                Category: "package-managers",
                Status:   "fail",
                Message:  "Health check failed",
                Issues:   []string{fmt.Sprintf("Error checking %s health: %v", managerName, err)},
            }
            checks = append(checks, check)
            continue
        }

        // Convert packages.HealthCheck to diagnostics.HealthCheck
        diagnosticsCheck := HealthCheck{
            Name:        healthCheck.Name,
            Category:    healthCheck.Category,
            Status:      healthCheck.Status,
            Message:     healthCheck.Message,
            Details:     healthCheck.Details,
            Issues:      healthCheck.Issues,
            Suggestions: healthCheck.Suggestions,
        }
        checks = append(checks, diagnosticsCheck)

        // Track homebrew availability for overall health
        if managerName == "brew" && healthCheck.Status == "pass" {
            homebrewAvailable = true
        }
    }

    // Add overall package manager status check
    overallCheck := calculateOverallPackageManagerHealth(checks, homebrewAvailable)
    checks = append(checks, overallCheck)

    return checks
}

func calculateOverallPackageManagerHealth(checks []HealthCheck, homebrewAvailable bool) HealthCheck {
    check := HealthCheck{
        Name:     "Package Manager Ecosystem",
        Category: "package-managers",
        Status:   "pass",
        Message:  "Package management ecosystem is healthy",
    }

    if !homebrewAvailable {
        check.Status = "fail"
        check.Message = "Critical package manager missing"
        check.Issues = []string{"Homebrew is required but not available"}
        check.Suggestions = []string{
            "Install Homebrew first, then other package managers as needed",
            "Homebrew is the foundational package manager for plonk",
        }
        return check
    }

    availableCount := 0
    for _, mgr := range checks {
        if mgr.Category == "package-managers" && mgr.Status == "pass" {
            availableCount++
        }
    }

    if availableCount == 1 {
        check.Status = "warn"
        check.Message = "Limited package manager availability"
        check.Suggestions = []string{
            "Consider installing additional package managers as needed",
            "Available: npm (Node.js), pip (Python), cargo (Rust), gem (Ruby), go",
        }
    } else {
        check.Message = fmt.Sprintf("Package management ecosystem healthy (%d managers available)", availableCount)
    }

    return check
}
```

#### 4.2 Update RunHealthChecks

**File**: `/Users/rdh/src/plonk/internal/diagnostics/health.go`

```go
func RunHealthChecks() HealthReport {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    report := HealthReport{
        Overall: HealthStatus{
            Status:  "healthy",
            Message: "All systems operational",
        },
        Checks: []HealthCheck{},
    }

    // System checks
    report.Checks = append(report.Checks, checkSystemRequirements())
    report.Checks = append(report.Checks, checkEnvironmentVariables())
    report.Checks = append(report.Checks, checkPermissions())

    // Configuration checks
    report.Checks = append(report.Checks, checkConfigurationFile())
    report.Checks = append(report.Checks, checkConfigurationValidity())

    // Lock file checks
    report.Checks = append(report.Checks, checkLockFile())
    report.Checks = append(report.Checks, checkLockFileValidity())

    // Package manager health checks (UPDATED - replaces old logic)
    packageHealthChecks := checkPackageManagerHealth(ctx)
    report.Checks = append(report.Checks, packageHealthChecks...)

    // Executable path check
    report.Checks = append(report.Checks, checkExecutablePath())

    // Determine overall health
    report.Overall = calculateOverallHealth(report.Checks)

    return report
}
```

### Phase 5: Testing Strategy

#### 5.1 Unit Tests

Each package manager's `CheckHealth()` method needs comprehensive unit tests:

**File**: `/Users/rdh/src/plonk/internal/resources/packages/homebrew_test.go`

```go
func TestHomebrewManager_CheckHealth(t *testing.T) {
    tests := []struct {
        name           string
        isAvailable    bool
        availableError error
        binDirOutput   []byte
        binDirError    error
        pathEnv        string
        wantStatus     string
        wantContains   []string
    }{
        {
            name:         "healthy homebrew",
            isAvailable:  true,
            binDirOutput: []byte("/opt/homebrew"),
            pathEnv:      "/opt/homebrew/bin:/usr/bin",
            wantStatus:   "pass",
            wantContains: []string{"Homebrew bin directory is in PATH"},
        },
        {
            name:         "homebrew not available",
            isAvailable:  false,
            wantStatus:   "fail",
            wantContains: []string{"required but not available"},
        },
        {
            name:         "homebrew available but not in path",
            isAvailable:  true,
            binDirOutput: []byte("/opt/homebrew"),
            pathEnv:      "/usr/bin:/usr/local/bin",
            wantStatus:   "warn",
            wantContains: []string{"exists but not in PATH"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation using mock executor
        })
    }
}
```

#### 5.2 Integration Tests

**File**: `/Users/rdh/src/plonk/internal/diagnostics/health_test.go`

Test the integration between package managers and diagnostics:

```go
func TestPackageManagerHealthIntegration(t *testing.T) {
    // Test that health checks are properly aggregated
    // Test overall health calculation
    // Test context handling
}
```

#### 5.3 BATS Tests

Add to existing BATS tests to verify `plonk doctor` output includes package manager health.

### Phase 6: Migration and Cleanup

#### 6.1 Remove Obsolete Code

**Files to clean up:**
- Remove `getHomebrewPath()`, `getPythonUserBinDir()`, `getGoBinDir()` from health.go
- Remove hardcoded `importantPaths` map
- Remove `checkPackageManagerAvailability()` and `checkPathConfiguration()`
- Keep shell detection helpers (still needed for PATH suggestions)

#### 6.2 Backwards Compatibility

- Maintain the same HealthCheck struct format for command output
- Ensure doctor command output format remains consistent
- Keep the same overall health calculation logic

## Implementation Rules Compliance

### Following `~/.claude/CLAUDE.md` Rules:

1. **No Unrequested Features**: Implementing exactly what was requested - CheckHealth() method and doctor integration
2. **No New Files Unless Necessary**: Only creating helper files that are essential for avoiding code duplication
3. **No Emojis**: Using clean, professional status indicators
4. **Professional Output**: Following existing plonk output patterns
5. **Prefer Editing Existing Files**: Extending existing package managers rather than creating new ones

### Safety Rules:

1. **Unit Tests Never Modify Host System**: All tests use mock executors and temporary directories
2. **No Real System Commands in Tests**: Using MockCommandExecutor throughout
3. **Safe Test Packages**: Only using approved safe packages in integration tests

## Success Criteria

1. **Complete Coverage**: All 10 package managers implement CheckHealth()
2. **Dynamic Discovery**: No hardcoded path assumptions remain in diagnostics layer
3. **Clean Integration**: doctor command seamlessly uses new health checks
4. **Test Coverage**: 100% unit test coverage for new CheckHealth() implementations
5. **No Regressions**: Existing doctor functionality continues working
6. **Performance**: Health checks complete within existing 30-second timeout
7. **Code Quality**: All hardcoded diagnostic logic removed, replaced with interface calls

## Timeline Estimate

- **Phase 1** (Interface Extension): 1 day
- **Phase 2** (Package Manager Implementations): 3 days
- **Phase 3** (PATH Checking Helpers): 1 day
- **Phase 4** (Diagnostics Integration): 1 day
- **Phase 5** (Testing): 2 days
- **Phase 6** (Cleanup): 1 day

**Total Estimated Time**: 9 days

This plan ensures complete implementation while following all development rules and maintaining the existing code quality standards.
