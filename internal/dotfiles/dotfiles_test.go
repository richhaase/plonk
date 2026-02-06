// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"strings"
	"testing"
)

func TestDotfileManager_List(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Files["/config/zshrc"] = []byte("# zsh config")
	fs.Files["/config/vimrc"] = []byte("# vim config")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	dotfiles, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(dotfiles) != 2 {
		t.Errorf("List() returned %d dotfiles, want 2", len(dotfiles))
	}

	// Check that files are found
	found := make(map[string]bool)
	for _, d := range dotfiles {
		found[d.Name] = true
	}

	if !found["zshrc"] {
		t.Error("List() missing zshrc")
	}
	if !found["vimrc"] {
		t.Error("List() missing vimrc")
	}
}

func TestDotfileManager_ToTarget(t *testing.T) {
	fs := NewMemoryFS()
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	tests := []struct {
		relPath string
		want    string
	}{
		{"zshrc", "/home/user/.zshrc"},
		{"config/nvim/init.lua", "/home/user/.config/nvim/init.lua"},
		{"bashrc", "/home/user/.bashrc"},
		{"gitconfig.tmpl", "/home/user/.gitconfig"},
		{"config/git/config.tmpl", "/home/user/.config/git/config"},
	}

	for _, tt := range tests {
		got := m.toTarget(tt.relPath)
		if got != tt.want {
			t.Errorf("toTarget(%q) = %q, want %q", tt.relPath, got, tt.want)
		}
	}
}

func TestDotfileManager_ToSource(t *testing.T) {
	fs := NewMemoryFS()
	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	tests := []struct {
		absTarget string
		want      string
	}{
		{"/home/user/.zshrc", "zshrc"},
		{"/home/user/.config/nvim/init.lua", "config/nvim/init.lua"},
		{"/home/user/.bashrc", "bashrc"},
	}

	for _, tt := range tests {
		got := m.toSource(tt.absTarget)
		if got != tt.want {
			t.Errorf("toSource(%q) = %q, want %q", tt.absTarget, got, tt.want)
		}
	}
}

func TestDotfileManager_ShouldIgnore(t *testing.T) {
	fs := NewMemoryFS()
	m := NewDotfileManagerWithFS("/config", "/home/user", []string{"*.bak", ".git"}, fs)

	tests := []struct {
		path string
		want bool
	}{
		{"zshrc", false},
		{"zshrc.bak", true},
		{".git", true},           // ignored by both dot-prefix rule and pattern
		{".gitignore", true},     // ignored by dot-prefix rule (internal file)
		{"config/app.yaml", false}, // nested config files are not ignored
	}

	for _, tt := range tests {
		got := m.shouldIgnore(tt.path)
		if got != tt.want {
			t.Errorf("shouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestDotfileManager_IsDrifted(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("source content")
	fs.Files["/home/user/.zshrc"] = []byte("different content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	d := Dotfile{
		Name:   "zshrc",
		Source: "/config/zshrc",
		Target: "/home/user/.zshrc",
	}

	drifted, err := m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}

	if !drifted {
		t.Error("IsDrifted() = false, want true")
	}

	// Now make them match
	fs.Files["/home/user/.zshrc"] = []byte("source content")
	drifted, err = m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}

	if drifted {
		t.Error("IsDrifted() = true, want false")
	}
}

func TestDotfileManager_Reconcile(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/managed"] = []byte("content")
	fs.Files["/home/user/.managed"] = []byte("content")
	fs.Files["/config/missing"] = []byte("content")
	// no /home/user/.missing
	fs.Files["/config/drifted"] = []byte("source")
	fs.Files["/home/user/.drifted"] = []byte("target")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	statuses, err := m.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	if len(statuses) != 3 {
		t.Errorf("Reconcile() returned %d statuses, want 3", len(statuses))
	}

	// Build map for easier assertions
	stateMap := make(map[string]SyncState)
	for _, s := range statuses {
		stateMap[s.Name] = s.State
	}

	if stateMap["managed"] != SyncStateManaged {
		t.Errorf("managed state = %v, want %v", stateMap["managed"], SyncStateManaged)
	}
	if stateMap["missing"] != SyncStateMissing {
		t.Errorf("missing state = %v, want %v", stateMap["missing"], SyncStateMissing)
	}
	if stateMap["drifted"] != SyncStateDrifted {
		t.Errorf("drifted state = %v, want %v", stateMap["drifted"], SyncStateDrifted)
	}
}

func TestDotfileManager_Add(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/home/user/.zshrc"] = []byte("zsh content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Add("/home/user/.zshrc")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify file was copied to config dir
	content, ok := fs.Files["/config/zshrc"]
	if !ok {
		t.Fatal("Add() did not create /config/zshrc")
	}
	if string(content) != "zsh content" {
		t.Errorf("Add() content = %q, want %q", string(content), "zsh content")
	}
}

func TestDotfileManager_Add_RelativePath(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/home/user/.zshrc"] = []byte("zsh content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	// Relative path should resolve to $HOME
	err := m.Add(".zshrc")
	if err != nil {
		t.Fatalf("Add(.zshrc) error = %v", err)
	}

	if _, ok := fs.Files["/config/zshrc"]; !ok {
		t.Fatal("Add(.zshrc) did not create /config/zshrc")
	}
}

func TestDotfileManager_Add_RejectsConfigDir(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Add("/config/zshrc")
	if err == nil {
		t.Fatal("Add(/config/zshrc) should return error, got nil")
	}

	err = m.ValidateAdd("/config/zshrc")
	if err == nil {
		t.Fatal("ValidateAdd(/config/zshrc) should return error, got nil")
	}
}

func TestDotfileManager_Add_RejectsHomeDirectory(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/home/user/.zshrc"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	// Adding $HOME itself must be rejected
	err := m.Add("/home/user")
	if err == nil {
		t.Fatal("Add($HOME) should return error, got nil")
	}

	// ValidateAdd should also reject it
	err = m.ValidateAdd("/home/user")
	if err == nil {
		t.Fatal("ValidateAdd($HOME) should return error, got nil")
	}
}

func TestDotfileManager_Add_RejectsNonDotPaths(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/home/user/vimrc"] = []byte("vim content")
	fs.Files["/home/user/bin/tool"] = []byte("tool content")
	fs.Dirs["/home/user/bin"] = true

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	// Non-dot file in $HOME should be rejected
	err := m.Add("/home/user/vimrc")
	if err == nil {
		t.Fatal("Add(~/vimrc) should return error for non-dot path, got nil")
	}

	// Non-dot directory in $HOME should be rejected
	err = m.Add("/home/user/bin/tool")
	if err == nil {
		t.Fatal("Add(~/bin/tool) should return error for non-dot path, got nil")
	}

	// Dot-prefixed paths should still work
	fs.Files["/home/user/.vimrc"] = []byte("vim content")
	err = m.Add("/home/user/.vimrc")
	if err != nil {
		t.Fatalf("Add(~/.vimrc) unexpected error: %v", err)
	}

	// Nested dot-prefixed paths should still work
	fs.Files["/home/user/.config/nvim/init.lua"] = []byte("nvim content")
	fs.Dirs["/home/user/.config"] = true
	fs.Dirs["/home/user/.config/nvim"] = true
	err = m.Add("/home/user/.config/nvim/init.lua")
	if err != nil {
		t.Fatalf("Add(~/.config/nvim/init.lua) unexpected error: %v", err)
	}
}

func TestDotfileManager_ValidateAdd_RejectsNonDotPaths(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/home/user/vimrc"] = []byte("content")
	fs.Files["/home/user/.vimrc"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	// Non-dot path should fail validation
	if err := m.ValidateAdd("/home/user/vimrc"); err == nil {
		t.Error("ValidateAdd(~/vimrc) should return error, got nil")
	}

	// Dot path should pass validation
	if err := m.ValidateAdd("/home/user/.vimrc"); err != nil {
		t.Errorf("ValidateAdd(~/.vimrc) unexpected error: %v", err)
	}
}

func TestDotfileManager_Remove(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Files["/config/zshrc"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Remove("zshrc")
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify file was removed
	if _, ok := fs.Files["/config/zshrc"]; ok {
		t.Error("Remove() did not delete /config/zshrc")
	}
}

func TestDotfileManager_Remove_Directory(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/config/dir"] = true
	fs.Files["/config/dir/file"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Remove("dir")
	if err != nil {
		t.Fatalf("Remove(dir) error = %v", err)
	}

	if _, ok := fs.Dirs["/config/dir"]; ok {
		t.Error("Remove(dir) did not delete /config/dir")
	}
	if _, ok := fs.Files["/config/dir/file"]; ok {
		t.Error("Remove(dir) did not delete /config/dir/file")
	}
}

func TestDotfileManager_Remove_RejectsInternalFiles(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Files["/config/plonk.lock"] = []byte("lock content")
	fs.Files["/config/plonk.yaml"] = []byte("yaml content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	// plonk.lock must not be removable
	err := m.Remove("plonk.lock")
	if err == nil {
		t.Fatal("Remove(plonk.lock) should return error, got nil")
	}
	// File must still exist
	if _, ok := fs.Files["/config/plonk.lock"]; !ok {
		t.Error("Remove(plonk.lock) deleted the lock file")
	}

	// plonk.yaml must not be removable
	err = m.Remove("plonk.yaml")
	if err == nil {
		t.Fatal("Remove(plonk.yaml) should return error, got nil")
	}
	// File must still exist
	if _, ok := fs.Files["/config/plonk.yaml"]; !ok {
		t.Error("Remove(plonk.yaml) deleted the config file")
	}
}

func TestDotfileManager_ValidateRemove(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Files["/config/zshrc"] = []byte("content")
	fs.Files["/config/plonk.lock"] = []byte("lock")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	// Valid dotfile should pass
	if err := m.ValidateRemove("zshrc"); err != nil {
		t.Errorf("ValidateRemove(zshrc) unexpected error: %v", err)
	}

	// Path traversal should fail
	if err := m.ValidateRemove("../etc/passwd"); err == nil {
		t.Error("ValidateRemove(../etc/passwd) should return error, got nil")
	}

	// Internal files should fail
	if err := m.ValidateRemove("plonk.lock"); err == nil {
		t.Error("ValidateRemove(plonk.lock) should return error, got nil")
	}

	// Nonexistent dotfile should fail
	if err := m.ValidateRemove("nonexistent"); err == nil {
		t.Error("ValidateRemove(nonexistent) should return error, got nil")
	}
}

func TestDotfileManager_Deploy(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/zshrc"] = []byte("source content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	err := m.Deploy("zshrc")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	// Verify file was deployed
	content, ok := fs.Files["/home/user/.zshrc"]
	if !ok {
		t.Fatal("Deploy() did not create /home/user/.zshrc")
	}
	if string(content) != "source content" {
		t.Errorf("Deploy() content = %q, want %q", string(content), "source content")
	}
}

func TestDotfileManager_ApplyAll(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/managed"] = []byte("content")
	fs.Files["/home/user/.managed"] = []byte("content")
	fs.Files["/config/missing"] = []byte("content")
	fs.Files["/config/drifted"] = []byte("source")
	fs.Files["/home/user/.drifted"] = []byte("target")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	result, err := m.ApplyAll(false)
	if err != nil {
		t.Fatalf("ApplyAll() error = %v", err)
	}

	if len(result.Deployed) != 2 {
		t.Errorf("ApplyAll() deployed %d files, want 2", len(result.Deployed))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("ApplyAll() skipped %d files, want 1", len(result.Skipped))
	}
	if len(result.Failed) != 0 {
		t.Errorf("ApplyAll() failed %d files, want 0", len(result.Failed))
	}

	// Verify missing file was deployed
	if _, ok := fs.Files["/home/user/.missing"]; !ok {
		t.Error("ApplyAll() did not deploy missing file")
	}

	// Verify drifted file was updated
	content := string(fs.Files["/home/user/.drifted"])
	if content != "source" {
		t.Errorf("ApplyAll() drifted content = %q, want %q", content, "source")
	}
}

func TestDotfileManager_ApplyAll_DryRun(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/missing"] = []byte("content")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)

	result, err := m.ApplyAll(true)
	if err != nil {
		t.Fatalf("ApplyAll(dryRun=true) error = %v", err)
	}

	if !result.DryRun {
		t.Error("ApplyAll(dryRun=true) result.DryRun = false")
	}

	if len(result.Deployed) != 1 {
		t.Errorf("ApplyAll(dryRun=true) deployed %d files, want 1", len(result.Deployed))
	}

	// Verify file was NOT actually deployed
	if _, ok := fs.Files["/home/user/.missing"]; ok {
		t.Error("ApplyAll(dryRun=true) actually deployed file")
	}
}

func TestDotfileManager_AddDirectory_IgnoresPatterns(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Dirs["/home/user/.config"] = true
	fs.Dirs["/home/user/.config/nvim"] = true
	fs.Dirs["/home/user/.config/.git"] = true
	fs.Files["/home/user/.config/nvim/init.lua"] = []byte("nvim content")
	fs.Files["/home/user/.config/nvim/init.lua.bak"] = []byte("backup")
	fs.Files["/home/user/.config/.git/ignored"] = []byte("git content")

	m := NewDotfileManagerWithFS("/config", "/home/user", []string{"*.bak"}, fs)

	err := m.Add("/home/user/.config")
	if err != nil {
		t.Fatalf("Add(~/.config) error = %v", err)
	}

	if _, ok := fs.Files["/config/config/nvim/init.lua"]; !ok {
		t.Error("Add(~/.config) did not copy init.lua")
	}
	if _, ok := fs.Files["/config/config/nvim/init.lua.bak"]; ok {
		t.Error("Add(~/.config) copied ignored .bak file")
	}
	if _, ok := fs.Files["/config/config/.git/ignored"]; ok {
		t.Error("Add(~/.config) copied .git directory contents")
	}
}

func TestRenderTemplate(t *testing.T) {
	lookup := func(key string) (string, bool) {
		vars := map[string]string{
			"EMAIL":         "user@example.com",
			"GIT_USER_NAME": "Test User",
		}
		v, ok := vars[key]
		return v, ok
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "single variable",
			input: "email = {{EMAIL}}",
			want:  "email = user@example.com",
		},
		{
			name:  "multiple variables",
			input: "email = {{EMAIL}}\nname = {{GIT_USER_NAME}}",
			want:  "email = user@example.com\nname = Test User",
		},
		{
			name:  "no placeholders",
			input: "just plain text",
			want:  "just plain text",
		},
		{
			name:    "missing variable",
			input:   "email = {{MISSING_VAR}}",
			wantErr: true,
		},
		{
			name:    "multiple missing variables",
			input:   "{{MISSING_ONE}} and {{MISSING_TWO}}",
			wantErr: true,
		},
		{
			name:  "same variable used twice",
			input: "{{EMAIL}} and {{EMAIL}}",
			want:  "user@example.com and user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate([]byte(tt.input), lookup)
			if tt.wantErr {
				if err == nil {
					t.Fatal("renderTemplate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("renderTemplate() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("renderTemplate() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestDotfileManager_Deploy_Template(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("[user]\n    email = {{EMAIL}}")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		if key == "EMAIL" {
			return "user@example.com", true
		}
		return "", false
	}

	err := m.Deploy("gitconfig.tmpl")
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	// Verify rendered content was deployed (not raw template)
	content, ok := fs.Files["/home/user/.gitconfig"]
	if !ok {
		t.Fatal("Deploy() did not create /home/user/.gitconfig")
	}
	want := "[user]\n    email = user@example.com"
	if string(content) != want {
		t.Errorf("Deploy() content = %q, want %q", string(content), want)
	}
}

func TestDotfileManager_Deploy_Template_MissingVar(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("email = {{MISSING_VAR}}")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		return "", false
	}

	err := m.Deploy("gitconfig.tmpl")
	if err == nil {
		t.Fatal("Deploy() expected error for missing variable, got nil")
	}
}

func TestDotfileManager_IsDrifted_Template(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("email = {{EMAIL}}")
	fs.Files["/home/user/.gitconfig"] = []byte("email = user@example.com")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		if key == "EMAIL" {
			return "user@example.com", true
		}
		return "", false
	}

	d := Dotfile{
		Name:   "gitconfig.tmpl",
		Source: "/config/gitconfig.tmpl",
		Target: "/home/user/.gitconfig",
	}

	// Rendered content matches target — not drifted
	drifted, err := m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}
	if drifted {
		t.Error("IsDrifted() = true, want false (rendered content matches)")
	}

	// Change the target — should be drifted
	fs.Files["/home/user/.gitconfig"] = []byte("email = different@example.com")
	drifted, err = m.IsDrifted(d)
	if err != nil {
		t.Fatalf("IsDrifted() error = %v", err)
	}
	if !drifted {
		t.Error("IsDrifted() = false, want true (rendered content differs)")
	}
}

func TestDotfileManager_Diff_Template(t *testing.T) {
	fs := NewMemoryFS()
	fs.Dirs["/config"] = true
	fs.Dirs["/home/user"] = true
	fs.Files["/config/gitconfig.tmpl"] = []byte("email = {{EMAIL}}")
	fs.Files["/home/user/.gitconfig"] = []byte("email = old@example.com")

	m := NewDotfileManagerWithFS("/config", "/home/user", nil, fs)
	m.lookupEnv = func(key string) (string, bool) {
		if key == "EMAIL" {
			return "new@example.com", true
		}
		return "", false
	}

	d := Dotfile{
		Name:   "gitconfig.tmpl",
		Source: "/config/gitconfig.tmpl",
		Target: "/home/user/.gitconfig",
	}

	diff, err := m.Diff(d)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// Diff should show rendered source vs target, not raw template
	if strings.Contains(diff, "{{EMAIL}}") {
		t.Error("Diff() contains raw template placeholder, should contain rendered value")
	}
	if !strings.Contains(diff, "new@example.com") {
		t.Error("Diff() should contain rendered value 'new@example.com'")
	}
}

func TestIsTemplate(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"gitconfig.tmpl", true},
		{"config/git/config.tmpl", true},
		{"zshrc", false},
		{"tmpl", false},
		{"file.tmpl.bak", false},
	}

	for _, tt := range tests {
		if got := isTemplate(tt.name); got != tt.want {
			t.Errorf("isTemplate(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
