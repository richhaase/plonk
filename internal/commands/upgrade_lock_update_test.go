package commands

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestUpgrade_UpdatesLockFileVersion(t *testing.T) {
	dir := t.TempDir()
	// Seed lock with brew:jq@1.0.0
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})

	// Mock brew upgrade success
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew upgrade jq": {Output: []byte("ok"), Error: nil},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })

	spec := upgradeSpec{ManagerTargets: map[string][]string{"brew": {"jq"}}}
	cfg := &config.Config{}
	reg := packages.GetRegistry()
	res, err := Upgrade(context.Background(), spec, cfg, svc, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary.Upgraded == 0 {
		t.Fatalf("expected at least one upgraded package, got %+v", res.Summary)
	}

	// Verify lock updated
	lf, err := svc.Read()
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	found := false
	for _, r := range lf.Resources {
		if r.Type == "package" && r.ID == "brew:jq" {
			found = true
		}
	}
	if !found {
		t.Fatalf("brew:jq not found in lock after upgrade")
	}
}
