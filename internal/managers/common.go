// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

// PackageManager defines the interface for package managers.
// Package managers handle availability checking, listing, installing, and uninstalling packages.
type PackageManager interface {
	IsAvailable() bool
	ListInstalled() ([]string, error)
	Install(name string) error
	Uninstall(name string) error
	IsInstalled(name string) bool
}