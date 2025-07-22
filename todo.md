# Todo List

## Completed
- [x] Fix Bug #2: Change error symbol for 'already managed'
- [x] Fix Bug #5: Add removed count to uninstall summary
- [x] Fix integration test - don't test with dev tools like goimports
- [x] Fix Bug #6: Ensure lock file updates on uninstall
- [x] Fix Bug #3: Info command should check lock file for manager
- [x] Refactor: Centralize package manager list to avoid duplication
- [x] Fix Bug #4: Debug and fix gem manager
- [x] Fix Bug #8: NPM namespaced packages show as 'missing' in state reconciliation
- [x] Fix Bug #7: Improve unavailable manager error messages
- [x] Critical: Remove orphaned and duplicate code throughout codebase (audit completed)
- [x] Fix failing Go manager integration tests (binary name vs module path)

## Integration Test Fixes (Critical Priority - Needed as guardrails for refactoring)
- [ ] Fix failing Cargo manager integration tests (showing wrong package in list)
- [ ] Fix failing Gem manager integration tests (bundler installation)
- [ ] Fix failing search command integration test (pip search error)

## Architecture Refactoring (High Priority - After tests are fixed)
- [ ] Phase 1: Extract business logic from commands (600+ lines in shared.go)
- [ ] Phase 2: Fix domain boundary violations (circular dependencies)
- [ ] Phase 3: Centralize lock file management in runtime
- [ ] Phase 4: Centralize all state management in runtime
- [ ] Phase 5: Clean architecture layers (presentation/application/domain/infrastructure)

## Code Audit Remediation (Medium Priority)
- [ ] Fix all fmt.Errorf usage to use plonk's error system (15 instances found)
- [ ] Resolve panic methods in yaml_config.go
- [ ] Implement table formatting for generic command output
- [ ] Handle multiple package manager installations in info command
- [ ] Fix failing Gem manager integration tests (bundler installation)
- [ ] Fix failing search command integration test (Go manager search)

## Feature Improvements (Medium Priority)
- [ ] Improve handling of packages installed via multiple managers
- [ ] Improve integration tests to be more robust and catch edge cases

## Bug Fixes (Low Priority)
- [ ] Fix Bug #1: Make doctor suggestions OS-aware
- [ ] Fix unsupported manager error message transformation issue
