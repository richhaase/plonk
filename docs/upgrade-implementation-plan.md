# Complete Implementation Plan: `plonk upgrade` Command

## Phase 1: Interface Extension

### 1.1 Add Upgrade Method to PackageManager Interface
**File**: `internal/resources/packages/interfaces.go`

Add to PackageManager interface:
```go
// Upgrade upgrades one or more packages to their latest versions
// If packages slice is empty, upgrades all installed packages for this manager
Upgrade(ctx context.Context, packages []string) error
```

## Phase 2: Package Manager Implementation

### 2.1 Implement Upgrade() for All 9 Package Managers

**Implementation Pattern for each manager**:

1. **Homebrew** (`homebrew.go`):
   - Empty packages: `brew upgrade` (all packages)
   - Specific packages: `brew upgrade <package1> <package2>...`

2. **NPM** (`npm.go`):
   - Empty packages: `npm update -g` (all global packages)
   - Specific packages: `npm update -g <package1> <package2>...`

3. **Cargo** (`cargo.go`):
   - Empty packages: Get installed list via `cargo install --list`, reinstall each with `cargo install <package>`
   - Specific packages: `cargo install <package1> <package2>...` (reinstalls latest)

4. **Go** (`go.go`):
   - Empty packages: Parse `go list -m all`, reinstall with `go install <package>@latest`
   - Specific packages: `go install <package1>@latest <package2>@latest...`

5. **Gem** (`gem.go`):
   - Empty packages: `gem update` (all gems)
   - Specific packages: `gem update <gem1> <gem2>...`

6. **UV** (`uv.go`):
   - Empty packages: `uv tool list` then `uv tool upgrade` each
   - Specific packages: `uv tool upgrade <tool1> <tool2>...`

7. **Pixi** (`pixi.go`):
   - Empty packages: `pixi global list` then `pixi global update` each
   - Specific packages: `pixi global update <package1> <package2>...`

8. **Composer** (`composer.go`):
   - Empty packages: `composer global update` (all packages)
   - Specific packages: `composer global update <package1> <package2>...`

9. **DotNet** (`dotnet.go`):
    - Empty packages: `dotnet tool list -g` then `dotnet tool update -g` each
    - Specific packages: `dotnet tool update -g <tool1> <tool2>...`

### 2.2 Error Handling Standards
Each implementation should:
- Return detailed error messages with package manager context
- Handle package not found scenarios gracefully
- Support context cancellation
- Use consistent error types across managers

## Phase 3: Command Implementation

### 3.1 Create Upgrade Command Structure
**File**: `internal/commands/upgrade.go`

```go
var upgradeCmd = &cobra.Command{
    Use:   "upgrade [manager:package|package|manager] ...",
    Short: "Upgrade packages across supported package managers",
    RunE:  runUpgrade,
}

func init() {
    rootCmd.AddCommand(upgradeCmd)
    upgradeCmd.Flags().Bool("dry-run", false, "Show upgrade commands without executing")
    upgradeCmd.Flags().String("format", "table", "Output format (table, json, yaml)")
    upgradeCmd.Flags().Bool("verbose", false, "Show detailed upgrade information")
    upgradeCmd.Flags().Bool("quiet", false, "Suppress non-error output")
}
```

### 3.2 Argument Parsing Logic
**Function**: `parseUpgradeArgs(args []string) (upgradeSpec, error)`

Parse into structure:
```go
type upgradeSpec struct {
    UpgradeAll     bool                    // plonk upgrade (no args)
    ManagerTargets map[string][]string     // manager -> packages to upgrade
}

type upgradeTarget struct {
    Manager  string   // empty means "all managers"
    Packages []string // empty means "all packages for manager"
}
```

**Parsing Rules**:
- `plonk upgrade` → `UpgradeAll = true`
- `plonk upgrade brew` → `ManagerTargets["brew"] = []` (empty slice = all packages)
- `plonk upgrade brew:` → Same as above
- `plonk upgrade ripgrep` → Add "ripgrep" to all applicable managers
- `plonk upgrade brew:ripgrep` → `ManagerTargets["brew"] = ["ripgrep"]`

### 3.3 Core Upgrade Orchestration
**Function**: `executeUpgrade(ctx context.Context, spec upgradeSpec) (upgradeResults, error)`

**Logic**:
1. Load current plonk.lock to identify managed packages
2. For each manager in upgradeSpec:
   - Check if manager is available via `IsAvailable()`
   - Filter packages to only those managed by plonk
   - Call `manager.Upgrade(ctx, packages)`
   - Capture results (success/failure per package)
3. Update plonk.lock with new versions via `InstalledVersion()` calls
4. Return comprehensive results

### 3.4 Results Structure
```go
type upgradeResults struct {
    Results []packageUpgradeResult `json:"upgrades"`
    Summary upgradeSummary         `json:"summary"`
}

type packageUpgradeResult struct {
    Manager     string `json:"manager"`
    Package     string `json:"package"`
    FromVersion string `json:"from_version,omitempty"`
    ToVersion   string `json:"to_version,omitempty"`
    Status      string `json:"status"` // "upgraded", "failed", "skipped"
    Error       string `json:"error,omitempty"`
}

type upgradeSummary struct {
    Total    int `json:"total"`
    Upgraded int `json:"upgraded"`
    Failed   int `json:"failed"`
    Skipped  int `json:"skipped"`
}
```

## Phase 4: Output Formatting

### 4.1 Create Upgrade Formatter
**File**: `internal/output/upgrade_formatter.go`

Implement table, JSON, and YAML output matching the documented format.

### 4.2 Progress Indication
- Use existing progress indication patterns from other commands
- Show per-manager progress during upgrades
- Display real-time status updates

## Phase 5: Integration & Lock File Updates

### 5.1 Lock File Updates
After successful upgrades:
- Query each upgraded package for new version via `InstalledVersion()`
- Update plonk.lock with new versions and timestamps
- Use atomic file operations (write to temp, then rename)

### 5.2 Error Recovery
- Partial failures should not prevent lock file updates for successful upgrades
- Provide clear messaging about which packages succeeded/failed
- Exit codes: 0=all success, 1=partial success, 2=total failure

## Phase 6: Testing Strategy

### 6.1 Unit Tests
- Test argument parsing logic with all syntax variations
- Test upgrade orchestration with mocked package managers
- Test error handling scenarios
- Test lock file update logic

### 6.2 Integration Tests
- Test with real package manager commands (using safe test packages)
- Test dry-run functionality
- Test output formatting

### 6.3 BATS Tests
- End-to-end CLI behavior testing
- Test all documented syntax patterns
- Test error scenarios and exit codes

## Implementation Dependencies

### Required Files to Examine:
- `internal/resources/packages/interfaces.go` - Add Upgrade method
- `internal/resources/packages/*.go` - All 10 package manager implementations
- `internal/commands/` - Command structure patterns
- `internal/output/` - Formatting patterns
- `internal/config/` - Lock file operations

### Integration Points:
- Package manager registry for loading managers
- Lock file loading/saving mechanisms
- Existing output formatting infrastructure
- Context and cancellation handling patterns

## Behavior Specification

### Command Syntax and Behavior:

1. **`plonk upgrade`** (no arguments):
   - Upgrades ALL packages managed by plonk across ALL package managers
   - Equivalent to running upgrade commands for each manager

2. **`plonk upgrade <package_name>`**:
   - Upgrades any packages managed by plonk that match that name
   - If ripgrep is installed by both brew and cargo, both versions are upgraded
   - Cross-manager package name matching

3. **`plonk upgrade <package_manager>:<package_name>`**:
   - Upgrades only the package installed by the specified package manager
   - Other managed versions of that package are NOT updated
   - Precise targeting

4. **`plonk upgrade <package_manager>`**:
   - Upgrades ALL packages managed by that specific package manager
   - Use clean syntax: `plonk upgrade brew` (not `brew:`)

## Acceptance Criteria

1. All syntax patterns work exactly as documented
2. Upgrade attempts are made for all requested packages
3. Lock file is updated atomically after successful upgrades
4. Clear error messages for failure scenarios
5. Dry-run mode works correctly
6. All output formats render properly
7. Exit codes reflect operation success/failure appropriately
8. Performance is reasonable for bulk operations
9. Safety is delegated to underlying package managers
10. All 10 package managers implement the Upgrade() interface method

## Implementation Questions - AWAITING RESOLUTION

The implementing agent has raised the following questions about the upgrade implementation. These need definitive answers before implementation can proceed:

### 1. **Lock File Update Timing**
**Question**: The plan mentions updating plonk.lock after successful upgrades by calling `InstalledVersion()`. Should I:
- Update the lock file immediately after each successful package upgrade?
- Or wait until all upgrades are complete and then update all successful ones at once?

**Impact**: Affects atomicity and error recovery behavior.

Update the lock file immediately after each successful package upgrade

### 2. **Cross-Manager Package Matching**
**Question**: For `plonk upgrade <package_name>` (without manager prefix), the plan says to upgrade any packages that match that name across managers. Should I:
- Only upgrade packages that are currently tracked in plonk.lock?
- Or search all managers for packages with that name (even if not in lock file)?

**Impact**: Defines scope of cross-manager operations and interaction with unmanaged packages.

Yes.  only upgrade plonk managed packages ever

### 3. **Package Name Validation**
**Question**: When upgrading specific packages (e.g., `plonk upgrade brew:ripgrep`), should I:
- Validate that the package exists before attempting upgrade?
- Or just attempt the upgrade and let the package manager handle "not found" errors?

**Impact**: Affects error handling patterns and performance.

yes.  only ever upgrade plonk managed packages

### 4. **All-Package Upgrade Scope**
**Question**: For `plonk upgrade` (no args) and `plonk upgrade <manager>`, should I:
- Only upgrade packages that are currently tracked in plonk.lock?
- Or upgrade all installed packages for that manager (even if not managed by plonk)?

**Impact**: Determines whether plonk manages only its own packages or acts as a global upgrade tool.

Only upgrade plonk managed packaged

### 5. **Dry-Run Implementation**
**Question**: For the `--dry-run` flag, should I:
- Show what commands would be executed without running them?
- Or actually query for available updates and show what would be upgraded?

**Impact**: Affects dry-run usefulness and implementation complexity.

Don't include dry-run it has no value in this context without an easy way to check outdated, so NO VALUE HERE.

### 6. **Error Handling Philosophy**
**Question**: Following the "NO PROMPTING" requirement from SelfInstall, should upgrade:
- Always attempt upgrades automatically without confirmation?
- Continue with remaining packages if some fail?
- Use the same non-interactive approach as the clone command?

**Impact**: Ensures consistency with project's no-prompting policy.

Always run upgrades with no prompting.
Always report errors, and continue when updating multiple packages

## Implementation Status

**Status**: ✅ Questions resolved - Ready for implementation

**Resolved Requirements**:
1. **Lock File Authority**: plonk.lock is the authoritative source - only upgrade packages listed there
2. **Immediate Updates**: Update plonk.lock immediately after each successful package upgrade
3. **No Dry-Run**: Remove --dry-run flag from command (no value without easy outdated checking)
4. **No Prompting**: Fully automated approach like clone command
5. **Continue on Errors**: Report failures but continue with remaining packages
6. **Clear Error Messages**: Inform users when they try to upgrade unmanaged packages

**Next Steps**: Proceed with Phase 1 (Interface Extension) implementation.
