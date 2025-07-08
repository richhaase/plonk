// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegration_FullDotfileWorkflow tests the complete dotfile management workflow
func TestIntegration_FullDotfileWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create config directory structure
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create home directory
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// 1. Create dotfiles in config directory
	dotfiles := map[string]string{
		"zshrc":     "# ZSH configuration\nexport TEST=zsh\n",
		"gitconfig": "[user]\n\tname = Test User\n",
		"vimrc":     "\" Vim configuration\nset number\n",
	}
	
	for source, content := range dotfiles {
		sourcePath := filepath.Join(configDir, source)
		err := os.WriteFile(sourcePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create dotfile %s: %v", source, err)
		}
	}
	
	// 2. Test file deployment
	deployments := map[string]string{
		"zshrc":     "~/.zshrc",
		"gitconfig": "~/.gitconfig",
		"vimrc":     "~/.vimrc",
	}
	
	options := DefaultCopyOptions()
	for source, destination := range deployments {
		// Check if file needs update (should be true initially)
		needsUpdate, err := fileOps.FileNeedsUpdate(source, destination)
		if err != nil {
			t.Errorf("FileNeedsUpdate() failed for %s: %v", source, err)
		}
		if !needsUpdate {
			t.Errorf("FileNeedsUpdate() should return true for initial deployment of %s", source)
		}
		
		// Deploy the file
		err = fileOps.CopyFile(source, destination, options)
		if err != nil {
			t.Errorf("CopyFile() failed for %s: %v", source, err)
		}
		
		// Verify deployment
		destPath := manager.GetDestinationPath(destination)
		if !manager.FileExists(destPath) {
			t.Errorf("Deployed file %s does not exist", destPath)
		}
		
		// Verify content
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read deployed file %s: %v", destPath, err)
		} else if string(content) != dotfiles[source] {
			t.Errorf("Deployed file %s content mismatch: got %s, expected %s", 
				destPath, string(content), dotfiles[source])
		}
		
		// Check if file needs update after deployment (should be false)
		needsUpdate, err = fileOps.FileNeedsUpdate(source, destination)
		if err != nil {
			t.Errorf("FileNeedsUpdate() failed after deployment for %s: %v", source, err)
		}
		if needsUpdate {
			t.Errorf("FileNeedsUpdate() should return false after deployment of %s", source)
		}
	}
	
	// 3. Test backup functionality
	// Update a source file and redeploy with backup
	updatedContent := "# Updated ZSH configuration\nexport TEST=updated\n"
	sourcePath := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(sourcePath, []byte(updatedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update source file: %v", err)
	}
	
	// Deploy with backup
	err = fileOps.CopyFile("zshrc", "~/.zshrc", options)
	if err != nil {
		t.Errorf("CopyFile() with backup failed: %v", err)
	}
	
	// Verify backup was created
	destPath := manager.GetDestinationPath("~/.zshrc")
	backupFiles, err := filepath.Glob(destPath + ".backup.*")
	if err != nil {
		t.Errorf("Failed to find backup files: %v", err)
	}
	if len(backupFiles) != 1 {
		t.Errorf("Expected 1 backup file, found %d", len(backupFiles))
	} else {
		// Verify backup contains original content
		backupContent, err := os.ReadFile(backupFiles[0])
		if err != nil {
			t.Errorf("Failed to read backup file: %v", err)
		} else if string(backupContent) != dotfiles["zshrc"] {
			t.Errorf("Backup content mismatch: got %s, expected %s", 
				string(backupContent), dotfiles["zshrc"])
		}
	}
	
	// Verify destination has updated content
	newContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Errorf("Failed to read updated file: %v", err)
	} else if string(newContent) != updatedContent {
		t.Errorf("Updated file content mismatch: got %s, expected %s", 
			string(newContent), updatedContent)
	}
	
	// 4. Test file removal
	err = fileOps.RemoveFile("~/.vimrc")
	if err != nil {
		t.Errorf("RemoveFile() failed: %v", err)
	}
	
	vimrcPath := manager.GetDestinationPath("~/.vimrc")
	if manager.FileExists(vimrcPath) {
		t.Error("File should have been removed")
	}
}

// TestIntegration_DirectoryWorkflow tests directory-based dotfile management
func TestIntegration_DirectoryWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create config directory structure
	nvimConfigDir := filepath.Join(configDir, "config", "nvim")
	err := os.MkdirAll(nvimConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nvim config directory: %v", err)
	}
	
	// Create subdirectories
	pluginDir := filepath.Join(nvimConfigDir, "plugin")
	err = os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}
	
	// Create home directory
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Create test files in the directory structure
	testFiles := map[string]string{
		"init.vim":           "\" Main neovim config\n",
		"plugins.vim":        "\" Plugins configuration\n",
		"plugin/keybinds.vim": "\" Keybindings\n",
	}
	
	for file, content := range testFiles {
		filePath := filepath.Join(nvimConfigDir, file)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// 1. Test directory expansion
	items, err := manager.ExpandDirectory("config/nvim/", "~/.config/nvim/")
	if err != nil {
		t.Fatalf("ExpandDirectory() failed: %v", err)
	}
	
	if len(items) != len(testFiles) {
		t.Errorf("ExpandDirectory() returned %d items, expected %d", len(items), len(testFiles))
	}
	
	// 2. Test directory copy
	options := DefaultCopyOptions()
	err = fileOps.CopyDirectory("config/nvim/", "~/.config/nvim/", options)
	if err != nil {
		t.Fatalf("CopyDirectory() failed: %v", err)
	}
	
	// 3. Verify all files were copied correctly
	for file, expectedContent := range testFiles {
		destPath := filepath.Join(homeDir, ".config", "nvim", file)
		
		if !manager.FileExists(destPath) {
			t.Errorf("File %s was not copied", destPath)
			continue
		}
		
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", destPath, err)
		} else if string(content) != expectedContent {
			t.Errorf("File %s content mismatch: got %s, expected %s", 
				destPath, string(content), expectedContent)
		}
	}
	
	// 4. Test individual file updates within directory
	// Update one file in source
	updatedContent := "\" Updated keybindings\nmap <leader>q :q<CR>\n"
	keybindsPath := filepath.Join(nvimConfigDir, "plugin", "keybinds.vim")
	err = os.WriteFile(keybindsPath, []byte(updatedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update keybinds file: %v", err)
	}
	
	// Copy directory again (should create backup of changed file)
	err = fileOps.CopyDirectory("config/nvim/", "~/.config/nvim/", options)
	if err != nil {
		t.Errorf("CopyDirectory() update failed: %v", err)
	}
	
	// Verify updated content
	destKeybindsPath := filepath.Join(homeDir, ".config", "nvim", "plugin", "keybinds.vim")
	content, err := os.ReadFile(destKeybindsPath)
	if err != nil {
		t.Errorf("Failed to read updated keybinds file: %v", err)
	} else if string(content) != updatedContent {
		t.Errorf("Updated keybinds content mismatch: got %s, expected %s", 
			string(content), updatedContent)
	}
	
	// Verify backup was created
	backupFiles, err := filepath.Glob(destKeybindsPath + ".backup.*")
	if err != nil {
		t.Errorf("Failed to find backup files: %v", err)
	}
	if len(backupFiles) != 1 {
		t.Errorf("Expected 1 backup file for keybinds, found %d", len(backupFiles))
	}
}

// TestIntegration_PathResolution tests path resolution and validation
func TestIntegration_PathResolution(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	
	// Create config directory
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create test source file
	testFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test path resolution
	tests := []struct {
		name        string
		source      string
		destination string
		expectValid bool
	}{
		{
			name:        "valid tilde path",
			source:      "zshrc",
			destination: "~/.zshrc",
			expectValid: true,
		},
		{
			name:        "valid absolute path",
			source:      "zshrc",
			destination: homeDir + "/.zshrc",
			expectValid: true,
		},
		{
			name:        "invalid relative path",
			source:      "zshrc",
			destination: "relative/path",
			expectValid: false,
		},
		{
			name:        "nonexistent source",
			source:      "nonexistent",
			destination: "~/.zshrc",
			expectValid: false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := manager.ValidatePaths(test.source, test.destination)
			
			if test.expectValid && err != nil {
				t.Errorf("ValidatePaths() should succeed but failed: %v", err)
			}
			if !test.expectValid && err == nil {
				t.Error("ValidatePaths() should fail but succeeded")
			}
			
			if test.expectValid {
				// Test path expansion
				expandedPath := manager.GetDestinationPath(test.destination)
				if test.destination == "~/.zshrc" {
					expectedPath := filepath.Join(homeDir, ".zshrc")
					if expandedPath != expectedPath {
						t.Errorf("GetDestinationPath() = %s, expected %s", expandedPath, expectedPath)
					}
				}
				
				// Test name generation
				name := manager.DestinationToName(test.destination)
				if test.destination == "~/.zshrc" {
					expectedName := ".zshrc"
					if name != expectedName {
						t.Errorf("DestinationToName() = %s, expected %s", name, expectedName)
					}
				}
			}
		})
	}
}

// TestIntegration_ErrorHandling tests error handling in various scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create config directory
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create home directory
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Test 1: Copy nonexistent file
	options := DefaultCopyOptions()
	err = fileOps.CopyFile("nonexistent", "~/.zshrc", options)
	if err == nil {
		t.Error("CopyFile() should fail for nonexistent source")
	}
	
	// Test 2: Copy directory as file
	testDir := filepath.Join(configDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	err = fileOps.CopyDirectory("nonexistent", "~/.config/", options)
	if err == nil {
		t.Error("CopyDirectory() should fail for nonexistent source")
	}
	
	// Test 3: Overwrite protection
	// Create source file
	sourceFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(sourceFile, []byte("new content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create existing destination
	destFile := filepath.Join(homeDir, ".zshrc")
	err = os.WriteFile(destFile, []byte("existing content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}
	
	// Try to copy with overwrite disabled
	noOverwriteOptions := CopyOptions{
		CreateBackup:      false,
		OverwriteExisting: false,
	}
	
	err = fileOps.CopyFile("zshrc", "~/.zshrc", noOverwriteOptions)
	if err == nil {
		t.Error("CopyFile() should fail when overwrite is disabled and destination exists")
	}
	
	// Test 4: Permission errors (simulate by creating read-only directory)
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.MkdirAll(readOnlyDir, 0555) // Read-only directory
	if err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}
	
	// This test might not work on all systems, so we'll check if it actually fails
	readOnlyFile := filepath.Join(readOnlyDir, "test.txt")
	err = fileOps.copyFileContents(sourceFile, readOnlyFile)
	// We can't guarantee this will fail on all systems, so we just log the result
	if err != nil {
		t.Logf("Expected permission error occurred: %v", err)
	}
}

// TestIntegration_ListDotfiles tests dotfile discovery functionality
func TestIntegration_ListDotfiles(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	
	// Create home directory with various files
	err := os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Create test files and directories
	testItems := []struct {
		name     string
		isDir    bool
		isDotfile bool
	}{
		{".zshrc", false, true},
		{".gitconfig", false, true},
		{".config", true, true},
		{".ssh", true, true},
		{"regular_file.txt", false, false},
		{"regular_dir", true, false},
		{".hidden_file", false, true},
	}
	
	for _, item := range testItems {
		path := filepath.Join(homeDir, item.name)
		if item.isDir {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", item.name, err)
			}
			// Add a file to make directory non-empty
			testFile := filepath.Join(path, "test.txt")
			err = os.WriteFile(testFile, []byte("test"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file in directory %s: %v", item.name, err)
			}
		} else {
			err := os.WriteFile(path, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", item.name, err)
			}
		}
	}
	
	// List dotfiles
	dotfiles, err := manager.ListDotfiles(homeDir)
	if err != nil {
		t.Fatalf("ListDotfiles() failed: %v", err)
	}
	
	// Count expected dotfiles
	expectedCount := 0
	expectedDotfiles := make(map[string]bool)
	for _, item := range testItems {
		if item.isDotfile && item.name != "." && item.name != ".." {
			expectedCount++
			expectedDotfiles[item.name] = true
		}
	}
	
	if len(dotfiles) != expectedCount {
		t.Errorf("ListDotfiles() returned %d items, expected %d", len(dotfiles), expectedCount)
	}
	
	// Verify all expected dotfiles are present
	foundDotfiles := make(map[string]bool)
	for _, df := range dotfiles {
		foundDotfiles[df] = true
	}
	
	for expected := range expectedDotfiles {
		if !foundDotfiles[expected] {
			t.Errorf("Expected dotfile %s not found", expected)
		}
	}
	
	// Verify non-dotfiles are not present
	for _, item := range testItems {
		if !item.isDotfile && foundDotfiles[item.name] {
			t.Errorf("Non-dotfile %s should not be in results", item.name)
		}
	}
}