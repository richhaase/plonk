# Plonk Diff Command - Implementation Tasks

## Implementation Checklist

### 1. Command Setup
- [ ] Create `internal/commands/diff.go`
- [ ] Define `diffCmd` with cobra
- [ ] Register command in root
- [ ] Add command help text

### 2. Configuration
- [ ] Add `DiffTool` field to Config struct
- [ ] Update config tests
- [ ] Document default value handling

### 3. Core Implementation
- [ ] Create `runDiff` function
- [ ] Add argument parsing logic
  - [ ] Handle no arguments (all drifted)
  - [ ] Handle file path argument
  - [ ] Resolve paths to config names
- [ ] Integrate with reconciliation
  - [ ] Get drifted dotfiles
  - [ ] Filter by argument if provided
- [ ] Execute diff tool
  - [ ] Load configured tool or default
  - [ ] Build command with file paths
  - [ ] Execute and stream output

### 4. File Resolution
- [ ] Create `resolveDotfilePath` helper
  - [ ] Handle `~/.vimrc` format
  - [ ] Handle `/home/user/.vimrc` format
  - [ ] Handle `vimrc` format
  - [ ] Handle `.vimrc` format
- [ ] Add tests for resolution

### 5. Error Handling
- [ ] No drifted files found
- [ ] Specified file not found
- [ ] Specified file not drifted
- [ ] Diff tool execution errors
- [ ] Missing diff tool (fall back)

### 6. Testing
- [ ] Unit tests
  - [ ] Test argument parsing
  - [ ] Test file resolution
  - [ ] Test diff tool selection
- [ ] Integration tests
  - [ ] Test full flow with mock diff tool
  - [ ] Test with actual file drift
  - [ ] Test error scenarios

### 7. Documentation
- [ ] Create `docs/cmds/diff.md`
- [ ] Update CLI reference
- [ ] Add to README.md commands
- [ ] Update plonk.yaml example

## Code Structure

### internal/commands/diff.go
```go
package commands

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
    "github.com/richhaase/plonk/internal/config"
    "github.com/richhaase/plonk/internal/resources"
    "github.com/richhaase/plonk/internal/resources/dotfiles"
)

var diffCmd = &cobra.Command{
    Use:   "diff [file]",
    Short: "Show differences for drifted dotfiles",
    Long: `Show differences between source and deployed dotfiles that have drifted.

With no arguments, shows diffs for all drifted dotfiles.
With a file argument, shows diff for that specific file only.

Examples:
  plonk diff                # Show all drifted files
  plonk diff ~/.vimrc       # Show diff for specific file
  plonk diff vimrc          # Use config name directly`,
    Args: cobra.MaximumNArgs(1),
    RunE: runDiff,
}

func init() {
    rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
    // Implementation here
}
```

### Key Functions Needed

1. `resolveDotfilePath(arg string, configDir string) (string, error)`
2. `getDriftedDotfiles(ctx context.Context) ([]resources.Item, error)`
3. `executeDiffTool(tool, source, dest string) error`
4. `getDefaultDiffTool() string`

## Definition of Done

- [ ] Command shows diffs for all drifted files when run without args
- [ ] Command shows diff for specific file when given argument
- [ ] Uses git diff --no-index by default
- [ ] Respects diff_tool configuration
- [ ] GUI diff tools work properly
- [ ] All tests pass
- [ ] Documentation complete
