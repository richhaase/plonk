package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// fake managers for upgrade execution tests
type fakeMgrUpgradeChanging struct{ ver string }

func (f *fakeMgrUpgradeChanging) IsAvailable(ctx context.Context) (bool, error) {
	if f.ver == "" {
		f.ver = "1.0"
	}
	return true, nil
}
func (f *fakeMgrUpgradeChanging) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (f *fakeMgrUpgradeChanging) Install(ctx context.Context, name string) error   { return nil }
func (f *fakeMgrUpgradeChanging) Uninstall(ctx context.Context, name string) error { return nil }
func (f *fakeMgrUpgradeChanging) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeMgrUpgradeChanging) InstalledVersion(ctx context.Context, name string) (string, error) {
	if f.ver == "" {
		f.ver = "1.0"
	}
	return f.ver, nil
}
func (f *fakeMgrUpgradeChanging) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "brew", Installed: true, Version: f.ver}, nil
}
func (f *fakeMgrUpgradeChanging) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}
func (f *fakeMgrUpgradeChanging) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "brew"}, nil
}
func (f *fakeMgrUpgradeChanging) SelfInstall(ctx context.Context) error { return nil }
func (f *fakeMgrUpgradeChanging) Upgrade(ctx context.Context, pkgs []string) error {
	f.ver = "1.1"
	return nil
}
func (f *fakeMgrUpgradeChanging) Dependencies() []string { return nil }

type fakeMgrUpgradeNoChange struct{}

func (f *fakeMgrUpgradeNoChange) IsAvailable(ctx context.Context) (bool, error) { return true, nil }
func (f *fakeMgrUpgradeNoChange) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (f *fakeMgrUpgradeNoChange) Install(ctx context.Context, name string) error   { return nil }
func (f *fakeMgrUpgradeNoChange) Uninstall(ctx context.Context, name string) error { return nil }
func (f *fakeMgrUpgradeNoChange) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeMgrUpgradeNoChange) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0", nil
}
func (f *fakeMgrUpgradeNoChange) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "npm", Installed: true, Version: "1.0"}, nil
}
func (f *fakeMgrUpgradeNoChange) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}
func (f *fakeMgrUpgradeNoChange) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "npm"}, nil
}
func (f *fakeMgrUpgradeNoChange) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeMgrUpgradeNoChange) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeMgrUpgradeNoChange) Dependencies() []string                           { return nil }

func seedUpgradeLock(t *testing.T, dir string) {
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})
	_ = svc.AddPackage("npm", "typescript", "1.0.0", map[string]interface{}{"manager": "npm", "name": "typescript", "version": "1.0.0"})
	_ = svc.AddPackage("foo", "bar", "1.0.0", map[string]interface{}{"manager": "foo", "name": "bar", "version": "1.0.0"})
}

func TestUpgrade_ManagerUnavailableAndMixedResults(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		seedUpgradeLock(env.T, env.ConfigDir)
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {
			r.Register("brew", func() packages.PackageManager { return &fakeMgrUpgradeChanging{} })
			r.Register("npm", func() packages.PackageManager { return &fakeMgrUpgradeNoChange{} })
			// intentionally do NOT register manager "foo" to trigger GetManager error path
		})
	})
	if err == nil {
		t.Fatalf("expected non-nil error due to failures, got nil. Output:\n%s", out)
	}
	// Ensure summary line present and mentions failed/skipped/upgraded in some combination
	if !containsAll(out, []string{"Summary:", "Failed:", "Skipped:"}) {
		t.Fatalf("expected summary with failed and skipped, got:\n%s", out)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
