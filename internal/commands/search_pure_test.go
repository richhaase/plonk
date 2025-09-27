package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

type fakeSearchMgr struct {
	name      string
	pkgs      []string
	available bool
}

func (f *fakeSearchMgr) IsAvailable(ctx context.Context) (bool, error)       { return f.available, nil }
func (f *fakeSearchMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fakeSearchMgr) Install(ctx context.Context, name string) error      { return nil }
func (f *fakeSearchMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fakeSearchMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (f *fakeSearchMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (f *fakeSearchMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: f.name}, nil
}
func (f *fakeSearchMgr) Search(ctx context.Context, q string) ([]string, error) { return f.pkgs, nil }
func (f *fakeSearchMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: f.name}, nil
}
func (f *fakeSearchMgr) SelfInstall(ctx context.Context) error            { return nil }
func (f *fakeSearchMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fakeSearchMgr) Dependencies() []string                           { return nil }

func TestSearch_SpecificManagerFound(t *testing.T) {
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager {
			return &fakeSearchMgr{name: "brew", pkgs: []string{"ripgrep"}, available: true}
		})
	})
	cfg := &config.Config{OperationTimeout: 1}
	res, err := Search(context.Background(), cfg, "brew:ripgrep")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Status != "found" {
		t.Fatalf("status=%s", res.Status)
	}
	if len(res.Results) != 1 || res.Results[0].Manager != "brew" {
		t.Fatalf("unexpected results: %+v", res.Results)
	}
}

func TestSearch_AllManagersMultiple(t *testing.T) {
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager {
			return &fakeSearchMgr{name: "brew", pkgs: []string{"ripgrep"}, available: true}
		})
		r.Register("npm", func() packages.PackageManager {
			return &fakeSearchMgr{name: "npm", pkgs: []string{"ripgrep"}, available: true}
		})
	})
	cfg := &config.Config{OperationTimeout: 1}
	res, err := Search(context.Background(), cfg, "ripgrep")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Status != "found-multiple" {
		t.Fatalf("status=%s", res.Status)
	}
	if len(res.Results) != 2 {
		t.Fatalf("expected 2 managers, got %d", len(res.Results))
	}
}

func TestSearch_ManagerUnavailable(t *testing.T) {
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("off", func() packages.PackageManager { return &fakeSearchMgr{name: "off", pkgs: []string{}, available: false} })
	})
	cfg := &config.Config{OperationTimeout: 1}
	res, err := Search(context.Background(), cfg, "off:pkg")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Status != "manager-unavailable" {
		t.Fatalf("status=%s", res.Status)
	}
}
