package diagnostics

import (
	"context"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

type blockingMgr struct{}

func (b *blockingMgr) IsAvailable(ctx context.Context) (bool, error)              { return true, nil }
func (b *blockingMgr) ListInstalled(ctx context.Context) ([]string, error)        { return nil, nil }
func (b *blockingMgr) Install(ctx context.Context, name string) error             { return nil }
func (b *blockingMgr) Uninstall(ctx context.Context, name string) error           { return nil }
func (b *blockingMgr) IsInstalled(ctx context.Context, name string) (bool, error) { return false, nil }
func (b *blockingMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (b *blockingMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "block"}, nil
}
func (b *blockingMgr) Search(ctx context.Context, q string) ([]string, error) { return nil, nil }
func (b *blockingMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	<-ctx.Done()
	return &packages.HealthCheck{Name: "block", Status: "warn"}, nil
}
func (b *blockingMgr) SelfInstall(ctx context.Context) error            { return nil }
func (b *blockingMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (b *blockingMgr) Dependencies() []string                           { return nil }

func TestRunHealthChecksWithContext_Timeout(t *testing.T) {
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("block", func() packages.PackageManager { return &blockingMgr{} })
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()                            // immediately cancel
	_ = RunHealthChecksWithContext(ctx) // ensure it returns without hang
}
