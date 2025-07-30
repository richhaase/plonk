# Dotfile Drift Detection Implementation Plan

**Status**: 🔄 In Planning (2025-07-30)

## Overview
Add drift detection to identify when deployed dotfiles have been modified and differ from their source versions in the plonk configuration directory. This is a critical gap for v1.0.

## Goals
1. **Detect Drift**: Know when deployed files differ from source
2. **Show in Status**: Clear indication of out-of-sync files
3. **Preview Changes**: Show what will change on `plonk apply`
4. **Support Diff Tools**: Allow users to see detailed differences

## Design

### State Model Enhancement

Currently we have three states:
- `StateManaged` - File exists in both config and deployment
- `StateMissing` - File in config but not deployed
- `StateUntracked` - File deployed but not in config

We'll repurpose the existing `StateDegraded` (reserved for future use) as:
- `StateDrifted` - File exists in both but content differs

### Detection Approach

**Phase 1: Quick Detection (MVP)**
- Use file checksums (SHA256) for fast comparison
- Store checksums in metadata during reconciliation
- Compare on each status/apply operation

**Phase 2: Detailed Diff (Enhancement)**
- Support external diff tools via configuration
- Default to internal line-by-line diff for basic output
- Show preview of changes before apply

### Implementation Details

#### 1. Add Comparison to Reconciliation

**File**: `internal/resources/dotfiles/manager.go`

Add comparison method:
```go
// CompareFiles checks if two files have identical content
func (m *Manager) CompareFiles(path1, path2 string) (bool, error) {
    // Phase 1: Use SHA256 checksum comparison
    hash1, err := m.computeFileHash(path1)
    if err != nil {
        return false, err
    }

    hash2, err := m.computeFileHash(path2)
    if err != nil {
        return false, err
    }

    return hash1 == hash2, nil
}

func (m *Manager) computeFileHash(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()

    h := sha256.New()
    if _, err := io.Copy(h, file); err != nil {
        return "", err
    }

    return hex.EncodeToString(h.Sum(nil)), nil
}
```

#### 2. Update Reconciliation Logic

**File**: `internal/resources/reconcile.go`

Modify `ReconcileItems` to detect drift:
```go
// In ReconcileItems, when item.State would be StateManaged:
if actualItem, exists := actualMap[desiredItem.Name]; exists {
    // Check if comparison function is provided in metadata
    if compareFn, ok := desiredItem.Metadata["compare_fn"].(func() (bool, error)); ok {
        if identical, err := compareFn(); err == nil && !identical {
            item.State = StateDegraded // Using existing reserved state
            item.Meta["drift_status"] = "modified"
        } else {
            item.State = StateManaged
        }
    } else {
        item.State = StateManaged
    }
}
```

#### 3. Update Dotfile Resource

**File**: `internal/resources/dotfiles/resource.go`

Enhance `GetConfiguredDotfiles` to include comparison function:
```go
// When creating items for reconciliation
item.Metadata["compare_fn"] = func() (bool, error) {
    sourcePath := m.GetSourcePath(info.Source)
    destPath, _ := m.GetDestinationPath(info.Destination)
    return m.CompareFiles(sourcePath, destPath)
}
```

#### 4. Update Status Display

**File**: `internal/commands/status.go`

Add drift status to output:
- Show "drifted" status with different color/icon
- Include count in summary
- Group drifted items separately in table

**File**: `internal/output/status.go`
```go
func Drifted() string {
    return color.New(color.FgYellow).Sprint("drifted")
}
```

#### 5. Update Apply Command

**File**: `internal/resources/dotfiles/resource.go`

Handle drifted state in Apply:
```go
case resources.StateDegraded: // Drifted files
    // Treat like missing - copy from source to destination
    return d.applyMissing(ctx, item)
```

### Configuration for Diff Tools

**File**: `internal/config/config.go`

Add configuration option:
```go
type Config struct {
    // ... existing fields

    // DiffTool specifies external diff command (e.g., "diff", "git diff", "code --diff")
    // Empty means use internal diff
    DiffTool string `yaml:"diff_tool,omitempty"`
}
```

### Diff Display Options

**Option 1: Internal Diff (Phase 1)**
- Simple unified diff output
- Show first N lines of changes
- Indicate if more changes exist

**Option 2: External Tool (Phase 2)**
- Execute configured diff tool
- Pass source and destination paths
- Capture and display output

### User Experience

#### Status Command Output
```
DOTFILES
--------
SOURCE              TARGET                STATUS
vimrc               ~/.vimrc              drifted
zshrc               ~/.zshrc              deployed
gitconfig           ~/.gitconfig          missing

Summary: 2 managed, 1 missing, 1 drifted
```

#### Apply Preview (Future)
```
$ plonk apply --preview
The following changes will be made:

DOTFILES
  ~/.vimrc (drifted)
    - set number
    + set nonumber
    + set relativenumber

Would deploy 1 dotfile (1 drifted)
```

## Implementation Phases

### Phase 1: Basic Detection (MVP)
1. Add checksum comparison function
2. Update reconciliation to detect drift
3. Update status to show drifted state
4. Update apply to handle drifted files
5. Add tests for drift detection

### Phase 2: Enhanced Features (Revised 2025-07-30)
After Phase 1 completion and review, decided to implement only:
1. Add `plonk diff` command - Show differences for drifted files
2. Add diff tool configuration - Simple `diff_tool` in plonk.yaml
3. Default to `git diff --no-index` for zero-config experience

Rejected features:
- Internal diff display - Too complex, users have diff tools
- `--preview` flag - Not worth the complexity
- Selective apply - Users can use `plonk add` to re-add files
- Reverse sync - Too dangerous and complex
- Three-way merge - Very complex, rare use case

## Testing Plan

### Unit Tests
1. Test checksum computation
2. Test file comparison function
3. Test reconciliation with drift
4. Test status output formatting

### Integration Tests
1. Create dotfile, deploy, modify deployed version
2. Verify status shows drift
3. Apply and verify file restored
4. Test with symlinks, directories

### Manual Testing
1. Various file types (text, binary)
2. Permission changes vs content changes
3. Large files performance
4. Cross-platform behavior

## Performance Considerations

- Checksum computation is fast for typical dotfiles
- Cache checksums during reconciliation
- Skip binary files or large files (configurable)
- Parallel computation for multiple files

## Future Enhancements

1. **Selective Apply**: Choose which drifted files to restore
2. **Reverse Sync**: Update source from deployed (careful!)
3. **Drift Notifications**: Hook system for alerts
4. **Ignore Patterns**: Skip certain differences (e.g., timestamps)
5. **Three-Way Merge**: Handle conflicts when both change

## Success Criteria

1. ✅ Status command shows drifted files clearly
2. ✅ Apply command restores drifted files
3. ✅ Performance impact < 100ms for typical configs
4. ✅ Clear documentation on drift behavior
5. ✅ No false positives for identical files

## Open Questions

1. **Binary Files**: Should we detect drift in binary files?
   - Pro: Completeness
   - Con: Can't show meaningful diff
   - Proposal: Detect but show only "binary file differs"

2. **Symlinks**: How to handle symlink drift?
   - Target change vs link itself
   - Proposal: Check link target, not content

3. **Permissions**: Include permission changes as drift?
   - Pro: Full fidelity
   - Con: More complex, platform-specific
   - Proposal: Phase 2 feature

4. **Performance**: Checksum all files on every status?
   - Pro: Always accurate
   - Con: Slower for large configs
   - Proposal: Cache with mtime check

## Risk Mitigation

1. **Data Loss**: Apply creates backups before overwriting
2. **Performance**: Start with checksums, profile before optimizing
3. **Compatibility**: Test on macOS and Linux thoroughly
4. **User Confusion**: Clear documentation and status messages
