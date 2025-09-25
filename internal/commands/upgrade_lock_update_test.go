package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

type fakeMgrVersioning struct{ ver string }

func (f *fakeMgrVersioning) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fakeMgrVersioning) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeMgrVersioning) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeMgrVersioning) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeMgrVersioning) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeMgrVersioning) InstalledVersion(ctx context.Context, name string) (string, error) {
	if f.ver == "" {
		f.ver = "1.0.0"
	}
	return f.ver, nil
}
func (f *fakeMgrVersioning) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "brew", Installed: true, Version: f.ver}, nil
}
func (f *fakeMgrVersioning) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (f *fakeMgrVersioning) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "brew"}, nil
}
func (f *fakeMgrVersioning) SelfInstall(ctx context.Context) error { return nil }
func (f *fakeMgrVersioning) Upgrade(ctx context.Context, pkgs []string) error {
	f.ver = "1.1.0"
	return nil
}
func (f *fakeMgrVersioning) Dependencies() []string { return nil }

func TestUpgrade_UpdatesLockFileVersion(t *testing.T) {
	dir := t.TempDir()
	// Seed lock with brew:jq@1.0.0
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})

	// Register fake brew manager
	var mgr fakeMgrVersioning
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &mgr })
	})

	spec := upgradeSpec{ManagerTargets: map[string][]string{"brew": {"jq"}}}
	cfg := &config.Config{}
	reg := packages.NewManagerRegistry()
	res, err := Upgrade(context.Background(), spec, cfg, svc, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary.Upgraded == 0 {
		t.Fatalf("expected at least one upgraded package, got %+v", res.Summary)
	}

	// Verify lock updated
	lf, err := svc.Read()
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	found := false
	for _, r := range lf.Resources {
		if r.Type == "package" && r.ID == "brew:jq" {
			if v, ok := r.Metadata["version"].(string); !ok || v != "1.1.0" {
				t.Fatalf("expected version 1.1.0, got %v", r.Metadata["version"])
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("brew:jq not found in lock after upgrade")
	}
}
