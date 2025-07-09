// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	homeDir := "/home/user"
	configDir := "/home/user/.config/plonk"

	manager := NewManager(homeDir, configDir)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.homeDir != homeDir {
		t.Errorf("manager.homeDir = %s, expected %s", manager.homeDir, homeDir)
	}

	if manager.configDir != configDir {
		t.Errorf("manager.configDir = %s, expected %s", manager.configDir, configDir)
	}
}

func TestManager_ExpandPath(t *testing.T) {
	homeDir := "/home/user"
	configDir := "/home/user/.config/plonk"
	manager := NewManager(homeDir, configDir)

	tests := []struct {
		path     string
		expected string
	}{
		{"~/.zshrc", "/home/user/.zshrc"},
		{"~/.config/nvim/init.vim", "/home/user/.config/nvim/init.vim"},
		{"~/dotfiles/test", "/home/user/dotfiles/test"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, test := range tests {
		result := manager.ExpandPath(test.path)
		if result != test.expected {
			t.Errorf("ExpandPath(%s) = %s, expected %s", test.path, result, test.expected)
		}
	}
}

func TestManager_DestinationToName(t *testing.T) {
	homeDir := "/home/user"
	configDir := "/home/user/.config/plonk"
	manager := NewManager(homeDir, configDir)

	tests := []struct {
		destination string
		expected    string
	}{
		{"~/.zshrc", ".zshrc"},
		{"~/.config/nvim/", ".config/nvim/"},
		{"~/.config/git/config", ".config/git/config"},
		{"/absolute/path/.vimrc", "/absolute/path/.vimrc"},
		{"relative/path/.bashrc", "relative/path/.bashrc"},
		{"", ""},
	}

	for _, test := range tests {
		result := manager.DestinationToName(test.destination)
		if result != test.expected {
			t.Errorf("DestinationToName(%s) = %s, expected %s", test.destination, result, test.expected)
		}
	}
}

func TestManager_GetSourcePath(t *testing.T) {
	homeDir := "/home/user"
	configDir := "/home/user/.config/plonk"
	manager := NewManager(homeDir, configDir)

	tests := []struct {
		source   string
		expected string
	}{
		{"zshrc", "/home/user/.config/plonk/zshrc"},
		{"config/nvim/init.vim", "/home/user/.config/plonk/config/nvim/init.vim"},
		{"dot_gitconfig", "/home/user/.config/plonk/dot_gitconfig"},
		{"", "/home/user/.config/plonk"},
	}

	for _, test := range tests {
		result := manager.GetSourcePath(test.source)
		if result != test.expected {
			t.Errorf("GetSourcePath(%s) = %s, expected %s", test.source, result, test.expected)
		}
	}
}

func TestManager_GetDestinationPath(t *testing.T) {
	homeDir := "/home/user"
	configDir := "/home/user/.config/plonk"
	manager := NewManager(homeDir, configDir)

	tests := []struct {
		destination string
		expected    string
	}{
		{"~/.zshrc", "/home/user/.zshrc"},
		{"~/.config/nvim/", "/home/user/.config/nvim"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, test := range tests {
		result := manager.GetDestinationPath(test.destination)
		if result != test.expected {
			t.Errorf("GetDestinationPath(%s) = %s, expected %s", test.destination, result, test.expected)
		}
	}
}

func TestManager_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, ".config", "plonk"))

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{testFile, true},
		{filepath.Join(tempDir, "nonexistent.txt"), false},
		{tempDir, true}, // Directory should also exist
	}

	for _, test := range tests {
		result := manager.FileExists(test.path)
		if result != test.expected {
			t.Errorf("FileExists(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestManager_IsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, ".config", "plonk"))

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{testFile, false},
		{testDir, true},
		{tempDir, true},
		{filepath.Join(tempDir, "nonexistent"), false},
	}

	for _, test := range tests {
		result := manager.IsDirectory(test.path)
		if result != test.expected {
			t.Errorf("IsDirectory(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestManager_ListDotfiles(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, filepath.Join(tempDir, ".config", "plonk"))

	// Create test files and directories
	testItems := []struct {
		name  string
		isDir bool
		isDot bool
	}{
		{".zshrc", false, true},
		{".gitconfig", false, true},
		{".config", true, true},
		{"regular_file.txt", false, false},
		{"regular_dir", true, false},
		{"..", true, false}, // Should be ignored
		{".", true, false},  // Should be ignored
	}

	for _, item := range testItems {
		path := filepath.Join(tempDir, item.name)
		if item.isDir {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", item.name, err)
			}
		} else {
			err := os.WriteFile(path, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", item.name, err)
			}
		}
	}

	dotfiles, err := manager.ListDotfiles(tempDir)
	if err != nil {
		t.Fatalf("ListDotfiles() failed: %v", err)
	}

	// Should only include dotfiles (starting with . but not . or ..)
	expectedDotfiles := []string{".zshrc", ".gitconfig", ".config"}
	if len(dotfiles) != len(expectedDotfiles) {
		t.Errorf("ListDotfiles() returned %d items, expected %d", len(dotfiles), len(expectedDotfiles))
	}

	dotfileMap := make(map[string]bool)
	for _, df := range dotfiles {
		dotfileMap[df] = true
	}

	for _, expected := range expectedDotfiles {
		if !dotfileMap[expected] {
			t.Errorf("Expected dotfile %s not found in results", expected)
		}
	}

	// Verify non-dotfiles are excluded
	if dotfileMap["regular_file.txt"] {
		t.Error("Regular file should not be included in dotfiles")
	}
	if dotfileMap["regular_dir"] {
		t.Error("Regular directory should not be included in dotfiles")
	}
}

func TestManager_CreateDotfileInfo(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	manager := NewManager(tempDir, configDir)

	// Create config directory and test file
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	testFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info := manager.CreateDotfileInfo("zshrc", "~/.zshrc")

	if info.Name != ".zshrc" {
		t.Errorf("info.Name = %s, expected .zshrc", info.Name)
	}

	if info.Source != "zshrc" {
		t.Errorf("info.Source = %s, expected zshrc", info.Source)
	}

	if info.Destination != "~/.zshrc" {
		t.Errorf("info.Destination = %s, expected ~/.zshrc", info.Destination)
	}

	if info.IsDirectory != false {
		t.Errorf("info.IsDirectory = %v, expected false", info.IsDirectory)
	}

	// Verify metadata
	if info.Metadata["source"] != "zshrc" {
		t.Errorf("info.Metadata[\"source\"] = %v, expected zshrc", info.Metadata["source"])
	}

	if info.Metadata["destination"] != "~/.zshrc" {
		t.Errorf("info.Metadata[\"destination\"] = %v, expected ~/.zshrc", info.Metadata["destination"])
	}
}

func TestManager_ValidatePaths(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	manager := NewManager(tempDir, configDir)

	// Create config directory and test file
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	testFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		source      string
		destination string
		expectError bool
	}{
		{
			name:        "valid paths",
			source:      "zshrc",
			destination: "~/.zshrc",
			expectError: false,
		},
		{
			name:        "valid absolute destination",
			source:      "zshrc",
			destination: "/home/user/.zshrc",
			expectError: false,
		},
		{
			name:        "nonexistent source",
			source:      "nonexistent",
			destination: "~/.zshrc",
			expectError: true,
		},
		{
			name:        "invalid destination format",
			source:      "zshrc",
			destination: "relative/path",
			expectError: true,
		},
		{
			name:        "empty destination",
			source:      "zshrc",
			destination: "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := manager.ValidatePaths(test.source, test.destination)
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestManager_ExpandDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	manager := NewManager(tempDir, configDir)

	// Create config directory structure
	testConfigDir := filepath.Join(configDir, "config", "nvim")
	err := os.MkdirAll(testConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	// Create test files in the directory
	testFiles := []string{
		"init.vim",
		"plugins.vim",
		"colors/theme.vim",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(testConfigDir, file)
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}

		err = os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	items, err := manager.ExpandDirectory("config/nvim/", "~/.config/nvim/")
	if err != nil {
		t.Fatalf("ExpandDirectory() failed: %v", err)
	}

	if len(items) != len(testFiles) {
		t.Errorf("ExpandDirectory() returned %d items, expected %d", len(items), len(testFiles))
	}

	// Verify each item
	itemMap := make(map[string]DotfileInfo)
	for _, item := range items {
		itemMap[item.Source] = item
	}

	expectedItems := []struct {
		source      string
		destination string
		name        string
	}{
		{"config/nvim/init.vim", "~/.config/nvim/init.vim", ".config/nvim/init.vim"},
		{"config/nvim/plugins.vim", "~/.config/nvim/plugins.vim", ".config/nvim/plugins.vim"},
		{"config/nvim/colors/theme.vim", "~/.config/nvim/colors/theme.vim", ".config/nvim/colors/theme.vim"},
	}

	for _, expected := range expectedItems {
		item, exists := itemMap[expected.source]
		if !exists {
			t.Errorf("Expected item with source %s not found", expected.source)
			continue
		}

		if item.Destination != expected.destination {
			t.Errorf("Item %s destination = %s, expected %s", expected.source, item.Destination, expected.destination)
		}

		if item.Name != expected.name {
			t.Errorf("Item %s name = %s, expected %s", expected.source, item.Name, expected.name)
		}

		if item.IsDirectory != false {
			t.Errorf("Item %s IsDirectory = %v, expected false", expected.source, item.IsDirectory)
		}

		if item.ParentDir != "config/nvim/" {
			t.Errorf("Item %s ParentDir = %s, expected config/nvim/", expected.source, item.ParentDir)
		}

		// Verify metadata
		if item.Metadata["source"] != expected.source {
			t.Errorf("Item %s metadata source = %v, expected %s", expected.source, item.Metadata["source"], expected.source)
		}

		if item.Metadata["destination"] != expected.destination {
			t.Errorf("Item %s metadata destination = %v, expected %s", expected.source, item.Metadata["destination"], expected.destination)
		}

		if item.Metadata["parent_dir"] != "config/nvim/" {
			t.Errorf("Item %s metadata parent_dir = %v, expected config/nvim/", expected.source, item.Metadata["parent_dir"])
		}
	}
}

func TestManager_ExpandDirectory_NonexistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	manager := NewManager(tempDir, configDir)

	_, err := manager.ExpandDirectory("nonexistent/", "~/.nonexistent/")
	if err == nil {
		t.Error("ExpandDirectory() should return error for nonexistent directory")
	}
}
