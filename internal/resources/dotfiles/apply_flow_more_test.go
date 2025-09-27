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
