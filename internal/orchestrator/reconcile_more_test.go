package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

func TestReconcileAllWithConfig_Simple(t *testing.T) {
	home := t.TempDir()
	cfgDir := t.TempDir()
	// Seed config dotfile
	if err := os.WriteFile(filepath.Join(cfgDir, "zshrc"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Seed lock with a desired package
	svc := lock.NewYAMLLockService(cfgDir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})
	// Rely on default managers; no real commands are executed during reconciliation

	results, err := ReconcileAllWithConfig(context.Background(), home, cfgDir, config.LoadWithDefaults(cfgDir))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
}
