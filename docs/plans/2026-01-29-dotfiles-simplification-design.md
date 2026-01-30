# Dotfiles Module Simplification

## Goal

Simplify the dotfiles module to match the clean pattern established in the packages module. Same features, fewer abstractions.

**Objectives:**
- Maintainability: Fewer abstractions to understand
- Performance: Remove unnecessary indirection
- Consistency: Match packages module pattern
- Code health: 80% reduction in lines

## Current State

- 38 files, 6,124 lines
- 9 injected dependencies in Manager
- Template support (unused)
- Multiple wrapper patterns and abstract interfaces

## Design

### File Structure

```
internal/dotfiles/
├── types.go        # ~50 lines
├── fs.go           # ~80 lines
├── dotfiles.go     # ~400-500 lines
├── reconcile.go    # ~150 lines
└── dotfiles_test.go # ~300-400 lines
```

### Types (types.go)

```go
type Dotfile struct {
    Name   string // "zshrc" (without dot)
    Source string // "$PLONK_DIR/zshrc"
    Target string // "$HOME/.zshrc"
}

type State string

const (
    StateManaged   State = "managed"
    StateMissing   State = "missing"
    StateDrifted   State = "drifted"
    StateUnmanaged State = "unmanaged"
)

type DotfileStatus struct {
    Dotfile
    State State
}
```

### FileSystem Interface (fs.go)

Single interface replacing 4 abstractions:

```go
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    Stat(path string) (os.FileInfo, error)
    ReadDir(path string) ([]os.DirEntry, error)
    MkdirAll(path string, perm os.FileMode) error
    Remove(path string) error
    Rename(old, new string) error
}

type OSFileSystem struct{}
// All methods delegate to os package
```

### Manager (dotfiles.go)

```go
type Manager struct {
    configDir string     // $PLONK_DIR
    homeDir   string     // $HOME
    fs        FileSystem
    ignore    []string   // Patterns to ignore
}

func New(configDir, homeDir string, ignorePatterns []string) *Manager
func NewWithFS(..., fs FileSystem) *Manager  // For testing

// Core operations
func (m *Manager) List() ([]Dotfile, error)
func (m *Manager) Add(targetPath string) error
func (m *Manager) Remove(name string) error
func (m *Manager) Deploy(name string) error
func (m *Manager) IsDrifted(d Dotfile) (bool, error)
func (m *Manager) Diff(d Dotfile) (string, error)
```

### Path Handling

```go
// $PLONK_DIR/zshrc → $HOME/.zshrc
// $PLONK_DIR/config/nvim/init.lua → $HOME/.config/nvim/init.lua
func (m *Manager) toTarget(sourcePath string) string

// $HOME/.zshrc → $PLONK_DIR/zshrc
func (m *Manager) toSource(targetPath string) string

func (m *Manager) shouldIgnore(path string) bool
```

### Reconciliation (reconcile.go)

```go
func (m *Manager) Reconcile() ([]DotfileStatus, error)
func (m *Manager) Apply(dryRun bool) (ApplyResult, error)
```

Reconcile walks `$PLONK_DIR`, checks each file against `$HOME`, returns states.

## What Gets Deleted

**Template code (unused):**
- template.go, template_fileops.go, template_comparator.go, expander.go
- Associated test files

**Over-abstracted components:**
- atomic.go (inline temp-file-then-rename)
- config_handler.go (pass config as parameters)
- directory_scanner.go (inline in List())
- file_comparator.go (inline bytes.Equal)
- fileops.go (replaced by FileSystem interface)
- filter.go (inline in shouldIgnore())
- path_resolver.go, path_validator.go (inline helpers)
- scanner.go (inline in List())

## Testing Strategy

**Unit tests with MemoryFS:**
```go
type MemoryFS struct {
    Files map[string][]byte
    Dirs  map[string]bool
}
```

Test: List(), toTarget(), toSource(), IsDrifted(), Reconcile(), shouldIgnore()

**BATS integration tests:** Keep existing tests for full command flows.

## Features Preserved

- `plonk add` - Copy from $HOME to $PLONK_DIR
- `plonk rm` - Remove from $PLONK_DIR
- `plonk apply` - Deploy missing/drifted dotfiles
- `plonk status` - Show managed/missing/drifted states
- `plonk diff` - Show content differences
- Ignore patterns
- Dot-prefix handling

## Features Removed

- Template expansion ({{.Home}}, etc.)
- Template-aware file comparison

## Expected Result

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Files | 38 | 5 | -87% |
| Lines | 6,124 | ~1,100 | -82% |
| Interfaces | 9+ | 1 | -89% |
