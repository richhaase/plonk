// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/resources"
)

// DotfileResource adapts dotfile operations to the Resource interface
type DotfileResource struct {
	manager *Manager
	desired []resources.Item
}

// NewDotfileResource creates a new dotfile resource adapter
func NewDotfileResource(manager *Manager) *DotfileResource {
	return &DotfileResource{manager: manager}
}

// ID returns a unique identifier for this resource
func (d *DotfileResource) ID() string {
	return "dotfiles"
}

// Desired returns the desired state (dotfiles that should be present)
func (d *DotfileResource) Desired() []resources.Item {
	return d.desired
}

// SetDesired sets the desired dotfiles
func (d *DotfileResource) SetDesired(items []resources.Item) {
	d.desired = items
}

// Actual returns the actual state (dotfiles currently present)
func (d *DotfileResource) Actual(ctx context.Context) []resources.Item {
	// Use the manager's GetActualDotfiles method which scans based on config
	actualItems, err := d.manager.GetActualDotfiles(ctx)
	if err != nil {
		// If scanning fails, fall back to only checking configured files
		return d.actualFromDesired(ctx)
	}

	// Set state to untracked for all items (will be reconciled)
	for i := range actualItems {
		actualItems[i].State = resources.StateUntracked
	}

	return actualItems
}

// actualFromDesired is a fallback that only checks desired files
func (d *DotfileResource) actualFromDesired(ctx context.Context) []resources.Item {
	var actualItems []resources.Item

	for _, desired := range d.desired {
		// Check if the destination file exists
		destPath, err := d.manager.GetDestinationPath(desired.Path)
		if err != nil {
			continue // Skip invalid paths
		}
		if d.manager.FileExists(destPath) {
			// File exists, so it's in the actual state
			actualItem := desired
			actualItem.State = resources.StateUntracked // Will be reconciled to StateManaged
			actualItems = append(actualItems, actualItem)
		}
	}

	return actualItems
}

// Apply performs the necessary action to move an item to its desired state
func (d *DotfileResource) Apply(ctx context.Context, item resources.Item) error {
	switch item.State {
	case resources.StateMissing:
		// Dotfile should be created/linked
		return d.applyMissing(ctx, item)
	case resources.StateUntracked:
		// Dotfile should be removed
		return d.applyUntracked(ctx, item)
	case resources.StateManaged:
		// Dotfile is already in desired state
		// Could check if content needs updating, but for now do nothing
		return nil
	default:
		return fmt.Errorf("unknown item state: %v", item.State)
	}
}

// applyMissing handles creating/linking a missing dotfile
func (d *DotfileResource) applyMissing(ctx context.Context, item resources.Item) error {
	// The item should have Path information from reconciliation
	if item.Path == "" {
		return fmt.Errorf("missing path information for dotfile %s", item.Name)
	}

	// Get source from metadata
	source, ok := item.Metadata["source"].(string)
	if !ok || source == "" {
		return fmt.Errorf("missing source information for dotfile %s", item.Name)
	}

	// Get destination from metadata
	destination, ok := item.Metadata["destination"].(string)
	if !ok || destination == "" {
		destination = item.Path // Fallback to Path if destination not in metadata
	}

	// Use the manager's ProcessDotfileForApply method
	opts := ApplyOptions{
		DryRun: false,
		Backup: true,
	}

	result, err := d.manager.ProcessDotfileForApply(ctx, source, destination, opts)
	if err != nil {
		return fmt.Errorf("applying dotfile %s: %w", item.Name, err)
	}

	if result.Status != "added" && result.Status != "updated" {
		return fmt.Errorf("unexpected status %s when applying dotfile %s", result.Status, item.Name)
	}

	return nil
}

// applyUntracked handles removing an untracked dotfile
func (d *DotfileResource) applyUntracked(ctx context.Context, item resources.Item) error {
	// For untracked files, we typically don't want to remove them automatically
	// This is a safety measure to prevent accidental deletion of user files
	// In the future, this could be configurable
	return fmt.Errorf("automatic removal of untracked dotfiles not supported for safety")
}
