# Plan: Merge CLI Package into Commands

## Objective
Merge the `internal/cli` package (1 file) into `internal/commands` package to reduce package sprawl.

## Current State
- `internal/cli/helpers.go` contains:
  - `GetMetadataString()` - extracts metadata from operation results
  - `SimpleFlags` struct - represents command flags
  - `ParseSimpleFlags()` - parses manager and common flags from cobra commands
  - `CompleteDotfilePaths()` - provides shell completion for dotfile paths

- Used by 7 files in `internal/commands/`:
  - shared.go
  - ls.go
  - uninstall.go
  - rm.go
  - install.go
  - add.go
  - dotfiles.go

## Plan

### Step 1: Move the file
1. Copy `internal/cli/helpers.go` to `internal/commands/cli_helpers.go`
2. Update package declaration from `package cli` to `package commands`
3. Add comment explaining these are CLI-specific helpers

### Step 2: Update imports
Replace all imports of `"github.com/richhaase/plonk/internal/cli"` with local references in:
- shared.go - uses `cli.ParseSimpleFlags`
- ls.go - uses `cli.ParseSimpleFlags`
- uninstall.go - uses `cli.ParseSimpleFlags`
- rm.go - uses `cli.CompleteDotfilePaths`
- install.go - uses `cli.ParseSimpleFlags`
- add.go - uses `cli.CompleteDotfilePaths`
- dotfiles.go - uses `cli.CompleteDotfilePaths`

### Step 3: Clean up
1. Delete the empty `internal/cli/` directory
2. Ensure all references are updated

## Expected Changes
- 7 files will have import statements removed
- Function calls will change from `cli.FunctionName` to just `FunctionName`
- Package count reduced by 1

## Risks
- Low risk - this is a simple mechanical refactoring
- All functionality remains the same
- No external API changes

## Validation
1. Run `go build ./...` to ensure compilation
2. Run unit tests: `just test`
3. Run UX/integration tests: `just test-ux`
