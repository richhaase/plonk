// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// templateTestEnv holds test environment for template file operations tests
type templateTestEnv struct {
	homeDir   string
	configDir string
	ops       *TemplateFileOperations
	t         *testing.T
}

// setupTemplateTestEnv creates a test environment with optional local.yaml content
func setupTemplateTestEnv(t *testing.T, localYamlContent string) *templateTestEnv {
	t.Helper()
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	if localYamlContent != "" {
		localDir := filepath.Join(configDir, LocalVarsDir)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			t.Fatalf("Failed to create local dir: %v", err)
		}
		localFile := filepath.Join(localDir, LocalVarsFile)
		if err := os.WriteFile(localFile, []byte(localYamlContent), 0644); err != nil {
			t.Fatalf("Failed to write local.yaml: %v", err)
		}
	}

	pathResolver := NewPathResolver(homeDir, configDir)
	templateProcessor := NewTemplateProcessor(configDir)
	ops := NewTemplateFileOperations(pathResolver, templateProcessor)

	return &templateTestEnv{
		homeDir:   homeDir,
		configDir: configDir,
		ops:       ops,
		t:         t,
	}
}

func (e *templateTestEnv) writeTemplate(name, content string) {
	e.t.Helper()
	path := filepath.Join(e.configDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		e.t.Fatalf("Failed to write template: %v", err)
	}
}

func (e *templateTestEnv) writeSource(name, content string) {
	e.t.Helper()
	path := filepath.Join(e.configDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		e.t.Fatalf("Failed to write source: %v", err)
	}
}

func (e *templateTestEnv) writeDest(name, content string) {
	e.t.Helper()
	path := filepath.Join(e.homeDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		e.t.Fatalf("Failed to write destination: %v", err)
	}
}

func (e *templateTestEnv) readDest(name string) string {
	e.t.Helper()
	path := filepath.Join(e.homeDir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		e.t.Fatalf("Failed to read destination: %v", err)
	}
	return string(content)
}

func TestTemplateFileOperations_CopyFile_RendersTemplate(t *testing.T) {
	env := setupTemplateTestEnv(t, "email: test@example.com")
	env.writeTemplate("gitconfig.tmpl", `[user]
    email = {{.email}}`)

	ctx := context.Background()
	options := CopyOptions{OverwriteExisting: true}

	err := env.ops.CopyFile(ctx, "gitconfig.tmpl", "~/.gitconfig", options)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	expected := `[user]
    email = test@example.com`
	if got := env.readDest(".gitconfig"); got != expected {
		t.Errorf("Rendered content mismatch:\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestTemplateFileOperations_CopyFile_CopiesRegularFile(t *testing.T) {
	env := setupTemplateTestEnv(t, "")
	sourceContent := "export PATH=$PATH:/custom/bin"
	env.writeSource("bashrc", sourceContent)

	ctx := context.Background()
	options := CopyOptions{OverwriteExisting: true}

	err := env.ops.CopyFile(ctx, "bashrc", "~/.bashrc", options)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	if got := env.readDest(".bashrc"); got != sourceContent {
		t.Errorf("Content mismatch: got %q, want %q", got, sourceContent)
	}
}

func TestTemplateFileOperations_CopyFile_ErrorsOnMissingVariable(t *testing.T) {
	env := setupTemplateTestEnv(t, "other_var: value")
	env.writeTemplate("config.tmpl", "value = {{.missing_var}}")

	ctx := context.Background()
	options := CopyOptions{OverwriteExisting: true}

	err := env.ops.CopyFile(ctx, "config.tmpl", "~/.config", options)
	if err == nil {
		t.Error("Expected error for missing variable, got nil")
	}
}

func TestTemplateFileOperations_CopyFile_CreatesBackup(t *testing.T) {
	env := setupTemplateTestEnv(t, "value: new")
	env.writeTemplate("config.tmpl", "v = {{.value}}")
	env.writeDest(".config", "old content")

	ctx := context.Background()
	options := CopyOptions{
		OverwriteExisting: true,
		CreateBackup:      true,
		BackupSuffix:      ".bak",
	}

	err := env.ops.CopyFile(ctx, "config.tmpl", "~/.config", options)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	if got := env.readDest(".config"); got != "v = new" {
		t.Errorf("New content mismatch: got %q", got)
	}

	// Verify backup exists
	entries, err := os.ReadDir(env.homeDir)
	if err != nil {
		t.Fatalf("Failed to read home dir: %v", err)
	}
	backupFound := false
	for _, e := range entries {
		if e.Name() != ".config" && len(e.Name()) > len(".config.bak") {
			backupFound = true
			break
		}
	}
	if !backupFound {
		t.Error("Expected backup file to be created")
	}
}

func TestTemplateFileOperations_CopyFile_RespectsContextCancellation(t *testing.T) {
	env := setupTemplateTestEnv(t, "value: test")
	env.writeTemplate("config.tmpl", "{{.value}}")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	options := CopyOptions{OverwriteExisting: true}
	err := env.ops.CopyFile(ctx, "config.tmpl", "~/.config", options)
	if err == nil {
		t.Error("Expected error for canceled context")
	}
}

func TestTemplateFileOperations_GetTemplateProcessor(t *testing.T) {
	tempDir := t.TempDir()

	pathResolver := NewPathResolver(tempDir, tempDir)
	templateProcessor := NewTemplateProcessor(tempDir)
	ops := NewTemplateFileOperations(pathResolver, templateProcessor)

	got := ops.GetTemplateProcessor()
	if got == nil {
		t.Error("GetTemplateProcessor returned nil")
	}
}

func TestTemplateFileOperations_GetBaseOperations(t *testing.T) {
	tempDir := t.TempDir()

	pathResolver := NewPathResolver(tempDir, tempDir)
	templateProcessor := NewTemplateProcessor(tempDir)
	ops := NewTemplateFileOperations(pathResolver, templateProcessor)

	got := ops.GetBaseOperations()
	if got == nil {
		t.Error("GetBaseOperations returned nil")
	}
}
