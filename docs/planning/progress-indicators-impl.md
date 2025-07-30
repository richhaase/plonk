# Progress Indicators Implementation Plan

**Status**: ✅ Completed (2025-07-30)

## Overview
Add progress feedback to long-running operations in plonk to improve user experience.

## Scope
Add progress indicators to these commands:
1. `plonk install` - Multiple packages
2. `plonk uninstall` - Multiple packages
3. `plonk apply` - Packages and dotfiles
4. ~~`plonk search` - Multiple package managers~~ (Not for v1.0)
5. `plonk clone` - Clone + apply operations

## Design Decisions

### Format
Use simple counter format: `[2/5] Installing: htop`

### Update Frequency
- Update for every item (not batched)
- Show progress before starting each operation

### No Additional Flags
- Always show progress (no --quiet flag for v1.0)
- Keep output clean and minimal

## Implementation Details

### 1. Package Operations (install/uninstall)

**Files to modify:**
- `internal/orchestrator/orchestrator.go` - Add progress to ProcessBatch()
- `internal/commands/install.go` - Pass total count to orchestrator
- `internal/commands/uninstall.go` - Pass total count to orchestrator

**Changes:**
```go
// In orchestrator.ProcessBatch
func (o *Orchestrator) ProcessBatch(ctx context.Context, specs []resources.Spec, options ProcessOptions) ([]resources.Result, error) {
    results := make([]resources.Result, 0, len(specs))

    for i, spec := range specs {
        // Add progress output
        if options.ShowProgress && len(specs) > 1 {
            fmt.Printf("[%d/%d] %s: %s\n", i+1, len(specs), options.OperationName, spec.Name)
        }

        // Existing processing code...
    }
}

// Add to ProcessOptions struct
type ProcessOptions struct {
    Parallel      bool
    ShowProgress  bool
    OperationName string // "Installing", "Uninstalling", etc.
}
```

### 2. Apply Command

**Files to modify:**
- `internal/commands/apply.go` - Add progress for packages and dotfiles
- `internal/orchestrator/apply.go` - Add progress tracking

**Changes:**
- Show overall progress: "Applying packages (3 missing)..."
- Then per-package: "[1/3] Installing: wget"
- Then dotfiles: "Applying dotfiles (5 missing)..."
- Then per-dotfile: "[1/5] Deploying: ~/.zshrc"

### 3. ~~Search Command~~ (Not implementing for v1.0)

Per user feedback, search command will not show progress indicators in v1.0.

### 4. Clone Command

**Files to modify:**
- `internal/commands/clone.go` - Add progress for multi-step operation

**Progress stages:**
1. "Cloning repository..."
2. "Detecting required package managers..."
3. "Installing package managers (2 required)..."
4. "[1/2] Installing: npm"
5. "[2/2] Installing: cargo"
6. "Running plonk apply..."
7. (Then normal apply progress)

### 5. Output Consistency

**Create new file:**
- `internal/output/progress.go` - Centralized progress formatting

```go
package output

import "fmt"

// ProgressUpdate prints a progress update in consistent format
func ProgressUpdate(current, total int, operation, item string) {
    if total <= 1 {
        return // No progress for single items
    }
    fmt.Printf("[%d/%d] %s: %s\n", current, total, operation, item)
}

// StageUpdate prints a stage update for multi-stage operations
func StageUpdate(stage string) {
    fmt.Printf("%s\n", stage)
}
```

## Test Plan

### Manual Testing
1. Install multiple packages: `plonk install wget htop jq`
2. Uninstall multiple: `plonk uninstall wget htop`
3. Clone with multiple managers: Test repo with npm and cargo packages
4. Apply with both packages and dotfiles missing
5. Search across all managers: `plonk search git`

### Edge Cases
1. Single item operations (no progress shown)
2. Empty operations (nothing to do)
3. Failed operations mid-batch
4. Ctrl+C during operation

### Expected Output Examples

```bash
$ plonk install wget htop jq
[1/3] Installing: wget
[2/3] Installing: htop
[3/3] Installing: jq
✓ Successfully installed 3 packages

$ plonk apply
Applying packages (2 missing)...
[1/2] Installing: ripgrep
[2/2] Installing: fd
Applying dotfiles (3 missing)...
[1/3] Deploying: ~/.zshrc
[2/3] Deploying: ~/.gitconfig
[3/3] Deploying: ~/.vimrc
✓ Applied 2 packages and 3 dotfiles

$ plonk search git
[Results displayed without progress indicators]
```

## Implementation Order
1. Create output/progress.go with helper functions
2. Update orchestrator.ProcessBatch for install/uninstall
3. Update apply command for two-phase progress
4. Update search command
5. Update clone command
6. Test all commands with various scenarios

## Success Criteria
- Progress shown for all multi-item operations
- Clean, consistent format across commands
- No progress for single-item operations
- No performance impact
- Existing output/formatting preserved
