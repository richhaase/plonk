package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestUpgrade_ManagerErrorCountsFailed(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "a", "1.0.0", map[string]interface{}{"manager": "brew", "name": "a", "version": "1.0.0"})

	// Mark brew available but make upgrade fail
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version": {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew upgrade a": {Output: []byte("failed"), Error: &packages.MockExitError{Code: 1}},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })

	spec := upgradeSpec{ManagerTargets: map[string][]string{"brew": {"a"}}}
	res2, err := Upgrade(context.Background(), spec, &config.Config{}, svc, packages.NewManagerRegistry())
	if err != nil {
		t.Fatalf("unexpected error from Upgrade: %v", err)
	}
	if res2.Summary.Failed == 0 {
		t.Fatalf("expected failed count > 0")
	}
}
