# Critical Architecture Review: Plonk Codebase

## Executive Summary

The plonk codebase has become unmaintainable due to over-engineering and excessive abstraction. What should be a simple tool for managing dotfiles and packages has evolved into a complex system with 50+ packages, 100+ interfaces, and 7-10 layers of abstraction for basic operations. This complexity makes it extremely difficult for both AI agents and human developers to understand, modify, or debug the code effectively.

## Core Architectural Failures

### 1. Adapter Pattern Abuse

The codebase contains 15+ adapter types that exist solely to work around circular dependencies:

- `StatePackageConfigAdapter`
- `StateDotfileConfigAdapter`
- `ConfigAdapter`
- `ManagerAdapter`
- `LockFileAdapter`

These adapters are symptoms of poor architectural design. Well-structured code doesn't need adapters at every boundary.

### 2. Business Logic Scatter

Simple operations are distributed across many files. For example, adding a dotfile touches:

1. `commands/add.go` - CLI parsing
2. `commands/pipeline.go` - Pipeline abstraction
3. `commands/shared.go` - Mixed utilities (1185 lines!)
4. `services/dotfile_operations.go` - Service layer wrapper
5. `dotfiles/operations.go` - Core logic
6. `state/dotfile_provider.go` - State management
7. `paths/resolver.go` - Path resolution
8. Multiple additional files

**Result**: 9+ files involved in copying a single file.

### 3. Duplicate Implementations

Core functionality is implemented multiple times with slight variations. Path resolution alone has 4 different implementations:

- `commands/shared.go:resolveDotfilePath()`
- `services/dotfile_operations.go:ResolveDotfilePath()`
- `dotfiles/operations.go:ExpandPath()`
- `paths/resolver.go:ResolveDotfilePath()`

### 4. Interface Explosion

The codebase defines interfaces for everything, most with only one implementation:

- `ConfigReader`, `ConfigWriter`, `ConfigValidator`
- `DotfileConfigReader`, `PackageConfigReader`
- `Provider`, `BatchProcessor`, `ProgressReporter`

This violates YAGNI (You Aren't Gonna Need It) and makes code navigation difficult.

### 5. Pipeline Anti-Pattern

The command pipeline adds unnecessary layers:

```
runAdd() → NewSimpleCommandPipeline() → ExecuteWithData() →
addDotfilesProcessor() → addSingleDotfiles() → addSingleDotfile() →
AddSingleDotfile() → AddSingleFile() → CopyFileWithAttributes()
```

Each layer adds error handling, type conversion, and complexity without clear value.

### 6. Configuration Over-Engineering

The configuration system spans 1500+ lines for what should be simple YAML handling:

- `Config` struct (all nullable pointers)
- `ResolvedConfig` struct
- `ConfigDefaults` struct
- Multiple loading methods
- Manual schema generation
- Validation at multiple levels

### 7. State Reconciliation Complexity

The state reconciliation system is over-engineered with:

- Generic `Provider` interface
- Complex diffing logic
- Multiple abstraction layers

What should be simple set operations (configured - actual = to_install) spans multiple files and interfaces.

## Why AI Agents Struggle

1. **No Clear Entry Points** - Understanding any feature requires reading 10+ files
2. **Hidden Dependencies** - Adapters obscure real relationships between components
3. **Scattered Context** - Related logic is distributed across packages
4. **Naming Confusion** - Same concepts have multiple names and implementations
5. **Abstraction Maze** - Cannot determine what code actually does without tracing through layers

## Impact on Maintainability

### Cognitive Load
Understanding any operation requires holding 7-10 files in mental context simultaneously.

### Change Amplification
Simple changes cascade through multiple files due to tight coupling disguised as loose coupling.

### Bug Hiding
Bugs can hide in the gaps between layers, making debugging extremely difficult.

### Testing Complexity
Tests require extensive mocking due to the layered architecture.

### Onboarding Difficulty
New developers must understand the entire architecture to make even simple changes.

## Concrete Examples

### Example 1: Adding a Dotfile

A simple `plonk add ~/.zshrc` operation involves:

1. Command parsing (`commands/add.go`)
2. Pipeline setup (`commands/pipeline.go`)
3. Processor function (`commands/add.go`)
4. Path resolution (multiple implementations)
5. Validation (scattered across files)
6. File operations (`commands/shared.go`)
7. Result formatting (`commands/output.go`)

### Example 2: Package Installation

Installing a package touches:

1. Command entry (`commands/install.go`)
2. Batch processing framework (`operations/batch.go`)
3. Manager registry (`managers/registry.go`)
4. Specific manager implementation
5. Lock file updates (`lock/yaml_lock.go`)
6. State reconciliation (if used)
7. Multiple error handling layers

## What Good Architecture Would Look Like

A maintainable version would have:

- **3-4 core packages** instead of 50+
- **Direct function calls** instead of 9 layers of indirection
- **Clear domain boundaries**
- **Business logic consolidated** in domain objects
- **No adapter layers**
- **Minimal, purposeful interfaces**

Example structure:
```
cmd/plonk/main.go
internal/
  core/           # Core business logic
    dotfiles.go   # All dotfile operations
    packages.go   # All package operations
    config.go     # Configuration
  managers/       # Package manager implementations
  cli/           # CLI command definitions
```

## Recommendations for Improvement

1. **Consolidate Business Logic** - Move related operations into single domain files
2. **Remove Adapters** - Restructure to eliminate circular dependencies
3. **Reduce Interfaces** - Only create interfaces when there are multiple implementations
4. **Flatten Call Chains** - Remove unnecessary abstraction layers
5. **Simplify Configuration** - Use struct tags and standard libraries
6. **Direct Operations** - Replace complex pipelines with straightforward function calls

## Conclusion

The plonk codebase is a cautionary tale of over-engineering. It prioritizes theoretical "clean architecture" principles over practical maintainability. For a tool that manages dotfiles and packages, this level of complexity is unjustifiable and actively harmful to development velocity and code quality.

The architecture must be simplified dramatically to make the codebase maintainable and accessible to both human developers and AI agents.
