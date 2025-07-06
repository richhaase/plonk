package commands

import (
	"os"
	"path/filepath"
	"testing"
	
	"plonk/internal/utils"
)

func TestDirectoryStructure_NewInstallation(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Clear PLONK_DIR to use default
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Unsetenv("PLONK_DIR")
	
	// Test that new directory structure is created correctly
	expectedPlonkDir := filepath.Join(tempHome, ".config", "plonk")
	expectedRepoDir := filepath.Join(expectedPlonkDir, "repo")
	expectedBackupsDir := filepath.Join(expectedPlonkDir, "backups")
	
	// Create config to trigger directory creation
	configContent := `settings:
  default_manager: homebrew

backup:
  location: default
  keep_count: 5
`
	
	// Create plonk directory
	err := os.MkdirAll(expectedPlonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Write config file
	configPath := filepath.Join(expectedPlonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test that getRepoDir creates and returns correct path
	repoDir := getRepoDir()
	if repoDir != expectedRepoDir {
		t.Errorf("Expected repo dir %s, got %s", expectedRepoDir, repoDir)
	}
	
	// Test that getBackupsDir creates and returns correct path
	backupsDir := getBackupsDir()
	if backupsDir != expectedBackupsDir {
		t.Errorf("Expected backups dir %s, got %s", expectedBackupsDir, backupsDir)
	}
	
	// Verify directories are created
	if !utils.FileExists(expectedRepoDir) {
		t.Error("Expected repo directory to be created")
	}
	
	if !utils.FileExists(expectedBackupsDir) {
		t.Error("Expected backups directory to be created")
	}
}

func TestDirectoryStructure_BackwardCompatibility(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Clear PLONK_DIR to use default
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Unsetenv("PLONK_DIR")
	
	// Create old-style directory structure (everything in ~/.config/plonk/)
	oldPlonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(oldPlonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create old plonk directory: %v", err)
	}
	
	// Create old-style files in root plonk directory
	oldConfigFile := filepath.Join(oldPlonkDir, "plonk.yaml")
	oldRepoFile := filepath.Join(oldPlonkDir, "README.md") // Simulates repo content
	oldBackupFile := filepath.Join(oldPlonkDir, "zshrc.backup.20241206-143022")
	
	configContent := `settings:
  default_manager: homebrew

backup:
  location: default
  keep_count: 5
`
	
	err = os.WriteFile(oldConfigFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write old config file: %v", err)
	}
	
	err = os.WriteFile(oldRepoFile, []byte("# Old repo content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write old repo file: %v", err)
	}
	
	err = os.WriteFile(oldBackupFile, []byte("# old backup"), 0644)
	if err != nil {
		t.Fatalf("Failed to write old backup file: %v", err)
	}
	
	// Test migration functionality
	err = migrateDirectoryStructure()
	if err != nil {
		t.Fatalf("Directory migration failed: %v", err)
	}
	
	// Verify new structure exists
	expectedRepoDir := filepath.Join(oldPlonkDir, "repo")
	expectedBackupsDir := filepath.Join(oldPlonkDir, "backups")
	
	if !utils.FileExists(expectedRepoDir) {
		t.Error("Expected repo directory to be created during migration")
	}
	
	if !utils.FileExists(expectedBackupsDir) {
		t.Error("Expected backups directory to be created during migration")
	}
	
	// Verify files were moved correctly
	newRepoFile := filepath.Join(expectedRepoDir, "README.md")
	newBackupFile := filepath.Join(expectedBackupsDir, "zshrc.backup.20241206-143022")
	
	if !utils.FileExists(newRepoFile) {
		t.Error("Expected repo file to be moved to repo/ directory")
	}
	
	if !utils.FileExists(newBackupFile) {
		t.Error("Expected backup file to be moved to backups/ directory")
	}
	
	// Verify config file stays in root
	if !utils.FileExists(oldConfigFile) {
		t.Error("Expected config file to remain in root plonk directory")
	}
	
	// Verify old files are gone from root
	if utils.FileExists(oldRepoFile) {
		t.Error("Expected old repo file to be removed from root")
	}
	
	if utils.FileExists(oldBackupFile) {
		t.Error("Expected old backup file to be removed from root")
	}
}

func TestDirectoryStructure_CustomBackupLocation(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Clear PLONK_DIR to use default
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Unsetenv("PLONK_DIR")
	
	// Create plonk directory
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config with custom backup location
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/my-custom-backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test that custom backup location is respected (not migrated to subdirectory)
	backupsDir := getBackupsDir()
	expectedExpandedPath := filepath.Join(tempHome, "my-custom-backups")
	if backupsDir != expectedExpandedPath {
		t.Errorf("Expected custom backup dir %s, got %s", expectedExpandedPath, backupsDir)
	}
	
	// Test that repo directory still uses new structure
	expectedRepoDir := filepath.Join(plonkDir, "repo")
	repoDir := getRepoDir()
	if repoDir != expectedRepoDir {
		t.Errorf("Expected repo dir %s, got %s", expectedRepoDir, repoDir)
	}
}

func TestDirectoryStructure_PlonkDirEnvironmentVariable(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Set custom PLONK_DIR
	customPlonkDir := filepath.Join(tempHome, "my-dotfiles")
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", customPlonkDir)
	
	// Test that subdirectories use custom plonk dir
	expectedRepoDir := filepath.Join(customPlonkDir, "repo")
	expectedBackupsDir := filepath.Join(customPlonkDir, "backups")
	
	repoDir := getRepoDir()
	if repoDir != expectedRepoDir {
		t.Errorf("Expected repo dir %s, got %s", expectedRepoDir, repoDir)
	}
	
	backupsDir := getBackupsDir()
	if backupsDir != expectedBackupsDir {
		t.Errorf("Expected backups dir %s, got %s", expectedBackupsDir, backupsDir)
	}
}

func TestDirectoryStructure_DetectExistingStructure(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Clear PLONK_DIR to use default
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Unsetenv("PLONK_DIR")
	
	// Create new-style directory structure
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	repoDir := filepath.Join(plonkDir, "repo")
	backupsDir := filepath.Join(plonkDir, "backups")
	
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo directory: %v", err)
	}
	
	err = os.MkdirAll(backupsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backups directory: %v", err)
	}
	
	// Test that existing structure is detected correctly
	isNewStructure := hasNewDirectoryStructure()
	if !isNewStructure {
		t.Error("Expected new directory structure to be detected")
	}
	
	// Test with old structure
	err = os.RemoveAll(repoDir)
	if err != nil {
		t.Fatalf("Failed to remove repo directory: %v", err)
	}
	
	err = os.RemoveAll(backupsDir)
	if err != nil {
		t.Fatalf("Failed to remove backups directory: %v", err)
	}
	
	isNewStructure = hasNewDirectoryStructure()
	if isNewStructure {
		t.Error("Expected old directory structure to be detected")
	}
}