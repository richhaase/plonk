package clone

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// fakeInstallMgr is a minimal PackageManager that reports unavailable and records SelfInstall calls
type fakeInstallMgr struct{ installs int }

func (f *fakeInstallMgr) IsAvailable(ctx context.Context) (bool, error)       { return false, nil }
func (f *fakeInstallMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeInstallMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeInstallMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeInstallMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (f *fakeInstallMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (f *fakeInstallMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "fake"}, nil
}
func (f *fakeInstallMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (f *fakeInstallMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "fake"}, nil
}
func (f *fakeInstallMgr) SelfInstall(ctx context.Context) error            { f.installs++; return nil }
func (f *fakeInstallMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeInstallMgr) Dependencies() []string                           { return nil }

func TestSetupFromClonedRepo_NoManagers_NoApply(t *testing.T) {
	dir := t.TempDir()
	// Create minimal plonk.yaml so hasConfig=true
	if err := createDefaultConfig(dir); err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	// No lock file => no detected managers
	if err := SetupFromClonedRepo(context.Background(), dir, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupFromClonedRepo_InstallsDetectedManagers(t *testing.T) {
	dir := t.TempDir()
	if err := createDefaultConfig(dir); err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	// Seed lock with two managers
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})
	_ = svc.AddPackage("npm", "typescript", "1.0.0", map[string]interface{}{"manager": "npm", "name": "typescript", "version": "1.0.0"})

	// Register fake managers that are not available so they get installed
	var brewMgr, npmMgr fakeInstallMgr
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &brewMgr })
		r.Register("npm", func() packages.PackageManager { return &npmMgr })
	})

	// Run setup without apply to isolate manager install path
	if err := SetupFromClonedRepo(context.Background(), dir, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if brewMgr.installs != 1 {
		t.Fatalf("expected brew SelfInstall to be called once, got %d", brewMgr.installs)
	}
	if npmMgr.installs != 1 {
		t.Fatalf("expected npm SelfInstall to be called once, got %d", npmMgr.installs)
	}

	// Sanity: lock still present
	if _, err := svc.Read(); err != nil {
		t.Fatalf("failed reading lock: %v", err)
	}
	_ = filepath.Join // silence import linters
}
