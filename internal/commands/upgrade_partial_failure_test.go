package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// fakeMgrPartialFailure simulates a manager where some packages upgrade successfully and others fail
type fakeMgrPartialFailure struct {
	failPackage string // Package that should fail
	upgraded    map[string]string
}

func newFakeMgrPartialFailure(failPackage string) *fakeMgrPartialFailure {
	return &fakeMgrPartialFailure{
		failPackage: failPackage,
		upgraded:    make(map[string]string),
	}
}

func (f *fakeMgrPartialFailure) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}

func (f *fakeMgrPartialFailure) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (f *fakeMgrPartialFailure) Install(ctx context.Context, name string) error {
	return nil
}

func (f *fakeMgrPartialFailure) Uninstall(ctx context.Context, name string) error {
	return nil
}

func (f *fakeMgrPartialFailure) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}

func (f *fakeMgrPartialFailure) InstalledVersion(ctx context.Context, name string) (string, error) {
	// If package was upgraded, return new version
	if newVer, ok := f.upgraded[name]; ok {
		return newVer, nil
	}
	// Otherwise return old version
	return "1.0.0", nil
}

func (f *fakeMgrPartialFailure) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}

func (f *fakeMgrPartialFailure) SelfInstall(ctx context.Context) error {
	return nil
}

func (f *fakeMgrPartialFailure) Upgrade(ctx context.Context, pkgs []string) error {
	// Should only be called with single package now (per-package upgrade)
	if len(pkgs) != 1 {
		return fmt.Errorf("expected single package, got %d", len(pkgs))
	}

	pkg := pkgs[0]
	if pkg == f.failPackage {
		return fmt.Errorf("simulated upgrade failure for %s", pkg)
	}

	// Success - mark as upgraded
	f.upgraded[pkg] = "2.0.0"
	return nil
}

func (f *fakeMgrPartialFailure) Dependencies() []string {
	return nil
}

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

	// Setup registry with fake manager that fails on "pkg-fail"
	fakeMgr := newFakeMgrPartialFailure("pkg-fail")
	packages.WithTemporaryRegistry(t, func(reg *packages.ManagerRegistry) {
		reg.Register("brew", func() packages.PackageManager { return fakeMgr })

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
			if r.Error != "simulated upgrade failure for pkg-fail" {
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
	})
}

// fakeMgrAlwaysFails always fails on upgrade
type fakeMgrAlwaysFails struct{}

func (f *fakeMgrAlwaysFails) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}

func (f *fakeMgrAlwaysFails) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (f *fakeMgrAlwaysFails) Install(ctx context.Context, name string) error {
	return nil
}

func (f *fakeMgrAlwaysFails) Uninstall(ctx context.Context, name string) error {
	return nil
}

func (f *fakeMgrAlwaysFails) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}

func (f *fakeMgrAlwaysFails) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}

func (f *fakeMgrAlwaysFails) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}

func (f *fakeMgrAlwaysFails) SelfInstall(ctx context.Context) error {
	return nil
}

func (f *fakeMgrAlwaysFails) Upgrade(ctx context.Context, pkgs []string) error {
	return fmt.Errorf("upgrade failed for all packages")
}

func (f *fakeMgrAlwaysFails) Dependencies() []string {
	return nil
}

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

	// Manager that fails all packages
	fakeMgr := &fakeMgrAlwaysFails{}

	packages.WithTemporaryRegistry(t, func(reg *packages.ManagerRegistry) {
		reg.Register("brew", func() packages.PackageManager { return fakeMgr })

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
	})
}

// fakeMgrUpgradeTracking tracks Upgrade calls for testing
type fakeMgrUpgradeTracking struct {
	calls [][]string // Each slice contains packages passed to Upgrade
}

func (f *fakeMgrUpgradeTracking) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}

func (f *fakeMgrUpgradeTracking) ListInstalled(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (f *fakeMgrUpgradeTracking) Install(ctx context.Context, name string) error {
	return nil
}

func (f *fakeMgrUpgradeTracking) Uninstall(ctx context.Context, name string) error {
	return nil
}

func (f *fakeMgrUpgradeTracking) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}

func (f *fakeMgrUpgradeTracking) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "2.0.0", nil
}

func (f *fakeMgrUpgradeTracking) Search(ctx context.Context, q string) ([]string, error) {
	return nil, nil
}

func (f *fakeMgrUpgradeTracking) SelfInstall(ctx context.Context) error {
	return nil
}

func (f *fakeMgrUpgradeTracking) Upgrade(ctx context.Context, pkgs []string) error {
	f.calls = append(f.calls, pkgs)
	return nil
}

func (f *fakeMgrUpgradeTracking) Dependencies() []string {
	return nil
}

// TestUpgrade_PerPackageUpgradeCalled validates that Upgrade() is called once per package,
// not once for all packages (the old behavior that caused the bug)
func TestUpgrade_PerPackageUpgradeCalled(t *testing.T) {
	dir := t.TempDir()

	fakeMgr := &fakeMgrUpgradeTracking{}

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

	packages.WithTemporaryRegistry(t, func(reg *packages.ManagerRegistry) {
		reg.Register("brew", func() packages.PackageManager { return fakeMgr })

		_, err := executeUpgrade(context.Background(), spec, &config.Config{}, svc, reg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have been called 3 times, once per package
		if len(fakeMgr.calls) != 3 {
			t.Fatalf("expected Upgrade() to be called 3 times, got %d", len(fakeMgr.calls))
		}

		// Each call should have exactly 1 package
		for i, call := range fakeMgr.calls {
			if len(call) != 1 {
				t.Errorf("call %d: expected 1 package, got %d: %v", i, len(call), call)
			}
		}

		// Verify we called with the right packages (in order)
		expectedPackages := []string{"pkg-1", "pkg-2", "pkg-3"}
		for i, call := range fakeMgr.calls {
			if len(call) > 0 && call[0] != expectedPackages[i] {
				t.Errorf("call %d: expected package %s, got %s", i, expectedPackages[i], call[0])
			}
		}
	})
}
