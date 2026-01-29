# Package Simplification Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Simplify package management to track/untrack/apply only, with 5 managers (brew, pnpm, cargo, go, uv).

**Architecture:** Replace complex PackageManager interface with 2-method interface (IsInstalled, Install). Remove CommandExecutor abstraction, call exec.Command directly. New lock format v3 groups packages by manager. Test with BATS instead of Go unit tests.

**Tech Stack:** Go 1.24, exec.Command, YAML v3 lock format, BATS for testing.

---

## Phase 1: New Lock Format (v3)

### Task 1: Define Lock v3 Types

**Files:**
- Create: `internal/lock/v3.go`

**Step 1: Create v3 types file**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

// LockV3 represents the simplified v3 lock format
type LockV3 struct {
	Version  int                 `yaml:"version"`
	Packages map[string][]string `yaml:"packages,omitempty"` // manager -> []package
}

// NewLockV3 creates an empty v3 lock
func NewLockV3() *LockV3 {
	return &LockV3{
		Version:  3,
		Packages: make(map[string][]string),
	}
}

// AddPackage adds a package under its manager (maintains sorted order)
func (l *LockV3) AddPackage(manager, pkg string) {
	if l.Packages == nil {
		l.Packages = make(map[string][]string)
	}

	// Check if already exists
	for _, existing := range l.Packages[manager] {
		if existing == pkg {
			return
		}
	}

	l.Packages[manager] = append(l.Packages[manager], pkg)
	sort.Strings(l.Packages[manager])
}

// RemovePackage removes a package from its manager
func (l *LockV3) RemovePackage(manager, pkg string) {
	if l.Packages == nil {
		return
	}

	pkgs := l.Packages[manager]
	for i, existing := range pkgs {
		if existing == pkg {
			l.Packages[manager] = append(pkgs[:i], pkgs[i+1:]...)
			break
		}
	}

	// Remove manager key if empty
	if len(l.Packages[manager]) == 0 {
		delete(l.Packages, manager)
	}
}

// HasPackage checks if a package is tracked
func (l *LockV3) HasPackage(manager, pkg string) bool {
	for _, existing := range l.Packages[manager] {
		if existing == pkg {
			return true
		}
	}
	return false
}

// GetPackages returns all packages for a manager
func (l *LockV3) GetPackages(manager string) []string {
	return l.Packages[manager]
}

// GetAllPackages returns all manager:package pairs
func (l *LockV3) GetAllPackages() []string {
	var result []string
	for manager, pkgs := range l.Packages {
		for _, pkg := range pkgs {
			result = append(result, manager+":"+pkg)
		}
	}
	sort.Strings(result)
	return result
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/lock/...`
Expected: Success, no errors

**Step 3: Commit**

```bash
git add internal/lock/v3.go
git commit -m "feat(lock): add v3 lock format types"
```

---

### Task 2: Add Lock v3 Service

**Files:**
- Modify: `internal/lock/v3.go` (add Read/Write methods)

**Step 1: Add file operations to v3.go**

Append to `internal/lock/v3.go`:

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// LockV3Service handles v3 lock file operations
type LockV3Service struct {
	lockPath string
}

// NewLockV3Service creates a new v3 lock service
func NewLockV3Service(configDir string) *LockV3Service {
	return &LockV3Service{
		lockPath: filepath.Join(configDir, LockFileName),
	}
}

// Read reads the lock file, auto-migrating v2 if needed
func (s *LockV3Service) Read() (*LockV3, error) {
	// If lock file doesn't exist, return empty lock
	if _, err := os.Stat(s.lockPath); os.IsNotExist(err) {
		return NewLockV3(), nil
	}

	data, err := os.ReadFile(s.lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	// Try to detect version
	var versionCheck struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &versionCheck); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	// Handle v2 migration
	if versionCheck.Version == 2 {
		return s.migrateV2(data)
	}

	// Parse v3
	var lock LockV3
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	if lock.Version != 3 {
		return nil, fmt.Errorf("unsupported lock version %d", lock.Version)
	}

	return &lock, nil
}

// Write writes the lock to disk
func (s *LockV3Service) Write(lock *LockV3) error {
	if lock == nil {
		return fmt.Errorf("cannot write nil lock")
	}

	lock.Version = 3

	data, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lock: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	if err := os.WriteFile(s.lockPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// migrateV2 converts a v2 lock to v3 format
func (s *LockV3Service) migrateV2(data []byte) (*LockV3, error) {
	var v2 Lock
	if err := yaml.Unmarshal(data, &v2); err != nil {
		return nil, fmt.Errorf("failed to parse v2 lock: %w", err)
	}

	v3 := NewLockV3()

	for _, resource := range v2.Resources {
		if resource.Type != "package" {
			continue
		}

		// Extract manager and name from metadata
		manager, _ := resource.Metadata["manager"].(string)
		name, _ := resource.Metadata["name"].(string)

		if manager != "" && name != "" {
			v3.AddPackage(manager, name)
		}
	}

	return v3, nil
}

// GetLockPath returns the path to the lock file
func (s *LockV3Service) GetLockPath() string {
	return s.lockPath
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/lock/...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/lock/v3.go
git commit -m "feat(lock): add v3 lock service with v2 migration"
```

---

## Phase 2: Simplified Package Managers

### Task 3: Create New Package Interface

**Files:**
- Create: `internal/packages/manager.go`

**Step 1: Create simplified interface**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "context"

// Manager defines the simplified package manager interface.
// Only two operations: check if installed, install if missing.
type Manager interface {
	// IsInstalled checks if a package is installed
	IsInstalled(ctx context.Context, name string) (bool, error)

	// Install installs a package (should be idempotent)
	Install(ctx context.Context, name string) error
}

// SupportedManagers lists all available package managers
var SupportedManagers = []string{"brew", "cargo", "go", "pnpm", "uv"}

// IsSupportedManager checks if a manager name is valid
func IsSupportedManager(name string) bool {
	for _, m := range SupportedManagers {
		if m == name {
			return true
		}
	}
	return false
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/packages/...`
Expected: Compilation errors (old code conflicts) - that's OK for now

**Step 3: Commit**

```bash
git add internal/packages/manager.go
git commit -m "feat(packages): add simplified Manager interface"
```

---

### Task 4: Implement Brew Manager (New)

**Files:**
- Create: `internal/packages/brew_simple.go`

**Step 1: Create simplified brew manager**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os/exec"
	"strings"
)

// BrewSimple implements Manager for Homebrew
type BrewSimple struct{}

// NewBrewSimple creates a new Homebrew manager
func NewBrewSimple() *BrewSimple {
	return &BrewSimple{}
}

// IsInstalled checks if a package is installed via brew
func (b *BrewSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "brew", "list", "--formula", name)
	err := cmd.Run()
	return err == nil, nil
}

// Install installs a package via brew
func (b *BrewSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed (idempotent)
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return err
	}
	return nil
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/packages/...`
Expected: May have errors from old code, but new file should be syntactically correct

**Step 3: Commit**

```bash
git add internal/packages/brew_simple.go
git commit -m "feat(packages): add simplified brew manager"
```

---

### Task 5: Implement Cargo Manager (New)

**Files:**
- Create: `internal/packages/cargo_simple.go`

**Step 1: Create simplified cargo manager**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os/exec"
	"strings"
)

// CargoSimple implements Manager for Rust's Cargo
type CargoSimple struct{}

// NewCargoSimple creates a new Cargo manager
func NewCargoSimple() *CargoSimple {
	return &CargoSimple{}
}

// IsInstalled checks if a package is installed via cargo
func (c *CargoSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "cargo", "install", "--list")
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	// Parse output: each installed package starts at column 0
	// Format: "package_name v1.2.3:\n    binary1\n"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true, nil
		}
	}
	return false, nil
}

// Install installs a package via cargo
func (c *CargoSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "cargo", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed (idempotent)
		outStr := strings.ToLower(string(output))
		if strings.Contains(outStr, "already exists") || strings.Contains(outStr, "already installed") {
			return nil
		}
		return err
	}
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/packages/cargo_simple.go
git commit -m "feat(packages): add simplified cargo manager"
```

---

### Task 6: Implement Go Manager (New)

**Files:**
- Create: `internal/packages/go_simple.go`

**Step 1: Create simplified go manager**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoSimple implements Manager for Go packages
type GoSimple struct{}

// NewGoSimple creates a new Go manager
func NewGoSimple() *GoSimple {
	return &GoSimple{}
}

// IsInstalled checks if a go package is installed by looking for its binary
func (g *GoSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	binDir := goBinDir()
	if binDir == "" {
		return false, nil
	}

	// Extract binary name from package path
	// e.g., "golang.org/x/tools/gopls" -> "gopls"
	binaryName := name
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		binaryName = parts[len(parts)-1]
	}
	// Remove @version suffix if present
	if idx := strings.Index(binaryName, "@"); idx != -1 {
		binaryName = binaryName[:idx]
	}

	binPath := filepath.Join(binDir, binaryName)
	_, err := os.Stat(binPath)
	return err == nil, nil
}

// Install installs a go package
func (g *GoSimple) Install(ctx context.Context, name string) error {
	// Add @latest if no version specified
	pkg := name
	if !strings.Contains(name, "@") {
		pkg = name + "@latest"
	}

	cmd := exec.CommandContext(ctx, "go", "install", pkg)
	return cmd.Run()
}

// goBinDir returns the directory where go install puts binaries
func goBinDir() string {
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		gopath = filepath.Join(home, "go")
	}

	return filepath.Join(gopath, "bin")
}
```

**Step 2: Commit**

```bash
git add internal/packages/go_simple.go
git commit -m "feat(packages): add simplified go manager"
```

---

### Task 7: Implement PNPM Manager (New)

**Files:**
- Create: `internal/packages/pnpm_simple.go`

**Step 1: Create simplified pnpm manager**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

// PNPMSimple implements Manager for pnpm
type PNPMSimple struct{}

// NewPNPMSimple creates a new pnpm manager
func NewPNPMSimple() *PNPMSimple {
	return &PNPMSimple{}
}

// IsInstalled checks if a package is globally installed via pnpm
func (p *PNPMSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "pnpm", "list", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	// pnpm outputs JSON array: [{"dependencies": {...}}]
	var result []struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return false, nil
	}

	for _, item := range result {
		if _, ok := item.Dependencies[name]; ok {
			return true, nil
		}
	}
	return false, nil
}

// Install installs a package globally via pnpm
func (p *PNPMSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "pnpm", "add", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return err
	}
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/packages/pnpm_simple.go
git commit -m "feat(packages): add simplified pnpm manager"
```

---

### Task 8: Implement UV Manager (New)

**Files:**
- Create: `internal/packages/uv_simple.go`

**Step 1: Create simplified uv manager**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os/exec"
	"strings"
)

// UVSimple implements Manager for uv (Python)
type UVSimple struct{}

// NewUVSimple creates a new uv manager
func NewUVSimple() *UVSimple {
	return &UVSimple{}
}

// IsInstalled checks if a tool is installed via uv
func (u *UVSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "uv", "tool", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	// Parse output: tool names are first token on each line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true, nil
		}
	}
	return false, nil
}

// Install installs a tool via uv
func (u *UVSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "uv", "tool", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return err
	}
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/packages/uv_simple.go
git commit -m "feat(packages): add simplified uv manager"
```

---

### Task 9: Create New Registry

**Files:**
- Create: `internal/packages/registry_simple.go`

**Step 1: Create simplified registry**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "fmt"

// GetManager returns a Manager by name
func GetManager(name string) (Manager, error) {
	switch name {
	case "brew":
		return NewBrewSimple(), nil
	case "cargo":
		return NewCargoSimple(), nil
	case "go":
		return NewGoSimple(), nil
	case "pnpm":
		return NewPNPMSimple(), nil
	case "uv":
		return NewUVSimple(), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s (supported: %v)", name, SupportedManagers)
	}
}
```

**Step 2: Commit**

```bash
git add internal/packages/registry_simple.go
git commit -m "feat(packages): add simplified manager registry"
```

---

## Phase 3: New Commands

### Task 10: Create Track Command

**Files:**
- Create: `internal/commands/track.go`

**Step 1: Create track command**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track <manager:package>...",
	Short: "Track installed packages",
	Long: `Track packages that are already installed on your system.

This command verifies that each package is installed, then adds it to your
lock file for management. Use this to record packages you want to keep
in sync across machines.

The package must already be installed - track only records existing packages.

Examples:
  plonk track brew:ripgrep           # Track a brew package
  plonk track cargo:bat go:gopls     # Track multiple packages
  plonk track pnpm:typescript        # Track a pnpm package`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runTrack,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(trackCmd)
}

func runTrack(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()
	lockSvc := lock.NewLockV3Service(configDir)

	lockFile, err := lockSvc.Read()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	ctx := context.Background()
	var tracked, skipped, failed int

	for _, arg := range args {
		manager, pkg, err := parsePackageSpec(arg)
		if err != nil {
			fmt.Printf("Error: %s: %v\n", arg, err)
			failed++
			continue
		}

		// Check if already tracked
		if lockFile.HasPackage(manager, pkg) {
			fmt.Printf("Skipping %s:%s (already tracked)\n", manager, pkg)
			skipped++
			continue
		}

		// Get manager and verify package is installed
		mgr, err := packages.GetManager(manager)
		if err != nil {
			fmt.Printf("Error: %s: %v\n", arg, err)
			failed++
			continue
		}

		installed, err := mgr.IsInstalled(ctx, pkg)
		if err != nil {
			fmt.Printf("Error checking %s:%s: %v\n", manager, pkg, err)
			failed++
			continue
		}

		if !installed {
			fmt.Printf("Error: %s:%s is not installed\n", manager, pkg)
			failed++
			continue
		}

		// Add to lock file
		lockFile.AddPackage(manager, pkg)
		fmt.Printf("Tracking %s:%s\n", manager, pkg)
		tracked++
	}

	// Write updated lock file
	if tracked > 0 {
		if err := lockSvc.Write(lockFile); err != nil {
			return fmt.Errorf("failed to write lock file: %w", err)
		}
	}

	// Summary
	if failed > 0 {
		return fmt.Errorf("tracked %d, skipped %d, failed %d", tracked, skipped, failed)
	}

	return nil
}

// parsePackageSpec parses "manager:package" format
func parsePackageSpec(spec string) (manager, pkg string, err error) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected manager:package")
	}

	manager = parts[0]
	pkg = parts[1]

	if !packages.IsSupportedManager(manager) {
		return "", "", fmt.Errorf("unsupported manager: %s (supported: %v)", manager, packages.SupportedManagers)
	}

	if pkg == "" {
		return "", "", fmt.Errorf("package name cannot be empty")
	}

	return manager, pkg, nil
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/commands/...`
Expected: May have conflicts with old code

**Step 3: Commit**

```bash
git add internal/commands/track.go
git commit -m "feat(commands): add track command"
```

---

### Task 11: Create Untrack Command

**Files:**
- Create: `internal/commands/untrack.go`

**Step 1: Create untrack command**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/spf13/cobra"
)

var untrackCmd = &cobra.Command{
	Use:   "untrack <manager:package>...",
	Short: "Stop tracking packages",
	Long: `Stop tracking packages without uninstalling them.

This command removes packages from your lock file but does NOT uninstall
them from your system. The packages remain installed, they're just no
longer managed by plonk.

Examples:
  plonk untrack brew:ripgrep           # Stop tracking a brew package
  plonk untrack cargo:bat go:gopls     # Stop tracking multiple packages`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runUntrack,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(untrackCmd)
}

func runUntrack(cmd *cobra.Command, args []string) error {
	configDir := config.GetDefaultConfigDirectory()
	lockSvc := lock.NewLockV3Service(configDir)

	lockFile, err := lockSvc.Read()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	var untracked, skipped, failed int

	for _, arg := range args {
		manager, pkg, err := parsePackageSpec(arg)
		if err != nil {
			fmt.Printf("Error: %s: %v\n", arg, err)
			failed++
			continue
		}

		// Check if tracked
		if !lockFile.HasPackage(manager, pkg) {
			fmt.Printf("Skipping %s:%s (not tracked)\n", manager, pkg)
			skipped++
			continue
		}

		// Remove from lock file
		lockFile.RemovePackage(manager, pkg)
		fmt.Printf("Untracking %s:%s\n", manager, pkg)
		untracked++
	}

	// Write updated lock file
	if untracked > 0 {
		if err := lockSvc.Write(lockFile); err != nil {
			return fmt.Errorf("failed to write lock file: %w", err)
		}
	}

	// Summary
	if failed > 0 {
		return fmt.Errorf("untracked %d, skipped %d, failed %d", untracked, skipped, failed)
	}

	return nil
}
```

**Step 2: Commit**

```bash
git add internal/commands/untrack.go
git commit -m "feat(commands): add untrack command"
```

---

### Task 12: Create Package Apply Logic

**Files:**
- Create: `internal/packages/apply_simple.go`

**Step 1: Create apply logic**

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/lock"
)

// ApplyResult holds the result of applying packages
type ApplyResult struct {
	Installed []string
	Skipped   []string
	Failed    []string
	Errors    []error
}

// Apply installs all tracked packages that are missing
func Apply(ctx context.Context, configDir string, dryRun bool) (*ApplyResult, error) {
	lockSvc := lock.NewLockV3Service(configDir)
	lockFile, err := lockSvc.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	result := &ApplyResult{}

	// Process each manager
	for manager, pkgs := range lockFile.Packages {
		mgr, err := GetManager(manager)
		if err != nil {
			result.Failed = append(result.Failed, manager+":*")
			result.Errors = append(result.Errors, fmt.Errorf("manager %s: %w", manager, err))
			continue
		}

		for _, pkg := range pkgs {
			spec := manager + ":" + pkg

			// Check if installed
			installed, err := mgr.IsInstalled(ctx, pkg)
			if err != nil {
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, err))
				continue
			}

			if installed {
				result.Skipped = append(result.Skipped, spec)
				continue
			}

			// Install
			if dryRun {
				fmt.Printf("Would install %s\n", spec)
				result.Installed = append(result.Installed, spec)
				continue
			}

			fmt.Printf("Installing %s...\n", spec)
			if err := mgr.Install(ctx, pkg); err != nil {
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, err))
				continue
			}

			result.Installed = append(result.Installed, spec)
		}
	}

	return result, nil
}
```

**Step 2: Commit**

```bash
git add internal/packages/apply_simple.go
git commit -m "feat(packages): add simplified apply logic"
```

---

## Phase 4: Cleanup Old Code

### Task 13: Delete Old Package Files

**Files:**
- Delete: `internal/packages/base.go`
- Delete: `internal/packages/executor.go`
- Delete: `internal/packages/interfaces.go`
- Delete: `internal/packages/helpers.go`
- Delete: `internal/packages/operations.go`
- Delete: `internal/packages/reconcile.go`
- Delete: `internal/packages/spec.go`
- Delete: `internal/packages/types.go`
- Delete: `internal/packages/validation.go`
- Delete: `internal/packages/brew.go`
- Delete: `internal/packages/cargo.go`
- Delete: `internal/packages/go.go`
- Delete: `internal/packages/pnpm.go`
- Delete: `internal/packages/uv.go`
- Delete: `internal/packages/npm.go`
- Delete: `internal/packages/gem.go`
- Delete: `internal/packages/bun.go`
- Delete: `internal/packages/registry.go`
- Delete: `internal/packages/apply.go`

**Step 1: Delete old files**

```bash
rm internal/packages/base.go
rm internal/packages/executor.go
rm internal/packages/interfaces.go
rm internal/packages/helpers.go
rm internal/packages/operations.go
rm internal/packages/reconcile.go
rm internal/packages/spec.go
rm internal/packages/types.go
rm internal/packages/validation.go
rm internal/packages/brew.go
rm internal/packages/cargo.go
rm internal/packages/go.go
rm internal/packages/pnpm.go
rm internal/packages/uv.go
rm internal/packages/npm.go
rm internal/packages/gem.go
rm internal/packages/bun.go
rm internal/packages/registry.go
rm internal/packages/apply.go
```

**Step 2: Delete test files**

```bash
rm internal/packages/*_test.go
```

**Step 3: Commit**

```bash
git add -A internal/packages/
git commit -m "refactor(packages): remove old complex implementation"
```

---

### Task 14: Rename Simple Files

**Files:**
- Rename: `internal/packages/brew_simple.go` → `internal/packages/brew.go`
- Rename: `internal/packages/cargo_simple.go` → `internal/packages/cargo.go`
- Rename: `internal/packages/go_simple.go` → `internal/packages/go.go`
- Rename: `internal/packages/pnpm_simple.go` → `internal/packages/pnpm.go`
- Rename: `internal/packages/uv_simple.go` → `internal/packages/uv.go`
- Rename: `internal/packages/registry_simple.go` → `internal/packages/registry.go`
- Rename: `internal/packages/apply_simple.go` → `internal/packages/apply.go`

**Step 1: Rename files**

```bash
mv internal/packages/brew_simple.go internal/packages/brew.go
mv internal/packages/cargo_simple.go internal/packages/cargo.go
mv internal/packages/go_simple.go internal/packages/go.go
mv internal/packages/pnpm_simple.go internal/packages/pnpm.go
mv internal/packages/uv_simple.go internal/packages/uv.go
mv internal/packages/registry_simple.go internal/packages/registry.go
mv internal/packages/apply_simple.go internal/packages/apply.go
```

**Step 2: Commit**

```bash
git add -A internal/packages/
git commit -m "refactor(packages): rename simplified implementations"
```

---

### Task 15: Delete Old Commands

**Files:**
- Delete: `internal/commands/install.go`
- Delete: `internal/commands/uninstall.go`
- Delete: `internal/commands/upgrade.go`
- Delete: `internal/commands/packages.go` (if exists)

**Step 1: Delete old command files**

```bash
rm -f internal/commands/install.go
rm -f internal/commands/uninstall.go
rm -f internal/commands/upgrade.go
rm -f internal/commands/packages.go
```

**Step 2: Delete old command tests**

```bash
rm -f internal/commands/upgrade_*_test.go
rm -f internal/commands/cli_upgrade_test.go
```

**Step 3: Commit**

```bash
git add -A internal/commands/
git commit -m "refactor(commands): remove old package commands"
```

---

### Task 16: Update Apply Command

**Files:**
- Modify: `internal/commands/apply.go`

**Step 1: Simplify apply command to use new package apply**

This will require updating the orchestrator or bypassing it. The apply command needs to:
1. Call `packages.Apply()` for packages
2. Call existing dotfiles apply for dotfiles

This is a larger refactor - update `apply.go` to use the new simplified package apply.

**Step 2: Commit**

```bash
git add internal/commands/apply.go
git commit -m "refactor(commands): update apply to use simplified packages"
```

---

## Phase 5: Fix Compilation & Integration

### Task 17: Fix Imports and Compilation

**Step 1: Try to build**

Run: `go build ./...`

**Step 2: Fix any import errors**

Update imports in files that reference old package types.

**Step 3: Keep fixing until it compiles**

Run: `go build ./...`
Expected: Success

**Step 4: Commit**

```bash
git add -A
git commit -m "fix: resolve compilation errors after refactor"
```

---

### Task 18: Update Orchestrator

**Files:**
- Modify: `internal/orchestrator/coordinator.go`

**Step 1: Update orchestrator to use new packages.Apply**

The orchestrator needs to call the new simplified `packages.Apply()` instead of the old complex flow.

**Step 2: Commit**

```bash
git add internal/orchestrator/
git commit -m "refactor(orchestrator): use simplified package apply"
```

---

### Task 19: Update Config Validation

**Files:**
- Modify: `internal/config/config.go`

**Step 1: Update manager validation to use new list**

Remove references to old manager list, use `packages.SupportedManagers`.

**Step 2: Commit**

```bash
git add internal/config/
git commit -m "refactor(config): update manager validation"
```

---

## Phase 6: BATS Tests

### Task 20: Update BATS Tests

**Files:**
- Delete: `tests/bats/behavioral/02-package-install.bats`
- Delete: `tests/bats/behavioral/03-package-uninstall.bats`
- Delete: `tests/bats/behavioral/07-package-upgrade.bats`
- Create: `tests/bats/behavioral/02-package-track.bats`

**Step 1: Remove old package test files**

```bash
rm tests/bats/behavioral/02-package-install.bats
rm tests/bats/behavioral/03-package-uninstall.bats
rm tests/bats/behavioral/07-package-upgrade.bats
```

**Step 2: Create new track test file**

```bash
# tests/bats/behavioral/02-package-track.bats

#!/usr/bin/env bats

load '../test_helper'

@test "track requires manager:package format" {
  run plonk track ripgrep
  [ "$status" -ne 0 ]
  [[ "$output" == *"invalid format"* ]]
}

@test "track rejects unsupported manager" {
  run plonk track npm:typescript
  [ "$status" -ne 0 ]
  [[ "$output" == *"unsupported manager"* ]]
}

@test "track fails for uninstalled package" {
  run plonk track brew:this-package-does-not-exist-12345
  [ "$status" -ne 0 ]
  [[ "$output" == *"not installed"* ]]
}

@test "untrack removes from lock file" {
  # First track something that's installed
  plonk track brew:jq || skip "jq not installed"

  run plonk untrack brew:jq
  [ "$status" -eq 0 ]
  [[ "$output" == *"Untracking"* ]]
}
```

**Step 3: Commit**

```bash
git add -A tests/bats/
git commit -m "test(bats): update package tests for track/untrack"
```

---

### Task 21: Run Full Test Suite

**Step 1: Build**

Run: `go build ./...`
Expected: Success

**Step 2: Run BATS tests**

Run: `bats tests/bats/behavioral/`

**Step 3: Fix any failures**

Address test failures as needed.

**Step 4: Final commit**

```bash
git add -A
git commit -m "test: fix remaining test failures"
```

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | 1-2 | New lock format v3 |
| 2 | 3-9 | Simplified managers (5 total) |
| 3 | 10-12 | New commands (track, untrack) |
| 4 | 13-16 | Delete old code |
| 5 | 17-19 | Fix compilation |
| 6 | 20-21 | Update BATS tests |

Total: ~21 tasks, each 2-5 minutes.
