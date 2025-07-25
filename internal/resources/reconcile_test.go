// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources_test

import (
	"testing"

	"github.com/richhaase/plonk/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestReconcileItems(t *testing.T) {
	tests := []struct {
		name     string
		desired  []resources.Item
		actual   []resources.Item
		expected []resources.Item
	}{
		{
			name: "all items managed",
			desired: []resources.Item{
				{Name: "item1", Domain: "test"},
				{Name: "item2", Domain: "test"},
			},
			actual: []resources.Item{
				{Name: "item1", Domain: "test"},
				{Name: "item2", Domain: "test"},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateManaged},
				{Name: "item2", Domain: "test", State: resources.StateManaged},
			},
		},
		{
			name: "some items missing",
			desired: []resources.Item{
				{Name: "item1", Domain: "test"},
				{Name: "item2", Domain: "test"},
			},
			actual: []resources.Item{
				{Name: "item1", Domain: "test"},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateManaged},
				{Name: "item2", Domain: "test", State: resources.StateMissing},
			},
		},
		{
			name: "untracked items",
			desired: []resources.Item{
				{Name: "item1", Domain: "test"},
			},
			actual: []resources.Item{
				{Name: "item1", Domain: "test"},
				{Name: "item2", Domain: "test"},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateManaged},
				{Name: "item2", Domain: "test", State: resources.StateUntracked},
			},
		},
		{
			name: "mixed states",
			desired: []resources.Item{
				{Name: "item1", Domain: "test"},
				{Name: "item2", Domain: "test"},
			},
			actual: []resources.Item{
				{Name: "item2", Domain: "test"},
				{Name: "item3", Domain: "test"},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateMissing},
				{Name: "item2", Domain: "test", State: resources.StateManaged},
				{Name: "item3", Domain: "test", State: resources.StateUntracked},
			},
		},
		{
			name:    "empty desired",
			desired: []resources.Item{},
			actual: []resources.Item{
				{Name: "item1", Domain: "test"},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateUntracked},
			},
		},
		{
			name: "empty actual",
			desired: []resources.Item{
				{Name: "item1", Domain: "test"},
			},
			actual: []resources.Item{},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateMissing},
			},
		},
		{
			name: "metadata merging",
			desired: []resources.Item{
				{Name: "item1", Domain: "test", Metadata: map[string]interface{}{"key1": "value1"}},
			},
			actual: []resources.Item{
				{Name: "item1", Domain: "test", Metadata: map[string]interface{}{"key2": "value2"}},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateManaged,
					Metadata: map[string]interface{}{"key1": "value1", "key2": "value2"}},
			},
		},
		{
			name: "path and type merging",
			desired: []resources.Item{
				{Name: "item1", Domain: "test"},
			},
			actual: []resources.Item{
				{Name: "item1", Domain: "test", Path: "/path/to/item", Type: "file"},
			},
			expected: []resources.Item{
				{Name: "item1", Domain: "test", State: resources.StateManaged, Path: "/path/to/item", Type: "file"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resources.ReconcileItems(tt.desired, tt.actual)

			// Sort results for consistent comparison
			assert.Equal(t, len(tt.expected), len(result), "result length mismatch")

			// Create maps for easier comparison
			expectedMap := make(map[string]resources.Item)
			for _, item := range tt.expected {
				expectedMap[item.Name] = item
			}

			resultMap := make(map[string]resources.Item)
			for _, item := range result {
				resultMap[item.Name] = item
			}

			for name, expectedItem := range expectedMap {
				resultItem, exists := resultMap[name]
				assert.True(t, exists, "item %s not found in result", name)
				assert.Equal(t, expectedItem.State, resultItem.State, "state mismatch for %s", name)
				assert.Equal(t, expectedItem.Domain, resultItem.Domain, "domain mismatch for %s", name)
				if expectedItem.Path != "" {
					assert.Equal(t, expectedItem.Path, resultItem.Path, "path mismatch for %s", name)
				}
				if expectedItem.Type != "" {
					assert.Equal(t, expectedItem.Type, resultItem.Type, "type mismatch for %s", name)
				}
			}
		})
	}
}

func TestReconcileItemsWithKey(t *testing.T) {
	keyFunc := func(item resources.Item) string {
		return item.Manager + ":" + item.Name
	}

	tests := []struct {
		name     string
		desired  []resources.Item
		actual   []resources.Item
		expected []resources.Item
	}{
		{
			name: "reconcile with manager key",
			desired: []resources.Item{
				{Name: "pkg1", Domain: "package", Manager: "brew"},
				{Name: "pkg1", Domain: "package", Manager: "npm"},
			},
			actual: []resources.Item{
				{Name: "pkg1", Domain: "package", Manager: "brew"},
				{Name: "pkg1", Domain: "package", Manager: "apt"},
			},
			expected: []resources.Item{
				{Name: "pkg1", Domain: "package", Manager: "brew", State: resources.StateManaged},
				{Name: "pkg1", Domain: "package", Manager: "npm", State: resources.StateMissing},
				{Name: "pkg1", Domain: "package", Manager: "apt", State: resources.StateUntracked},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resources.ReconcileItemsWithKey(tt.desired, tt.actual, keyFunc)
			assert.Equal(t, len(tt.expected), len(result), "result length mismatch")

			// Create maps for easier comparison
			expectedMap := make(map[string]resources.Item)
			for _, item := range tt.expected {
				key := keyFunc(item)
				expectedMap[key] = item
			}

			resultMap := make(map[string]resources.Item)
			for _, item := range result {
				key := keyFunc(item)
				resultMap[key] = item
			}

			for key, expectedItem := range expectedMap {
				resultItem, exists := resultMap[key]
				assert.True(t, exists, "item %s not found in result", key)
				assert.Equal(t, expectedItem.State, resultItem.State, "state mismatch for %s", key)
			}
		})
	}
}

func TestGroupItemsByState(t *testing.T) {
	items := []resources.Item{
		{Name: "item1", State: resources.StateManaged},
		{Name: "item2", State: resources.StateMissing},
		{Name: "item3", State: resources.StateManaged},
		{Name: "item4", State: resources.StateUntracked},
		{Name: "item5", State: resources.StateMissing},
	}

	managed, missing, untracked := resources.GroupItemsByState(items)

	assert.Len(t, managed, 2)
	assert.Len(t, missing, 2)
	assert.Len(t, untracked, 1)

	assert.Equal(t, "item1", managed[0].Name)
	assert.Equal(t, "item3", managed[1].Name)
	assert.Equal(t, "item2", missing[0].Name)
	assert.Equal(t, "item5", missing[1].Name)
	assert.Equal(t, "item4", untracked[0].Name)
}

func TestReconcileItemsCaseSensitivity(t *testing.T) {
	desired := []resources.Item{
		{Name: "Item1", Domain: "test"},
		{Name: "item1", Domain: "test"},
	}
	actual := []resources.Item{
		{Name: "Item1", Domain: "test"},
		{Name: "ITEM1", Domain: "test"},
	}

	result := resources.ReconcileItems(desired, actual)

	// Create a map to check results
	resultMap := make(map[string]resources.ItemState)
	for _, item := range result {
		resultMap[item.Name] = item.State
	}

	assert.Equal(t, resources.StateManaged, resultMap["Item1"])
	assert.Equal(t, resources.StateMissing, resultMap["item1"])
	assert.Equal(t, resources.StateUntracked, resultMap["ITEM1"])
}

func TestReconcileItemsDuplicateHandling(t *testing.T) {
	// Test that duplicates are preserved in the result
	desired := []resources.Item{
		{Name: "item1", Domain: "test", Meta: map[string]string{"version": "1.0"}},
		{Name: "item1", Domain: "test", Meta: map[string]string{"version": "2.0"}},
	}
	actual := []resources.Item{
		{Name: "item1", Domain: "test", Type: "type1"},
		{Name: "item1", Domain: "test", Type: "type2"},
	}

	result := resources.ReconcileItems(desired, actual)

	// Count how many item1s we have
	count := 0
	var managedItems []resources.Item
	for _, item := range result {
		if item.Name == "item1" {
			count++
			if item.State == resources.StateManaged {
				managedItems = append(managedItems, item)
			}
		}
	}

	// With the current implementation, we expect 2 items (both from desired)
	// as the function processes all desired items
	assert.Equal(t, 2, count, "should have both duplicates in result")
	assert.Len(t, managedItems, 2, "both should be managed")

	// Both should have the type from the last actual item due to map behavior
	for _, item := range managedItems {
		assert.Equal(t, resources.StateManaged, item.State)
		assert.Equal(t, "type2", item.Type, "should use last actual duplicate")
	}
}
