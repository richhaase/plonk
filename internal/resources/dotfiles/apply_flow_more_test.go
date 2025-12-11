package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestDotfilesApply_DeploysMissing(t *testing.T) {
	cfgDir := t.TempDir()
	home := t.TempDir()
	// seed a managed dotfile in config dir (maps to ~/.zshrc)
	if err := os.WriteFile(filepath.Join(cfgDir, "zshrc"), []byte("export X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Apply(context.Background(), cfgDir, home, config.LoadWithDefaults(cfgDir), false)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.Summary.Added == 0 {
		t.Fatalf("expected added > 0, got %+v", res)
	}
}

func TestApplySelective_FiltersCorrectly(t *testing.T) {
	cfgDir := t.TempDir()
	home := t.TempDir()

	// Seed two managed dotfiles in config dir
	if err := os.WriteFile(filepath.Join(cfgDir, "zshrc"), []byte("export X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "vimrc"), []byte("set nocompatible\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create filter set for only zshrc
	filter := map[string]bool{
		filepath.Join(home, ".zshrc"): true,
	}

	opts := ApplyFilterOptions{
		DryRun: false,
		Filter: filter,
	}

	res, err := ApplySelective(context.Background(), cfgDir, home, config.LoadWithDefaults(cfgDir), opts)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	// Should only have added one file (zshrc), not two
	if res.Summary.Added != 1 {
		t.Errorf("expected 1 file added, got %d", res.Summary.Added)
	}

	// zshrc should exist
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); os.IsNotExist(err) {
		t.Error("expected .zshrc to be created")
	}

	// vimrc should NOT exist (it was filtered out)
	if _, err := os.Stat(filepath.Join(home, ".vimrc")); !os.IsNotExist(err) {
		t.Error("expected .vimrc to NOT be created (filtered out)")
	}
}

func TestApplySelective_EmptyFilterAppliesAll(t *testing.T) {
	cfgDir := t.TempDir()
	home := t.TempDir()

	// Seed two managed dotfiles
	if err := os.WriteFile(filepath.Join(cfgDir, "zshrc"), []byte("export X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "vimrc"), []byte("set nocompatible\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Empty filter should apply all
	opts := ApplyFilterOptions{
		DryRun: false,
		Filter: nil,
	}

	res, err := ApplySelective(context.Background(), cfgDir, home, config.LoadWithDefaults(cfgDir), opts)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	// Both files should have been added
	if res.Summary.Added != 2 {
		t.Errorf("expected 2 files added, got %d", res.Summary.Added)
	}

	// Both files should exist
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); os.IsNotExist(err) {
		t.Error("expected .zshrc to be created")
	}
	if _, err := os.Stat(filepath.Join(home, ".vimrc")); os.IsNotExist(err) {
		t.Error("expected .vimrc to be created")
	}
}

func TestApplySelective_DryRunDoesNotCreate(t *testing.T) {
	cfgDir := t.TempDir()
	home := t.TempDir()

	// Seed a managed dotfile
	if err := os.WriteFile(filepath.Join(cfgDir, "zshrc"), []byte("export X=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	filter := map[string]bool{
		filepath.Join(home, ".zshrc"): true,
	}

	opts := ApplyFilterOptions{
		DryRun: true,
		Filter: filter,
	}

	res, err := ApplySelective(context.Background(), cfgDir, home, config.LoadWithDefaults(cfgDir), opts)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	// Should report file as added
	if res.Summary.Added != 1 {
		t.Errorf("expected 1 file would be added, got %d", res.Summary.Added)
	}

	// But file should NOT actually exist (dry run)
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); !os.IsNotExist(err) {
		t.Error("expected .zshrc to NOT be created in dry-run mode")
	}
}
