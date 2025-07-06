package commands

import (
	"os"
	"path/filepath"
	"testing"
	
	"plonk/internal/utils"
)

func TestBackupCommand_SpecificFiles(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create existing files to backup
	existingZshrc := filepath.Join(tempHome, ".zshrc")
	existingContent := "# My existing zshrc\nalias ls='ls -la'"
	err := os.WriteFile(existingZshrc, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .zshrc: %v", err)
	}
	
	existingVimrc := filepath.Join(tempHome, ".vimrc")
	vimrcContent := "\" My existing vimrc\nset number"
	err = os.WriteFile(existingVimrc, []byte(vimrcContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .vimrc: %v", err)
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
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test backing up specific files
	err = backupCmdRun(nil, []string{existingZshrc, existingVimrc})
	if err != nil {
		t.Fatalf("Backup command failed: %v", err)
	}
	
	// Verify backups were created
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	if !utils.FileExists(backupDir) {
		t.Fatal("Expected backup directory to be created")
	}
	
	// Check for .zshrc backup
	zshrcBackups, err := filepath.Glob(filepath.Join(backupDir, "zshrc.backup.*"))
	if err != nil {
		t.Fatalf("Failed to search for zshrc backup files: %v", err)
	}
	
	if len(zshrcBackups) != 1 {
		t.Errorf("Expected 1 .zshrc backup file, found %d: %v", len(zshrcBackups), zshrcBackups)
	}
	
	// Check for .vimrc backup
	vimrcBackups, err := filepath.Glob(filepath.Join(backupDir, "vimrc.backup.*"))
	if err != nil {
		t.Fatalf("Failed to search for vimrc backup files: %v", err)
	}
	
	if len(vimrcBackups) != 1 {
		t.Errorf("Expected 1 .vimrc backup file, found %d: %v", len(vimrcBackups), vimrcBackups)
	}
	
	// Verify backup contents
	if len(zshrcBackups) > 0 {
		backupContent, err := os.ReadFile(zshrcBackups[0])
		if err != nil {
			t.Fatalf("Failed to read zshrc backup file: %v", err)
		}
		
		if string(backupContent) != existingContent {
			t.Errorf("Zshrc backup doesn't contain original content.\nExpected: %s\nGot: %s", 
				existingContent, string(backupContent))
		}
	}
}

func TestBackupCommand_AllFiles(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create existing .zshrc that would be overwritten by apply
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
	
	// Create config file with ZSH configuration (so apply would overwrite .zshrc)
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
	
	// Test backing up all files (no arguments)
	err = backupCmdRun(nil, []string{})
	if err != nil {
		t.Fatalf("Backup command failed: %v", err)
	}
	
	// Verify backup was created for .zshrc
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
}

func TestBackupCommand_NoExistingFiles(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config but no existing files
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
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
	
	// Test backing up all files when no files exist - should not error
	err = backupCmdRun(nil, []string{})
	if err != nil {
		t.Fatalf("Backup command should not fail when no files exist: %v", err)
	}
	
	// Backup directory should be created but empty of actual backups
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	if !utils.FileExists(backupDir) {
		t.Fatal("Expected backup directory to be created")
	}
	
	// No backup files should exist since no source files existed
	backupFiles, err := filepath.Glob(filepath.Join(backupDir, "*.backup.*"))
	if err != nil {
		t.Fatalf("Failed to search for backup files: %v", err)
	}
	
	if len(backupFiles) != 0 {
		t.Errorf("Expected 0 backup files when no source files exist, found %d: %v", len(backupFiles), backupFiles)
	}
}