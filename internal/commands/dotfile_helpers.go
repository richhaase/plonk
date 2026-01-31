// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
)

// convertDotfileStatusToOutput converts []dotfiles.DotfileStatus to separate managed and missing slices.
// Drifted items are included in managed with StateDegraded state.
func convertDotfileStatusToOutput(statuses []dotfiles.DotfileStatus) (managed, missing []output.Item) {
	for _, s := range statuses {
		item := output.Item{
			Name: "." + s.Name,
			Path: s.Target,
			Metadata: map[string]interface{}{
				"source":      s.Source,
				"destination": s.Target,
			},
		}

		switch s.State {
		case dotfiles.SyncStateManaged:
			item.State = output.StateManaged
			managed = append(managed, item)
		case dotfiles.SyncStateMissing:
			item.State = output.StateMissing
			missing = append(missing, item)
		case dotfiles.SyncStateDrifted:
			item.State = output.StateDegraded
			managed = append(managed, item)
		}
	}
	return managed, missing
}

// resolveDotfilePath resolves a path to an absolute path, trying cwd first then home
func resolveDotfilePath(path, homeDir string) string {
	// Absolute paths used as-is
	if filepath.IsAbs(path) {
		return path
	}

	// Handle tilde expansion
	if len(path) > 0 && path[0] == '~' {
		if len(path) == 1 {
			return homeDir
		}
		if path[1] == '/' {
			return filepath.Join(homeDir, path[2:])
		}
	}

	// Relative path - try current directory first, then home directory
	absPath, err := filepath.Abs(path)
	if err == nil {
		if _, statErr := os.Stat(absPath); statErr == nil {
			return absPath // File exists in cwd
		}
	}

	// Try home directory
	homePath := filepath.Join(homeDir, path)
	if _, statErr := os.Stat(homePath); statErr == nil {
		return homePath // File exists in home
	}

	// Fall back to cwd-resolved path (will likely error later, but preserves original behavior)
	if absPath != "" {
		return absPath
	}
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

// resolveDotfileNameForRemoval resolves a path to the dotfile name for removal.
// Unlike resolveDotfilePath, this doesn't check file existence because
// the target in $HOME may not exist (not deployed yet) while the source
// in $PLONK_DIR does exist.
func resolveDotfileNameForRemoval(path, homeDir string) string {
	var absPath string

	if filepath.IsAbs(path) {
		absPath = path
	} else if len(path) > 0 && path[0] == '~' {
		if len(path) == 1 {
			absPath = homeDir
		} else if path[1] == '/' {
			absPath = filepath.Join(homeDir, path[2:])
		} else {
			absPath = filepath.Join(homeDir, path)
		}
	} else {
		// For removal, always resolve relative to home (not cwd)
		// because we're finding which managed dotfile to remove
		absPath = filepath.Join(homeDir, path)
	}

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

		// Validate the path (security checks + existence)
		if err := dm.ValidateAdd(absPath); err != nil {
			result.Status = AddStatusFailed
			result.Error = err
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
		name := resolveDotfileNameForRemoval(path, homeDir)

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
