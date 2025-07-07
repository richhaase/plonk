// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package directories

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Error("NewManager should return a non-nil manager")
	}
}

func TestPlonkDir_Default(t *testing.T) {
	// Clear any PLONK_DIR environment variable.
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Unsetenv("PLONK_DIR")

	manager := NewManager()
	plonkDir := manager.PlonkDir()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	expected := filepath.Join(homeDir, ".config", "plonk")
	if plonkDir != expected {
		t.Errorf("Expected PlonkDir to be %s, got %s", expected, plonkDir)
	}
}

func TestPlonkDir_CustomEnvironment(t *testing.T) {
	// Set custom PLONK_DIR.
	tempDir := t.TempDir()
	customPlonkDir := filepath.Join(tempDir, "custom-plonk")

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", customPlonkDir)

	manager := NewManager()
	plonkDir := manager.PlonkDir()

	if plonkDir != customPlonkDir {
		t.Errorf("Expected PlonkDir to be %s, got %s", customPlonkDir, plonkDir)
	}
}

func TestPlonkDir_Caching(t *testing.T) {
	manager := NewManager()

	// First call.
	plonkDir1 := manager.PlonkDir()

	// Second call should return the same cached value.
	plonkDir2 := manager.PlonkDir()

	if plonkDir1 != plonkDir2 {
		t.Errorf("PlonkDir should return consistent cached values: %s != %s", plonkDir1, plonkDir2)
	}
}

func TestRepoDir_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", tempDir)

	manager := NewManager()
	repoDir := manager.RepoDir()

	expectedRepoDir := filepath.Join(tempDir, "repo")
	if repoDir != expectedRepoDir {
		t.Errorf("Expected RepoDir to be %s, got %s", expectedRepoDir, repoDir)
	}

	// Verify directory was created.
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Errorf("RepoDir should create the directory, but %s does not exist", repoDir)
	}
}

func TestBackupsDir_Default(t *testing.T) {
	tempDir := t.TempDir()

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", tempDir)

	manager := NewManager()
	backupsDir := manager.BackupsDir()

	expectedBackupsDir := filepath.Join(tempDir, "backups")
	if backupsDir != expectedBackupsDir {
		t.Errorf("Expected BackupsDir to be %s, got %s", expectedBackupsDir, backupsDir)
	}

	// Verify directory was created.
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		t.Errorf("BackupsDir should create the directory, but %s does not exist", backupsDir)
	}
}

func TestBackupsDir_CustomConfigLocation(t *testing.T) {
	tempDir := t.TempDir()
	customBackupDir := filepath.Join(tempDir, "custom-backups")

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", tempDir)

	// Create a config file with custom backup location.
	configContent := `backup:
  location: ` + customBackupDir + `
settings:
  default_manager: homebrew`

	configPath := filepath.Join(tempDir, "plonk.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	manager := NewManager()
	backupsDir := manager.BackupsDir()

	if backupsDir != customBackupDir {
		t.Errorf("Expected BackupsDir to be %s, got %s", customBackupDir, backupsDir)
	}

	// Verify directory was created.
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		t.Errorf("BackupsDir should create the custom directory, but %s does not exist", backupsDir)
	}
}

func TestExpandHomeDir_NoTilde(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		input    string
		expected string
	}{
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, test := range tests {
		result := manager.ExpandHomeDir(test.input)
		if result != test.expected {
			t.Errorf("ExpandHomeDir(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestExpandHomeDir_WithTilde(t *testing.T) {
	manager := NewManager()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~", homeDir},
		{"~/", homeDir},
		{"~/Documents", filepath.Join(homeDir, "Documents")},
		{"~/path/to/file", filepath.Join(homeDir, "path/to/file")},
	}

	for _, test := range tests {
		result := manager.ExpandHomeDir(test.input)
		if result != test.expected {
			t.Errorf("ExpandHomeDir(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestReset_ClearsCachedPaths(t *testing.T) {
	tempDir := t.TempDir()

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", tempDir)

	manager := NewManager()

	// Cache some paths.
	plonkDir1 := manager.PlonkDir()
	repoDir1 := manager.RepoDir()

	// Change environment.
	newTempDir := t.TempDir()
	os.Setenv("PLONK_DIR", newTempDir)

	// Without reset, should return cached values.
	plonkDir2 := manager.PlonkDir()
	if plonkDir2 != plonkDir1 {
		t.Errorf("Without reset, should return cached PlonkDir: %s != %s", plonkDir2, plonkDir1)
	}

	// After reset, should respect new environment.
	manager.Reset()
	plonkDir3 := manager.PlonkDir()
	repoDir3 := manager.RepoDir()

	if plonkDir3 == plonkDir1 {
		t.Errorf("After reset, PlonkDir should reflect new environment")
	}
	if repoDir3 == repoDir1 {
		t.Errorf("After reset, RepoDir should reflect new environment")
	}

	expectedPlonkDir := newTempDir
	expectedRepoDir := filepath.Join(newTempDir, "repo")

	if plonkDir3 != expectedPlonkDir {
		t.Errorf("Expected PlonkDir to be %s, got %s", expectedPlonkDir, plonkDir3)
	}
	if repoDir3 != expectedRepoDir {
		t.Errorf("Expected RepoDir to be %s, got %s", expectedRepoDir, repoDir3)
	}
}

func TestEnsureStructure_NewInstallation(t *testing.T) {
	tempDir := t.TempDir()

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", tempDir)

	manager := NewManager()
	err := manager.EnsureStructure()
	if err != nil {
		t.Fatalf("EnsureStructure failed: %v", err)
	}

	// Verify both directories exist.
	repoDir := filepath.Join(tempDir, "repo")
	backupsDir := filepath.Join(tempDir, "backups")

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Errorf("EnsureStructure should create repo directory")
	}
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		t.Errorf("EnsureStructure should create backups directory")
	}
}

func TestEnsureStructure_MigrationNeeded(t *testing.T) {
	tempDir := t.TempDir()

	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	os.Setenv("PLONK_DIR", tempDir)

	// Create some files in the flat structure (simulating old installation).
	err := os.WriteFile(filepath.Join(tempDir, "some-file.txt"), []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "backup.backup.20241206"), []byte("backup"), 0644)
	if err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "plonk.yaml"), []byte("settings:\n  default_manager: homebrew"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	manager := NewManager()
	err = manager.EnsureStructure()
	if err != nil {
		t.Fatalf("EnsureStructure failed: %v", err)
	}

	// Verify migration occurred.
	repoFile := filepath.Join(tempDir, "repo", "some-file.txt")
	backupFile := filepath.Join(tempDir, "backups", "backup.backup.20241206")
	configFile := filepath.Join(tempDir, "plonk.yaml")

	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		t.Errorf("Regular file should be migrated to repo directory")
	}
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Errorf("Backup file should be migrated to backups directory")
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file should remain in root directory")
	}

	// Original files should be gone.
	if _, err := os.Stat(filepath.Join(tempDir, "some-file.txt")); !os.IsNotExist(err) {
		t.Errorf("Original file should be moved from root")
	}
}
