// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepoCommand_Success(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()

	cloneCalled := false
	pullCalled := false
	mockGit := &MockGit{
		IsRepoFunc: func(dir string) bool {
			return false // No existing repo, should clone
		},
		CloneFunc: func(repoURL, targetDir string) error {
			cloneCalled = true

			// Create a mock config file to simulate successful clone
			err := os.MkdirAll(targetDir, 0755)
			if err != nil {
				return err
			}

			configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
`
			configPath := filepath.Join(targetDir, "plonk.yaml")
			err = os.WriteFile(configPath, []byte(configContent), 0644)
			if err != nil {
				return err
			}

			// Create source file for dotfile
			return os.WriteFile(filepath.Join(targetDir, "zshrc"), []byte("# test zshrc"), 0644)
		},
		PullFunc: func(repoDir string) error {
			pullCalled = true
			return nil
		},
	}
	gitClient = mockGit

	// Test
	err := runRepo([]string{"git@github.com/user/dotfiles.git"})
	if err != nil {
		t.Fatalf("Setup command failed: %v", err)
	}

	// Verify clone was called (not pull, since no existing repo)
	if !cloneCalled {
		t.Error("Expected clone to be called")
	}

	if pullCalled {
		t.Error("Expected pull NOT to be called when no existing repo")
	}
}

func TestRepoCommand_ExistingRepo(t *testing.T) {
	// Setup
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create existing plonk directory with config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
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

	// Create source file for dotfile
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()

	cloneCalled := false
	pullCalled := false
	mockGit := &MockGit{
		IsRepoFunc: func(dir string) bool {
			return true // Existing repo, should pull
		},
		CloneFunc: func(repoURL, targetDir string) error {
			cloneCalled = true
			return nil
		},
		PullFunc: func(repoDir string) error {
			pullCalled = true
			return nil
		},
	}
	gitClient = mockGit

	// Test
	err = runRepo([]string{"git@github.com/user/dotfiles.git"})
	if err != nil {
		t.Fatalf("Setup command failed: %v", err)
	}

	// Verify pull was called (not clone, since repo exists)
	if cloneCalled {
		t.Error("Expected clone NOT to be called when repo exists")
	}

	if !pullCalled {
		t.Error("Expected pull to be called when repo exists")
	}
}

func TestRepoCommand_NoRepository(t *testing.T) {
	// Test - should error when no repository URL provided
	err := runRepo([]string{})
	if err == nil {
		t.Error("Expected error when no repository URL provided")
	}
}

func TestRepoCommand_TooManyArguments(t *testing.T) {
	// Test - should error when too many arguments provided
	err := runRepo([]string{"repo1", "repo2"})
	if err == nil {
		t.Error("Expected error when too many arguments provided")
	}
}
