// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

// Test helper functions

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
