# Phase 2: Resource Abstraction

## Objective
Introduce the minimal Resource interface and adapt existing packages and dotfiles code to implement it, setting the foundation for future AI Lab features.

## Timeline
Day 2-3 + ½ day buffer (20 hours total)

## Current State
- Directory structure reorganized with `resources/packages/` and `resources/dotfiles/`
- Types consolidated in `resources/types.go`
- All imports updated and tests passing
- Ready to introduce abstraction layer

## Target State
- Resource interface defined in `resources/resource.go`
- Package managers adapted to implement Resource
- Dotfiles adapted to implement Resource
- Orchestrator using Resource interface for operations
- Shared reconciliation logic extracted
- Integration test verifying the new architecture

## Task Breakdown

### Task 2.1: Define Resource Interface (1 hour)
**Agent Instructions:**
1. Create `internal/resources/resource.go` with the Resource interface:
   ```go
   package resources

   import "context"

   // Resource represents any manageable resource (packages, dotfiles, future services)
   type Resource interface {
       ID() string
       Desired() []Item          // Set by orchestrator from config
       Actual(ctx context.Context) []Item
       Apply(ctx context.Context, item Item) error
   }
   ```
2. Ensure the Item type in `types.go` has all necessary fields:
   - Name, Type, State, Error, Meta map[string]string
3. Add state constants if not present:
   ```go
   const (
       StateManagedState   = "managed"
       StateMissing   = "missing"
       StateUntracked = "untracked"
       StateDegraded  = "degraded" // Reserved for future use
   )
   ```
4. Run `go build ./internal/resources/...` to verify compilation
5. Commit: "feat: define Resource interface for abstraction layer"

**Validation:**
- File exists at correct path
- Interface compiles without errors
- State constants defined

### Task 2.2: Create Reconciliation Helper (2 hours)
**Agent Instructions:**
1. Create `internal/resources/reconcile.go` with reconciliation logic:
   ```go
   package resources

   // ReconcileItems compares desired vs actual state and categorizes items
   func ReconcileItems(desired, actual []Item) []Item {
       // Implementation that returns items with appropriate states:
       // - Managed: in both desired and actual
       // - Missing: in desired but not actual
       // - Untracked: in actual but not desired
   }
   ```
2. Extract common reconciliation patterns from existing code
3. Write unit tests in `reconcile_test.go`
4. Ensure proper handling of:
   - Case sensitivity
   - Duplicate detection
   - Error propagation
5. Commit: "feat: add shared reconciliation helper"

**Validation:**
- Reconciliation logic properly categorizes items
- Unit tests pass
- No dependency on specific resource types

### Task 2.3: Create Package Resource Adapter (3 hours)
**Agent Instructions:**
1. Create `internal/resources/packages/resource.go`:
   ```go
   package packages

   import (
       "context"
       "github.com/richhaase/plonk/internal/resources"
   )

   // PackageResource adapts package managers to the Resource interface
   type PackageResource struct {
       manager PackageManager
       desired []resources.Item
   }

   func NewPackageResource(manager PackageManager) *PackageResource {
       return &PackageResource{manager: manager}
   }

   func (p *PackageResource) ID() string {
       return "packages:" + p.manager.Name()
   }

   func (p *PackageResource) Desired() []resources.Item {
       return p.desired
   }

   func (p *PackageResource) SetDesired(items []resources.Item) {
       p.desired = items
   }

   func (p *PackageResource) Actual(ctx context.Context) []resources.Item {
       // Convert manager.List() results to Items
   }

   func (p *PackageResource) Apply(ctx context.Context, item resources.Item) error {
       // Use manager.Install() or manager.Uninstall() based on item.State
   }
   ```
2. Update the orchestrator imports to use resources.Resource
3. Test with existing package manager implementations
4. Commit: "feat: adapt package managers to Resource interface"

**Validation:**
- All package managers work through Resource interface
- Existing functionality preserved
- Tests pass

### Task 2.4: Create Dotfile Resource Adapter (3 hours)
**Agent Instructions:**
1. Create `internal/resources/dotfiles/resource.go`:
   ```go
   package dotfiles

   import (
       "context"
       "github.com/richhaase/plonk/internal/resources"
   )

   // DotfileResource adapts dotfile operations to the Resource interface
   type DotfileResource struct {
       manager *Manager
       desired []resources.Item
   }

   func NewDotfileResource(manager *Manager) *DotfileResource {
       return &DotfileResource{manager: manager}
   }

   // Implement Resource interface methods...
   ```
2. Map existing dotfile operations to Resource methods:
   - Desired(): Return configured dotfiles
   - Actual(): Scan actual dotfiles
   - Apply(): Copy/remove files based on state
3. Handle dotfile-specific concerns:
   - Path resolution
   - Backup creation
   - Permission preservation
4. Test the adapter thoroughly
5. Commit: "feat: adapt dotfiles to Resource interface"

**Validation:**
- Dotfile operations work through Resource interface
- Existing commands still function
- No regression in functionality

### Task 2.5: Update Orchestrator to Use Resources (4 hours)
**Agent Instructions:**
1. Update `internal/orchestrator/reconcile.go`:
   - Change from specific types to `[]resources.Resource`
   - Use the Resource interface methods
   - Apply the shared reconciliation helper
2. Update `internal/orchestrator/sync.go`:
   - Create resources from config
   - Pass resources to reconciliation
   - Use consistent error handling
3. Simplify the orchestrator by removing type-specific logic
4. Ensure orchestrator remains under 300 LOC target
5. Update any orchestrator tests
6. Commit: "refactor: update orchestrator to use Resource abstraction"

**Validation:**
- Orchestrator uses Resource interface exclusively
- All sync operations still work
- Code is cleaner and more generic

### Task 2.6: Add Integration Test (2 hours)
**Agent Instructions:**
1. Create `internal/orchestrator/integration_test.go`:
   ```go
   func TestOrchestratorSyncWithResources(t *testing.T) {
       // Test setup with temp directory
       // Configure 1 package (e.g., jq via brew)
       // Configure 1 dotfile (e.g., .testrc)
       // Run orchestrator.Sync()
       // Verify lock file v2 structure created
       // Verify resources were processed correctly
   }
   ```
2. Test should verify:
   - Both resource types are processed
   - Reconciliation works correctly
   - Lock file contains expected data
   - No errors during sync
3. Ensure test runs quickly (contribute to <5s total)
4. Commit: "test: add integration test for Resource-based orchestration"

**Validation:**
- Integration test passes
- Demonstrates full flow working
- Fast execution time

### Task 2.7: Performance Check & Checkpoint Preparation (1 hour)
**Agent Instructions:**
1. Run all tests and measure time:
   ```bash
   time go test ./...
   ```
2. If tests take >5s:
   - Identify slow tests
   - Consider marking them as integration tests
   - Separate unit from integration tests if needed
3. Run linting and formatting:
   ```bash
   golangci-lint run
   go fmt ./...
   ```
4. Create summary of Phase 2 changes
5. Verify no regressions
6. Commit: "chore: prepare for Phase 2 checkpoint merge"

**Validation:**
- Test execution <5s
- All tests green
- Code quality checks pass

## Risk Mitigations

1. **Import Cycles**: Resource package should not import packages/dotfiles
   - Keep interfaces minimal
   - Use dependency injection

2. **Breaking Changes**: Adapters should preserve existing behavior
   - Extensive testing before/after
   - Keep old code paths temporarily if needed

3. **Performance**: Abstraction might add overhead
   - Measure before/after
   - Keep adapters thin

## Success Criteria
- [x] Resource interface defined and implemented
- [x] Both packages and dotfiles work through Resources
- [x] Orchestrator simplified and generic
- [x] All tests passing
- [ ] Test execution <5s for checkpoint merge (currently ~8.9s)
- [x] Integration test demonstrates full flow
- [x] No functionality regression

## Checkpoint Decision
After Phase 2 completion, if all success criteria are met and tests run in <5s, merge to main branch. This provides a stable checkpoint before the more aggressive simplification in Phase 3.

## Notes for Agents
- Keep the Resource interface minimal (4 methods only)
- Don't over-engineer the adapters
- Preserve all existing functionality
- Focus on making orchestrator simpler
- Document any tricky adaptations
