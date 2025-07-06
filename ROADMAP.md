# Plonk Roadmap

This document outlines the future development plans for Plonk, organized by priority and complexity.

## Prioritization Strategy
Balanced approach considering implementation simplicity, user value, dependencies, and project impact. Focus on quick wins that improve codebase quality and provide immediate user benefits.

---

## 🎯 TIER 1: HIGH VALUE + LOW COMPLEXITY + NO DEPENDENCIES
*Quick wins that immediately improve user experience*

### Group A: Code Quality & Cleanup (Simple Infrastructure)
**Value**: 🟢 **High** | **Complexity**: 🟢 **Low** | **Dependencies**: 🟢 **None**

- [ ] Organize imports consistently across all files
- [ ] Standardize function documentation
- [ ] Convert remaining tests to table-driven format

**Why Tier 1**: Simple refactoring tasks that improve maintainability with minimal risk. Can be done independently and make future development easier.

### Group B: Diff Command (Builds on Existing Infrastructure)
**Value**: 🟢 **High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

- [ ] Add tests for diff command structure (Red phase)
- [ ] Implement basic diff command (Green phase)
- [ ] Add tests for file content comparison (Red phase)
- [ ] Implement file diff logic (Green phase)
- [ ] Add tests for config state comparison (Red phase)
- [ ] Implement config vs reality diff (Green phase)
- [ ] Refactor with colored diff output (Refactor phase)

**Why Tier 1**: Builds directly on existing drift detection and dry-run work. High user value for seeing configuration differences. Well-defined scope.

---

## 🚀 TIER 2: HIGH VALUE + MEDIUM COMPLEXITY + SOME DEPENDENCIES
*Significant user value with manageable implementation*

### Group C: Integration Testing (Foundation)
**Value**: 🟢 **High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟡 **Should complete Tier 1 first**

- [ ] Add integration tests for end-to-end workflows (Red phase)
- [ ] Implement comprehensive integration test suite (Green phase)
- [ ] Refactor integration tests with CI/CD support (Refactor phase)

**Why Tier 2**: Critical for stability but needs existing codebase to be clean first. High value for preventing regressions.

### Group D: Additional Shell Support (Natural Extension)
**Value**: 🟡 **Medium-High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

- [ ] Add Bash shell config generation tests (Red phase)
- [ ] Implement Bash shell config generation functionality (Green phase)
- [ ] Add Fish shell config generation tests (Red phase)
- [ ] Implement Fish shell config generation functionality (Green phase)
- [ ] Refactor shell config generation with multi-shell support (Refactor phase)

**Why Tier 2**: Natural extension of existing ZSH work. Clear user value for non-ZSH users. Well-defined patterns to follow.

### Group E: Import Command (User Onboarding)
**Value**: 🟡 **Medium-High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

- [ ] Add shell config parsing tests for common formats (Red phase)
- [ ] Implement basic .zshrc/.bashrc parsing functionality (Green phase)
- [ ] Add tests for plonk.yaml generation from parsed configs (Red phase)
- [ ] Implement plonk import command with YAML suggestion (Green phase)
- [ ] Refactor import command with support for multiple shell types (Refactor phase)

**Why Tier 2**: High value for user adoption. One-time use but critical for migration to plonk.

---

## ⚡ TIER 3: COMPLEX BUT HIGH IMPACT
*Advanced features requiring significant implementation*

### Group F: Watch Mode (Complex File Operations)
**Value**: 🟡 **Medium** | **Complexity**: 🔴 **High** | **Dependencies**: 🟡 **Should have stable foundation**

- [ ] Add tests for watch command structure (Red phase)
- [ ] Implement basic watch command (Green phase)
- [ ] Add tests for file change detection (Red phase)
- [ ] Implement file watcher (Green phase)
- [ ] Add tests for auto-apply on change (Red phase)
- [ ] Implement auto-apply logic (Green phase)
- [ ] Refactor with debouncing and error handling (Refactor phase)

**Why Tier 3**: Complex file watching, debouncing, error handling. High complexity with moderate value. Needs stable base.

### Group G: Repository Infrastructure (DevOps Setup)
**Value**: 🟡 **Medium** | **Complexity**: 🔴 **High** | **Dependencies**: 🔴 **Needs stable codebase**

- [ ] Add pre-commit hook tests for Go formatting (Red phase)
- [ ] Implement pre-commit hooks for Go formatting (Green phase)
- [ ] Add linting tests with golangci-lint (Red phase)
- [ ] Implement golangci-lint configuration and hooks (Green phase)
- [ ] Refactor code quality setup with development workflow integration (Refactor phase)
- [ ] Add development workflow tests (Red phase)
- [ ] Implement development workflow tool (Green phase)
- [ ] Add test coverage enforcement tests (Red phase)
- [ ] Implement test coverage tooling (Green phase)
- [ ] Refactor development workflow with documentation and optimization (Refactor phase)

**Why Tier 3**: Important for project health but complex setup. Should wait until core functionality is stable.

---

## 🎁 TIER 4: NICE-TO-HAVE ENHANCEMENTS
*Lower priority features with specific use cases*

### Group H: Advanced Backup Features
**Value**: 🟡 **Low-Medium** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

- [ ] Add selective restore tests for granular file restoration (Red phase)
- [ ] Implement selective restore functionality (Green phase)
- [ ] Add backup compression tests for space optimization (Red phase)
- [ ] Implement backup compression functionality (Green phase)
- [ ] Add remote backup tests for cloud sync (Red phase)
- [ ] Implement remote backup sync functionality (Green phase)
- [ ] Add backup encryption tests for sensitive data protection (Red phase)
- [ ] Implement backup encryption functionality (Green phase)
- [ ] Refactor advanced backup features with unified management (Refactor phase)

### Group I: Cross-Platform Support
**Value**: 🟡 **Low** | **Complexity**: 🔴 **High** | **Dependencies**: 🟢 **None**

- [ ] Add Windows PowerShell profile tests for cross-platform support (Red phase)
- [ ] Implement Windows PowerShell profile generation (Green phase)
- [ ] Add Linux distribution package manager tests (Red phase)
- [ ] Implement Linux distribution package manager support (Green phase)
- [ ] Refactor cross-platform support with unified configuration (Refactor phase)

### Group J: Package Manager Extensions
**Value**: 🔴 **Low** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

- [ ] Add mas command support tests for Mac App Store integration (Red phase)
- [ ] Implement mas command functionality for App Store apps (Green phase)
- [ ] Refactor package manager integration with mas support (Refactor phase)

### Group K: Environment Snapshots
**Value**: 🔴 **Low** | **Complexity**: 🔴 **High** | **Dependencies**: 🟢 **None**

- [ ] Add full environment snapshot tests (Red phase)
- [ ] Implement plonk snapshot create functionality (Green phase)
- [ ] Add snapshot restoration tests (Red phase)
- [ ] Implement plonk snapshot restore functionality (Green phase)
- [ ] Add snapshot management tests (list, delete) (Red phase)
- [ ] Implement plonk snapshot list/delete functionality (Green phase)
- [ ] Refactor snapshot system with metadata and cross-platform support (Refactor phase)

---

## 🎯 RECOMMENDED EXECUTION ORDER

### Phase 1: Foundation (Tier 1) - 3-5 weeks
1. **Group A**: Code Quality & Cleanup - 1-2 weeks
2. **Group B**: Diff Command - 2-3 weeks

**Rationale**: Quick wins that improve codebase quality and provide immediate user value.

### Phase 2: Core Extensions (Tier 2) - 6-8 weeks
3. **Group C**: Integration Testing - 2 weeks
4. **Group D**: Additional Shell Support - 2-3 weeks
5. **Group E**: Import Command - 2-3 weeks

**Rationale**: Builds on stable foundation to extend core value proposition.

### Phase 3+: Advanced Features (Tier 3+)
6. **Group F**: Watch Mode - if user demand exists
7. **Group G**: Repository Infrastructure - when codebase is mature
8. **Groups H-K**: Nice-to-have features based on user feedback

---

## Additional Future Ideas

- [ ] Create comprehensive user guide
- [ ] Plugin system for custom package managers
- [ ] Configuration templates/presets
- [ ] Multi-machine sync capabilities
- [ ] Configuration conflict resolution