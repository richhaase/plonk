package packages

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
)

// fake manager variants for targeted branches
type fakeAvailMgr struct{ available bool }

func (f *fakeAvailMgr) IsAvailable(ctx context.Context) (bool, error)              { return f.available, nil }
func (f *fakeAvailMgr) ListInstalled(ctx context.Context) ([]string, error)        { return nil, nil }
func (f *fakeAvailMgr) Install(ctx context.Context, name string) error             { return nil }
func (f *fakeAvailMgr) Uninstall(ctx context.Context, name string) error           { return nil }
func (f *fakeAvailMgr) IsInstalled(ctx context.Context, name string) (bool, error) { return true, nil }
func (f *fakeAvailMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (f *fakeAvailMgr) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name}, nil
}
func (f *fakeAvailMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (f *fakeAvailMgr) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "fake"}, nil
}
func (f *fakeAvailMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeAvailMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeAvailMgr) Dependencies() []string                           { return nil }

type fakeUninstallErrorMgr struct{}

func (f *fakeUninstallErrorMgr) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fakeUninstallErrorMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeUninstallErrorMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeUninstallErrorMgr) Uninstall(ctx context.Context, name string) error {
	return context.DeadlineExceeded
}
func (f *fakeUninstallErrorMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fakeUninstallErrorMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (f *fakeUninstallErrorMgr) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name}, nil
}
func (f *fakeUninstallErrorMgr) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}
func (f *fakeUninstallErrorMgr) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "fake"}, nil
}
func (f *fakeUninstallErrorMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeUninstallErrorMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeUninstallErrorMgr) Dependencies() []string                           { return nil }

func TestInstall_ManagerUnavailable_Suggestion(t *testing.T) {
	configDir := t.TempDir()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		// Register npm but mark unavailable
		r.Register("npm", func() PackageManager { return &fakeAvailMgr{available: false} })
	})

	res, err := InstallPackages(context.Background(), configDir, []string{"prettier"}, InstallOptions{Manager: "npm"})
	if err != nil {
		t.Fatalf("InstallPackages error: %v", err)
	}
	if len(res) != 1 || res[0].Status != "failed" || res[0].Error == nil {
		t.Fatalf("unexpected result: %+v", res)
	}
	if !strings.Contains(res[0].Error.Error(), "install Node.js from") {
		t.Fatalf("expected suggestion in error, got: %v", res[0].Error)
	}
}

func TestInstall_NpmScoped_MetadataSaved(t *testing.T) {
	configDir := t.TempDir()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("npm", func() PackageManager { return &fakeAvailMgr{available: true} })
	})

	pkg := "@scope/typescript"
	res, err := InstallPackages(context.Background(), configDir, []string{pkg}, InstallOptions{Manager: "npm"})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if res[0].Status != "added" {
		t.Fatalf("expected added, got: %+v", res[0])
	}

	// Verify lock metadata includes scope and full_name
	svc := lock.NewYAMLLockService(configDir)
	lk, _ := svc.Read()
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}
	md := lk.Resources[0].Metadata
	if md["scope"] != "@scope" || md["full_name"] != pkg {
		t.Fatalf("expected npm metadata saved, got: %#v", md)
	}
}

func TestInstall_GoSourcePath_SavedAndBinaryNamedInLock(t *testing.T) {
	configDir := t.TempDir()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("go", func() PackageManager { return &fakeAvailMgr{available: true} })
	})

	src := "github.com/foo/bar"
	res, err := InstallPackages(context.Background(), configDir, []string{src}, InstallOptions{Manager: "go"})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if res[0].Status != "added" {
		t.Fatalf("expected added, got: %+v", res[0])
	}

	svc := lock.NewYAMLLockService(configDir)
	lk, _ := svc.Read()
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}
	if lk.Resources[0].ID != "go:bar" {
		t.Fatalf("expected ID go:bar, got %s", lk.Resources[0].ID)
	}
	if lk.Resources[0].Metadata["source_path"] != src {
		t.Fatalf("missing source_path")
	}
}

func TestInstall_LockWriteFailure(t *testing.T) {
	configDir := t.TempDir()
	// Make directory read-only to trigger writer failure
	_ = os.Chmod(configDir, 0500)
	t.Cleanup(func() { _ = os.Chmod(configDir, 0700) })
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &fakeAvailMgr{available: true} })
	})

	res, err := InstallPackages(context.Background(), configDir, []string{"jq"}, InstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if res[0].Status != "failed" || res[0].Error == nil {
		t.Fatalf("expected failed due to lock write failure, got: %+v", res[0])
	}
}

func TestUninstall_PartialSuccess_WhenSystemUninstallFails(t *testing.T) {
	configDir := t.TempDir()
	// Seed lock as managed item
	svc := lock.NewYAMLLockService(configDir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq"})

	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &fakeUninstallErrorMgr{} })
	})

	res, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("UninstallPackages: %v", err)
	}
	if len(res) != 1 || res[0].Status != "removed" || res[0].Error == nil {
		t.Fatalf("expected removed with error detail, got: %+v", res)
	}
	if !strings.Contains(res[0].Error.Error(), "system uninstall failed") {
		t.Fatalf("expected system uninstall failed detail, got: %v", res[0].Error)
	}
}

func TestUninstall_NotManaged_PassThrough(t *testing.T) {
	configDir := t.TempDir()
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &fakeAvailMgr{available: true} })
	})

	res, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("UninstallPackages: %v", err)
	}
	if res[0].Status != "removed" || res[0].Error != nil {
		t.Fatalf("expected removed without error, got: %+v", res[0])
	}
}

func TestUninstall_LockWriteFailure_Error(t *testing.T) {
	configDir := t.TempDir()
	// Seed lock as managed item
	svc := lock.NewYAMLLockService(configDir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq"})

	// Make directory read-only so RemovePackage's write fails
	_ = os.Chmod(configDir, 0500)
	t.Cleanup(func() { _ = os.Chmod(configDir, 0700) })
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &fakeAvailMgr{available: true} })
	})

	res, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("UninstallPackages: %v", err)
	}
	if res[0].Status != "removed" || res[0].Error == nil {
		t.Fatalf("expected removed with lock update error, got: %+v", res[0])
	}
	if !strings.Contains(res[0].Error.Error(), "failed to update lock") {
		t.Fatalf("expected failed to update lock detail, got: %v", res[0].Error)
	}
}
