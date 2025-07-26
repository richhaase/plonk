# Phase 13: Search and Info Commands Enhancement

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase enhances the search and info commands to be more useful and performant. Search will query all managers in parallel with a timeout, and info will show the most relevant information based on package state (managed → installed → available).

## Objectives

1. Implement parallel search across all package managers
2. Add timeout handling (2-3 seconds) for search operations
3. Update info command with smart priority logic
4. Support prefix syntax from Phase 12
5. Display results in clear table format showing sources

## Current State

- Search requires manager flags to search specific managers
- Search is sequential, can be slow
- Info command shows basic package information
- No clear indication of package state (managed/installed/available)

## Implementation Details

### 1. Parallel Search Implementation

**In `internal/commands/search.go`:**

```go
func searchAllManagers(ctx context.Context, query string) []SearchResult {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    // Channel for results
    resultsChan := make(chan SearchResult, len(managers)*10)
    var wg sync.WaitGroup

    // Search each manager in parallel
    for _, mgr := range managers {
        wg.Add(1)
        go func(m Manager) {
            defer wg.Done()
            results := m.Search(ctx, query)
            for _, r := range results {
                select {
                case resultsChan <- r:
                case <-ctx.Done():
                    return
                }
            }
        }(mgr)
    }

    // Close channel when done
    go func() {
        wg.Wait()
        close(resultsChan)
    }()

    // Collect results
    var allResults []SearchResult
    for r := range resultsChan {
        allResults = append(allResults, r)
    }

    return allResults
}
```

### 2. Search Result Display

Show results grouped by manager or in a unified table:

```
Searching for "ripgrep"...

PACKAGE    MANAGER    VERSION    DESCRIPTION
ripgrep    homebrew   14.0.3     Recursively search directories
ripgrep    cargo      14.0.3     Fast grep replacement
rg         scoop      14.0.3     Ripgrep for Windows
```

### 3. Info Command Priority Logic

**In `internal/commands/info.go`:**

Priority order:
1. If package is managed by plonk → show managed info
2. Else if package is installed → show installed info
3. Else search all managers → show available info

```go
func getPackageInfo(name string) (*PackageInfo, error) {
    // Check if prefix syntax used
    manager, pkgName := ParsePackageSpec(name)

    // 1. Check if managed by plonk
    if state := getManagedState(pkgName); state != nil {
        return &PackageInfo{
            Source: "managed",
            Manager: state.Manager,
            // ... other fields
        }, nil
    }

    // 2. Check if installed (search installed packages)
    if info := getInstalledInfo(pkgName); info != nil {
        return &PackageInfo{
            Source: "installed",
            // ... other fields
        }, nil
    }

    // 3. Search all available
    results := searchAllManagers(ctx, pkgName)
    if len(results) > 0 {
        return &PackageInfo{
            Source: "available",
            // ... from first/best result
        }, nil
    }

    return nil, fmt.Errorf("package %q not found", name)
}
```

### 4. Info Output Format

Clear indication of package state:

```
Package: ripgrep
Status: Managed by plonk
Manager: homebrew
Version: 14.0.3 (installed)
Latest: 14.0.3
Description: Recursively search directories
```

Or for available packages:

```
Package: ripgrep
Status: Available (not installed)
Manager: homebrew
Version: 14.0.3 (latest)
Description: Recursively search directories
```

### 5. Prefix Syntax Support

Both commands should support the prefix syntax:
- `plonk search brew:rip` - Search only in brew
- `plonk info npm:typescript` - Get info specifically from npm

When prefix is used:
- Search: Only search that specific manager
- Info: Only check that specific manager (skip priority logic)

### 6. Error Handling

- Handle timeout gracefully: "Search timed out after 3 seconds. Partial results shown."
- Handle manager failures: Continue with other managers
- Clear message when nothing found

## Testing Requirements

### Unit Tests
- Test parallel search with mock managers
- Test timeout handling
- Test info priority logic
- Test prefix syntax parsing

### Integration Tests
1. Search returns results from multiple managers
2. Search respects timeout (use slow mock manager)
3. Info shows managed packages first
4. Info with prefix only checks specific manager
5. Search with prefix only searches one manager
6. Error handling for timeouts and failures

### Manual Testing
- Search for common packages
- Test info on managed, installed, and available packages
- Verify timeout behavior with slow network
- Test prefix syntax variations

## Expected Changes

1. **Modified files:**
   - `internal/commands/search.go` - Parallel implementation
   - `internal/commands/info.go` - Priority logic
   - Both: Remove manager flags, add prefix support

2. **Performance changes:**
   - Search is now parallel (much faster)
   - 3-second timeout prevents hanging

3. **Output changes:**
   - Search shows results from all managers
   - Info clearly indicates package state
   - Better error messages

## Note for Phase 15 (Output Standardization)

The user asked about output preferences for the info command. This should be addressed in Phase 15:
- Unified view vs separated sections
- Verbose vs minimal by default
- Consistent table formatting across commands

For now, implement a reasonable default that can be refined in Phase 15.

## Validation Checklist

Before marking complete:
- [ ] Search queries all managers in parallel
- [ ] 3-second timeout works correctly
- [ ] Search results show manager source clearly
- [ ] Info prioritizes managed → installed → available
- [ ] Prefix syntax works for both commands
- [ ] All manager flags removed
- [ ] Clear indication of package state in info
- [ ] Timeout/error messages are helpful
- [ ] All tests updated and passing

## Notes

- Parallel search will significantly improve performance
- The timeout prevents slow managers from blocking results
- Info priority ensures users see the most relevant information
- This completes the prefix syntax migration started in Phase 12
- Output format details will be refined in Phase 15

Remember to create `PHASE_13_COMPLETION.md` when finished!
