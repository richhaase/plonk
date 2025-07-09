// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"
)

func TestConfigAdapter_GetDotfileTargets(t *testing.T) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Dotfiles: []DotfileEntry{
			{Source: "zshrc", Destination: "~/.zshrc"},
			{Source: "gitconfig", Destination: "~/.gitconfig"},
			{Source: "config/nvim/", Destination: "~/.config/nvim/"},
			{Source: "", Destination: "~/.vimrc"}, // Source should be inferred
			{Source: "bashrc", Destination: ""},   // Destination should be inferred
		},
	}
	
	adapter := NewConfigAdapter(config)
	targets := adapter.GetDotfileTargets()
	
	// Should have 5 entries
	if len(targets) != 5 {
		t.Errorf("Expected 5 dotfile targets, got %d", len(targets))
	}
	
	// Test specific mappings
	expectedTargets := map[string]string{
		"zshrc":        "~/.zshrc",
		"gitconfig":    "~/.gitconfig",
		"config/nvim/": "~/.config/nvim/",
		"dot_vimrc":    "~/.vimrc",      // Should be inferred from destination
		"bashrc":       "~/.bashrc",     // Should be inferred from source
	}
	
	for source, expectedDest := range expectedTargets {
		if actualDest, exists := targets[source]; !exists {
			t.Errorf("Expected source %s to exist in targets", source)
		} else if actualDest != expectedDest {
			t.Errorf("Expected source %s to map to %s, got %s", source, expectedDest, actualDest)
		}
	}
}

func TestConfigAdapter_GetPackagesForManager(t *testing.T) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Homebrew: []HomebrewPackage{
			{Name: "git"},
			{Name: "curl"},
			{Name: "htop"},
			{Name: "firefox"},
			{Name: "vscode"},
		},
		NPM: []NPMPackage{
			{Name: "typescript"},
			{Name: "prettier"},
		},
	}
	
	adapter := NewConfigAdapter(config)
	
	t.Run("homebrew packages", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("homebrew")
		if err != nil {
			t.Fatalf("GetPackagesForManager(homebrew) failed: %v", err)
		}
		
		// Should have 5 packages
		if len(packages) != 5 {
			t.Errorf("Expected 5 homebrew packages, got %d", len(packages))
		}
		
		// Check that all expected packages are present
		expectedNames := map[string]bool{
			"git":      true,
			"curl":     true,
			"htop":     true,
			"firefox":  true,
			"vscode":   true,
		}
		
		for _, pkg := range packages {
			if !expectedNames[pkg.Name] {
				t.Errorf("Unexpected package: %s", pkg.Name)
			}
			delete(expectedNames, pkg.Name)
		}
		
		if len(expectedNames) > 0 {
			t.Errorf("Missing expected packages: %v", expectedNames)
		}
	})
	
	t.Run("npm packages", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("npm")
		if err != nil {
			t.Fatalf("GetPackagesForManager(npm) failed: %v", err)
		}
		
		// Should have 2 npm packages
		if len(packages) != 2 {
			t.Errorf("Expected 2 npm packages, got %d", len(packages))
		}
		
		// Check that all expected packages are present
		expectedNames := map[string]bool{
			"typescript": true,
			"prettier":   true,
		}
		
		for _, pkg := range packages {
			if !expectedNames[pkg.Name] {
				t.Errorf("Unexpected package: %s", pkg.Name)
			}
			delete(expectedNames, pkg.Name)
		}
		
		if len(expectedNames) > 0 {
			t.Errorf("Missing expected packages: %v", expectedNames)
		}
	})
	
	t.Run("unknown manager", func(t *testing.T) {
		_, err := adapter.GetPackagesForManager("unknown")
		if err == nil {
			t.Error("Expected error for unknown package manager")
		}
	})
}

func TestStatePackageConfigAdapter(t *testing.T) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Homebrew: []HomebrewPackage{
			{Name: "git"},
			{Name: "curl"},
		},
		NPM: []NPMPackage{
			{Name: "typescript"},
		},
	}
	
	configAdapter := NewConfigAdapter(config)
	stateAdapter := NewStatePackageConfigAdapter(configAdapter)
	
	t.Run("homebrew packages", func(t *testing.T) {
		packages, err := stateAdapter.GetPackagesForManager("homebrew")
		if err != nil {
			t.Fatalf("GetPackagesForManager(homebrew) failed: %v", err)
		}
		
		// Should have 2 packages
		if len(packages) != 2 {
			t.Errorf("Expected 2 homebrew packages, got %d", len(packages))
		}
		
		// Check that packages are correctly converted to state.PackageConfigItem
		expectedNames := map[string]bool{
			"git":  true,
			"curl": true,
		}
		
		for _, pkg := range packages {
			if !expectedNames[pkg.Name] {
				t.Errorf("Unexpected package: %s", pkg.Name)
			}
			delete(expectedNames, pkg.Name)
		}
		
		if len(expectedNames) > 0 {
			t.Errorf("Missing expected packages: %v", expectedNames)
		}
	})
	
	t.Run("npm packages", func(t *testing.T) {
		packages, err := stateAdapter.GetPackagesForManager("npm")
		if err != nil {
			t.Fatalf("GetPackagesForManager(npm) failed: %v", err)
		}
		
		// Should have 1 package
		if len(packages) != 1 {
			t.Errorf("Expected 1 npm package, got %d", len(packages))
		}
		
		if packages[0].Name != "typescript" {
			t.Errorf("Expected package name 'typescript', got %s", packages[0].Name)
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
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Dotfiles: []DotfileEntry{
			{Source: "zshrc", Destination: "~/.zshrc"},
			{Source: "gitconfig", Destination: "~/.gitconfig"},
			{Source: "config/nvim/", Destination: "~/.config/nvim/"},
		},
	}
	
	configAdapter := NewConfigAdapter(config)
	stateAdapter := NewStateDotfileConfigAdapter(configAdapter)
	
	targets := stateAdapter.GetDotfileTargets()
	
	// Should have 3 entries
	if len(targets) != 3 {
		t.Errorf("Expected 3 dotfile targets, got %d", len(targets))
	}
	
	// Test specific mappings
	expectedTargets := map[string]string{
		"zshrc":        "~/.zshrc",
		"gitconfig":    "~/.gitconfig",
		"config/nvim/": "~/.config/nvim/",
	}
	
	for source, expectedDest := range expectedTargets {
		if actualDest, exists := targets[source]; !exists {
			t.Errorf("Expected source %s to exist in targets", source)
		} else if actualDest != expectedDest {
			t.Errorf("Expected source %s to map to %s, got %s", source, expectedDest, actualDest)
		}
	}
}

func TestConfigAdapter_EmptyConfig(t *testing.T) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
	}
	
	adapter := NewConfigAdapter(config)
	
	t.Run("empty dotfiles", func(t *testing.T) {
		targets := adapter.GetDotfileTargets()
		if len(targets) != 0 {
			t.Errorf("Expected 0 dotfile targets for empty config, got %d", len(targets))
		}
	})
	
	t.Run("empty homebrew packages", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("homebrew")
		if err != nil {
			t.Fatalf("GetPackagesForManager(homebrew) failed: %v", err)
		}
		
		if len(packages) != 0 {
			t.Errorf("Expected 0 homebrew packages for empty config, got %d", len(packages))
		}
	})
	
	t.Run("empty npm packages", func(t *testing.T) {
		packages, err := adapter.GetPackagesForManager("npm")
		if err != nil {
			t.Fatalf("GetPackagesForManager(npm) failed: %v", err)
		}
		
		if len(packages) != 0 {
			t.Errorf("Expected 0 npm packages for empty config, got %d", len(packages))
		}
	})
}

func TestConfigAdapter_InterfaceCompliance(t *testing.T) {
	config := &Config{
		Settings: Settings{DefaultManager: "homebrew"},
	}
	
	adapter := NewConfigAdapter(config)
	
	// Test that ConfigAdapter implements DotfileConfigReader
	var _ DotfileConfigReader = adapter
	
	// Test that ConfigAdapter implements PackageConfigReader
	var _ PackageConfigReader = adapter
	
	// Test that StatePackageConfigAdapter implements the state interface
	statePackageAdapter := NewStatePackageConfigAdapter(adapter)
	// We can't easily test this without importing state package here, but compilation will catch issues
	_ = statePackageAdapter
	
	// Test that StateDotfileConfigAdapter implements the state interface
	stateDotfileAdapter := NewStateDotfileConfigAdapter(adapter)
	// We can't easily test this without importing state package here, but compilation will catch issues
	_ = stateDotfileAdapter
}