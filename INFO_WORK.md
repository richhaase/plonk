# Info Command Implementation Considerations

## Phase 15.5 Implementation Guidance

**IMPORTANT**: For Phase 15.5, implement only the basic info structure with version/status information. Rich metadata (descriptions, latest versions) is deferred to future work.

## Overview

The Phase 15 plan specifies detailed designs for the `plonk info` command output based on package state (managed, installed but not managed, available). This document analyzes the implementation considerations and architectural requirements.

## Current State Analysis

### Existing Info Command Implementation
- **File**: `internal/commands/info.go`
- **Current Approach**: Uses priority logic (managed → installed → available)
- **Manager Support**: Searches specific manager or uses priority across all managers
- **Output**: Basic `InfoOutput` structure with status-based messages

### Current Information Sources
1. **Managed packages**: From plonk lock file via lock service
2. **Installed packages**: Via package manager's `ListInstalled()` method
3. **Available packages**: Via package manager's `Search()` method (limited info)

## Proposed Design Requirements

The Phase 15 plan specifies three distinct output formats:

### 1. For Managed Packages
```
Package: ripgrep
Status: Managed by plonk
Manager: homebrew
Installed: 14.0.3
Latest: 14.0.3
Description: Recursively search directories for patterns
```

### 2. For Installed (Not Managed)
```
Package: wget
Status: Installed (not managed)
Manager: homebrew
Version: 1.21.4
Description: Internet file retriever
Note: Run 'plonk install brew:wget' to manage this package
```

### 3. For Available (Not Installed)
```
Package: jq
Status: Available
Manager: homebrew
Version: 1.7 (latest)
Description: Command-line JSON processor
Install: plonk install brew:jq
```

## Implementation Challenges

### Challenge 1: Package Description Retrieval
**Issue**: Most package managers don't provide rich metadata through their CLI interfaces.

**Current Capabilities by Manager**:
- **Homebrew**: `brew info <package>` provides description, homepage, version
- **npm**: `npm view <package>` provides description, version, homepage
- **pip**: `pip show <package>` provides summary, version (for installed only)
- **cargo**: Limited metadata available through CLI
- **gem**: `gem spec <package>` provides description
- **go**: No standard description mechanism

**Recommendation**: Implement progressive enhancement:
1. Start with basic version/status info
2. Add description support where package managers provide it
3. Consider this a future enhancement rather than blocking requirement

### Challenge 2: "Latest Version" Retrieval
**Issue**: Determining latest available version requires additional API calls.

**Current Capabilities**:
- **Homebrew**: `brew info <package>` shows latest version
- **npm**: `npm view <package> version` shows latest
- **pip**: Requires separate `pip index versions <package>` call
- **Others**: Limited or no support

**Recommendation**:
1. Implement where straightforward (brew, npm)
2. Mark as "unknown" or omit for managers without easy access
3. Consider this an enhancement feature

### Challenge 3: Package Manager Interface Extensions
**Issue**: Current `PackageManager` interface may need extensions for info retrieval.

**Current Interface** (estimated):
```go
type PackageManager interface {
    IsAvailable(ctx context.Context) (bool, error)
    Search(ctx context.Context, query string) ([]string, error)
    ListInstalled(ctx context.Context) ([]string, error)
    // ... install/uninstall methods
}
```

**Needed Extensions**:
```go
type PackageManager interface {
    // ... existing methods
    GetPackageInfo(ctx context.Context, packageName string) (*PackageInfo, error)
}

type PackageInfo struct {
    Name         string
    Version      string
    LatestVersion string
    Description  string
    Homepage     string
    IsInstalled  bool
    IsManaged    bool  // determined by plonk, not manager
}
```

### Challenge 4: Cross-Manager Priority Logic
**Issue**: When no manager prefix is specified, determining which manager to query.

**Current Approach**: Check each manager in sequence until found
**Proposed Enhancement**:
1. Check managed packages first (from lock file)
2. Check installed packages across all managers
3. Search available packages across all managers
4. Present results with clear manager indication

## Architectural Considerations

### 1. Interface Design Decision
**Option A**: Extend existing `PackageManager` interface
- Pros: Consistent with current architecture
- Cons: Requires updating all manager implementations

**Option B**: Create separate `PackageInfoProvider` interface
- Pros: Optional implementation, gradual rollout
- Cons: Additional interface complexity

**Recommendation**: Start with Option A, implement basic info for core managers (brew, npm)

### 2. Performance Considerations
- Package info retrieval may be slower than simple search/list operations
- Consider implementing timeout/cancellation
- Cache results for repeated queries in same session

### 3. Error Handling Strategy
- Manager unavailable: Clear error message
- Package not found: Suggest similar packages or managers
- Network/timeout errors: Graceful degradation

## Implementation Phases

### Phase 1: Basic Structure (Immediate)
1. Update `InfoOutput` to support detailed package information
2. Implement basic package info for Homebrew and npm
3. Add "not implemented" responses for other managers
4. Focus on table output format standardization

### Phase 2: Enhanced Information (Future)
1. Add description/homepage retrieval where available
2. Implement latest version checking
3. Add cross-manager search when package not found in specified manager

### Phase 3: Advanced Features (Future)
1. Package dependency information
2. Installation size/requirements
3. Security/vulnerability information

## Phase 15.5 Implementation Plan

**Worker Guidance**:
1. Implement basic info command structure with standardized output
2. Support version and basic status information only
3. Use existing package manager methods (ListInstalled, Search)
4. Rich metadata (descriptions, latest versions) is explicitly deferred
5. Focus on consistent table/JSON/YAML formatting using StandardTableBuilder

**What to Implement**:
- Basic package status (managed/installed/available)
- Version information for installed packages
- Installation hints for available packages
- Manager indication when found
- Consistent error messages using the new error formatting utilities

**What NOT to Implement**:
- Package descriptions
- Latest version checking
- Homepage URLs
- Dependency information
- New PackageManager interface methods

## Alternative Approach

If rich metadata is deemed essential for Phase 15:
1. Implement for Homebrew and npm first (most capable package managers)
2. Provide graceful degradation for other managers
3. Accept that some managers will show limited information initially
4. Document the limitation and plan for future enhancement

This approach balances the Phase 15 requirements with implementation complexity and maintains consistent output formatting across all commands.
