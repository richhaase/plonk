package packages

import (
	"context"
	"testing"
)

type depMgr struct{ deps []string }

func (d *depMgr) IsAvailable(ctx context.Context) (bool, error)                     { return true, nil }
func (d *depMgr) ListInstalled(ctx context.Context) ([]string, error)               { return nil, nil }
func (d *depMgr) Install(ctx context.Context, name string) error                    { return nil }
func (d *depMgr) Uninstall(ctx context.Context, name string) error                  { return nil }
func (d *depMgr) IsInstalled(ctx context.Context, name string) (bool, error)        { return false, nil }
func (d *depMgr) InstalledVersion(ctx context.Context, name string) (string, error) { return "", nil }
func (d *depMgr) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name}, nil
}
func (d *depMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (d *depMgr) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "x"}, nil
}
func (d *depMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (d *depMgr) Dependencies() []string                           { return d.deps }

func TestDependencyResolver_Cycle(t *testing.T) {
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("a", func() PackageManager { return &depMgr{deps: []string{"b"}} })
		r.Register("b", func() PackageManager { return &depMgr{deps: []string{"a"}} })
	})
	reg := NewManagerRegistry()
	res := NewDependencyResolver(reg)
	_, err := res.ResolveDependencyOrder([]string{"a", "b"})
	if err == nil {
		t.Fatalf("expected cycle error")
	}
}
