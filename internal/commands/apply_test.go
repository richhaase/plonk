// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"plonk/internal/utils"

	"github.com/spf13/cobra"
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

	// Create source dotfile in plonk directory
	newZshrcContent := "# New zshrc from plonk\nalias ll='eza -la'"
	sourceZshrc := filepath.Join(plonkDir, "zshrc")
	err = os.WriteFile(sourceZshrc, []byte(newZshrcContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source zshrc: %v", err)
	}

	// Create config file with backup configuration and dotfiles
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5

dotfiles:
  - zshrc
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

	// Verify new .zshrc was copied from dotfiles
	actualContent, err := os.ReadFile(existingZshrc)
	if err != nil {
		t.Fatalf("Failed to read new .zshrc: %v", err)
	}

	if string(actualContent) != newZshrcContent {
		t.Errorf("Expected .zshrc to contain dotfile content.\nExpected: %s\nGot: %s",
			newZshrcContent, string(actualContent))
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

	// Create source dotfile
	newZshrcContent := "# New zshrc from plonk\nalias ll='eza -la'"
	sourceZshrc := filepath.Join(plonkDir, "zshrc")
	err = os.WriteFile(sourceZshrc, []byte(newZshrcContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source zshrc: %v", err)
	}

	// Create config file
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
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

	// Verify new .zshrc was copied from dotfiles
	actualContent, err := os.ReadFile(existingZshrc)
	if err != nil {
		t.Fatalf("Failed to read new .zshrc: %v", err)
	}

	if string(actualContent) != newZshrcContent {
		t.Errorf("Expected .zshrc to contain dotfile content.\nExpected: %s\nGot: %s",
			newZshrcContent, string(actualContent))
	}
}

func TestApplyCommand_DryRunFlagParsing(t *testing.T) {
	tests := []struct {
		name         string
		cmdArgs      []string
		dryRunFlag   bool
		backupFlag   bool
		expectDryRun bool
		expectBackup bool
	}{
		{
			name:         "no flags",
			cmdArgs:      []string{},
			dryRunFlag:   false,
			backupFlag:   false,
			expectDryRun: false,
			expectBackup: false,
		},
		{
			name:         "dry-run flag only",
			cmdArgs:      []string{},
			dryRunFlag:   true,
			backupFlag:   false,
			expectDryRun: true,
			expectBackup: false,
		},
		{
			name:         "backup flag only",
			cmdArgs:      []string{},
			dryRunFlag:   false,
			backupFlag:   true,
			expectDryRun: false,
			expectBackup: true,
		},
		{
			name:         "both flags",
			cmdArgs:      []string{},
			dryRunFlag:   true,
			backupFlag:   true,
			expectDryRun: true,
			expectBackup: true,
		},
		{
			name:         "dry-run with package arg",
			cmdArgs:      []string{"neovim"},
			dryRunFlag:   true,
			backupFlag:   false,
			expectDryRun: true,
			expectBackup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance to test flag parsing
			cmd := &cobra.Command{}
			cmd.Flags().Bool("backup", false, "test backup flag")
			cmd.Flags().Bool("dry-run", false, "test dry-run flag")

			// Set the flags
			if tt.backupFlag {
				cmd.Flags().Set("backup", "true")
			}
			if tt.dryRunFlag {
				cmd.Flags().Set("dry-run", "true")
			}

			// Test flag parsing
			backup, err := cmd.Flags().GetBool("backup")
			if err != nil {
				t.Fatalf("Failed to get backup flag: %v", err)
			}

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				t.Fatalf("Failed to get dry-run flag: %v", err)
			}

			// Verify the flags are parsed correctly
			if backup != tt.expectBackup {
				t.Errorf("Expected backup=%v, got %v", tt.expectBackup, backup)
			}

			if dryRun != tt.expectDryRun {
				t.Errorf("Expected dry-run=%v, got %v", tt.expectDryRun, dryRun)
			}
		})
	}
}

func TestApplyCommand_DryRunDoesNotCreateFiles(t *testing.T) {
	// This test verifies that dry-run mode doesn't actually create files
	// Setup test environment
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create plonk directory and config with dotfiles
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

	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
`
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test dry-run mode - should not create any files
	err = runApplyWithAllFlags([]string{}, false, true) // backup=false, dryRun=true
	if err != nil {
		t.Fatalf("Dry-run apply should not fail: %v", err)
	}

	// Verify NO files were actually created in dry-run mode
	if utils.FileExists(filepath.Join(tempHome, ".zshrc")) {
		t.Error("Expected .zshrc NOT to be created in dry-run mode")
	}
}
