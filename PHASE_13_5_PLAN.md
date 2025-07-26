# Phase 13.5: Separate Status and Doctor Commands

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase separates the concerns of `plonk status` and `plonk doctor` commands. Currently, these commands have overlapping functionality that confuses their purpose. This separation will make each command's role clear and focused.

## Objectives

1. Refocus `plonk status` to show only what plonk is managing
2. Refocus `plonk doctor` to show system readiness and health checks
3. Remove overlapping functionality between the commands
4. Update help text to clearly differentiate the commands
5. Maintain the `st` alias for status

## Current State

- `plonk status` shows both managed resources AND system information
- `plonk doctor` shows system health checks and package manager availability
- There's confusion about which command to use for what purpose
- Both commands have some overlapping output

## Desired State

### `plonk status` (and `plonk st`)
**Purpose**: Show what plonk is currently managing

Should display:
- Summary of managed packages (count by manager)
- Summary of managed dotfiles (count)
- Missing packages/dotfiles that need to be installed/linked
- Quick overview of plonk's current state

Should NOT display:
- System information
- Package manager availability
- Configuration file paths
- Health checks

### `plonk doctor`
**Purpose**: Check system readiness for using plonk

Should display:
- System information (OS, arch, etc.)
- Package manager availability
- Configuration file status and location
- Environment variables (PLONK_DIR, etc.)
- Any issues that would prevent plonk from working

Should NOT display:
- Managed package counts
- Missing packages
- Dotfile status

## Implementation Details

### 1. Update Status Command

**In `internal/commands/status.go`:**

Remove all system/health check code and focus on managed resources:

```go
func runStatus(cmd *cobra.Command, args []string) error {
    // Get managed state
    ctx := context.Background()

    // Get package status
    packageResult, _ := packages.Reconcile(ctx, configDir)

    // Get dotfile status
    dotfileResult, _ := dotfiles.Reconcile(ctx, homeDir, configDir)

    // Create focused status output
    output := &StatusOutput{
        Packages: PackageStatus{
            Managed: len(packageResult.Managed),
            Missing: len(packageResult.Missing),
            ByManager: groupByManager(packageResult),
        },
        Dotfiles: DotfileStatus{
            Managed: len(dotfileResult.Managed),
            Missing: len(dotfileResult.Missing),
            Linked: countLinked(dotfileResult),
        },
    }

    return RenderOutput(output, format)
}
```

Example output:
```
Plonk Status

Packages: 42 managed, 3 missing
  homebrew: 30 packages
  npm:      10 packages
  cargo:    2 packages

Dotfiles: 15 managed, 2 missing
  linked:   13 files
  missing:  2 files

Run 'plonk apply' to install missing items.
```

### 2. Update Doctor Command

**In `internal/commands/doctor.go`:**

Remove any managed resource information and focus on system health:

```go
func runDoctor(cmd *cobra.Command, args []string) error {
    checks := []Check{
        checkSystem(),
        checkPackageManagers(),
        checkConfiguration(),
        checkEnvironment(),
        checkPermissions(),
    }

    output := &DoctorOutput{
        System: getSystemInfo(),
        Checks: checks,
        ConfigPath: config.GetConfigPath(),
        PlonkDir: config.GetConfigDir(),
    }

    return RenderOutput(output, format)
}
```

Example output:
```
Plonk Doctor Report

System Information:
  OS:       Darwin 23.5.0 (macOS)
  Arch:     arm64
  Plonk:    v0.8.0

Configuration:
  Config:   ~/.config/plonk/plonk.yaml (exists)
  Lock:     ~/.config/plonk/plonk.lock (exists)
  PLONK_DIR: /Users/user/.config/plonk

Package Manager Availability:
  ✅ homebrew:  v4.2.0
  ✅ npm:       v10.2.0
  ❌ cargo:     not found (install: https://rustup.rs)
  ✅ pip:       v23.3.1
  ❌ gem:       not found
  ✅ go:        v1.21.5

All checks passed! Plonk is ready to use.
```

### 3. Update Help Text

**Status command help:**
```
Show the current state of plonk-managed resources

Usage:
  plonk status [flags]

Aliases:
  status, st

Description:
  Display a summary of all packages and dotfiles managed by plonk,
  including any that are missing and need to be installed.

Flags:
  -o, --output string   Output format (table|json|yaml) (default "table")
  -h, --help           help for status
```

**Doctor command help:**
```
Check system readiness for using plonk

Usage:
  plonk doctor [flags]

Description:
  Perform health checks to ensure your system is properly configured
  for plonk. This includes checking for required package managers,
  configuration files, and system compatibility.

Flags:
  -o, --output string   Output format (table|json|yaml) (default "table")
  -h, --help           help for doctor
```

### 4. Remove Overlap

Ensure these items are in the correct command:
- Package counts → status only
- Manager availability → doctor only
- Config file location → doctor only
- Missing packages → status only
- System info → doctor only

## Testing Requirements

### Unit Tests
- Test status command shows only managed resources
- Test doctor command shows only system health
- Test output formats for both commands
- Ensure no overlap in functionality

### Integration Tests
1. `plonk status` shows managed package/dotfile counts
2. `plonk st` works as alias
3. `plonk doctor` shows system and manager info
4. Neither command shows the other's information
5. Both support -o json/yaml/table

### Manual Testing
- Run both commands and verify clean separation
- Verify help text clearly explains each command
- Check that status focuses on "what" is managed
- Check that doctor focuses on "can" plonk work

## Expected Changes

1. **Modified files:**
   - `internal/commands/status.go` - Remove system checks, focus on managed resources
   - `internal/commands/doctor.go` - Remove resource counts, focus on system health
   - Both: Update help text

2. **Output changes:**
   - Status: Simpler, focused on managed resources
   - Doctor: Comprehensive system health report

3. **No breaking changes:**
   - Commands still exist
   - `st` alias still works
   - Output formats preserved

## Validation Checklist

Before marking complete:
- [ ] Status shows only managed resources
- [ ] Doctor shows only system health
- [ ] No overlapping information between commands
- [ ] Help text clearly differentiates purpose
- [ ] `st` alias continues to work
- [ ] All output formats work correctly
- [ ] Tests updated and passing
- [ ] Clear action items in output (e.g., "Run plonk apply")

## Notes

- This separation makes each command's purpose crystal clear
- Status = "What is plonk managing?"
- Doctor = "Is my system ready for plonk?"
- This should be done before Phase 14 to avoid testing the wrong behavior
- Consider adding summary line to status like "All resources up to date!" or "3 items need attention"

Remember to create `PHASE_13_5_COMPLETION.md` when finished!
