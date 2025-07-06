package commands

import (
	"testing"
)

func TestPullCommand_Success(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()

	pullCalled := false
	mockGit := &MockGit{
		IsRepoFunc: func(dir string) bool {
			return true // Existing repo
		},
		PullFunc: func(repoDir string) error {
			pullCalled = true
			return nil
		},
	}
	gitClient = mockGit

	// Test
	err := runPull([]string{})
	if err != nil {
		t.Fatalf("Pull command failed: %v", err)
	}

	// Verify pull was called
	if !pullCalled {
		t.Error("Expected pull to be called")
	}
}

func TestPullCommand_NoRepository(t *testing.T) {
	// Setup
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Mock git client
	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()

	mockGit := &MockGit{
		IsRepoFunc: func(dir string) bool {
			return false // No existing repo
		},
	}
	gitClient = mockGit

	// Test - should error when no repo exists
	err := runPull([]string{})
	if err == nil {
		t.Error("Expected error when no repository exists")
	}
}

func TestPullCommand_WithArguments(t *testing.T) {
	// Test - should error when arguments are provided (pull takes no args)
	err := runPull([]string{"some-arg"})
	if err == nil {
		t.Error("Expected error when arguments are provided to pull")
	}
}
