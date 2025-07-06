package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	
	"plonk/internal/directories"
)

func TestCloneCommand_Success(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	plonkDir := directories.Default.PlonkDir()
	if _, err := os.Stat(plonkDir); os.IsNotExist(err) {
		t.Error("Expected plonk directory to be created")
	}
	
	// Verify plonk.yaml exists in cloned repo (now in repo subdirectory)
	repoDir := directories.Default.RepoDir()
	configPath := filepath.Join(repoDir, "plonk.yaml")
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
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Create existing directory
	plonkDir := directories.Default.PlonkDir()
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
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
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
	customDir := filepath.Join(tempHome, "my-dotfiles")
	_, cleanup := setupTestEnvWithPlonkDir(t, customDir)
	defer cleanup()
	
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
	
	// Verify it cloned to the custom location (with repo subdirectory)
	expectedRepoDir := filepath.Join(customDir, "repo")
	if clonedTo != expectedRepoDir {
		t.Errorf("Expected clone to %s, got %s", expectedRepoDir, clonedTo)
	}
}

func TestCloneCommand_WithBranchFlag(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()
	
	cloneCalled := false
	var capturedBranch string
	mockGit := &MockGitWithBranch{
		CloneBranchFunc: func(repoURL, targetDir, branch string) error {
			cloneCalled = true
			capturedBranch = branch
			// Simulate successful clone
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return err
			}
			configPath := filepath.Join(targetDir, "plonk.yaml")
			return os.WriteFile(configPath, []byte("test config"), 0644)
		},
	}
	gitClient = mockGit
	
	// Test with branch flag
	err := runCloneWithBranch([]string{"git@github.com/user/dotfiles.git"}, "develop")
	if err != nil {
		t.Fatalf("Clone command with branch failed: %v", err)
	}
	
	// Verify
	if !cloneCalled {
		t.Error("Expected clone to be called")
	}
	
	if capturedBranch != "develop" {
		t.Errorf("Expected branch 'develop', got '%s'", capturedBranch)
	}
}

func TestCloneCommand_WithBranchInURL(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()
	
	cloneCalled := false
	var capturedRepoURL, capturedBranch string
	mockGit := &MockGitWithBranch{
		CloneBranchFunc: func(repoURL, targetDir, branch string) error {
			cloneCalled = true
			capturedRepoURL = repoURL
			capturedBranch = branch
			// Simulate successful clone
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return err
			}
			configPath := filepath.Join(targetDir, "plonk.yaml")
			return os.WriteFile(configPath, []byte("test config"), 0644)
		},
	}
	gitClient = mockGit
	
	// Test with branch in URL
	err := runClone([]string{"git@github.com/user/dotfiles.git#feature-branch"})
	if err != nil {
		t.Fatalf("Clone command with branch in URL failed: %v", err)
	}
	
	// Verify
	if !cloneCalled {
		t.Error("Expected clone to be called")
	}
	
	if capturedRepoURL != "git@github.com/user/dotfiles.git" {
		t.Errorf("Expected clean repo URL 'git@github.com/user/dotfiles.git', got '%s'", capturedRepoURL)
	}
	
	if capturedBranch != "feature-branch" {
		t.Errorf("Expected branch 'feature-branch', got '%s'", capturedBranch)
	}
}

func TestCloneCommand_BranchFlagOverridesURL(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()
	
	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()
	
	var capturedBranch string
	mockGit := &MockGitWithBranch{
		CloneBranchFunc: func(repoURL, targetDir, branch string) error {
			capturedBranch = branch
			// Simulate successful clone
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return err
			}
			configPath := filepath.Join(targetDir, "plonk.yaml")
			return os.WriteFile(configPath, []byte("test config"), 0644)
		},
	}
	gitClient = mockGit
	
	// Test with both branch in URL and flag (flag should win)
	err := runCloneWithBranch([]string{"git@github.com/user/dotfiles.git#url-branch"}, "flag-branch")
	if err != nil {
		t.Fatalf("Clone command failed: %v", err)
	}
	
	// Verify flag takes precedence
	if capturedBranch != "flag-branch" {
		t.Errorf("Expected flag branch 'flag-branch', got '%s'", capturedBranch)
	}
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		input      string
		expectURL  string
		expectBranch string
	}{
		{"git@github.com/user/repo.git", "git@github.com/user/repo.git", ""},
		{"git@github.com/user/repo.git#main", "git@github.com/user/repo.git", "main"},
		{"https://github.com/user/repo.git#develop", "https://github.com/user/repo.git", "develop"},
		{"git@github.com/user/repo.git#feature/new-ui", "git@github.com/user/repo.git", "feature/new-ui"},
		{"simple-repo", "simple-repo", ""},
	}
	
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			url, branch := parseRepoURL(test.input)
			if url != test.expectURL {
				t.Errorf("Expected URL '%s', got '%s'", test.expectURL, url)
			}
			if branch != test.expectBranch {
				t.Errorf("Expected branch '%s', got '%s'", test.expectBranch, branch)
			}
		})
	}
}