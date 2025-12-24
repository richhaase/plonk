// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

// Tests use config + mock executor

// failingAddLockService wraps MockLockService to fail AddPackage
type failingAddLockService struct{ *MockLockService }

func (f *failingAddLockService) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	return errors.New("simulated add failure")
}

func TestInstallPackagesWith_AlreadyManaged_Skips(t *testing.T) {
	cfg := &config.Config{DefaultManager: "brew"}
	lockSvc := NewMockLockService()
	lockSvc.SetPackageExists("brew", "jq")

	reg := GetRegistry()

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
	// Use mock executor for brew install (hardcoded managers are used)
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":  {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew install jq": {Output: []byte("installed"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	reg := GetRegistry()

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
	// Use mock executor for brew uninstall (hardcoded managers are used)
	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{
		"brew --version":    {Output: []byte("Homebrew 4.0"), Error: nil},
		"brew uninstall jq": {Output: []byte("uninstalled"), Error: nil},
	}}
	SetDefaultExecutor(mock)
	t.Cleanup(func() { SetDefaultExecutor(&RealCommandExecutor{}) })
	reg := GetRegistry()

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
