package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
)

func TestPackagesApply_InstallsMissing(t *testing.T) {
	dir := t.TempDir()
	// Seed lock with one desired package
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "jq", "1.0.0", map[string]interface{}{"manager": "brew", "name": "jq", "version": "1.0.0"})

	// Mark brew available and treat install as success
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew install jq": {Output: []byte("ok"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	res, err := Apply(context.Background(), dir, config.LoadWithDefaults(dir), false)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.TotalInstalled == 0 {
		t.Fatalf("expected installs > 0, got: %+v", res)
	}
	// Also exercise StructuredData path for coverage
	_ = output.NewPackageOperationFormatter(output.PackageOperationOutput{Command: "install"}).StructuredData()
}
