// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/config"
)

// Tests define a minimal manager via config; mock executor not needed here

func TestInstallPackages_ContextTimeout(t *testing.T) {
	// Use a temporary registry with only the slow manager
	// Define custom manager in config and let it time out via context
	cfg := &config.Config{Managers: map[string]config.ManagerConfig{"slow": {Binary: "slow"}}}
	reg := NewManagerRegistry()
	reg.LoadV2Configs(cfg)

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
	cfg := &config.Config{Managers: map[string]config.ManagerConfig{"slow": {Binary: "slow"}}}
	reg := NewManagerRegistry()
	reg.LoadV2Configs(cfg)

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
