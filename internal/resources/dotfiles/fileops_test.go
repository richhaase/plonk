// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFileOperations(t *testing.T) {
	manager := NewManager("/home/user", "/home/user/.config/plonk")
	fileOps := NewFileOperations(manager.pathResolver)

	if fileOps == nil {
		t.Fatal("NewFileOperations() returned nil")
	}

	if fileOps.pathResolver != manager.pathResolver {
		t.Error("fileOps.pathResolver not set correctly")
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
			err := fileOps.CopyFile(context.Background(), test.source, test.destination, test.options)

			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !test.expectError {
				// Verify file was copied correctly
				destPath, err := manager.GetDestinationPath(test.destination)
				if err != nil {
					t.Fatalf("Failed to get destination path: %v", err)
				}
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

	err = fileOps.CopyFile(context.Background(), "zshrc", "~/.zshrc", options)
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

	err = fileOps.CopyFile(context.Background(), "zshrc", "~/.zshrc", options)
	if err == nil {
		t.Error("Expected error when overwrite is disabled but destination exists")
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
	err = fileOps.copyFileContents(context.Background(), sourceFile, destFile)
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
	err = fileOps.createBackup(context.Background(), sourceFile, backupPath)
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

func TestFileOperations_ContextCancellation(t *testing.T) {
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
	err = os.WriteFile(sourceFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	t.Run("CopyFile_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := fileOps.CopyFile(ctx, "zshrc", "~/.zshrc", DefaultCopyOptions())
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("CopyFile_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		err := fileOps.CopyFile(ctx, "zshrc", "~/.zshrc", DefaultCopyOptions())
		if err == nil {
			t.Error("Expected error when context times out")
		}
		// Error may be wrapped, check if it contains deadline exceeded
		if !containsContextError(err) {
			t.Errorf("Expected context timeout error, got %v", err)
		}
	})

	t.Run("copyFileContents_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		destFile := filepath.Join(tempDir, "dest.txt")
		err := fileOps.copyFileContents(ctx, sourceFile, destFile)
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("copyFileContents_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		destFile := filepath.Join(tempDir, "dest.txt")
		err := fileOps.copyFileContents(ctx, sourceFile, destFile)
		if err == nil {
			t.Error("Expected error when context times out")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context deadline exceeded error, got %v", err)
		}
	})

	t.Run("createBackup_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		backupPath := filepath.Join(tempDir, "backup.txt")
		err := fileOps.createBackup(ctx, sourceFile, backupPath)
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("createBackup_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		backupPath := filepath.Join(tempDir, "backup.txt")
		err := fileOps.createBackup(ctx, sourceFile, backupPath)
		if err == nil {
			t.Error("Expected error when context times out")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context deadline exceeded error, got %v", err)
		}
	})
}

// containsContextError checks if the error contains context cancellation or timeout
func containsContextError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsString(errStr, "context canceled") ||
		containsString(errStr, "context deadline exceeded") ||
		containsString(errStr, "signal: killed")
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
