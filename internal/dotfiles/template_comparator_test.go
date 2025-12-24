// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateAwareComparator_CompareFiles(t *testing.T) {
	t.Run("template matches deployed when rendered content equals", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create local.yaml with variables
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}
		localFile := filepath.Join(localDir, LocalVarsFile)
		if err := os.WriteFile(localFile, []byte("name: TestUser"), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		// Create template file
		templatePath := filepath.Join(tempDir, "config.tmpl")
		if err := os.WriteFile(templatePath, []byte("user = {{.name}}"), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		// Create deployed file with rendered content
		deployedPath := filepath.Join(tempDir, "deployed_config")
		if err := os.WriteFile(deployedPath, []byte("user = TestUser"), 0644); err != nil {
			t.Fatalf("Failed to write deployed file: %v", err)
		}

		templateProcessor := NewTemplateProcessor(tempDir)
		baseComparator := NewFileComparator()
		comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

		same, err := comparator.CompareFiles(templatePath, deployedPath)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if !same {
			t.Error("Expected template and deployed to match")
		}
	})

	t.Run("template differs when deployed content differs", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create local.yaml with variables
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}
		localFile := filepath.Join(localDir, LocalVarsFile)
		if err := os.WriteFile(localFile, []byte("name: NewUser"), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		// Create template file
		templatePath := filepath.Join(tempDir, "config.tmpl")
		if err := os.WriteFile(templatePath, []byte("user = {{.name}}"), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		// Create deployed file with old content
		deployedPath := filepath.Join(tempDir, "deployed_config")
		if err := os.WriteFile(deployedPath, []byte("user = OldUser"), 0644); err != nil {
			t.Fatalf("Failed to write deployed file: %v", err)
		}

		templateProcessor := NewTemplateProcessor(tempDir)
		baseComparator := NewFileComparator()
		comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

		same, err := comparator.CompareFiles(templatePath, deployedPath)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if same {
			t.Error("Expected template and deployed to differ")
		}
	})

	t.Run("non-template uses base comparator", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create two identical regular files
		file1 := filepath.Join(tempDir, "file1.txt")
		file2 := filepath.Join(tempDir, "file2.txt")
		content := []byte("same content")
		if err := os.WriteFile(file1, content, 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(file2, content, 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		templateProcessor := NewTemplateProcessor(tempDir)
		baseComparator := NewFileComparator()
		comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

		same, err := comparator.CompareFiles(file1, file2)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if !same {
			t.Error("Expected identical files to match")
		}
	})

	t.Run("non-template different files differ", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create two different regular files
		file1 := filepath.Join(tempDir, "file1.txt")
		file2 := filepath.Join(tempDir, "file2.txt")
		if err := os.WriteFile(file1, []byte("content A"), 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(file2, []byte("content B"), 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		templateProcessor := NewTemplateProcessor(tempDir)
		baseComparator := NewFileComparator()
		comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

		same, err := comparator.CompareFiles(file1, file2)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if same {
			t.Error("Expected different files to differ")
		}
	})
}

func TestTemplateAwareComparator_ComputeRenderedHash(t *testing.T) {
	tempDir := t.TempDir()

	// Create local.yaml with variables
	localDir := filepath.Join(tempDir, LocalVarsDir)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("Failed to create local dir: %v", err)
	}
	localFile := filepath.Join(localDir, LocalVarsFile)
	if err := os.WriteFile(localFile, []byte("value: test"), 0644); err != nil {
		t.Fatalf("Failed to write local.yaml: %v", err)
	}

	// Create template file
	templatePath := filepath.Join(tempDir, "test.tmpl")
	if err := os.WriteFile(templatePath, []byte("data = {{.value}}"), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	templateProcessor := NewTemplateProcessor(tempDir)
	baseComparator := NewFileComparator()
	comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

	hash, err := comparator.ComputeRenderedHash(templatePath)
	if err != nil {
		t.Fatalf("ComputeRenderedHash failed: %v", err)
	}

	// Hash should be non-empty
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Hash should be consistent
	hash2, err := comparator.ComputeRenderedHash(templatePath)
	if err != nil {
		t.Fatalf("Second ComputeRenderedHash failed: %v", err)
	}
	if hash != hash2 {
		t.Errorf("Hash not consistent: %q vs %q", hash, hash2)
	}
}

func TestTemplateAwareComparator_ComputeFileHash(t *testing.T) {
	tempDir := t.TempDir()

	// Create a regular file
	filePath := filepath.Join(tempDir, "test.txt")
	content := []byte("test content for hashing")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	templateProcessor := NewTemplateProcessor(tempDir)
	baseComparator := NewFileComparator()
	comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

	hash, err := comparator.ComputeFileHash(filePath)
	if err != nil {
		t.Fatalf("ComputeFileHash failed: %v", err)
	}

	// Hash should be non-empty
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Hash should be consistent
	hash2, err := comparator.ComputeFileHash(filePath)
	if err != nil {
		t.Fatalf("Second ComputeFileHash failed: %v", err)
	}
	if hash != hash2 {
		t.Errorf("Hash not consistent: %q vs %q", hash, hash2)
	}
}

func TestTemplateAwareComparator_GetTemplateProcessor(t *testing.T) {
	tempDir := t.TempDir()

	templateProcessor := NewTemplateProcessor(tempDir)
	baseComparator := NewFileComparator()
	comparator := NewTemplateAwareComparator(baseComparator, templateProcessor)

	got := comparator.GetTemplateProcessor()
	if got == nil {
		t.Error("GetTemplateProcessor returned nil")
	}
}
