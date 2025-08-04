# Plan: Redirect Progress/Status Messages to stderr

**Created**: 2025-08-04
**Status**: COMPLETED
**Objective**: Redirect all progress and status messages to stderr to ensure clean JSON/YAML output on stdout

## Background

Current issue: Progress messages contaminate JSON/YAML output because everything goes to stdout. Industry standard (kubectl, docker, gh) is to send progress/status to stderr, structured data to stdout.

## Current State Summary

### Output Infrastructure
- **Writer Interface**: Clean abstraction in `internal/output/writer.go`
- **Progress Functions**: `ProgressUpdate()` and `StageUpdate()` in `progress.go`
- **Current Flow**: All output → stdout via `writer` variable
- **Existing stderr**: Used only for errors/warnings in 4 locations

### Key Components
1. `writer` package variable (currently StdoutWriter)
2. Progress/status functions that use `writer.Printf()`
3. Color system that checks stdout terminal status
4. Test infrastructure with BufferWriter

## Proposed Solution: Dual Writer System

### Phase 1: Infrastructure Setup

#### 1.1 Create Progress Writer
```go
// internal/output/progress_writer.go
package output

import (
    "fmt"
    "os"
    "github.com/mattn/go-isatty"
)

// progressWriter is used for all progress/status messages
var progressWriter Writer = &StderrWriter{}

type StderrWriter struct{}

func (s *StderrWriter) Printf(format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, format, args...)
}

func (s *StderrWriter) IsTerminal() bool {
    return isatty.IsTerminal(os.Stderr.Fd()) ||
           isatty.IsCygwinTerminal(os.Stderr.Fd())
}
```

#### 1.2 Update Progress Functions
```go
// internal/output/progress.go
func ProgressUpdate(current, total int, operation, item string) {
    // Use progressWriter instead of writer
    if total > 1 {
        progressWriter.Printf("[%d/%d] %s: %s\n", current, total, operation, item)
    } else if total == 1 {
        progressWriter.Printf("%s: %s\n", operation, item)
    }
}

func StageUpdate(stage string) {
    progressWriter.Printf("%s\n", stage)
}
```

### Phase 2: Direct Printf Replacements

#### 2.1 Create Printf Wrapper
```go
// internal/output/print.go
package output

import "fmt"

// Printf writes formatted output to the appropriate stream
// Use for status/progress messages that should go to stderr
func Printf(format string, args ...interface{}) {
    progressWriter.Printf(format, args...)
}

// Println writes output with newline to the appropriate stream
func Println(args ...interface{}) {
    progressWriter.Printf("%s\n", fmt.Sprint(args...))
}
```

#### 2.2 Replace Direct fmt.Printf Calls

**Files with status/progress printf calls to update:**
- `internal/commands/install.go`: Lines 106, 109 (status icons)
- `internal/commands/uninstall.go`: Lines 103, 106 (status icons)
- `internal/clone/setup.go`: Multiple lines (setup progress)
- `internal/clone/tools.go`: Lines 46-51 (tool checks)

**Pattern to replace:**
```go
// Before
fmt.Printf("%s %s %s\n", icon, result.Status, result.Name)

// After
output.Printf("%s %s %s\n", icon, result.Status, result.Name)
```

### Phase 3: Color System Updates

#### 3.1 Dual Terminal Detection
```go
// internal/output/colors.go
func InitColors() {
    // Check both stdout and stderr for terminal
    stdoutIsTerminal := writer.IsTerminal()
    stderrIsTerminal := progressWriter.IsTerminal()

    // Disable colors if neither is a terminal
    if !stdoutIsTerminal && !stderrIsTerminal {
        color.NoColor = true
    }
}
```

### Phase 4: Testing Infrastructure

#### 4.1 Enhanced Test Writer
```go
// internal/testutil/writer.go
type DualBufferWriter struct {
    StdoutBuffer bytes.Buffer
    StderrBuffer bytes.Buffer
    isTerminal   bool
}

// Methods for stdout writer
func (w *DualBufferWriter) Printf(format string, args ...interface{}) {
    fmt.Fprintf(&w.StdoutBuffer, format, args...)
}

// Methods for stderr writer
func (w *DualBufferWriter) StderrPrintf(format string, args ...interface{}) {
    fmt.Fprintf(&w.StderrBuffer, format, args...)
}
```

#### 4.2 Update Tests
- Progress tests to verify stderr output
- Integration tests to verify clean JSON/YAML

## Implementation Order

### Step 1: Core Infrastructure (Day 1)
1. Create `progress_writer.go` with StderrWriter
2. Update `progress.go` to use progressWriter
3. Create basic tests

### Step 2: Command Updates (Day 2)
1. Create Printf/Println wrappers
2. Update install/uninstall commands
3. Test JSON output is clean

### Step 3: Clone/Setup Updates (Day 3)
1. Update all clone setup progress messages
2. Update tool check messages
3. Comprehensive testing

### Step 4: Testing & Documentation (Day 4)
1. Update test infrastructure
2. Add integration tests
3. Update documentation

## Files to Modify

### Core Files (Phase 1)
- [x] Create `internal/output/progress_writer.go`
- [x] Update `internal/output/progress.go`
- [x] Update `internal/output/colors.go`
- [x] Create `internal/output/print.go` with Printf/Println wrappers

### Command Files (Phase 2)
- [x] `internal/commands/install.go`
- [x] `internal/commands/uninstall.go`
- [x] `internal/commands/config_edit.go`
- [x] `internal/commands/diff.go`
- [x] Other commands checked - no updates needed

### Clone/Setup Files (Phase 3)
- [x] `internal/clone/setup.go`
- [x] `internal/clone/tools.go`
- [x] `internal/clone/prompts.go`

### Test Files (Phase 4)
- [x] Updated existing tests to use progressWriter
- [x] Integration test validates clean JSON output

## Success Criteria

1. ✅ `plonk install -o json package` produces clean JSON on stdout
2. ✅ Progress messages appear on stderr during execution
3. ✅ Piping works correctly: `plonk status -o json | jq .`
4. ✅ No regression in table format output
5. ✅ Colors work correctly on both streams
6. ✅ All tests pass

## Implementation Results (2025-08-04)

Successfully implemented the dual writer system:
- All progress/status messages now go to stderr
- JSON/YAML output on stdout is clean
- Integration tests can parse JSON output reliably
- No breaking changes to existing functionality
- Colors work correctly for both stdout and stderr

## Risks & Mitigations

### Risk 1: Breaking Changes
**Mitigation**: Keep existing interfaces, add new ones alongside

### Risk 2: Missing Printf Calls
**Mitigation**: Grep for all `fmt.Printf` and `fmt.Println` calls

### Risk 3: Test Failures
**Mitigation**: Update tests incrementally as we go

### Risk 4: Color Confusion
**Mitigation**: Test thoroughly with different terminal configurations

## Validation Plan

### Manual Testing
```bash
# Test JSON is clean
plonk install -o json brew:jq 2>/dev/null | jq .

# Test progress still shows
plonk install brew:jq

# Test piping works
plonk status -o json | jq '.packages | length'

# Test stderr redirection
plonk install brew:jq 2>progress.log
cat progress.log  # Should contain progress messages
```

### Automated Testing
- Unit tests for new writer
- Integration tests for JSON output
- Progress message tests
