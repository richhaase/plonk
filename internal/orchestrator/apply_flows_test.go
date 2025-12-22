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

// --- Test helpers for mock executor ---

// setupBrewMockExecutor wires a mock executor for brew commands
func setupBrewMockExecutor(t *testing.T, failInstalls map[string]bool) {
	t.Helper()

	// Configure a mock executor so no real commands run
	mock := &pkgs.MockCommandExecutor{Responses: map[string]pkgs.CommandResponse{}}

	// Make the binary discoverable and available
	mock.Responses["brew --version"] = pkgs.CommandResponse{Output: []byte("Homebrew 4.0.0"), Error: nil}
	mock.Responses["brew list --formula -1"] = pkgs.CommandResponse{Output: []byte(""), Error: nil}

	// Provide install responses; optionally fail select packages
	for _, name := range []string{"pkg1", "pkg2", "pkg-only", "a", "b"} {
		key := "brew install " + name
		if failInstalls != nil && failInstalls[name] {
			mock.Responses[key] = pkgs.CommandResponse{Output: []byte("permission denied"), Error: &pkgs.MockExitError{Code: 1}}
		} else {
			mock.Responses[key] = pkgs.CommandResponse{Output: []byte("installed"), Error: nil}
		}
	}

	// Inject executor for package manager operations
	pkgs.SetDefaultExecutor(mock)
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
	configDir := t.TempDir()
	homeDir := t.TempDir()

	// Packages: two brew packages, both missing -> installed
	writeLockPackage(t, configDir, "brew", "pkg1", "1.0.0")
	writeLockPackage(t, configDir, "brew", "pkg2", "2.0.0")

	// Dotfiles: one source present in configDir, not in homeDir -> added
	writeDotfileSource(t, configDir, "zshrc", "export TEST=1\n")

	cfg := config.LoadWithDefaults(configDir)
	setupBrewMockExecutor(t, nil)
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
	configDir := t.TempDir()
	homeDir := t.TempDir()

	writeLockPackage(t, configDir, "brew", "pkg-only", "1.0.0")

	cfg := config.LoadWithDefaults(configDir)
	setupBrewMockExecutor(t, nil)
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
	configDir := t.TempDir()
	homeDir := t.TempDir()

	writeDotfileSource(t, configDir, "gitconfig", "[user]\n\tname = Test\n")

	cfg := config.LoadWithDefaults(configDir)
	// No package manager needed for this test
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
	configDir := t.TempDir()
	homeDir := t.TempDir()

	writeLockPackage(t, configDir, "brew", "a", "0.1.0")
	writeLockPackage(t, configDir, "brew", "b", "0.2.0")
	writeDotfileSource(t, configDir, "bashrc", "export DRY=1\n")

	cfg := config.LoadWithDefaults(configDir)
	setupBrewMockExecutor(t, nil)
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
	// No package manager needed; failure happens during lock read
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
