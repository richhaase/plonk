// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"testing"

	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/stretchr/testify/assert"
)

func TestPackageResults_Aggregation(t *testing.T) {
	tests := []struct {
		name              string
		managers          []output.ManagerResults
		expectedMissing   int
		expectedInstalled int
		expectedFailed    int
		expectedWould     int
	}{
		{
			name: "single manager with mixed results",
			managers: []output.ManagerResults{
				{
					Name:         "brew",
					MissingCount: 3,
					Packages: []output.PackageOperation{
						{Name: "ripgrep", Status: "installed"},
						{Name: "fd", Status: "installed"},
						{Name: "bat", Status: "failed", Error: "network error"},
					},
				},
			},
			expectedMissing:   3,
			expectedInstalled: 2,
			expectedFailed:    1,
			expectedWould:     0,
		},
		{
			name: "multiple managers",
			managers: []output.ManagerResults{
				{
					Name:         "brew",
					MissingCount: 2,
					Packages: []output.PackageOperation{
						{Name: "ripgrep", Status: "installed"},
						{Name: "fd", Status: "installed"},
					},
				},
				{
					Name:         "npm",
					MissingCount: 3,
					Packages: []output.PackageOperation{
						{Name: "prettier", Status: "installed"},
						{Name: "eslint", Status: "failed", Error: "permission denied"},
						{Name: "typescript", Status: "would-install"},
					},
				},
			},
			expectedMissing:   5,
			expectedInstalled: 3,
			expectedFailed:    1,
			expectedWould:     1,
		},
		{
			name: "dry run results",
			managers: []output.ManagerResults{
				{
					Name:         "brew",
					MissingCount: 4,
					Packages: []output.PackageOperation{
						{Name: "ripgrep", Status: "would-install"},
						{Name: "fd", Status: "would-install"},
						{Name: "bat", Status: "would-install"},
						{Name: "jq", Status: "would-install"},
					},
				},
			},
			expectedMissing:   4,
			expectedInstalled: 0,
			expectedFailed:    0,
			expectedWould:     4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the counting logic from ApplyPackages
			totalMissing := 0
			totalInstalled := 0
			totalFailed := 0
			totalWouldInstall := 0

			for _, manager := range tt.managers {
				totalMissing += manager.MissingCount
				for _, pkg := range manager.Packages {
					switch pkg.Status {
					case "installed":
						totalInstalled++
					case "failed":
						totalFailed++
					case "would-install":
						totalWouldInstall++
					}
				}
			}

			assert.Equal(t, tt.expectedMissing, totalMissing)
			assert.Equal(t, tt.expectedInstalled, totalInstalled)
			assert.Equal(t, tt.expectedFailed, totalFailed)
			assert.Equal(t, tt.expectedWould, totalWouldInstall)
		})
	}
}

func TestDotfileResults_Summary(t *testing.T) {
	tests := []struct {
		name              string
		actions           []output.DotfileOperation
		expectedAdded     int
		expectedUpdated   int
		expectedUnchanged int
		expectedFailed    int
	}{
		{
			name: "mixed results",
			actions: []output.DotfileOperation{
				{Source: ".bashrc", Destination: "~/.bashrc", Action: "copy", Status: "added"},
				{Source: ".vimrc", Destination: "~/.vimrc", Action: "copy", Status: "added"},
				{Source: ".gitconfig", Destination: "~/.gitconfig", Action: "error", Status: "failed", Error: "permission denied"},
			},
			expectedAdded:     2,
			expectedUpdated:   0,
			expectedUnchanged: 0,
			expectedFailed:    1,
		},
		{
			name: "dry run results",
			actions: []output.DotfileOperation{
				{Source: ".bashrc", Destination: "~/.bashrc", Action: "would-copy", Status: "would-add"},
				{Source: ".vimrc", Destination: "~/.vimrc", Action: "would-copy", Status: "would-add"},
				{Source: ".gitconfig", Destination: "~/.gitconfig", Action: "would-copy", Status: "would-add"},
			},
			expectedAdded:     3,
			expectedUpdated:   0,
			expectedUnchanged: 0,
			expectedFailed:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the counting logic from ApplyDotfiles
			summary := output.DotfileSummary{}

			for _, action := range tt.actions {
				switch action.Status {
				case "added", "would-add":
					summary.Added++
				case "updated", "would-update":
					summary.Updated++
				case "unchanged":
					summary.Unchanged++
				case "failed":
					summary.Failed++
				}
			}

			assert.Equal(t, tt.expectedAdded, summary.Added)
			assert.Equal(t, tt.expectedUpdated, summary.Updated)
			assert.Equal(t, tt.expectedUnchanged, summary.Unchanged)
			assert.Equal(t, tt.expectedFailed, summary.Failed)
		})
	}
}

func TestApplyResults_StateMapping(t *testing.T) {
	t.Run("resource state to action mapping", func(t *testing.T) {
		tests := []struct {
			state          dotfiles.ItemState
			expectedAction string
			expectedStatus string
		}{
			{
				state:          dotfiles.StateMissing,
				expectedAction: "copy",
				expectedStatus: "added",
			},
			{
				state:          dotfiles.StateDegraded,
				expectedAction: "copy",
				expectedStatus: "added",
			},
			{
				state:          dotfiles.StateManaged,
				expectedAction: "",
				expectedStatus: "unchanged",
			},
		}

		for _, tt := range tests {
			t.Run(tt.state.String(), func(t *testing.T) {
				// Test items with different states
				item := dotfiles.DotfileItem{
					Name:  ".testfile",
					State: tt.state,
				}

				// Simulate the logic from ApplyDotfiles
				var action, status string
				if item.State == dotfiles.StateMissing || item.State == dotfiles.StateDegraded {
					action = "copy"
					status = "added"
				} else if item.State == dotfiles.StateManaged {
					action = ""
					status = "unchanged"
				}

				if tt.expectedAction != "" {
					assert.Equal(t, tt.expectedAction, action)
				}
				assert.Equal(t, tt.expectedStatus, status)
			})
		}
	})
}

// packageItem is a test-only type for simulating package items
type packageItem struct {
	Name    string
	Manager string
}

func TestApplyPackages_ProgressReporting(t *testing.T) {
	// Test that progress is properly tracked
	missingItems := []packageItem{
		{Name: "ripgrep", Manager: "brew"},
		{Name: "fd", Manager: "brew"},
		{Name: "prettier", Manager: "npm"},
	}

	// Group by manager like ApplyPackages does
	missingByManager := make(map[string][]packageItem)
	for _, item := range missingItems {
		if item.Manager != "" {
			missingByManager[item.Manager] = append(missingByManager[item.Manager], item)
		}
	}

	// Verify grouping
	assert.Len(t, missingByManager, 2)
	assert.Len(t, missingByManager["brew"], 2)
	assert.Len(t, missingByManager["npm"], 1)

	// Simulate progress tracking
	totalMissing := len(missingItems)
	packageIndex := 0

	for _, items := range missingByManager {
		for range items {
			packageIndex++
			assert.LessOrEqual(t, packageIndex, totalMissing)
		}
	}

	assert.Equal(t, totalMissing, packageIndex)
}

func TestApplyDotfiles_ProgressReporting(t *testing.T) {
	// Test that progress is properly tracked for dotfiles
	reconciled := []dotfiles.DotfileItem{
		{Name: "~/.bashrc", State: dotfiles.StateMissing},
		{Name: "~/.vimrc", State: dotfiles.StateDegraded},
		{Name: "~/.gitconfig", State: dotfiles.StateManaged},
		{Name: "~/.zshrc", State: dotfiles.StateMissing},
	}

	// Count items that need to be applied
	applyCount := 0
	for _, item := range reconciled {
		if item.State == dotfiles.StateMissing || item.State == dotfiles.StateDegraded {
			applyCount++
		}
	}

	assert.Equal(t, 3, applyCount)

	// Simulate progress tracking
	dotfileIndex := 0
	for _, item := range reconciled {
		if item.State == dotfiles.StateMissing || item.State == dotfiles.StateDegraded {
			dotfileIndex++
			assert.LessOrEqual(t, dotfileIndex, applyCount)
		}
	}

	assert.Equal(t, applyCount, dotfileIndex)
}

func TestManagerResults_Structure(t *testing.T) {
	// Test the structure of manager results
	result := output.ManagerResults{
		Name:         "brew",
		MissingCount: 3,
		Packages: []output.PackageOperation{
			{
				Name:   "ripgrep",
				Status: "installed",
			},
			{
				Name:   "fd",
				Status: "failed",
				Error:  "connection timeout",
			},
			{
				Name:   "bat",
				Status: "would-install",
			},
		},
	}

	assert.Equal(t, "brew", result.Name)
	assert.Equal(t, 3, result.MissingCount)
	assert.Len(t, result.Packages, 3)

	// Verify each package result
	assert.Equal(t, "ripgrep", result.Packages[0].Name)
	assert.Equal(t, "installed", result.Packages[0].Status)
	assert.Empty(t, result.Packages[0].Error)

	assert.Equal(t, "fd", result.Packages[1].Name)
	assert.Equal(t, "failed", result.Packages[1].Status)
	assert.Equal(t, "connection timeout", result.Packages[1].Error)

	assert.Equal(t, "bat", result.Packages[2].Name)
	assert.Equal(t, "would-install", result.Packages[2].Status)
	assert.Empty(t, result.Packages[2].Error)
}

func TestDotfileOperation_Structure(t *testing.T) {
	// Test the structure of dotfile action results
	actions := []output.DotfileOperation{
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
			Error:       "file exists",
		},
		{
			Source:      "dotfiles/.gitconfig",
			Destination: "~/.gitconfig",
			Action:      "would-copy",
			Status:      "would-add",
		},
	}

	// Verify normal operation
	assert.Equal(t, "dotfiles/.bashrc", actions[0].Source)
	assert.Equal(t, "~/.bashrc", actions[0].Destination)
	assert.Equal(t, "copy", actions[0].Action)
	assert.Equal(t, "added", actions[0].Status)
	assert.Empty(t, actions[0].Error)

	// Verify error case
	assert.Equal(t, "dotfiles/.vimrc", actions[1].Source)
	assert.Equal(t, "~/.vimrc", actions[1].Destination)
	assert.Equal(t, "error", actions[1].Action)
	assert.Equal(t, "failed", actions[1].Status)
	assert.Equal(t, "file exists", actions[1].Error)

	// Verify dry run case
	assert.Equal(t, "dotfiles/.gitconfig", actions[2].Source)
	assert.Equal(t, "~/.gitconfig", actions[2].Destination)
	assert.Equal(t, "would-copy", actions[2].Action)
	assert.Equal(t, "would-add", actions[2].Status)
	assert.Empty(t, actions[2].Error)
}

// Note: The actual ApplyPackages and ApplyDotfiles functions are difficult to test
// without significant refactoring because they directly call package-level functions
// like packages.Reconcile and create new resources. This is a limitation of the
// current design that was identified in the architecture decision document.
// These tests focus on the data structures and logic that can be tested in isolation.
