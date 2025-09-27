// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	pkgs "github.com/richhaase/plonk/internal/resources/packages"
	"github.com/stretchr/testify/assert"
)

// --- Fake package manager for tests ---

// Global knobs for the fake manager behavior (kept simple: tests here do not run in parallel)
var fakeAvailable = true
var fakeFail = map[string]bool{}

type fakeManager struct{}

func (f *fakeManager) IsAvailable(ctx context.Context) (bool, error)       { return fakeAvailable, nil }
func (f *fakeManager) ListInstalled(ctx context.Context) ([]string, error) { return []string{}, nil }
func (f *fakeManager) Install(ctx context.Context, name string) error {
	if fakeFail[name] {
		return os.ErrPermission
	}
	return nil
}
func (f *fakeManager) Uninstall(ctx context.Context, name string) error           { return nil }
func (f *fakeManager) IsInstalled(ctx context.Context, name string) (bool, error) { return false, nil }
func (f *fakeManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (f *fakeManager) Info(ctx context.Context, name string) (*pkgs.PackageInfo, error) {
	return &pkgs.PackageInfo{Name: name, Manager: "fake", Installed: false}, nil
}
func (f *fakeManager) Search(ctx context.Context, query string) ([]string, error) {
	return []string{}, nil
}
func (f *fakeManager) CheckHealth(ctx context.Context) (*pkgs.HealthCheck, error) {
	return &pkgs.HealthCheck{Name: "fake", Category: "package-manager", Status: "PASS"}, nil
}
func (f *fakeManager) SelfInstall(ctx context.Context) error                { return nil }
func (f *fakeManager) Upgrade(ctx context.Context, packages []string) error { return nil }
func (f *fakeManager) Dependencies() []string                               { return nil }

func registerFakeManager() {
	pkgs.RegisterManager("fake", func() pkgs.PackageManager { return &fakeManager{} })
}

// --- Helpers ---

func writeLockPackage(t *testing.T, configDir, manager, name, version string) {
	t.Helper()
	svc := lock.NewYAMLLockService(configDir)
	meta := map[string]interface{}{"manager": manager, "name": name, "version": version}
	if err := svc.AddPackage(manager, name, version, meta); err != nil {
		t.Fatalf("failed to add package to lock: %v", err)
	}
}

func writeDotfileSource(t *testing.T, configDir, name, contents string) {
	t.Helper()
	path := filepath.Join(configDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write dotfile source: %v", err)
	}
}

// --- Tests ---

func TestApply_Combined_Success(t *testing.T) {
	registerFakeManager()
	fakeAvailable = true
	fakeFail = map[string]bool{}

	configDir := t.TempDir()
	homeDir := t.TempDir()

	// Packages: two fake packages, both missing -> installed
	writeLockPackage(t, configDir, "fake", "pkg1", "1.0.0")
	writeLockPackage(t, configDir, "fake", "pkg2", "2.0.0")

	// Dotfiles: one source present in configDir, not in homeDir -> added
	writeDotfileSource(t, configDir, "zshrc", "export TEST=1\n")

	cfg := config.LoadWithDefaults(configDir)
	orch := New(
		WithConfig(cfg),
		WithConfigDir(configDir),
		WithHomeDir(homeDir),
		WithDryRun(false),
	)

	res, err := orch.Apply(context.Background())
	assert.NoError(t, err)
	assert.True(t, res.Success)
	if assert.NotNil(t, res.Packages) {
		assert.Equal(t, 2, res.Packages.TotalInstalled)
		assert.Equal(t, 0, res.Packages.TotalFailed)
	}
	if assert.NotNil(t, res.Dotfiles) {
		assert.Equal(t, 1, res.Dotfiles.Summary.Added)
		assert.Equal(t, 0, res.Dotfiles.Summary.Failed)
	}
}

func TestApply_PackagesOnly(t *testing.T) {
	registerFakeManager()
	fakeAvailable = true
	fakeFail = map[string]bool{}

	configDir := t.TempDir()
	homeDir := t.TempDir()

	writeLockPackage(t, configDir, "fake", "pkg-only", "1.0.0")

	cfg := config.LoadWithDefaults(configDir)
	orch := New(
		WithConfig(cfg),
		WithConfigDir(configDir),
		WithHomeDir(homeDir),
		WithPackagesOnly(true),
	)

	res, err := orch.Apply(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, res.Packages)
	assert.Nil(t, res.Dotfiles)
}

func TestApply_DotfilesOnly(t *testing.T) {
	registerFakeManager() // not used here but safe

	configDir := t.TempDir()
	homeDir := t.TempDir()

	writeDotfileSource(t, configDir, "gitconfig", "[user]\n\tname = Test\n")

	cfg := config.LoadWithDefaults(configDir)
	orch := New(
		WithConfig(cfg),
		WithConfigDir(configDir),
		WithHomeDir(homeDir),
		WithDotfilesOnly(true),
	)

	res, err := orch.Apply(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, res.Packages)
	if assert.NotNil(t, res.Dotfiles) {
		assert.Equal(t, 1, res.Dotfiles.Summary.Added)
	}
}

func TestApply_DryRun(t *testing.T) {
	registerFakeManager()
	fakeAvailable = true
	fakeFail = map[string]bool{}

	configDir := t.TempDir()
	homeDir := t.TempDir()

	writeLockPackage(t, configDir, "fake", "a", "0.1.0")
	writeLockPackage(t, configDir, "fake", "b", "0.2.0")
	writeDotfileSource(t, configDir, "bashrc", "export DRY=1\n")

	cfg := config.LoadWithDefaults(configDir)
	orch := New(
		WithConfig(cfg),
		WithConfigDir(configDir),
		WithHomeDir(homeDir),
		WithDryRun(true),
	)

	res, err := orch.Apply(context.Background())
	assert.NoError(t, err)
	assert.True(t, res.Success)
	if assert.NotNil(t, res.Packages) {
		assert.Equal(t, 2, res.Packages.TotalWouldInstall)
		assert.Equal(t, 0, res.Packages.TotalInstalled)
	}
	if assert.NotNil(t, res.Dotfiles) {
		assert.Equal(t, 1, res.Dotfiles.Summary.Added)
	}
}

func TestApply_PackageError_Propagates(t *testing.T) {
	registerFakeManager()

	configDir := t.TempDir()
	homeDir := t.TempDir()

	// Force packages.Apply to fail by writing a lock file with a bad version
	bad := []byte("version: 999\nresources: []\n")
	if err := os.WriteFile(filepath.Join(configDir, lock.LockFileName), bad, 0o644); err != nil {
		t.Fatalf("failed to write bad lock: %v", err)
	}

	// Dotfiles still have work to do (should succeed)
	writeDotfileSource(t, configDir, "vimrc", "set number\n")

	cfg := config.LoadWithDefaults(configDir)
	orch := New(
		WithConfig(cfg),
		WithConfigDir(configDir),
		WithHomeDir(homeDir),
	)

	res, err := orch.Apply(context.Background())
	assert.Error(t, err)
	assert.True(t, res.HasErrors())
	assert.Greater(t, len(res.PackageErrors), 0)
	// Dotfiles should still report success on their side
	if assert.NotNil(t, res.Dotfiles) {
		assert.Equal(t, 1, res.Dotfiles.Summary.Added)
	}
}
