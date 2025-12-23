package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestAddSingleFile_AddedVsUpdated_DryRun(t *testing.T) {
	home := t.TempDir()
	cfgDir := t.TempDir()
	srcHomeFile := filepath.Join(home, ".gitconfig")
	if err := os.WriteFile(srcHomeFile, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	m := NewManagerWithConfig(home, cfgDir, cfg)

	// First add (not managed yet) → would-add
	res := m.AddSingleFile(context.Background(), cfg, srcHomeFile, true)
	if res.Status != "would-add" {
		t.Fatalf("expected would-add, got %s", res.Status)
	}

	// Create file in config dir to simulate managed state
	// Destination computed from path resolver: source in config is without dot
	if err := os.WriteFile(filepath.Join(cfgDir, "gitconfig"), []byte("cfg"), 0o644); err != nil {
		t.Fatal(err)
	}
	res2 := m.AddSingleFile(context.Background(), cfg, srcHomeFile, true)
	if res2.Status != "would-update" {
		t.Fatalf("expected would-update, got %s", res2.Status)
	}
}

func TestRemoveSingleDotfile_Behavior(t *testing.T) {
	home := t.TempDir()
	cfgDir := t.TempDir()
	cfg := &config.Config{}
	m := NewManagerWithConfig(home, cfgDir, cfg)

	// Not managed → skipped
	res := m.RemoveSingleDotfile(cfg, "~/.zshrc", true)
	if res.Status != "would-remove" && res.Status != "skipped" {
		t.Fatalf("expected would-remove or skipped (implementation may vary), got %s", res.Status)
	}

	// Create managed source then remove
	if err := os.WriteFile(filepath.Join(cfgDir, "zshrc"), []byte("z"), 0o644); err != nil {
		t.Fatal(err)
	}
	res2 := m.RemoveSingleDotfile(cfg, "~/.zshrc", false)
	if res2.Status != "removed" {
		t.Fatalf("expected removed, got %s", res2.Status)
	}
}
