// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"
)

func TestConfigAdapter_GetDotfileTargets(t *testing.T) {
	config := &Config{
		Settings: &Settings{
			DefaultManager: StringPtr("homebrew"),
		},
	}

	adapter := NewConfigAdapter(config)
	targets := adapter.GetDotfileTargets()

	// We can't predict exact dotfiles, but we can verify the function works
	if targets == nil {
		t.Error("GetDotfileTargets returned nil")
	}

	// Log what was discovered for debugging
	t.Logf("Auto-discovered %d dotfiles:", len(targets))
	for source, target := range targets {
		t.Logf("  %s -> %s", source, target)
	}
}

func TestConfigAdapter_GetPackagesForManager(t *testing.T) {
	config := &Config{
		Settings: &Settings{
			DefaultManager: StringPtr("homebrew"),
		},
	}

	adapter := NewConfigAdapter(config)

	t.Run("homebrew packages (empty - now in lock file)", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("homebrew")
		if err != nil {
			t.Fatalf("GetPackagesForManager(homebrew) failed: %v", err)
		}

		if len(packages) != 0 {
			t.Errorf("Expected 0 homebrew packages (now in lock file), got %d", len(packages))
		}
	})

	t.Run("npm packages (empty - now in lock file)", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("npm")
		if err != nil {
			t.Fatalf("GetPackagesForManager(npm) failed: %v", err)
		}

		if len(packages) != 0 {
			t.Errorf("Expected 0 npm packages (now in lock file), got %d", len(packages))
		}
	})

	t.Run("unknown manager", func(t *testing.T) {
		_, err := adapter.GetPackagesForManager("unknown")
		if err == nil {
			t.Error("Expected error for unknown package manager")
		}
	})

	t.Run("cargo packages (empty - now in lock file)", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("cargo")
		if err != nil {
			t.Fatalf("GetPackagesForManager(cargo) failed: %v", err)
		}

		if len(packages) != 0 {
			t.Errorf("Expected 0 cargo packages (now in lock file), got %d", len(packages))
		}
	})
}

func TestStatePackageConfigAdapter(t *testing.T) {
	config := &Config{
		Settings: &Settings{
			DefaultManager: StringPtr("homebrew"),
		},
	}

	adapter := NewConfigAdapter(config)
	stateAdapter := NewStatePackageConfigAdapter(adapter)

	t.Run("homebrew packages (empty - now in lock file)", func(t *testing.T) {
		packages, err := stateAdapter.GetPackagesForManager("homebrew")
		if err != nil {
			t.Fatalf("GetPackagesForManager(homebrew) failed: %v", err)
		}

		if len(packages) != 0 {
			t.Errorf("Expected 0 homebrew packages (now in lock file), got %d", len(packages))
		}
	})

	t.Run("npm packages (empty - now in lock file)", func(t *testing.T) {
		packages, err := stateAdapter.GetPackagesForManager("npm")
		if err != nil {
			t.Fatalf("GetPackagesForManager(npm) failed: %v", err)
		}

		if len(packages) != 0 {
			t.Errorf("Expected 0 npm packages (now in lock file), got %d", len(packages))
		}
	})

	t.Run("unknown manager", func(t *testing.T) {
		_, err := stateAdapter.GetPackagesForManager("unknown")
		if err == nil {
			t.Error("Expected error for unknown package manager")
		}
	})
}

func TestStateDotfileConfigAdapter(t *testing.T) {
	config := &Config{
		Settings: &Settings{
			DefaultManager: StringPtr("homebrew"),
		},
	}

	configAdapter := NewConfigAdapter(config)
	adapter := NewStateDotfileConfigAdapter(configAdapter)

	// Test GetDotfileTargets
	targets := adapter.GetDotfileTargets()
	if targets == nil {
		t.Error("GetDotfileTargets returned nil")
	}

	// Test GetIgnorePatterns
	patterns := adapter.GetIgnorePatterns()
	if patterns == nil {
		t.Error("GetIgnorePatterns returned nil")
	}

	// Test GetExpandDirectories
	dirs := adapter.GetExpandDirectories()
	if dirs == nil {
		t.Error("GetExpandDirectories returned nil")
	}
}

func TestConfigAdapter_EmptyConfig(t *testing.T) {
	config := &Config{
		Settings: &Settings{
			DefaultManager: StringPtr("homebrew"),
		},
	}

	adapter := NewConfigAdapter(config)

	// Test empty package lists
	for _, manager := range []string{"homebrew", "npm", "cargo"} {
		packages, err := adapter.GetPackagesForManager(manager)
		if err != nil {
			t.Errorf("GetPackagesForManager(%s) failed: %v", manager, err)
		}
		if len(packages) != 0 {
			t.Errorf("Expected 0 packages for %s (now in lock file), got %d", manager, len(packages))
		}
	}

	// Test dotfile targets (should still work)
	targets := adapter.GetDotfileTargets()
	if targets == nil {
		t.Error("GetDotfileTargets returned nil")
	}
}

func TestConfigAdapter_InterfaceCompliance(t *testing.T) {
	config := &Config{
		Settings: &Settings{
			DefaultManager: StringPtr("homebrew"),
		},
	}

	adapter := NewConfigAdapter(config)

	// Verify adapter implements the expected interfaces
	var _ PackageConfigReader = adapter
	var _ DotfileConfigReader = adapter
}
