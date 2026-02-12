// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a temp dir with git init and returns the path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run(t, dir, "git", "init", "-b", "main")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "initial")

	return dir
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestIsRepo(t *testing.T) {
	dir := initTestRepo(t)
	client := New(dir)

	if !client.IsRepo() {
		t.Error("expected IsRepo to return true for a git repo")
	}
}

func TestIsRepoFalse(t *testing.T) {
	dir := t.TempDir()
	client := New(dir)

	if client.IsRepo() {
		t.Error("expected IsRepo to return false for a plain directory")
	}
}

func TestHasRemote(t *testing.T) {
	dir := initTestRepo(t)

	// Add a bare remote
	remoteDir := t.TempDir()
	run(t, remoteDir, "git", "init", "--bare")
	run(t, dir, "git", "remote", "add", "origin", remoteDir)

	client := New(dir)
	if !client.HasRemote() {
		t.Error("expected HasRemote to return true")
	}
}

func TestHasRemoteFalse(t *testing.T) {
	dir := initTestRepo(t)
	client := New(dir)

	if client.HasRemote() {
		t.Error("expected HasRemote to return false")
	}
}

func TestIsDirtyClean(t *testing.T) {
	dir := initTestRepo(t)
	client := New(dir)

	dirty, err := client.IsDirty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Error("expected clean repo to not be dirty")
	}
}

func TestIsDirty(t *testing.T) {
	dir := initTestRepo(t)
	client := New(dir)

	// Create an untracked file
	if err := os.WriteFile(filepath.Join(dir, "newfile"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dirty, err := client.IsDirty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dirty {
		t.Error("expected dirty repo after creating a file")
	}
}

func TestCommit(t *testing.T) {
	dir := initTestRepo(t)
	client := New(dir)

	// Create a file
	if err := os.WriteFile(filepath.Join(dir, "testfile"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Commit
	if err := client.Commit("test commit message"); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	// Should be clean now
	dirty, err := client.IsDirty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Error("expected clean repo after commit")
	}

	// Verify commit message
	cmd := exec.Command("git", "-C", dir, "log", "--oneline", "-1")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if !strings.Contains(string(out), "test commit message") {
		t.Errorf("expected commit message in log, got: %s", out)
	}
}

func TestCommitNoop(t *testing.T) {
	dir := initTestRepo(t)
	client := New(dir)

	// Commit on clean repo should be a no-op
	if err := client.Commit("should not appear"); err != nil {
		t.Fatalf("commit on clean repo should not error: %v", err)
	}
}

func TestPushPull(t *testing.T) {
	// Set up repo with bare remote
	dir := initTestRepo(t)
	remoteDir := t.TempDir()
	run(t, remoteDir, "git", "init", "--bare")
	run(t, dir, "git", "remote", "add", "origin", remoteDir)
	run(t, dir, "git", "push", "-u", "origin", "main")

	client := New(dir)

	// Create a file and commit
	if err := os.WriteFile(filepath.Join(dir, "pushed"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := client.Commit("push test"); err != nil {
		t.Fatal(err)
	}

	// Push
	if err := client.Push(); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Clone into a second dir and verify the file arrived
	cloneDir := t.TempDir()
	run(t, cloneDir, "git", "clone", remoteDir, ".")

	data, err := os.ReadFile(filepath.Join(cloneDir, "pushed"))
	if err != nil {
		t.Fatalf("pushed file not found in clone: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("unexpected file content: %s", data)
	}

	// Now create a commit in the clone and push it back
	if err := os.WriteFile(filepath.Join(cloneDir, "pulled"), []byte("remote"), 0644); err != nil {
		t.Fatal(err)
	}
	run(t, cloneDir, "git", "add", "-A")
	run(t, cloneDir, "git", "config", "user.email", "test@test.com")
	run(t, cloneDir, "git", "config", "user.name", "Test")
	run(t, cloneDir, "git", "commit", "-m", "remote commit")
	run(t, cloneDir, "git", "push")

	// Pull from original
	if err := client.Pull(); err != nil {
		t.Fatalf("pull failed: %v", err)
	}

	// Verify pulled file
	data, err = os.ReadFile(filepath.Join(dir, "pulled"))
	if err != nil {
		t.Fatalf("pulled file not found: %v", err)
	}
	if string(data) != "remote" {
		t.Errorf("unexpected pulled content: %s", data)
	}
}

func TestCommitMessage(t *testing.T) {
	tests := []struct {
		command string
		args    []string
		want    string
	}{
		{"add", nil, "plonk: add"},
		{"track", []string{"brew:ripgrep"}, "plonk: track brew:ripgrep"},
		{"add", []string{".zshrc", ".vimrc"}, "plonk: add .zshrc .vimrc"},
		{"rm", []string{"a", "b", "c", "d", "e", "f"}, "plonk: rm a b c d e (+1 more)"},
		{"push", nil, "plonk: push"},
		{"config edit", nil, "plonk: config edit"},
	}

	for _, tt := range tests {
		got := CommitMessage(tt.command, tt.args)
		if got != tt.want {
			t.Errorf("CommitMessage(%q, %v) = %q, want %q", tt.command, tt.args, got, tt.want)
		}
	}
}
