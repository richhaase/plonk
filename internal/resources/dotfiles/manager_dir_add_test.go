package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestAddFiles_DirectoryExpansion_DryRun(t *testing.T) {
	home := t.TempDir()
	cfgDir := t.TempDir()
	cfg := &config.Config{}
	m := NewManagerWithConfig(home, cfgDir, cfg)

	// Create a directory with a couple of files
	dir := filepath.Join(home, ".config", "myapp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := m.AddFiles(context.Background(), cfg, []string{dir}, AddOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	// All should be would-add on first pass
	for _, r := range results {
		if r.Status != "would-add" && r.Status != "would-update" { // tolerate update if scanning recognizes pre-existing
			t.Fatalf("unexpected status: %s for %s", r.Status, r.Name)
		}
	}
}
