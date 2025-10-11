package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// stubLockService is a minimal LockService for upgrade tests
type stubLockService struct{}

func (s *stubLockService) Read() (*lock.Lock, error) {
	return &lock.Lock{Version: lock.LockFileVersion, Resources: []lock.ResourceEntry{}}, nil
}
func (s *stubLockService) Write(l *lock.Lock) error { return nil }
func (s *stubLockService) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	return nil
}
func (s *stubLockService) RemovePackage(manager, name string) error                 { return nil }
func (s *stubLockService) GetPackages(manager string) ([]lock.ResourceEntry, error) { return nil, nil }
func (s *stubLockService) HasPackage(manager, name string) bool                     { return false }
func (s *stubLockService) FindPackage(name string) []lock.ResourceEntry             { return nil }

// fixedVersionManager allows observing calls without affecting order
type fixedVersionManager struct{}

func (f *fixedVersionManager) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (f *fixedVersionManager) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (f *fixedVersionManager) Install(ctx context.Context, name string) error      { return nil }
func (f *fixedVersionManager) Uninstall(ctx context.Context, name string) error    { return nil }
func (f *fixedVersionManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (f *fixedVersionManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (f *fixedVersionManager) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}
func (f *fixedVersionManager) SelfInstall(ctx context.Context) error            { return nil }
func (f *fixedVersionManager) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (f *fixedVersionManager) Dependencies() []string                           { return nil }

func TestExecuteUpgrade_DeterministicOrdering(t *testing.T) {
	cfg := &config.Config{}
	lockSvc := &stubLockService{}

	// Build spec with intentionally unsorted map keys and package lists
	spec := upgradeSpec{ManagerTargets: map[string][]string{
		"npm":  {"b", "a"},
		"brew": {"d", "c"},
	}}

	// Inject registry with fake managers for both entries
	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
		r.Register("brew", func() packages.PackageManager { return &fixedVersionManager{} })
		r.Register("npm", func() packages.PackageManager { return &fixedVersionManager{} })
	})
	reg := packages.NewManagerRegistry()

	res, err := executeUpgrade(context.Background(), spec, cfg, lockSvc, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]string, 0, len(res.Results))
	for _, r := range res.Results {
		got = append(got, r.Manager+":"+r.Package)
	}

	// Expect managers then packages sorted: brew:c, brew:d, npm:a, npm:b
	want := []string{"brew:c", "brew:d", "npm:a", "npm:b"}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %q want %q (full: %#v)", i, got[i], want[i], got)
		}
	}
}
