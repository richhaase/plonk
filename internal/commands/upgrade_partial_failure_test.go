package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// fakeMgrPartialFailure simulates a manager where some packages upgrade successfully and others fail

// TestUpgrade_PartialFailureWithinSameManager validates that when upgrading multiple packages
// from the same manager, if one fails the others still succeed independently
func TestUpgrade_PartialFailureWithinSameManager(t *testing.T) {
	dir := t.TempDir()

	// Create lock file with 3 packages from same manager
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "pkg-success-1", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-success-1",
		"version": "1.0.0",
	})
	_ = svc.AddPackage("brew", "pkg-fail", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-fail",
		"version": "1.0.0",
	})
	_ = svc.AddPackage("brew", "pkg-success-2", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-success-2",
		"version": "1.0.0",
	})

	// Create spec for upgrading all 3 packages
	spec := upgradeSpec{
		ManagerTargets: map[string][]string{
			"brew": {"pkg-success-1", "pkg-fail", "pkg-success-2"},
		},
	}

	// v2-only: mock brew availability and per-package results (pkg-fail fails)
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version":             {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew upgrade pkg-success-1": {Output: []byte("ok"), Error: nil},
		"brew upgrade pkg-fail":      {Output: []byte("fail"), Error: &packages.MockExitError{Code: 1}},
		"brew upgrade pkg-success-2": {Output: []byte("ok"), Error: nil},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })
	reg := packages.NewManagerRegistry()

	// Execute upgrade
	results, err := executeUpgrade(context.Background(), spec, &config.Config{}, svc, reg)

	// Should return error because one package failed
	if err != nil {
		t.Logf("Expected error returned: %v", err)
	}

	// Verify results
	if len(results.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results.Results))
	}

	// Check each package individually
	resultMap := make(map[string]packageUpgradeResult)
	for _, r := range results.Results {
		resultMap[r.Package] = r
	}

	// pkg-success-1 should succeed
	if r, ok := resultMap["pkg-success-1"]; ok {
		if r.Status != "upgraded" {
			t.Errorf("pkg-success-1 should be upgraded, got status: %s", r.Status)
		}
		if r.Error != "" {
			t.Errorf("pkg-success-1 should have no error, got: %s", r.Error)
		}
	} else {
		t.Error("pkg-success-1 result not found")
	}

	// pkg-fail should fail
	if r, ok := resultMap["pkg-fail"]; ok {
		if r.Status != "failed" {
			t.Errorf("pkg-fail should be failed, got status: %s", r.Status)
		}
		if r.Error == "" {
			t.Error("pkg-fail should have an error message")
		}
		if !strings.Contains(r.Error, "failed to upgrade pkg-fail") {
			t.Errorf("unexpected error message: %s", r.Error)
		}
	} else {
		t.Error("pkg-fail result not found")
	}

	// pkg-success-2 should succeed (even though pkg-fail failed)
	if r, ok := resultMap["pkg-success-2"]; ok {
		if r.Status != "upgraded" {
			t.Errorf("pkg-success-2 should be upgraded, got status: %s", r.Status)
		}
		if r.Error != "" {
			t.Errorf("pkg-success-2 should have no error, got: %s", r.Error)
		}
	} else {
		t.Error("pkg-success-2 result not found")
	}

	// Verify summary counts
	if results.Summary.Total != 3 {
		t.Errorf("expected 3 total packages, got %d", results.Summary.Total)
	}
	if results.Summary.Upgraded != 2 {
		t.Errorf("expected 2 upgraded packages, got %d", results.Summary.Upgraded)
	}
	if results.Summary.Failed != 1 {
		t.Errorf("expected 1 failed package, got %d", results.Summary.Failed)
	}
	if results.Summary.Skipped != 0 {
		t.Errorf("expected 0 skipped packages, got %d", results.Summary.Skipped)
	}

	// Verify lock file was updated for successful packages
	updatedLock, err := svc.Read()
	if err != nil {
		t.Fatalf("failed to read updated lock: %v", err)
	}

	// Verify lock file contains all packages
	if len(updatedLock.Resources) != 3 {
		t.Errorf("expected 3 resources in lock file, got %d", len(updatedLock.Resources))
	}
}

// fakeMgrAlwaysFails always fails on upgrade

// TestUpgrade_AllPackagesFailInManager validates that when all packages fail,
// we get appropriate error reporting
func TestUpgrade_AllPackagesFailInManager(t *testing.T) {
	dir := t.TempDir()

	// Create lock file with 2 packages
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "pkg-a", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-a",
		"version": "1.0.0",
	})
	_ = svc.AddPackage("brew", "pkg-b", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-b",
		"version": "1.0.0",
	})

	// Create spec for upgrading both
	spec := upgradeSpec{
		ManagerTargets: map[string][]string{
			"brew": {"pkg-a", "pkg-b"},
		},
	}

	// v2-only: mock failures for all packages
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version":     {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew upgrade pkg-a": {Output: []byte("fail"), Error: &packages.MockExitError{Code: 1}},
		"brew upgrade pkg-b": {Output: []byte("fail"), Error: &packages.MockExitError{Code: 1}},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })
	reg := packages.NewManagerRegistry()
	results, _ := executeUpgrade(context.Background(), spec, &config.Config{}, svc, reg)

	// executeUpgrade doesn't return error, but sets Failed count
	// The runUpgrade wrapper converts Failed count to error

	// All should be marked as failed
	if results.Summary.Failed != 2 {
		t.Errorf("expected 2 failed, got %d", results.Summary.Failed)
	}
	if results.Summary.Upgraded != 0 {
		t.Errorf("expected 0 upgraded, got %d", results.Summary.Upgraded)
	}
}

// fakeMgrUpgradeTracking tracks Upgrade calls for testing

// TestUpgrade_PerPackageUpgradeCalled validates that Upgrade() is called once per package,
// not once for all packages (the old behavior that caused the bug)
func TestUpgrade_PerPackageUpgradeCalled(t *testing.T) {
	dir := t.TempDir()

	// Create lock file with 3 packages
	svc := lock.NewYAMLLockService(dir)
	_ = svc.AddPackage("brew", "pkg-1", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-1",
		"version": "1.0.0",
	})
	_ = svc.AddPackage("brew", "pkg-2", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-2",
		"version": "1.0.0",
	})
	_ = svc.AddPackage("brew", "pkg-3", "1.0.0", map[string]interface{}{
		"manager": "brew",
		"name":    "pkg-3",
		"version": "1.0.0",
	})

	spec := upgradeSpec{
		ManagerTargets: map[string][]string{
			"brew": {"pkg-1", "pkg-2", "pkg-3"},
		},
	}

	// v2-only: verify three separate upgrade commands were issued
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version":     {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew upgrade pkg-1": {Output: []byte("ok"), Error: nil},
		"brew upgrade pkg-2": {Output: []byte("ok"), Error: nil},
		"brew upgrade pkg-3": {Output: []byte("ok"), Error: nil},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })
	reg := packages.NewManagerRegistry()
	_, err := executeUpgrade(context.Background(), spec, &config.Config{}, svc, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Count upgrade calls
	count := 0
	for _, c := range mock.Commands {
		if c.Name == "brew" && len(c.Args) >= 2 && c.Args[0] == "upgrade" {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("expected 3 upgrade calls, got %d (commands: %+v)", count, mock.Commands)
	}
}
