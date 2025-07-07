// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test DotfileStatus string representation
func TestDotfileStatusString(t *testing.T) {
	tests := []struct {
		status   DotfileStatus
		expected string
	}{
		{DotfileManaged, "managed"},
		{DotfileUntracked, "untracked"},
		{DotfileMissing, "missing"},
		{DotfileModified, "modified"},
		{DotfileStatus(999), "unknown"},
	}

	for _, test := range tests {
		if got := test.status.String(); got != test.expected {
			t.Errorf("DotfileStatus(%d).String() = %s, want %s", test.status, got, test.expected)
		}
	}
}

// Test DotfileInfo structure
func TestDotfileInfo(t *testing.T) {
	info := DotfileInfo{
		Path:         "/Users/test/.zshrc",
		Name:         ".zshrc",
		Status:       DotfileManaged,
		Source:       "/Users/test/.config/plonk/repo/.zshrc",
		LastModified: time.Now(),
		Size:         1024,
		IsDir:        false,
	}

	if info.Path != "/Users/test/.zshrc" {
		t.Errorf("Expected path to be '/Users/test/.zshrc', got %s", info.Path)
	}

	if info.Status != DotfileManaged {
		t.Errorf("Expected status to be DotfileManaged, got %v", info.Status)
	}

	if info.Status.String() != "managed" {
		t.Errorf("Expected status string to be 'managed', got %s", info.Status.String())
	}
}

// MockDotfilesManager for testing
type MockDotfilesManager struct {
	ManagedFiles   []DotfileInfo
	UntrackedFiles []DotfileInfo
	MissingFiles   []DotfileInfo
	ModifiedFiles  []DotfileInfo
	Error          error
}

func (m *MockDotfilesManager) ListManaged() ([]DotfileInfo, error) {
	return m.ManagedFiles, m.Error
}

func (m *MockDotfilesManager) ListUntracked() ([]DotfileInfo, error) {
	return m.UntrackedFiles, m.Error
}

func (m *MockDotfilesManager) ListMissing() ([]DotfileInfo, error) {
	return m.MissingFiles, m.Error
}

func (m *MockDotfilesManager) ListModified() ([]DotfileInfo, error) {
	return m.ModifiedFiles, m.Error
}

func (m *MockDotfilesManager) ListAll() ([]DotfileInfo, error) {
	all := make([]DotfileInfo, 0)
	all = append(all, m.ManagedFiles...)
	all = append(all, m.UntrackedFiles...)
	all = append(all, m.MissingFiles...)
	all = append(all, m.ModifiedFiles...)
	return all, m.Error
}

// Test DotfilesManager interface compliance
func TestDotfilesManagerInterface(t *testing.T) {
	mock := &MockDotfilesManager{
		ManagedFiles: []DotfileInfo{
			{Path: "/Users/test/.zshrc", Name: ".zshrc", Status: DotfileManaged},
		},
		UntrackedFiles: []DotfileInfo{
			{Path: "/Users/test/.bashrc", Name: ".bashrc", Status: DotfileUntracked},
		},
	}

	// Test that mock implements DotfilesManager interface
	var _ DotfilesManager = mock

	// Test ListManaged
	managed, err := mock.ListManaged()
	if err != nil {
		t.Errorf("ListManaged() returned error: %v", err)
	}
	if len(managed) != 1 {
		t.Errorf("Expected 1 managed file, got %d", len(managed))
	}
	if managed[0].Status != DotfileManaged {
		t.Errorf("Expected managed file status to be DotfileManaged, got %v", managed[0].Status)
	}

	// Test ListUntracked
	untracked, err := mock.ListUntracked()
	if err != nil {
		t.Errorf("ListUntracked() returned error: %v", err)
	}
	if len(untracked) != 1 {
		t.Errorf("Expected 1 untracked file, got %d", len(untracked))
	}
	if untracked[0].Status != DotfileUntracked {
		t.Errorf("Expected untracked file status to be DotfileUntracked, got %v", untracked[0].Status)
	}
}

// Test NewDotfilesManager constructor
func TestNewDotfilesManager(t *testing.T) {
	homeDir := "/tmp/test"
	plonkDir := "/tmp/test/.config/plonk"

	manager := NewDotfilesManager(homeDir, plonkDir)

	if manager == nil {
		t.Error("NewDotfilesManager should not return nil")
	}

	if manager.homeDir != homeDir {
		t.Errorf("Expected homeDir to be %s, got %s", homeDir, manager.homeDir)
	}

	if manager.plonkDir != plonkDir {
		t.Errorf("Expected plonkDir to be %s, got %s", plonkDir, manager.plonkDir)
	}

	// Test that it implements the interface
	var _ DotfilesManager = manager
}

// Test DotfilesManager with real filesystem (integration test)
func TestDotfilesManagerIntegration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "plonk-dotfiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test dotfiles
	testFiles := []string{".zshrc", ".gitconfig", ".tmux.conf"}
	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		err := os.WriteFile(path, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create some files that should be ignored
	ignoredFiles := []string{".DS_Store", ".git"}
	for _, file := range ignoredFiles {
		path := filepath.Join(tempDir, file)
		var err error
		if file == ".git" {
			err = os.Mkdir(path, 0755)
		} else {
			err = os.WriteFile(path, []byte("ignored"), 0644)
		}
		if err != nil {
			t.Fatalf("Failed to create ignored file %s: %v", file, err)
		}
	}

	// Create dotfiles manager
	plonkDir := filepath.Join(tempDir, ".config", "plonk")
	manager := NewDotfilesManager(tempDir, plonkDir)

	// Test ListUntracked (since we have no config, all should be untracked)
	untracked, err := manager.ListUntracked()
	if err != nil {
		t.Errorf("ListUntracked failed: %v", err)
	}

	// Should find our test files but not ignored ones
	if len(untracked) != len(testFiles) {
		t.Errorf("Expected %d untracked files, got %d", len(testFiles), len(untracked))
	}

	// Check that ignored files are not in the results
	for _, dotfile := range untracked {
		for _, ignored := range ignoredFiles {
			if dotfile.Name == ignored {
				t.Errorf("Ignored file %s should not appear in untracked files", ignored)
			}
		}
	}
}
