// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
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
