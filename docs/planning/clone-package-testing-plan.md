# Clone Package Testing Plan

**Date**: 2025-08-03
**Author**: Analysis by Claude
**Current Coverage**: 0% (no test files exist)
**Target Coverage**: 70-80%
**Estimated Effort**: 6 days

## Executive Summary

The `internal/clone` package is responsible for cloning dotfile repositories and setting up plonk. Currently, it has 0% test coverage with no test files. This document provides a comprehensive plan to achieve 70-80% coverage through systematic refactoring and comprehensive testing.

## Package Overview

### Purpose
The clone package handles the `plonk clone` command, which:
1. Clones a git repository containing dotfiles
2. Creates default configuration if needed
3. Detects required package managers from lock file
4. Installs missing package managers
5. Runs `plonk apply` to set up dotfiles

### Current Structure
```
internal/clone/
├── setup.go      # Main orchestration logic (173 lines)
├── git.go        # Git URL parsing and cloning (67 lines)
├── tools.go      # Tool installation logic (90 lines)
└── prompts.go    # User interaction (36 lines)
```

### Key Functions

#### setup.go
- `CloneAndSetup()` - Main entry point
- `createDefaultConfig()` - Generates plonk.yaml
- `DetectRequiredManagers()` - Reads lock file for package managers
- `installDetectedManagers()` - Installs missing tools
- `installLanguagePackage()` - Delegates to package system

#### git.go
- `parseGitURL()` - Parses various git URL formats (PURE FUNCTION)
- `cloneRepository()` - Executes git clone command

#### tools.go
- `installCargo()` - Installs Rust/Cargo via curl
- `checkNetworkConnectivity()` - HTTP connectivity check

#### prompts.go
- `promptYesNo()` - Interactive user prompts

## External Dependencies Analysis

### Direct System Calls
```go
// Command execution
exec.LookPath("git")
exec.LookPath("cargo")
exec.Command("git", "clone", url, dir)
exec.CommandContext(ctx, "bash", "-c", script)

// File system
os.Stat(path)
os.WriteFile(path, data, perm)
os.RemoveAll(path)
os.Stdin

// Network
http.Client
http.NewRequestWithContext()
```

### Package Dependencies
```go
orchestrator.New()              // Run apply after clone
diagnostics.RunHealthChecks()   // Check system state
packages.InstallPackages()      // Install tools
lock.NewYAMLLockService()       // Read lock files
config.GetConfigDir()           // Configuration paths
output.StageUpdate()            // UI updates
```

## Testing Challenges

1. **Heavy I/O operations** - Git cloning, file creation, network calls
2. **System modifications** - Installing package managers
3. **User interaction** - Stdin prompts
4. **Complex orchestration** - Multiple steps with error handling
5. **No existing test infrastructure** - Starting from scratch

## Proposed Solution: Comprehensive Interface Abstraction

### Design Principles

1. **Dependency Injection** - All external dependencies through interfaces
2. **Backward Compatibility** - Maintain existing public API
3. **Testability First** - Design for easy mocking
4. **Separation of Concerns** - Clear boundaries between components

### Architecture

```
┌─────────────────────────────────────────────┐
│          Public API (unchanged)              │
│                                              │
│  func CloneAndSetup(ctx, repo, cfg) error   │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│           CloneService                       │
│                                              │
│  - git: GitOperations                       │
│  - fs: FileSystem                           │
│  - network: NetworkChecker                  │
│  - tools: ToolInstaller                     │
│  - ui: UserInterface                        │
│  - orchestrator: OrchestratorFactory        │
│  - diagnostics: DiagnosticsRunner           │
│  - packages: PackageManager                 │
│  - lock: LockReader                         │
└─────────────────────────────────────────────┘
```

## Implementation Plan

### Phase 1: Create Core Interfaces (Day 1)

**1. Create `internal/clone/interfaces.go`:**

```go
package clone

import (
    "context"
    "os"
    "github.com/richhaase/plonk/internal/orchestrator"
    "github.com/richhaase/plonk/internal/diagnostics"
    "github.com/richhaase/plonk/internal/lock"
)

// GitOperations handles git-related operations
type GitOperations interface {
    LookPath(tool string) (string, error)
    Clone(url, targetDir string) error
}

// FileSystem handles file system operations
type FileSystem interface {
    Stat(path string) (os.FileInfo, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    RemoveAll(path string) error
    Exists(path string) bool
}

// NetworkChecker handles network connectivity checks
type NetworkChecker interface {
    CheckConnectivity(ctx context.Context, url string) error
}

// ToolInstaller handles tool installation
type ToolInstaller interface {
    InstallCargo(ctx context.Context) error
    InstallViaPackageManager(ctx context.Context, manager, packageName string) error
    IsInstalled(tool string) bool
}

// UserInterface handles user interaction
type UserInterface interface {
    PromptYesNo(question string, defaultYes bool) bool
    Printf(format string, args ...interface{})
}

// OrchestratorFactory creates orchestrator instances
type OrchestratorFactory interface {
    Create(configDir, homeDir string, cfg interface{}) OrchestratorRunner
}

// OrchestratorRunner runs the apply operation
type OrchestratorRunner interface {
    Apply(ctx context.Context) (orchestrator.ApplyResult, error)
}

// DiagnosticsRunner runs health checks
type DiagnosticsRunner interface {
    RunHealthChecks() diagnostics.HealthReport
}

// PackageManager handles package operations
type PackageManager interface {
    InstallPackages(ctx context.Context, configDir string, packages []string, opts interface{}) (interface{}, error)
    GetRegistry() interface{}
}

// LockReader reads lock files
type LockReader interface {
    ReadLockFile(path string) (*lock.Lock, error)
}

// ConfigManager handles configuration
type ConfigManager interface {
    GetConfigDir() string
    GetHomeDir() string
    LoadWithDefaults(configDir string) interface{}
    GetDefaults() interface{}
}
```

**2. Create Real Implementations:**

```go
// internal/clone/implementations.go
package clone

// RealGitOperations implements GitOperations
type RealGitOperations struct{}

func (r *RealGitOperations) LookPath(tool string) (string, error) {
    return exec.LookPath(tool)
}

func (r *RealGitOperations) Clone(url, targetDir string) error {
    cmd := exec.Command("git", "clone", url, targetDir)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("git clone failed: %s\nOutput: %s", err, string(output))
    }
    return nil
}

// ... implement other real versions
```

**3. Create CloneService:**

```go
// internal/clone/service.go
package clone

// CloneService orchestrates the clone and setup process
type CloneService struct {
    git          GitOperations
    fs           FileSystem
    network      NetworkChecker
    tools        ToolInstaller
    ui           UserInterface
    orchestrator OrchestratorFactory
    diagnostics  DiagnosticsRunner
    packages     PackageManager
    lock         LockReader
    config       ConfigManager
}

// NewCloneService creates a new clone service with real implementations
func NewCloneService() *CloneService {
    return &CloneService{
        git:          &RealGitOperations{},
        fs:           &RealFileSystem{},
        network:      &RealNetworkChecker{},
        tools:        &RealToolInstaller{},
        ui:           &RealUserInterface{},
        orchestrator: &RealOrchestratorFactory{},
        diagnostics:  &RealDiagnosticsRunner{},
        packages:     &RealPackageManager{},
        lock:         &RealLockReader{},
        config:       &RealConfigManager{},
    }
}
```

### Phase 2: Refactor Existing Code (Days 2-3)

**1. Maintain Backward Compatibility:**

```go
// internal/clone/setup.go (modified)
package clone

// CloneAndSetup maintains the original API
func CloneAndSetup(ctx context.Context, gitRepo string, cfg Config) error {
    service := NewCloneService()
    return service.CloneAndSetup(ctx, gitRepo, cfg)
}

// Move implementation to CloneService
func (s *CloneService) CloneAndSetup(ctx context.Context, gitRepo string, cfg Config) error {
    // Refactored implementation using interfaces
}
```

**2. Extract Pure Functions:**

```go
// internal/clone/config_generator.go
package clone

type ConfigGenerator struct {
    defaults interface{}
}

func (c *ConfigGenerator) GenerateYAML() string {
    // Extract config generation logic
}

// internal/clone/url_parser.go
package clone

type URLParser struct{}

func (u *URLParser) Parse(input string) (string, error) {
    // Move parseGitURL logic here
}

// internal/clone/manager_detector.go
package clone

type ManagerDetector struct{}

func (m *ManagerDetector) DetectFromLock(lockFile *lock.Lock) []string {
    // Extract detection logic
}
```

### Phase 3: Implement Comprehensive Tests (Days 4-5)

**1. Create Mock Implementations:**

```go
// internal/clone/mocks_test.go
package clone

type MockGitOperations struct {
    LookPathFunc func(tool string) (string, error)
    CloneFunc    func(url, targetDir string) error
    calls        []string
}

func (m *MockGitOperations) LookPath(tool string) (string, error) {
    m.calls = append(m.calls, "LookPath:"+tool)
    if m.LookPathFunc != nil {
        return m.LookPathFunc(tool)
    }
    return "", nil
}

// ... implement other mocks
```

**2. Test Pure Functions:**

```go
// internal/clone/git_test.go
package clone

import "testing"

func TestParseGitURL(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "github shorthand",
            input: "user/repo",
            want:  "https://github.com/user/repo.git",
        },
        {
            name:  "https url",
            input: "https://github.com/user/repo",
            want:  "https://github.com/user/repo.git",
        },
        {
            name:  "ssh url",
            input: "git@github.com:user/repo.git",
            want:  "git@github.com:user/repo.git",
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    parser := &URLParser{}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parser.Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**3. Test Orchestration:**

```go
// internal/clone/setup_test.go
package clone

func TestCloneService_CloneAndSetup(t *testing.T) {
    tests := []struct {
        name      string
        gitRepo   string
        cfg       Config
        setupMock func(*CloneService)
        wantErr   bool
        validate  func(*testing.T, *CloneService)
    }{
        {
            name:    "successful clone and setup",
            gitRepo: "user/dotfiles",
            cfg:     Config{Interactive: false},
            setupMock: func(s *CloneService) {
                s.git = &MockGitOperations{
                    LookPathFunc: func(tool string) (string, error) {
                        return "/usr/bin/" + tool, nil
                    },
                    CloneFunc: func(url, dir string) error {
                        return nil
                    },
                }
                s.fs = &MockFileSystem{
                    ExistsFunc: func(path string) bool {
                        return false // PLONK_DIR doesn't exist
                    },
                    WriteFileFunc: func(path string, data []byte, perm os.FileMode) error {
                        return nil
                    },
                }
                // ... setup other mocks
            },
            validate: func(t *testing.T, s *CloneService) {
                // Verify expected calls were made
            },
        },
        {
            name:    "clone fails",
            gitRepo: "user/dotfiles",
            cfg:     Config{Interactive: false},
            setupMock: func(s *CloneService) {
                s.git = &MockGitOperations{
                    CloneFunc: func(url, dir string) error {
                        return fmt.Errorf("authentication failed")
                    },
                }
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := &CloneService{}
            tt.setupMock(s)

            err := s.CloneAndSetup(context.Background(), tt.gitRepo, tt.cfg)
            if (err != nil) != tt.wantErr {
                t.Errorf("CloneAndSetup() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.validate != nil {
                tt.validate(t, s)
            }
        })
    }
}
```

### Phase 4: Integration Tests (Day 6)

**1. Create Integration Test Suite:**

```go
// internal/clone/integration_test.go
// +build integration

package clone

import (
    "testing"
    "os"
    "path/filepath"
)

func TestCloneAndSetup_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Create temp directory
    tempDir := t.TempDir()

    // Set PLONK_DIR to temp location
    os.Setenv("PLONK_DIR", filepath.Join(tempDir, "plonk"))
    defer os.Unsetenv("PLONK_DIR")

    // Test with a real test repository
    testRepo := "https://github.com/plonk-test/minimal-dotfiles.git"

    cfg := Config{
        Interactive: false,
        NoApply:     true, // Don't run apply in tests
    }

    err := CloneAndSetup(context.Background(), testRepo, cfg)
    if err != nil {
        t.Fatalf("CloneAndSetup() failed: %v", err)
    }

    // Verify results
    plonkDir := filepath.Join(tempDir, "plonk")
    if _, err := os.Stat(filepath.Join(plonkDir, "plonk.yaml")); err != nil {
        t.Error("Expected plonk.yaml to be created")
    }

    if _, err := os.Stat(filepath.Join(plonkDir, ".git")); err != nil {
        t.Error("Expected git repository to be cloned")
    }
}
```

## Test Coverage Strategy

### Unit Tests (Target: 80%)

1. **Pure Functions (100% coverage)**
   - `parseGitURL()`
   - `getManagerDescription()`
   - `getManualInstallInstructions()`
   - Config generation logic

2. **Service Methods (75% coverage)**
   - All CloneService methods with mocked dependencies
   - Error scenarios
   - Edge cases

3. **Integration Points (70% coverage)**
   - Mock external package calls
   - Test orchestration flow

### Integration Tests

1. **Git Operations**
   - Test with local git repositories
   - Test various URL formats

2. **File System**
   - Use temp directories
   - Test cleanup on failure

3. **Tool Detection**
   - Test with sample lock files
   - Test manager detection logic

## Implementation Guidelines

### Do's
- ✅ Maintain backward compatibility
- ✅ Test error scenarios thoroughly
- ✅ Use table-driven tests
- ✅ Document mock behavior
- ✅ Keep interfaces focused and small

### Don'ts
- ❌ Don't modify public API
- ❌ Don't test external tools directly
- ❌ Don't rely on network in unit tests
- ❌ Don't modify system state in unit tests

## Success Criteria

- [ ] All interfaces defined and documented
- [ ] Real implementations created
- [ ] CloneService refactored with DI
- [ ] Unit tests achieve 75%+ coverage
- [ ] Integration tests cover main flows
- [ ] Public API remains unchanged
- [ ] All existing functionality preserved

## Risk Mitigation

1. **Feature Flags**
   ```go
   var useNewImplementation = os.Getenv("PLONK_USE_NEW_CLONE") == "true"
   ```

2. **Gradual Rollout**
   - Test with subset of users first
   - Monitor for issues
   - Full rollout after validation

3. **Rollback Plan**
   - Keep old implementation available
   - Quick switch via environment variable

## Maintenance Notes

### Adding New Features
1. Add method to appropriate interface
2. Implement in real version
3. Add to mock
4. Write tests first

### Debugging Tests
- Use `-v` flag for verbose output
- Check mock call logs
- Verify mock setup matches test scenario

## Conclusion

This plan transforms the clone package from 0% to 70-80% test coverage through systematic refactoring and comprehensive testing. The approach maintains backward compatibility while significantly improving code quality and maintainability.
