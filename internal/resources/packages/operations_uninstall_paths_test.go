package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

type fakeUninstallMgr struct{ err error }

func (f *fakeUninstallMgr) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fakeUninstallMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeUninstallMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeUninstallMgr) Uninstall(ctx context.Context, name string) error    { return f.err }
func (f *fakeUninstallMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeUninstallMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (f *fakeUninstallMgr) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name}, nil
}
func (f *fakeUninstallMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (f *fakeUninstallMgr) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "u"}, nil
}
func (f *fakeUninstallMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeUninstallMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeUninstallMgr) Dependencies() []string                           { return nil }

// lockSvc that reports package as managed and can fail RemovePackage
type failingRemoveLock struct{ removeErr error }

func (f *failingRemoveLock) Read() (*lock.Lock, error) {
	return &lock.Lock{Version: lock.LockFileVersion}, nil
}
func (f *failingRemoveLock) Write(l *lock.Lock) error { return nil }
func (f *failingRemoveLock) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	return nil
}
func (f *failingRemoveLock) RemovePackage(manager, name string) error { return f.removeErr }
func (f *failingRemoveLock) GetPackages(manager string) ([]lock.ResourceEntry, error) {
	return nil, nil
}
func (f *failingRemoveLock) HasPackage(manager, name string) bool         { return true }
func (f *failingRemoveLock) FindPackage(name string) []lock.ResourceEntry { return nil }

func TestUninstall_Managed_UninstallErrorButRemovedFromLock(t *testing.T) {
	cfg := &config.Config{}
	l := &failingRemoveLock{removeErr: nil}
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &fakeUninstallMgr{err: errors.New("boom")} })
	})
	reg := NewManagerRegistry()
	res, err := UninstallPackagesWith(context.Background(), cfg, l, reg, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" {
		t.Fatalf("expected removed with error, got %+v", res)
	}
}

func TestUninstall_Managed_RemoveFromLockFails(t *testing.T) {
	cfg := &config.Config{}
	l := &failingRemoveLock{removeErr: errors.New("lock fail")}
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("brew", func() PackageManager { return &fakeUninstallMgr{} }) })
	reg := NewManagerRegistry()
	res, err := UninstallPackagesWith(context.Background(), cfg, l, reg, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" || res[0].Error == nil {
		t.Fatalf("expected removed with error due to lock removal, got %+v", res)
	}
}

func TestUninstall_Unmanaged_PassThrough(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	// Use real YAML lock in temp dir (empty means unmanaged)
	l := lock.NewYAMLLockService(t.TempDir())
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("brew", func() PackageManager { return &fakeUninstallMgr{} }) })
	reg := NewManagerRegistry()
	res, err := UninstallPackagesWith(context.Background(), cfg, l, reg, []string{"jq"}, UninstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" {
		t.Fatalf("expected removed pass-through, got %+v", res)
	}
}
