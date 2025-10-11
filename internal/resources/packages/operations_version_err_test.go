package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

type verErrMgr struct{}

func (v *verErrMgr) IsAvailable(ctx context.Context) (bool, error)              { return true, nil }
func (v *verErrMgr) ListInstalled(ctx context.Context) ([]string, error)        { return nil, nil }
func (v *verErrMgr) Install(ctx context.Context, name string) error             { return nil }
func (v *verErrMgr) Uninstall(ctx context.Context, name string) error           { return nil }
func (v *verErrMgr) IsInstalled(ctx context.Context, name string) (bool, error) { return false, nil }
func (v *verErrMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", errors.New("oops")
}
func (v *verErrMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (v *verErrMgr) Dependencies() []string                           { return nil }

func TestInstall_VersionErrorStillAdded(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	ls := lock.NewYAMLLockService(t.TempDir())
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("brew", func() PackageManager { return &verErrMgr{} }) })
	reg := NewManagerRegistry()
	res, err := InstallPackagesWith(context.Background(), cfg, ls, reg, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "added" {
		t.Fatalf("expected added, got %+v", res)
	}
}
