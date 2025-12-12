package packages

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

func TestInstallPackagesWith_UnsupportedManager(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	lockSvc := lock.NewYAMLLockService(t.TempDir())
	reg := GetRegistry() // no registrations

	results, err := InstallPackagesWith(context.Background(), cfg, lockSvc, reg, []string{"tool"}, InstallOptions{Manager: "fake"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected failed, got %s", results[0].Status)
	}
}

func TestInstallPackagesWith_ManagerUnavailable(t *testing.T) {
	cfg := &config.Config{
		Managers: map[string]config.ManagerConfig{
			"off": {Binary: "off"},
		},
	}
	// Mock executor with no responses for "off" so LookPath fails → unavailable
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	lockSvc := lock.NewYAMLLockService(t.TempDir())
	reg := GetRegistry()

	results, err := InstallPackagesWith(context.Background(), cfg, lockSvc, reg, []string{"x"}, InstallOptions{Manager: "off"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected failed, got %s", results[0].Status)
	}
}
