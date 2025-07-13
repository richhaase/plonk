// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package interfaces provides unified interface definitions for the plonk codebase.
// This package consolidates common interfaces to reduce duplication and improve maintainability.
package interfaces

import "context"

// Provider defines the universal state provider interface for all domains.
// Implementations handle state reconciliation for packages, dotfiles, etc.
type Provider interface {
	// Domain returns the domain name (e.g., "package", "dotfile")
	Domain() string

	// GetConfiguredItems returns items defined in configuration
	GetConfiguredItems() ([]ConfigItem, error)

	// GetActualItems returns items currently present in the system
	GetActualItems(ctx context.Context) ([]ActualItem, error)

	// CreateItem creates an Item from configured and actual data
	CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
}

// ConfigItem represents an item as defined in configuration
type ConfigItem struct {
	Name     string
	Metadata map[string]interface{}
}

// ActualItem represents an item as it exists in the system
type ActualItem struct {
	Name     string
	Path     string
	Metadata map[string]interface{}
}

// ItemState represents the reconciliation state of any managed item
type ItemState int

const (
	StateManaged   ItemState = iota // In config AND present/installed
	StateMissing                    // In config BUT not present/installed
	StateUntracked                  // Present/installed BUT not in config
)

// String returns a human-readable representation of the item state
func (s ItemState) String() string {
	switch s {
	case StateManaged:
		return "managed"
	case StateMissing:
		return "missing"
	case StateUntracked:
		return "untracked"
	default:
		return "unknown"
	}
}

// Item represents any manageable item (package, dotfile, etc.) with its current state
type Item struct {
	Name     string                 `json:"name" yaml:"name"`
	State    ItemState              `json:"state" yaml:"state"`
	Domain   string                 `json:"domain" yaml:"domain"`                         // "package", "dotfile", etc.
	Manager  string                 `json:"manager,omitempty" yaml:"manager,omitempty"`   // "homebrew", "npm", etc.
	Path     string                 `json:"path,omitempty" yaml:"path,omitempty"`         // For dotfiles
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"` // Additional data
}
