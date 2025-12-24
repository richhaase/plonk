// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateProcessor_IsTemplate(t *testing.T) {
	tempDir := t.TempDir()
	processor := NewTemplateProcessor(tempDir)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"template file", "gitconfig.tmpl", true},
		{"template in subdir", "config/settings.tmpl", true},
		{"regular file", "bashrc", false},
		{"dotfile", ".zshrc", false},
		{"tmpl in name but not extension", "tmpl-file.txt", false},
		{"empty extension", "file", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.IsTemplate(tt.path)
			if result != tt.expected {
				t.Errorf("IsTemplate(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestTemplateProcessor_GetTemplateName(t *testing.T) {
	tempDir := t.TempDir()
	processor := NewTemplateProcessor(tempDir)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"simple template", "gitconfig.tmpl", "gitconfig"},
		{"template with dot", "config.yaml.tmpl", "config.yaml"},
		{"template in path", "~/.gitconfig.tmpl", "~/.gitconfig"},
		{"non-template unchanged", "bashrc", "bashrc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.GetTemplateName(tt.path)
			if result != tt.expected {
				t.Errorf("GetTemplateName(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestTemplateProcessor_LoadVariables(t *testing.T) {
	t.Run("loads valid yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}

		localFile := filepath.Join(localDir, LocalVarsFile)
		content := `git_email: "test@example.com"
git_name: "Test User"
is_work: true
count: 42`
		if err := os.WriteFile(localFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		vars, err := processor.LoadVariables()
		if err != nil {
			t.Fatalf("LoadVariables failed: %v", err)
		}

		if vars["git_email"] != "test@example.com" {
			t.Errorf("Expected git_email='test@example.com', got %v", vars["git_email"])
		}
		if vars["git_name"] != "Test User" {
			t.Errorf("Expected git_name='Test User', got %v", vars["git_name"])
		}
		if vars["is_work"] != true {
			t.Errorf("Expected is_work=true, got %v", vars["is_work"])
		}
		if vars["count"] != 42 {
			t.Errorf("Expected count=42, got %v", vars["count"])
		}
	})

	t.Run("returns empty map for missing file", func(t *testing.T) {
		tempDir := t.TempDir()
		processor := NewTemplateProcessor(tempDir)

		vars, err := processor.LoadVariables()
		if err != nil {
			t.Errorf("Unexpected error for missing local.yaml: %v", err)
		}
		if vars == nil {
			t.Error("Expected non-nil map for missing local.yaml")
		}
		if len(vars) != 0 {
			t.Errorf("Expected empty map, got %d entries", len(vars))
		}
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}

		localFile := filepath.Join(localDir, LocalVarsFile)
		// Invalid YAML - tabs are not allowed
		content := "key:\n\t- invalid"
		if err := os.WriteFile(localFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		_, err := processor.LoadVariables()
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
}

func TestTemplateProcessor_HasLocalVars(t *testing.T) {
	t.Run("returns true when file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}

		localFile := filepath.Join(localDir, LocalVarsFile)
		if err := os.WriteFile(localFile, []byte("key: value"), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		if !processor.HasLocalVars() {
			t.Error("HasLocalVars() = false, want true")
		}
	})

	t.Run("returns false when file missing", func(t *testing.T) {
		tempDir := t.TempDir()
		processor := NewTemplateProcessor(tempDir)

		if processor.HasLocalVars() {
			t.Error("HasLocalVars() = true, want false")
		}
	})
}

func TestTemplateProcessor_Render(t *testing.T) {
	t.Run("renders simple template", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create template file
		templateContent := `[user]
    email = {{.git_email}}
    name = {{.git_name}}`
		templatePath := filepath.Join(tempDir, "gitconfig.tmpl")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		vars := map[string]interface{}{
			"git_email": "user@example.com",
			"git_name":  "Test User",
		}

		rendered, err := processor.Render(templatePath, vars)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}

		expected := `[user]
    email = user@example.com
    name = Test User`
		if string(rendered) != expected {
			t.Errorf("Render mismatch:\ngot:\n%s\n\nwant:\n%s", string(rendered), expected)
		}
	})

	t.Run("renders template with conditionals", func(t *testing.T) {
		tempDir := t.TempDir()

		templateContent := `{{if .is_work}}[include]
    path = ~/.gitconfig-work{{end}}`
		templatePath := filepath.Join(tempDir, "test.tmpl")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)

		// Test with is_work = true
		vars := map[string]interface{}{"is_work": true}
		rendered, err := processor.Render(templatePath, vars)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if !strings.Contains(string(rendered), "gitconfig-work") {
			t.Error("Expected conditional block to be included")
		}

		// Test with is_work = false
		vars = map[string]interface{}{"is_work": false}
		rendered, err = processor.Render(templatePath, vars)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if strings.Contains(string(rendered), "gitconfig-work") {
			t.Error("Expected conditional block to be excluded")
		}
	})

	t.Run("errors on missing variable", func(t *testing.T) {
		tempDir := t.TempDir()

		templateContent := `email = {{.git_email}}`
		templatePath := filepath.Join(tempDir, "test.tmpl")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		vars := map[string]interface{}{} // Missing git_email

		_, err := processor.Render(templatePath, vars)
		if err == nil {
			t.Error("Expected error for missing variable, got nil")
		}
	})

	t.Run("errors on invalid template syntax", func(t *testing.T) {
		tempDir := t.TempDir()

		templateContent := `{{.unclosed`
		templatePath := filepath.Join(tempDir, "invalid.tmpl")
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		vars := map[string]interface{}{}

		_, err := processor.Render(templatePath, vars)
		if err == nil {
			t.Error("Expected error for invalid template syntax, got nil")
		}
	})
}

func TestTemplateProcessor_RenderToBytes(t *testing.T) {
	tempDir := t.TempDir()

	// Create local.yaml
	localDir := filepath.Join(tempDir, LocalVarsDir)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("Failed to create local dir: %v", err)
	}
	localFile := filepath.Join(localDir, LocalVarsFile)
	if err := os.WriteFile(localFile, []byte("name: TestValue"), 0644); err != nil {
		t.Fatalf("Failed to write local.yaml: %v", err)
	}

	// Create template
	templateContent := `value = {{.name}}`
	templatePath := filepath.Join(tempDir, "test.tmpl")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}

	processor := NewTemplateProcessor(tempDir)
	rendered, err := processor.RenderToBytes(templatePath)
	if err != nil {
		t.Fatalf("RenderToBytes failed: %v", err)
	}

	expected := "value = TestValue"
	if string(rendered) != expected {
		t.Errorf("RenderToBytes = %q, want %q", string(rendered), expected)
	}
}

func TestTemplateProcessor_ValidateTemplate(t *testing.T) {
	t.Run("valid template passes", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create local.yaml with required variable
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}
		localFile := filepath.Join(localDir, LocalVarsFile)
		if err := os.WriteFile(localFile, []byte("name: value"), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		// Create template using that variable
		templatePath := filepath.Join(tempDir, "test.tmpl")
		if err := os.WriteFile(templatePath, []byte("{{.name}}"), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		err := processor.ValidateTemplate(templatePath)
		if err != nil {
			t.Errorf("ValidateTemplate failed for valid template: %v", err)
		}
	})

	t.Run("template with missing variable fails", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create local.yaml without the required variable
		localDir := filepath.Join(tempDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}
		localFile := filepath.Join(localDir, LocalVarsFile)
		if err := os.WriteFile(localFile, []byte("other: value"), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}

		// Create template using missing variable
		templatePath := filepath.Join(tempDir, "test.tmpl")
		if err := os.WriteFile(templatePath, []byte("{{.missing_var}}"), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}

		processor := NewTemplateProcessor(tempDir)
		err := processor.ValidateTemplate(templatePath)
		if err == nil {
			t.Error("Expected validation to fail for missing variable")
		}
	})
}

func TestTemplateProcessor_ListTemplates(t *testing.T) {
	tempDir := t.TempDir()

	// Create some template files
	templates := []string{
		"gitconfig.tmpl",
		"bashrc.tmpl",
		"config/settings.tmpl",
	}

	for _, tmpl := range templates {
		path := filepath.Join(tempDir, tmpl)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", tmpl, err)
		}
	}

	// Create a non-template file
	nonTemplate := filepath.Join(tempDir, "regular.txt")
	if err := os.WriteFile(nonTemplate, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write non-template: %v", err)
	}

	processor := NewTemplateProcessor(tempDir)
	listed, err := processor.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(listed) != len(templates) {
		t.Errorf("ListTemplates returned %d templates, want %d", len(listed), len(templates))
	}

	// Verify each expected template is found
	for _, expected := range templates {
		found := false
		for _, got := range listed {
			if got == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected template %q not found in list", expected)
		}
	}
}

func TestTemplateProcessor_GetLocalVarsPath(t *testing.T) {
	tempDir := t.TempDir()
	processor := NewTemplateProcessor(tempDir)

	expected := filepath.Join(tempDir, LocalVarsDir, LocalVarsFile)
	got := processor.GetLocalVarsPath()

	if got != expected {
		t.Errorf("GetLocalVarsPath() = %q, want %q", got, expected)
	}
}
