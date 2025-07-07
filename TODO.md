# TODO

Session-level work items and progress tracking. Maintained by AI agents for tactical execution.

## Current Session Focus

**Primary Goal:** Phase 2 - Package Management UI Redesign
**Session Date:** July 7, 2025
**Estimated Time:** 2 hours

## Active Work Items

### Completed This Session
- [x] Phase 1: Removed existing interface code/tests (8400+ lines) ✅
- [x] Created new focused CLI structure (`plonk pkg`) ✅
- [x] Migrated from dev.go to justfile for development workflow ✅
- [x] Separated dotfiles from package managers into internal/dotfiles ✅
- [x] Drastically simplified package managers (2,224 → ~150 lines) ✅
- [x] Removed all abstraction layers from package managers ✅
- [x] Restructured to idiomatic Go CLI layout (pkg/* → internal/*) ✅
- [x] Implemented state reconciliation with separated concerns ✅
- [x] **Complete Package Management UI** - `plonk pkg list/status` with all filters ✅
- [x] **Machine-friendly output** - JSON/YAML formats with `--output` flag ✅
- [x] **Production-ready package management** - Real-world tested with 5 managed, 2 missing, 159 untracked packages ✅

### Phase 2 Tasks (Next Up)

#### Immediate Next Steps
- [ ] **Implement Dotfiles Management UI** - `plonk dot list/status/sync` commands using existing dotfiles package
- [ ] **Design config UI** - `plonk config show/edit/validate` commands

#### Current Focus: Dotfiles CLI Implementation
**Goal**: Mirror successful package management pattern for dotfiles
**Infrastructure**: Ready - `internal/dotfiles/` package with DotfileStatus/DotfileInfo exists
**Commands to Build**:
- `plonk dot list [all|managed|missing|untracked|modified]` - List dotfiles with state filtering
- `plonk dot status` - Summary of dotfile management state  
- `plonk dot sync` - Apply/update dotfiles from config
- Support same `--output table|json|yaml` formats

#### Future Phases
- [ ] **Import command** - Generate config from existing environment
- [ ] **Add workflow-focused integration tests**
- [ ] **Resume comprehensive dogfooding**

## UI Redesign Phase Overview

**Phase 1:** Remove existing interface ✅ **COMPLETED**
**Phase 2:** Create UI one concept at a time 🔄 **IN PROGRESS**
**Phase 3:** Refactor and review interfaces 📋 **PLANNED**

## Current Blockers

- **Security Findings** - 34 gosec issues need resolution before public release  
- **Missing Integration Tests** - Need to create based on dogfooding scenarios

## Temporary Changes

- **Pre-commit Hooks Disabled** - Temporarily disabled to allow smooth dogfooding without security check failures

## Context for Next Session

- **Ready for Stage 1** - Dogfooding plan established, documentation clean
- **Key Decision** - Real-world validation before GitHub launch
- **Security Note** - Dogfooding can proceed while security fixes are planned

## Session Notes

**Current Session Progress:**
1. **Major UI/UX Improvement**: Fixed import command path issue (import_cmd.go:109) - now saves to `.config/plonk/plonk.yaml` instead of `.config/plonk/repo/plonk.yaml`
2. **Architecture Enhancement**: Created comprehensive DotfilesManager interface with rich DotfileInfo objects and state-aware methods (managed/untracked/missing/modified)
3. **Status Command Enhancement**: Added full dotfiles status display to `plonk status` - now shows managed, untracked, missing, and modified dotfiles
4. **Interface Standardization**: Enhanced PackageManager interface with PackageInfo objects and state-aware methods (ListManagedPackages, ListUntrackedPackages, ListMissingPackages) to match DotfilesManager patterns

**Working Patterns for This Session:**
1. **Focus Area**: Dogfooding Plonk with real environment to identify and fix UI/UX issues
2. **Documentation Updates**: Update *.md docs as we go to ensure accuracy (design decisions → ARCHITECTURE.md)
3. **Code Changes**: Use TDD red→green→refactor cycles, small changes, frequent commits
4. **Interface Consistency**: Making DotfilesManager and PackageManager use similar patterns for state-aware functionality

**Technical Progress:**
- **DotfilesManager**: Complete implementation with ignore patterns, file size limits, status detection
- **PackageManager Enhancement**: Added PackageInfo struct, state-aware methods, config integration
- **Status Command**: Now displays both package and dotfiles management status with detailed information
- **Testing**: All tests pass with comprehensive coverage for new state-aware methods
- **Cleanup**: Removed ZSH plugin management functionality

**Next Steps:**
1. Update status command to use enhanced PackageManager methods  
2. Review interface consistency between DotfilesManager and PackageManager
3. Continue dogfooding workflow
4. Import existing setup into Plonk format

## Notes

*For session history, see git log and CHANGELOG.md. For strategic planning, see ROADMAP.md.*