// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import "context"

// PackageManager defines the interface for package managers.
// Package managers handle availability checking, listing, installing, and uninstalling packages.
// All methods accept a context for cancellation and timeout support.
type PackageManager interface {
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
	Search(ctx context.Context, query string) ([]string, error)
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Package string `json:"package"`
	Manager string `json:"manager"`
	Found   bool   `json:"found"`
}