# Task 017: Extract Business Logic from Commands Package

## Objective
Refactor the commands package to be thin CLI handlers by extracting business logic to appropriate domain packages, achieving 20-30% code reduction while improving separation of concerns.

## Quick Context
- **Current state**: 5,076 LOC (34.8% of entire codebase)
- **Target reduction**: 20-30% (5,076 → ~3,500-4,000 LOC)
- **Core principle**: Commands should only handle CLI concerns (parsing, validation, output)
- **Business logic**: Should live in domain packages (dotfiles, managers, orchestrator)

## Current Problems

### 1. Commands Contain Business Logic
Commands currently handle:
- File operations and dotfile management
- Package installation orchestration
- State reconciliation logic
- Complex validation rules
- Output formatting decisions

### 2. Poor Separation of Concerns
```go
// Example: add.go contains dotfile business logic that belongs in dotfiles package
func runAdd(cmd *cobra.Command, args []string) error {
    // CLI parsing (appropriate)
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // Business logic (should be extracted)
    for _, path := range args {
        // Complex file validation
        // Dotfile copying logic
        // Lock file updates
        // Error handling specific to dotfile operations
    }
}
```

### 3. Duplicated Logic Across Commands
Similar patterns repeated in:
- install.go / uninstall.go (package management)
- add.go / rm.go (dotfile management)
- sync.go / doctor.go (state validation)

## Refactoring Strategy

### Phase 1: Extract Dotfile Business Logic
**Target files**: add.go, rm.go, dotfiles.go, dotfile_operations.go

**Move to dotfiles package**:
- File validation and path resolution
- Dotfile copying and linking operations
- Backup handling
- Lock file updates for dotfiles

**Commands become**:
```go
func runAdd(cmd *cobra.Command, args []string) error {
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // Just call domain package
    results, err := dotfiles.AddFiles(args, dotfiles.AddOptions{
        DryRun: dryRun,
        Backup: true,
    })
    if err != nil {
        return err
    }

    // Just format output
    return ui.DisplayResults(results, cmd.OutOrStdout())
}
```

### Phase 2: Extract Package Management Logic
**Target files**: install.go, uninstall.go, sync.go

**Move to managers package**:
- Package installation orchestration
- Multi-manager coordination
- Dependency resolution
- Lock file updates for packages

**Commands become**:
```go
func runInstall(cmd *cobra.Command, args []string) error {
    // Parse flags only
    managerName, _ := cmd.Flags().GetString("manager")
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // Call domain package
    results, err := managers.InstallPackages(args, managers.InstallOptions{
        Manager: managerName,
        DryRun:  dryRun,
    })
    if err != nil {
        return err
    }

    // Format output
    return ui.DisplayResults(results, cmd.OutOrStdout())
}
```

### Phase 3: Extract Orchestration Logic
**Target files**: sync.go, doctor.go, status.go

**Move to orchestrator package**:
- State reconciliation coordination
- System health checks
- Cross-domain status aggregation
- Sync planning and execution

### Phase 4: Simplify Output Logic
**Target files**: output.go, output_utils.go, helpers.go

**Consolidate with ui package**:
- Move formatting logic to ui package
- Reduce duplication between structured and human output
- Simplify command output responsibilities

## Specific Extraction Targets

### High-Impact Extractions (>200 LOC each)

1. **dotfile_operations.go** (343 LOC)
   - Extract to `dotfiles.Operations()`
   - Keep only CLI interfacing in commands

2. **doctor.go** (771 LOC)
   - Extract system checks to `orchestrator.RunHealthChecks()`
   - Extract diagnostics to `orchestrator.GenerateDiagnostics()`
   - Keep only output formatting in commands

3. **sync.go** (471 LOC)
   - Extract sync logic to `orchestrator.PlanSync()` and `orchestrator.ExecuteSync()`
   - Keep only progress reporting in commands

4. **install.go/uninstall.go** (282 + 318 = 600 LOC)
   - Extract package operations to `managers.InstallPackages()` / `managers.UninstallPackages()`
   - Consolidate common patterns

### Medium-Impact Extractions (100-200 LOC each)

5. **shared.go** (396 LOC)
   - Move shared business logic to appropriate domains
   - Keep only CLI utilities in commands

6. **output.go** (436 LOC)
   - Move formatting logic to ui package
   - Simplify command output responsibilities

7. **search.go/info.go** (352 + 349 = 701 LOC)
   - Extract search/info logic to managers package
   - Simplify result formatting

## Implementation Guidelines

### Extract Pattern
```go
// Before: Business logic in command
func runCommand(cmd *cobra.Command, args []string) error {
    // 50+ lines of business logic
    // Complex domain operations
    // Output formatting mixed in
}

// After: Thin command handler
func runCommand(cmd *cobra.Command, args []string) error {
    // 1. Parse CLI inputs
    opts := parseCommandOptions(cmd, args)

    // 2. Call domain package
    results, err := domain.DoOperation(opts)
    if err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }

    // 3. Format output
    return ui.Display(results, opts.Format, cmd.OutOrStdout())
}
```

### Domain Package APIs
Design clean, focused APIs:
```go
// dotfiles package
type AddOptions struct {
    DryRun bool
    Backup bool
    Force  bool
}

func AddFiles(paths []string, opts AddOptions) ([]OperationResult, error)
func RemoveFiles(paths []string, opts RemoveOptions) ([]OperationResult, error)

// managers package
type InstallOptions struct {
    Manager string
    DryRun  bool
    Global  bool
}

func InstallPackages(packages []string, opts InstallOptions) ([]OperationResult, error)
```

## Success Criteria
1. ✅ **20-30% code reduction** achieved (5,076 → ~3,500-4,000 LOC)
2. ✅ **Commands are thin handlers** - mostly CLI parsing and output formatting
3. ✅ **Business logic in domain packages** - dotfiles, managers, orchestrator
4. ✅ **No functionality lost** - All CLI commands work identically
5. ✅ **Better testability** - Domain logic easier to unit test
6. ✅ **Reduced duplication** - Common patterns consolidated
7. ✅ **Clean APIs** - Domain packages have focused, well-designed interfaces

## Risk Mitigation
- **Extract one command at a time**: Don't refactor everything simultaneously
- **Preserve CLI interface**: User-facing behavior must remain identical
- **Test after each extraction**: Run integration tests continuously
- **Start with isolated commands**: Begin with commands that don't share logic

## Testing Strategy
1. **Before each extraction**: Document current command behavior
2. **After each extraction**: Verify CLI interface unchanged
3. **Domain package testing**: Add focused unit tests for extracted logic
4. **Integration testing**: Ensure `just test-ux` continues to pass

## Completion Report
Create `TASK_017_COMPLETION_REPORT.md` with:
- LOC reduction metrics per command
- List of business logic moved to each domain package
- New domain package APIs created
- Verification that all CLI commands work identically
- Examples of before/after command implementations
