// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

// PackageManager defines the minimal interface for package managers.
type PackageManager interface {
	IsAvailable() bool
	ListInstalled() ([]string, error)
}