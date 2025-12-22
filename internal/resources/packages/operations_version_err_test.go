package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

func TestInstall_VersionErrorStillAdded(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	// Mock executor to allow install (hardcoded managers are used)
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew install jq": {Output: []byte("installed"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	ls := lock.NewYAMLLockService(t.TempDir())
	reg := GetRegistry()
	res, err := InstallPackagesWith(context.Background(), cfg, ls, reg, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res) != 1 || res[0].Status != "added" {
		t.Fatalf("expected added, got %+v", res)
	}
}
