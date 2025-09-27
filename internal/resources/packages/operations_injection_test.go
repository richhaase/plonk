// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

// simpleFakeManager is a minimal PackageManager used for injection tests
type simpleFakeManager struct{}

func (m *simpleFakeManager) IsAvailable(ctx context.Context) (bool, error)       { return true, nil }
func (m *simpleFakeManager) ListInstalled(ctx context.Context) ([]string, error) { return nil, nil }
func (m *simpleFakeManager) Install(ctx context.Context, name string) error      { return nil }
func (m *simpleFakeManager) Uninstall(ctx context.Context, name string) error    { return nil }
func (m *simpleFakeManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (m *simpleFakeManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "1.0.0", nil
}
func (m *simpleFakeManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name, Manager: "brew", Installed: true, Version: "1.0.0"}, nil
}
func (m *simpleFakeManager) Search(ctx context.Context, query string) ([]string, error) {
	return nil, nil
}
func (m *simpleFakeManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "brew"}, nil
}
func (m *simpleFakeManager) SelfInstall(ctx context.Context) error            { return nil }
func (m *simpleFakeManager) Upgrade(ctx context.Context, pkgs []string) error { return nil }
func (m *simpleFakeManager) Dependencies() []string                           { return nil }

// failingAddLockService wraps MockLockService to fail AddPackage
type failingAddLockService struct{ *MockLockService }

func (f *failingAddLockService) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	return errors.New("simulated add failure")
}

func TestInstallPackagesWith_AlreadyManaged_Skips(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	lockSvc := NewMockLockService()
	lockSvc.SetPackageExists("brew", "jq")

	// Inject fake manager into temporary registry
	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &simpleFakeManager{} })
	})
	reg := NewManagerRegistry()

	ctx := context.Background()
	results, err := InstallPackagesWith(ctx, cfg, lockSvc, reg, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "skipped" || !results[0].AlreadyManaged {
		t.Fatalf("expected skipped due to already-managed, got %+v", results[0])
	}
}

func TestInstallPackagesWith_LockAddFailure_ReportsFailed(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	lockSvc := &failingAddLockService{NewMockLockService()}

	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &simpleFakeManager{} })
	})
	reg := NewManagerRegistry()

	ctx := context.Background()
	results, err := InstallPackagesWith(ctx, cfg, lockSvc, reg, []string{"jq"}, InstallOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected failed due to lock add failure, got %+v", results[0])
	}
}

func TestUninstallPackagesWith_PassThrough_WhenNotManaged(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	lockSvc := NewMockLockService() // empty, so package is not managed

	WithTemporaryRegistry(t, func(r *ManagerRegistry) {
		r.Register("brew", func() PackageManager { return &simpleFakeManager{} })
	})
	reg := NewManagerRegistry()

	ctx := context.Background()
	results, err := UninstallPackagesWith(ctx, cfg, lockSvc, reg, []string{"jq"}, UninstallOptions{Manager: "brew"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "removed" {
		t.Fatalf("expected removed (pass-through uninstall), got %+v", results[0])
	}
}
