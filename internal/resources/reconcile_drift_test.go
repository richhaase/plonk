// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

import (
	"testing"
)

func TestReconcileWithDrift(t *testing.T) {
	t.Run("detects drift when comparison fails", func(t *testing.T) {
		// Create desired items with comparison functions
		desired := []Item{
			{
				Name:   "file1",
				Domain: "dotfile",
				Metadata: map[string]interface{}{
					"compare_fn": func() (bool, error) {
						return true, nil // Files are identical
					},
				},
			},
			{
				Name:   "file2",
				Domain: "dotfile",
				Metadata: map[string]interface{}{
					"compare_fn": func() (bool, error) {
						return false, nil // Files differ (drift detected)
					},
				},
			},
		}

		// Create actual items (both exist)
		actual := []Item{
			{Name: "file1", Domain: "dotfile"},
			{Name: "file2", Domain: "dotfile"},
		}

		// Reconcile
		result := ReconcileItems(desired, actual)

		// Check results
		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// file1 should be managed (no drift)
		if result[0].Name != "file1" || result[0].State != StateManaged {
			t.Errorf("file1: expected StateManaged, got %v", result[0].State)
		}

		// file2 should be degraded (drifted)
		if result[1].Name != "file2" || result[1].State != StateDegraded {
			t.Errorf("file2: expected StateDegraded, got %v", result[1].State)
		}

		// Check drift metadata
		if result[1].Metadata == nil || result[1].Metadata["drift_status"] != "modified" {
			t.Errorf("file2: expected drift_status=modified in metadata, got %v", result[1].Metadata)
		}
	})

	t.Run("handles comparison function errors gracefully", func(t *testing.T) {
		// Create desired item with failing comparison function
		desired := []Item{
			{
				Name:   "file1",
				Domain: "dotfile",
				Metadata: map[string]interface{}{
					"compare_fn": func() (bool, error) {
						return false, nil // Simulate error - treat as managed
					},
				},
			},
		}

		actual := []Item{
			{Name: "file1", Domain: "dotfile"},
		}

		result := ReconcileItems(desired, actual)

		// Should treat as drifted when comparison returns false
		if result[0].State != StateDegraded {
			t.Errorf("Expected StateDegraded when comparison returns false, got %v", result[0].State)
		}
	})

	t.Run("handles missing comparison function", func(t *testing.T) {
		// Create desired item without comparison function
		desired := []Item{
			{
				Name:   "file1",
				Domain: "dotfile",
				Metadata: map[string]interface{}{
					"source": "file1",
				},
			},
		}

		actual := []Item{
			{Name: "file1", Domain: "dotfile"},
		}

		result := ReconcileItems(desired, actual)

		// Should treat as managed when no comparison function
		if result[0].State != StateManaged {
			t.Errorf("Expected StateManaged without comparison function, got %v", result[0].State)
		}
	})

	t.Run("preserves other states correctly", func(t *testing.T) {
		desired := []Item{
			{Name: "managed", Domain: "dotfile"},
			{Name: "missing", Domain: "dotfile"},
			{
				Name:   "drifted",
				Domain: "dotfile",
				Metadata: map[string]interface{}{
					"compare_fn": func() (bool, error) {
						return false, nil
					},
				},
			},
		}

		actual := []Item{
			{Name: "managed", Domain: "dotfile"},
			{Name: "untracked", Domain: "dotfile"},
			{Name: "drifted", Domain: "dotfile"},
		}

		result := ReconcileItems(desired, actual)

		// Verify all states
		foundStates := make(map[string]ItemState)
		for _, item := range result {
			foundStates[item.Name] = item.State
		}

		if foundStates["managed"] != StateManaged {
			t.Errorf("managed: expected StateManaged, got %v", foundStates["managed"])
		}
		if foundStates["missing"] != StateMissing {
			t.Errorf("missing: expected StateMissing, got %v", foundStates["missing"])
		}
		if foundStates["drifted"] != StateDegraded {
			t.Errorf("drifted: expected StateDegraded, got %v", foundStates["drifted"])
		}
		if foundStates["untracked"] != StateUntracked {
			t.Errorf("untracked: expected StateUntracked, got %v", foundStates["untracked"])
		}
	})
}

func TestReconcileItemsWithKeyDrift(t *testing.T) {
	t.Run("detects drift with custom key function", func(t *testing.T) {
		keyFunc := func(item Item) string {
			return item.Domain + ":" + item.Name
		}

		desired := []Item{
			{
				Name:   "file1",
				Domain: "dotfile",
				Metadata: map[string]interface{}{
					"compare_fn": func() (bool, error) {
						return false, nil // Drift detected
					},
				},
			},
		}

		actual := []Item{
			{Name: "file1", Domain: "dotfile"},
		}

		result := ReconcileItemsWithKey(desired, actual, keyFunc)

		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		if result[0].State != StateDegraded {
			t.Errorf("Expected StateDegraded, got %v", result[0].State)
		}
	})
}
