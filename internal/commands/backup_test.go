// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plonk/internal/utils"
)

func TestBackupExistingFile_CreatesBackup(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create existing file to backup
	existingFile := filepath.Join(tempHome, ".zshrc")
	originalContent := "# Original zshrc content\nexport PATH=/usr/local/bin:$PATH"
	err := os.WriteFile(existingFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Test backup functionality
	backupPath, err := BackupExistingFile(existingFile)
	if err != nil {
		t.Fatalf("BackupExistingFile failed: %v", err)
	}

	// Verify backup was created
	if !utils.FileExists(backupPath) {
		t.Fatalf("Backup file was not created at %s", backupPath)
	}

	// Verify backup contains original content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != originalContent {
		t.Errorf("Backup content doesn't match original.\nOriginal: %s\nBackup: %s",
			originalContent, string(backupContent))
	}

	// Verify backup path format (should include timestamp)
	if !strings.Contains(backupPath, ".zshrc.backup.") {
		t.Errorf("Expected backup path to contain '.zshrc.backup.', got %s", backupPath)
	}
}

func TestBackupExistingFile_NonExistentFile(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test backup of non-existent file
	nonExistentFile := filepath.Join(tempHome, ".nonexistent")
	backupPath, err := BackupExistingFile(nonExistentFile)

	// Should not error, but should return empty string
	if err != nil {
		t.Errorf("BackupExistingFile should not error for non-existent file: %v", err)
	}

	if backupPath != "" {
		t.Errorf("Expected empty backup path for non-existent file, got %s", backupPath)
	}
}

func TestBackupConfigurationFiles_CreatesBackupsInConfigurableLocation(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create existing .zshrc
	existingZshrc := filepath.Join(tempHome, ".zshrc")
	originalZshrcContent := "# My existing zshrc\nalias ls='ls -la'"
	err := os.WriteFile(existingZshrc, []byte(originalZshrcContent), 0644)
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

	// Test backup functionality
	err = BackupConfigurationFiles([]string{existingZshrc})
	if err != nil {
		t.Fatalf("BackupConfigurationFiles failed: %v", err)
	}

	// Verify backup was created in configured location
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	if !utils.FileExists(backupDir) {
		t.Fatal("Expected backup directory to be created")
	}

	backupFiles, err := filepath.Glob(filepath.Join(backupDir, "zshrc.backup.*"))
	if err != nil {
		t.Fatalf("Failed to search for backup files: %v", err)
	}

	if len(backupFiles) != 1 {
		t.Errorf("Expected 1 .zshrc backup file, found %d: %v", len(backupFiles), backupFiles)
	}

	// Verify backup contains original content
	if len(backupFiles) > 0 {
		backupContent, err := os.ReadFile(backupFiles[0])
		if err != nil {
			t.Fatalf("Failed to read backup file: %v", err)
		}

		if string(backupContent) != originalZshrcContent {
			t.Errorf("Backup doesn't contain original content.\nExpected: %s\nGot: %s",
				originalZshrcContent, string(backupContent))
		}
	}
}
