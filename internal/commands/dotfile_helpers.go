// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
)

// convertDotfileStatusToOutput converts []dotfiles.DotfileStatus to []output.Item
// This function bridges the dotfiles domain types to output types for rendering.
func convertDotfileStatusToOutput(statuses []dotfiles.DotfileStatus) []output.Item {
	items := make([]output.Item, len(statuses))
	for i, s := range statuses {
		// Convert state
		var state output.ItemState
		switch s.State {
		case dotfiles.SyncStateManaged:
			state = output.StateManaged
		case dotfiles.SyncStateMissing:
			state = output.StateMissing
		case dotfiles.SyncStateDrifted:
			state = output.StateDegraded
		default:
			state = output.StateManaged
		}

		// Build metadata
		metadata := map[string]interface{}{
			"source":      s.Source,
			"destination": s.Target,
		}

		items[i] = output.Item{
			Name:     "." + s.Name, // Add dot prefix for display
			Path:     s.Target,
			State:    state,
			Metadata: metadata,
		}
	}
	return items
}

// resolveDotfilePath resolves a path to an absolute path in home directory
func resolveDotfilePath(path, homeDir string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// Handle tilde
	if len(path) > 0 && path[0] == '~' {
		if len(path) == 1 {
			return homeDir
		}
		if path[1] == '/' {
			return filepath.Join(homeDir, path[2:])
		}
	}

	// Try as relative to home with dot prefix
	return filepath.Join(homeDir, path)
}

// resolveDotfileName resolves a path to the dotfile name (without leading dot)
func resolveDotfileName(path, homeDir string) string {
	absPath := resolveDotfilePath(path, homeDir)

	// Get relative to home
	rel, err := filepath.Rel(homeDir, absPath)
	if err != nil {
		return path
	}

	// Remove leading dot if present
	if len(rel) > 0 && rel[0] == '.' {
		rel = rel[1:]
	}

	return rel
}

// AddStatus represents the status of an add operation
type AddStatus string

const (
	AddStatusAdded       AddStatus = "added"
	AddStatusUpdated     AddStatus = "updated"
	AddStatusWouldAdd    AddStatus = "would-add"
	AddStatusWouldUpdate AddStatus = "would-update"
	AddStatusFailed      AddStatus = "failed"
)

// String returns the string representation
func (s AddStatus) String() string {
	return string(s)
}

// AddResult represents the result of an add operation
type AddResult struct {
	Path           string
	Source         string
	Destination    string
	Status         AddStatus
	AlreadyManaged bool
	Error          error
}

// AddOptions configures dotfile addition
type AddOptions struct {
	DryRun bool
}

// addDotfiles adds files from $HOME to $PLONK_DIR using DotfileManager
func addDotfiles(dm *dotfiles.DotfileManager, configDir, homeDir string, paths []string, opts AddOptions) []AddResult {
	var results []AddResult

	for _, path := range paths {
		result := AddResult{
			Path: path,
		}

		// Resolve path to absolute
		absPath := resolveDotfilePath(path, homeDir)

		// Calculate the relative source path (without dot)
		relPath := resolveDotfileName(absPath, homeDir)
		sourcePath := filepath.Join(configDir, relPath)

		// Check if already managed
		_, existsErr := os.Stat(sourcePath)
		alreadyManaged := existsErr == nil

		// Check if source file exists in home dir
		if _, err := os.Stat(absPath); err != nil {
			result.Status = AddStatusFailed
			result.Error = fmt.Errorf("%s does not exist", absPath)
			results = append(results, result)
			continue
		}

		result.Source = relPath
		result.Destination = absPath
		result.AlreadyManaged = alreadyManaged

		if opts.DryRun {
			if alreadyManaged {
				result.Status = AddStatusWouldUpdate
			} else {
				result.Status = AddStatusWouldAdd
			}
		} else {
			err := dm.Add(absPath)
			if err != nil {
				result.Status = AddStatusFailed
				result.Error = err
			} else if alreadyManaged {
				result.Status = AddStatusUpdated
			} else {
				result.Status = AddStatusAdded
			}
		}

		results = append(results, result)
	}

	return results
}

// RemoveStatus represents the status of a remove operation
type RemoveStatus string

const (
	RemoveStatusRemoved     RemoveStatus = "removed"
	RemoveStatusWouldRemove RemoveStatus = "would-remove"
	RemoveStatusSkipped     RemoveStatus = "skipped"
	RemoveStatusFailed      RemoveStatus = "failed"
)

// String returns the string representation
func (s RemoveStatus) String() string {
	return string(s)
}

// RemoveResult represents the result of a remove operation
type RemoveResult struct {
	Path        string
	Source      string
	Destination string
	Status      RemoveStatus
	Error       error
}

// RemoveOptions configures dotfile removal
type RemoveOptions struct {
	DryRun bool
}

// removeDotfiles removes files from $PLONK_DIR using DotfileManager
func removeDotfiles(dm *dotfiles.DotfileManager, configDir, homeDir string, paths []string, opts RemoveOptions) []RemoveResult {
	var results []RemoveResult

	for _, path := range paths {
		result := RemoveResult{
			Path: path,
		}

		// Resolve path to get the name in config dir
		name := resolveDotfileName(path, homeDir)

		// Check if it exists first
		sourcePath := filepath.Join(configDir, name)
		if _, err := os.Stat(sourcePath); err != nil {
			// File doesn't exist - skip it rather than fail
			result.Status = RemoveStatusSkipped
			result.Source = name
			result.Destination = toTargetPath(name, homeDir)
			results = append(results, result)
			continue
		}

		if opts.DryRun {
			result.Status = RemoveStatusWouldRemove
			result.Source = name
			result.Destination = toTargetPath(name, homeDir)
		} else {
			err := dm.Remove(name)
			if err != nil {
				result.Status = RemoveStatusFailed
				result.Error = err
			} else {
				result.Status = RemoveStatusRemoved
				result.Source = name
				result.Destination = toTargetPath(name, homeDir)
			}
		}

		results = append(results, result)
	}

	return results
}

// toTargetPath converts a source name to its target path in home
// e.g., "zshrc" -> "/home/user/.zshrc"
func toTargetPath(name, homeDir string) string {
	return filepath.Join(homeDir, "."+name)
}
