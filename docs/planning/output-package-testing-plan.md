# Output Package Testing Plan

**Date**: 2025-08-03
**Author**: Analysis by Claude
**Current Coverage**: 0% (no test files exist)
**Target Coverage**: 80-90%
**Estimated Effort**: 3 days

## Executive Summary

The `internal/output` package handles all user-facing output for plonk, including colored text, progress updates, and formatted results. Currently, it has 0% test coverage with no test files. This document provides a plan to achieve 80-90% coverage by introducing a minimal abstraction layer that maintains backward compatibility while enabling comprehensive testing.

## Package Overview

### Purpose
The output package provides:
- Colored terminal output with automatic detection
- Progress indicators for long-running operations
- Formatted output for various command results
- Consistent status messaging across the application

### Current Structure
```
internal/output/
├── colors.go      # Color initialization and status words (60 lines)
├── progress.go    # Progress and stage updates (23 lines)
└── formatters.go  # Output formatting and data structures (234 lines)
```

### Key Functions

#### colors.go
- `InitColors()` - Detects terminal and initializes color support
- `colorize()` - Internal helper for applying colors
- Status words: `Available()`, `Missing()`, `Drifted()`, etc.
- Color helpers: `ColorError()`, `ColorInfo()`

#### progress.go
- `ProgressUpdate(current, total, operation, item)` - Shows [1/5] style progress
- `StageUpdate(stage)` - Shows stage messages

#### formatters.go
- Data structures for various output types
- `TableOutput()` methods for human-readable formatting
- Helper functions for data conversion
- Complex conditional formatting logic

## Current Implementation Analysis

### Direct Dependencies
```go
// Direct stdout printing
fmt.Printf("[%d/%d] %s: %s\n", current, total, operation, item)

// Terminal detection
isatty.IsTerminal(os.Stdout.Fd())

// Color library
color.New(attrs...).Sprint(text)
```

### Usage Patterns
The package is used extensively throughout plonk:
- **17 files** import the output package
- Commands use it for user feedback
- Orchestrator uses it for progress during apply
- Clone package uses it for setup stages
- Diagnostics uses colored status words

### Testing Challenges

1. **Direct I/O** - All functions print directly to stdout
2. **No abstraction** - No way to intercept or verify output
3. **Terminal detection** - Hard-coded to check real terminal
4. **External behavior** - Color library modifies global state
5. **Complex formatting** - Many conditional branches in formatters

## Proposed Solution: Output Writer Interface

### Design Principles

1. **Minimal changes** - Add abstraction without breaking existing code
2. **Backward compatible** - All current usage continues to work
3. **Testable** - Enable comprehensive testing of all output
4. **Consistent** - Follow patterns already in the codebase

### Architecture

```
┌─────────────────────────────────────────────┐
│         Current Callers (unchanged)          │
│                                              │
│  output.ProgressUpdate(1, 5, "Installing")  │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│          Output Package                      │
│                                              │
│  writer: Writer = &StdoutWriter{}           │
│                                              │
│  func ProgressUpdate(...) {                 │
│      writer.Printf(...)                     │
│  }                                          │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│     Writer Interface (new)                   │
│                                              │
│  Printf(format string, args ...interface{}) │
│  IsTerminal() bool                          │
└─────────────────────────────────────────────┘
         │                    │
         ▼                    ▼
┌──────────────────┐ ┌──────────────────────┐
│  StdoutWriter    │ │  BufferWriter (test) │
│  (production)    │ │  (captures output)   │
└──────────────────┘ └──────────────────────┘
```

## Implementation Plan

### Phase 1: Create Writer Abstraction (Day 1 Morning)

**1. Add writer.go file:**
```go
// internal/output/writer.go
package output

import (
    "fmt"
    "os"
    "github.com/mattn/go-isatty"
)

// Writer abstracts output operations for testing
type Writer interface {
    Printf(format string, args ...interface{})
    IsTerminal() bool
}

// StdoutWriter implements Writer for real stdout
type StdoutWriter struct{}

func (s *StdoutWriter) Printf(format string, args ...interface{}) {
    fmt.Printf(format, args...)
}

func (s *StdoutWriter) IsTerminal() bool {
    return isatty.IsTerminal(os.Stdout.Fd()) ||
           isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// Package-level writer instance
var writer Writer = &StdoutWriter{}

// SetWriter allows tests to override the writer
func SetWriter(w Writer) {
    writer = w
}

// GetWriter returns the current writer (useful for tests)
func GetWriter() Writer {
    return writer
}
```

**2. Create test writer:**
```go
// internal/output/test_writer.go
package output

import (
    "bytes"
    "fmt"
)

// BufferWriter captures output for testing
type BufferWriter struct {
    buf      bytes.Buffer
    terminal bool
}

// NewBufferWriter creates a test writer
func NewBufferWriter(terminal bool) *BufferWriter {
    return &BufferWriter{terminal: terminal}
}

func (b *BufferWriter) Printf(format string, args ...interface{}) {
    fmt.Fprintf(&b.buf, format, args...)
}

func (b *BufferWriter) IsTerminal() bool {
    return b.terminal
}

func (b *BufferWriter) String() string {
    return b.buf.String()
}

func (b *BufferWriter) Reset() {
    b.buf.Reset()
}
```

### Phase 2: Update Existing Functions (Day 1 Afternoon)

**1. Update colors.go:**
```go
// Modify InitColors to use writer
func InitColors() {
    // Use writer to check terminal
    isTerminal := writer.IsTerminal()

    if !isTerminal {
        color.NoColor = true
    }
}
```

**2. Update progress.go:**
```go
// Replace fmt.Printf with writer.Printf
func ProgressUpdate(current, total int, operation, item string) {
    if total > 1 {
        writer.Printf("[%d/%d] %s: %s\n", current, total, operation, item)
    } else if total == 1 {
        writer.Printf("%s: %s\n", operation, item)
    }
}

func StageUpdate(stage string) {
    writer.Printf("%s\n", stage)
}
```

**3. Note on formatters.go:**
- The formatter functions return strings, not print directly
- They don't need writer updates
- Focus testing on the complex logic

### Phase 3: Implement Comprehensive Tests (Days 2-3)

**1. Create colors_test.go:**
```go
package output

import (
    "testing"
    "github.com/fatih/color"
)

func TestInitColors(t *testing.T) {
    // Save original state
    originalWriter := writer
    originalNoColor := color.NoColor
    defer func() {
        writer = originalWriter
        color.NoColor = originalNoColor
    }()

    tests := []struct {
        name         string
        isTerminal   bool
        wantNoColor  bool
    }{
        {
            name:        "terminal output enables colors",
            isTerminal:  true,
            wantNoColor: false,
        },
        {
            name:        "non-terminal output disables colors",
            isTerminal:  false,
            wantNoColor: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set up test writer
            writer = NewBufferWriter(tt.isTerminal)
            color.NoColor = false // Reset state

            InitColors()

            if color.NoColor != tt.wantNoColor {
                t.Errorf("color.NoColor = %v, want %v",
                    color.NoColor, tt.wantNoColor)
            }
        })
    }
}

func TestStatusWords(t *testing.T) {
    // Test with colors enabled
    color.NoColor = false

    tests := []struct {
        name     string
        fn       func() string
        wantText string
        wantColor bool
    }{
        {
            name:      "Available shows green",
            fn:        Available,
            wantText:  "available",
            wantColor: true,
        },
        {
            name:      "Missing shows red",
            fn:        Missing,
            wantText:  "missing",
            wantColor: true,
        },
        {
            name:      "Drifted shows yellow",
            fn:        Drifted,
            wantText:  "drifted",
            wantColor: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.fn()

            // Check text content
            if !strings.Contains(result, tt.wantText) {
                t.Errorf("result %q does not contain %q",
                    result, tt.wantText)
            }

            // With colors enabled, output should have ANSI codes
            if tt.wantColor && !strings.Contains(result, "\033[") {
                t.Error("expected colored output but got plain text")
            }
        })
    }
}
```

**2. Create progress_test.go:**
```go
package output

import (
    "testing"
)

func TestProgressUpdate(t *testing.T) {
    // Save and restore original writer
    originalWriter := writer
    defer func() { writer = originalWriter }()

    tests := []struct {
        name      string
        current   int
        total     int
        operation string
        item      string
        want      string
    }{
        {
            name:      "single item shows simple format",
            current:   1,
            total:     1,
            operation: "Installing",
            item:      "vim",
            want:      "Installing: vim\n",
        },
        {
            name:      "multiple items shows progress",
            current:   2,
            total:     5,
            operation: "Installing",
            item:      "git",
            want:      "[2/5] Installing: git\n",
        },
        {
            name:      "zero total shows nothing",
            current:   1,
            total:     0,
            operation: "Installing",
            item:      "test",
            want:      "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            buf := NewBufferWriter(true)
            writer = buf

            ProgressUpdate(tt.current, tt.total, tt.operation, tt.item)

            if got := buf.String(); got != tt.want {
                t.Errorf("ProgressUpdate() output = %q, want %q",
                    got, tt.want)
            }
        })
    }
}

func TestStageUpdate(t *testing.T) {
    originalWriter := writer
    defer func() { writer = originalWriter }()

    buf := NewBufferWriter(true)
    writer = buf

    StageUpdate("Cloning repository...")

    want := "Cloning repository...\n"
    if got := buf.String(); got != want {
        t.Errorf("StageUpdate() = %q, want %q", got, want)
    }
}
```

**3. Create formatters_test.go:**
```go
package output

import (
    "testing"
    "strings"
)

func TestDotfileAddOutput_TableOutput(t *testing.T) {
    tests := []struct {
        name     string
        output   DotfileAddOutput
        wantStrs []string
    }{
        {
            name: "successful add",
            output: DotfileAddOutput{
                Source:      "~/.config/plonk/.vimrc",
                Destination: "~/.vimrc",
                Action:      "added",
                Path:        "/home/user/.vimrc",
            },
            wantStrs: []string{
                "Added dotfile to plonk configuration",
                "Source: ~/.config/plonk/.vimrc",
                "Destination: ~/.vimrc",
                "Original: /home/user/.vimrc",
                "has been copied to your plonk config",
            },
        },
        {
            name: "dry run would add",
            output: DotfileAddOutput{
                Source:      "~/.config/plonk/.bashrc",
                Destination: "~/.bashrc",
                Action:      "would-add",
                Path:        "/home/user/.bashrc",
            },
            wantStrs: []string{
                "Would add dotfile",
                "(dry-run)",
            },
        },
        {
            name: "failed action",
            output: DotfileAddOutput{
                Action: "failed",
                Path:   "/home/user/.config/test",
                Error:  "permission denied",
            },
            wantStrs: []string{
                "✗",
                "permission denied",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.output.TableOutput()

            for _, want := range tt.wantStrs {
                if !strings.Contains(result, want) {
                    t.Errorf("TableOutput() missing %q in:\n%s",
                        want, result)
                }
            }
        })
    }
}

func TestMapStatusToAction(t *testing.T) {
    tests := []struct {
        status string
        want   string
    }{
        {"added", "added"},
        {"updated", "updated"},
        {"would-add", "would-add"},
        {"would-update", "would-update"},
        {"unknown", "failed"},
        {"error", "failed"},
    }

    for _, tt := range tests {
        t.Run(tt.status, func(t *testing.T) {
            if got := MapStatusToAction(tt.status); got != tt.want {
                t.Errorf("MapStatusToAction(%q) = %q, want %q",
                    tt.status, got, tt.want)
            }
        })
    }
}
```

## Test Coverage Strategy

### Unit Tests (Target: 90%)

1. **colors.go (100% coverage)**
   - Terminal detection scenarios
   - Each status word function
   - Color enabled/disabled states

2. **progress.go (100% coverage)**
   - All branches in ProgressUpdate
   - StageUpdate output
   - Edge cases (0 total, negative numbers)

3. **formatters.go (80% coverage)**
   - All TableOutput methods
   - Helper function logic
   - Error message extraction
   - Data conversion functions

### Integration Considerations

- Test with NO_COLOR environment variable
- Test with different terminal types
- Verify actual color codes in output
- Test formatting with various data sizes

## Implementation Guidelines

### Do's
- ✅ Preserve exact current behavior
- ✅ Use defer to restore package state in tests
- ✅ Test both terminal and non-terminal modes
- ✅ Cover all conditional branches
- ✅ Use table-driven tests

### Don'ts
- ❌ Don't change any public APIs
- ❌ Don't modify output format
- ❌ Don't add unnecessary complexity
- ❌ Don't test third-party library internals

## Verification Steps

1. **All tests pass:**
   ```bash
   go test ./internal/output -v
   ```

2. **Coverage meets target:**
   ```bash
   go test ./internal/output -cover
   # Should show 80%+ coverage
   ```

3. **No behavior changes:**
   ```bash
   # Run plonk commands and verify output unchanged
   plonk status
   plonk apply --dry-run
   ```

4. **Colors work correctly:**
   ```bash
   # Terminal should show colors
   plonk doctor

   # Pipe should not show colors
   plonk doctor | cat
   ```

## Success Criteria

- [ ] Writer interface implemented
- [ ] All functions use writer instead of fmt.Printf
- [ ] Tests achieve 80%+ coverage
- [ ] No changes to external behavior
- [ ] Terminal detection works in tests
- [ ] All existing commands work unchanged

## Alternative Approaches Considered

### Why Not Return Strings?
- Would require updating all 17 files that use the package
- Breaking change for all callers
- More significant refactoring effort

### Why Not Just Mock fmt.Printf?
- Go doesn't support mocking package functions
- Would require build-time tricks or code generation
- Less idiomatic Go

### Why Not Use Dependency Injection?
- Would require changing function signatures
- Breaking change for existing code
- Over-engineering for this use case

## Maintenance Notes

### Adding New Output Functions
1. Use `writer.Printf()` instead of `fmt.Printf()`
2. Add corresponding tests
3. Consider both terminal and non-terminal cases

### Debugging Output Issues
- Use `GetWriter()` to inspect current writer
- Check `color.NoColor` value
- Verify terminal detection with `writer.IsTerminal()`

## Conclusion

This plan provides a minimal, backward-compatible approach to making the output package testable. By introducing a simple Writer interface, we can achieve 80-90% test coverage while maintaining all existing behavior. The implementation follows patterns already established in the codebase and requires no changes to existing callers.
