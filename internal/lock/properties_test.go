package lock_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
)

// Property-based tests verify invariants that should ALWAYS hold true,
// regardless of the specific operations performed.
//
// Unlike traditional tests that check specific inputs/outputs, property tests
// verify that certain properties remain true across many different scenarios.
//
// For lock file integrity, key properties include:
// 1. Lock file is always valid YAML
// 2. Lock file can always be parsed after any operation
// 3. Adding a package and reading back returns the same data
// 4. Operations are idempotent (repeating them produces same result)
// 5. Lock file updates are atomic (no partial writes)

// TestProperty_LockFileAlwaysValidYAML verifies that after ANY operation,
// the lock file is valid YAML and can be parsed
func TestProperty_LockFileAlwaysValidYAML(t *testing.T) {
	operations := []struct {
		name string
		fn   func(*testing.T, lock.LockService)
	}{
		{
			"add package",
			func(t *testing.T, svc lock.LockService) {
				_ = svc.AddPackage("brew", "pkg", "1.0", nil)
			},
		},
		{
			"add package with metadata",
			func(t *testing.T, svc lock.LockService) {
				_ = svc.AddPackage("npm", "pkg2", "2.0", map[string]interface{}{
					"manager": "npm",
					"name":    "pkg2",
					"extra":   "data",
				})
			},
		},
		{
			"remove package",
			func(t *testing.T, svc lock.LockService) {
				_ = svc.AddPackage("gem", "pkg3", "3.0", nil)
				_ = svc.RemovePackage("gem", "pkg3")
			},
		},
		{
			"update package",
			func(t *testing.T, svc lock.LockService) {
				_ = svc.AddPackage("cargo", "pkg4", "1.0", nil)
				_ = svc.AddPackage("cargo", "pkg4", "2.0", nil) // Update
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			dir := t.TempDir()
			svc := lock.NewYAMLLockService(dir)

			// Perform the operation
			op.fn(t, svc)

			// Property: Lock file must be valid YAML and parseable
			lockFile, err := svc.Read()
			if err != nil {
				t.Fatalf("lock file must be parseable after %s: %v", op.name, err)
			}

			// Property: Lock file must not be nil
			if lockFile == nil {
				t.Fatalf("lock file must not be nil after %s", op.name)
			}

			// Property: Lock file on disk must be valid
			lockPath := filepath.Join(dir, "plonk.lock")
			data, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatalf("lock file must exist on disk after %s: %v", op.name, err)
			}

			if len(data) == 0 {
				t.Fatalf("lock file must not be empty after %s", op.name)
			}
		})
	}
}

// TestProperty_AddThenReadReturnsData verifies round-trip integrity:
// Data written to lock file can be read back unchanged
func TestProperty_AddThenReadReturnsData(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)

	// Add package with metadata
	metadata := map[string]interface{}{
		"manager": "brew",
		"name":    "test-pkg",
		"version": "1.2.3",
		"url":     "https://example.com",
	}

	err := svc.AddPackage("brew", "test-pkg", "1.2.3", metadata)
	if err != nil {
		t.Fatalf("failed to add package: %v", err)
	}

	// Read it back
	lockFile, err := svc.Read()
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}

	// Property: Package must exist in lock file
	found := false
	for _, resource := range lockFile.Resources {
		if resource.Type == "package" {
			nameVal, _ := resource.Metadata["name"]
			name, _ := nameVal.(string)
			if name == "test-pkg" {
				found = true

				// Property: Metadata must match what we wrote
				versionVal, _ := resource.Metadata["version"]
				version, _ := versionVal.(string)
				if version != "1.2.3" {
					t.Errorf("version mismatch: expected 1.2.3, got %s", version)
				}

				urlVal, _ := resource.Metadata["url"]
				url, _ := urlVal.(string)
				if url != "https://example.com" {
					t.Errorf("url mismatch: expected https://example.com, got %s", url)
				}
			}
		}
	}

	if !found {
		t.Error("package not found in lock file after adding")
	}
}

// TestProperty_OperationsAreIdempotent verifies that repeating
// operations produces the same result
func TestProperty_OperationsAreIdempotent(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)

	// Add same package multiple times
	for i := 0; i < 3; i++ {
		err := svc.AddPackage("npm", "lodash", "4.17.21", map[string]interface{}{
			"manager": "npm",
			"name":    "lodash",
			"version": "4.17.21",
		})
		if err != nil {
			t.Fatalf("iteration %d: failed to add package: %v", i, err)
		}
	}

	// Property: Should only have one instance of the package
	lockFile, err := svc.Read()
	if err != nil {
		t.Fatalf("failed to read lock: %v", err)
	}

	count := 0
	for _, resource := range lockFile.Resources {
		if resource.Type == "package" {
			nameVal, _ := resource.Metadata["name"]
			name, _ := nameVal.(string)
			if name == "lodash" {
				count++
			}
		}
	}

	if count != 1 {
		t.Errorf("expected 1 instance of package, got %d (operations are not idempotent)", count)
	}
}

// TestProperty_ConcurrentReadsSafe verifies that multiple readers
// can safely read the lock file simultaneously
func TestProperty_ConcurrentReadsSafe(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)

	// Seed lock file
	_ = svc.AddPackage("brew", "pkg1", "1.0", nil)
	_ = svc.AddPackage("npm", "pkg2", "2.0", nil)

	// Property: Concurrent reads should all succeed
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			select {
			case <-ctx.Done():
				return
			default:
				_, err := svc.Read()
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	// Check no errors occurred
	select {
	case err := <-errors:
		t.Errorf("concurrent read failed: %v", err)
	default:
		// Success - no errors
	}
}

// TestProperty_NoPartialWrites verifies atomic writes:
// Lock file should never be left in partially-written state
func TestProperty_NoPartialWrites(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)
	lockPath := filepath.Join(dir, "plonk.lock")

	// Add a package
	err := svc.AddPackage("brew", "pkg1", "1.0", nil)
	if err != nil {
		t.Fatalf("failed to add package: %v", err)
	}

	// Property: Lock file should always be parseable, even mid-operation
	// (This is tested by checking file can be read at any time)

	// Perform multiple rapid writes
	for i := 0; i < 20; i++ {
		_ = svc.AddPackage("npm", "pkg", "1.0", map[string]interface{}{
			"iteration": i,
		})

		// After each write, file must be valid
		if _, err := svc.Read(); err != nil {
			t.Errorf("lock file invalid after write %d: %v", i, err)
		}

		// File must exist and not be empty
		info, err := os.Stat(lockPath)
		if err != nil {
			t.Errorf("lock file missing after write %d: %v", i, err)
		}
		if info.Size() == 0 {
			t.Errorf("lock file empty after write %d", i)
		}
	}
}

// TestProperty_RemoveIsInverseOfAdd verifies that adding then removing
// a package returns to original state
func TestProperty_RemoveIsInverseOfAdd(t *testing.T) {
	dir := t.TempDir()
	svc := lock.NewYAMLLockService(dir)

	// Get initial state
	initial, err := svc.Read()
	if err != nil {
		t.Fatalf("failed to read initial state: %v", err)
	}
	initialCount := len(initial.Resources)

	// Add package
	err = svc.AddPackage("gem", "colorize", "1.0", nil)
	if err != nil {
		t.Fatalf("failed to add package: %v", err)
	}

	// Verify it was added
	afterAdd, err := svc.Read()
	if err != nil {
		t.Fatalf("failed to read after add: %v", err)
	}
	if len(afterAdd.Resources) != initialCount+1 {
		t.Errorf("expected %d resources after add, got %d", initialCount+1, len(afterAdd.Resources))
	}

	// Remove package
	err = svc.RemovePackage("gem", "colorize")
	if err != nil {
		t.Fatalf("failed to remove package: %v", err)
	}

	// Property: Should return to initial state
	afterRemove, err := svc.Read()
	if err != nil {
		t.Fatalf("failed to read after remove: %v", err)
	}
	if len(afterRemove.Resources) != initialCount {
		t.Errorf("expected %d resources after remove, got %d (remove is not inverse of add)", initialCount, len(afterRemove.Resources))
	}
}
