package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_BasicTOML(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.toml")

	// Write a basic config file
	configContent := `[settings]
default_manager = "homebrew"

[packages.neovim]
config_files = [
  { source = "config/nvim/", target = "~/.config/nvim/" }
]

[packages.nodejs]
manager = "asdf"
version = "24.2.0"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// RED: This should fail because LoadConfig doesn't exist yet
	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify settings
	if config.Settings.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager 'homebrew', got '%s'", config.Settings.DefaultManager)
	}

	// Verify neovim package
	neovim, exists := config.Packages["neovim"]
	if !exists {
		t.Fatal("neovim package not found")
	}
	
	if neovim.Manager != "homebrew" {
		t.Errorf("Expected neovim manager 'homebrew', got '%s'", neovim.Manager)
	}
	
	if len(neovim.ConfigFiles) != 1 {
		t.Errorf("Expected 1 config file, got %d", len(neovim.ConfigFiles))
	}

	// Verify nodejs package
	nodejs, exists := config.Packages["nodejs"]
	if !exists {
		t.Fatal("nodejs package not found")
	}
	
	if nodejs.Manager != "asdf" {
		t.Errorf("Expected nodejs manager 'asdf', got '%s'", nodejs.Manager)
	}
	
	if nodejs.Version != "24.2.0" {
		t.Errorf("Expected nodejs version '24.2.0', got '%s'", nodejs.Version)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	
	_, err := LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestLoadConfig_WithLocalOverrides(t *testing.T) {
	tempDir := t.TempDir()
	
	// Write main config
	mainConfig := `[settings]
default_manager = "homebrew"

[packages.neovim]
config_files = [
  { source = "config/nvim/", target = "~/.config/nvim/" }
]
`
	
	// Write local config with overrides
	localConfig := `[packages.neovim]
manager = "homebrew"

[packages.local-tool]
manager = "npm"
name = "some-local-tool"
`

	err := os.WriteFile(filepath.Join(tempDir, "plonk.toml"), []byte(mainConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write main config: %v", err)
	}
	
	err = os.WriteFile(filepath.Join(tempDir, "plonk.local.toml"), []byte(localConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write local config: %v", err)
	}

	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Should still have default manager from main config
	if config.Settings.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager 'homebrew', got '%s'", config.Settings.DefaultManager)
	}

	// Should have both packages
	if len(config.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(config.Packages))
	}

	// Check local package
	localTool, exists := config.Packages["local-tool"]
	if !exists {
		t.Fatal("local-tool package not found")
	}
	
	if localTool.Manager != "npm" {
		t.Errorf("Expected local-tool manager 'npm', got '%s'", localTool.Manager)
	}
}

func TestPackageValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.toml")

	// Test ASDF package without version should fail
	configContent := `[settings]
default_manager = "homebrew"

[packages.nodejs]
manager = "asdf"
# Missing version for ASDF package
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err = LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for ASDF package without version")
	}
}