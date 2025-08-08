# Implementation Plan: Conda Package Manager Support

**Status**: Planning
**Priority**: Medium
**Estimated Effort**: 4-5 days
**Target Release**: Future minor version (after pnpm)

## Executive Summary

This document outlines the implementation plan for adding conda package manager support to plonk. Conda is the dominant package manager for data science, machine learning, and scientific computing workflows, managing both Python packages and system-level dependencies. Adding conda support addresses a critical gap in plonk's coverage of the data science ecosystem.

## Background & Justification

### Market Trends
- **Data Science Dominance**: Conda remains the primary package manager for data science and ML workflows
- **Scientific Computing**: Essential for bioinformatics, computational science, and research environments
- **Enterprise Adoption**: Many organizations standardize on Anaconda/Miniconda distributions
- **Unique Capabilities**: Only Python package manager that handles non-Python system dependencies (C libraries, R packages, system tools)

### Strategic Alignment
- **Market Coverage**: Addresses the massive data science and research community not served by pipx/uv
- **Unique Value**: Conda provides capabilities no other Python package manager offers
- **Enterprise Reach**: Extends plonk into scientific computing and enterprise data science environments
- **Complementary Coverage**: Works alongside existing Python managers (pipx, uv, pixi) without overlap

## Technical Analysis

### Command Mapping
```bash
# Core Operations (focused on base environment global packages)
Install:    conda install -n base <package>     → PackageManager.Install()
Uninstall:  conda remove -n base <package>      → PackageManager.Uninstall()
List:       conda list -n base --json           → PackageManager.ListInstalled()
Version:    conda list -n base <pkg> --json     → PackageManager.InstalledVersion()
Info:       conda search <package> --info       → PackageManager.Info()
Check:      conda list -n base <package>        → PackageManager.IsInstalled()
Upgrade:    conda update -n base [packages]     → PackageManager.Upgrade()
Search:     conda search <query>                → PackageManager.Search()
Health:     conda info                          → PackageManager.CheckHealth()
Available:  which conda                         → PackageManager.IsAvailable()
```

### Interface Compatibility
```go
type PackageManager interface {
    // Core operations - all supported by conda
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
    Search(ctx context.Context, query string) ([]string, error) ✅ Fully supported
}
```

### Full Interface Compatibility
Conda provides **complete implementation** for all required methods:
- ✅ **All core operations** - full package management capabilities
- ✅ **CheckHealth()** - via `conda info` and environment analysis
- ✅ **SelfInstall()** - multiple installation methods available
- ✅ **Upgrade()** - via `conda update -n base [packages]`
- ✅ **Search()** - via `conda search <query>` (full support unlike many managers)

### Base Environment Strategy
**Design Decision**: Focus on base environment global package management
- **Rationale**: Aligns with plonk's global package paradigm
- **Implementation**: Use `-n base` flag for all operations
- **Benefits**: Consistent with other package managers' global approach
- **Documentation**: Clear guidance on conda vs other Python managers

### Multi-Variant Detection Strategy
**IMPORTANT**: This introduces a new architectural pattern to plonk - intelligent binary detection within a single package manager.

#### Conda Ecosystem Variants
The conda ecosystem has three main variants that serve different use cases:

| Variant | Size | Speed | Use Case | API Compatibility |
|---------|------|--------|----------|-------------------|
| **conda** | ~100 MiB | Standard | Traditional workflows | 100% (original) |
| **mamba** | ~100 MiB | 10-100x faster | Performance-focused | 100% (drop-in replacement) |
| **micromamba** | ~13 MiB | Fast | CI/CD, containers | ~90% (different CLI for environments) |

#### Unified Implementation Decision
**Approach**: Single `CondaManager` with intelligent binary detection
**Rationale**: All variants use identical commands for package management operations

```bash
# Core operations are IDENTICAL across all variants
{conda|mamba|micromamba} install -n base <package>
{conda|mamba|micromamba} remove -n base <package>
{conda|mamba|micromamba} list -n base --json
{conda|mamba|micromamba} search <package> --json
```

#### Detection Priority Strategy
```bash
# Proposed priority order (fastest to slowest)
1. mamba (fastest dependency resolution)
2. micromamba (fast + minimal)
3. conda (standard fallback)
```

## Implementation Plan

### Phase 1: Core Implementation (2-3 days)

#### File Creation
- `internal/resources/packages/conda.go` - Main implementation
- `internal/resources/packages/conda_test.go` - Unit tests

#### Core Structure with Intelligent Detection
```go
// CondaManager manages conda packages using the best available conda variant
type CondaManager struct {
    binary       string           // Detected binary: "conda", "mamba", or "micromamba"
    variant      CondaVariant     // Enum for variant-specific behavior
    capabilities VariantCapabilities // What this variant supports
}

// CondaVariant represents the detected conda ecosystem variant
type CondaVariant int

const (
    VariantConda CondaVariant = iota
    VariantMamba
    VariantMicromamba
)

// VariantCapabilities tracks variant-specific features
type VariantCapabilities struct {
    FastDependencyResolution bool
    MinimalFootprint        bool
    FullEnvironmentSupport  bool
    PerformanceOptimized    bool
}

// NewCondaManager creates a new conda manager with intelligent binary detection
func NewCondaManager() *CondaManager {
    binary, variant, capabilities := detectCondaVariant()
    return &CondaManager{
        binary:       binary,
        variant:      variant,
        capabilities: capabilities,
    }
}
```

#### Intelligent Binary Detection Implementation
**NEW ARCHITECTURAL FEATURE**: This is the first plonk package manager with intelligent binary detection.

```go
// detectCondaVariant performs comprehensive conda variant detection
func detectCondaVariant() (string, CondaVariant, VariantCapabilities) {
    // Detection strategy with priority order
    candidates := []struct {
        binary       string
        variant      CondaVariant
        capabilities VariantCapabilities
    }{
        {
            binary:  "mamba",
            variant: VariantMamba,
            capabilities: VariantCapabilities{
                FastDependencyResolution: true,
                PerformanceOptimized:    true,
                FullEnvironmentSupport:  true,
            },
        },
        {
            binary:  "micromamba",
            variant: VariantMicromamba,
            capabilities: VariantCapabilities{
                FastDependencyResolution: true,
                MinimalFootprint:        true,
                PerformanceOptimized:    true,
                FullEnvironmentSupport:  false, // Different CLI for environments
            },
        },
        {
            binary:  "conda",
            variant: VariantConda,
            capabilities: VariantCapabilities{
                FullEnvironmentSupport: true,
            },
        },
    }

    for _, candidate := range candidates {
        if CheckCommandAvailable(candidate.binary) {
            // Additional validation that the binary is functional
            if isCondaVariantFunctional(candidate.binary) {
                return candidate.binary, candidate.variant, candidate.capabilities
            }
        }
    }

    // Fallback to conda (will fail later in IsAvailable if not found)
    return "conda", VariantConda, VariantCapabilities{
        FullEnvironmentSupport: true,
    }
}

// isCondaVariantFunctional validates that the detected binary actually works
func isCondaVariantFunctional(binary string) bool {
    // Test with a quick --version check or similar
    // This prevents false positives from broken installations
    // Implementation details TBD - see Open Questions
}
```

#### Method Implementation Priority
1. **detectCondaVariant()** - **NEW**: Intelligent binary detection system
2. **IsAvailable()** - Check if detected binary exists and is functional
3. **ListInstalled()** - Parse `{binary} list -n base --json` output
4. **Install()/Uninstall()** - Base environment package management operations
5. **IsInstalled()** - Check specific package installation in base environment
6. **InstalledVersion()** - Extract version from list output
7. **Info()** - Use `{binary} search <package> --info`
8. **Search()** - Parse `{binary} search <query>` output
9. **CheckHealth()** - Variant-aware health checking with performance insights
10. **SelfInstall()** - Install best available conda variant with fallbacks
11. **Upgrade()** - Use `{binary} update -n base [packages]`

#### JSON Output Parsing
```go
// conda list -n base --json structure
type CondaListItem struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Build       string `json:"build"`
    Channel     string `json:"channel"`
    Platform    string `json:"platform"`
    BuildString string `json:"build_string"`
}

// conda search --info output structure
type CondaSearchInfo struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Build        string            `json:"build"`
    Channel      string            `json:"channel"`
    Dependencies []string          `json:"depends"`
    License      string            `json:"license"`
    Summary      string            `json:"summary"`
    Description  string            `json:"description"`
    Homepage     string            `json:"home"`
    Size         int64             `json:"size"`
}

// conda info output structure (for health checking)
type CondaInfo struct {
    Platform           string   `json:"platform"`
    CondaVersion       string   `json:"conda_version"`
    PythonVersion      string   `json:"python_version"`
    BaseEnvironment    string   `json:"base_environment"`
    CondaPrefix        string   `json:"conda_prefix"`
    Channels           []string `json:"channels"`
    PackageCacheDir    []string `json:"pkgs_dirs"`
    EnvironmentDirs    []string `json:"envs_dirs"`
    VirtualPackages    []string `json:"virtual_packages"`
}
```

#### Error Handling Patterns
```go
// Handle conda-specific error patterns
func (c *CondaManager) handleInstallError(err error, output []byte, packageName string) error {
    outputStr := strings.ToLower(string(output))

    if exitCode, ok := ExtractExitCode(err); ok {
        // Conda-specific error patterns
        if strings.Contains(outputStr, "packagenotfounderror") ||
           strings.Contains(outputStr, "no packages found matching") {
            return fmt.Errorf("package '%s' not found", packageName)
        }
        if strings.Contains(outputStr, "unsatisfiableerror") ||
           strings.Contains(outputStr, "conflicting dependencies") {
            return fmt.Errorf("dependency conflicts installing '%s'", packageName)
        }
        if strings.Contains(outputStr, "channelnotavailableerror") {
            return fmt.Errorf("conda channels unavailable for package '%s'", packageName)
        }
        if strings.Contains(outputStr, "environmentlockederror") {
            return fmt.Errorf("conda environment is locked")
        }
        if strings.Contains(outputStr, "condahttperror") ||
           strings.Contains(outputStr, "connection failed") {
            return fmt.Errorf("network error during conda operation")
        }

        // Standard exit code handling
        if exitCode != 0 {
            if len(output) > 0 {
                errorOutput := strings.TrimSpace(string(output))
                if len(errorOutput) > 500 {
                    errorOutput = errorOutput[:500] + "..."
                }
                return fmt.Errorf("conda installation failed: %s", errorOutput)
            }
            return fmt.Errorf("conda installation failed (exit code %d): %w", exitCode, err)
        }
        return nil
    }

    return err
}
```

#### Base Environment Focus
```go
// All operations target the base environment for global package management
const baseEnvironmentFlag = "-n base"

func (c *CondaManager) Install(ctx context.Context, name string) error {
    // Install to base environment only
    output, err := ExecuteCommandCombined(ctx, c.binary, "install", baseEnvironmentFlag, "-y", name)
    if err != nil {
        return c.handleInstallError(err, output, name)
    }
    return nil
}

func (c *CondaManager) ListInstalled(ctx context.Context) ([]string, error) {
    // List base environment packages only
    output, err := ExecuteCommand(ctx, c.binary, "list", baseEnvironmentFlag, "--json")
    if err != nil {
        return nil, fmt.Errorf("failed to list conda packages: %w", err)
    }
    return c.parseListOutput(output), nil
}
```

#### Registration
```go
func init() {
    RegisterManager("conda", func() PackageManager {
        return NewCondaManager()
    })
}
```

### Phase 2: Required Operations (1-2 days)

#### Health Checking
```go
func (c *CondaManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
    check := &HealthCheck{
        Name:     "Conda Manager",
        Category: "package-managers",
        Status:   "pass",
        Message:  "Conda is available and properly configured",
    }

    // Check availability
    available, err := c.IsAvailable(ctx)
    if err != nil {
        if IsContextError(err) {
            return nil, err
        }
        check.Status = "warn"
        check.Message = "Conda availability check failed"
        check.Issues = []string{fmt.Sprintf("Error checking conda: %v", err)}
        return check, nil
    }

    if !available {
        check.Status = "warn"
        check.Message = "Conda is not available"
        check.Issues = []string{"conda command not found"}
        check.Suggestions = []string{
            "Install Miniconda: https://docs.conda.io/en/latest/miniconda.html",
            "Install Anaconda: https://www.anaconda.com/products/distribution",
            "Install via Homebrew: brew install --cask anaconda",
            "Install Mamba (faster alternative): brew install mambaforge",
        }
        return check, nil
    }

    // Get conda info for detailed diagnostics
    info, err := c.getCondaInfo(ctx)
    if err != nil {
        check.Status = "warn"
        check.Message = "Could not retrieve conda information"
        check.Issues = []string{fmt.Sprintf("Error getting conda info: %v", err)}
        return check, nil
    }

    // Add conda version and environment details
    check.Details = append(check.Details, fmt.Sprintf("Conda version: %s", info.CondaVersion))
    check.Details = append(check.Details, fmt.Sprintf("Python version: %s", info.PythonVersion))
    check.Details = append(check.Details, fmt.Sprintf("Base environment: %s", info.BaseEnvironment))
    check.Details = append(check.Details, fmt.Sprintf("Platform: %s", info.Platform))

    // Check channels configuration
    if len(info.Channels) > 0 {
        check.Details = append(check.Details, fmt.Sprintf("Configured channels: %d", len(info.Channels)))
        // Show top 3 channels
        channelStr := strings.Join(info.Channels[:min(3, len(info.Channels))], ", ")
        check.Details = append(check.Details, fmt.Sprintf("Primary channels: %s", channelStr))
    }

    // Validate base environment access
    if info.BaseEnvironment == "" {
        check.Status = "warn"
        check.Message = "Conda base environment not properly configured"
        check.Issues = []string{"Base environment path not detected"}
        check.Suggestions = []string{"Reinitialize conda: conda init"}
    }

    return check, nil
}

// getCondaInfo retrieves comprehensive conda system information
func (c *CondaManager) getCondaInfo(ctx context.Context) (*CondaInfo, error) {
    output, err := ExecuteCommand(ctx, c.binary, "info", "--json")
    if err != nil {
        return nil, fmt.Errorf("failed to get conda info: %w", err)
    }

    var info CondaInfo
    if err := json.Unmarshal(output, &info); err != nil {
        return nil, fmt.Errorf("failed to parse conda info: %w", err)
    }

    return &info, nil
}
```

#### Self-Installation
```go
func (c *CondaManager) SelfInstall(ctx context.Context) error {
    // Check if already available (idempotent)
    if available, _ := c.IsAvailable(ctx); available {
        return nil
    }

    // Multiple installation methods with fallbacks
    methods := []struct {
        name string
        fn   func(context.Context) error
    }{
        {"Homebrew (Miniconda)", c.installViaHomebrew},
        {"Official Miniconda Script", c.installViaMinicondaScript},
    }

    var lastErr error
    for _, method := range methods {
        err := method.fn(ctx)
        if err == nil {
            return nil // Success
        }
        lastErr = err
    }

    // All methods failed
    return fmt.Errorf("failed to install conda via any available method - last error: %w", lastErr)
}

// installViaHomebrew installs Miniconda via Homebrew
func (c *CondaManager) installViaHomebrew(ctx context.Context) error {
    if available, _ := checkPackageManagerAvailable(ctx, "brew"); !available {
        return fmt.Errorf("homebrew not available")
    }
    return executeInstallCommand(ctx, "brew", []string{"install", "--cask", "miniconda"}, "Miniconda")
}

// installViaMinicondaScript uses the official Miniconda installer
func (c *CondaManager) installViaMinicondaScript(ctx context.Context) error {
    // Platform-specific installation
    script := c.getMinicondaInstallScript()
    return executeInstallScript(ctx, script, "Miniconda")
}

// getMinicondaInstallScript returns the appropriate installation script for the platform
func (c *CondaManager) getMinicondaInstallScript() string {
    // Simplified cross-platform script
    return `curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-$(uname)-$(uname -m).sh -o /tmp/miniconda.sh && bash /tmp/miniconda.sh -b && rm /tmp/miniconda.sh`
}
```

#### Channel Management
```go
// ensureCondaForge ensures conda-forge channel is available (best practice)
func (c *CondaManager) ensureCondaForge(ctx context.Context) error {
    // Check if conda-forge is already configured
    output, err := ExecuteCommand(ctx, c.binary, "config", "--show", "channels")
    if err != nil {
        return fmt.Errorf("failed to check conda channels: %w", err)
    }

    if !strings.Contains(string(output), "conda-forge") {
        // Add conda-forge channel
        _, err := ExecuteCommandCombined(ctx, c.binary, "config", "--add", "channels", "conda-forge")
        if err != nil {
            return fmt.Errorf("failed to add conda-forge channel: %w", err)
        }
    }

    return nil
}
```

### Phase 3: Advanced Features (1 day)

#### Search Implementation
```go
func (c *CondaManager) Search(ctx context.Context, query string) ([]string, error) {
    // Use conda search with minimal output
    output, err := ExecuteCommand(ctx, c.binary, "search", query, "--json")
    if err != nil {
        // Check for no results vs real errors
        if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
            outputStr := string(output)
            if strings.Contains(outputStr, "no packages found") ||
               strings.Contains(outputStr, "PackagesNotFoundError") {
                return []string{}, nil // No results is not an error
            }
        }
        return nil, fmt.Errorf("conda search failed for '%s': %w", query, err)
    }

    return c.parseSearchOutput(output), nil
}

// parseSearchOutput parses conda search JSON output
func (c *CondaManager) parseSearchOutput(output []byte) []string {
    var searchResults map[string][]CondaSearchInfo
    if err := json.Unmarshal(output, &searchResults); err != nil {
        return []string{} // Parsing error returns empty results
    }

    // Extract unique package names
    packages := make(map[string]bool)
    for packageName := range searchResults {
        packages[packageName] = true
    }

    // Convert to sorted slice
    result := make([]string, 0, len(packages))
    for pkg := range packages {
        result = append(result, pkg)
    }
    sort.Strings(result)

    return result
}
```

#### Info Implementation
```go
func (c *CondaManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
    // Check if package is installed first
    installed, err := c.IsInstalled(ctx, name)
    if err != nil {
        return nil, fmt.Errorf("failed to check package installation status for %s: %w", name, err)
    }

    // Get package information from conda search
    output, err := ExecuteCommand(ctx, c.binary, "search", name, "--info", "--json")
    if err != nil {
        if exitCode, ok := ExtractExitCode(err); ok && exitCode == 1 {
            return nil, fmt.Errorf("package '%s' not found", name)
        }
        return nil, fmt.Errorf("failed to get package info for %s: %w", name, err)
    }

    info := c.parseInfoOutput(output, name)
    if info == nil {
        return nil, fmt.Errorf("package '%s' not found", name)
    }

    info.Manager = "conda"
    info.Installed = installed

    return info, nil
}

// parseInfoOutput parses conda search --info JSON output
func (c *CondaManager) parseInfoOutput(output []byte, name string) *PackageInfo {
    var searchResults map[string][]CondaSearchInfo
    if err := json.Unmarshal(output, &searchResults); err != nil {
        return nil
    }

    // Find the requested package
    packages, exists := searchResults[name]
    if !exists || len(packages) == 0 {
        return nil
    }

    // Use the latest version (first in list)
    pkg := packages[0]

    return &PackageInfo{
        Name:        pkg.Name,
        Version:     pkg.Version,
        Description: pkg.Summary,
        Homepage:    pkg.Homepage,
        Manager:     "conda",
        Dependencies: pkg.Dependencies,
        InstalledSize: fmt.Sprintf("%d bytes", pkg.Size),
    }
}
```

### Phase 4: Testing & Integration (1 day)

#### Unit Test Coverage
```go
func TestCondaManager_ListInstalled(t *testing.T) {
    tests := []struct {
        name           string
        mockOutput     []byte
        mockError      error
        expectedPkgs   []string
        expectedError  bool
    }{
        {
            name: "successful list with packages",
            mockOutput: []byte(`[
                {
                    "name": "numpy",
                    "version": "1.24.3",
                    "build": "py311h08b1b3b_0",
                    "channel": "conda-forge"
                },
                {
                    "name": "pandas",
                    "version": "2.0.3",
                    "build": "py311hd9cd6c9_0",
                    "channel": "conda-forge"
                }
            ]`),
            expectedPkgs: []string{"numpy", "pandas"},
        },
        {
            name: "empty environment",
            mockOutput: []byte(`[]`),
            expectedPkgs: []string{},
        },
        // Additional test cases for error conditions...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := NewCondaManager()
            // Use mock executor
            // Test implementation...
        })
    }
}

func TestCondaManager_Search(t *testing.T) {
    tests := []struct {
        name           string
        query          string
        mockOutput     []byte
        mockError      error
        expectedPkgs   []string
        expectedError  bool
    }{
        {
            name:  "successful search",
            query: "numpy",
            mockOutput: []byte(`{
                "numpy": [
                    {
                        "name": "numpy",
                        "version": "1.24.3",
                        "build": "py311h08b1b3b_0",
                        "channel": "conda-forge"
                    }
                ]
            }`),
            expectedPkgs: []string{"numpy"},
        },
        // Additional search test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := NewCondaManager()
            // Test search implementation...
        })
    }
}
```

#### BATS Integration Tests
- Add BATS tests for conda package installation/uninstall scenarios
- Test conda integration with `plonk clone` workflows
- Verify channel configuration and package resolution
- Test base environment isolation from project environments
- Cross-platform compatibility testing via BATS

#### Test Helper Functions
```go
// Test helpers for conda operations
func setupMockCondaManager(t *testing.T) *CondaManager {
    return &CondaManager{binary: "mock-conda"}
}

func createMockCondaListOutput(packages map[string]string) []byte {
    var items []CondaListItem
    for name, version := range packages {
        items = append(items, CondaListItem{
            Name:    name,
            Version: version,
            Build:   "py311_0",
            Channel: "conda-forge",
        })
    }

    data, _ := json.Marshal(items)
    return data
}

func createMockCondaSearchOutput(packages map[string]string) []byte {
    results := make(map[string][]CondaSearchInfo)
    for name, version := range packages {
        results[name] = []CondaSearchInfo{{
            Name:    name,
            Version: version,
            Build:   "py311_0",
            Channel: "conda-forge",
            Summary: fmt.Sprintf("Test package %s", name),
        }}
    }

    data, _ := json.Marshal(results)
    return data
}
```

### Phase 5: Documentation & Polish (0.5 days)

#### Documentation Updates

**Update `docs/CLI.md`**:
```markdown
- `conda:` - Conda (scientific computing and data science packages)

Examples:
```bash
plonk install conda:numpy
plonk install conda:pandas
plonk install conda:matplotlib
```

**Update `docs/package-managers.md`**:
```markdown
### Current Implementation Status

#### 3. **Conda Global Support** - ✅ **IMPLEMENTED**
- **Status**: Complete support for base environment package management
- **Use Case**: Data science, machine learning, and scientific computing
- **Capabilities**: Handles both Python packages and system-level dependencies
- **Commands**: `conda install -n base`, `conda remove -n base`, `conda list -n base`
- **Unique Value**: Only Python package manager that manages non-Python dependencies
- **Examples**: `numpy`, `pandas`, `jupyter`, `matplotlib`, `scipy`
```

**Update `README.md`**:
```markdown
### Core Package Managers
- **Homebrew** (brew) - macOS/Linux packages and system tools
- **NPM** (npm) - Node.js packages (global)
- **PNPM** (pnpm) - Fast, efficient Node.js packages (global)
- **Cargo** (cargo) - Rust packages
- **Pipx** (pipx) - Python applications in isolated environments
- **Conda** (conda) - Scientific computing and data science packages
```

#### CLI Examples
```bash
# Python ecosystem coverage
plonk install pipx:black uv:ruff conda:numpy pixi:jupyter

# Data science workflow
plonk install conda:pandas conda:matplotlib conda:scipy conda:seaborn
```

#### Usage Guidelines
**Document conda vs other Python managers**:
```markdown
### Python Package Manager Selection Guide

- **pipx** - Isolated Python CLI applications (black, flake8, poetry)
- **uv** - Fast modern Python package management and project tools
- **conda** - Data science, ML, scientific computing (numpy, pandas, jupyter)
- **pixi** - Modern conda-forge alternative with better reproducibility

**When to use conda:**
- Installing data science packages (numpy, pandas, scipy)
- Need non-Python dependencies (C libraries, R packages)
- Scientific computing workflows
- Machine learning environments
- When packages require specific system-level dependencies
```

## Implementation Details

### File Structure
```
internal/resources/packages/
├── conda.go           # Main implementation (~800 lines)
├── conda_test.go      # Comprehensive unit tests (~1000 lines)
├── npm.go            # Reference for JSON parsing patterns
├── cargo.go          # Reference for self-installation patterns
└── ...
```

### Code Patterns to Follow
1. **Error Handling**: Follow npm.go patterns with conda-specific error messages
2. **JSON Parsing**: Robust parsing for conda's complex JSON structures
3. **Command Execution**: Use existing ExecuteCommand/ExecuteCommandCombined utilities
4. **Context Support**: Proper context cancellation handling for long operations
5. **Health Checking**: Comprehensive diagnostics following cargo.go patterns

### Performance Considerations
- Handle large conda package lists efficiently (conda environments can be large)
- Implement proper timeout handling for slow conda operations
- Cache conda info and base environment discovery
- Optimize JSON parsing for conda's verbose output structures

### Platform Considerations
- Support conda, mamba, and micromamba detection
- Handle different conda installation paths (Anaconda, Miniconda, Mambaforge)
- Cross-platform installer script generation
- Path configuration for conda binary discovery

## Open Questions & Design Considerations

### **Critical Design Questions for Intelligent Binary Detection**

#### **1. Detection Timing and Caching**
**Question**: When should variant detection occur?
- **Option A**: At package manager creation time (current proposal)
- **Option B**: Lazy detection on first method call
- **Option C**: On-demand detection with caching

**Trade-offs**:
- **Performance**: Eager detection adds startup cost, lazy detection adds first-call latency
- **Reliability**: User might install/uninstall variants between plonk operations
- **Complexity**: Caching introduces state management complexity

**Open Issues**:
- Should detection results be cached across plonk invocations?
- How do we handle the case where a user installs mamba after conda detection?
- What happens if the detected binary is uninstalled between operations?

#### **2. Binary Validation Strategy**
**Question**: How thoroughly should we validate detected binaries?
- **Option A**: Simple existence check (`CheckCommandAvailable`)
- **Option B**: Version check (`{binary} --version`)
- **Option C**: Full functionality check (`{binary} list -n base`)

**Open Issues**:
```go
func isCondaVariantFunctional(binary string) bool {
    // What level of validation is appropriate?
    // - Just version check? (fast, might miss broken installs)
    // - Test base environment access? (more reliable, slower)
    // - Test package listing? (comprehensive, potential side effects)
}
```

#### **3. Priority Order Rationale**
**Current Proposal**: mamba → micromamba → conda

**Questions**:
- Is speed the right primary criterion, or should we consider other factors?
- Should micromamba come before mamba due to minimal footprint?
- How do we handle user preference overrides?

**Alternative Priority Strategies**:
```go
// Strategy A: Speed-first (current)
priorities := []string{"mamba", "micromamba", "conda"}

// Strategy B: Compatibility-first
priorities := []string{"conda", "mamba", "micromamba"}

// Strategy C: User-configurable
priorities := getUserPreferredPriority()
```

#### **4. Error Handling and Fallbacks**
**Question**: What happens when the detected variant fails during operation?

**Scenarios**:
- Mamba detected but fails during `install` operation
- Should we fall back to conda, or report the mamba error?
- How do we distinguish between variant-specific vs general failures?

**Open Issues**:
```go
func (c *CondaManager) Install(ctx context.Context, name string) error {
    err := ExecuteCommand(ctx, c.binary, "install", "-n", "base", "-y", name)
    if err != nil {
        // Should we try fallback variants here?
        // Or is the detection decision final?
        return handleInstallError(err, name)
    }
}
```

#### **5. Lock File Representation**
**Question**: How should the lock file represent packages installed via different conda variants?

**Options**:
```yaml
# Option A: Unified (current proposal)
resources:
  - type: package
    id: conda:numpy
    metadata:
      manager: conda
      variant: mamba  # Track but don't distinguish

# Option B: Variant-aware
resources:
  - type: package
    id: conda:numpy
    metadata:
      manager: conda
      detected_binary: mamba

# Option C: Separate tracking
resources:
  - type: package
    id: conda:numpy
    metadata:
      manager: conda
      installation_method: mamba
```

#### **6. User Override Mechanisms**
**Question**: Should users be able to force a specific conda variant?

**Use Cases**:
- User prefers conda over mamba for compatibility
- CI environments want to test specific variants
- Debugging variant-specific issues

**Implementation Options**:
```go
// Option A: Environment variable
PLONK_CONDA_VARIANT=conda plonk install conda:numpy

// Option B: Configuration setting
// plonk.yaml
conda:
  preferred_variant: "conda"

// Option C: Command flag
plonk install conda:numpy --conda-variant=mamba
```

### **Architectural Impact Assessment**

#### **1. Precedent Setting**
**Impact**: This is the first package manager with intelligent detection in plonk.

**Questions**:
- Should this pattern be applied to other package managers?
- Could npm benefit from detecting npm vs yarn vs pnpm?
- How do we maintain consistency across the codebase?

#### **2. Testing Complexity**
**New Requirements**:
```go
// Need to test all variants
func TestCondaManager_WithMamba(t *testing.T) { }
func TestCondaManager_WithMicromamba(t *testing.T) { }
func TestCondaManager_WithConda(t *testing.T) { }

// Need to test detection logic
func TestCondaVariantDetection(t *testing.T) { }

// Need to test fallback scenarios
func TestDetectionFailures(t *testing.T) { }
```

#### **3. Documentation Burden**
**New Documentation Required**:
- How variant detection works
- Why certain variants are prioritized
- How to override detection
- Troubleshooting variant issues
- Performance implications of each variant

#### **4. Maintenance Complexity**
**Ongoing Responsibilities**:
- Monitor conda ecosystem changes (new variants?)
- Update priority logic as variants evolve
- Handle version-specific compatibility issues
- Test across all supported variants

### **Implementation Decisions Required**

#### **Phase 1 Decisions (Required for MVP)**
1. **Detection timing**: Eager vs lazy vs cached?
2. **Validation depth**: Simple check vs comprehensive validation?
3. **Priority order**: Speed-first vs compatibility-first?
4. **Error handling**: Fail-fast vs fallback strategy?

#### **Phase 2 Decisions (Can be deferred)**
1. **User overrides**: Environment variable vs config file vs command flag?
2. **Lock file representation**: How much variant detail to track?
3. **Performance optimization**: Detection result caching strategy?
4. **Ecosystem integration**: Apply pattern to other package managers?

### **Recommended Decision Framework**

#### **For MVP (Keep It Simple)**
1. **Eager detection** with simple validation
2. **Speed-first priority** (mamba → micromamba → conda)
3. **Fail-fast error handling** (no cross-variant fallbacks)
4. **Minimal variant tracking** in lock file
5. **No user overrides** initially

#### **Rationale**
- Establishes the pattern without over-engineering
- Provides immediate performance benefits
- Allows learning from real-world usage
- Enables iterative improvement based on user feedback

## Risk Analysis

### Technical Risks: **HIGH** (Elevated due to intelligent detection)
- **Conda Complexity**: More complex than other package managers due to environment model
- **Mitigation**: Focus on base environment only, comprehensive testing
- **JSON Parsing**: Conda outputs more complex JSON than npm/pnpm
- **Mitigation**: Robust error handling and parsing validation
- **🚨 NEW: Detection Logic Complexity**: First-of-its-kind intelligent binary detection
- **Mitigation**: Extensive testing across all variants, fail-fast approach, clear error messages

### Implementation Risks: **HIGH** (New architectural pattern)
- **Environment Model**: Conda's environment paradigm differs from plonk's global approach
- **Mitigation**: Clear documentation and base environment focus
- **🚨 NEW: Variant Detection**: Complex detection logic with multiple failure modes
- **Mitigation**: Systematic detection logic with comprehensive validation
- **🚨 NEW: Testing Complexity**: Need to test 3 variants × multiple platforms × various scenarios
- **Mitigation**: Structured test matrix, mock-based testing, BATS integration
- **🚨 NEW: Precedent Setting**: This pattern may be applied to other package managers
- **Mitigation**: Design for reusability, document patterns clearly

### User Experience Risks: **MEDIUM** (Elevated due to detection transparency)
- **Confusion**: Users might expect full environment management
- **Mitigation**: Clear documentation about base environment approach
- **🚨 NEW: Detection Transparency**: Users may be confused about which variant is being used
- **Mitigation**: Clear status messages, health check reporting, documentation
- **🚨 NEW: Inconsistent Performance**: Different variants have different performance characteristics
- **Mitigation**: Document performance differences, consistent timeout handling
- **Performance**: Conda can be slower than other Python package managers
- **Mitigation**: Intelligent detection prioritizes faster variants (mamba, micromamba)

### Operational Risks: **MEDIUM** (New category)
- **🚨 NEW: Variant Ecosystem Changes**: New conda variants may appear, existing ones may be deprecated
- **Mitigation**: Monitor conda ecosystem, modular detection design
- **🚨 NEW: Priority Order Maintenance**: Detection priority may need updates as ecosystem evolves
- **Mitigation**: Document rationale, make priorities configurable in future
- **🚨 NEW: Cross-Platform Behavior**: Detection may behave differently across platforms
- **Mitigation**: Platform-specific testing, consistent detection logic

## Success Criteria

### Functional Requirements
- ✅ Install/uninstall global packages in conda base environment
- ✅ List installed packages with versions from base environment
- ✅ Check package installation status in base environment
- ✅ Search conda package repositories
- ✅ Retrieve package information including dependencies
- ✅ Upgrade packages to latest versions
- ✅ Health checking and diagnostics with conda-specific details
- ✅ Self-installation via multiple methods (Homebrew, official installer)

### Non-Functional Requirements
- **Performance**: Handle conda's slower operations within timeout limits
- **Reliability**: Robust error handling for conda's complex failure modes
- **Usability**: Clear guidance on conda vs other Python package managers
- **Maintainability**: Code follows established patterns with conda-specific adaptations

### Integration Requirements
- ✅ Works with plonk's lock file system for conda packages
- ✅ Integrates with plonk's orchestration layer
- ✅ Compatible with existing CLI commands (`plonk install conda:numpy`)
- ✅ Supports all output formats (table, JSON, YAML)
- ✅ Proper prefix handling (`conda:` prefix recognition)

## Timeline

**Updated Timeline**: Adjusted for intelligent binary detection complexity

### Week 1: Core Implementation with Detection
- **Day 1**: Design and implement intelligent binary detection system
- **Day 2**: Core CondaManager with variant-aware operations (Phase 1)
- **Day 3**: Health checking with variant reporting, basic self-installation (Phase 2)

### Week 2: Advanced Features and Comprehensive Testing
- **Day 1**: Advanced features - search, info, variant-aware self-installation (Phase 3)
- **Day 2**: Comprehensive testing across all variants - unit tests (Phase 4)
- **Day 3**: Integration testing and variant detection edge cases

### Week 3: Documentation and Validation
- **Day 1**: Documentation updates including detection behavior (Phase 5)
- **Day 2**: BATS integration testing across conda variants
- **Day 3**: Cross-platform validation and performance testing

### Week 4: Review and Polish
- **Day 1**: Code review with focus on detection logic
- **Day 2**: Address open questions from detection design
- **Day 3**: Final testing and release preparation

**Rationale for Extended Timeline**:
- Intelligent binary detection adds ~2 days of implementation
- Comprehensive testing across 3 variants adds ~2 days
- Additional documentation and design validation adds ~1 day
- **Total**: 4-5 days becomes 6-7 days (40% increase in complexity)

## Future Considerations

### Post-Implementation Enhancements
1. **Channel Management**: Advanced channel configuration and management
2. **Environment Integration**: Consider supporting named environments (beyond base)
3. **Performance Optimization**: Mamba detection and usage for faster operations
4. **Build Support**: Integration with conda-build for package creation
5. **Lock File Integration**: Enhanced integration with conda-lock format

### Monitoring & Maintenance
1. **Usage Analytics**: Track conda adoption among plonk users vs other Python managers
2. **Performance Metrics**: Monitor conda operation times and timeout issues
3. **Error Monitoring**: Track common failure patterns specific to conda
4. **Channel Health**: Monitor default channel availability and performance
5. **Version Compatibility**: Track conda/mamba/micromamba version compatibility

### Ecosystem Integration
1. **Jupyter Support**: Special handling for Jupyter ecosystem packages
2. **R Integration**: Consider R package management through conda
3. **GPU Packages**: Special consideration for CUDA/GPU-enabled packages
4. **Bioinformatics**: Integration with bioconda channel for life sciences

## Conclusion

Adding conda support to plonk addresses a significant gap in coverage of the data science and scientific computing ecosystem. While conda presents more implementation complexity than other package managers due to its environment model and rich feature set, the strategic value is substantial.

### Key Benefits
- **Market Coverage**: Serves the massive data science and research communities
- **Unique Capabilities**: Only package manager handling both Python and system dependencies
- **Enterprise Value**: Extends plonk into scientific computing and enterprise data science
- **Complete Python Ecosystem**: Completes comprehensive Python package management coverage

### Implementation Strategy
- **Base Environment Focus**: Aligns conda with plonk's global package paradigm
- **Multi-Variant Support**: Supports conda, mamba, and micromamba for flexibility
- **Comprehensive Testing**: Robust testing strategy addresses conda's complexity
- **Clear Documentation**: Guidance helps users choose between Python package managers

The comprehensive implementation plan, detailed risk analysis, and clear success criteria make conda support a valuable addition to plonk's package manager ecosystem, serving users in data science, research, and scientific computing workflows while maintaining compatibility with plonk's design principles.
