package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"plonk/internal/utils"
)

func TestApplyCommand_NoConfig(t *testing.T) {
	// Setup temporary directory
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Test - should error when no config exists
	err := runApply([]string{})
	if err == nil {
		t.Error("Expected error when no config file exists")
	}
}

func TestApplyCommand_AllDotfiles(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	if !utils.FileExists(filepath.Join(tempHome, ".zshrc")) {
		t.Error("Expected ~/.zshrc to be created")
	}
	
	if !utils.FileExists(filepath.Join(tempHome, ".gitconfig")) {
		t.Error("Expected ~/.gitconfig to be created")
	}
	
	if !utils.FileExists(filepath.Join(tempHome, ".config", "nvim", "init.vim")) {
		t.Error("Expected ~/.config/nvim/init.vim to be created")
	}
}

func TestApplyCommand_PackageSpecific(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	if !utils.FileExists(filepath.Join(tempHome, ".config", "nvim", "init.vim")) {
		t.Error("Expected ~/.config/nvim/init.vim to be created")
	}
	
	if utils.FileExists(filepath.Join(tempHome, ".zshrc")) {
		t.Error("Expected ~/.zshrc NOT to be created when applying package-specific config")
	}
	
	if utils.FileExists(filepath.Join(tempHome, ".config", "mcfly", "config.yaml")) {
		t.Error("Expected ~/.config/mcfly/config.yaml NOT to be created when applying only neovim")
	}
}

func TestApplyCommand_InvalidPackage(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	if !utils.FileExists(zshrcPath) {
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
	if !utils.FileExists(zshenvPath) {
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
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	if !utils.FileExists(zshrcPath) {
		t.Fatal("Expected ~/.zshrc to be generated")
	}
	
	// Verify .zshenv was NOT created (no env vars)
	zshenvPath := filepath.Join(tempHome, ".zshenv")
	if utils.FileExists(zshenvPath) {
		t.Error("Expected ~/.zshenv NOT to be created when no env vars are present")
	}
}

func TestApplyCommand_WithBackupFlag(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Create existing .zshrc that should be backed up
	existingZshrc := filepath.Join(tempHome, ".zshrc")
	existingContent := "# My existing zshrc\nalias ls='ls -la'"
	err := os.WriteFile(existingZshrc, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .zshrc: %v", err)
	}
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err = os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file with backup configuration
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5

zsh:
  aliases:
    ll: "eza -la"
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying with backup flag - this should backup .zshrc before applying
	err = runApplyWithFlags([]string{}, true) // backup=true
	if err != nil {
		t.Fatalf("Apply command with backup failed: %v", err)
	}
	
	// Verify backup was created
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	if !utils.FileExists(backupDir) {
		t.Fatal("Expected backup directory to be created")
	}
	
	backupFiles, err := filepath.Glob(filepath.Join(backupDir, "zshrc.backup.*"))
	if err != nil {
		t.Fatalf("Failed to search for backup files: %v", err)
	}
	
	if len(backupFiles) != 1 {
		t.Errorf("Expected 1 backup file, found %d: %v", len(backupFiles), backupFiles)
	}
	
	// Verify backup contains original content
	if len(backupFiles) > 0 {
		backupContent, err := os.ReadFile(backupFiles[0])
		if err != nil {
			t.Fatalf("Failed to read backup file: %v", err)
		}
		
		if string(backupContent) != existingContent {
			t.Errorf("Backup doesn't contain original content.\nExpected: %s\nGot: %s", 
				existingContent, string(backupContent))
		}
	}
	
	// Verify new .zshrc was created with plonk content
	newZshrcContent, err := os.ReadFile(existingZshrc)
	if err != nil {
		t.Fatalf("Failed to read new .zshrc: %v", err)
	}
	
	newZshrcStr := string(newZshrcContent)
	if !strings.Contains(newZshrcStr, "alias ll='eza -la'") {
		t.Error("Expected new .zshrc to contain plonk-generated content")
	}
	
	if strings.Contains(newZshrcStr, "alias ls='ls -la'") {
		t.Error("Expected new .zshrc to not contain old content")
	}
}

func TestApplyCommand_WithoutBackupFlag(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Create existing .zshrc that should NOT be backed up
	existingZshrc := filepath.Join(tempHome, ".zshrc")
	existingContent := "# My existing zshrc\nalias ls='ls -la'"
	err := os.WriteFile(existingZshrc, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .zshrc: %v", err)
	}
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err = os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

zsh:
  aliases:
    ll: "eza -la"
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying without backup flag - should not backup
	err = runApplyWithFlags([]string{}, false) // backup=false
	if err != nil {
		t.Fatalf("Apply command failed: %v", err)
	}
	
	// Verify no backup directory was created
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	if utils.FileExists(backupDir) {
		t.Error("Expected backup directory NOT to be created when --backup flag is not used")
	}
	
	// Verify new .zshrc was still created with plonk content
	newZshrcContent, err := os.ReadFile(existingZshrc)
	if err != nil {
		t.Fatalf("Failed to read new .zshrc: %v", err)
	}
	
	newZshrcStr := string(newZshrcContent)
	if !strings.Contains(newZshrcStr, "alias ll='eza -la'") {
		t.Error("Expected new .zshrc to contain plonk-generated content")
	}
}

func TestApplyCommand_GitConfiguration(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file with Git configuration
	configContent := `settings:
  default_manager: homebrew

git:
  user:
    name: "Test User"
    email: "test@example.com"
  core:
    pager: "delta"
    excludesfile: "~/.gitignore_global"
  aliases:
    st: "status"
    co: "checkout"
    br: "branch"
  color:
    ui: "auto"
    branch: "auto"
  push:
    default: "current"
    autoSetupRemote: "true"
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying Git configuration
	err = runApply([]string{})
	if err != nil {
		t.Fatalf("Apply command with Git config failed: %v", err)
	}
	
	// Verify .gitconfig was created
	gitconfigPath := filepath.Join(tempHome, ".gitconfig")
	if !utils.FileExists(gitconfigPath) {
		t.Fatal("Expected .gitconfig to be created")
	}
	
	// Read and verify .gitconfig content
	gitconfigContent, err := os.ReadFile(gitconfigPath)
	if err != nil {
		t.Fatalf("Failed to read .gitconfig: %v", err)
	}
	
	gitconfigStr := string(gitconfigContent)
	
	// Verify header comment
	if !strings.Contains(gitconfigStr, "# Generated by plonk - Do not edit manually") {
		t.Error("Expected .gitconfig to contain plonk header comment")
	}
	
	// Verify user section
	if !strings.Contains(gitconfigStr, "[user]") {
		t.Error("Expected .gitconfig to contain [user] section")
	}
	if !strings.Contains(gitconfigStr, `name = "Test User"`) {
		t.Error("Expected .gitconfig to contain user name")
	}
	if !strings.Contains(gitconfigStr, `email = "test@example.com"`) {
		t.Error("Expected .gitconfig to contain user email")
	}
	
	// Verify core section
	if !strings.Contains(gitconfigStr, "[core]") {
		t.Error("Expected .gitconfig to contain [core] section")
	}
	if !strings.Contains(gitconfigStr, `pager = "delta"`) {
		t.Error("Expected .gitconfig to contain core pager setting")
	}
	if !strings.Contains(gitconfigStr, `excludesfile = "~/.gitignore_global"`) {
		t.Error("Expected .gitconfig to contain core excludesfile setting")
	}
	
	// Verify alias section
	if !strings.Contains(gitconfigStr, "[alias]") {
		t.Error("Expected .gitconfig to contain [alias] section")
	}
	if !strings.Contains(gitconfigStr, `st = "status"`) {
		t.Error("Expected .gitconfig to contain st alias")
	}
	if !strings.Contains(gitconfigStr, `co = "checkout"`) {
		t.Error("Expected .gitconfig to contain co alias")
	}
	if !strings.Contains(gitconfigStr, `br = "branch"`) {
		t.Error("Expected .gitconfig to contain br alias")
	}
	
	// Verify color section
	if !strings.Contains(gitconfigStr, "[color]") {
		t.Error("Expected .gitconfig to contain [color] section")
	}
	if !strings.Contains(gitconfigStr, `ui = "auto"`) {
		t.Error("Expected .gitconfig to contain color ui setting")
	}
	if !strings.Contains(gitconfigStr, `branch = "auto"`) {
		t.Error("Expected .gitconfig to contain color branch setting")
	}
	
	// Verify push section
	if !strings.Contains(gitconfigStr, "[push]") {
		t.Error("Expected .gitconfig to contain [push] section")
	}
	if !strings.Contains(gitconfigStr, `default = "current"`) {
		t.Error("Expected .gitconfig to contain push default setting")
	}
	if !strings.Contains(gitconfigStr, `autoSetupRemote = "true"`) {
		t.Error("Expected .gitconfig to contain push autoSetupRemote setting")
	}
}

func TestApplyCommand_GitConfigurationWithBackup(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Create existing .gitconfig that should be backed up
	existingGitconfig := filepath.Join(tempHome, ".gitconfig")
	existingContent := `[user]
	name = "Old User"
	email = "old@example.com"`
	err := os.WriteFile(existingGitconfig, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .gitconfig: %v", err)
	}
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err = os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file with backup and Git configuration
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5

git:
  user:
    name: "New User"
    email: "new@example.com"
  core:
    pager: "delta"
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test applying with backup flag
	err = runApplyWithFlags([]string{}, true) // backup=true
	if err != nil {
		t.Fatalf("Apply command with backup failed: %v", err)
	}
	
	// Verify backup was created
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	if !utils.FileExists(backupDir) {
		t.Fatal("Expected backup directory to be created")
	}
	
	// Verify .gitconfig backup exists
	backupFiles, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}
	
	var gitconfigBackupExists bool
	for _, file := range backupFiles {
		if strings.Contains(file.Name(), "gitconfig.backup.") {
			gitconfigBackupExists = true
			break
		}
	}
	
	if !gitconfigBackupExists {
		t.Error("Expected gitconfig backup file to be created")
	}
	
	// Verify new .gitconfig was created with plonk content
	newGitconfigContent, err := os.ReadFile(existingGitconfig)
	if err != nil {
		t.Fatalf("Failed to read new .gitconfig: %v", err)
	}
	
	newGitconfigStr := string(newGitconfigContent)
	if !strings.Contains(newGitconfigStr, "# Generated by plonk") {
		t.Error("Expected new .gitconfig to contain plonk header")
	}
	if !strings.Contains(newGitconfigStr, `name = "New User"`) {
		t.Error("Expected new .gitconfig to contain new user name")
	}
	if !strings.Contains(newGitconfigStr, `email = "new@example.com"`) {
		t.Error("Expected new .gitconfig to contain new user email")
	}
}