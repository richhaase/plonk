# Plonk Development Guidelines for AI Agents

This document provides essential guidelines and context for AI coding agents working on the plonk codebase.

## Development Rules

### Error Handling
- **Always use plonk's error handling** - `errors.Wrap()` not `fmt.Errorf()`
- **Package manager availability** - Return `(false, nil)` not error when binary missing
- **Structured errors** - Use appropriate error codes and domains for user-friendly messages
- **Continue-on-failure** - Process all items even if some fail, provide comprehensive summaries
- **Contextual suggestions** - Include actionable error messages with suggested commands

### Testing & Quality
- **Test before release** - Run `just precommit` before any release
- **Graceful degradation** - Features should fail gracefully (e.g., unavailable package managers)
- **Context everywhere** - All long-running operations must accept context for cancellation
- **Table-driven tests** - Standard pattern throughout codebase
- **Mock interfaces** - All external dependencies must be mockable
- **Isolated environments** - Use `t.TempDir()` and environment variables for test isolation

### Developer Experience
- **Zero-config first** - Features must work without configuration
- **One-command setup** - New developers should be productive with `just dev-setup`
- **Fast feedback loops** - Use pre-commit framework for file-specific checks
- **Sequential processing** - Prefer sequential over parallel for user feedback and conflict avoidance

### Build & Release
- **Binary location** - Output to `bin/` not `build/`
- **Automated releases** - Use GoReleaser via `just release-auto`
- **Composite actions** - Reuse GitHub Actions for consistency

## Code Architecture Patterns

### Command Structure
```go
// Standard pattern for all commands
func runCommand(cmd *cobra.Command, args []string) error {
    // 1. Parse flags and validate input
    // 2. Create operation context with timeout
    // 3. Process items sequentially with progress reporting
    // 4. Handle output based on format (table vs structured)
    // 5. Determine exit code based on results
}
```

### Error Handling Pattern
```go
// Use structured errors with domain context
return errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "add",
    "dotfile does not exist").WithSuggestionMessage("Check if path exists: ls -la " + path)

// Wrap existing errors with context
return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy",
    "failed to copy dotfile")
```

### Shared Operations Pattern
```go
// Use shared OperationResult type for batch operations
type OperationResult struct {
    Name           string
    Status         string // "added", "updated", "failed", "would-add"
    Error          error
    FilesProcessed int
    Metadata       map[string]interface{}
}

// Sequential processing with progress reporting
reporter := operations.NewProgressReporter("operation", showProgress)
for _, item := range items {
    result := processItem(ctx, item)
    reporter.ShowItemProgress(result)
    results = append(results, result)
}
```

## Key Architectural Concepts

### State Reconciliation
Plonk compares desired state (configuration) with actual state (system):
- **Packages**: Lock file defines desired state, system inspection shows actual state
- **Dotfiles**: Config directory defines desired state, home directory shows actual state
- **Apply operations**: Reconcile differences between desired and actual state

### Filesystem-Based Dotfile Detection
- Uses `filepath.Walk()` to scan `$PLONK_DIR` (default: `~/.config/plonk`)
- Auto-discovery without manual configuration required
- Convention-based mapping: `zshrc` → `~/.zshrc`, `config/nvim/init.lua` → `~/.config/nvim/init.lua`
- Respects configurable ignore patterns

### Provider Pattern
Extensible architecture through well-defined interfaces:
- Package managers implement `PackageManager` interface from `internal/interfaces/`
- Commands use dependency injection for testability
- State reconciliation through provider abstractions

### Shared Runtime Context (Phase 4)
Use the singleton pattern for expensive resources:
- `runtime.GetSharedContext()` provides cached access to ManagerRegistry, Reconciler, Config
- Manager availability cached for 5 minutes to avoid repeated checks
- Configuration caching with `ConfigWithDefaults()` fallback pattern

### Logging System (Phase 4)
Industry-standard logging levels with domain-specific control:
- Use `runtime.Error/Warn/Info/Debug/Trace(domain, format, args...)`
- Domains: `DomainCommand`, `DomainConfig`, `DomainManager`, `DomainState`, `DomainFile`, `DomainLock`
- Environment: `PLONK_DEBUG=debug:manager,state` for targeted debugging

## Essential Commands

### Development Workflow
```bash
# Complete setup for new developers
just dev-setup

# Daily development cycle
just build && just test && just precommit

# After interface changes
just generate-mocks

# Before any release
just precommit
```

### Testing Commands
```bash
# Run all tests
just test

# Run tests with coverage
just test-coverage

# Test specific package
go test -v ./internal/commands/

# Test with environment override
PLONK_DIR=/tmp/test-config go test -v ./...
```

## Directory Structure

```
cmd/plonk/              # CLI entry point
internal/
├── commands/           # CLI command implementations using CommandPipeline
├── config/             # Configuration management (YAML-based)
├── dotfiles/           # Dotfile operations and atomic updates
├── errors/             # Structured error handling system
├── interfaces/         # Unified interface definitions (Phase 4)
├── managers/           # Package manager implementations
├── operations/         # Shared utilities for batch operations
├── runtime/            # Shared context and logging (Phase 4)
├── state/              # State reconciliation engine
└── testing/            # Test helpers and utilities (Phase 4)
docs/                   # Technical documentation
```

## Testing Infrastructure

### Test Patterns
- **Table-driven tests**: Comprehensive scenario coverage
- **Mock usage**: Isolated testing of business logic from `internal/interfaces/mocks/`
- **Test helpers**: Use `testing.NewTestContext(t)` for isolated environments (Phase 4)
- **Context testing**: Timeout and cancellation scenarios
- **Error scenarios**: Validate error codes and user messages
- **Environment isolation**: Use `PLONK_DIR` for test separation

### Key Test Files
- `internal/commands/*_test.go`: Command implementations
- `internal/operations/types_test.go`: Shared utilities
- `internal/managers/*_test.go`: Package manager implementations
- `internal/config/*_test.go`: Configuration parsing and validation

## Documentation Standards

### Command Documentation
- Update `docs/CLI.md` with comprehensive examples
- Include both single and multiple operation examples
- Show realistic output examples
- Document error scenarios and troubleshooting

### Code Documentation
- Interface documentation with usage examples
- Error scenarios and expected behavior
- Performance characteristics for operations
- Context requirements and timeout behavior

## Common Development Tasks

### Adding New Commands
1. Create in `internal/commands/` following existing patterns
2. Implement comprehensive error handling with suggestions
3. Support structured output formats (table, JSON, YAML)
4. Add table-driven tests with mocks and error scenarios
5. Update CLI.md with examples and usage patterns

### Extending Package Managers
1. Implement `PackageManager` interface
2. Add version detection via `GetInstalledVersion()`
3. Handle unavailable manager gracefully
4. Create comprehensive tests with mocks
5. Update configuration validation

### Working with Mocks
- Generate mocks after interface changes: `just generate-mocks`
- Use mocks for external dependencies (package managers, file system)
- Test both success and failure scenarios
- Verify correct interface usage in tests

## References

- **Session History**: See `DEVELOPMENT_HISTORY.md` for detailed development sessions
- **Architecture**: `docs/ARCHITECTURE.md` for technical deep-dive
- **CLI Reference**: `docs/CLI.md` for complete command documentation
- **Development Setup**: `docs/DEVELOPMENT.md` for contributing guidelines

## Code Review Memories

### Updated Code Review (2024-03-02)
- Perform a critical code review of the plonk project
- Project is very young (only a week old) with a largely fixed interface as of yesterday
- Multiple refactors to enhance UX have occurred, with code potentially renamed multiple times
- Do not assume the architecture is correct
- Look for opportunities to extract common functionality
- Primary goal: Determine necessary code refinements to ensure well-structured, clearly designed code with clear responsibility boundaries
