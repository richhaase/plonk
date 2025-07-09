// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// MockDotfileConfigLoader implements DotfileConfigLoader for testing
type MockDotfileConfigLoader struct {
	targets map[string]string
}

func NewMockDotfileConfigLoader() *MockDotfileConfigLoader {
	return &MockDotfileConfigLoader{
		targets: make(map[string]string),
	}
}

func (m *MockDotfileConfigLoader) GetDotfileTargets() map[string]string {
	return m.targets
}

func (m *MockDotfileConfigLoader) SetTargets(targets map[string]string) {
	m.targets = targets
}

func TestNewDotfileProvider(t *testing.T) {
	homeDir := "/home/user"
	configDir := "/home/user/.config/plonk"
	configLoader := NewMockDotfileConfigLoader()
	
	provider := NewDotfileProvider(homeDir, configDir, configLoader)
	
	if provider == nil {
		t.Fatal("NewDotfileProvider() returned nil")
	}
	
	if provider.homeDir != homeDir {
		t.Errorf("provider.homeDir = %s, expected %s", provider.homeDir, homeDir)
	}
	
	if provider.configLoader != configLoader {
		t.Error("provider.configLoader not set correctly")
	}
}

func TestDotfileProvider_Domain(t *testing.T) {
	provider := NewDotfileProvider("/home/user", "/home/user/.config/plonk", NewMockDotfileConfigLoader())
	
	domain := provider.Domain()
	if domain != "dotfile" {
		t.Errorf("Domain() = %s, expected dotfile", domain)
	}
}

func TestDotfileProvider_GetConfiguredItems(t *testing.T) {
	configLoader := NewMockDotfileConfigLoader()
	configLoader.SetTargets(map[string]string{
		"zshrc":         "~/.zshrc",
		"gitconfig":     "~/.gitconfig",
		"config/nvim/":  "~/.config/nvim/",
	})
	
	provider := NewDotfileProvider("/home/user", "/home/user/.config/plonk", configLoader)
	
	items, err := provider.GetConfiguredItems()
	if err != nil {
		t.Fatalf("GetConfiguredItems() failed: %v", err)
	}
	
	if len(items) != 3 {
		t.Errorf("GetConfiguredItems() returned %d items, expected 3", len(items))
	}
	
	// Verify items structure
	itemsByName := make(map[string]ConfigItem)
	for _, item := range items {
		itemsByName[item.Name] = item
	}
	
	// Test .zshrc
	zshrcItem, exists := itemsByName[".zshrc"]
	if !exists {
		t.Error("Expected .zshrc item not found")
	} else {
		if zshrcItem.Metadata["source"] != "zshrc" {
			t.Errorf("zshrc item source = %v, expected zshrc", zshrcItem.Metadata["source"])
		}
		if zshrcItem.Metadata["destination"] != "~/.zshrc" {
			t.Errorf("zshrc item destination = %v, expected ~/.zshrc", zshrcItem.Metadata["destination"])
		}
	}
	
	// Test .gitconfig
	gitconfigItem, exists := itemsByName[".gitconfig"]
	if !exists {
		t.Error("Expected .gitconfig item not found")
	} else {
		if gitconfigItem.Metadata["source"] != "gitconfig" {
			t.Errorf("gitconfig item source = %v, expected gitconfig", gitconfigItem.Metadata["source"])
		}
	}
	
	// Test .config/nvim/
	nvimItem, exists := itemsByName[".config/nvim/"]
	if !exists {
		t.Error("Expected .config/nvim/ item not found")
	} else {
		if nvimItem.Metadata["source"] != "config/nvim/" {
			t.Errorf("nvim item source = %v, expected config/nvim/", nvimItem.Metadata["source"])
		}
	}
}

func TestDotfileProvider_GetActualItems(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create test dotfiles
	testFiles := []string{
		".zshrc",
		".gitconfig",
		".vimrc",
		"regular_file.txt", // Should be ignored (not a dotfile)
	}
	
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Create test dotfile directories
	testDirs := []string{
		".config",
		".ssh",
	}
	
	for _, dir := range testDirs {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
		
		// Add a file inside the directory to make it non-empty
		testFile := filepath.Join(dirPath, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file in directory %s: %v", dir, err)
		}
	}
	
	provider := NewDotfileProvider(tempDir, tempDir+"/.config/plonk", NewMockDotfileConfigLoader())
	
	items, err := provider.GetActualItems(context.Background())
	if err != nil {
		t.Fatalf("GetActualItems() failed: %v", err)
	}
	
	// Should find dotfiles but not regular files
	expectedDotfiles := []string{".zshrc", ".gitconfig", ".vimrc", ".config", ".ssh"}
	if len(items) != len(expectedDotfiles) {
		t.Errorf("GetActualItems() returned %d items, expected %d", len(items), len(expectedDotfiles))
	}
	
	// Verify items structure
	itemsByName := make(map[string]ActualItem)
	for _, item := range items {
		itemsByName[item.Name] = item
	}
	
	for _, expected := range expectedDotfiles {
		item, exists := itemsByName[expected]
		if !exists {
			t.Errorf("Expected dotfile %s not found in actual items", expected)
			continue
		}
		
		expectedPath := filepath.Join(tempDir, expected)
		if item.Path != expectedPath {
			t.Errorf("Item %s path = %s, expected %s", expected, item.Path, expectedPath)
		}
		
		if item.Metadata["path"] != expectedPath {
			t.Errorf("Item %s metadata path = %v, expected %s", expected, item.Metadata["path"], expectedPath)
		}
	}
	
	// Verify regular file is not included
	if _, exists := itemsByName["regular_file.txt"]; exists {
		t.Error("Regular file should not be included in dotfiles")
	}
}

func TestDotfileProvider_CreateItem(t *testing.T) {
	provider := NewDotfileProvider("/home/user", "/home/user/.config/plonk", NewMockDotfileConfigLoader())
	
	tests := []struct {
		name         string
		state        ItemState
		configured   *ConfigItem
		actual       *ActualItem
		expectedName string
		expectedPath string
	}{
		{
			name:         "managed dotfile",
			state:        StateManaged,
			configured:   &ConfigItem{Name: ".zshrc", Metadata: map[string]interface{}{"source": "zshrc", "destination": "~/.zshrc"}},
			actual:       &ActualItem{Name: ".zshrc", Path: "/home/user/.zshrc", Metadata: map[string]interface{}{"path": "/home/user/.zshrc"}},
			expectedName: ".zshrc",
			expectedPath: "/home/user/.zshrc",
		},
		{
			name:         "missing dotfile",
			state:        StateMissing,
			configured:   &ConfigItem{Name: ".gitconfig", Metadata: map[string]interface{}{"source": "gitconfig", "destination": "~/.gitconfig"}},
			actual:       nil,
			expectedName: ".gitconfig",
			expectedPath: "/home/user/.gitconfig", // expandPath converts ~/ to /home/user/
		},
		{
			name:         "untracked dotfile",
			state:        StateUntracked,
			configured:   nil,
			actual:       &ActualItem{Name: ".vimrc", Path: "/home/user/.vimrc", Metadata: map[string]interface{}{"path": "/home/user/.vimrc"}},
			expectedName: ".vimrc",
			expectedPath: "/home/user/.vimrc",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := provider.CreateItem(test.expectedName, test.state, test.configured, test.actual)
			
			if item.Name != test.expectedName {
				t.Errorf("item.Name = %s, expected %s", item.Name, test.expectedName)
			}
			
			if item.State != test.state {
				t.Errorf("item.State = %s, expected %s", item.State, test.state)
			}
			
			if item.Domain != "dotfile" {
				t.Errorf("item.Domain = %s, expected dotfile", item.Domain)
			}
			
			if item.Path != test.expectedPath {
				t.Errorf("item.Path = %s, expected %s", item.Path, test.expectedPath)
			}
			
			// Verify metadata is merged correctly
			if test.configured != nil {
				for key, value := range test.configured.Metadata {
					if item.Metadata[key] != value {
						t.Errorf("item.Metadata[%s] = %v, expected %v", key, item.Metadata[key], value)
					}
				}
			}
			
			if test.actual != nil {
				for key, value := range test.actual.Metadata {
					if item.Metadata[key] != value {
						t.Errorf("item.Metadata[%s] = %v, expected %v", key, item.Metadata[key], value)
					}
				}
			}
		})
	}
}

func TestDotfileProvider_GetActualItems_WithConfiguredDirectories(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create test dotfiles in home directory
	testFiles := []string{
		".zshrc",
		".gitconfig",
	}
	
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Create configured directory structure
	nvimDir := filepath.Join(tempDir, ".config", "nvim", "lua", "config")
	err := os.MkdirAll(nvimDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nvim directory: %v", err)
	}
	
	// Add files in the configured directory
	nvimFiles := []string{
		filepath.Join(tempDir, ".config", "nvim", "init.lua"),
		filepath.Join(tempDir, ".config", "nvim", "lua", "config", "options.lua"),
		filepath.Join(tempDir, ".config", "nvim", "lua", "config", "keymaps.lua"),
	}
	
	for _, file := range nvimFiles {
		err := os.WriteFile(file, []byte("-- nvim config"), 0644)
		if err != nil {
			t.Fatalf("Failed to create nvim file %s: %v", file, err)
		}
	}
	
	// Set up mock config loader with directory mapping
	configLoader := NewMockDotfileConfigLoader()
	configLoader.SetTargets(map[string]string{
		"zshrc":        "~/.zshrc",
		"gitconfig":    "~/.gitconfig",
		"config/nvim":  "~/.config/nvim",
	})
	
	provider := NewDotfileProvider(tempDir, tempDir+"/.config/plonk", configLoader)
	
	items, err := provider.GetActualItems(context.Background())
	if err != nil {
		t.Fatalf("GetActualItems() failed: %v", err)
	}
	
	// Should find dotfiles plus nvim directory files
	expectedItems := []string{
		".zshrc",
		".gitconfig",
		".config", // This will be found as a dotfile directory
		".config/nvim/init.lua",
		".config/nvim/lua/config/options.lua",
		".config/nvim/lua/config/keymaps.lua",
	}
	
	if len(items) != len(expectedItems) {
		t.Errorf("GetActualItems() returned %d items, expected %d", len(items), len(expectedItems))
		for i, item := range items {
			t.Logf("Item %d: %s", i, item.Name)
		}
	}
	
	// Verify all expected items are present
	itemsByName := make(map[string]ActualItem)
	for _, item := range items {
		itemsByName[item.Name] = item
	}
	
	for _, expected := range expectedItems {
		item, exists := itemsByName[expected]
		if !exists {
			t.Errorf("Expected item %s not found in actual items", expected)
			continue
		}
		
		// Verify the item has correct path
		expectedPath := filepath.Join(tempDir, expected)
		if item.Path != expectedPath {
			t.Errorf("Item %s path = %s, expected %s", expected, item.Path, expectedPath)
		}
	}
}

func TestDotfileProvider_GetActualItems_NoDuplicates(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create .vimrc both as a dotfile and as a configured file
	vimrcPath := filepath.Join(tempDir, ".vimrc")
	err := os.WriteFile(vimrcPath, []byte("vimrc content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .vimrc: %v", err)
	}
	
	// Set up mock config loader that includes .vimrc
	configLoader := NewMockDotfileConfigLoader()
	configLoader.SetTargets(map[string]string{
		"vimrc": "~/.vimrc",
	})
	
	provider := NewDotfileProvider(tempDir, tempDir+"/.config/plonk", configLoader)
	
	items, err := provider.GetActualItems(context.Background())
	if err != nil {
		t.Fatalf("GetActualItems() failed: %v", err)
	}
	
	// Should only find .vimrc once, not duplicated
	vimrcCount := 0
	for _, item := range items {
		if item.Name == ".vimrc" {
			vimrcCount++
		}
	}
	
	if vimrcCount != 1 {
		t.Errorf("Expected .vimrc to appear exactly once, found %d times", vimrcCount)
	}
}

func TestDotfileProvider_DestinationToName(t *testing.T) {
	provider := NewDotfileProvider("/home/user", "/home/user/.config/plonk", NewMockDotfileConfigLoader())
	
	tests := []struct {
		destination string
		expected    string
	}{
		{"~/.zshrc", ".zshrc"},
		{"~/.config/nvim/", ".config/nvim/"},
		{"~/.config/git/config", ".config/git/config"},
		{"/absolute/path/.vimrc", "/absolute/path/.vimrc"}, // Fixed: absolute paths are returned as-is
		{"relative/path/.bashrc", "relative/path/.bashrc"}, // Fixed: relative paths are returned as-is
	}
	
	for _, test := range tests {
		result := provider.manager.DestinationToName(test.destination)
		if result != test.expected {
			t.Errorf("DestinationToName(%s) = %s, expected %s", test.destination, result, test.expected)
		}
	}
}