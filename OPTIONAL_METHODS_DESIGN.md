# Optional Methods Design for Package Managers

## Problem Statement

Currently, all methods in the `PackageManager` interface are required, but not all package managers support all operations. This leads to:

1. **Inconsistent error handling** - Some managers return `ErrUnsupported`, others use `ErrCommandExecution`
2. **Poor user experience** - Users don't know which operations are supported until they try them
3. **Interface pollution** - Managers must implement methods they don't support

### Current State

- **Go Install**: Search returns `ErrUnsupported` (go has no search command)
- **Pip**: Search returns `ErrCommandExecution` (pip search is deprecated)
- **Others**: All other operations are universally supported

## Proposed Solution

### Option 1: Capability Interface (Recommended)

Add a capability discovery mechanism using a separate interface:

```go
// PackageManagerCapabilities describes which optional operations are supported
type PackageManagerCapabilities interface {
    // SupportsSearch returns true if the manager supports searching for packages
    SupportsSearch() bool

    // Future optional operations can be added here
    // SupportsUpgrade() bool
    // SupportsDependencyTree() bool
}

// Update PackageManager to embed capabilities
type PackageManager interface {
    PackageManagerCapabilities

    // Core operations (always supported)
    IsAvailable(ctx context.Context) (bool, error)
    ListInstalled(ctx context.Context) ([]string, error)
    Install(ctx context.Context, name string) error
    Uninstall(ctx context.Context, name string) error
    IsInstalled(ctx context.Context, name string) (bool, error)
    GetInstalledVersion(ctx context.Context, name string) (string, error)

    // Optional operations (check capabilities first)
    Search(ctx context.Context, query string) ([]string, error)
    Info(ctx context.Context, name string) (*PackageInfo, error)
}
```

### Option 2: Optional Interface Pattern

Use Go's interface composition to make Search optional:

```go
// Searcher is an optional interface for package managers that support search
type Searcher interface {
    Search(ctx context.Context, query string) ([]string, error)
}

// Core PackageManager interface with only required methods
type PackageManager interface {
    IsAvailable(ctx context.Context) (bool, error)
    ListInstalled(ctx context.Context) ([]string, error)
    Install(ctx context.Context, name string) error
    Uninstall(ctx context.Context, name string) error
    IsInstalled(ctx context.Context, name string) (bool, error)
    GetInstalledVersion(ctx context.Context, name string) (string, error)
    Info(ctx context.Context, name string) (*PackageInfo, error)
}
```

### Option 3: Sentinel Error Pattern

Standardize on returning a specific error for unsupported operations:

```go
var ErrOperationNotSupported = errors.New("operation not supported")

// Managers return this error for unsupported operations
func (g *GoInstallManager) Search(ctx context.Context, query string) ([]string, error) {
    return nil, ErrOperationNotSupported
}
```

## Recommendation: Option 1 - Capability Interface

This approach provides:

1. **Clear capability discovery** - Callers can check support before calling
2. **Better UX** - UI can disable/hide unsupported operations
3. **Type safety** - Methods remain on the interface
4. **Extensibility** - Easy to add new optional operations
5. **Backward compatibility** - Existing code continues to work

## Implementation Plan

### Phase 1: Add Capability Interface

1. Add `PackageManagerCapabilities` interface to `internal/interfaces/`
2. Update `PackageManager` interface to embed it
3. Add `SupportsSearch()` method to `BaseManager` (default: true)
4. Override in managers that don't support search (Go, Pip)

### Phase 2: Update Managers

1. Go Install: `SupportsSearch() = false`, Search returns standard error
2. Pip: `SupportsSearch() = false`, Search returns standard error
3. Others: `SupportsSearch() = true`, no changes needed

### Phase 3: Update Command Layer

1. Check capabilities before calling Search
2. Provide clear user messaging when operation not supported
3. Update help/documentation to show supported operations

### Phase 4: Standardize Error Handling

1. Create `ErrOperationNotSupported` in errors package
2. Update all managers to return this for unsupported operations
3. Include helpful suggestion messages

## Example Implementation

```go
// In BaseManager
func (b *BaseManager) SupportsSearch() bool {
    return true // Default: most managers support search
}

// In GoInstallManager
func (g *GoInstallManager) SupportsSearch() bool {
    return false
}

func (g *GoInstallManager) Search(ctx context.Context, query string) ([]string, error) {
    return nil, errors.NewError(errors.ErrOperationNotSupported, errors.DomainPackages, "search",
        "go does not support package search").
        WithSuggestionMessage(fmt.Sprintf("Search at https://pkg.go.dev/search?q=%s", query))
}

// In command layer
func searchPackages(manager interfaces.PackageManager, query string) error {
    if !manager.SupportsSearch() {
        fmt.Printf("Search is not supported by %s\n", manager.Name())
        // Could also get suggestion from Search() error
        return nil
    }

    results, err := manager.Search(ctx, query)
    // ... handle results
}
```

## Benefits

1. **Clear contracts** - Interface explicitly shows what's optional
2. **Better errors** - Consistent error handling for unsupported operations
3. **Improved UX** - Users know upfront what's supported
4. **Future-proof** - Easy to add more optional operations
5. **Type-safe** - No runtime type assertions needed

## Migration Path

1. Add capability interface (non-breaking)
2. Update managers to implement it (non-breaking)
3. Update commands to check capabilities (improved UX)
4. Deprecate returning errors for capability checking (future)

## Future Considerations

Other potentially optional operations:
- `Upgrade(name string)` - Upgrade a specific package
- `UpgradeAll()` - Upgrade all packages
- `ShowDependencies(name string)` - Show dependency tree
- `ShowReverseDependencies(name string)` - Show what depends on package
- `Verify()` - Verify package integrity
- `Clean()` - Clean package cache
