# Task 007: Replace Emojis with Minimal Colorization

## Overview
Replace all emoji usage in plonk with colored status indicators. This creates a more professional appearance while maintaining clarity and scannability.

## Design Principles
1. **Color status words only** - Not full lines or sentences
2. **Consistent meaning** - Same status always uses same color
3. **Remove all emojis** - No Unicode emoji characters anywhere
4. **Minimal visual noise** - Use color sparingly for maximum impact

## Implementation Plan

### Phase 1: Create Color Infrastructure

#### 1.1 Create Color Utilities
**File**: `/internal/output/colors.go` (new file)

```go
package output

import (
    "os"
    "github.com/fatih/color"
)

// Status color functions
var (
    StatusSuccess = color.New(color.FgGreen).SprintFunc()
    StatusError   = color.New(color.FgRed).SprintFunc()
    StatusWarning = color.New(color.FgYellow).SprintFunc()
    StatusInfo    = color.New(color.FgBlue).SprintFunc()
    StatusSkip    = color.New(color.Faint).SprintFunc()
)

// Common status words
func Managed() string    { return StatusSuccess("managed") }
func Missing() string    { return StatusError("missing") }
func Unmanaged() string  { return StatusWarning("unmanaged") }
func Available() string  { return StatusSuccess("available") }
func NotAvailable() string { return StatusError("not available") }
// ... etc
```

### Phase 2: Systematic Emoji Replacement

#### 2.1 Status Command (`/internal/commands/status.go`)
**Replace**:
- `✅ managed` → `managed` (green)
- `❌ missing` → `missing` (red)
- `⚠️  unmanaged` → `unmanaged` (yellow)
- `✅ deployed` → `deployed` (green)

#### 2.2 Doctor Command (`/internal/commands/doctor.go`)
**Replace**:
- `❌` in issues → Remove, let error text stand alone
- `💡` in suggestions → Remove entirely
- `%s: ✅` → `%s: available` (green)
- `%s: ❌` → `%s: not available` (red)

#### 2.3 Setup Command (`/internal/setup/setup.go`)
**Replace**:
- `✅ Repository cloned successfully` → `Repository cloned successfully`
- `✅ All required tools are available` → `All required tools are available`
- `❌ Failed to install %s` → `Failed to install %s` (color "Failed")
- `💡 Manual installation:` → `Manual installation:`
- Remove all `⚠️`, `ℹ️` prefixes

#### 2.4 Apply Command (`/internal/commands/apply.go`)
**Replace**:
- `⚠️  Some operations failed` → `Some operations failed` (color "failed")

#### 2.5 Info Command (`/internal/commands/info.go`)
**Replace**:
- `✅ Installed (not managed)` → `Installed (not managed)` (color "Installed")
- `❌ Not found` → `Not found` (color entire phrase)
- `⚠️ No package managers available` → `No package managers available`
- `⚠️ Manager unavailable` → `Manager unavailable`

#### 2.6 Search Command (`/internal/commands/search.go`)
**Replace**:
- `❌ %s` → Just the message, color if it contains "not found"
- `⚠️  %s` → Just the message

#### 2.7 Output Formatters (`/internal/output/formatters.go`)
**Replace**:
- `✅` → Color the action word (deployed, added, etc.)
- `❌` → Color "failed" or "error"
- `⏭️` → `skipped` (dim color)
- Remove emoji status assignment logic

#### 2.8 Other Commands
**Remove**:
- All `ℹ️` info emojis
- All `🔍` search emojis
- All `📦` package emojis
- All `🔧` fix emojis

### Phase 3: Special Cases

#### 3.1 Diagnostics/Health (`/internal/diagnostics/health.go`)
**Replace**:
- `✅ %s: %s` → `%s: %s` with "available" colored green
- `⚠️  %s: %s (exists but not in PATH)` → Same without emoji, color "not in PATH"
- `ℹ️  %s: %s (directory does not exist)` → Remove emoji

#### 3.2 Environment Command (`/internal/commands/env.go`)
**Replace**:
- `✅ Valid` → `Valid` (green)
- `❌ Invalid` → `Invalid` (red)
- `✅ Available` → `Available` (green)
- `❌ Not available` → `Not available` (red)

#### 3.3 Remove Command (`/internal/commands/rm.go`)
**Replace**:
- `✅ Removed dotfile` → `Removed dotfile` (color "Removed")
- `⏭️ Skipped` → `Skipped` (dim)

### Phase 4: Testing and Validation

#### 4.1 Verify No Emojis Remain
```bash
# Search for any remaining emoji characters
grep -r "[✅❌⚠️💡ℹ️🔍📦🔧⏭️]" internal/
```

#### 4.2 Test Color Output
- Run each command and verify colors appear correctly
- Test with `NO_COLOR=1` environment variable
- Test output piped to file (should have no color codes)

### Implementation Order
1. Create color utilities
2. Update high-visibility commands first: status, doctor, setup
3. Update remaining commands
4. Remove unused emoji constants from `output_utils.go`
5. Final sweep for any missed emojis

### Success Criteria
- Zero emoji characters in codebase
- Consistent color usage across all commands
- Status indicators are colored appropriately
- Professional, clean output
- No regression in functionality

### Notes
- Use existing `github.com/fatih/color` package
- Don't add new dependencies
- Keep changes minimal - only replace emojis and add color
- Don't restructure output format or add new features
