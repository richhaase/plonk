package commands

import (
	"os"
	"path/filepath"
	"strings"
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

func TestApplyCommand_ZSHConfiguration(t *testing.T) {
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
	
	// Create config file with ZSH configuration
	configContent := `settings:
  default_manager: homebrew

zsh:
  env_vars:
    EDITOR: nvim
    PAGER: bat
  
  aliases:
    ll: "eza -la"
    cat: "bat"
  
  inits:
    - 'eval "$(starship init zsh)"'
    - 'eval "$(zoxide init zsh)"'
  
  completions:
    - 'source <(kubectl completion zsh)'
  
  shell_options:
    - AUTO_MENU
  
  functions:
    mkcd: 'mkdir -p "$1" && cd "$1"'
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying ZSH configuration
	err = runApply([]string{})
	if err != nil {
		t.Fatalf("Apply command failed: %v", err)
	}
	
	// Verify .zshrc was created and contains expected content
	zshrcPath := filepath.Join(tempHome, ".zshrc")
	if !fileExists(zshrcPath) {
		t.Fatal("Expected ~/.zshrc to be generated")
	}
	
	zshrcContent, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("Failed to read generated .zshrc: %v", err)
	}
	
	zshrcStr := string(zshrcContent)
	
	// Check for generated header
	if !strings.Contains(zshrcStr, "# Generated by plonk") {
		t.Error("Expected generated header in .zshrc")
	}
	
	// Check for aliases
	if !strings.Contains(zshrcStr, "alias ll='eza -la'") {
		t.Error("Expected ll alias in generated .zshrc")
	}
	if !strings.Contains(zshrcStr, "alias cat='bat'") {
		t.Error("Expected cat alias in generated .zshrc")
	}
	
	// Check for inits
	if !strings.Contains(zshrcStr, `eval "$(starship init zsh)"`) {
		t.Error("Expected starship init in generated .zshrc")
	}
	if !strings.Contains(zshrcStr, `eval "$(zoxide init zsh)"`) {
		t.Error("Expected zoxide init in generated .zshrc")
	}
	
	// Check for completions
	if !strings.Contains(zshrcStr, `source <(kubectl completion zsh)`) {
		t.Error("Expected kubectl completion in generated .zshrc")
	}
	
	// Check for shell options
	if !strings.Contains(zshrcStr, "setopt AUTO_MENU") {
		t.Error("Expected AUTO_MENU setopt in generated .zshrc")
	}
	
	// Check for functions
	if !strings.Contains(zshrcStr, "function mkcd() {") {
		t.Error("Expected mkcd function in generated .zshrc")
	}
	
	// Verify .zshenv was created and contains expected content
	zshenvPath := filepath.Join(tempHome, ".zshenv")
	if !fileExists(zshenvPath) {
		t.Fatal("Expected ~/.zshenv to be generated")
	}
	
	zshenvContent, err := os.ReadFile(zshenvPath)
	if err != nil {
		t.Fatalf("Failed to read generated .zshenv: %v", err)
	}
	
	zshenvStr := string(zshenvContent)
	
	// Check for environment variables in .zshenv
	if !strings.Contains(zshenvStr, "export EDITOR='nvim'") {
		t.Error("Expected EDITOR export in generated .zshenv")
	}
	if !strings.Contains(zshenvStr, "export PAGER='bat'") {
		t.Error("Expected PAGER export in generated .zshenv")
	}
	
	// Check that .zshenv doesn't contain aliases (should only be in .zshrc)
	if strings.Contains(zshenvStr, "alias") {
		t.Error("Expected .zshenv to not contain aliases")
	}
}

func TestApplyCommand_ZSHConfiguration_NoEnvVars(t *testing.T) {
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
	
	// Create config file with ZSH configuration but no env vars
	configContent := `settings:
  default_manager: homebrew

zsh:
  aliases:
    ll: "eza -la"
  
  inits:
    - 'eval "$(starship init zsh)"'
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying ZSH configuration
	err = runApply([]string{})
	if err != nil {
		t.Fatalf("Apply command failed: %v", err)
	}
	
	// Verify .zshrc was created
	zshrcPath := filepath.Join(tempHome, ".zshrc")
	if !fileExists(zshrcPath) {
		t.Fatal("Expected ~/.zshrc to be generated")
	}
	
	// Verify .zshenv was NOT created (no env vars)
	zshenvPath := filepath.Join(tempHome, ".zshenv")
	if fileExists(zshenvPath) {
		t.Error("Expected ~/.zshenv NOT to be created when no env vars are present")
	}
}