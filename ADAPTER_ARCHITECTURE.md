# Adapter Architecture Guidelines

**Created**: 2025-07-13
**Purpose**: Document the adapter pattern usage in plonk architecture

## Overview

Adapters in plonk serve as bridges between package boundaries, preventing circular dependencies while maintaining clean separation of concerns. This document outlines when and how to use adapters effectively.

## Why Adapters?

### The Problem
Go's import system doesn't allow circular dependencies. When package A needs to use types from package B, but package B also needs types from package A, we hit a circular dependency error.

### The Solution
Adapters act as translators between packages, allowing each package to define its own interfaces without importing from other packages.

```
Package A → Adapter → Package B
         ↖          ↙
          Interface
```

## Current Adapters

### 1. Config Adapters (`/internal/config/adapters.go`)
- **StatePackageConfigAdapter**: Bridges config → state for package configs
- **StateDotfileConfigAdapter**: Bridges config → state for dotfile configs

### 2. State Adapters (`/internal/state/adapters.go`)
- **ConfigAdapter**: Adapts config types to state interfaces
- **ManagerAdapter**: Adapts package managers to state interface

### 3. Lock File Adapter (`/internal/lock/adapter.go`)
- **LockFileAdapter**: Bridges lock file service to state interfaces

## When to Use Adapters

### Use Adapters When:
1. **Cross-Package Communication**: When package A needs to use functionality from package B
2. **Circular Dependency Prevention**: When direct imports would create cycles
3. **Interface Translation**: When packages have similar but not identical interfaces
4. **Type Conversion**: When converting between package-specific types

### Use Type Aliases When:
1. **Identical Interfaces**: When interfaces match exactly
2. **Same Package Boundary**: When consolidating within a package
3. **No Circular Risk**: When imports are unidirectional

## Adapter Implementation Pattern

### Basic Structure
```go
// adapter.go
type SourceTargetAdapter struct {
    source SourceInterface
}

func NewSourceTargetAdapter(source SourceInterface) *SourceTargetAdapter {
    return &SourceTargetAdapter{source: source}
}

// Implement target interface methods
func (a *SourceTargetAdapter) TargetMethod() error {
    // Translate and delegate to source
    return a.source.SourceMethod()
}
```

### Naming Convention
- **Pattern**: `<Source><Target>Adapter`
- **Examples**:
  - `StatePackageConfigAdapter` (state → package config)
  - `ConfigAdapter` (config → generic interface)
  - `LockFileAdapter` (lock file → state interface)

### Interface Compliance
Always add compile-time checks:
```go
var _ TargetInterface = (*SourceTargetAdapter)(nil)
```

## Best Practices

### 1. Keep Adapters Thin
Adapters should only translate between interfaces, not contain business logic.

```go
// Good: Simple translation
func (a *Adapter) GetItems() []TargetItem {
    sourceItems := a.source.GetSourceItems()
    return convertItems(sourceItems)
}

// Bad: Business logic in adapter
func (a *Adapter) GetItems() []TargetItem {
    items := a.source.GetSourceItems()
    // Don't filter or process here
    filtered := filterInvalidItems(items)
    return filtered
}
```

### 2. Document Purpose
Always document why an adapter exists:
```go
// StatePackageConfigAdapter bridges the config package's ConfigAdapter
// to the state package's PackageConfigLoader interface, preventing
// circular dependencies between these packages.
type StatePackageConfigAdapter struct {
    configAdapter *ConfigAdapter
}
```

### 3. Test Adapters
Ensure adapters correctly translate between interfaces:
```go
func TestAdapterTranslation(t *testing.T) {
    source := &mockSource{items: []string{"a", "b"}}
    adapter := NewAdapter(source)

    result := adapter.GetItems()
    assert.Equal(t, 2, len(result))
}
```

### 4. Performance Considerations
- Adapters add a layer of indirection
- Usually negligible performance impact
- Profile if performance is critical
- Consider caching for expensive translations

## Migration Strategy

When consolidating interfaces:

1. **Identify Duplicates**: Find interfaces that serve the same purpose
2. **Check Signatures**: Determine if they're identical or need adaptation
3. **Apply Strategy**:
   - Identical → Use type alias
   - Different → Keep adapter
   - Complex → Document translation logic

## Examples

### Type Alias (Simple Case)
```go
// When interfaces are identical
package state

import "github.com/richhaase/plonk/internal/interfaces"

// Direct alias - no adapter needed
type DotfileConfigLoader = interfaces.DotfileConfigLoader
```

### Adapter (Complex Case)
```go
// When interfaces differ
package config

type StatePackageConfigAdapter struct {
    configAdapter *ConfigAdapter
}

func (s *StatePackageConfigAdapter) GetPackagesForManager(name string) ([]state.PackageConfigItem, error) {
    // Convert from config.PackageConfigItem to state.PackageConfigItem
    items, err := s.configAdapter.GetPackagesForManager(name)
    if err != nil {
        return nil, err
    }

    // Type conversion
    stateItems := make([]state.PackageConfigItem, len(items))
    for i, item := range items {
        stateItems[i] = state.PackageConfigItem{Name: item.Name}
    }
    return stateItems, nil
}
```

## Future Considerations

### Generics
With Go generics, we might create generic adapters:
```go
type Adapter[S any, T any] struct {
    source S
    converter func(S) T
}
```

### Code Generation
For many similar adapters, consider code generation to reduce boilerplate.

## Conclusion

Adapters are a fundamental part of plonk's architecture, not technical debt. They enable:
- Clean package boundaries
- Circular dependency prevention
- Interface evolution
- Type safety

When used correctly, adapters make the codebase more maintainable and flexible.
