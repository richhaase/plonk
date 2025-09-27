package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestApply_DryRun_PackagesAndDotfiles(t *testing.T) {
	configDir := t.TempDir()
	homeDir := t.TempDir()

	// Seed a dotfile in config dir (maps to ~/.zshrc)
	if err := os.WriteFile(filepath.Join(configDir, "zshrc"), []byte("echo hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Do not seed lock packages to avoid relying on host package managers in test environment

	cfg := config.LoadWithDefaults(configDir)
	orch := New(
		WithConfig(cfg),
		WithConfigDir(configDir),
		WithHomeDir(homeDir),
		WithDryRun(true),
	)

	res, err := orch.Apply(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !res.DryRun {
		t.Fatalf("expected dryrun")
	}
	if res.Dotfiles == nil {
		t.Fatalf("expected dotfiles result present")
	}
	if res.Dotfiles.Summary.Added == 0 && len(res.Dotfiles.Actions) == 0 {
		t.Fatalf("expected dotfiles would be added")
	}
}
