// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

// DriftComparator provides a type-safe way to check if an item has drifted from its desired state
type DriftComparator interface {
	// Compare checks if the actual state matches the desired state
	// Returns true if identical (no drift), false if different (drift detected)
	Compare() (bool, error)
}

// FuncComparator wraps a function as a DriftComparator (for backward compatibility)
type FuncComparator struct {
	compareFunc func() (bool, error)
}

// NewFuncComparator creates a DriftComparator from a function
func NewFuncComparator(fn func() (bool, error)) DriftComparator {
	return &FuncComparator{compareFunc: fn}
}

// Compare implements DriftComparator
func (f *FuncComparator) Compare() (bool, error) {
	return f.compareFunc()
}

// GetDriftComparator retrieves a DriftComparator from item metadata
// Returns nil if no comparator is present
func GetDriftComparator(item Item) DriftComparator {
	if item.Metadata == nil {
		return nil
	}

	// Check for typed comparator first (new style)
	if comparator, ok := item.Metadata["drift_comparator"].(DriftComparator); ok {
		return comparator
	}

	// Fall back to function-in-metadata for backward compatibility (old style)
	if compareFn, ok := item.Metadata["compare_fn"].(func() (bool, error)); ok {
		return NewFuncComparator(compareFn)
	}

	return nil
}
