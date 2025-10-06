package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
)

type installOnlyMgr struct{}

func (i *installOnlyMgr) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (i *installOnlyMgr) ListInstalled(ctx context.Context) ([]string, error) { return []string{}, nil }
func (i *installOnlyMgr) Install(ctx context.Context, name string) error      { return nil }
func (i *installOnlyMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (i *installOnlyMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (i *installOnlyMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (i *installOnlyMgr) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name, Manager: "brew"}, nil
}
func (i *installOnlyMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (i *installOnlyMgr) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "brew"}, nil
}
func (i *installOnlyMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (i *installOnlyMgr) Dependencies() []string                           { return nil }

func TestPackagesApply_InstallsMissing(t *testing.T) {
	dir := t.TempDir()
	// Seed lock with one desired package
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})

	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("brew", func() PackageManager { return &installOnlyMgr{} }) })

	res, err := Apply(context.Background(), dir, config.LoadWithDefaults(dir), false)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.TotalInstalled == 0 {
		t.Fatalf("expected installs > 0, got: %+v", res)
	}
	// Also exercise StructuredData path for coverage
	_ = output.NewPackageOperationFormatter(output.PackageOperationOutput{Command: "install"}).StructuredData()
}
