# Plonk Codebase Refactoring Implementation Plan

## Plan Summary and Progress Tracker

### Overall Goal
Systematically address code review findings through phases that eliminate duplication, improve consistency, and enhance maintainability while preserving all public APIs and build system compatibility.

### Phase Overview
- **Phase 1**: Foundation & Cleanup (Low Risk, High Impact) - **PLANNED**
- **Phase 2**: Core Abstractions (Medium Risk, High Impact) - **PLANNED**
- **Phase 3**: Structural Improvements (Medium Risk, Medium Impact) - **PLANNED**
- **Phase 4**: Advanced Optimizations (High Risk, High Impact) - **PLANNED**

### Progress Checklist

#### Phase 1: Foundation & Cleanup ⏸️ **NOT STARTED**
- [ ] **P1.1**: Remove backup files and legacy artifacts
- [ ] **P1.2**: Implement logging infrastructure
- [ ] **P1.3**: Standardize package manager error handling
- [ ] **P1.4**: Extract ManagerRegistry pattern
- [ ] **P1.5**: Validation and testing

#### Phase 2: Core Abstractions ⏸️ **NOT STARTED**
- [ ] **P2.1**: Create CommandPipeline abstraction
- [ ] **P2.2**: Extract PathResolver utility
- [ ] **P2.3**: Centralize configuration management
- [ ] **P2.4**: Standardize output rendering
- [ ] **P2.5**: Migration and validation

#### Phase 3: Structural Improvements ⏸️ **NOT STARTED**
- [ ] **P3.1**: Simplify dotfile provider complexity
- [ ] **P3.2**: Consolidate interface hierarchies
- [ ] **P3.3**: Enhance operations package
- [ ] **P3.4**: Improve error context and suggestions
- [ ] **P3.5**: Optimization and validation

#### Phase 4: Advanced Optimizations ⏸️ **NOT STARTED**
- [ ] **P4.1**: Introduce service layer
- [ ] **P4.2**: Add infrastructure abstractions
- [ ] **P4.3**: Implement dependency injection
- [ ] **P4.4**: Performance optimizations
- [ ] **P4.5**: Final validation and documentation

### Key Metrics Tracking
- **Code Duplication**: Target 70% reduction (baseline: ~500 lines identified)
- **Test Coverage**: Maintain/improve current levels (commands: 3.8% → 25%+)
- **Build Times**: Maintain current performance
- **Interface Stability**: Zero breaking changes to public APIs

### Risk Mitigation
- All phases include comprehensive testing
- Incremental changes with immediate validation
- Rollback strategies for each major change
- Mock interface preservation throughout

---

## Detailed Implementation Plan

### Constraints and Requirements

#### Immutable Public API Constraints
- **CLI Interface**: All command names, flags, and behaviors must remain unchanged
- **Output Formats**: table/json/yaml schemas must be preserved
- **Build System**: justfile, precommit hooks, GitHub Actions must continue working
- **Testing**: All 123 existing tests must pass (may be enhanced but not broken)
- **Mock Generation**: Interface locations hardcoded in justfile must remain stable

#### Success Criteria
1. Zero breaking changes to public APIs
2. All existing tests continue to pass
3. Build system functions without modification
4. Significant reduction in code duplication
5. Improved error handling consistency
6. Enhanced maintainability and testability

---

## Phase 1: Foundation & Cleanup (Week 1-2)

**Objective**: Establish stable foundation with minimal risk changes that provide immediate value.

### P1.1: Remove Legacy Artifacts and Cleanup (Day 1-2)

**Context**: CLI 2.0 refactoring left backup files and legacy patterns that add confusion.

**Tasks**:
1. **Remove backup files**: `rm.go.bak`, `install.go.bak`
2. **Address global variables**: Remove `outputFormat` global variable coupling
3. **Standardize naming patterns**: Fix inconsistent function names
4. **Clean git artifacts**: Ensure clean working state

**Implementation Details**:
```bash
# Remove backup files
rm internal/commands/rm.go.bak internal/commands/install.go.bak

# Audit for other legacy artifacts
find . -name "*.bak" -o -name "*_old*" -o -name "*_legacy*"
```

**Files to Modify**:
- Remove: `internal/commands/rm.go.bak`, `internal/commands/install.go.bak`
- Clean: `internal/commands/root.go` (remove global `outputFormat`)
- Standardize: Function naming in `internal/commands/shared.go`

**Validation**:
- [ ] All tests pass: `just test`
- [ ] Build succeeds: `just build`
- [ ] No backup files remain
- [ ] Global variable coupling eliminated

### P1.2: Implement Logging Infrastructure (Day 3-4)

**Context**: Multiple TODO comments indicate missing logging mechanism affecting observability.

**Tasks**:
1. **Create logging interface**: `internal/logging/interface.go`
2. **Implement default logger**: Simple structured logging with levels
3. **Add logger to critical operations**: Address specific TODOs
4. **Environment variable control**: `PLONK_LOG_LEVEL` and `PLONK_DEBUG`

**Implementation Details**:
```go
// internal/logging/interface.go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, err error, fields ...Field)
}

type Field struct {
    Key   string
    Value interface{}
}

// Default implementation using structured format
type DefaultLogger struct {
    level Level
    writer io.Writer
}
```

**Files to Create**:
- `internal/logging/interface.go` - Logger interface definition
- `internal/logging/default.go` - Default implementation
- `internal/logging/context.go` - Context integration

**Files to Modify**:
- `internal/commands/status.go` - Address TODOs (lines 131, 142, 153)
- `internal/commands/shared.go` - Address TODOs (lines 266, 278, 289, 347)

**Critical Note**: This logging interface might be used throughout later phases, so design must be stable.

**Validation**:
- [ ] All TODO comments resolved
- [ ] Tests pass with debug logging enabled
- [ ] No performance regression in build times
- [ ] Logger can be disabled for production

### P1.3: Standardize Package Manager Error Handling (Day 5-7)

**Context**: Inconsistent error handling patterns across package managers (20+ `fmt.Errorf` in homebrew, mixed patterns elsewhere).

**Tasks**:
1. **Audit current patterns**: Catalog all error handling in managers
2. **Convert to structured errors**: Replace `fmt.Errorf` with `errors.Wrap`
3. **Add error suggestions**: Enhance user experience with actionable guidance
4. **Standardize error domains**: Ensure consistent domain usage

**Implementation Strategy**:
```go
// Before (homebrew.go):
return fmt.Errorf("failed to install %s (exit code %d): %w\nOutput: %s", name, exitError.ExitCode(), err, outputStr)

// After:
return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", name,
    fmt.Sprintf("package installation failed (exit code %d)", exitError.ExitCode())).
    WithSuggestionMessage(fmt.Sprintf("Check package availability: brew search %s", name))
```

**Files to Modify**:
- `internal/managers/homebrew.go` - Convert 20+ `fmt.Errorf` instances
- `internal/managers/npm.go` - Standardize mixed patterns
- `internal/managers/cargo.go` - Ensure consistency (already mostly structured)
- `internal/state/dotfile_provider.go` - Convert remaining `fmt.Errorf`

**Validation**:
- [ ] No `fmt.Errorf` usage in managers package
- [ ] All errors include suggestions where appropriate
- [ ] Error messages are user-friendly
- [ ] Integration tests pass with better error context

### P1.4: Extract ManagerRegistry Pattern (Day 8-10)

**Context**: Package manager creation boilerplate appears in 5+ locations, highest-impact duplication.

**Tasks**:
1. **Create ManagerRegistry**: `internal/managers/registry.go`
2. **Centralize manager availability checking**
3. **Replace all creation patterns**: Update 5+ locations
4. **Add configuration support**: Manager preferences and overrides

**Implementation Details**:
```go
// internal/managers/registry.go
type ManagerRegistry struct {
    managers map[string]ManagerFactory
    logger   logging.Logger
}

type ManagerFactory func() PackageManager

func NewManagerRegistry(logger logging.Logger) *ManagerRegistry {
    return &ManagerRegistry{
        managers: map[string]ManagerFactory{
            "homebrew": func() PackageManager { return NewHomebrewManager() },
            "npm":      func() PackageManager { return NewNpmManager() },
            "cargo":    func() PackageManager { return NewCargoManager() },
        },
        logger: logger,
    }
}

func (r *ManagerRegistry) GetManager(name string) (PackageManager, error)
func (r *ManagerRegistry) GetAvailableManagers(ctx context.Context) []string
func (r *ManagerRegistry) CreateMultiProvider(ctx context.Context, lockAdapter LockAdapter) (*MultiManagerPackageProvider, error)
```

**Files to Create**:
- `internal/managers/registry.go` - Registry implementation

**Files to Modify**:
- `internal/commands/shared.go` (lines 262-293) - Replace factory pattern
- `internal/commands/status.go` (lines 120-160) - Use registry
- `internal/commands/install.go` (lines 183-194) - Use registry
- `internal/commands/uninstall.go` - Use registry
- `internal/commands/info.go` - Use registry

**Validation**:
- [ ] All package commands use registry
- [ ] Manager availability checking centralized
- [ ] No duplicate manager creation logic
- [ ] Mocks update correctly: `just generate-mocks`

### P1.5: Phase 1 Validation and Testing (Day 11-12)

**Tasks**:
1. **Comprehensive testing**: Run full test suite
2. **Integration validation**: Test all CLI commands end-to-end
3. **Performance baseline**: Ensure no regression
4. **Documentation updates**: Update any relevant docs

**Validation Checklist**:
- [ ] All 123 tests pass: `just test`
- [ ] Precommit hooks pass: `just precommit`
- [ ] Build system works: `just build`
- [ ] Mock generation works: `just generate-mocks`
- [ ] CLI interface unchanged: Manual testing of all commands
- [ ] Code coverage maintained or improved

---

## Phase 2: Core Abstractions (Week 3-4)

**Objective**: Extract common patterns into reusable abstractions that eliminate the majority of command-level duplication.

### P2.1: Create CommandPipeline Abstraction (Day 13-16)

**Context**: Parse flags → Process → Render pattern duplicated across 8+ commands.

**Tasks**:
1. **Design pipeline interface**: Flexible command execution abstraction
2. **Implement standard pipeline**: Flag parsing, processing, and rendering
3. **Create command-specific processors**: Business logic injection points
4. **Migrate commands gradually**: Start with simplest commands

**Implementation Details**:
```go
// internal/commands/pipeline.go
type CommandPipeline struct {
    flags    *SimpleFlags
    format   OutputFormat
    reporter operations.ProgressReporter
    logger   logging.Logger
}

type ProcessorFunc func(ctx context.Context, args []string) ([]operations.OperationResult, error)

func NewCommandPipeline(cmd *cobra.Command, itemType string, logger logging.Logger) (*CommandPipeline, error)
func (p *CommandPipeline) ExecuteWithResults(ctx context.Context, processor ProcessorFunc, args []string) error
func (p *CommandPipeline) HandleOutput(results []operations.OperationResult, domain, operation string) error
```

**Files to Create**:
- `internal/commands/pipeline.go` - Pipeline implementation
- `internal/commands/processors.go` - Standard processor functions

**Files to Modify** (Gradual Migration):
1. `internal/commands/install.go` - Convert to pipeline (simplest first)
2. `internal/commands/uninstall.go` - Convert to pipeline
3. `internal/commands/add.go` - Convert to pipeline
4. `internal/commands/rm.go` - Convert to pipeline

**Critical Decision Point**: The processor interface design will affect all commands. This needs validation before proceeding to other commands.

**Validation**:
- [ ] Pipeline handles all flag combinations correctly
- [ ] Output formats work identically to before
- [ ] Error handling preserves exit codes
- [ ] Progress reporting functions correctly

### P2.2: Extract PathResolver Utility (Day 17-18)

**Context**: Dotfile path resolution logic duplicated 10+ times with complex expansion rules.

**Tasks**:
1. **Create PathResolver**: `internal/paths/resolver.go`
2. **Centralize expansion logic**: Directory expansion, home resolution
3. **Add path validation**: Security checks and permissions
4. **Replace scattered logic**: Update all path resolution calls

**Implementation Details**:
```go
// internal/paths/resolver.go
type PathResolver struct {
    homeDir   string
    configDir string
    logger    logging.Logger
}

func NewPathResolver(logger logging.Logger) (*PathResolver, error)
func (p *PathResolver) ResolveDotfilePath(path string) (string, error)
func (p *PathResolver) GenerateDestinationPath(resolvedPath, configDir string) (string, error)
func (p *PathResolver) ValidatePath(path string) error
func (p *PathResolver) ExpandDirectory(path string) ([]string, error)
```

**Files to Create**:
- `internal/paths/resolver.go` - Path resolution implementation
- `internal/paths/validator.go` - Path validation and security

**Files to Modify**:
- `internal/commands/shared.go` (lines 666-671, 953-993) - Replace path logic
- Multiple dotfile operation functions throughout commands

**Validation**:
- [ ] All existing dotfile operations work identically
- [ ] Path resolution is consistent across commands
- [ ] Security validation prevents directory traversal
- [ ] Performance is maintained or improved

### P2.3: Centralize Configuration Management (Day 19-20)

**Context**: Configuration loading inconsistency with multiple approaches and adapter proliferation.

**Tasks**:
1. **Create ConfigManager**: Single configuration service
2. **Consolidate loading mechanisms**: `LoadConfig` vs `GetOrCreateConfig`
3. **Simplify adapter chain**: Reduce unnecessary indirection
4. **Add configuration validation**: Centralized validation logic

**Implementation Details**:
```go
// internal/config/manager.go
type ConfigManager struct {
    configDir string
    logger    logging.Logger
    cache     *ResolvedConfig // Optional caching
}

func NewConfigManager(configDir string, logger logging.Logger) *ConfigManager
func (c *ConfigManager) Load() (*ResolvedConfig, error)
func (c *ConfigManager) Save(config *ResolvedConfig) error
func (c *ConfigManager) Validate(config *ResolvedConfig) error
func (c *ConfigManager) GetOrCreate() (*ResolvedConfig, error)
```

**Files to Create**:
- `internal/config/manager.go` - Centralized configuration management

**Files to Modify**:
- Simplify: `internal/config/adapters.go` - Reduce adapter complexity
- Update: All commands using configuration loading
- Remove: Duplicate configuration loading patterns

**Critical Note**: Configuration interface changes may affect mock generation. Interface locations must be preserved for build system compatibility.

**Validation**:
- [ ] All commands use consistent configuration loading
- [ ] Adapter complexity reduced without breaking functionality
- [ ] Configuration validation is centralized
- [ ] Mocks generate correctly: `just generate-mocks`

### P2.4: Standardize Output Rendering (Day 21-22)

**Context**: Output formatting spread across multiple files with inconsistent schemas.

**Tasks**:
1. **Create output package**: `internal/output/`
2. **Implement unified renderer**: Single function for all output types
3. **Standardize data structures**: Common result types
4. **Migrate command outputs**: Gradual conversion

**Implementation Details**:
```go
// internal/output/renderer.go
type Renderer struct {
    format OutputFormat
    logger logging.Logger
}

func NewRenderer(format OutputFormat, logger logging.Logger) *Renderer
func (r *Renderer) Render(data interface{}) error
func (r *Renderer) RenderResults(results []operations.OperationResult, operation string) error

// Standard data contracts
type RenderableData interface {
    TableOutput() string
    StructuredData() interface{}
}
```

**Files to Create**:
- `internal/output/renderer.go` - Unified output rendering
- `internal/output/types.go` - Standard data structures
- `internal/output/table.go` - Table formatting utilities

**Files to Modify**:
- Consolidate: Various command output types into standard structures
- Update: Commands to use unified renderer
- Remove: Duplicate output formatting logic

**Validation**:
- [ ] All output formats produce identical results
- [ ] JSON/YAML schemas remain stable for external tools
- [ ] Table output maintains readability
- [ ] Performance is maintained

### P2.5: Phase 2 Migration and Validation (Day 23-24)

**Tasks**:
1. **Complete command migration**: Ensure all commands use new abstractions
2. **Integration testing**: Full CLI workflow testing
3. **Performance validation**: Ensure no regression
4. **Documentation updates**: Update architecture documentation

**Validation Checklist**:
- [ ] All commands migrated to new abstractions
- [ ] Zero breaking changes to CLI interface
- [ ] Build system functions correctly
- [ ] Test coverage improved (target: commands 3.8% → 15%+)
- [ ] Code duplication significantly reduced

---

## Phase 3: Structural Improvements (Week 5-6)

**Objective**: Address complex components and improve overall architecture quality.

### P3.1: Simplify Dotfile Provider Complexity (Day 25-27)

**Context**: `GetActualItems()` method is 300+ lines with multiple responsibilities.

**Tasks**:
1. **Break down complex method**: Extract scanner, filter, expander components
2. **Improve memory efficiency**: Streaming for large directories
3. **Add performance optimization**: Caching and lazy loading
4. **Enhance error handling**: Better error context in dotfile operations

**Implementation Details**:
```go
// internal/dotfiles/scanner.go
type Scanner struct {
    pathResolver *paths.PathResolver
    logger       logging.Logger
}

// internal/dotfiles/filter.go
type Filter struct {
    ignorePatterns []string
    logger         logging.Logger
}

// internal/dotfiles/expander.go
type Expander struct {
    homeDir  string
    logger   logging.Logger
}
```

**Files to Create**:
- `internal/dotfiles/scanner.go` - File system scanning
- `internal/dotfiles/filter.go` - Ignore pattern filtering
- `internal/dotfiles/expander.go` - Directory expansion

**Files to Modify**:
- `internal/state/dotfile_provider.go` - Refactor complex method
- Related dotfile operation functions

**Validation**:
- [ ] Dotfile operations maintain identical behavior
- [ ] Large directory performance improved
- [ ] Memory usage reduced for directory scanning
- [ ] Error messages improved

### P3.2: Consolidate Interface Hierarchies (Day 28-29)

**Context**: Interface proliferation with too many small, related interfaces.

**Tasks**:
1. **Audit interface usage**: Identify consolidation opportunities
2. **Merge related interfaces**: Reduce `ConfigReader`/`ConfigWriter` proliferation
3. **Simplify adapter patterns**: Direct interface implementation where possible
4. **Update mock generation**: Ensure compatibility with build system

**Implementation Strategy**:
```go
// Before: Multiple small interfaces
type ConfigReader interface { ... }
type ConfigWriter interface { ... }
type ConfigValidator interface { ... }

// After: Consolidated interfaces
type ConfigService interface {
    Load(configDir string) (*ResolvedConfig, error)
    Save(configDir string, config *ResolvedConfig) error
    Validate(config *ResolvedConfig) error
}
```

**Files to Modify**:
- `internal/config/interfaces.go` - Consolidate interfaces
- Remove unnecessary adapter files
- Update all interface implementations

**Critical Note**: Interface changes affect mock generation. Must preserve mockable interfaces at expected locations.

**Validation**:
- [ ] Mock generation continues to work: `just generate-mocks`
- [ ] Interface usage simplified without breaking functionality
- [ ] Adapter complexity reduced
- [ ] All tests pass with consolidated interfaces

### P3.3: Enhance Operations Package (Day 30-31)

**Context**: Operations package underutilized as shared services layer.

**Tasks**:
1. **Expand shared functionality**: Generic batch processing
2. **Improve progress reporting**: More flexible and configurable
3. **Add operation contexts**: Enhanced metadata and tracing
4. **Standardize result handling**: Consistent success/failure processing

**Implementation Details**:
```go
// internal/operations/batch.go
type BatchProcessor struct {
    reporter ProgressReporter
    logger   logging.Logger
}

func (b *BatchProcessor) ProcessItems(ctx context.Context, items []string, processor ItemProcessor) []OperationResult

// internal/operations/enhanced_reporter.go
type EnhancedReporter struct {
    operation string
    itemType  string
    logger    logging.Logger
}
```

**Files to Create**:
- `internal/operations/batch.go` - Generic batch processing
- `internal/operations/enhanced_reporter.go` - Improved progress reporting

**Files to Modify**:
- `internal/operations/reporter.go` - Enhance existing reporter
- Commands using operations - Migrate to enhanced patterns

**Validation**:
- [ ] Operations package provides comprehensive shared services
- [ ] Progress reporting improved across all commands
- [ ] Batch processing standardized
- [ ] Error handling enhanced

### P3.4: Improve Error Context and Suggestions (Day 32-33)

**Context**: Error enhancement features underutilized throughout codebase.

**Tasks**:
1. **Add error suggestions**: Actionable guidance for common failures
2. **Enhance error context**: Rich metadata for debugging
3. **Improve user messages**: User-friendly error explanations
4. **Add error categories**: Better error classification

**Implementation Details**:
- Add suggestion messages to common error scenarios
- Enhance error metadata with operation context
- Improve user-facing error messages
- Add error recovery suggestions

**Files to Modify**:
- All error creation sites throughout codebase
- `internal/errors/types.go` - Enhance error types if needed

**Validation**:
- [ ] Error messages significantly improved
- [ ] Users receive actionable guidance
- [ ] Debug information enhanced
- [ ] Error categorization improved

### P3.5: Phase 3 Optimization and Validation (Day 34-35)

**Tasks**:
1. **Performance optimization**: Address any performance regressions
2. **Memory usage optimization**: Efficient resource utilization
3. **Comprehensive testing**: Full system validation
4. **Documentation updates**: Architecture and design updates

**Validation Checklist**:
- [ ] Performance maintained or improved
- [ ] Memory usage optimized
- [ ] Code complexity reduced
- [ ] Test coverage improved (target: 20%+ for commands)

---

## Phase 4: Advanced Optimizations (Week 7-8)

**Objective**: Implement advanced architectural improvements for long-term maintainability.

### P4.1: Introduce Service Layer (Day 36-38)

**Context**: Business logic embedded in command layer violates clean architecture.

**Tasks**:
1. **Extract business services**: `PackageService`, `DotfileService`
2. **Move logic from shared.go**: 1,300+ lines of business logic
3. **Implement service interfaces**: Clean contracts for business operations
4. **Add service composition**: Combine services for complex operations

**Implementation Details**:
```go
// internal/services/package_service.go
type PackageService interface {
    ApplyPackages(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
    AddPackage(ctx context.Context, pkg, manager string) (*AddResult, error)
    RemovePackage(ctx context.Context, pkg, manager string) (*RemoveResult, error)
}

// internal/services/dotfile_service.go
type DotfileService interface {
    ApplyDotfiles(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
    AddDotfile(ctx context.Context, path string, options AddOptions) (*AddResult, error)
}
```

**Files to Create**:
- `internal/services/package_service.go` - Package business logic
- `internal/services/dotfile_service.go` - Dotfile business logic
- `internal/services/interfaces.go` - Service contracts

**Files to Modify**:
- `internal/commands/shared.go` - Extract business logic
- All commands - Use services instead of direct business logic

**Validation**:
- [ ] Business logic properly separated from commands
- [ ] Services testable in isolation
- [ ] Commands simplified to orchestration only
- [ ] Clean architecture principles followed

### P4.2: Add Infrastructure Abstractions (Day 39-40)

**Context**: Missing abstractions for file system and command execution impact testability.

**Tasks**:
1. **Create FileSystem interface**: Abstract file operations
2. **Create CommandExecutor interface**: Abstract command execution
3. **Implement test doubles**: Mock implementations for testing
4. **Integrate with existing code**: Replace direct file system calls

**Implementation Details**:
```go
// internal/infrastructure/filesystem.go
type FileSystem interface {
    Stat(path string) (os.FileInfo, error)
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    Walk(root string, fn filepath.WalkFunc) error
}

// internal/infrastructure/executor.go
type CommandExecutor interface {
    Execute(ctx context.Context, name string, args ...string) error
    Output(ctx context.Context, name string, args ...string) ([]byte, error)
}
```

**Files to Create**:
- `internal/infrastructure/filesystem.go` - File system abstraction
- `internal/infrastructure/executor.go` - Command execution abstraction
- `internal/infrastructure/mocks.go` - Test implementations

**Files to Modify**:
- Package managers - Use CommandExecutor
- Dotfile operations - Use FileSystem interface
- Services - Inject infrastructure dependencies

**Validation**:
- [ ] Infrastructure properly abstracted
- [ ] Testing improved through mocks
- [ ] No regression in functionality
- [ ] Better isolation in unit tests

### P4.3: Implement Dependency Injection (Day 41-42)

**Context**: Missing service container for managing dependencies.

**Tasks**:
1. **Create service container**: Dependency management system
2. **Implement constructor injection**: Services receive dependencies
3. **Update command construction**: Commands receive services via injection
4. **Add configuration-based setup**: Service configuration and wiring

**Implementation Details**:
```go
// internal/container/container.go
type Container struct {
    logger     logging.Logger
    fileSystem infrastructure.FileSystem
    executor   infrastructure.CommandExecutor
    // ... other dependencies
}

func (c *Container) PackageService() services.PackageService
func (c *Container) DotfileService() services.DotfileService
```

**Files to Create**:
- `internal/container/container.go` - Service container
- `internal/container/builder.go` - Container construction

**Files to Modify**:
- `cmd/plonk/main.go` - Use container for setup
- All commands - Receive services via injection

**Validation**:
- [ ] Dependencies properly managed
- [ ] Commands simplified through injection
- [ ] Testing improved through controllable dependencies
- [ ] Configuration-driven service setup

### P4.4: Performance Optimizations (Day 43-44)

**Context**: Opportunities for performance improvements identified during refactoring.

**Tasks**:
1. **Implement caching**: Configuration and state caching
2. **Add streaming operations**: Large directory handling
3. **Optimize memory usage**: Efficient resource utilization
4. **Add concurrent operations**: Where safe and beneficial

**Implementation Areas**:
- Configuration caching for repeated access
- Streaming file discovery for large directories
- Memory-efficient state reconciliation
- Concurrent package manager operations where safe

**Validation**:
- [ ] Performance improved measurably
- [ ] Memory usage optimized
- [ ] No regression in functionality
- [ ] Concurrency safe where implemented

### P4.5: Final Validation and Documentation (Day 45-48)

**Tasks**:
1. **Comprehensive system testing**: Full end-to-end validation
2. **Performance benchmarking**: Measure improvements
3. **Update documentation**: Architecture and design documentation
4. **Create migration guide**: For future developers

**Final Validation Checklist**:
- [ ] All 123+ tests pass
- [ ] CLI interface unchanged
- [ ] Build system functions correctly
- [ ] Performance improved
- [ ] Code duplication reduced by 70%+
- [ ] Test coverage improved significantly
- [ ] Documentation updated

---

## Implementation Notes and Considerations

### Mock Interface Preservation
The justfile contains hardcoded paths for mock generation:
```bash
mockgen -source=internal/managers/common.go -destination=internal/managers/mock_manager.go
mockgen -source=internal/config/interfaces.go -destination=internal/config/mock_config.go
```

**Critical**: Interface files must remain at these locations or justfile must be updated.

### Testing Strategy
- **Maintain existing tests**: All 123 tests must continue to pass
- **Add tests for new abstractions**: New interfaces and implementations need test coverage
- **Improve command testing**: Current 3.8% coverage should improve to 25%+
- **Integration testing**: Validate CLI behavior remains identical

### Risk Mitigation
- **Incremental changes**: Each task includes validation before proceeding
- **Rollback capability**: Each phase can be rolled back independently
- **Interface stability**: Public APIs preserved throughout
- **Build system compatibility**: Continuous validation of build processes

### Decision Points Requiring Validation
1. **P2.1 CommandPipeline Interface**: Processor function signature affects all commands
2. **P2.3 Configuration Interface Changes**: May affect mock generation locations
3. **P3.2 Interface Consolidation**: Must preserve mockable interface locations
4. **P4.1 Service Layer Design**: Business logic extraction strategy needs validation

### Success Metrics
- **Code Duplication**: Target 70% reduction from ~500 identified lines
- **Test Coverage**: Commands from 3.8% to 25%+, Operations from 14.6% to 50%+
- **Build Performance**: Maintain or improve current build times
- **Error Quality**: Measurable improvement in error message actionability

This implementation plan provides comprehensive guidance for systematic refactoring while preserving all external contracts and improving code quality significantly.
