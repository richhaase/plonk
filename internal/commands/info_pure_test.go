package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

type fakeInfoMgr struct {
	name      string
	installed bool
}

func (f *fakeInfoMgr) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fakeInfoMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeInfoMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeInfoMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeInfoMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return f.installed, nil
}
func (f *fakeInfoMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (f *fakeInfoMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: f.name, Installed: f.installed, Version: "1.0.0"}, nil
}
func (f *fakeInfoMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (f *fakeInfoMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: f.name}, nil
}
func (f *fakeInfoMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeInfoMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeInfoMgr) Dependencies() []string                           { return nil }

func TestInfo_InvalidManagerPrefix(t *testing.T) {
	if _, err := Info(context.Background(), "badmgr:pkg"); err == nil {
		t.Fatalf("expected error for invalid manager prefix")
	}
}

func TestInfo_ManagedFromLock(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PLONK_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("PLONK_DIR") })
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})

	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &fakeInfoMgr{name: "brew", installed: true} })
	})

	res, err := Info(context.Background(), "jq")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.Status != "managed" {
		t.Fatalf("expected managed, got %s", res.Status)
	}
}

func TestInfo_InstalledNotManaged(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PLONK_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("PLONK_DIR") })
	// No lock entry

	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &fakeInfoMgr{name: "brew", installed: true} })
	})

	res, err := Info(context.Background(), "jq")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.Status != "installed" {
		t.Fatalf("expected installed, got %s", res.Status)
	}
}

func TestInfo_AvailableNotInstalled(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PLONK_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("PLONK_DIR") })

	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &fakeInfoMgr{name: "brew", installed: false} })
	})

	res, err := Info(context.Background(), "jq")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.Status != "available" {
		t.Fatalf("expected available, got %s", res.Status)
	}
	_ = filepath.Join // silence import tools
}
