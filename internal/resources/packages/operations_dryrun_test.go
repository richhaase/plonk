package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

func TestInstall_Uninstall_DryRun_Statuses(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	ls := lock.NewYAMLLockService(t.TempDir())

	// No manager needed for dry-run
	reg := NewManagerRegistry()

	res, err := InstallPackagesWith(context.Background(), cfg, ls, reg, []string{"jq"}, InstallOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "would-add" {
		t.Fatalf("expected would-add, got %+v", res)
	}

	res2, err := UninstallPackagesWith(context.Background(), cfg, ls, reg, []string{"jq"}, UninstallOptions{DryRun: true, Manager: "brew"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res2) != 1 || res2[0].Status != "would-remove" {
		t.Fatalf("expected would-remove, got %+v", res2)
	}
}
