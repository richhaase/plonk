# Doctor Code Consolidation Plan

**Status**: ⏸️ Skipped (2025-07-30) - Needs design decisions

## Current State Analysis

### Duplication Found

1. **Health Check Execution**
   - `diagnostics.RunHealthChecks()` is called in:
     - `doctor` command (doctor.go)
     - `setup.installDetectedManagers()` (setup.go)
     - Both use the same health report structure

2. **Package Manager Installation**
   - `setup.installSingleManager()` - Used by clone/setup
   - `setup.CheckAndInstallToolsFromReport()` - Used by doctor --fix
   - Both install missing package managers with similar logic

3. **User Prompting**
   - Both paths have their own prompting logic
   - Both show missing managers and ask for confirmation
   - Both show progress during installation

4. **Error Handling**
   - Both provide manual installation instructions on failure
   - Both reference "plonk doctor --fix" for retry

### Key Differences

1. **Context**: Clone operates on detected managers from lock file, doctor operates on all missing managers
2. **Progress**: Clone uses new progress indicators, doctor uses older format
3. **Output**: Slightly different messaging but same underlying operations

## Proposed Solution

### 1. Create Shared Package Manager Service

Create `internal/packagemanager/installer.go`:
```go
package packagemanager

type Installer struct {
    verbose bool
    interactive bool
}

type InstallRequest struct {
    Managers []string
    Context context.Context
    ShowProgress bool
}

type InstallResult struct {
    Manager string
    Success bool
    Error error
}

func (i *Installer) InstallManagers(req InstallRequest) ([]InstallResult, error)
func (i *Installer) InstallSingleManager(ctx context.Context, manager string) error
```

### 2. Consolidate Health Check Usage

Create `internal/packagemanager/detector.go`:
```go
func DetectMissingManagers() ([]string, error)
func DetectRequiredManagers(lockPath string) ([]string, error)
func GetManagersFromHealthReport(report diagnostics.HealthReport) []string
```

### 3. Unify Progress and Output

- Use the new `output.ProgressUpdate()` for all installations
- Standardize messages across doctor and clone commands
- Remove duplicate prompt logic

### 4. Migration Plan

**Phase 1: Create new packagemanager package**
- Move core installation logic from setup
- Create installer service with progress support
- Add tests

**Phase 2: Update doctor command**
- Use new installer service
- Remove CheckAndInstallToolsFromReport
- Use progress indicators

**Phase 3: Update setup/clone**
- Use new installer service
- Remove installSingleManager and related functions
- Simplify installDetectedManagers

**Phase 4: Cleanup**
- Remove old installation functions
- Update tests
- Update documentation

## Benefits

1. **Single Source of Truth**: One place for package manager installation
2. **Consistent UX**: Same progress indicators and messages everywhere
3. **Easier Testing**: Centralized logic is easier to test
4. **Future Expansion**: APT support can be added in one place
5. **Reduced Maintenance**: Fix bugs once, not in multiple places

## Implementation Details

### File Structure
```
internal/
├── packagemanager/         (NEW)
│   ├── installer.go       # Core installation logic
│   ├── installer_test.go  # Tests
│   ├── detector.go        # Detection utilities
│   └── detector_test.go   # Tests
├── commands/
│   └── doctor.go          # Updated to use installer
├── setup/
│   └── setup.go          # Updated to use installer
└── diagnostics/          # Unchanged
```

### Key Functions to Consolidate

1. **From setup.go:**
   - `installSingleManager()` → `Installer.InstallSingleManager()`
   - `installHomebrew()` → Internal to installer
   - `installWithHomebrew()` → Internal to installer
   - `CheckAndInstallToolsFromReport()` → Use `Installer.InstallManagers()`

2. **Shared Logic:**
   - Progress indicators
   - Error handling and manual instructions
   - User prompting (when interactive)
   - Success/failure tracking

### Testing Strategy

1. Unit tests for installer package
2. Integration tests for doctor command
3. Integration tests for clone command
4. Verify identical behavior before/after

## Risk Assessment

**Low Risk**: This is internal refactoring with no user-facing changes
**Testing**: Comprehensive tests ensure no regressions
**Rollback**: Can be done in phases if issues arise

## Success Criteria

1. ✅ No duplicate installation code
2. ✅ Consistent progress indicators
3. ✅ Same user experience for doctor and clone
4. ✅ All tests passing
5. ✅ Easier to add APT support later

## Estimated Effort

- Creating packagemanager package: 0.5 days
- Migrating doctor command: 0.25 days
- Migrating setup/clone: 0.25 days
- Testing and cleanup: 0.5 days
- **Total**: 1.5 days

## Revised Understanding (After User Feedback)

### Key Clarifications
1. **Location**: Code should go in `internal/resources/packages/` (not a new packagemanager directory)
2. **No Prompting**: Remove ALL interactive prompts - use `--no-npm`, `--no-cargo` flags instead
3. **Progress Format**: Use consistent `[1/3] Installing: npm` format everywhere
4. **Scope**: Keep detection logic separate (clone detects from lock, doctor detects from system)
5. **Purpose**: Ensure identical behavior when installing managers, not consolidation for its own sake

### Refined Approach

The consolidation should focus on:
- Moving `installSingleManager()` and related functions from `setup/` to `packages/`
- Ensuring both `clone` and `doctor --fix` use the same installation code
- Removing all prompting logic (rely on flags)
- Using consistent progress indicators

## Outstanding Questions

1. **File Structure**: Should manager installation go in:
   - Existing `operations.go` as `InstallPackageManager()`?
   - New file `manager_installer.go` in packages directory?

2. **Functions to Move**: Should these all move from setup to packages:
   - `installSingleManager()`
   - `installHomebrew()`
   - `installWithHomebrew()`
   - Manual installation instructions?

3. **Interface Design**: Which approach for the shared function:
   ```go
   // Option A: Simple
   func InstallPackageManager(ctx context.Context, manager string) error

   // Option B: With options for future extensibility
   func InstallPackageManager(ctx context.Context, manager string, opts ManagerInstallOptions) error
   ```

4. **Flag Handling**: Where should `--no-npm`, `--no-cargo` flags be checked:
   - In the commands before calling installer?
   - Inside the shared installer function?

5. **Error Messages**: Should manual installation instructions:
   - Stay in shared code (consistent messages)?
   - Be handled by each command (customized context)?

## Decision Needed

This task is being skipped pending design decisions on the above questions. The implementation is straightforward once these decisions are made, but proceeding without clarity could lead to rework.
