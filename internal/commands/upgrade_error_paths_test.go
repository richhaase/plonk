package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

type errUpgradeMgr struct{}

func (e *errUpgradeMgr) IsAvailable(ctx context.Context) (bool, error)              { return true, nil }
func (e *errUpgradeMgr) ListInstalled(ctx context.Context) ([]string, error)        { return nil, nil }
func (e *errUpgradeMgr) Install(ctx context.Context, name string) error             { return nil }
func (e *errUpgradeMgr) Uninstall(ctx context.Context, name string) error           { return nil }
func (e *errUpgradeMgr) IsInstalled(ctx context.Context, name string) (bool, error) { return true, nil }
func (e *errUpgradeMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (e *errUpgradeMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Installed: true, Version: "1.0.0"}, nil
}
func (e *errUpgradeMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (e *errUpgradeMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "x"}, nil
}
func (e *errUpgradeMgr) SelfInstall(ctx context.Context) error { return nil }
func (e *errUpgradeMgr) Upgrade(ctx context.Context, pkgs []string) error {
	return fmt.Errorf("upgrade failed")
}
func (e *errUpgradeMgr) Dependencies() []string { return nil }

func TestUpgrade_ManagerErrorCountsFailed(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "a", "1.0.0", map[string]interface{}{"manager": "brew", "name": "a", "version": "1.0.0"})

	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &errUpgradeMgr{} })
	})

	spec := upgradeSpec{ManagerTargets: map[string][]string{"brew": {"a"}}}
	res2, err := Upgrade(context.Background(), spec, &config.Config{}, svc, packages.NewManagerRegistry())
	if err != nil {
		t.Fatalf("unexpected error from Upgrade: %v", err)
	}
	if res2.Summary.Failed == 0 {
		t.Fatalf("expected failed count > 0")
	}
}
