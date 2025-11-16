# Package Manager References - Enhanced Refactoring Guide

> **Status**: 🔄 Revised After Multi-Model Review - 11 of 13 violations resolved (~85% complete)
> **Priority**: Complete manager-agnostic core architecture
> **Target**: v2.1 release milestone
> **Last Updated**: 2025-11-02 (Post-review revisions by Codex CLI, Claude Code, Gemini CLI)

This document catalogs manager-specific code references that violate our goal of maintaining a manager-agnostic core, with enhanced tracking, priorities, and implementation guidance.

**🆕 Recent Changes**: Document has been revised based on comprehensive multi-model review feedback. Key changes include realistic effort estimates (3x increase), new Phase 0 for foundation work, simplified npm/pnpm approach, and explicit testing/migration phases. See revision history at end of document.

## 📊 Progress Overview

### Completion Status
- [ ] **Core Package Code** (4/4 completed - 100%)
- [ ] **CLI/Orchestration** (3/3 completed - 100%)
- [ ] **Shared Types/Config** (5/6 completed - 83%)

> **Phase 0 status (2025-11-15)**: Configuration schema fields and metadata pipeline design are now implemented:
> - `parse_strategy`, `name_transform`, and `metadata_extractors` are defined in `internal/config/managers.go`.
> - `GenericManager` supports `parse_strategy` as an alias for `parse`.
> - A `json-map` parsing mode and metadata pipeline design doc (`docs/plans/pkg-mgr-metadata-pipeline.md`) are in place and used by the default npm manager configuration.

**Total Progress**: 11/13 violations resolved (~85% complete)
**Phase 0**: Foundation work is completed in code; Phase 1 has begun by switching npm/pnpm to JSON-based parsing and removing legacy path-based logic from `GenericManager`.

## ⚠️ Critical Review Findings (2025-11-02)

**Multi-model review conducted** by Codex CLI, Claude Code, and Gemini CLI identified key issues:

### 🔴 **Blocking Issues**
1. **Missing Config Schema Fields**: Proposed fields (`parse_strategy`, `name_transform`, `metadata_extractors`) are now defined in code, but still need to be fully validated and adopted by default manager configs
2. **npm/pnpm Solution Incomplete**: Path extraction approach is overly complex; simpler to migrate to JSON parsing
3. **npm Scoped Package Metadata**: Need consistent metadata storage for scoped packages (`@scope/name`) across lock file and operations
4. **Lock File Migration Unaddressed**: Breaking changes to npm metadata storage will invalidate existing lock files without migration tooling

### ⚠️ **Effort Estimate Corrections**
- **Planned**: 28 hours for core violations
- **Realistic**: 54-72 hours (2.5x multiplier)
- **With testing**: +20-30 hours
- **With migration**: +8-12 hours
- **Total Realistic**: 96-114 hours (~3 weeks)

### 📋 **Recommended Changes**
1. Add **Phase 0** (12-16h): Define config schemas, write metadata pipeline design doc, fix documentation
2. Simplify npm/pnpm: Use JSON parsing instead of path extraction
3. Reprioritize: Manager validation → CRITICAL, npm scopes → LOW, CLI help → HIGH
4. Remove dead code: Delete Go special-case logic (not a built-in manager)
5. Address edge cases: npm workspaces, pnpm store symlinks, scoped package metadata

---

## 🔴 Critical Violations (Core Logic)

### Core Package Code

#### ✅ COMPLETED - npm/pnpm Path Parsing
- **File**: `internal/resources/packages/generic.go` (historical reference at :182)
- **Impact**: ~20 lines of manager-specific parsing logic (now removed)
- **Description**: Previously special-cased npm/pnpm parseable paths (`node_modules`) when parsing `list` output.
- **Resolution**:
  - Default npm and pnpm managers now use JSON-based list commands (`--json`).
  - `GenericManager` gained a `json-map` parser used by the npm default (`dependencies` keys), and pnpm uses JSON-array parsing with `json_field: name`.
  - All npm/pnpm-specific path parsing has been removed from core logic; behavior is entirely configuration-driven.
 - **Status**: ✅ **COMPLETED** (implemented November 2025)

#### ✅ COMPLETED - Go Special-Case Code (Dead Code)
- **File**: `internal/resources/packages/operations.go` (historical references at :124,171,188,223)
- **Impact**: Removed ~15 lines of manager-specific logic for Go, which is not a built-in manager.
- **Previous Behavior**: Special logic for Go packages (binary name extraction, `source_path` metadata) with custom upgrade targeting.
- **Current Design**:
  - All `if manager == "go"` branches have been removed from core operations and upgrade logic.
  - Go can now be added purely as a config-defined manager in `plonk.yaml` without any special handling in core code.
- **Status**: ✅ **COMPLETED** (implemented November 2025)

#### ✅ COMPLETED - Homebrew Alias Expansion
- **File**: `internal/resources/packages/generic.go:104`
- **Status**: ✅ **COMPLETED** - This violation has been resolved
- **Note**: Homebrew alias expansion logic has been removed in recent versions

#### ✅ COMPLETED - Darwin/Linux Conditional
- **File**: `internal/resources/packages/generic.go:85`
- **Status**: ✅ **COMPLETED** - Platform-specific logic moved to configuration

## 🟡 High Priority (Configuration Coupling)

### CLI/Orchestration

#### ✅ COMPLETED - npm Scoped Package Handling
- **File**: `internal/resources/packages/operations.go`
- **Impact**: Previously contained inline special-casing for npm scoped packages; now migrated to a generic metadata pipeline.
- **Previous Behavior**: Hard-coded logic extracted scope/full_name directly from npm package names in core operations.
- **Current Design**:
  - `ManagerConfig.MetadataExtractors` lets npm (and others) define how to derive fields like `scope` and `full_name`.
  - `applyMetadataExtractors` in `operations.go` applies these extractors generically based on config, without referencing manager names in core logic.
  - Default npm config now defines extractors for `scope` and `full_name` using the parsed package name.
- **Status**: ✅ **COMPLETED** (inline npm-special code removed; behavior is config-driven)

#### 🟡 HIGH - Upgrade FullName Tracking
- **Files**: `internal/commands/upgrade.go:54,97,178,194-195`
- **Impact**: ~10 lines across 4 locations
- **Previous Description**: npm-specific FullName tracking for package upgrades.
- **Current Design**:
  - `ManagerConfig.UpgradeTarget` controls how each manager chooses the upgrade target (`name` vs `full_name_preferred`).
  - npm uses `full_name_preferred` so scoped packages upgrade by `full_name` when present; other managers default to `name`.
  - All explicit `info.Manager == "npm"` checks for FullName tracking have been removed from `upgrade.go`.
- **Status**: ✅ **COMPLETED**

#### ✅ COMPLETED - Manager Validation Logic
- **File**: `internal/config/validators.go`
- **Impact**: Removed hard-coded manager lists in validation path
- **Previous State**: Used a static `knownManagers` slice (`apt, brew, npm, uv, gem, go, cargo, test-unavailable`) as a fallback when dynamic registration was absent.
- **Current Design**:
  - Validation is driven by the set registered via `SetValidManagers` (fed from `ManagerRegistry`) when present.
  - When nothing has been registered yet, validation falls back to the keys of `GetDefaultManagers()` (config-driven defaults, not hand-maintained constants).
  - Tests can explicitly control the allowed managers by calling `SetValidManagers`.
- **Status**: ✅ **COMPLETED** (hard-coded `knownManagers` list removed; now derived from config and registry)

### Shared Types/Config

#### ✅ COMPLETED - Default Manager Configuration
- **File**: `internal/config/managers.go:15-25`
- **Status**: ✅ **COMPLETED** - Now uses v2 YAML-based configuration system

#### ✅ COMPLETED - Manager Name Constants
- **File**: `internal/types/package.go:12-18`
- **Status**: ✅ **COMPLETED** - Constants removed, using dynamic registry

#### ❌ NOT STARTED - Help Text Generation
- **File**: `internal/commands/root.go`, `internal/commands/install.go`, etc.
- **Status**: ❌ **NOT STARTED** - Help text still contains hard-coded manager examples
- **Note**: Examples in CLI help are acceptable, but dynamic generation from registry would be better

#### ✅ COMPLETED - Error Message Templates
- **File**: `internal/errors/messages.go:23,67`
- **Status**: ✅ **COMPLETED** - Using templated error messages

#### ✅ COMPLETED - Manager Feature Matrix
- **File**: `internal/config/features.go:12-45`
- **Status**: ✅ **COMPLETED** - Feature matrix now configuration-driven

## 🟠 Medium Priority (UX References)

### Documentation/Comments

#### 🟠 MEDIUM - CLI Help Manager Lists
- **File**: `cmd/plonk/main.go`, `internal/commands/*.go`
- **Impact**: User-facing help text mentions specific managers
- **Description**: Hard-coded manager examples in CLI help (install.go, uninstall.go, upgrade.go) have been replaced with dynamically generated examples for those commands, driven by configured managers; other commands still use static examples.
- **Refactoring Solution**: Generate help examples from registry/ManagerConfig for all commands, or keep remaining examples generic.
- **Effort**: 2 hours
- **Status**: 🟡 PARTIALLY COMPLETED

## 🟢 Low Priority (Documentation)

All documentation violations have been resolved through the v2 configuration system.

## 🚀 Implementation Strategy

### Phase 0: Foundation (NEW - Week 1)
**Goal**: Address blocking issues before implementation
**Duration**: 12-16 hours

1. **Fix Documentation Errors** (2 hours)
   - Correct file references (✅ COMPLETED 2025-11-02)
   - Update completion status markers (✅ COMPLETED 2025-11-02)
   - Document reviewer feedback (✅ COMPLETED 2025-11-02)

2. **Define Missing Config Schema Fields** (4-6 hours)
   - Add `ParseStrategy`, `NameTransform`, `MetadataExtractors` to `ManagerConfig` (✅ COMPLETED 2025-11-15)
   - Write validation logic for regex patterns (ReDoS protection)
   - Define `TransformConfig` and `ExtractorConfig` types (✅ COMPLETED 2025-11-15 as `NameTransformConfig` and `MetadataExtractorConfig`)

3. **Write Metadata Pipeline Design Doc** (4-6 hours)
   - Sequence diagram: raw output → parse → extract → transform → lock
   - Error handling strategy at each stage
   - Backward compatibility approach for lock files
   - Lock file versioning scheme
   - ✅ COMPLETED 2025-11-15 in `docs/plans/pkg-mgr-metadata-pipeline.md`

4. **Prototype Critical Changes** (2-4 hours)
   - Test npm JSON parsing approach (prototype `json-map` parser implemented in `GenericManager`, but not yet used by default configs)
   - Validate Go metadata storage design
   - Confirm no breaking changes to existing workflows (✅ CONFIRMED via `just test`; defaults unchanged for npm/pnpm)

### Phase 1: Critical Path (Week 2)
**Goal**: Eliminate core logic violations with simplified approach
**Duration**: 14-18 hours

1. **Manager Validation Refactor** (2 hours) → ELEVATED TO CRITICAL
   - Move from hard-coded list to registry-based validation
   - Update `internal/config/validators.go`
   - Remove `knownManagers` fallback

2. **Migrate npm/pnpm to JSON Parsing** (6-8 hours) → SIMPLIFIED APPROACH
   - Change npm/pnpm configs to use `--json` output
   - Implement JSON parsing with nested object support
   - Remove 22 lines of path parsing logic from `generic.go`
   - Handle scoped packages via JSON structure
   - Test on Windows/Linux paths

3. **CLI Help Generation** (2 hours) → ELEVATED TO HIGH
   - Generate help examples from manager registry
   - Quick win with high visibility
   - Update `internal/commands/install.go`, `uninstall.go`, etc.

4. **Metadata Framework Design** (4-6 hours)
   - Design metadata storage system for npm scoped packages
   - Preserve backward compatibility with existing lock files
   - Plan npm `scope` + `full_name` handling in lock file v2

### Phase 2: Enhancement (Week 3)
**Goal**: Complete configuration migration and code cleanup
**Duration**: 8-12 hours

1. **Remove Go Special-Case Code** (1 hour) → DEAD CODE REMOVAL
   - Delete `if manager == "go"` blocks from operations.go:124,171,188,223
   - Remove `ExtractBinaryNameFromPath` helper if unused elsewhere
   - Update tests to remove Go-specific test cases
   - Document that Go can be added as custom manager if needed

2. **Upgrade FullName Tracking** (6 hours) → COMPLETED
   - Per-manager upgrade targeting is now driven by configuration via `ManagerConfig.UpgradeTarget`.
   - npm uses `UpgradeTarget: "full_name_preferred"` so scoped packages upgrade by `full_name` when present; other managers default to using `name`.

3. **npm Scoped Packages** (4-6 hours) → DEMOTED TO LOW PRIORITY
   - Implement metadata extraction framework (if not covered by JSON parsing)
   - Add scope/name pattern matching to configs
   - Test with complex npm package structures
   - NOTE: May be resolved by JSON parsing migration in Phase 1

### Phase 3: Testing & Documentation (Week 4)
**Goal**: Comprehensive testing and edge case coverage
**Duration**: 20-30 hours

1. **Comprehensive Test Suite** (20-30 hours) → ADDED PER REVIEWER FEEDBACK
   - Table-driven tests for all parse strategies (10+ cases each)
   - Edge cases: empty output, malformed JSON, missing fields
   - Windows/Linux path handling tests
   - npm workspaces, pnpm store symlinks, Go replace directives
   - Integration tests with mocked executor
   - Regression tests for all supported managers
   - Target: >95% coverage for refactored components

2. **Lock File Migration** (8-12 hours) → ADDED PER REVIEWER FEEDBACK
   - Lock file version bump (v1 → v2)
   - Backward compatibility layer
   - Migration tool or auto-migration on load
   - Documentation for manual migration if needed
   - Test migrations with real-world lock files

3. **Documentation & Error Handling** (4-6 hours)
   - Document new config schema fields
   - Add examples for custom managers
   - Improve error messages for parse failures
   - Add validation error messages for invalid regex patterns
   - Document edge cases and limitations

### Phase 4: Polish (Week 5)
**Goal**: Final UX improvements
**Duration**: 2-4 hours

1. **Install Suggestions & Clone Descriptions** (2-4 hours) → ADDED PER CODEX REVIEW
   - Move hard-coded install suggestions from `operations.go:304-319` to config
   - Move manager descriptions from `clone/setup.go:176-214` to config
   - Add optional fields to ManagerConfig: `description`, `install_hint`, `help_url`
   - Update orchestration to use config-driven suggestions

## 🏗️ Architectural Patterns

### Enhanced ManagerConfig Schema
```yaml
# Example enhanced manager configuration
name: "npm"
parse_strategy: "json"  # or "parseable", "plain"
name_transform:
  type: "regex"
  pattern: "^(@[^/]+/)?(.*)$"
  replacement: "$2"
metadata_extractors:
  scope:
    pattern: "^@([^/]+)/.*$"
    group: 1
  version:
    source: "json_field"
    field: "version"
commands:
  list:
    json_output: ["--json", "--depth=0"]
    parseable_output: ["--parseable", "--depth=0"]
```

### Config-Driven Parsing Strategy
```go
type ParseStrategy interface {
    Parse(output string, config ManagerConfig) ([]Package, error)
}

type JSONStrategy struct{}
type ParseableStrategy struct{}
type PlainTextStrategy struct{}

func (p *PackageParser) GetStrategy(config ManagerConfig) ParseStrategy {
    switch config.ParseStrategy {
    case "json":
        return &JSONStrategy{}
    case "parseable":
        return &ParseableStrategy{}
    default:
        return &PlainTextStrategy{}
    }
}
```

### Registry-Driven Validation
```go
type ManagerRegistry struct {
    managers map[string]*ManagerConfig
}

func (r *ManagerRegistry) Validate(name string) error {
    if _, exists := r.managers[name]; !exists {
        return fmt.Errorf("unknown manager: %s", name)
    }
    return nil
}
```

## 📋 Developer Guidelines

### ❌ Never Do
```go
// DON'T: Manager-specific logic in core code
if manager == "npm" {
    return parseNpmOutput(output)
}

// DON'T: Hard-coded manager lists
validManagers := []string{"brew", "npm", "cargo"}

// DON'T: Manager-specific data structures
type NpmPackage struct {
    Scope string
    Name  string
}
```

### ✅ Always Do
```go
// DO: Configuration-driven behavior
strategy := getParseStrategy(config.ParseStrategy)
return strategy.Parse(output)

// DO: Registry-based validation
return registry.ValidateManager(name)

// DO: Extensible data structures
type Package struct {
    Name     string
    Metadata map[string]string // for scope, etc.
}
```

## 🎯 Success Metrics

- [ ] **Zero manager names in core logic** (`internal/resources/packages/`)
- [ ] **Configuration-driven parsing** (All managers use same generic parser)
- [ ] **Registry-based validation** (No hard-coded manager lists)
- [ ] **Custom manager support** (Users can add managers without code changes)
- [ ] **Test coverage >95%** for all refactored components

## 📊 Repository Analysis Summary

**Analysis Date**: 2025-11-02
**Total Files Analyzed**: 47
**Documentation Accuracy**: 100% ✓
**References Verified**: 13/13 accurate, 1 marked as resolved

### Current Architecture Quality
- **v2 Config System**: ✅ Excellent implementation
- **Manager Registry**: ✅ Well-designed, extensible
- **Custom Manager Support**: ✅ YAML-based, no code required
- **Technical Debt**: 🟡 Low (~50 lines in 1,500+ total)

### Supported Managers
Currently supporting 8 built-in managers: `brew`, `npm`, `pnpm`, `cargo`, `gem`, `uv`, `conda`, `pipx` + custom user-defined managers via YAML configuration.

---

## Notes

The violations above represent the remaining work to achieve complete manager-agnosticism. The v2 architecture foundation is excellent - this document provides the roadmap to complete the migration.

**Key Principle**: Manager-specific logic belongs in configuration files, not in core Go code. The generic package management engine should be capable of handling any package manager through configuration alone.

**Testing Strategy**: Each refactoring should maintain 100% backward compatibility and include comprehensive tests covering edge cases for all supported managers. Based on reviewer feedback, testing is now explicitly tracked as Phase 3 with 20-30 hours allocated.

**Performance Impact**: Configuration-driven approaches may have minimal performance overhead but provide significant architectural benefits and extensibility.

## 📝 Revision History

### 2025-11-02 (Evening): Go Manager Correction
**Reviewer**: User verification + code analysis

**Critical Correction**:
- ❌ **Removed**: Go manager references (Go is NOT a built-in manager)
- ✅ **Corrected**: Changed from "Go import path metadata" to "Remove Go dead code"
- ✅ **Updated**: Lock file migration examples to use npm scoped packages instead
- ✅ **Fixed**: Supported managers list (8 built-in, not including Go)
- ✅ **Clarified**: Blocking issue #3 now focuses on npm scoped packages
- **Impact**: Reduces Phase 2 effort from 12-16h to 1h (dead code removal)

### 2025-11-02 (Afternoon): Multi-Model Review & Major Updates
**Reviewers**: Codex CLI, Claude Code, Gemini CLI

**Changes Made**:
1. ✅ Fixed documentation errors (file references, completion status)
2. ✅ Added Phase 0 (Foundation) - 12-16 hours
3. ✅ Updated effort estimates (28h → 96-114h realistic)
4. ✅ Added comprehensive testing phase (20-30h)
5. ✅ Added lock file migration phase (8-12h)
6. ✅ Reprioritized violations:
   - Manager validation: HIGH → CRITICAL
   - CLI help generation: MEDIUM → HIGH
   - npm scoped packages: HIGH → LOW
7. ✅ Simplified npm/pnpm approach: Use JSON parsing instead of path extraction
8. ✅ Redesigned Go handling: Metadata storage system instead of name transform
9. ✅ Added install suggestions & clone descriptions to roadmap
10. ✅ Updated total progress: 54% → 46% (status corrections)

**Key Insights from Review**:
- All three reviewers agreed: architecture is sound, but implementation needs redesign
- Critical gap: Missing config schema fields must be defined before implementation
- Risk identified: Lock file format changes require migration tooling
- Consensus: npm/pnpm JSON parsing is simpler and more robust than path extraction
- Testing was underspecified; now has dedicated phase with explicit coverage targets

**Estimated Timeline After Revisions**:
- Phase 0 (Foundation): 12-16 hours (1-2 days)
- Phase 1 (Critical Path): 14-18 hours (2-3 days)
- Phase 2 (Enhancement): 8-12 hours (1-2 days) ← REDUCED (Go removal instead of metadata system)
- Phase 3 (Testing): 20-30 hours (3-4 days)
- Phase 4 (Polish): 2-4 hours (0.5 day)
- **Total**: 56-80 hours (~1.5-2 weeks focused work)
- **With buffer**: 70-100 hours (~2-3 weeks realistic)
