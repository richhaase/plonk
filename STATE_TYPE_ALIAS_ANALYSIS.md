# State Type Alias Analysis

## Type Aliases Defined in `internal/state/types.go`

1. **ItemState** → `interfaces.ItemState`
2. **StateManaged** → `interfaces.StateManaged` (constant)
3. **StateMissing** → `interfaces.StateMissing` (constant)
4. **StateUntracked** → `interfaces.StateUntracked` (constant)
5. **Item** → `interfaces.Item`
6. **ConfigItem** → `interfaces.ConfigItem`
7. **ActualItem** → `interfaces.ActualItem`
8. **Result** → `types.Result`
9. **Summary** → `types.Summary`

## Usage Analysis

### Aliases NOT Used in Code
- `state.ItemState` - 0 occurrences
- `state.StateManaged` - 0 occurrences
- `state.StateMissing` - 0 occurrences
- `state.StateUntracked` - 0 occurrences
- `state.ConfigItem` - 0 occurrences
- `state.ActualItem` - 0 occurrences

### Aliases Actually Used

#### 1. `state.Item` (18 occurrences)
- **internal/services/package_operations.go**: 1
  - Line 60: `missingByManager := make(map[string][]state.Item)`
- **internal/commands/shared.go**: 17
  - Lines 69-71: Creating filtered slices
  - Line 96: Clearing untracked items
  - Lines 134, 278: In itemWithState struct
  - Lines 216-223, 226: Clearing result slices
  - Line 356: Function parameter

#### 2. `state.Result` (16 occurrences)
- **internal/commands/status.go**: 5
  - Lines 99, 104, 221, 239: Function signatures and variable declarations
- **internal/commands/shared.go**: 8
  - Line 110: In packageListResultWrapper struct
  - Line 206: Creating new Result
  - Line 241: In dotfileListResultWrapper struct
- **internal/runtime/context.go**: 3
  - Lines 200, 206, 212: Return types

#### 3. `state.Summary` (4 occurrences)
- **internal/commands/status.go**: 4
  - Lines 98-100, 126: Comments, function signatures, and struct fields

## Files Importing `internal/state`

1. **internal/services/package_operations.go** - Uses `state.Item`
2. **internal/services/dotfile_operations.go** - No type aliases used
3. **internal/commands/status.go** - Uses `state.Result`, `state.Summary`
4. **internal/commands/shared.go** - Uses `state.Item`, `state.Result`
5. **internal/runtime/context_simple.go** - No type aliases used
6. **internal/runtime/context.go** - Uses `state.Result`
7. **internal/managers/registry.go** - No type aliases used
8. **internal/core/state.go** - No type aliases used

## Refactoring Impact

### High Priority (Used Aliases)
1. **state.Item** → `interfaces.Item` (18 replacements)
2. **state.Result** → `types.Result` (16 replacements)
3. **state.Summary** → `types.Summary` (4 replacements)

### Low Priority (Unused Aliases)
These can be removed immediately as they have no callers:
- ItemState
- StateManaged, StateMissing, StateUntracked constants
- ConfigItem
- ActualItem

### Import Changes Required
After replacing aliases, the following imports will need to be added:
- Files using `state.Item` will need `"github.com/richhaase/plonk/internal/interfaces"`
- Files using `state.Result` or `state.Summary` will need `"github.com/richhaase/plonk/internal/types"`

Some files may be able to remove the `internal/state` import entirely if they only use type aliases.
