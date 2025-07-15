# Using Package Manager Capabilities in Commands

This document shows how to use the new capability discovery feature in the command layer.

## Example: Search Command Implementation

```go
// internal/commands/package_search.go

func runPackageSearch(cmd *cobra.Command, args []string) error {
    if len(args) < 2 {
        return errors.NewError(errors.ErrInvalidInput, errors.DomainCommands, "search",
            "missing required arguments: MANAGER QUERY")
    }

    managerName := args[0]
    query := args[1]

    // Get the package manager
    registry := managers.NewManagerRegistry()
    manager, err := registry.GetManager(managerName)
    if err != nil {
        return err
    }

    // Check if manager is available
    ctx := context.Background()
    available, err := manager.IsAvailable(ctx)
    if err != nil {
        return errors.Wrap(err, errors.ErrCommandExecution, errors.DomainCommands, "search",
            fmt.Sprintf("failed to check %s availability", managerName))
    }
    if !available {
        return errors.NewError(errors.ErrManagerUnavailable, errors.DomainCommands, "search",
            fmt.Sprintf("%s is not available on this system", managerName))
    }

    // Check if manager supports search
    if !manager.SupportsSearch() {
        // Get the error with suggestion from the manager
        _, err := manager.Search(ctx, query)
        if plonkErr, ok := err.(*errors.PlonkError); ok {
            // Return the manager's error which includes helpful suggestions
            return plonkErr
        }
        // Fallback error if something unexpected happens
        return errors.NewError(errors.ErrOperationNotSupported, errors.DomainCommands, "search",
            fmt.Sprintf("search is not supported by %s", managerName))
    }

    // Perform the search
    results, err := manager.Search(ctx, query)
    if err != nil {
        return err
    }

    // Display results
    if len(results) == 0 {
        fmt.Printf("No packages found matching '%s'\n", query)
    } else {
        fmt.Printf("Found %d packages:\n", len(results))
        for _, pkg := range results {
            fmt.Printf("  %s\n", pkg)
        }
    }

    return nil
}
```

## Example: Status Command Showing Capabilities

```go
// internal/commands/status.go

func showManagerCapabilities(registry *managers.ManagerRegistry) error {
    ctx := context.Background()

    fmt.Println("Package Manager Capabilities:")
    fmt.Println("=============================")

    for _, name := range registry.GetAllManagerNames() {
        manager, err := registry.GetManager(name)
        if err != nil {
            continue
        }

        available, _ := manager.IsAvailable(ctx)
        supportsSearch := manager.SupportsSearch()

        status := "✗ Not Available"
        if available {
            status = "✓ Available"
        }

        searchStatus := "✓"
        if !supportsSearch {
            searchStatus = "✗"
        }

        fmt.Printf("%-12s %s    Search: %s\n", name, status, searchStatus)
    }

    return nil
}
```

## Example: Interactive Mode

```go
// internal/commands/interactive.go

func handleSearchCommand(registry *managers.ManagerRegistry, input string) error {
    parts := strings.Fields(input)
    if len(parts) < 3 {
        return fmt.Errorf("usage: search <manager> <query>")
    }

    managerName := parts[1]
    query := strings.Join(parts[2:], " ")

    manager, err := registry.GetManager(managerName)
    if err != nil {
        return err
    }

    // Check capability first
    if !manager.SupportsSearch() {
        // Provide user-friendly message
        fmt.Printf("Search is not supported by %s.\n", managerName)

        // Get suggestion from the manager's error
        ctx := context.Background()
        _, searchErr := manager.Search(ctx, query)
        if plonkErr, ok := searchErr.(*errors.PlonkError); ok && plonkErr.Suggestion != nil {
            fmt.Printf("Suggestion: %s\n", plonkErr.Suggestion.Message)
        }

        return nil
    }

    // Proceed with search...
    return performSearch(manager, query)
}
```

## Example: Help Text Generation

```go
// internal/commands/help.go

func generateManagerHelp(manager interfaces.PackageManager, name string) string {
    var help strings.Builder

    help.WriteString(fmt.Sprintf("Package Manager: %s\n", name))
    help.WriteString("Supported Operations:\n")
    help.WriteString("  - List installed packages\n")
    help.WriteString("  - Install packages\n")
    help.WriteString("  - Uninstall packages\n")
    help.WriteString("  - Check if package is installed\n")
    help.WriteString("  - Get package information\n")

    if manager.SupportsSearch() {
        help.WriteString("  - Search for packages\n")
    } else {
        help.WriteString("  - Search for packages (NOT SUPPORTED)\n")
    }

    // Future capabilities
    // if manager.SupportsUpgrade() {
    //     help.WriteString("  - Upgrade packages\n")
    // }

    return help.String()
}
```

## Benefits of This Approach

1. **Better User Experience**: Users know upfront what operations are supported
2. **Cleaner Error Handling**: No need to try-and-fail to discover capabilities
3. **UI Adaptation**: UIs can disable/hide unsupported operations
4. **Help Generation**: Documentation can be dynamic based on capabilities
5. **Future-Proof**: Easy to add new optional operations

## Migration Guide

### Before (without capability checking):
```go
results, err := manager.Search(ctx, query)
if err != nil {
    // User sees error after trying
    return err
}
```

### After (with capability checking):
```go
if !manager.SupportsSearch() {
    // User knows upfront it's not supported
    // Can get helpful suggestion from Search() error
    _, err := manager.Search(ctx, query)
    return err
}

results, err := manager.Search(ctx, query)
if err != nil {
    // This is a real error, not just "unsupported"
    return err
}
```

## Testing Capabilities

```go
func TestCommandWithCapabilities(t *testing.T) {
    // Create mock manager
    mockManager := &MockPackageManager{}

    // Test with search supported
    mockManager.On("SupportsSearch").Return(true)
    mockManager.On("Search", ctx, "query").Return([]string{"result"}, nil)

    err := runSearchCommand(mockManager, "query")
    assert.NoError(t, err)

    // Test with search not supported
    mockManager2 := &MockPackageManager{}
    mockManager2.On("SupportsSearch").Return(false)
    mockManager2.On("Search", ctx, "query").
        Return(nil, errors.NewError(errors.ErrOperationNotSupported, ...))

    err = runSearchCommand(mockManager2, "query")
    assert.Error(t, err)
    assert.Equal(t, errors.ErrOperationNotSupported, err.(*errors.PlonkError).Code)
}
```
