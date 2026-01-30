// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
)

// Result contains reconciliation results for commands
type Result struct {
	Domain    string
	Managed   []DotfileItem
	Missing   []DotfileItem
	Drifted   []DotfileItem
	Untracked []DotfileItem
}

// ReconcileWithConfig reconciles dotfiles and returns status
func ReconcileWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (Result, error) {
	var patterns []string
	if cfg != nil {
		patterns = cfg.IgnorePatterns
	}

	_ = ctx // suppress unused warning

	dm := NewDotfileManager(configDir, homeDir, patterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return Result{}, err
	}

	result := Result{Domain: "dotfile"}
	for _, s := range statuses {
		item := DotfileItem{
			Name:        "." + s.Name,
			Source:      s.Source,
			Destination: s.Target,
		}

		switch s.State {
		case SyncStateManaged:
			item.State = StateManaged
			result.Managed = append(result.Managed, item)
		case SyncStateMissing:
			item.State = StateMissing
			result.Missing = append(result.Missing, item)
		case SyncStateDrifted:
			item.State = StateDegraded
			// Drifted items go to Managed with StateDegraded state - the formatter expects this
			result.Managed = append(result.Managed, item)
		}
	}

	return result, nil
}

// TemplateProcessor is a no-op since templates are no longer supported
type TemplateProcessor struct {
	configDir string
}

// NewTemplateProcessor creates a no-op template processor
func NewTemplateProcessor(configDir string) *TemplateProcessor {
	return &TemplateProcessor{configDir: configDir}
}

// RenderTemplate returns the original content unchanged (templates not supported)
func (p *TemplateProcessor) RenderTemplate(sourcePath string) ([]byte, error) {
	return os.ReadFile(sourcePath)
}

// RenderToBytes returns the original content unchanged (templates not supported)
func (p *TemplateProcessor) RenderToBytes(sourcePath string) ([]byte, error) {
	return os.ReadFile(sourcePath)
}

// ReconcileItems reconciles desired vs actual for legacy compatibility
func ReconcileItems(desired, actual []DotfileItem) []DotfileItem {
	// Build map of actual by destination
	actualMap := make(map[string]DotfileItem)
	for _, item := range actual {
		actualMap[item.Destination] = item
	}

	var result []DotfileItem
	for _, d := range desired {
		if a, exists := actualMap[d.Destination]; exists {
			// File exists, check if drifted
			if d.CompareFunc != nil {
				same, _ := d.CompareFunc()
				if !same {
					d.State = StateDegraded
				} else {
					d.State = StateManaged
				}
			} else {
				// Compare directly
				srcContent, _ := os.ReadFile(d.Source)
				dstContent, _ := os.ReadFile(d.Destination)
				if string(srcContent) != string(dstContent) {
					d.State = StateDegraded
				} else {
					d.State = StateManaged
				}
			}
			_ = a // use the actual item
		} else {
			d.State = StateMissing
		}
		result = append(result, d)
	}

	return result
}

// CommandManager wraps DotfileManager to provide the command interface
type CommandManager struct {
	dm        *DotfileManager
	configDir string
	homeDir   string
}

// NewCommandManager creates a command-compatible manager
func NewCommandManager(homeDir, configDir string, cfg *config.Config) *CommandManager {
	var patterns []string
	if cfg != nil {
		patterns = cfg.IgnorePatterns
	}
	return &CommandManager{
		dm:        NewDotfileManager(configDir, homeDir, patterns),
		configDir: configDir,
		homeDir:   homeDir,
	}
}

// AddFiles adds files from $HOME to $PLONK_DIR
func (m *CommandManager) AddFiles(_ context.Context, _ *config.Config, paths []string, opts AddOptions) ([]AddResult, error) {
	var results []AddResult

	for _, path := range paths {
		result := AddResult{
			Path: path,
		}

		// Resolve path to absolute
		absPath := resolvePath(path, m.homeDir)

		// Check if source file exists in config dir (to determine add vs update)
		relPath := m.dm.toSource(absPath)
		sourcePath := filepath.Join(m.configDir, relPath)
		_, existsErr := m.dm.fs.Stat(sourcePath)
		alreadyManaged := existsErr == nil

		// Check if source file exists in home dir
		if _, err := m.dm.fs.Stat(absPath); err != nil {
			result.Status = AddStatusFailed
			result.Error = fmt.Errorf("%s does not exist", absPath)
			results = append(results, result)
			continue
		}

		// Calculate source and destination
		result.Source = relPath
		result.Destination = absPath
		result.AlreadyManaged = alreadyManaged

		if opts.DryRun {
			// Don't actually add the file in dry-run mode
			if alreadyManaged {
				result.Status = AddStatusWouldUpdate
			} else {
				result.Status = AddStatusWouldAdd
			}
		} else {
			// Actually add the file
			err := m.dm.Add(absPath)
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

	return results, nil
}

// RemoveFiles removes files from $PLONK_DIR
func (m *CommandManager) RemoveFiles(_ *config.Config, paths []string, opts RemoveOptions) ([]RemoveResult, error) {
	var results []RemoveResult

	for _, path := range paths {
		result := RemoveResult{
			Path: path,
		}

		// Resolve path to get the name in config dir
		name := resolveName(path, m.homeDir)

		// Check if it exists first
		sourcePath := filepath.Join(m.configDir, name)
		_, existsErr := m.dm.fs.Stat(sourcePath)
		if existsErr != nil {
			// File doesn't exist - skip it rather than fail
			result.Status = RemoveStatusSkipped
			result.Source = name
			result.Destination = m.dm.toTarget(name)
			results = append(results, result)
			continue
		}

		if opts.DryRun {
			result.Status = RemoveStatusWouldRemove
			result.Source = name
			result.Destination = m.dm.toTarget(name)
		} else {
			err := m.dm.Remove(name)
			if err != nil {
				result.Status = RemoveStatusFailed
				result.Error = err
			} else {
				result.Status = RemoveStatusRemoved
				result.Source = name
				result.Destination = m.dm.toTarget(name)
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// GetConfiguredDotfiles returns list of dotfiles in config directory
func (m *CommandManager) GetConfiguredDotfiles() ([]DotfileItem, error) {
	dotfiles, err := m.dm.List()
	if err != nil {
		return nil, err
	}

	var items []DotfileItem
	for _, d := range dotfiles {
		items = append(items, DotfileItem{
			Name:        "." + d.Name, // Add dot prefix for legacy compatibility
			Source:      d.Source,
			Destination: d.Target,
			State:       StateManaged,
		})
	}

	return items, nil
}

// GetActualDotfiles returns status of all managed dotfiles
func (m *CommandManager) GetActualDotfiles(_ context.Context) ([]DotfileItem, error) {
	statuses, err := m.dm.Reconcile()
	if err != nil {
		return nil, err
	}

	var items []DotfileItem
	for _, s := range statuses {
		item := DotfileItem{
			Name:        "." + s.Name,
			Source:      s.Source,
			Destination: s.Target,
		}

		// Map new states to old states
		switch s.State {
		case SyncStateManaged:
			item.State = StateManaged
		case SyncStateMissing:
			item.State = StateMissing
		case SyncStateDrifted:
			item.State = StateDegraded
		default:
			item.State = StateManaged
		}

		items = append(items, item)
	}

	return items, nil
}

// Diff returns the diff for a dotfile
func (m *CommandManager) Diff(name string) (string, error) {
	// Find the dotfile
	dotfiles, err := m.dm.List()
	if err != nil {
		return "", err
	}

	for _, d := range dotfiles {
		if d.Name == name || "."+d.Name == name {
			return m.dm.Diff(d)
		}
	}

	return "", nil
}

// resolvePath resolves a path to an absolute path in home directory
func resolvePath(path, homeDir string) string {
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

// resolveName resolves a path to the dotfile name (without dot)
func resolveName(path, homeDir string) string {
	absPath := resolvePath(path, homeDir)

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

// NewManagerWithConfig creates a CommandManager (alias for backwards compatibility)
func NewManagerWithConfig(homeDir, configDir string, cfg *config.Config) *CommandManager {
	return NewCommandManager(homeDir, configDir, cfg)
}

