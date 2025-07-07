// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"testing"

	"plonk/internal/directories"
)

// setupTestEnv sets up a temporary test environment with isolated HOME directory
// and returns a cleanup function that should be called with defer.
func setupTestEnv(t *testing.T) (tempHome string, cleanup func()) {
	tempHome = t.TempDir()
	originalHome := os.Getenv("HOME")

	cleanup = func() {
		os.Setenv("HOME", originalHome)
		directories.Default.Reset()
	}

	os.Setenv("HOME", tempHome)
	return tempHome, cleanup
}

// setupTestEnvWithPlonkDir sets up a temporary test environment with both HOME and PLONK_DIR
// and returns a cleanup function that should be called with defer.
func setupTestEnvWithPlonkDir(t *testing.T, plonkDir string) (tempHome string, cleanup func()) {
	tempHome = t.TempDir()
	originalHome := os.Getenv("HOME")
	originalPlonkDir := os.Getenv("PLONK_DIR")

	cleanup = func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("PLONK_DIR", originalPlonkDir)
		directories.Default.Reset()
	}

	os.Setenv("HOME", tempHome)
	os.Setenv("PLONK_DIR", plonkDir)
	return tempHome, cleanup
}

// MockGit implements GitInterface for testing.
type MockGit struct {
	CloneFunc  func(repoURL, targetDir string) error
	PullFunc   func(repoDir string) error
	IsRepoFunc func(dir string) bool
}

// MockGitWithBranch implements GitInterface with branch support for testing.
type MockGitWithBranch struct {
	CloneBranchFunc func(repoURL, targetDir, branch string) error
	PullFunc        func(repoDir string) error
	IsRepoFunc      func(dir string) bool
}

func (m *MockGitWithBranch) Clone(repoURL, targetDir string) error {
	return m.CloneBranch(repoURL, targetDir, "")
}

func (m *MockGitWithBranch) CloneBranch(repoURL, targetDir, branch string) error {
	if m.CloneBranchFunc != nil {
		return m.CloneBranchFunc(repoURL, targetDir, branch)
	}
	return nil
}

func (m *MockGitWithBranch) Pull(repoDir string) error {
	if m.PullFunc != nil {
		return m.PullFunc(repoDir)
	}
	return nil
}

func (m *MockGitWithBranch) IsRepo(dir string) bool {
	if m.IsRepoFunc != nil {
		return m.IsRepoFunc(dir)
	}
	return false
}

func (m *MockGit) Clone(repoURL, targetDir string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(repoURL, targetDir)
	}
	return nil
}

func (m *MockGit) CloneBranch(repoURL, targetDir, branch string) error {
	// For basic MockGit, just call Clone.
	return m.Clone(repoURL, targetDir)
}

func (m *MockGit) Pull(repoDir string) error {
	if m.PullFunc != nil {
		return m.PullFunc(repoDir)
	}
	return nil
}

func (m *MockGit) IsRepo(dir string) bool {
	if m.IsRepoFunc != nil {
		return m.IsRepoFunc(dir)
	}
	return false
}

// runApplyWithFlags runs the apply command with flag support for testing.
func runApplyWithFlags(args []string, backup bool) error {
	return runApplyWithBackup(args, backup)
}

// runApplyWithAllFlags runs the apply command with both backup and dry-run flag support for testing.
func runApplyWithAllFlags(args []string, backup bool, dryRun bool) error {
	// This will call a function that doesn't exist yet - that's fine for Red phase TDD.
	return runApplyWithAllOptions(args, backup, dryRun)
}
