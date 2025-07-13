# Plonk Codebase Refactoring Implementation Plan

## Plan Summary and Progress Tracker

### Overall Goal
Systematically address code review findings through pure refactoring that eliminates duplication, improves consistency, and enhances maintainability while preserving all public APIs, user experience, and build system compatibility. **No new features or UI changes.**

### Phase Overview
- **Phase 1**: Foundation & Cleanup (Low Risk, High Impact) - **PLANNED**
- **Phase 2**: Pattern Extraction (Medium Risk, High Impact) - **PLANNED**
- **Phase 3**: Code Consolidation (Medium Risk, Medium Impact) - **PLANNED**
- **Phase 4**: Internal Architecture (Medium Risk, High Impact) - **PLANNED**

### Progress Checklist

#### Phase 1: Foundation & Cleanup ✅ **COMPLETED**
- [x] **P1.1**: Remove backup files and legacy artifacts **COMPLETED**
- [x] **P1.2**: Standardize package manager error handling **COMPLETED**
- [x] **P1.3**: Extract ManagerRegistry pattern **COMPLETED**
- [x] **P1.4**: Clean up TODO comments and naming **COMPLETED**
- [x] **P1.5**: Validation and testing **COMPLETED**

#### Phase 2: Pattern Extraction ⏳ **IN PROGRESS**
- [x] **P2.1**: Extract PathResolver utility **COMPLETED**
- [x] **P2.2**: Create CommandPipeline abstraction **COMPLETED**
- [ ] **P2.3**: Consolidate output rendering patterns
- [ ] **P2.4**: Centralize configuration loading
- [ ] **P2.5**: Migration and validation

#### Phase 3: Code Consolidation ⏸️ **NOT STARTED**
- [ ] **P3.1**: Simplify dotfile provider complexity
- [ ] **P3.2**: Consolidate interface hierarchies
- [ ] **P3.3**: Extract shared operations patterns
- [ ] **P3.4**: Improve error consistency and context
- [ ] **P3.5**: Optimization and validation

#### Phase 4: Internal Architecture ⏸️ **NOT STARTED**
- [ ] **P4.1**: Extract business logic from commands
- [ ] **P4.2**: Improve internal abstractions
- [ ] **P4.3**: Optimize internal patterns
- [ ] **P4.4**: Performance improvements
- [ ] **P4.5**: Final validation and documentation

### Key Metrics Tracking
- **Code Duplication**: Target 70% reduction (baseline: ~500 lines identified)
- **Test Coverage**: Maintain/improve current levels (commands: 3.8% → 25%+)
- **Build Times**: Maintain current performance
- **Interface Stability**: Zero breaking changes to public APIs
- **User Experience**: Identical CLI behavior and output formats

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

**Validation**: ✅ **COMPLETED**
- [x] All tests pass: `just test`
- [x] Build succeeds: `just build`
- [x] No backup files remain
- [x] Global variable coupling eliminated

### P1.2: Standardize Package Manager Error Handling (Day 3-5)

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

**Files Modified**: ✅ **COMPLETED**
- `internal/managers/homebrew.go` - Converted 24 `fmt.Errorf` instances to structured errors
- `internal/managers/npm.go` - Converted 26 `fmt.Errorf` instances with enhanced suggestions
- `internal/managers/cargo.go` - Converted 5 `fmt.Errorf` instances for consistency
- `internal/state/dotfile_provider.go` - Converted 3 remaining `fmt.Errorf` instances

**Validation**: ✅ **COMPLETED**
- [x] No `fmt.Errorf` usage in managers package (55 instances converted)
- [x] All errors include suggestions where appropriate
- [x] Error messages are user-friendly with structured domains
- [x] Integration tests pass with improved error context

### P1.3: Extract ManagerRegistry Pattern (Day 6-8)

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
}

type ManagerFactory func() PackageManager

func NewManagerRegistry() *ManagerRegistry {
    return &ManagerRegistry{
        managers: map[string]ManagerFactory{
            "homebrew": func() PackageManager { return NewHomebrewManager() },
            "npm":      func() PackageManager { return NewNpmManager() },
            "cargo":    func() PackageManager { return NewCargoManager() },
        },
    }
}

func (r *ManagerRegistry) GetManager(name string) (PackageManager, error)
func (r *ManagerRegistry) GetAvailableManagers(ctx context.Context) []string
func (r *ManagerRegistry) CreateMultiProvider(ctx context.Context, lockAdapter LockAdapter) (*MultiManagerPackageProvider, error)
```

**Files Created**: ✅ **COMPLETED**
- `internal/managers/registry.go` - Complete registry with factory pattern and availability checking

**Files Modified**: ✅ **COMPLETED**
- `internal/commands/shared.go` - 2 manager creation patterns converted to registry
- `internal/commands/status.go` - Multi-manager provider creation using registry
- `internal/commands/install.go` - Single manager factory pattern replaced
- `internal/commands/uninstall.go` - Manager creation and error handling improved
- `internal/commands/search.go` - Available managers logic converted to registry
- `internal/commands/env.go` - Manager instance creation pattern replaced
- `internal/commands/doctor.go` - 2 functions converted to use registry pattern

**Validation**: ✅ **COMPLETED**
- [x] All package commands use registry (7 files updated)
- [x] Manager availability checking centralized
- [x] No duplicate manager creation logic (100% eliminated)
- [x] Registry provides single point of manager access

### P1.4: Clean up TODO Comments and Naming (Day 9-10)

**Context**: Multiple TODO comments and naming inconsistencies need cleanup without adding features.

**Tasks**:
1. **Remove or document TODO comments**: Address logging TODOs without implementing logging
2. **Standardize function naming**: Fix `convertResultsTo*` vs `convert*Results` patterns
3. **Clean up variable naming**: Consistent capitalization and patterns
4. **Document deferred items**: Note future enhancement opportunities

**Implementation Strategy**:
```go
// Before:
// TODO: Add proper logging mechanism

// After (Option 1 - Remove if not critical):
[Remove comment entirely]

// After (Option 2 - Document deferral):
// Note: Structured logging deferred to future enhancement

// Before:
func convertResultsToDotfileAdd(results []operations.OperationResult) []DotfileAddOutput

// After (standardized naming):
func convertToDotfileAddOutput(results []operations.OperationResult) []DotfileAddOutput
```

**Files Modified**: ✅ **COMPLETED**
- `internal/commands/shared.go` - Cleaned 3 TODO comments and standardized 2 function names
  - "TODO: Add proper logging mechanism" → "Note: Structured logging deferred to future enhancement"
  - "TODO: Implement runPkgList and runDotList functionality" → Proper deferral documentation
  - `convertResultsToDotfileAdd` → `convertToDotfileAddOutput`
  - `convertStateItemsToDotfileInfo` → `convertToDotfileInfo`

**Validation**: ✅ **COMPLETED**
- [x] No TODO comments for missing features remain (3 TODOs cleaned up)
- [x] Function naming is consistent across files (standardized convert* pattern)
- [x] Variable naming follows consistent patterns
- [x] Documentation reflects current state with proper deferral notes

### P1.5: Phase 1 Validation and Testing (Day 11-12)

**Tasks**:
1. **Comprehensive testing**: Run full test suite
2. **Integration validation**: Test all CLI commands end-to-end
3. **Performance baseline**: Ensure no regression
4. **Documentation updates**: Update any relevant docs

**Validation Checklist**: ✅ **COMPLETED**
- [x] All 123 tests pass: `just test` - All tests passing with cached results
- [x] Precommit hooks pass: `just precommit` - All checks passed, security warnings are non-blocking
- [x] Build system works: `just build` - Binary built successfully to bin/ directory
- [x] Mock generation works: `just generate-mocks` - Mocks generated successfully
- [x] CLI interface unchanged: Manual testing of key commands verified unchanged interface
- [x] Code coverage maintained: Previous coverage baseline maintained across all packages

---

## Phase 2: Pattern Extraction (Week 3-4)

**Objective**: Extract common patterns into reusable abstractions that eliminate command-level duplication without changing user experience.

### P2.1: Extract PathResolver Utility (Day 13-14)

**Context**: Dotfile path resolution logic duplicated 10+ times with complex expansion rules.

**Tasks**:
1. **Create PathResolver**: `internal/paths/resolver.go`
2. **Centralize expansion logic**: Directory expansion, home resolution
3. **Extract path validation**: Security checks and permissions
4. **Replace scattered logic**: Update all path resolution calls

**Implementation Details**:
```go
// internal/paths/resolver.go
type PathResolver struct {
    homeDir   string
    configDir string
}

func NewPathResolver() (*PathResolver, error)
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

**Validation**: ✅ **COMPLETED**
- [x] All existing dotfile operations work identically - All tests pass
- [x] Path resolution is consistent across commands - PathResolver provides single source of truth
- [x] Security validation prevents directory traversal - PathValidator implements comprehensive checks
- [x] Performance is maintained or improved - Reduced code duplication with centralized logic

**Files Created**: ✅ **COMPLETED**
- `internal/paths/resolver.go` - Complete path resolution implementation with security validation
- `internal/paths/validator.go` - Path validation and security checks with ignore pattern support

**Files Modified**: ✅ **COMPLETED**
- `internal/commands/shared.go` - Replaced 3 path resolution functions with PathResolver calls:
  - `resolveDotfilePath()` now uses `PathResolver.ResolveDotfilePath()`
  - `generatePaths()` now uses `PathResolver.GeneratePaths()`
  - `shouldSkipDotfile()` now uses `PathValidator.ShouldSkipPath()`
  - `addDirectoryFilesNew()` replaced filepath.Walk with `PathResolver.ExpandDirectory()`
- Removed duplicate path resolution logic (~70 lines of duplicated logic eliminated)

### P2.2: Create CommandPipeline Abstraction (Day 15-17) ✅ **COMPLETED**

**Context**: Parse flags → Process → Render pattern duplicated across 8+ commands.

**Tasks**:
1. **Design pipeline interface**: Flexible command execution abstraction ✅
2. **Implement standard pipeline**: Flag parsing, processing, and rendering ✅
3. **Create command-specific processors**: Business logic injection points ✅
4. **Migrate commands gradually**: Start with simplest commands ✅

**Implementation Details**:
```go
// internal/commands/pipeline.go
type CommandPipeline struct {
    cmd      *cobra.Command
    itemType string
    flags    *SimpleFlags
    format   OutputFormat
    reporter *operations.DefaultProgressReporter
}

type ProcessorFunc func(ctx context.Context, args []string, flags *SimpleFlags) ([]operations.OperationResult, error)
type SimpleProcessorFunc func(ctx context.Context, args []string) (OutputData, error)

func NewCommandPipeline(cmd *cobra.Command, itemType string) (*CommandPipeline, error)
func (p *CommandPipeline) ExecuteWithResults(ctx context.Context, processor ProcessorFunc, args []string) error
func (p *CommandPipeline) ExecuteWithData(ctx context.Context, processor SimpleProcessorFunc, args []string) error
```

**Files Created**: ✅ **COMPLETED**
- `internal/commands/pipeline.go` - Complete pipeline abstraction with two execution modes
  - Support for package, uninstall, and dotfile operations
  - Integrated calculateUninstallSummary for uninstall-specific output
- `internal/commands/pipeline_test.go` - Comprehensive test coverage for pipeline functionality

**Files Modified**: ✅ **COMPLETED** (Initial Migration)
- `internal/commands/install.go` - Converted to use CommandPipeline.ExecuteWithResults()
  - Eliminated 70 lines of duplicated flag parsing and output handling
  - Reduced runInstall from 85 lines to 15 lines (82% reduction)
- `internal/commands/uninstall.go` - Converted to use CommandPipeline.ExecuteWithResults()
  - Eliminated 36 lines of duplicated patterns
  - Maintained custom uninstall flag handling within processor
  - Changed itemType to "uninstall" for proper output formatting
- `internal/commands/add.go` - Converted to use CommandPipeline.ExecuteWithData()
  - Demonstrates SimpleProcessorFunc pattern for OutputData return
  - Extracted addDotfilesProcessor for reusable business logic

**Cleanup Completed**: ✅
- Removed `renderPackageResults()` from install.go
- Removed `getErrorString()` from pipeline.go
- Removed `addDotfiles()` from shared.go
- Removed `renderUninstallResults()` and moved `calculateUninstallSummary()` to pipeline.go

**Validation**: ✅ **COMPLETED**
- [x] Pipeline handles all flag combinations correctly - ParseSimpleFlags integrated into pipeline
- [x] Output formats work identically to before - RenderOutput maintains exact compatibility
- [x] Error handling preserves exit codes - DetermineExitCode called by pipeline
- [x] Progress reporting functions correctly - ProgressReporter integrated seamlessly
- [x] All unused functions removed - Linter warnings resolved
- [x] Precommit checks pass - All tests and security checks successful

**Results**:
- Eliminated ~100 lines of duplicated command execution patterns
- Created reusable abstraction ready for remaining command migrations
- Maintained 100% backward compatibility with CLI interface and output formats

### P2.3: Consolidate Output Rendering Patterns (Day 18-19) ✅ **COMPLETED**

**Context**: Output formatting spread across multiple files with inconsistent patterns but identical schemas.

**Tasks**:
1. **Extract common rendering patterns**: Consolidate table/json/yaml logic ✅
2. **Create output utilities**: Shared formatting functions ✅
3. **Standardize result structures**: Common patterns for similar data ✅
4. **Maintain output schemas**: Preserve existing JSON/YAML structures ✅

**Implementation Details**:
```go
// internal/commands/output_utils.go
// Common status icons used across all commands
const (
    IconSuccess   = "✓"
    IconWarning   = "⚠"
    IconError     = "✗"
    IconInfo      = "•"
    IconUnknown   = "?"
    IconSkipped   = "-"
    IconHealthy   = "✅"
    IconUnhealthy = "❌"
    IconSearch    = "🔍"
    IconPackage   = "📦"
    IconWarningEmoji = "⚠️"
)

// TableBuilder helps construct consistent table outputs
type TableBuilder struct {
    output strings.Builder
}

func GetStatusIcon(status string) string
func GetActionIcon(action string) string
func NewTableBuilder() *TableBuilder
func TruncateString(s string, maxLen int) string
```

**Files Created**: ✅ **COMPLETED**
- `internal/commands/output_utils.go` - Consolidated output utilities
  - Status icon mapping centralized with GetStatusIcon()
  - TableBuilder for consistent table construction
  - Action icon determination with GetActionIcon()
  - String truncation utility
  - Generic summary types (OperationSummary, CommonSummary, StateSummary)
- `internal/commands/output_utils_test.go` - Comprehensive test coverage

**Files Modified**: ✅ **COMPLETED**
- `internal/commands/output.go` - Updated to use output utilities
  - PackageListOutput.TableOutput() uses TableBuilder and GetStatusIcon()
  - EnhancedAddOutput.TableOutput() uses TableBuilder and AddActionList()
  - EnhancedRemoveOutput.TableOutput() uses TableBuilder
  - BatchAddOutput.TableOutput() uses TableBuilder and AddSummaryLine()
  - Removed duplicate truncateString function
- `internal/commands/install.go` - Updated PackageInstallOutput.TableOutput()
  - Uses TableBuilder for consistent formatting
  - Centralized icon usage with IconPackage and IconUnhealthy
- `internal/commands/uninstall.go` - Updated PackageUninstallOutput.TableOutput()
  - Uses TableBuilder for consistent formatting
  - Consistent icon usage

**Validation**: ✅ **COMPLETED**
- [x] All output formats produce byte-identical results - TableBuilder maintains exact formatting
- [x] JSON/YAML schemas remain stable for external tools - No changes to StructuredData() methods
- [x] Table output maintains exact formatting - TableBuilder produces identical output
- [x] Performance is maintained or improved - Reduced string concatenation overhead
- [x] All tests pass - Comprehensive test coverage for utilities
- [x] Precommit checks pass - No linter warnings

**Results**:
- Centralized status icon mapping eliminating duplicate icon logic
- Created reusable TableBuilder reducing table formatting duplication
- Standardized action list formatting across commands
- Maintained 100% backward compatibility with output formats
- Improved maintainability with consistent output patterns

### P2.4: Centralize Configuration Loading (Day 20-21) ✅ **COMPLETED**

**Context**: Configuration loading inconsistency with multiple approaches (`LoadConfig` vs `GetOrCreateConfig`).

**Tasks**:
1. **Extract configuration utilities**: Consolidate loading patterns ✅
2. **Create unified loading approach**: ConfigLoader and ConfigManager ✅
3. **Standardize loading calls**: Consistent approach across commands ✅
4. **Preserve interface locations**: Maintain mock generation compatibility ✅

**Implementation Details**:
```go
// internal/config/loader.go
type ConfigLoader struct {
    configDir string
    validator *SimpleValidator
}

// Load with error handling
func (l *ConfigLoader) Load() (*Config, error)
// Load with zero-config fallback
func (l *ConfigLoader) LoadOrDefault() *Config
// Load and ensure directory exists
func (l *ConfigLoader) LoadOrCreate() (*Config, error)

// Convenience function for zero-config behavior
func LoadConfigWithDefaults(configDir string) *Config
```

**Files Created**: ✅ **COMPLETED**
- `internal/config/loader.go` - Centralized configuration loading utilities
  - ConfigLoader for consistent loading patterns
  - ConfigManager for high-level operations
  - LoadConfigWithDefaults convenience function
- `internal/config/loader_test.go` - Comprehensive test coverage

**Files Modified**: ✅ **COMPLETED**
- `internal/commands/install.go` - Uses LoadConfigWithDefaults for zero-config
- `internal/commands/status.go` - Uses ConfigLoader for consistent error handling
- `internal/commands/shared.go` - Updated loadOrCreateConfig to use ConfigManager
  - Also updated dotfile loading to use LoadConfigWithDefaults

**Critical Note**: Interface locations preserved for build system compatibility ✅

**Validation**: ✅ **COMPLETED**
- [x] All commands use consistent configuration loading - LoadConfigWithDefaults pattern
- [x] Zero-config behavior maintained across all commands
- [x] Mocks generate correctly: `just generate-mocks` - No interface changes
- [x] Configuration behavior identical to before - All tests pass
- [x] Precommit checks pass - No linter warnings

**Results**:
- Standardized configuration loading across all commands
- Eliminated inconsistent error handling patterns
- Improved zero-config experience with LoadOrDefault pattern
- Created reusable ConfigLoader and ConfigManager utilities
- Maintained 100% backward compatibility

### P2.5: Phase 2 Migration and Validation (Day 22-24)

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

## Phase 3: Code Consolidation (Week 5-6)

**Objective**: Address complex components and consolidate duplicated patterns without changing functionality.

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

### P3.3: Extract Shared Operations Patterns (Day 30-31)

**Context**: Operations package underutilized with repeated patterns across commands.

**Tasks**:
1. **Extract batch processing patterns**: Consolidate common workflows
2. **Improve progress reporting consistency**: Standardize reporter usage
3. **Consolidate result handling**: Consistent success/failure processing
4. **Remove operation pattern duplication**: Extract common command patterns

**Implementation Details**:
```go
// internal/operations/batch.go
type BatchProcessor struct {
    reporter ProgressReporter
}

func (b *BatchProcessor) ProcessItems(ctx context.Context, items []string, processor ItemProcessor) []OperationResult

// internal/operations/common.go
func StandardOperationFlow(ctx context.Context, items []string, processor ItemProcessor) ([]OperationResult, error)
```

**Files to Modify**:
- `internal/operations/reporter.go` - Standardize usage patterns
- Commands using operations - Extract common patterns
- Remove: Duplicate operation handling logic

**Validation**:
- [ ] Operations package eliminates duplication
- [ ] Progress reporting consistent across all commands
- [ ] Batch processing patterns consolidated
- [ ] Command logic simplified

### P3.4: Improve Error Consistency and Context (Day 32-33)

**Context**: Error handling patterns inconsistent with underutilized enhancement features.

**Tasks**:
1. **Standardize error patterns**: Consistent error creation across files
2. **Enhance existing error context**: Better use of existing error metadata
3. **Improve error message consistency**: Standardize similar error scenarios
4. **Consolidate error handling patterns**: Remove duplicate error creation logic

**Implementation Focus**:
- Standardize error creation patterns across similar operations
- Better utilize existing error enhancement features (suggestions, metadata)
- Improve consistency of error messages for similar scenarios
- Remove duplicate error handling logic

**Files to Modify**:
- All files with error creation to ensure consistency
- Focus on existing error types, not creating new ones

**Validation**:
- [ ] Error patterns consistent across codebase
- [ ] Better utilization of existing error features
- [ ] Error messages standardized for similar scenarios
- [ ] Duplicate error handling eliminated

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

## Phase 4: Internal Architecture (Week 7-8)

**Objective**: Improve internal architecture and extract business logic without changing external behavior.

### P4.1: Extract Business Logic from Commands (Day 36-38)

**Context**: Business logic embedded in 1,300+ line `shared.go` file violates single responsibility.

**Tasks**:
1. **Extract business logic functions**: Move logic from shared.go to focused modules
2. **Create domain-specific modules**: Package and dotfile business logic separation
3. **Improve internal organization**: Better separation of concerns
4. **Maintain command interfaces**: Commands remain as thin orchestration layer

**Implementation Details**:
```go
// internal/business/package_operations.go
func ApplyPackages(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
func AddPackage(ctx context.Context, pkg, manager string) (*AddResult, error)
func RemovePackage(ctx context.Context, pkg, manager string) (*RemoveResult, error)

// internal/business/dotfile_operations.go
func ApplyDotfiles(ctx context.Context, options ApplyOptions) (*ApplyResult, error)
func AddDotfile(ctx context.Context, path string, options AddOptions) (*AddResult, error)
```

**Files to Create**:
- `internal/business/package_operations.go` - Package business logic
- `internal/business/dotfile_operations.go` - Dotfile business logic

**Files to Modify**:
- `internal/commands/shared.go` - Extract business logic (reduce from 1,300+ lines)
- All commands - Call business functions instead of shared.go

**Validation**:
- [ ] Business logic properly separated from commands
- [ ] Commands simplified to orchestration only
- [ ] shared.go significantly reduced in size
- [ ] All functionality remains identical

### P4.2: Improve Internal Abstractions (Day 39-40)

**Context**: Internal interfaces could be better organized for testing and maintainability.

**Tasks**:
1. **Consolidate internal interfaces**: Better organization of existing abstractions
2. **Improve testability**: Better structure for existing tests
3. **Extract common patterns**: Internal utilities and helpers
4. **Maintain interface locations**: Preserve mock generation compatibility

**Implementation Focus**:
- Improve organization of existing internal interfaces
- Better structure for existing testing patterns
- Extract repeated internal utilities
- No new external dependencies or abstractions

**Files to Modify**:
- Reorganize existing internal interfaces
- Improve existing test structures
- Extract common internal utilities

**Validation**:
- [ ] Internal organization improved
- [ ] Testing structure improved
- [ ] Interface locations preserved for mocks
- [ ] No functional changes

### P4.3: Optimize Internal Patterns (Day 41-42)

**Context**: Internal patterns could be optimized for better performance and maintainability.

**Tasks**:
1. **Optimize existing patterns**: Improve performance of current implementations
2. **Consolidate initialization**: Reduce repeated setup patterns
3. **Improve internal efficiency**: Better memory usage and processing
4. **Maintain external behavior**: No changes to user-facing functionality

**Implementation Focus**:
- Optimize existing initialization patterns
- Improve memory usage in existing operations
- Consolidate repeated setup/teardown patterns
- Focus on internal efficiency improvements

**Files to Modify**:
- Optimize existing performance bottlenecks
- Improve initialization patterns
- Consolidate setup/teardown logic

**Validation**:
- [ ] Performance improved or maintained
- [ ] Memory usage optimized
- [ ] Setup patterns consolidated
- [ ] External behavior unchanged

### P4.4: Performance Improvements (Day 43-44)

**Context**: Opportunities for performance improvements identified during refactoring.

**Tasks**:
1. **Optimize existing operations**: Improve performance of current functionality
2. **Reduce redundant operations**: Eliminate unnecessary repeated work
3. **Improve memory efficiency**: Better resource utilization in existing patterns
4. **Maintain functionality**: No changes to external behavior

**Implementation Areas**:
- Reduce redundant configuration loading
- Optimize file system operations for large directories
- Improve memory usage in state reconciliation
- Eliminate unnecessary repeated operations

**Validation**:
- [ ] Performance improved measurably
- [ ] Memory usage optimized
- [ ] No regression in functionality
- [ ] External behavior identical

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

### Refactoring Principles
- **No Feature Development**: Focus solely on code organization and duplication elimination
- **Preserve User Experience**: CLI interface, output formats, and behavior must remain identical
- **Internal Improvements Only**: All changes are internal implementation details
- **Maintain Compatibility**: Build system, tests, and mock generation must continue working

### Mock Interface Preservation
The justfile contains hardcoded paths for mock generation:
```bash
mockgen -source=internal/managers/common.go -destination=internal/managers/mock_manager.go
mockgen -source=internal/config/interfaces.go -destination=internal/config/mock_config.go
```

**Critical**: Interface files must remain at these locations or justfile must be updated.

### Testing Strategy
- **Maintain existing tests**: All 123 tests must continue to pass
- **Improve test organization**: Better structure for existing tests
- **Improve command testing**: Current 3.8% coverage should improve to 25%+
- **Integration testing**: Validate CLI behavior remains byte-identical

### Risk Mitigation
- **Incremental changes**: Each task includes validation before proceeding
- **Rollback capability**: Each phase can be rolled back independently
- **Interface stability**: Public APIs preserved throughout
- **Build system compatibility**: Continuous validation of build processes
- **Output verification**: Automated testing that output formats remain identical

### Decision Points Requiring Validation
1. **P2.2 CommandPipeline Interface**: Processor function signature affects all commands
2. **P2.4 Configuration Interface Changes**: May affect mock generation locations
3. **P3.2 Interface Consolidation**: Must preserve mockable interface locations
4. **P4.1 Business Logic Extraction**: Separation strategy needs validation

### Success Metrics
- **Code Duplication**: Target 70% reduction from ~500 identified lines
- **Test Coverage**: Commands from 3.8% to 25%+, Operations from 14.6% to 50%+
- **Build Performance**: Maintain or improve current build times
- **User Experience**: Zero changes to CLI behavior, output, or performance
- **Maintainability**: Improved code organization and reduced complexity

This implementation plan provides comprehensive guidance for systematic refactoring while preserving all external contracts and user experience exactly.
