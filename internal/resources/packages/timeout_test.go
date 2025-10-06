// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
	"time"
)

// slowManager simulates long-running operations that respect context deadlines.
type slowManager struct{}

func (s *slowManager) IsAvailable(ctx context.Context) (bool, error)              { return true, nil }
func (s *slowManager) ListInstalled(ctx context.Context) ([]string, error)        { return []string{}, nil }
func (s *slowManager) Install(ctx context.Context, name string) error             { return blockUntilDone(ctx) }
func (s *slowManager) Uninstall(ctx context.Context, name string) error           { return blockUntilDone(ctx) }
func (s *slowManager) IsInstalled(ctx context.Context, name string) (bool, error) { return false, nil }
func (s *slowManager) InstalledVersion(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (s *slowManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
	return &PackageInfo{Name: name, Manager: "slow", Installed: false}, nil
}
func (s *slowManager) Search(ctx context.Context, query string) ([]string, error) {
	return []string{}, nil
}
func (s *slowManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return &HealthCheck{Name: "slow", Category: "package-manager", Status: "PASS"}, nil
}
func (s *slowManager) Upgrade(ctx context.Context, packages []string) error {
	return blockUntilDone(ctx)
}
func (s *slowManager) Dependencies() []string { return nil }

func blockUntilDone(ctx context.Context) error {
	// Sleep longer than the test timeout; return when context is done
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(500 * time.Millisecond):
		return nil
	}
}

func TestInstallPackages_ContextTimeout(t *testing.T) {
	// Use a temporary registry with only the slow manager
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("slow", func() PackageManager { return &slowManager{} }) })

	// Very short timeout to trigger cancellation
	parent, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	// Run install using the slow manager
	results, err := InstallPackages(parent, t.TempDir(), []string{"tool"}, InstallOptions{Manager: "slow", DryRun: false})
	if err != nil {
		t.Fatalf("InstallPackages returned unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected status 'failed', got %q", results[0].Status)
	}
	if results[0].Error == nil || (results[0].Error != nil && results[0].Error.Error() == "") {
		t.Fatalf("expected contextual error, got: %v", results[0].Error)
	}
}

func TestUninstallPackages_ContextTimeout(t *testing.T) {
	// Use a temporary registry with only the slow manager
	WithTemporaryRegistry(t, func(r *ManagerRegistry) { r.Register("slow", func() PackageManager { return &slowManager{} }) })

	parent, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	results, err := UninstallPackages(parent, t.TempDir(), []string{"tool"}, UninstallOptions{Manager: "slow", DryRun: false})
	if err != nil {
		t.Fatalf("UninstallPackages returned unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "failed" && results[0].Status != "removed" {
		// Depending on timing, slow manager may return failure; 'removed' would only occur if it returned before timeout
		t.Fatalf("expected status 'failed' or 'removed', got %q", results[0].Status)
	}
}
