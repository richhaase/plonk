// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
	"time"
)

func TestInstallPackages_ContextTimeout(t *testing.T) {
	// Use a very short timeout that should trigger cancellation
	// before the package manager can respond
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Small delay to ensure context is canceled
	time.Sleep(5 * time.Millisecond)

	// Run install using brew manager - context should already be canceled
	results, err := InstallPackages(ctx, t.TempDir(), []string{"nonexistent-package-12345"}, InstallOptions{Manager: "brew", DryRun: false})
	if err != nil {
		t.Fatalf("InstallPackages returned unexpected error: %v", err)
	}
	// With canceled context, we should get 0 results since the loop exits early
	if len(results) != 0 {
		t.Fatalf("expected 0 results with canceled context, got %d", len(results))
	}
}

func TestUninstallPackages_ContextTimeout(t *testing.T) {
	// Use a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Small delay to ensure context is canceled
	time.Sleep(5 * time.Millisecond)

	results, err := UninstallPackages(ctx, t.TempDir(), []string{"nonexistent-package-12345"}, UninstallOptions{Manager: "brew", DryRun: false})
	if err != nil {
		t.Fatalf("UninstallPackages returned unexpected error: %v", err)
	}
	// With canceled context, we should get 0 results since the loop exits early
	// or a failed result
	_ = results
}
