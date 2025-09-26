package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	pkgs "github.com/richhaase/plonk/internal/resources/packages"
)

type listOnlyMgr struct{}

func (l *listOnlyMgr) IsAvailable(ctx context.Context) (bool, error)              { return true, nil }
func (l *listOnlyMgr) ListInstalled(ctx context.Context) ([]string, error)        { return []string{}, nil }
func (l *listOnlyMgr) Install(ctx context.Context, name string) error             { return nil }
func (l *listOnlyMgr) Uninstall(ctx context.Context, name string) error           { return nil }
func (l *listOnlyMgr) IsInstalled(ctx context.Context, name string) (bool, error) { return false, nil }
func (l *listOnlyMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (l *listOnlyMgr) Info(ctx context.Context, name string) (*pkgs.PackageInfo, error) {
	return &pkgs.PackageInfo{Name: name}, nil
}
func (l *listOnlyMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (l *listOnlyMgr) CheckHealth(ctx context.Context) (*pkgs.HealthCheck, error) {
	return &pkgs.HealthCheck{Name: "x"}, nil
}
func (l *listOnlyMgr) SelfInstall(ctx context.Context) error         { return nil }
func (l *listOnlyMgr) Upgrade(ctx context.Context, p []string) error { return nil }
func (l *listOnlyMgr) Dependencies() []string                        { return nil }

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
	// Register fake manager
	pkgs.WithTemporaryRegistry(t, func(r *pkgs.ManagerRegistry) {
		r.Register("brew", func() pkgs.PackageManager { return &listOnlyMgr{} })
	})

	results, err := ReconcileAllWithConfig(context.Background(), home, cfgDir, config.LoadWithDefaults(cfgDir))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
}
