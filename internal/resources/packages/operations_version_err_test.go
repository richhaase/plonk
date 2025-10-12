package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

func TestInstall_VersionErrorStillAdded(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew", Managers: map[string]config.ManagerConfig{
		"brew": {Binary: "brew", Install: config.CommandConfig{Command: []string{"brew", "install", "{{.Package}}"}}},
	}}
	// Mock executor to allow install
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew install jq": {Output: []byte("installed"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	ls := lock.NewYAMLLockService(t.TempDir())
	reg := NewManagerRegistry()
	res, err := InstallPackagesWith(context.Background(), cfg, ls, reg, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "added" {
		t.Fatalf("expected added, got %+v", res)
	}
}
