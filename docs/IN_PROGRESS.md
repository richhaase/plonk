# Plonk Code Review & Improvement Plan

## **Strengths: Clean Architecture & Separation of Concerns**

The code demonstrates excellent separation into the 4 core buckets:

1. **Configuration** (`internal/config/`) - Clean YAML parsing with validation
2. **Package Management** (`internal/managers/`) - Pluggable interface design
3. **Dotfile Management** (`internal/state/dotfile_provider.go`) - Well-abstracted file operations
4. **State Management** (`internal/state/`) - Unified reconciliation pattern

## **Key Architectural Improvements Needed**

### 1. **Configuration Loading - Interface Segregation**
**Current Issue**: `yaml_config.go:241` - Config struct directly implements provider interfaces
```go
// GetDotfileTargets returns dotfiles with their target paths.
func (c *Config) GetDotfileTargets() map[string]string {
```

**Improvement**: Create separate config reader/writer interfaces:
```go
type ConfigReader interface {
    LoadConfig(path string) (*Config, error)
}

type ConfigWriter interface {
    SaveConfig(path string, config *Config) error
}

type DotfileConfigReader interface {
    GetDotfileTargets() map[string]string
}
```

### 2. **Package Manager Interface - Error Handling**
**Current Issue**: `managers/homebrew.go:62` - Inconsistent error handling patterns
```go
func (h *HomebrewManager) IsInstalled(name string) bool {
    cmd := exec.Command("brew", "list", name)
    err := cmd.Run()
    return err == nil  // Loses error context
}
```

**Improvement**: Use proper error propagation:
```go
func (h *HomebrewManager) IsInstalled(name string) (bool, error) {
    cmd := exec.Command("brew", "list", name)
    if err := cmd.Run(); err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            return false, nil // Package not installed
        }
        return false, fmt.Errorf("failed to check package %s: %w", name, err)
    }
    return true, nil
}
```

### 3. **State Provider Pattern - Generic Implementation**
**Current Issue**: Code duplication between `package_provider.go` and `dotfile_provider.go`

**Improvement**: Extract common provider logic:
```go
type BaseProvider[T ConfigItem, U ActualItem] struct {
    domain string
    configLoader ConfigLoader[T]
    actualLoader ActualLoader[U]
}

func (b *BaseProvider[T, U]) Domain() string {
    return b.domain
}

func (b *BaseProvider[T, U]) GetConfiguredItems() ([]ConfigItem, error) {
    items, err := b.configLoader.LoadConfigured()
    if err != nil {
        return nil, fmt.Errorf("failed to load configured items: %w", err)
    }
    return items, nil
}
```

### 4. **Error Handling - Consistent Patterns**
**Current Issue**: Mixed error handling approaches throughout codebase

**Improvement**: Standardize error types:
```go
type PlonkError struct {
    Op      string // Operation
    Domain  string // package, dotfile, etc.
    Item    string // specific item name
    Err     error  // underlying error
}

func (e *PlonkError) Error() string {
    return fmt.Sprintf("plonk %s %s [%s]: %v", e.Op, e.Domain, e.Item, e.Err)
}
```

### 5. **Individual Item Focus - Command Structure**
**Current Issue**: Commands handle multiple items but core abstractions are per-item

**Improvement**: Enforce single-item operations in core:
```go
type ItemManager interface {
    Add(item string) error
    Remove(item string) error
    Apply(item string) error
    Status(item string) (ItemState, error)
}
```

## **Go Idioms & Best Practices**

### 1. **Context Propagation**
Add context support throughout:
```go
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
    cmd := exec.CommandContext(ctx, "brew", "install", name)
    // ...
}
```

### 2. **Functional Options Pattern**
For provider configuration:
```go
type ProviderOption func(*DotfileProvider)

func WithBackup(enabled bool) ProviderOption {
    return func(p *DotfileProvider) {
        p.backupEnabled = enabled
    }
}

func NewDotfileProvider(opts ...ProviderOption) *DotfileProvider {
    p := &DotfileProvider{}
    for _, opt := range opts {
        opt(p)
    }
    return p
}
```

### 3. **Proper Interface Definitions**
Move interfaces to separate files:
```go
// internal/interfaces/providers.go
type Provider interface {
    Domain() string
    GetConfiguredItems() ([]ConfigItem, error)
    GetActualItems() ([]ActualItem, error)
    CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
}
```

## **Specific Improvements**

### 1. **Config Path Resolution**
**File**: `config/yaml_config.go:289`
```go
func targetToSource(target string) string {
    // Remove ~/ prefix if present
    if len(target) > 2 && target[:2] == "~/" {
        target = target[2:]
    }
    // Should use strings.TrimPrefix for clarity
    target = strings.TrimPrefix(target, "~/")
    target = strings.TrimPrefix(target, ".")
    // ...
}
```

### 2. **Directory Expansion Logic**
**File**: `state/dotfile_provider.go:177`
The `expandConfigDirectory` function is solid but could benefit from:
- Early returns for empty directories
- Better error context
- Consistent path handling utilities

### 3. **State Reconciliation**
**File**: `state/reconciler.go:96`
The reconciliation logic is excellent but could use:
- Concurrent provider reconciliation
- Better error aggregation
- Metrics collection

## **Recommendations for Next Steps**

1. **Refactor configuration loading** to use proper interfaces
2. **Add context support** throughout the codebase
3. **Implement proper error types** for better debugging
4. **Extract common provider logic** into generic base types
5. **Add comprehensive logging** with structured output
6. **Implement metrics collection** for operations

The core architecture is sound with excellent separation of concerns. The main improvements focus on Go idioms, error handling consistency, and reducing code duplication while maintaining the strong individual item focus.