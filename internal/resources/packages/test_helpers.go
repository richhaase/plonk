// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

// Test helper functions that are only used in tests

import "strings"

// stringContains checks if string contains substring (case-insensitive)
func stringContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Convert to lowercase for case-insensitive comparison
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// equalPackageInfo compares two PackageInfo structs for equality
func equalPackageInfo(a, b *PackageInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Name == b.Name &&
		a.Version == b.Version &&
		a.Description == b.Description &&
		a.Homepage == b.Homepage &&
		a.Manager == b.Manager &&
		a.Installed == b.Installed
}
