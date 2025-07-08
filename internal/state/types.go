// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package state provides unified state management capabilities for plonk.
// This package implements the core state reconciliation patterns that are used
// across both package management and dotfile management domains.
package state

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
	Name     string    `json:"name" yaml:"name"`
	State    ItemState `json:"state" yaml:"state"`
	Domain   string    `json:"domain" yaml:"domain"`     // "package", "dotfile", etc.
	Manager  string    `json:"manager,omitempty" yaml:"manager,omitempty"` // "homebrew", "npm", etc.
	Path     string    `json:"path,omitempty" yaml:"path,omitempty"`       // For dotfiles
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"` // Additional data
}

// Result contains the results of state reconciliation for a domain
type Result struct {
	Domain    string `json:"domain" yaml:"domain"`
	Manager   string `json:"manager,omitempty" yaml:"manager,omitempty"`
	Managed   []Item `json:"managed" yaml:"managed"`
	Missing   []Item `json:"missing" yaml:"missing"`
	Untracked []Item `json:"untracked" yaml:"untracked"`
}

// Summary provides aggregate counts across all states
type Summary struct {
	TotalManaged   int `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int `json:"total_untracked" yaml:"total_untracked"`
	Results        []Result `json:"results" yaml:"results"`
}

// Count returns the total number of items in this result
func (r *Result) Count() int {
	return len(r.Managed) + len(r.Missing) + len(r.Untracked)
}

// IsEmpty returns true if this result contains no items
func (r *Result) IsEmpty() bool {
	return r.Count() == 0
}

// AddToSummary adds this result's counts to the provided summary
func (r *Result) AddToSummary(summary *Summary) {
	summary.TotalManaged += len(r.Managed)
	summary.TotalMissing += len(r.Missing)
	summary.TotalUntracked += len(r.Untracked)
	summary.Results = append(summary.Results, *r)
}