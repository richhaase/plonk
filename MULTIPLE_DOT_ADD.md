# Multiple Dotfile Add Implementation Plan

## Overview

This document provides a detailed implementation plan for adding multiple dotfile support to the `plonk dot add` command. The enhancement allows users to add multiple dotfiles and directories in a single command while maintaining full backward compatibility.

## Goals

- Enable `plonk dot add ~/.vimrc ~/.zshrc ~/.gitconfig`
- Support glob patterns: `plonk dot add ~/.config/nvim/* ~/.ssh/config`
- Support mixed files and directories: `plonk dot add ~/.vimrc ~/.config/nvim/ ~/.ssh/`
- Maintain backward compatibility with single dotfile usage
- Provide clear progress indication and error handling
- Support dry-run mode for multiple dotfiles

## Current Implementation Analysis

### Current Command Structure
```go
// internal/commands/dot_add.go (current)
func runDotAdd(cmd *cobra.Command, args []string) error {
    // Currently expects exactly 1 argument
    if len(args) != 1 {
        return errors.NewError(...)
    }

    dotfilePath := args[0]
    // Process single dotfile...
}
```

### Current Add Logic Flow
1. Validate single dotfile argument
2. Resolve and validate path
3. Check if dotfile already managed
4. Copy file to plonk config directory
5. Show result

## Implementation Plan

**⚠️ IMPORTANT: See CONTEXT.md for pre-work requirements and shared utilities that must be implemented first.**

### Phase 0: Pre-work (✅ COMPLETED)

The following pre-work has been completed:

1. **✅ Create shared utilities in `internal/operations/`**
   - ✅ Common result types and progress reporting interfaces (`types.go`)
   - ✅ Error suggestion formatting utilities (`progress.go`)
   - ✅ Summary display logic and context management (`context.go`)
   - ✅ Comprehensive test coverage (`types_test.go`)

2. **✅ Extend error system**
   - ✅ Added suggestion support to PlonkError type
   - ✅ Created helper methods: `WithSuggestion`, `WithSuggestionCommand`, `WithSuggestionMessage`
   - ✅ Updated `UserMessage()` to include suggestions

**Note:** Dotfile implementation does not require PackageManager interface changes.

### Phase 1: Core Multiple Dotfile Support (✅ COMPLETED)

**Implementation Summary:**
- ✅ Updated command interface to accept multiple dotfile arguments
- ✅ Implemented sequential dotfile processing with shared utilities
- ✅ Added glob expansion and file discovery for directories
- ✅ Added file attribute preservation (permissions, timestamps)
- ✅ Added progress reporting for each dotfile
- ✅ Enhanced error handling with contextual suggestions
- ✅ Added dry-run support for preview mode
- ✅ Comprehensive test coverage with filesystem-based testing
- ✅ **RESOLVED**: Dotfile management detection working correctly

### Phase 2: Implementation Complete (✅ COMPLETED)

**Final Implementation Status:**
- ✅ All core functionality implemented in `internal/commands/dot_add.go`
- ✅ Multiple dotfile processing with `addDotfiles()` function
- ✅ Single file processing with `addSingleFileNew()`
- ✅ Directory processing with `addDirectoryFilesNew()`
- ✅ File attribute preservation with `copyFileWithAttributes()`
- ✅ Path resolution with `resolveDotfilePath()`
- ✅ Source/destination mapping with `generatePaths()`
- ✅ Progress reporting using shared operations utilities
- ✅ Structured output support for both single and batch operations
- ✅ Error handling with continue-on-failure approach
- ✅ All tests passing including filesystem-based dotfile detection
- ✅ PLONK_DIR environment variable handling for test isolation

**Key Finding - Dotfile Management Detection (✅ VERIFIED):**
Investigation of `GetDotfileTargets()` confirms filesystem-based behavior:

**How Plonk Determines "Already Managed" Dotfiles:**
1. **Filesystem-Based Discovery**: Uses `filepath.Walk()` to scan `$PLONK_DIR`
2. **Auto-Discovery**: No manual configuration - files in config dir = managed dotfiles
3. **Convention-Based Mapping**:
   - Source: `zshrc` → Target: `~/.zshrc`
   - Source: `config/nvim/init.lua` → Target: `~/.config/nvim/init.lua`
4. **Ignore Pattern Filtering**: Skips `.DS_Store`, `.git`, `*.tmp`, etc.
5. **Zero-Config Philosophy**: Works immediately without explicit configuration

**Test Fix Required:**
- "Already managed" = source file physically exists in `$PLONK_DIR` filesystem
- Tests must create actual files in config directory to simulate existing dotfiles
- No need to modify config objects - pure filesystem detection

**Implementation Approach:**
- Using `config.GetDotfileTargets()` to detect existing managed dotfiles
- Sequential processing with immediate progress reporting
- File attribute preservation during copy operations
- Continue-on-failure with comprehensive error reporting

#### 1.1 Update Command Arguments Handling (✅ COMPLETED)

**File:** `internal/commands/dot_add.go`

```go
func runDotAdd(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
        return errors.NewError(errors.ErrInvalidInput, errors.DomainCommands,
            "add", "at least one dotfile path required")
    }

    // Handle both single and multiple dotfiles with same logic
    return addDotfiles(cmd, args)
}

func addDotfiles(cmd *cobra.Command, dotfilePaths []string) error {
    // New function to handle both single and multiple dotfiles
}
```

#### 1.2 Path Expansion (Discover Issues During Processing)

```go
func expandGlobPatterns(dotfilePaths []string) ([]DotfileEntry, error) {
    var entries []DotfileEntry

    for _, path := range dotfilePaths {
        // Handle tilde expansion first
        expandedPath := path
        if strings.HasPrefix(path, "~/") {
            homeDir, err := os.UserHomeDir()
            if err != nil {
                return nil, errors.Wrap(err, errors.ErrInvalidInput,
                    errors.DomainCommands, "expand-home",
                    "failed to get home directory")
            }
            expandedPath = filepath.Join(homeDir, path[2:])
        }

        // Expand globs
        matches, err := filepath.Glob(expandedPath)
        if err != nil {
            return nil, errors.Wrap(err, errors.ErrInvalidInput,
                errors.DomainCommands, "expand-glob",
                fmt.Sprintf("invalid glob pattern: %s", path))
        }

        if len(matches) == 0 {
            // No matches - add as literal path for validation during processing
            matches = []string{expandedPath}
        }

        for _, matchedPath := range matches {
            entry := DotfileEntry{
                SourcePath:   path,        // Original user input for error reporting
                ResolvedPath: matchedPath, // Expanded path
                TargetPath:   "", // Will be determined during processing
            }
            entries = append(entries, entry)
        }
    }

    return entries, nil
}

type DotfileEntry struct {
    SourcePath   string // Original path specified by user (for error reporting)
    ResolvedPath string // Absolute path after expansion
    TargetPath   string // Path in plonk config directory
    IsDirectory  bool   // Determined during processing
    IsFile       bool   // Determined during processing
}
```

#### 1.3 Implement Multiple Dotfile Processing

```go
// NOTE: This will be replaced by shared OperationResult type from internal/operations/
type DotfileAddResult struct {
    Entry           DotfileEntry
    Status          string // "added", "updated", "skipped", "failed"
    Error           error
    FilesProcessed  int    // For directories
}

func addDotfiles(cmd *cobra.Command, dotfilePaths []string) error {
    // Expand glob patterns only
    entries, err := expandGlobPatterns(dotfilePaths)
    if err != nil {
        return err
    }

    var allResults []DotfileAddResult

    // Load configuration once
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(),
        time.Duration(cfg.DotfileTimeout)*time.Second)
    defer cancel()

    // Process each dotfile entry (discover issues during processing)
    for _, entry := range entries {
        // Check for cancellation
        if ctx.Err() != nil {
            return errors.Wrap(ctx.Err(), errors.ErrInternal,
                errors.DomainDotfiles, "add-multiple",
                "operation cancelled or timed out")
        }

        results := processEntryWithIndividualProgress(ctx, cfg, entry)
        allResults = append(allResults, results...)
    }

    // Show summary
    showAddSummary(allResults)

    // Determine exit code
    return determineExitCode(allResults)
}
```

#### 1.4 Individual File Processing with Progress

```go
func processEntryWithIndividualProgress(ctx context.Context, cfg *config.Config, entry DotfileEntry) []DotfileAddResult {
    // Discover file type and existence during processing
    stat, err := os.Stat(entry.ResolvedPath)
    if err != nil {
        result := DotfileAddResult{
            Entry: entry,
            Status: "failed",
            Error: errors.Wrap(err, errors.ErrFileNotFound,
                errors.DomainDotfiles, "add",
                fmt.Sprintf("file not found: %s", entry.SourcePath)),
        }
        showDotfileProgress(result)
        return []DotfileAddResult{result}
    }

    entry.IsDirectory = stat.IsDir()
    entry.IsFile = !stat.IsDir()

    if entry.IsDirectory {
        return processDirectoryWithIndividualFiles(ctx, cfg, entry)
    } else {
        result := processSingleFile(ctx, cfg, entry)
        showDotfileProgress(result)
        return []DotfileAddResult{result}
    }
}

func processDirectoryWithIndividualFiles(ctx context.Context, cfg *config.Config, dirEntry DotfileEntry) []DotfileAddResult {
    var results []DotfileAddResult

    // Walk through directory and process each file individually
    err := filepath.Walk(dirEntry.ResolvedPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip directories themselves, only process files
        if info.IsDir() {
            return nil
        }

        // Check for cancellation
        if ctx.Err() != nil {
            return ctx.Err()
        }

        // Create relative path for target
        relPath, err := filepath.Rel(dirEntry.ResolvedPath, path)
        if err != nil {
            return err
        }

        // Calculate target path in plonk config
        targetPath := calculateTargetPath(dirEntry.ResolvedPath, dirEntry.SourcePath)
        fileTargetPath := filepath.Join(targetPath, relPath)

        fileEntry := DotfileEntry{
            SourcePath:   fmt.Sprintf("%s/%s", dirEntry.SourcePath, relPath),
            ResolvedPath: path,
            TargetPath:   fileTargetPath,
            IsFile:       true,
            IsDirectory:  false,
        }

        result := processSingleFile(ctx, cfg, fileEntry)
        results = append(results, result)
        showDotfileProgress(result)

        return nil
    })

    if err != nil {
        result := DotfileAddResult{
            Entry: dirEntry,
            Status: "failed",
            Error: errors.Wrap(err, errors.ErrFileIO,
                errors.DomainDotfiles, "walk-directory",
                fmt.Sprintf("failed to process directory: %s", dirEntry.SourcePath)),
        }
        showDotfileProgress(result)
        return []DotfileAddResult{result}
    }

    return results
}

func processSingleFile(ctx context.Context, cfg *config.Config, entry DotfileEntry) DotfileAddResult {
    result := DotfileAddResult{Entry: entry}

    // Calculate target path if not already set
    if entry.TargetPath == "" {
        entry.TargetPath = calculateTargetPath(entry.ResolvedPath, entry.SourcePath)
    }

    // Check if already managed (default behavior is to update)
    alreadyManaged := isAlreadyManaged(cfg, entry.TargetPath)

    // Dry run check
    if dryRun {
        if alreadyManaged {
            result.Status = "would-update"
        } else {
            result.Status = "would-add"
        }
        return result
    }

    // Copy file with permission/timestamp preservation
    err := copyFilePreservingAttributes(entry.ResolvedPath, entry.TargetPath)
    if err != nil {
        result.Status = "failed"
        result.Error = err
        return result
    }

    if alreadyManaged {
        result.Status = "updated"
    } else {
        result.Status = "added"
    }
    result.FilesProcessed = 1
    return result
}

func copyFilePreservingAttributes(src, dst string) error {
    // Ensure target directory exists
    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return errors.Wrap(err, errors.ErrFileIO,
            errors.DomainDotfiles, "create-directory",
            "failed to create target directory")
    }

    // Get source file info for preservation
    srcInfo, err := os.Stat(src)
    if err != nil {
        return errors.Wrap(err, errors.ErrFileNotFound,
            errors.DomainDotfiles, "stat-source",
            "failed to get source file info")
    }

    // Copy file content
    srcFile, err := os.Open(src)
    if err != nil {
        return errors.Wrap(err, errors.ErrFileIO,
            errors.DomainDotfiles, "open-source",
            "failed to open source file")
    }
    defer srcFile.Close()

    dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
    if err != nil {
        return errors.Wrap(err, errors.ErrFileIO,
            errors.DomainDotfiles, "create-target",
            "failed to create target file")
    }
    defer dstFile.Close()

    _, err = io.Copy(dstFile, srcFile)
    if err != nil {
        return errors.Wrap(err, errors.ErrFileIO,
            errors.DomainDotfiles, "copy-content",
            "failed to copy file content")
    }

    // Preserve timestamps
    err = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
    if err != nil {
        // Log warning but don't fail - timestamp preservation is best effort
        log.Printf("Warning: failed to preserve timestamps for %s: %v", dst, err)
    }

    return nil
}
```

### Phase 2: Output and User Experience

#### 2.1 Progress Indication (Individual File Progress)

```go
func showDotfileProgress(result DotfileAddResult) {
    entry := result.Entry
    sourcePath := entry.SourcePath
    targetPath := getRelativeTargetPath(entry.TargetPath) // Show relative to config dir

    switch result.Status {
    case "added":
        fmt.Printf("✓ %s → %s\n", sourcePath, targetPath)
    case "updated":
        fmt.Printf("↻ %s → %s (updated)\n", sourcePath, targetPath)
    case "failed":
        fmt.Printf("✗ %s - %s\n", sourcePath, formatErrorWithSuggestion(result.Error, sourcePath))
    case "would-add":
        fmt.Printf("+ %s → %s (would add)\n", sourcePath, targetPath)
    case "would-update":
        fmt.Printf("+ %s → %s (would update)\n", sourcePath, targetPath)
    }
}

func getRelativeTargetPath(fullPath string) string {
    // Convert full target path to relative path from config directory
    // e.g., "/home/user/.config/plonk/config/nvim/init.lua" → "config/nvim/init.lua"
    configDir := getConfigDir()
    if rel, err := filepath.Rel(configDir, fullPath); err == nil {
        return rel
    }
    return fullPath
}

// NOTE: This will be replaced by shared utility from internal/operations/
func formatErrorWithSuggestion(err error, sourcePath string) string {
    // This function will be replaced by operations.FormatErrorWithSuggestion
    // See CONTEXT.md for the shared implementation
    msg := err.Error()

    // Add suggestions based on error type
    if strings.Contains(msg, "not found") {
        return fmt.Sprintf("%s\n     Check if path exists: ls -la %s", msg, sourcePath)
    }
    if strings.Contains(msg, "permission") {
        return fmt.Sprintf("%s\n     Try: chmod +r %s", msg, sourcePath)
    }
    if strings.Contains(msg, "already exists") {
        return fmt.Sprintf("%s\n     Use --force to overwrite", msg)
    }

    return msg
}
```

#### 2.2 Summary Output

```go
func showAddSummary(results []DotfileAddResult) {
    added := countByStatus(results, "added")
    updated := countByStatus(results, "updated")
    skipped := countByStatus(results, "skipped")
    failed := countByStatus(results, "failed")

    totalFiles := 0
    for _, result := range results {
        totalFiles += result.FilesProcessed
    }

    fmt.Printf("\nSummary: %d added, %d updated, %d skipped, %d failed (%d total files)\n",
        added, updated, skipped, failed, totalFiles)

    // Show failed dotfiles with suggestions
    if failed > 0 {
        fmt.Println("\nFailed dotfiles:")
        for _, result := range results {
            if result.Status == "failed" {
                fmt.Printf("  %s: %v\n", result.Entry.SourcePath, result.Error)
            }
        }
    }
}
```

#### 2.3 Structured Output Support

```go
type MultipleDotAddOutput struct {
    Summary struct {
        Total     int `json:"total"`
        Added     int `json:"added"`
        Updated   int `json:"updated"`
        Skipped   int `json:"skipped"`
        Failed    int `json:"failed"`
        TotalFiles int `json:"total_files"`
    } `json:"summary"`
    Results []DotfileAddResult `json:"results"`
}
```

### Phase 3: Error Handling and Edge Cases

#### 3.1 Error Handling Strategy

**Continue on Error Approach (Sequential Processing):**
- Process dotfiles one at a time in order specified
- Report success or failure immediately after each dotfile
- Continue processing remaining dotfiles even if some fail
- Exit code 0 if any dotfiles succeeded
- Exit code 1 only if all dotfiles failed

**Error Scenarios:**
- **Cancellation (Ctrl-C)**: Clean termination, partially copied directories left as-is
- **File not found**: Clear error with suggestion to check path
- **Permission errors**: Clear error with suggestion to check file permissions
- **Target already exists**: Option to skip or overwrite (future enhancement)
- **Disk space issues**: Clear error about available space
- **Invalid glob patterns**: Clear error about pattern syntax

#### 3.2 Context and Timeout Handling

```go
func addDotfiles(cmd *cobra.Command, dotfilePaths []string) error {
    // Create context with timeout for entire operation
    ctx, cancel := context.WithTimeout(context.Background(),
        time.Duration(cfg.DotfileTimeout)*time.Second)
    defer cancel()

    for _, entry := range entries {
        // Check if context cancelled
        if ctx.Err() != nil {
            return errors.Wrap(ctx.Err(), errors.ErrInternal,
                errors.DomainDotfiles, "add-multiple",
                "operation cancelled or timed out")
        }

        result := addSingleDotfileWithContext(ctx, cfg, entry)
        // ...
    }
}
```

#### 3.3 Glob Pattern Handling

```go
func expandGlobPatterns(patterns []string) ([]string, error) {
    var allPaths []string

    for _, pattern := range patterns {
        // Handle tilde expansion
        if strings.HasPrefix(pattern, "~/") {
            homeDir, err := os.UserHomeDir()
            if err != nil {
                return nil, errors.Wrap(err, errors.ErrInvalidInput,
                    errors.DomainCommands, "expand-home",
                    "failed to get home directory")
            }
            pattern = filepath.Join(homeDir, pattern[2:])
        }

        // Expand glob
        matches, err := filepath.Glob(pattern)
        if err != nil {
            return nil, errors.Wrap(err, errors.ErrInvalidInput,
                errors.DomainCommands, "expand-glob",
                fmt.Sprintf("invalid glob pattern: %s", pattern))
        }

        if len(matches) == 0 {
            // No matches - treat as literal path for later validation
            allPaths = append(allPaths, pattern)
        } else {
            allPaths = append(allPaths, matches...)
        }
    }

    return allPaths, nil
}
```

### Phase 4: Testing Strategy

#### 4.1 Unit Tests

**File:** `internal/commands/dot_add_test.go`

```go
func TestMultipleDotfileAdd(t *testing.T) {
    tests := []struct {
        name          string
        dotfiles      []string
        setupFiles    map[string]string // file path -> content
        expectedAdded int
        expectedFailed int
        expectError   bool
    }{
        {
            name:     "add multiple dotfiles successfully",
            dotfiles: []string{"~/.vimrc", "~/.zshrc", "~/.gitconfig"},
            setupFiles: map[string]string{
                ".vimrc":     "set number",
                ".zshrc":     "export PATH=$PATH:/usr/local/bin",
                ".gitconfig": "[user]\nname = Test User",
            },
            expectedAdded: 3,
            expectedFailed: 0,
            expectError: false,
        },
        {
            name:     "continue on partial failure",
            dotfiles: []string{"~/.vimrc", "~/.nonexistent", "~/.zshrc"},
            setupFiles: map[string]string{
                ".vimrc": "set number",
                ".zshrc": "export PATH=$PATH:/usr/local/bin",
            },
            expectedAdded: 2,
            expectedFailed: 1,
            expectError: false,
        },
        {
            name:     "handle glob patterns",
            dotfiles: []string{"~/.config/nvim/*"},
            setupFiles: map[string]string{
                ".config/nvim/init.lua":     "vim.cmd('set number')",
                ".config/nvim/lua/config.lua": "return {}",
            },
            expectedAdded: 2,
            expectedFailed: 0,
            expectError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### 4.2 Integration Tests

```go
func TestBackwardCompatibility(t *testing.T) {
    // Test that single dotfile add still works
    // Test that existing behavior is unchanged
}

func TestGlobExpansion(t *testing.T) {
    // Test glob pattern expansion
    // Test tilde expansion
    // Test invalid patterns
}

func TestDirectoryHandling(t *testing.T) {
    // Test directory copying
    // Test nested directory structures
    // Test mixed files and directories
}
```

### Phase 5: Documentation Updates

#### 5.1 CLI Help Text

```go
var dotAddCmd = &cobra.Command{
    Use:   "add [dotfile1] [dotfile2] ...",
    Short: "Add dotfiles to plonk management",
    Long: `Add one or more dotfiles to plonk management.

Supports glob patterns and mixed files/directories.

Examples:
  plonk dot add ~/.vimrc                    # Add single dotfile
  plonk dot add ~/.vimrc ~/.zshrc ~/.gitconfig  # Add multiple dotfiles
  plonk dot add ~/.config/nvim/*            # Add using glob pattern
  plonk dot add ~/.vimrc ~/.config/nvim/    # Mix files and directories
  plonk dot add --dry-run ~/.vimrc ~/.zshrc # Preview changes`,
    Args: cobra.MinimumNArgs(1),
}
```

### Phase 6: Implementation Checklist

#### Code Changes
- [ ] Update `dot_add.go` argument handling
- [ ] Implement `addDotfiles()` function
- [ ] Add path expansion and validation
- [ ] Implement glob pattern support
- [ ] Add directory handling
- [ ] Refactor single dotfile logic into `addSingleDotfile()`
- [ ] Add progress indication
- [ ] Add summary output
- [ ] Add structured output support
- [ ] Update error handling

#### Testing
- [ ] Add unit tests for multiple dotfile scenarios
- [ ] Add backward compatibility tests
- [ ] Add glob pattern tests
- [ ] Add directory handling tests
- [ ] Add error handling tests
- [ ] Add dry-run tests
- [ ] Test with different output formats

#### Documentation
- [ ] Update command help text
- [ ] Update CLI.md with examples
- [ ] Add usage examples to README.md

## Migration and Compatibility

### Backward Compatibility Guarantees

1. **Existing single dotfile usage unchanged**
   - `plonk dot add ~/.vimrc` works exactly as before
   - All existing flags continue to work
   - Output format for single dotfiles unchanged

2. **Path resolution behavior preserved**
   - Same path resolution logic for single files
   - Same error messages for invalid paths

### Performance Considerations

1. **Sequential processing** - Process files one at a time
2. **Early validation** - Expand and validate paths before processing
3. **Context checking** - Cancel gracefully on timeout
4. **Memory efficient** - Don't load large files into memory

## Future Enhancements

Once basic multiple dotfile support is implemented, these enhancements could be considered:

1. **Large directory handling** - For directories with many files (e.g., > 100), consider:
   - Progress indicators or counters during processing
   - Limits on file count with user confirmation for very large directories
   - Batch progress reporting instead of individual file progress

2. **Selective directory processing** - Allow users to specify which files within a directory to include/exclude

3. **Symlink handling** - Smart handling of symbolic links in dotfile directories

4. **Backup integration** - Automatic backup of existing dotfiles before overwriting

5. **Interactive mode** - Allow users to select which glob matches to include

6. **Conflict resolution** - Handle cases where target files already exist

7. **Skip vs Update flags** - Allow users to override default update behavior with `--skip-existing` flag

## Implementation Requirements Summary

Based on design review, the dotfile implementation should:

1. **Sequential processing** - Process dotfiles one at a time with discover-as-we-go validation
2. **Individual file progress** - Show progress for each individual file, even within directories
3. **Update by default** - Default behavior for existing managed dotfiles is to update (maintain consistency)
4. **Informative error messages** - Include helpful suggestions for common error scenarios
5. **Continue on failure** - Process all files even if some fail, show summary at end
6. **Glob pattern support** - Expand patterns like `~/.config/nvim/*` during processing
7. **Mixed input handling** - Support files, directories, and glob patterns in same command
8. **Preserve file attributes** - Maintain permissions, ownership, and timestamps when copying

## Expected User Experience

```bash
$ plonk dot add ~/.vimrc ~/.config/nvim/ ~/.nonexistent ~/.zshrc
✓ ~/.vimrc → vimrc
✓ ~/.config/nvim/init.lua → config/nvim/init.lua
✓ ~/.config/nvim/lua/config.lua → config/nvim/lua/config.lua
↻ ~/.config/nvim/lua/plugins.lua → config/nvim/lua/plugins.lua (updated)
✗ ~/.nonexistent - file not found
     Check if path exists: ls -la ~/.nonexistent
↻ ~/.zshrc → zshrc (updated)

Summary: 3 added, 2 updated, 0 skipped, 1 failed (5 total files)
```

## Implementation Notes

- **Directory processing**: Use `filepath.Walk()` to process each file individually
- **Path discovery**: Validate file existence during processing, not upfront
- **Attribute preservation**: Use `os.Stat()` and `os.Chtimes()` to preserve metadata
- **Target path calculation**: Map source paths to appropriate plonk config directory structure
- **Error handling**: Continue processing on individual file failures, collect results for summary
