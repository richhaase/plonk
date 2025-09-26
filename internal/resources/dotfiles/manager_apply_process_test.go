package dotfiles

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

type failingWriter struct{}

func (f *failingWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return errors.New("nope")
}
func (f *failingWriter) CopyFile(ctx context.Context, src, dst string, perm os.FileMode) error {
	return errors.New("copy-failed")
}

func TestProcessDotfileForApply_AddUpdateAndFailure(t *testing.T) {
	home := t.TempDir()
	cfgDir := t.TempDir()
	cfg := &config.Config{}
	m := NewManagerWithConfig(home, cfgDir, cfg)

	// Inject failing writer for later
	resolver := NewPathResolver(home, cfgDir)
	failingOps := NewFileOperationsWithWriter(resolver, &failingWriter{})

	// Create source in config dir (gitconfig -> ~/.gitconfig)
	if err := os.WriteFile(filepath.Join(cfgDir, "gitconfig"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Ensure destination does not exist yet → added
	op, err := m.ProcessDotfileForApply(context.Background(), "gitconfig", "~/.gitconfig", ApplyOptions{DryRun: false})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if op.Status != "added" {
		t.Fatalf("expected added, got %s", op.Status)
	}

	// Second call should update (destination exists)
	op2, err := m.ProcessDotfileForApply(context.Background(), "gitconfig", "~/.gitconfig", ApplyOptions{DryRun: false, Backup: true})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if op2.Status != "updated" {
		t.Fatalf("expected updated, got %s", op2.Status)
	}

	// Now simulate copy failure
	m.fileOperations = failingOps
	op3, err := m.ProcessDotfileForApply(context.Background(), "gitconfig", "~/.gitconfig", ApplyOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if op3.Status != "failed" {
		t.Fatalf("expected failed, got %s", op3.Status)
	}
}
