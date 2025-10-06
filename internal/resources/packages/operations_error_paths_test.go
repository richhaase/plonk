package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

type unavailMgr struct{}

func (u *unavailMgr) IsAvailable(ctx context.Context) (bool, error)              { return false, nil }
func (u *unavailMgr) ListInstalled(ctx context.Context) ([]string, error)        { return nil, nil }
func (u *unavailMgr) Install(ctx context.Context, name string) error             { return nil }
func (u *unavailMgr) Uninstall(ctx context.Context, name string) error           { return nil }
func (u *unavailMgr) IsInstalled(ctx context.Context, name string) (bool, error) { return false, nil }
func (u *unavailMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (u *unavailMgr) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name, Manager: "u"}, nil
}
func (u *unavailMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (u *unavailMgr) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "u"}, nil
}
func (u *unavailMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (u *unavailMgr) Dependencies() []string                           { return nil }

func TestInstallPackagesWith_UnsupportedManager(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	lockSvc := lock.NewYAMLLockService(t.TempDir())
	reg := NewManagerRegistry() // no registrations

	results, err := InstallPackagesWith(context.Background(), cfg, lockSvc, reg, []string{"tool"}, InstallOptions{Manager: "fake"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected failed, got %s", results[0].Status)
	}
}

func TestInstallPackagesWith_ManagerUnavailable(t *testing.T) {
	cfg := &config.Config{}
	lockSvc := lock.NewYAMLLockService(t.TempDir())
	var u unavailMgr
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("off", func() PackageManager { return &u }) })
	reg := NewManagerRegistry()

	results, err := InstallPackagesWith(context.Background(), cfg, lockSvc, reg, []string{"x"}, InstallOptions{Manager: "off"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected failed, got %s", results[0].Status)
	}
}
