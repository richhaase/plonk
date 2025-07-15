# Adding a New Package Manager to Plonk

This guide walks through the process of adding a new package manager to plonk. With the refactored architecture, this should take 1-2 hours.

## Prerequisites

Before starting, ensure you have:
- Basic understanding of the package manager's CLI interface
- Test environment with the package manager installed (for integration testing)
- Go development environment set up

## Step-by-Step Guide

### 1. Choose a Similar Manager as Template

Look at existing managers to find one with similar behavior:
- **Simple managers** (list/install/uninstall): Use `gem` or `cargo` as template
- **JSON output support**: Use `npm` or `pip` as template
- **System package managers**: Use `apt` or `homebrew` as template
- **Non-standard patterns**: Use `goinstall` as template

### 2. Create the Manager File

Create `internal/managers/yourmanager.go`:

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
    "context"
    "strings"

    "github.com/richhaase/plonk/internal/errors"
    "github.com/richhaase/plonk/internal/executor"
)

// YourManager manages packages via your-tool using BaseManager for common functionality.
type YourManager struct {
    *BaseManager
}

// NewYourManager creates a new manager with the default executor.
func NewYourManager() *YourManager {
    return newYourManager(nil)
}

// NewYourManagerWithExecutor creates a new manager with a custom executor for testing.
func NewYourManagerWithExecutor(exec executor.CommandExecutor) *YourManager {
    return newYourManager(exec)
}

// newYourManager creates a manager with the given executor.
func newYourManager(exec executor.CommandExecutor) *YourManager {
    config := ManagerConfig{
        BinaryName:  "your-tool",
        VersionArgs: []string{"--version"},
        ListArgs: func() []string {
            return []string{"list"}
        },
        InstallArgs: func(pkg string) []string {
            return []string{"install", pkg}
        },
        UninstallArgs: func(pkg string) []string {
            return []string{"uninstall", pkg}
        },
    }

    // Add tool-specific error patterns
    errorMatcher := NewCommonErrorMatcher()
    errorMatcher.AddPattern(ErrorTypeNotFound, "not found", "no such package")
    errorMatcher.AddPattern(ErrorTypeAlreadyInstalled, "already installed")
    errorMatcher.AddPattern(ErrorTypeNotInstalled, "not installed")

    var base *BaseManager
    if exec == nil {
        base = NewBaseManager(config)
    } else {
        base = NewBaseManagerWithExecutor(config, exec)
    }
    base.ErrorMatcher = errorMatcher

    return &YourManager{
        BaseManager: base,
    }
}
```

### 3. Implement Required Methods

At minimum, implement these methods:

```go
// ListInstalled lists all installed packages.
func (y *YourManager) ListInstalled(ctx context.Context) ([]string, error) {
    output, err := y.ExecuteList(ctx)
    if err != nil {
        return nil, err
    }

    return y.parseListOutput(output), nil
}

// parseListOutput parses the list command output
func (y *YourManager) parseListOutput(output []byte) []string {
    result := strings.TrimSpace(string(output))
    if result == "" {
        return []string{}
    }

    // Parse according to your tool's output format
    var packages []string
    lines := strings.Split(result, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            packages = append(packages, line)
        }
    }

    return packages
}

// Install installs a package.
func (y *YourManager) Install(ctx context.Context, name string) error {
    return y.ExecuteInstall(ctx, name)
}

// Uninstall removes a package.
func (y *YourManager) Uninstall(ctx context.Context, name string) error {
    return y.ExecuteUninstall(ctx, name)
}

// IsInstalled checks if a package is installed.
func (y *YourManager) IsInstalled(ctx context.Context, name string) (bool, error) {
    packages, err := y.ListInstalled(ctx)
    if err != nil {
        return false, err
    }

    for _, pkg := range packages {
        if pkg == name {
            return true, nil
        }
    }

    return false, nil
}
```

### 4. Implement Search (if supported)

If your package manager supports search:

```go
// Search searches for packages.
func (y *YourManager) Search(ctx context.Context, query string) ([]string, error) {
    output, err := y.Executor.Execute(ctx, y.GetBinary(), "search", query)
    if err != nil {
        return nil, errors.WrapWithItem(err, errors.ErrCommandExecution,
            errors.DomainPackages, "search", query,
            "failed to search packages")
    }

    return y.parseSearchOutput(output), nil
}
```

If search is NOT supported:

```go
// SupportsSearch returns false as this tool doesn't have search.
func (y *YourManager) SupportsSearch() bool {
    return false
}

// Search returns an error as this operation is not supported.
func (y *YourManager) Search(ctx context.Context, query string) ([]string, error) {
    return nil, errors.NewError(errors.ErrOperationNotSupported,
        errors.DomainPackages, "search",
        "your-tool does not have a search command").
        WithSuggestionMessage("Visit https://your-tool.com to search for packages")
}
```

### 5. Implement Info and GetInstalledVersion

```go
// Info retrieves detailed information about a package.
func (y *YourManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
    // Implementation depends on your tool's capabilities
    // See existing managers for examples
}

// GetInstalledVersion retrieves the installed version of a package.
func (y *YourManager) GetInstalledVersion(ctx context.Context, name string) (string, error) {
    // Implementation depends on your tool's output format
    // See existing managers for examples
}
```

### 6. Add Custom Error Patterns

Update the error matcher in your constructor with tool-specific patterns:

```go
errorMatcher.AddPattern(ErrorTypeNotFound, "E404", "404 Not Found", "package not found")
errorMatcher.AddPattern(ErrorTypePermission, "permission denied", "access denied")
errorMatcher.AddPattern(ErrorTypeNetwork, "network error", "connection refused")
```

### 7. Create Unit Tests

Create `internal/managers/yourmanager_test.go`:

```go
package managers

import (
    "context"
    "testing"

    "github.com/golang/mock/gomock"
    "github.com/richhaase/plonk/internal/executor/mocks"
    "github.com/stretchr/testify/assert"
)

func TestYourManager_Install(t *testing.T) {
    tests := []struct {
        name        string
        packageName string
        setupMock   func(m *mocks.MockCommandExecutor)
        wantErr     bool
    }{
        {
            name:        "successful install",
            packageName: "example",
            setupMock: func(m *mocks.MockCommandExecutor) {
                m.EXPECT().
                    ExecuteCombined(gomock.Any(), "your-tool", "install", "example").
                    Return([]byte("Successfully installed example"), nil)
            },
            wantErr: false,
        },
        {
            name:        "package not found",
            packageName: "nonexistent",
            setupMock: func(m *mocks.MockCommandExecutor) {
                m.EXPECT().
                    ExecuteCombined(gomock.Any(), "your-tool", "install", "nonexistent").
                    Return([]byte("Error: package not found"), &executor.ExecError{Code: 1})
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockExec := mocks.NewMockCommandExecutor(ctrl)
            tt.setupMock(mockExec)

            manager := NewYourManagerWithExecutor(mockExec)
            err := manager.Install(context.Background(), tt.packageName)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Add tests for other methods...
```

### 8. Register the Manager

Add your manager to the registry in `internal/managers/registry.go`:

```go
func NewManagerRegistry() *ManagerRegistry {
    return &ManagerRegistry{
        managers: map[string]ManagerFactory{
            "apt":      func() PackageManager { return NewAptManager() },
            "homebrew": func() PackageManager { return NewHomebrewManager() },
            "npm":      func() PackageManager { return NewNpmManager() },
            "cargo":    func() PackageManager { return NewCargoManager() },
            "pip":      func() PackageManager { return NewPipManager() },
            "gem":      func() PackageManager { return NewGemManager() },
            "go":       func() PackageManager { return NewGoInstallManager() },
            "yourmanager": func() PackageManager { return NewYourManager() }, // ADD THIS
        },
    }
}
```

### 9. Update Command Flags

If your manager needs special flags, update the relevant commands in `internal/commands/`:
- `install.go` - Add flag like `--yourmanager`
- `uninstall.go` - Add corresponding flag
- `list.go` - Add to manager selection logic

### 10. Add Integration Tests

Create `internal/managers/yourmanager_integration_test.go`:

```go
//go:build integration

package managers

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestYourManagerIntegration(t *testing.T) {
    manager := NewYourManager()
    ctx := context.Background()

    // Check availability
    available, err := manager.IsAvailable(ctx)
    require.NoError(t, err)
    if !available {
        t.Skip("your-tool not available")
    }

    // Test list
    packages, err := manager.ListInstalled(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, packages)
}
```

## Common Patterns

### Handling JSON Output

If your tool supports JSON output:

```go
config := ManagerConfig{
    ListArgs: func() []string {
        return []string{"list", "--json"}
    },
    PreferJSON: true,
}

func (y *YourManager) parseListOutput(output []byte) ([]string, error) {
    var packages []struct {
        Name string `json:"name"`
    }
    if err := json.Unmarshal(output, &packages); err != nil {
        return nil, err
    }
    // Extract names...
}
```

### Platform-Specific Availability

For OS-specific tools:

```go
func (y *YourManager) IsAvailable(ctx context.Context) (bool, error) {
    if runtime.GOOS != "linux" {
        return false, nil
    }
    return y.BaseManager.IsAvailable(ctx)
}
```

### Custom Install Logic

If you need retry logic or special handling:

```go
func (y *YourManager) Install(ctx context.Context, name string) error {
    // Try with default args first
    err := y.ExecuteInstall(ctx, name)
    if err == nil {
        return nil
    }

    // Check error and potentially retry
    if isRetryableError(err) {
        // Try alternative approach
        output, err := y.Executor.Execute(ctx, y.GetBinary(), "install", "--force", name)
        // Handle result...
    }

    return err
}
```

## Testing Your Manager

1. **Run unit tests**: `go test ./internal/managers -run TestYourManager`
2. **Run with mocks**: Ensure all tests pass without the tool installed
3. **Integration test**: `go test ./internal/managers -tags integration -run TestYourManagerIntegration`
4. **Manual testing**: Use plonk commands to verify behavior

## Checklist

- [ ] Manager struct and constructors created
- [ ] BaseManager configuration set up correctly
- [ ] Error patterns added for common errors
- [ ] ListInstalled implemented with proper parsing
- [ ] Install/Uninstall use BaseManager or have custom logic
- [ ] Search implemented or SupportsSearch returns false
- [ ] Info and GetInstalledVersion implemented
- [ ] Unit tests cover all methods
- [ ] Integration tests verify real behavior
- [ ] Manager registered in registry
- [ ] Command flags updated if needed

## Tips

1. **Start simple**: Get basic list/install/uninstall working first
2. **Use existing managers**: Copy patterns from similar tools
3. **Test early**: Write tests as you implement each method
4. **Handle errors gracefully**: Add specific error patterns for better UX
5. **Document quirks**: Add comments for tool-specific behavior

## Common Pitfalls

1. **Exit codes**: Some tools use non-standard exit codes
2. **Output formats**: Tools may have different output for different versions
3. **Global vs user**: Consider where packages are installed
4. **Permissions**: Some operations may require elevated privileges
5. **Platform differences**: Test on all supported platforms

## Getting Help

- Review existing managers for patterns
- Check the test files for examples
- Look at BaseManager for available helper methods
- Consult PKG_MGR_REFACTOR.md for architecture details
