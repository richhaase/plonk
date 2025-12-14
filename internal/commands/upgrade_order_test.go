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

// Test uses config + mock executor to simulate availability

func TestExecuteUpgrade_DeterministicOrdering(t *testing.T) {
	cfg := &config.Config{}
	lockSvc := &stubLockService{}

	// Build spec with intentionally unsorted map keys and package lists
	spec := upgradeSpec{ManagerTargets: map[string][]string{
		"npm":  {"b", "a"},
		"brew": {"d", "c"},
	}}

	// Use default managers; simulate availability and success
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"npm --version":   {Output: []byte("10.0.0"), Error: nil},
		"brew upgrade d":  {Output: []byte("ok"), Error: nil},
		"brew upgrade c":  {Output: []byte("ok"), Error: nil},
		"npm update a -g": {Output: []byte("ok"), Error: nil},
		"npm update b -g": {Output: []byte("ok"), Error: nil},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })
	reg := packages.GetRegistry()

	res, err := executeUpgrade(context.Background(), spec, cfg, lockSvc, reg, false)
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
