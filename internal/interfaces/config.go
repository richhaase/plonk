// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package interfaces

// DotfileConfigLoader defines how to load dotfile configuration
type DotfileConfigLoader interface {
	GetDotfileTargets() map[string]string // source -> destination mapping
	GetIgnorePatterns() []string          // ignore patterns for file filtering
	GetExpandDirectories() []string       // directories to expand in dot list
}
