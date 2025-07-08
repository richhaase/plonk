// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFileOperations(t *testing.T) {
	manager := NewManager("/home/user", "/home/user/.config/plonk")
	fileOps := NewFileOperations(manager)
	
	if fileOps == nil {
		t.Fatal("NewFileOperations() returned nil")
	}
	
	if fileOps.manager != manager {
		t.Error("fileOps.manager not set correctly")
	}
}

func TestDefaultCopyOptions(t *testing.T) {
	options := DefaultCopyOptions()
	
	if !options.CreateBackup {
		t.Error("DefaultCopyOptions().CreateBackup should be true")
	}
	
	if options.BackupSuffix != ".backup" {
		t.Errorf("DefaultCopyOptions().BackupSuffix = %s, expected .backup", options.BackupSuffix)
	}
	
	if !options.OverwriteExisting {
		t.Error("DefaultCopyOptions().OverwriteExisting should be true")
	}
}

func TestFileOperations_CopyFile(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create config directory and source file
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	sourceFile := filepath.Join(configDir, "zshrc")
	sourceContent := "# Test zshrc content\nexport TEST=value\n"
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create home directory
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	tests := []struct {
		name            string
		source          string
		destination     string
		options         CopyOptions
		expectError     bool
		expectedContent string
	}{
		{
			name:            "basic copy",
			source:          "zshrc",
			destination:     "~/.zshrc",
			options:         DefaultCopyOptions(),
			expectError:     false,
			expectedContent: sourceContent,
		},
		{
			name:        "nonexistent source",
			source:      "nonexistent",
			destination: "~/.zshrc",
			options:     DefaultCopyOptions(),
			expectError: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fileOps.CopyFile(test.source, test.destination, test.options)
			
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !test.expectError {
				// Verify file was copied correctly
				destPath := manager.GetDestinationPath(test.destination)
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
				} else if string(content) != test.expectedContent {
					t.Errorf("Destination content = %s, expected %s", string(content), test.expectedContent)
				}
			}
		})
	}
}

func TestFileOperations_CopyFile_WithBackup(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Create source file
	sourceFile := filepath.Join(configDir, "zshrc")
	sourceContent := "# New zshrc content\n"
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create existing destination file
	destPath := filepath.Join(homeDir, ".zshrc")
	originalContent := "# Original zshrc content\n"
	err = os.WriteFile(destPath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}
	
	// Copy with backup
	options := CopyOptions{
		CreateBackup:      true,
		BackupSuffix:      ".backup",
		OverwriteExisting: true,
	}
	
	err = fileOps.CopyFile("zshrc", "~/.zshrc", options)
	if err != nil {
		t.Fatalf("CopyFile() failed: %v", err)
	}
	
	// Verify destination file was updated
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(content) != sourceContent {
		t.Errorf("Destination content = %s, expected %s", string(content), sourceContent)
	}
	
	// Verify backup was created
	backupFiles, err := filepath.Glob(destPath + ".backup.*")
	if err != nil {
		t.Fatalf("Failed to find backup files: %v", err)
	}
	if len(backupFiles) != 1 {
		t.Errorf("Expected 1 backup file, found %d", len(backupFiles))
	} else {
		backupContent, err := os.ReadFile(backupFiles[0])
		if err != nil {
			t.Errorf("Failed to read backup file: %v", err)
		} else if string(backupContent) != originalContent {
			t.Errorf("Backup content = %s, expected %s", string(backupContent), originalContent)
		}
	}
}

func TestFileOperations_CopyFile_OverwriteDisabled(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Create source file
	sourceFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(sourceFile, []byte("new content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create existing destination file
	destPath := filepath.Join(homeDir, ".zshrc")
	err = os.WriteFile(destPath, []byte("original content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}
	
	// Copy with overwrite disabled
	options := CopyOptions{
		CreateBackup:      false,
		OverwriteExisting: false,
	}
	
	err = fileOps.CopyFile("zshrc", "~/.zshrc", options)
	if err == nil {
		t.Error("Expected error when overwrite is disabled but destination exists")
	}
}

func TestFileOperations_CopyDirectory(t *testing.T) {
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
	
	// Create subdirectory
	colorsDir := filepath.Join(nvimConfigDir, "colors")
	err = os.MkdirAll(colorsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create colors directory: %v", err)
	}
	
	// Create test files
	testFiles := map[string]string{
		"init.vim":        "\" Neovim init file\n",
		"plugins.vim":     "\" Plugins configuration\n",
		"colors/theme.vim": "\" Color theme\n",
	}
	
	for file, content := range testFiles {
		filePath := filepath.Join(nvimConfigDir, file)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Create home directory
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Copy directory
	options := DefaultCopyOptions()
	err = fileOps.CopyDirectory("config/nvim/", "~/.config/nvim/", options)
	if err != nil {
		t.Fatalf("CopyDirectory() failed: %v", err)
	}
	
	// Verify all files were copied
	for file, expectedContent := range testFiles {
		destPath := filepath.Join(homeDir, ".config", "nvim", file)
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", file, err)
		} else if string(content) != expectedContent {
			t.Errorf("File %s content = %s, expected %s", file, string(content), expectedContent)
		}
	}
}

func TestFileOperations_CopyDirectory_NotADirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create config directory and a file (not directory)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	testFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Try to copy file as directory
	options := DefaultCopyOptions()
	err = fileOps.CopyDirectory("zshrc", "~/.zshrc", options)
	if err == nil {
		t.Error("Expected error when source is not a directory")
	}
}

func TestFileOperations_RemoveFile(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create home directory and test file
	err := os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	testFile := filepath.Join(homeDir, ".zshrc")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Remove the file
	err = fileOps.RemoveFile("~/.zshrc")
	if err != nil {
		t.Errorf("RemoveFile() failed: %v", err)
	}
	
	// Verify file was removed
	if manager.FileExists(testFile) {
		t.Error("File should have been removed")
	}
}

func TestFileOperations_RemoveFile_Nonexistent(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Try to remove nonexistent file (should not error)
	err := fileOps.RemoveFile("~/.nonexistent")
	if err != nil {
		t.Errorf("RemoveFile() should not error for nonexistent file: %v", err)
	}
}

func TestFileOperations_FileNeedsUpdate(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create directories
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}
	
	// Create source file
	sourceFile := filepath.Join(configDir, "zshrc")
	err = os.WriteFile(sourceFile, []byte("source content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	tests := []struct {
		name             string
		createDest       bool
		makeSourceNewer  bool
		expected         bool
	}{
		{
			name:       "destination doesn't exist",
			createDest: false,
			expected:   true,
		},
		{
			name:            "source is newer",
			createDest:      true,
			makeSourceNewer: true,
			expected:        true,
		},
		{
			name:            "destination is newer",
			createDest:      true,
			makeSourceNewer: false,
			expected:        false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destFile := filepath.Join(homeDir, ".zshrc")
			
			// Clean up from previous test
			os.Remove(destFile)
			
			if test.createDest {
				// Create destination file
				err = os.WriteFile(destFile, []byte("dest content"), 0644)
				if err != nil {
					t.Fatalf("Failed to create destination file: %v", err)
				}
				
				if test.makeSourceNewer {
					// Sleep to ensure different timestamps
					time.Sleep(10 * time.Millisecond)
					
					// Touch source file to make it newer
					err = os.WriteFile(sourceFile, []byte("updated source content"), 0644)
					if err != nil {
						t.Fatalf("Failed to update source file: %v", err)
					}
				}
			}
			
			needsUpdate, err := fileOps.FileNeedsUpdate("zshrc", "~/.zshrc")
			if err != nil {
				t.Errorf("FileNeedsUpdate() failed: %v", err)
			}
			
			if needsUpdate != test.expected {
				t.Errorf("FileNeedsUpdate() = %v, expected %v", needsUpdate, test.expected)
			}
		})
	}
}

func TestFileOperations_FileNeedsUpdate_NonexistentSource(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	_, err := fileOps.FileNeedsUpdate("nonexistent", "~/.zshrc")
	if err == nil {
		t.Error("FileNeedsUpdate() should error for nonexistent source")
	}
	
	expectedError := "source file does not exist"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got: %v", expectedError, err)
	}
}

func TestFileOperations_GetFileInfo(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	info, err := fileOps.GetFileInfo(testFile)
	if err != nil {
		t.Errorf("GetFileInfo() failed: %v", err)
	}
	
	if info.Size() != int64(len(testContent)) {
		t.Errorf("File size = %d, expected %d", info.Size(), len(testContent))
	}
	
	if info.IsDir() {
		t.Error("File should not be a directory")
	}
}

func TestFileOperations_GetFileInfo_Nonexistent(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	_, err := fileOps.GetFileInfo("/nonexistent/file.txt")
	if err == nil {
		t.Error("GetFileInfo() should error for nonexistent file")
	}
}

func TestFileOperations_copyFileContents_PreservesPermissions(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create source file with specific permissions
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	destFile := filepath.Join(tempDir, "dest.txt")
	
	// Copy file contents
	err = fileOps.copyFileContents(sourceFile, destFile)
	if err != nil {
		t.Fatalf("copyFileContents() failed: %v", err)
	}
	
	// Verify permissions were preserved
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}
	
	destInfo, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}
	
	if sourceInfo.Mode() != destInfo.Mode() {
		t.Errorf("File permissions not preserved: source=%v, dest=%v", sourceInfo.Mode(), destInfo.Mode())
	}
}

func TestFileOperations_createBackup_WithTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "plonk")
	homeDir := filepath.Join(tempDir, "home")
	
	manager := NewManager(homeDir, configDir)
	fileOps := NewFileOperations(manager)
	
	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	sourceContent := "backup test content"
	err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	backupPath := filepath.Join(tempDir, "backup.txt")
	
	// Create backup
	err = fileOps.createBackup(sourceFile, backupPath)
	if err != nil {
		t.Fatalf("createBackup() failed: %v", err)
	}
	
	// Verify backup file was created with timestamp
	backupFiles, err := filepath.Glob(backupPath + ".*")
	if err != nil {
		t.Fatalf("Failed to find backup files: %v", err)
	}
	
	if len(backupFiles) != 1 {
		t.Errorf("Expected 1 backup file, found %d", len(backupFiles))
	} else {
		// Verify timestamp format in filename
		filename := filepath.Base(backupFiles[0])
		if !strings.Contains(filename, "backup.txt.") {
			t.Errorf("Backup filename doesn't contain expected pattern: %s", filename)
		}
		
		// Verify content
		content, err := os.ReadFile(backupFiles[0])
		if err != nil {
			t.Errorf("Failed to read backup file: %v", err)
		} else if string(content) != sourceContent {
			t.Errorf("Backup content = %s, expected %s", string(content), sourceContent)
		}
	}
}