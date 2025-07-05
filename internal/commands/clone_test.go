package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCloneCommand_Success(t *testing.T) {
	// Setup
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()
	
	mockGit := &MockGit{
		CloneFunc: func(repoURL, targetDir string) error {
			// Simulate successful clone by creating the directory and plonk.yaml
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return err
			}
			configPath := filepath.Join(targetDir, "plonk.yaml")
			return os.WriteFile(configPath, []byte("test config"), 0644)
		},
	}
	gitClient = mockGit
	
	// Test
	err := runClone([]string{"git@github.com/user/dotfiles.git"})
	if err != nil {
		t.Fatalf("Clone command failed: %v", err)
	}
	
	// Verify the repo was cloned to the plonk directory
	plonkDir := getPlonkDir()
	if _, err := os.Stat(plonkDir); os.IsNotExist(err) {
		t.Error("Expected plonk directory to be created")
	}
	
	// Verify plonk.yaml exists in cloned repo
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected plonk.yaml to exist in cloned repo")
	}
}

func TestCloneCommand_NoRepository(t *testing.T) {
	// Test - should error when no repository is provided
	err := runClone([]string{})
	if err == nil {
		t.Error("Expected error when no repository is provided")
	}
}

func TestCloneCommand_ExistingDirectory(t *testing.T) {
	// Setup
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create existing directory
	plonkDir := getPlonkDir()
	if err := os.MkdirAll(plonkDir, 0755); err != nil {
		t.Fatalf("Failed to create existing directory: %v", err)
	}
	
	// Test - should error when directory already exists
	err := runClone([]string{"git@github.com/user/dotfiles.git"})
	if err == nil {
		t.Error("Expected error when target directory already exists")
	}
}

func TestCloneCommand_CloneError(t *testing.T) {
	// Setup
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Mock git client to fail
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()
	
	mockGit := &MockGit{
		CloneFunc: func(repoURL, targetDir string) error {
			return fmt.Errorf("clone failed")
		},
	}
	gitClient = mockGit
	
	// Test
	err := runClone([]string{"invalid-repo-url"})
	if err == nil {
		t.Error("Expected error for failed clone")
	}
}

func TestCloneCommand_CustomLocation(t *testing.T) {
	// Setup
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Set custom location
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)
	
	customDir := filepath.Join(tempHome, "my-dotfiles")
	os.Setenv("PLONK_DIR", customDir)
	
	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()
	
	var clonedTo string
	mockGit := &MockGit{
		CloneFunc: func(repoURL, targetDir string) error {
			clonedTo = targetDir
			// Simulate successful clone
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return err
			}
			configPath := filepath.Join(targetDir, "plonk.yaml")
			return os.WriteFile(configPath, []byte("test config"), 0644)
		},
	}
	gitClient = mockGit
	
	// Test
	err := runClone([]string{"git@github.com/user/dotfiles.git"})
	if err != nil {
		t.Fatalf("Clone command failed: %v", err)
	}
	
	// Verify it cloned to the custom location
	if clonedTo != customDir {
		t.Errorf("Expected clone to %s, got %s", customDir, clonedTo)
	}
}