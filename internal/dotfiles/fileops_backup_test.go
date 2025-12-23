package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileOperations_BackupAndCopy(t *testing.T) {
	home := t.TempDir()
	cfgDir := t.TempDir()
	resolver := NewPathResolver(home, cfgDir)
	ops := NewFileOperations(resolver)

	// Create source in config dir
	srcRel := "gitconfig"
	srcPath := filepath.Join(cfgDir, srcRel)
	if err := os.WriteFile(srcPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create existing destination to force backup
	dest := "~/.gitconfig"
	destPath, err := resolver.GetDestinationPath(dest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	err = ops.CopyFile(context.Background(), srcRel, dest, CopyOptions{CreateBackup: true, OverwriteExisting: true, BackupSuffix: ".backup"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	// Destination exists with new content
	if b, _ := os.ReadFile(destPath); string(b) != "x" {
		t.Fatalf("dest not updated")
	}

	// Backup file with timestamp should exist
	dir := filepath.Dir(destPath)
	entries, _ := os.ReadDir(dir)
	foundBackup := false
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) >= len(filepath.Base(destPath)+".backup.") && e.Name()[:len(filepath.Base(destPath)+".backup.")] == filepath.Base(destPath)+".backup." {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Fatalf("expected a backup file present")
	}
}
