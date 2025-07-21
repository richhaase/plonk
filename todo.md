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

## Pending
- [ ] Fix Bug #1: Make doctor suggestions OS-aware
- [ ] Fix Bug #7: Improve unavailable manager error messages (moved to end as it benefits from OS-awareness)
- [ ] Improve handling of packages installed via multiple managers
- [ ] Improve integration tests to be more robust and catch edge cases
