// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

import (
	"context"
)

// ReconcileItems compares desired vs actual state and categorizes items.
// This is a convenience wrapper around ReconcileItemsWithKey using Item.Name as the key.
func ReconcileItems(desired, actual []Item) []Item {
	return ReconcileItemsWithKey(desired, actual, func(item Item) string {
		return item.Name
	})
}

// ReconcileItemsWithKey compares desired vs actual state using a custom key function
// This is useful when items need to be compared by more than just name (e.g., manager:name)
func ReconcileItemsWithKey(desired, actual []Item, keyFunc func(Item) string) []Item {
	// Build lookup map for actual items by key
	actualMap := make(map[string]*Item)
	for i := range actual {
		key := keyFunc(actual[i])
		actualMap[key] = &actual[i]
	}

	// Build lookup map for desired items by key
	desiredMap := make(map[string]*Item)
	for i := range desired {
		key := keyFunc(desired[i])
		desiredMap[key] = &desired[i]
	}

	// Result slice to hold all items with their states
	result := make([]Item, 0, len(desired)+len(actual))

	// Check each desired item against actual
	for _, desiredItem := range desired {
		item := desiredItem // Copy the item
		key := keyFunc(desiredItem)
		if actualItem, exists := actualMap[key]; exists {
			// Item is managed (in both desired and actual)
			item.State = StateManaged

			// Check for drift using typed comparator
			if comparator := GetDriftComparator(desiredItem); comparator != nil {
				if identical, err := comparator.Compare(); err == nil && !identical {
					item.State = StateDegraded // Drift detected
					if item.Meta == nil {
						item.Meta = make(map[string]string)
					}
					item.Meta["drift_status"] = "modified"
				}
			}

			// Merge metadata from actual using deep copy to avoid aliasing
			item.Metadata = MergeMetadata(item.Metadata, actualItem.Metadata)

			// Merge Meta string map using deep copy
			item.Meta = MergeStringMap(item.Meta, actualItem.Meta)
			// Use actual path if available (for dotfiles)
			if item.Path == "" && actualItem.Path != "" {
				item.Path = actualItem.Path
			}
			// Use actual type if available
			if item.Type == "" && actualItem.Type != "" {
				item.Type = actualItem.Type
			}
		} else {
			// Item is missing (in desired but not actual)
			item.State = StateMissing
		}
		result = append(result, item)
	}

	// Check each actual item against desired
	for _, actualItem := range actual {
		key := keyFunc(actualItem)
		if _, exists := desiredMap[key]; !exists {
			// Item is untracked (in actual but not desired)
			item := actualItem // Copy the item
			item.State = StateUntracked
			result = append(result, item)
		}
	}

	return result
}

// GroupItemsByState separates items into managed, missing, and untracked slices
// Note: Degraded items (drifted dotfiles) are grouped with managed items
func GroupItemsByState(items []Item) (managed, missing, untracked []Item) {
	managed = make([]Item, 0)
	missing = make([]Item, 0)
	untracked = make([]Item, 0)

	for _, item := range items {
		switch item.State {
		case StateManaged, StateDegraded:
			managed = append(managed, item)
		case StateMissing:
			missing = append(missing, item)
		case StateUntracked:
			untracked = append(untracked, item)
		}
	}

	return managed, missing, untracked
}

// ReconcileResource performs reconciliation for a Resource interface
func ReconcileResource(ctx context.Context, resource Resource) ([]Item, error) {
	desired := resource.Desired()
	actual := resource.Actual(ctx)

	// Use the reconciliation helper from resources package
	reconciled := ReconcileItems(desired, actual)
	return reconciled, nil
}
