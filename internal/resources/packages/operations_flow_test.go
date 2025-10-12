package packages

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
)

// Tests use config + mock executor for package manager behavior

func TestOperations_Install_Uninstall_Flow_WithLock(t *testing.T) {
	// temp config dir
	configDir := t.TempDir()

	// v2: use default brew manager with mock executor
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew install jq":   {Output: []byte("installed"), Error: nil},
		"brew uninstall jq": {Output: []byte("uninstalled"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })

	// Install package (non-dry-run) → writes to lock
	res, err := InstallPackages(context.Background(), configDir, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("InstallPackages error: %v", err)
	}
	if len(res) != 1 || res[0].Status != "added" {
		t.Fatalf("unexpected results: %+v", res)
	}

	// Validate lock file has entry
	svc := lock.NewYAMLLockService(configDir)
	lk, err := svc.Read()
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	if len(lk.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(lk.Resources))
	}

	// Uninstall package (non-dry-run) → removes from lock
	ures, err := UninstallPackages(context.Background(), configDir, []string{"jq"}, UninstallOptions{})
	if err != nil {
		t.Fatalf("UninstallPackages error: %v", err)
	}
	if len(ures) != 1 || ures[0].Status != "removed" {
		t.Fatalf("unexpected uninstall results: %+v", ures)
	}

	lk2, err := svc.Read()
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	if len(lk2.Resources) != 0 {
		t.Fatalf("expected 0 resources after removal, got %d", len(lk2.Resources))
	}
}

func TestOperations_ManagerSpecified_And_DryRun(t *testing.T) {
	configDir := t.TempDir()

	// Dry-run should not write lock
	res, err := InstallPackages(context.Background(), configDir, []string{"jq"}, InstallOptions{Manager: "brew", DryRun: true})
	if err != nil {
		t.Fatalf("InstallPackages: %v", err)
	}
	if len(res) != 1 || res[0].Status != "would-add" {
		t.Fatalf("unexpected res: %+v", res)
	}

	// Lock file should not exist
	if _, err := os.Stat(filepath.Join(configDir, lock.LockFileName)); !os.IsNotExist(err) {
		t.Fatalf("lock file should not exist on dry-run")
	}
}
