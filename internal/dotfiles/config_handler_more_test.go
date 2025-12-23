package dotfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestGetConfiguredDotfiles_RespectsIgnorePatterns(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(t.TempDir())
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create two files in config dir layout
	// plonk stores without leading dot; e.g., zshrc maps to ~/.zshrc
	if err := os.WriteFile(filepath.Join(cfgDir, "keepme"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "ignoreme"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{IgnorePatterns: []string{"ignoreme"}}
	resolver := NewPathResolver(home, cfgDir)
	validator := NewPathValidator(home, cfgDir)
	scanner := NewDirectoryScanner(home, cfgDir, validator, resolver)
	handler := NewConfigHandlerWithConfig(home, cfgDir, cfg, resolver, scanner, NewFileComparator())

	items, err := handler.GetConfiguredDotfiles()
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	foundKeep, foundIgnore := false, false
	for _, it := range items {
		if it.Source == "keepme" {
			foundKeep = true
		}
		if it.Source == "ignoreme" {
			foundIgnore = true
		}
	}
	if !foundKeep {
		t.Fatalf("expected keepme present")
	}
	if foundIgnore {
		t.Fatalf("expected ignoreme filtered out")
	}
}
