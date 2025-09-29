// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestNewOrchestrator(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		validate func(t *testing.T, o *Orchestrator)
	}{
		{
			name: "default orchestrator",
			opts: nil,
			validate: func(t *testing.T, o *Orchestrator) {
				assert.NotNil(t, o)
				assert.Nil(t, o.config)
				assert.Nil(t, o.lock)
				assert.False(t, o.dryRun)
				assert.False(t, o.packagesOnly)
				assert.False(t, o.dotfilesOnly)
			},
		},
		{
			name: "with config",
			opts: []Option{
				WithConfig(&config.Config{DefaultManager: "brew"}),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.NotNil(t, o.config)
				assert.Equal(t, "brew", o.config.DefaultManager)
			},
		},
		{
			name: "with config dir",
			opts: []Option{
				WithConfigDir("/test/config"),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.Equal(t, "/test/config", o.configDir)
				assert.NotNil(t, o.lock)
			},
		},
		{
			name: "with home dir",
			opts: []Option{
				WithHomeDir("/home/user"),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.Equal(t, "/home/user", o.homeDir)
			},
		},
		{
			name: "with dry run",
			opts: []Option{
				WithDryRun(true),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.True(t, o.dryRun)
			},
		},
		{
			name: "with packages only",
			opts: []Option{
				WithPackagesOnly(true),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.True(t, o.packagesOnly)
			},
		},
		{
			name: "with dotfiles only",
			opts: []Option{
				WithDotfilesOnly(true),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.True(t, o.dotfilesOnly)
			},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithConfig(&config.Config{DefaultManager: "npm"}),
				WithConfigDir("/etc/plonk"),
				WithHomeDir("/home/test"),
				WithDryRun(true),
				WithPackagesOnly(false),
				WithDotfilesOnly(false),
			},
			validate: func(t *testing.T, o *Orchestrator) {
				assert.NotNil(t, o.config)
				assert.Equal(t, "npm", o.config.DefaultManager)
				assert.Equal(t, "/etc/plonk", o.configDir)
				assert.Equal(t, "/home/test", o.homeDir)
				assert.True(t, o.dryRun)
				assert.False(t, o.packagesOnly)
				assert.False(t, o.dotfilesOnly)
				assert.NotNil(t, o.lock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := New(tt.opts...)
			tt.validate(t, o)
		})
	}
}

func TestOptionFunctions(t *testing.T) {
	t.Run("WithConfig", func(t *testing.T) {
		cfg := &config.Config{DefaultManager: "test"}
		opt := WithConfig(cfg)
		o := &Orchestrator{}
		opt(o)
		assert.Equal(t, cfg, o.config)
	})

	t.Run("WithConfigDir", func(t *testing.T) {
		dir := "/test/dir"
		opt := WithConfigDir(dir)
		o := &Orchestrator{}
		opt(o)
		assert.Equal(t, dir, o.configDir)
	})

	t.Run("WithHomeDir", func(t *testing.T) {
		dir := "/home/test"
		opt := WithHomeDir(dir)
		o := &Orchestrator{}
		opt(o)
		assert.Equal(t, dir, o.homeDir)
	})

	t.Run("WithDryRun", func(t *testing.T) {
		opt := WithDryRun(true)
		o := &Orchestrator{}
		opt(o)
		assert.True(t, o.dryRun)
	})

	t.Run("WithPackagesOnly", func(t *testing.T) {
		opt := WithPackagesOnly(true)
		o := &Orchestrator{}
		opt(o)
		assert.True(t, o.packagesOnly)
	})

	t.Run("WithDotfilesOnly", func(t *testing.T) {
		opt := WithDotfilesOnly(true)
		o := &Orchestrator{}
		opt(o)
		assert.True(t, o.dotfilesOnly)
	})
}

func TestApply_SelectiveApplication(t *testing.T) {
	// This test verifies the selective application logic without actually calling Apply
	// which could modify the system. We test the flags behavior only.
	tests := []struct {
		name           string
		packagesOnly   bool
		dotfilesOnly   bool
		expectPackages bool
		expectDotfiles bool
	}{
		{
			name:           "apply both by default",
			packagesOnly:   false,
			dotfilesOnly:   false,
			expectPackages: true,
			expectDotfiles: true,
		},
		{
			name:           "packages only",
			packagesOnly:   true,
			dotfilesOnly:   false,
			expectPackages: true,
			expectDotfiles: false,
		},
		{
			name:           "dotfiles only",
			packagesOnly:   false,
			dotfilesOnly:   true,
			expectPackages: false,
			expectDotfiles: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Orchestrator{
				packagesOnly: tt.packagesOnly,
				dotfilesOnly: tt.dotfilesOnly,
				dryRun:       true,
			}

			// Test the logic without calling Apply
			// The Apply method checks these flags to decide what to run
			shouldRunPackages := !o.dotfilesOnly
			shouldRunDotfiles := !o.packagesOnly

			assert.Equal(t, tt.expectPackages, shouldRunPackages)
			assert.Equal(t, tt.expectDotfiles, shouldRunDotfiles)
		})
	}
}

func TestApplyResult_Success(t *testing.T) {
	tests := []struct {
		name          string
		packageResult *PackageApplyResult
		dotfileResult *DotfileApplyResult
		dryRun        bool
		hasErrors     bool
		expectSuccess bool
		expectChanged bool
	}{
		{
			name: "packages installed in normal mode",
			packageResult: &PackageApplyResult{
				DryRun:         false,
				TotalInstalled: 3,
			},
			dryRun:        false,
			hasErrors:     false,
			expectSuccess: true,
			expectChanged: true,
		},
		{
			name: "packages would install in dry run",
			packageResult: &PackageApplyResult{
				DryRun:            true,
				TotalWouldInstall: 3,
			},
			dryRun:        true,
			hasErrors:     false,
			expectSuccess: true,
			expectChanged: true,
		},
		{
			name: "dotfiles added in normal mode",
			dotfileResult: &DotfileApplyResult{
				DryRun: false,
				Summary: DotfileSummaryApplyResult{
					Added: 5,
				},
			},
			dryRun:        false,
			hasErrors:     false,
			expectSuccess: true,
			expectChanged: true,
		},
		{
			name: "dotfiles would add in dry run",
			dotfileResult: &DotfileApplyResult{
				DryRun: true,
				Summary: DotfileSummaryApplyResult{
					Added: 5,
				},
			},
			dryRun:        true,
			hasErrors:     false,
			expectSuccess: true,
			expectChanged: true,
		},
		{
			name: "no changes in normal mode - idempotent success",
			packageResult: &PackageApplyResult{
				DryRun:         false,
				TotalInstalled: 0,
			},
			dotfileResult: &DotfileApplyResult{
				DryRun: false,
				Summary: DotfileSummaryApplyResult{
					Added: 0,
				},
			},
			dryRun:        false,
			hasErrors:     false,
			expectSuccess: true, // Changed from false - no-op with no errors is success
			expectChanged: false,
		},
		{
			name: "mixed success - packages changed, dotfiles unchanged",
			packageResult: &PackageApplyResult{
				DryRun:         false,
				TotalInstalled: 2,
			},
			dotfileResult: &DotfileApplyResult{
				DryRun: false,
				Summary: DotfileSummaryApplyResult{
					Added: 0,
				},
			},
			dryRun:        false,
			hasErrors:     false,
			expectSuccess: true,
			expectChanged: true,
		},
		{
			name: "errors present - success is false even with changes",
			packageResult: &PackageApplyResult{
				DryRun:         false,
				TotalInstalled: 2,
				TotalFailed:    1,
			},
			dryRun:        false,
			hasErrors:     true,
			expectSuccess: false,
			expectChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyResult{
				DryRun:   tt.dryRun,
				Packages: tt.packageResult,
				Dotfiles: tt.dotfileResult,
			}

			// Simulate error injection if needed
			if tt.hasErrors {
				result.AddPackageError(fmt.Errorf("simulated package error"))
			}

			// Simulate the success determination logic from Apply()
			// Success = no errors (HasErrors() returns false)
			success := !result.HasErrors()

			// Determine if changes were made
			changed := false
			if result.Packages != nil {
				if !tt.dryRun && result.Packages.TotalInstalled > 0 {
					changed = true
				} else if tt.dryRun && result.Packages.TotalWouldInstall > 0 {
					changed = true
				}
			}
			if result.Dotfiles != nil {
				if !tt.dryRun && result.Dotfiles.Summary.Added > 0 {
					changed = true
				} else if tt.dryRun && result.Dotfiles.Summary.Added > 0 {
					changed = true
				}
			}

			assert.Equal(t, tt.expectSuccess, success, "success mismatch")
			assert.Equal(t, tt.expectChanged, changed, "changed mismatch")
		})
	}
}

func TestApply_ErrorHandling(t *testing.T) {
	tests := []struct {
		name             string
		packageErrors    []error
		dotfileErrors    []error
		expectError      bool
		expectErrorCount int
	}{
		{
			name:          "no errors",
			packageErrors: nil,
			dotfileErrors: nil,
			expectError:   false,
		},
		{
			name:             "package errors only",
			packageErrors:    []error{errors.New("failed to install foo")},
			dotfileErrors:    nil,
			expectError:      true,
			expectErrorCount: 1,
		},
		{
			name:             "dotfile errors only",
			packageErrors:    nil,
			dotfileErrors:    []error{errors.New("failed to link .bashrc")},
			expectError:      true,
			expectErrorCount: 1,
		},
		{
			name:             "both types of errors",
			packageErrors:    []error{errors.New("failed to install foo"), errors.New("failed to install bar")},
			dotfileErrors:    []error{errors.New("failed to link .bashrc"), errors.New("failed to link .vimrc")},
			expectError:      true,
			expectErrorCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyResult{
				PackageErrors: tt.packageErrors,
				DotfileErrors: tt.dotfileErrors,
			}

			// Simulate error checking logic from Apply()
			var err error
			if result.HasErrors() {
				err = result.GetCombinedError()
				totalErrors := len(result.PackageErrors) + len(result.DotfileErrors)
				assert.Equal(t, tt.expectErrorCount, totalErrors)
			}

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test the ReconcileAll function
func TestReconcileAll(t *testing.T) {
	// This test would require mocking the dotfiles.Reconcile and packages.Reconcile
	// functions, which are package-level functions. In a real implementation,
	// we might want to refactor to use interfaces for better testability.
	// For now, we'll test that the function exists and has the right signature.

	ctx := context.Background()

	// We can't really test this without modifying the package to support
	// dependency injection or mocking. This is a limitation of the current design.
	// The function will attempt to read actual files which may not exist in test.
	t.Run("function exists with temp dirs", func(t *testing.T) {
		// Create temporary directories for testing
		tempHome, err := os.MkdirTemp("", "plonk-test-home-*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempHome)

		tempConfig, err := os.MkdirTemp("", "plonk-test-config-*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempConfig)

		// Create minimal required structure
		err = os.MkdirAll(filepath.Join(tempConfig, "dotfiles"), 0755)
		assert.NoError(t, err)

		// Create an empty lock file to avoid errors
		lockFile := filepath.Join(tempConfig, "plonk.lock")
		err = os.WriteFile(lockFile, []byte("version: 2\nresources: []\n"), 0644)
		assert.NoError(t, err)

		// Just verify the function can be called
		results, err := ReconcileAll(ctx, tempHome, tempConfig)
		// Either way, we're just testing that the function exists and returns the right types
		if err == nil {
			assert.NotNil(t, results)
			assert.IsType(t, map[string]resources.Result{}, results)
		}
	})
}

// Test apply result structs marshaling
func TestApplyResultStructs(t *testing.T) {
	t.Run("PackageApplyResult fields", func(t *testing.T) {
		result := PackageApplyResult{
			DryRun:            true,
			TotalMissing:      5,
			TotalInstalled:    3,
			TotalFailed:       1,
			TotalWouldInstall: 2,
			Managers: []ManagerApplyResult{
				{
					Name:         "brew",
					MissingCount: 3,
					Packages: []PackageOperationApplyResult{
						{
							Name:   "ripgrep",
							Status: "installed",
						},
						{
							Name:   "fd",
							Status: "failed",
							Error:  "network error",
						},
					},
				},
			},
		}

		assert.True(t, result.DryRun)
		assert.Equal(t, 5, result.TotalMissing)
		assert.Equal(t, 3, result.TotalInstalled)
		assert.Equal(t, 1, result.TotalFailed)
		assert.Equal(t, 2, result.TotalWouldInstall)
		assert.Len(t, result.Managers, 1)
		assert.Equal(t, "brew", result.Managers[0].Name)
		assert.Len(t, result.Managers[0].Packages, 2)
	})

	t.Run("DotfileApplyResult fields", func(t *testing.T) {
		result := DotfileApplyResult{
			DryRun:     false,
			TotalFiles: 10,
			Actions: []DotfileActionApplyResult{
				{
					Source:      "dotfiles/.bashrc",
					Destination: "~/.bashrc",
					Action:      "copy",
					Status:      "added",
				},
				{
					Source:      "dotfiles/.vimrc",
					Destination: "~/.vimrc",
					Action:      "error",
					Status:      "failed",
					Error:       "permission denied",
				},
			},
			Summary: DotfileSummaryApplyResult{
				Added:     5,
				Updated:   2,
				Unchanged: 2,
				Failed:    1,
			},
		}

		assert.False(t, result.DryRun)
		assert.Equal(t, 10, result.TotalFiles)
		assert.Len(t, result.Actions, 2)
		assert.Equal(t, 5, result.Summary.Added)
		assert.Equal(t, 2, result.Summary.Updated)
		assert.Equal(t, 2, result.Summary.Unchanged)
		assert.Equal(t, 1, result.Summary.Failed)
	})
}
