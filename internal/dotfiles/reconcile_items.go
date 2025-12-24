// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

// ReconcileItems reconciles desired and actual dotfiles to determine their states
func ReconcileItems(desired, actual []DotfileItem) []DotfileItem {
	// Build lookup maps
	desiredMap := make(map[string]DotfileItem, len(desired))
	actualMap := make(map[string]DotfileItem, len(actual))

	for _, item := range desired {
		desiredMap[item.Name] = item
	}

	for _, item := range actual {
		actualMap[item.Name] = item
	}

	// Reconcile: determine state for each item
	var reconciled []DotfileItem

	// Process desired items (configured dotfiles)
	for name, desiredItem := range desiredMap {
		if actualItem, exists := actualMap[name]; exists {
			// Item is both configured and present
			// Check for drift if we have a comparison function
			if desiredItem.CompareFunc != nil {
				identical, err := desiredItem.CompareFunc()
				if err != nil {
					// Error during comparison - mark as degraded with error
					desiredItem.State = StateDegraded
					desiredItem.Error = err.Error()
					if desiredItem.Metadata == nil {
						desiredItem.Metadata = make(map[string]interface{})
					}
					desiredItem.Metadata["drift_status"] = "error"
				} else if identical {
					// Content matches - managed
					desiredItem.State = StateManaged
				} else {
					// Content differs - drifted (degraded)
					desiredItem.State = StateDegraded
					if desiredItem.Metadata == nil {
						desiredItem.Metadata = make(map[string]interface{})
					}
					desiredItem.Metadata["drift_status"] = "modified"
				}
			} else {
				// No comparison function - assume managed
				desiredItem.State = StateManaged
			}

			// Use actual item's destination if it's more specific
			if actualItem.Destination != "" && desiredItem.Destination == "" {
				desiredItem.Destination = actualItem.Destination
			}

			reconciled = append(reconciled, desiredItem)
			delete(actualMap, name)
		} else {
			// Item is configured but not present
			desiredItem.State = StateMissing
			reconciled = append(reconciled, desiredItem)
		}
	}

	// Process remaining actual items (untracked dotfiles)
	for _, actualItem := range actualMap {
		actualItem.State = StateUntracked
		reconciled = append(reconciled, actualItem)
	}

	return reconciled
}

// GroupItemsByState groups reconciled items by their state
func GroupItemsByState(items []DotfileItem) (managed, missing, untracked []DotfileItem) {
	for _, item := range items {
		switch item.State {
		case StateManaged:
			managed = append(managed, item)
		case StateMissing:
			missing = append(missing, item)
		case StateUntracked:
			untracked = append(untracked, item)
		case StateDegraded:
			// Drifted items are included in managed list but marked as degraded
			managed = append(managed, item)
		}
	}
	return managed, missing, untracked
}
