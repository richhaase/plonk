package diagnostics

import (
	"context"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

type passMgr struct{}

func (p *passMgr) IsAvailable(ctx context.Context) (bool, error)                     { return true, nil }
func (p *passMgr) ListInstalled(ctx context.Context) ([]string, error)               { return nil, nil }
func (p *passMgr) Install(ctx context.Context, name string) error                    { return nil }
func (p *passMgr) Uninstall(ctx context.Context, name string) error                  { return nil }
func (p *passMgr) IsInstalled(ctx context.Context, name string) (bool, error)        { return false, nil }
func (p *passMgr) InstalledVersion(ctx context.Context, name string) (string, error) { return "", nil }
func (p *passMgr) Search(ctx context.Context, q string) ([]string, error)            { return nil, nil }
func (p *passMgr) SelfInstall(ctx context.Context) error                             { return nil }
func (p *passMgr) Upgrade(ctx context.Context, pkgs []string) error                  { return nil }
func (p *passMgr) Dependencies() []string                                            { return nil }

func TestPackageManagerEcosystem_Check(t *testing.T) {
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &passMgr{} })
		r.Register("npm", func() packages.PackageManager { return &passMgr{} })
	})
	rep := RunHealthChecks()
	found := false
	for _, c := range rep.Checks {
		if c.Name == "Package Managers" && c.Category == "package-managers" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected package managers check present")
	}
}
