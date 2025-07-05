package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCommand_NoConfig(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Test - should error when no config exists
	err := runApply([]string{})
	if err == nil {
		t.Error("Expected error when no config file exists")
	}
}

func TestApplyCommand_AllDotfiles(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create source files
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	err = os.WriteFile(filepath.Join(plonkDir, "dot_gitconfig"), []byte("# test gitconfig"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create config directory structure
	configDir := filepath.Join(plonkDir, "config", "nvim")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	err = os.WriteFile(filepath.Join(configDir, "init.vim"), []byte("# test nvim config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create nvim config: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
  - dot_gitconfig

homebrew:
  brews:
    - name: neovim
      config: config/nvim/
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying all dotfiles
	err = runApply([]string{})
	if err != nil {
		t.Fatalf("Apply command failed: %v", err)
	}
	
	// Verify files were created
	if !fileExists(filepath.Join(tempHome, ".zshrc")) {
		t.Error("Expected ~/.zshrc to be created")
	}
	
	if !fileExists(filepath.Join(tempHome, ".gitconfig")) {
		t.Error("Expected ~/.gitconfig to be created")
	}
	
	if !fileExists(filepath.Join(tempHome, ".config", "nvim", "init.vim")) {
		t.Error("Expected ~/.config/nvim/init.vim to be created")
	}
}

func TestApplyCommand_PackageSpecific(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create source files
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create config directory for neovim
	nvimConfigDir := filepath.Join(plonkDir, "config", "nvim")
	err = os.MkdirAll(nvimConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nvim config directory: %v", err)
	}
	
	err = os.WriteFile(filepath.Join(nvimConfigDir, "init.vim"), []byte("# test nvim config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create nvim config: %v", err)
	}
	
	// Create config directory for mcfly
	mcflyConfigDir := filepath.Join(plonkDir, "config", "mcfly")
	err = os.MkdirAll(mcflyConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create mcfly config directory: %v", err)
	}
	
	err = os.WriteFile(filepath.Join(mcflyConfigDir, "config.yaml"), []byte("# test mcfly config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mcfly config: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc

homebrew:
  brews:
    - name: neovim
      config: config/nvim/
    - name: mcfly
      config: config/mcfly/
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying only neovim package config
	err = runApply([]string{"neovim"})
	if err != nil {
		t.Fatalf("Apply command failed: %v", err)
	}
	
	// Verify only neovim config was applied, not global dotfiles or mcfly
	if !fileExists(filepath.Join(tempHome, ".config", "nvim", "init.vim")) {
		t.Error("Expected ~/.config/nvim/init.vim to be created")
	}
	
	if fileExists(filepath.Join(tempHome, ".zshrc")) {
		t.Error("Expected ~/.zshrc NOT to be created when applying package-specific config")
	}
	
	if fileExists(filepath.Join(tempHome, ".config", "mcfly", "config.yaml")) {
		t.Error("Expected ~/.config/mcfly/config.yaml NOT to be created when applying only neovim")
	}
}

func TestApplyCommand_InvalidPackage(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create simple config file
	configContent := `settings:
  default_manager: homebrew

homebrew:
  brews:
    - name: neovim
      config: config/nvim/
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying non-existent package
	err = runApply([]string{"non-existent-package"})
	if err == nil {
		t.Error("Expected error when applying config for non-existent package")
	}
}