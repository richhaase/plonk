// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import "github.com/richhaase/plonk/internal/interfaces"

// Result contains the results of state reconciliation for a domain
type Result struct {
	Domain    string            `json:"domain" yaml:"domain"`
	Manager   string            `json:"manager,omitempty" yaml:"manager,omitempty"`
	Managed   []interfaces.Item `json:"managed" yaml:"managed"`
	Missing   []interfaces.Item `json:"missing" yaml:"missing"`
	Untracked []interfaces.Item `json:"untracked" yaml:"untracked"`
}

// Summary provides aggregate counts across all states
type Summary struct {
	TotalManaged   int      `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int      `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int      `json:"total_untracked" yaml:"total_untracked"`
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
