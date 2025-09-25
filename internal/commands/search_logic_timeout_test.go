package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// slowSearchMgr simulates a manager that times out when searching
type slowSearchMgr struct{}

func (s *slowSearchMgr) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (s *slowSearchMgr) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (s *slowSearchMgr) Install(ctx context.Context, name string) error      { return nil }
func (s *slowSearchMgr) Uninstall(ctx context.Context, name string) error    { return nil }
func (s *slowSearchMgr) IsInstalled(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (s *slowSearchMgr) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (s *slowSearchMgr) Info(ctx context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "slow"}, nil
}
func (s *slowSearchMgr) Search(ctx context.Context, q string) ([]string, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
func (s *slowSearchMgr) CheckHealth(ctx context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "slow"}, nil
}
func (s *slowSearchMgr) SelfInstall(ctx context.Context) error            { return nil }
func (s *slowSearchMgr) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (s *slowSearchMgr) Dependencies() []string                           { return nil }

func TestSearch_TimeoutReportsNotFoundWithErrors(t *testing.T) {
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("slow", func() packages.PackageManager { return &slowSearchMgr{} })
	})

	cfg := &config.Config{OperationTimeout: 1} // seconds
	res, err := Search(context.Background(), cfg, "tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "not-found" {
		t.Fatalf("expected not-found, got %s", res.Status)
	}
	if len(res.Message) == 0 {
		t.Fatalf("expected message to mention errors/timeouts")
	}
}
