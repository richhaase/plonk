// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

// Expander handles directory expansion for dotfiles
type Expander struct {
	homeDir        string
	expandDirs     []string
	maxDepth       int
	scanner        *Scanner
	duplicateCheck map[string]bool
}

// NewExpander creates a new expander
func NewExpander(homeDir string, expandDirs []string, scanner *Scanner) *Expander {
	return &Expander{
		homeDir:        homeDir,
		expandDirs:     expandDirs,
		maxDepth:       2, // Default max depth for expansion
		scanner:        scanner,
		duplicateCheck: make(map[string]bool),
	}
}

// CheckDuplicate checks if a path has already been processed
func (e *Expander) CheckDuplicate(name string) bool {
	if e.duplicateCheck[name] {
		return true
	}
	e.duplicateCheck[name] = true
	return false
}
