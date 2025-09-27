package commands

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// Fake manager that reports available and simulates version change on Upgrade.
type fakeUpgradeMgr struct{ ver string }

func (f *fakeUpgradeMgr) IsAvailable(ctx context.Context) (bool, error) {
	if f.ver == "" {
		f.ver = "1.0"
	}
	return true, nil
}
func (f *fakeUpgradeMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeUpgradeMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeUpgradeMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeUpgradeMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeUpgradeMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	if f.ver == "" {
		f.ver = "1.0"
	}
	return f.ver, nil
}
func (f *fakeUpgradeMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "brew", Installed: true, Version: f.ver}, nil
}
func (f *fakeUpgradeMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (f *fakeUpgradeMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "brew"}, nil
}
func (f *fakeUpgradeMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeUpgradeMgr) Upgrade(ctx context.Context, pkgs []string) error { f.ver = "2.0"; return nil }
func (f *fakeUpgradeMgr) Dependencies() []string                           { return nil }

func TestUpgrade_UpdatesLockMetadata(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		// Seed one brew package with starting version
		svc := lock.NewYAMLLockService(env.ConfigDir)
		if e := svc.AddPackage("brew", "jq", "1.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0"}); e != nil {
			t.Fatalf("seed lock: %v", e)
		}
		// Registry: only brew available
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {
			r.Register("brew", func() packages.PackageManager { return &fakeUpgradeMgr{} })
		})
	})
	if err != nil {
		t.Fatalf("upgrade error: %v\n%s", err, out)
	}

	// Read lock and verify version updated and installed_at present
	svc := lock.NewYAMLLockService(os.Getenv("PLONK_DIR"))
	// Note: RunCLI sets PLONK_DIR to env.ConfigDir
	lk, e := svc.Read()
	if e != nil {
		t.Fatalf("read lock: %v", e)
	}
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}
	md := lk.Resources[0].Metadata
	if md["version"] != "2.0" {
		t.Fatalf("expected version 2.0, got %#v", md["version"])
	}
	if lk.Resources[0].InstalledAt == "" {
		t.Fatalf("expected InstalledAt to be set")
	}
}

func TestUpgrade_InstalledAtUpdated(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade"}, func(env CLITestEnv) {
		// Seed one brew package, then force a known InstalledAt timestamp
		svc := lock.NewYAMLLockService(env.ConfigDir)
		if e := svc.AddPackage("brew", "jq", "1.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0"}); e != nil {
			t.Fatalf("seed lock: %v", e)
		}
		lk, e := svc.Read()
		if e != nil {
			t.Fatalf("read lock: %v", e)
		}
		if len(lk.Resources) != 1 {
			t.Fatalf("setup expected 1 resource, got %d", len(lk.Resources))
		}
		lk.Resources[0].InstalledAt = "2000-01-01T00:00:00Z"
		if e := svc.Write(lk); e != nil {
			t.Fatalf("write lock: %v", e)
		}

		// Registry: brew available and will upgrade
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {
			r.Register("brew", func() packages.PackageManager { return &fakeUpgradeMgr{} })
		})
	})
	if err != nil {
		t.Fatalf("upgrade error: %v\n%s", err, out)
	}

	svc := lock.NewYAMLLockService(os.Getenv("PLONK_DIR"))
	lk2, e := svc.Read()
	if e != nil {
		t.Fatalf("read lock: %v", e)
	}
	if lk2.Resources[0].InstalledAt == "2000-01-01T00:00:00Z" {
		t.Fatalf("expected InstalledAt to be updated, still %s", lk2.Resources[0].InstalledAt)
	}
}

// Manager IsAvailable=false should mark failures.
type fakeUnavailableUpgradeMgr struct{}

func (f *fakeUnavailableUpgradeMgr) IsAvailable(ctx context.Context) (bool, error) { return false, nil }
func (f *fakeUnavailableUpgradeMgr) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (f *fakeUnavailableUpgradeMgr) Install(ctx context.Context, name string) error   { return nil }
func (f *fakeUnavailableUpgradeMgr) Uninstall(ctx context.Context, name string) error { return nil }
func (f *fakeUnavailableUpgradeMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeUnavailableUpgradeMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (f *fakeUnavailableUpgradeMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "pipx"}, nil
}
func (f *fakeUnavailableUpgradeMgr) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}
func (f *fakeUnavailableUpgradeMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "pipx"}, nil
}
func (f *fakeUnavailableUpgradeMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeUnavailableUpgradeMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeUnavailableUpgradeMgr) Dependencies() []string                           { return nil }

func TestUpgrade_ManagerAvailableFalse_Fails(t *testing.T) {
	out, err := RunCLI(t, []string{"upgrade", "pipx"}, func(env CLITestEnv) {
		// Seed one pipx package
		svc := lock.NewYAMLLockService(env.ConfigDir)
		_ = svc.AddPackage("pipx", "httpx", "0.1", map[string]interface{}{"manager": "pipx", "name": "httpx", "version": "0.1"})
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {
			r.Register("pipx", func() packages.PackageManager { return &fakeUnavailableUpgradeMgr{} })
		})
	})
	if err == nil {
		t.Fatalf("expected error due to manager unavailable, got nil\n%s", out)
	}
	// Validate JSON path would be too heavy; check that output summary exists with Failed
	var payload struct{}
	_ = json.Unmarshal([]byte(out), &payload) // don't fail; out may be table
	if !strings.Contains(out, "Failed:") {
		t.Fatalf("expected Failed in summary, got:\n%s", out)
	}
}
