// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"slices"
	"sort"
)

// LockV3 represents the simplified v3 lock format
type LockV3 struct {
	Version  int                 `yaml:"version"`
	Packages map[string][]string `yaml:"packages,omitempty"` // manager -> []package
}

// NewLockV3 creates an empty v3 lock
func NewLockV3() *LockV3 {
	return &LockV3{
		Version:  3,
		Packages: make(map[string][]string),
	}
}

// AddPackage adds a package under its manager (maintains sorted order)
func (l *LockV3) AddPackage(manager, pkg string) {
	if l.Packages == nil {
		l.Packages = make(map[string][]string)
	}

	// Check if already exists
	if slices.Contains(l.Packages[manager], pkg) {
		return
	}

	l.Packages[manager] = append(l.Packages[manager], pkg)
	sort.Strings(l.Packages[manager])
}

// RemovePackage removes a package from its manager
func (l *LockV3) RemovePackage(manager, pkg string) {
	if l.Packages == nil {
		return
	}

	pkgs := l.Packages[manager]
	for i, existing := range pkgs {
		if existing == pkg {
			l.Packages[manager] = append(pkgs[:i], pkgs[i+1:]...)
			break
		}
	}

	// Remove manager key if empty
	if len(l.Packages[manager]) == 0 {
		delete(l.Packages, manager)
	}
}

// HasPackage checks if a package is tracked
func (l *LockV3) HasPackage(manager, pkg string) bool {
	return slices.Contains(l.Packages[manager], pkg)
}

// GetPackages returns all packages for a manager
func (l *LockV3) GetPackages(manager string) []string {
	return l.Packages[manager]
}

// GetAllPackages returns all manager:package pairs
func (l *LockV3) GetAllPackages() []string {
	var result []string
	for manager, pkgs := range l.Packages {
		for _, pkg := range pkgs {
			result = append(result, manager+":"+pkg)
		}
	}
	sort.Strings(result)
	return result
}
