package packages

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
)

// fakePM is a simple in-memory PackageManager for flow tests
type fakePM struct {
	available bool
	installed map[string]bool
	versions  map[string]string
}

func newFakePM() *fakePM {
	return &fakePM{available: true, installed: map[string]bool{}, versions: map[string]string{}}
}

func (f *fakePM) IsAvailable(ctx context.Context) (bool, error) { return f.available, nil }
func (f *fakePM) ListInstalled(ctx context.Context) ([]string, error) {
	var out []string
	for k, v := range f.installed {
		if v {
			out = append(out, k)
		}
	}
	return out, nil
}
func (f *fakePM) Install(ctx context.Context, name string) error {
	f.installed[name] = true
	if f.versions[name] == "" {
		f.versions[name] = "1.0.0"
	}
	return nil
}
func (f *fakePM) Uninstall(ctx context.Context, name string) error {
	delete(f.installed, name)
	return nil
}
func (f *fakePM) IsInstalled(ctx context.Context, name string) (bool, error) {
	return f.installed[name], nil
}
func (f *fakePM) InstalledVersion(ctx context.Context, name string) (string, error) {
	return f.versions[name], nil
}
func (f *fakePM) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name, Manager: "brew", Installed: f.installed[name], Version: f.versions[name]}, nil
}
func (f *fakePM) Search(ctx context.Context, query string) ([]string, error) { return []string{}, nil }
func (f *fakePM) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "fake", Category: "pm", Status: "pass"}, nil
}
func (f *fakePM) Upgrade(ctx context.Context, packages []string) error {
	for _, p := range packages {
		if f.installed[p] {
			f.versions[p] = "1.0.1"
		}
	}
	return nil
}
func (f *fakePM) Dependencies() []string { return nil }

func TestOperations_Install_Uninstall_Flow_WithLock(t *testing.T) {
	// temp config dir
	configDir := t.TempDir()

	// Isolate registry and register fake for brew
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return newFakePM() })
	})

	// Install package (non-dry-run) → writes to lock
	res, err := InstallPackages(context.Background(), configDir, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("InstallPackages error: %v", err)
	}
	if len(res) != 1 || res[0].Status != "added" {
		t.Fatalf("unexpected results: %+v", res)
	}

	// Validate lock file has entry
	svc := lock.NewYAMLLockService(configDir)
	lk, err := svc.Read()
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}

	// Uninstall package (non-dry-run) → removes from lock
	ures, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{})
	if err != nil {
		t.Fatalf("UninstallPackages error: %v", err)
	}
	if len(ures) != 1 || ures[0].Status != "removed" {
		t.Fatalf("unexpected uninstall results: %+v", ures)
	}

	lk2, err := svc.Read()
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	if len(lk2.Resources) != 0 {
		t.Fatalf("expected 0 resources after removal, got %d", len(lk2.Resources))
	}
}

func TestOperations_ManagerSpecified_And_DryRun(t *testing.T) {
	configDir := t.TempDir()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("brew", func() PackageManager { return newFakePM() }) })

	// Dry-run should not write lock
	res, err := InstallPackages(context.Background(), configDir, []string{"jq"}, InstallOptions{Manager: "brew", DryRun: true})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if len(res) != 1 || res[0].Status != "would-add" {
		t.Fatalf("unexpected res: %+v", res)
	}

	// Lock file should not exist
	if _, err := os.Stat(filepath.Join(configDir, lock.LockFileName)); !os.IsNotExist(err) {
		t.Fatalf("lock file should not exist on dry-run")
	}
}
