# Plan: Delete Constants Package

## Objective
Delete the `internal/constants` package and move constants to their logical domain packages.

## Current State
The constants package contains two files:

1. **files.go** - File-related constants:
   - `LockFileName = "plonk.lock"`
   - `ConfigFileName = "plonk.yaml"`
   - `LockFileVersion = 1`
   - `DefaultOperationTimeout = 60`
   - `DefaultPackageTimeout = 300`
   - `DefaultDotfileTimeout = 30`

2. **managers.go** - Manager-related constants:
   - `SupportedManagers` - slice of supported package managers
   - `DefaultManager = "homebrew"`

Used by 7 files (including 1 test file).

## Analysis
Constants should live close to where they're used. The current split suggests:
- File/config constants → `config` package
- Lock file constants → `lock` package
- Manager constants → `managers` package
- Timeout constants → `config` package (since they're config defaults)

## Plan

### Step 1: Move constants to appropriate packages

1. **To `internal/config/constants.go`** (new file):
   - `ConfigFileName`
   - `DefaultOperationTimeout`
   - `DefaultPackageTimeout`
   - `DefaultDotfileTimeout`

2. **To `internal/lock/constants.go`** (new file):
   - `LockFileName`
   - `LockFileVersion`

3. **To `internal/managers/constants.go`** (new file):
   - `SupportedManagers`
   - `DefaultManager`

### Step 2: Update imports
Replace `"github.com/richhaase/plonk/internal/constants"` with appropriate package imports in:
- commands/install.go → needs managers
- commands/uninstall.go → needs managers
- commands/doctor.go → needs managers
- config/compat.go → needs managers
- lock/yaml_lock.go → needs lock constants
- lock/yaml_lock_test.go → needs lock constants
- managers/registry.go → needs managers

### Step 3: Update references
Change constant references:
- `constants.LockFileName` → `lock.LockFileName`
- `constants.LockFileVersion` → `lock.LockFileVersion`
- `constants.SupportedManagers` → `managers.SupportedManagers`
- `constants.DefaultManager` → `managers.DefaultManager`

### Step 4: Clean up
Delete the `internal/constants/` directory

## Expected Changes
- 6 source files + 1 test file will have import changes
- Constants will be co-located with their domains
- Package count reduced by 1

## Risks
- Low risk - simple constant relocation
- Need to ensure all references are updated
- No functional changes

## Validation
1. Run `go build ./...` to ensure compilation
2. Run unit tests: `just test`
3. Run UX/integration tests: `just test-ux`
